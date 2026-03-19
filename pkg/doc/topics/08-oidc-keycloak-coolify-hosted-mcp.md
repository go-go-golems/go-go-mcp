---
Title: OIDC, Keycloak, and Coolify for Hosted MCP
Slug: oidc-keycloak-coolify-hosted-mcp
Short: Intern-level guide to how go-go-mcp uses OIDC, how Keycloak fits in, and how to operate a real hosted deployment on Coolify for OpenAI and Claude-compatible MCP clients.
Topics:
- mcp
- oidc
- oauth
- keycloak
- coolify
- security
- deployments
- openai
- claude
Commands:
- mcp start
Flags:
- auth-mode
- auth-resource-url
- oidc-issuer-url
- oidc-discovery-url
- oidc-audience
- oidc-required-scope
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Tutorial
---

This guide explains the full hosted authentication story for `go-go-mcp` in the order a new engineer should learn it. It starts with the general OIDC and OAuth concepts, then narrows to how `go-go-mcp` acts as a protected MCP resource, then narrows further to how Keycloak should be configured as the external issuer, and finally ends with a concrete Coolify deployment model based on the `smailnail` production work.

This guide matters because most MCP auth failures are not caused by one broken line of code. They usually come from a mismatch between four separate systems:

- the MCP server
- the authorization server
- the reverse proxy or hosting platform
- the remote MCP client such as OpenAI or Claude

If you do not understand where each responsibility begins and ends, it is very easy to fix the wrong layer.

## Who This Guide Is For

This guide is for an intern or any engineer who is new to:

- MCP
- OAuth 2.0 and OpenID Connect
- Keycloak
- hosted deployment debugging

You should read this if you need to:

- expose a `go-go-mcp` server over HTTP
- protect `/mcp` with bearer tokens
- use Keycloak as the issuer
- make a deployment work with OpenAI and Claude remote MCP connectors
- debug why a connector keeps getting `401 Unauthorized`

## The Short Mental Model

This section gives the mental model first so the rest of the document is easier to place. In a hosted MCP deployment:

- `go-go-mcp` is usually the protected resource server
- Keycloak is the authorization server and OIDC provider
- the AI product is the OAuth client
- Coolify is only the deployment and routing layer

That means:

- `go-go-mcp` should protect `/mcp`
- `go-go-mcp` should advertise where auth metadata lives
- Keycloak should issue tokens and publish discovery metadata
- the client should discover metadata, register or use a client, complete OAuth, and retry `/mcp` with `Authorization: Bearer ...`

The most important consequence is this: a `401` from `/mcp` is not automatically a server bug. In a healthy OAuth system, the first anonymous `/mcp` request is often supposed to fail. What matters is whether the client then follows the metadata trail and comes back with a valid bearer token.

## Part 1: OIDC and OAuth 2.0 in General

This section explains the protocol pieces without assuming you already know the terminology. It matters because Keycloak configuration only makes sense once you know which actor owns which job.

### OAuth 2.0

OAuth 2.0 is an authorization framework. It lets a client ask an authorization server for permission to call a protected resource on behalf of a user.

In normal prose, that means:

- the user wants Claude or OpenAI to use some tool
- the tool is hosted behind an HTTP endpoint such as `/mcp`
- the tool server does not want anonymous callers using it
- the AI client must obtain an access token first

The access token is the credential the protected resource accepts.

### OpenID Connect

OpenID Connect, usually shortened to OIDC, is an identity layer built on top of OAuth 2.0. It adds standard discovery documents and identity claims so clients and servers can agree on:

- the issuer
- the authorization endpoint
- the token endpoint
- the JWKS endpoint
- optional registration endpoints
- the meaning of identity claims such as `iss`, `sub`, `email`, and `preferred_username`

For `go-go-mcp`, the most important OIDC outputs are:

- discovery metadata so the client knows where to go
- signing keys so the server can verify JWTs
- issuer identity so the server can reject tokens from the wrong system

### Actors in the Protocol

These roles are easy to blur together, so keep them separate.

- Resource server:
  The server that protects the API or MCP endpoint. In this project, that is `go-go-mcp` or an app built on top of it.
- Authorization server:
  The server that logs the user in and issues tokens. In production, this is usually Keycloak.
- OAuth client:
  The application that wants a token. For this guide, that is a remote MCP client like OpenAI or Claude.
- User:
  The human who approves the connection.

### The Normal Authorization Code Flow

This is the normal happy-path sequence for a hosted MCP deployment:

```text
1. Client calls /mcp without a token
2. Resource server returns 401 + WWW-Authenticate
3. Client fetches protected resource metadata
4. Client discovers the issuer and auth server metadata
5. Client registers or selects an OAuth client
6. Client sends the user through login and consent
7. Client exchanges the code for an access token
8. Client retries /mcp with Authorization: Bearer <token>
9. Resource server validates the token and serves MCP
```

When something goes wrong, the system usually breaks at one of those numbered steps. That is why careful logs are so important.

### Why `WWW-Authenticate` Matters

When an anonymous caller hits `/mcp`, the server needs to do more than just say "no." It needs to tell the client where to learn how auth works.

The `WWW-Authenticate` response header is the standardized place for that. In the current `go-go-mcp` external OIDC model, the important piece is `resource_metadata`, for example:

```text
WWW-Authenticate: Bearer realm="mcp", resource_metadata="https://example.com/.well-known/oauth-protected-resource"
```

This matters because brittle clients may fail if the challenge is noisy, ambiguous, or points at the wrong resource metadata URL.

### Protected Resource Metadata

Protected resource metadata is a JSON document served from:

```text
/.well-known/oauth-protected-resource
```

Its job is to answer two questions:

- Which resource URL is protected?
- Which authorization server protects it?

Example:

```json
{
  "authorization_servers": [
    "https://auth.example.com/realms/myrealm"
  ],
  "resource": "https://mcp.example.com/mcp"
}
```

The `resource` value is not decorative. It must exactly match the public resource URL the client is trying to reach. Trailing slash mismatches, scheme mismatches, or host mismatches can break the flow.

### Dynamic Client Registration

Dynamic Client Registration, or DCR, is the OAuth mechanism where the client asks the authorization server to create a new client automatically at runtime instead of using a manually pre-created client.

DCR is useful because it removes manual setup work, but it also introduces policy complexity:

- which redirect URIs are allowed
- which scopes are allowed
- whether public or confidential clients are allowed
- whether anonymous registration is allowed at all

This matters directly for Claude and OpenAI because remote MCP connectors often try to use DCR. If DCR is blocked by policy, the MCP server may look broken even though the failure is actually inside Keycloak.

## Part 2: How `go-go-mcp` Uses OIDC

This section explains how the framework behaves. It matters because you need to know what `go-go-mcp` is responsible for and what it deliberately delegates to an external issuer.

### The Two Auth Modes

The main framework reference is [07-embedded-oidc.md](./07-embedded-oidc.md). `go-go-mcp` supports two HTTP auth modes:

- `embedded_dev`
- `external_oidc`

`embedded_dev` means the process itself acts as both:

- the MCP resource server
- the authorization server

`external_oidc` means:

- `go-go-mcp` is only the MCP resource server
- an external issuer such as Keycloak owns login, token issuance, discovery, and JWKS

For real hosted deployments, `external_oidc` is the production mode to prefer.

### Where the Framework Wires Auth

The main auth wiring lives in these files:

- `pkg/embeddable/auth_provider.go`
- `pkg/embeddable/auth_provider_external.go`
- `pkg/embeddable/mcpgo_backend.go`
- `pkg/embeddable/server.go`
- `pkg/embeddable/command.go`

The important design point is that the HTTP transport mounts:

- `/mcp`
- `/.well-known/oauth-protected-resource`

and wraps `/mcp` with bearer-token validation.

In external OIDC mode, the framework validates tokens by using:

- issuer metadata
- discovery metadata
- JWKS
- optional audience checks
- optional scope checks

The framework does not log the user in itself in this mode. That job belongs to Keycloak.

### The Main Runtime Contract

If you enable `external_oidc`, the following statements should be true:

- unauthenticated `/mcp` returns `401`
- `/.well-known/oauth-protected-resource` is public
- the `resource` value equals the real public `/mcp` URL
- the issuer URL points at the real external issuer
- a valid bearer token from that issuer is accepted

If any of those statements is false, you should expect connector failures.

### The Core Flags

The important framework flags are:

- `--auth-mode=external_oidc`
- `--auth-resource-url=https://public.example.com/mcp`
- `--oidc-issuer-url=https://auth.example.com/realms/myrealm`
- `--oidc-discovery-url=...` if you must override discovery
- `--oidc-audience=...` if you want audience enforcement
- `--oidc-required-scope=...` if you want scope enforcement

The practical meaning is:

- `auth-resource-url` tells clients what the protected resource actually is
- `oidc-issuer-url` tells the framework which issuer to trust
- audience and scope checks let you harden the deployment once the client behavior is stable

### Pseudocode for External OIDC Mode

This pseudocode shows the server's responsibility boundary:

```text
on request to /mcp:
  if Authorization header missing:
    return 401 with WWW-Authenticate pointing to resource metadata

  token = parse bearer token
  discovery = load or refresh issuer metadata
  jwks = load or refresh issuer keys
  claims = validate JWT signature and issuer

  if audience configured:
    reject token if audience does not match

  if required scopes configured:
    reject token if scopes do not include them

  inject authenticated identity into request context
  pass request to MCP handler
```

This is why an upstream DCR or login failure still appears as repeated anonymous `/mcp` hits in the resource-server logs. The resource server can only validate a token after the client actually gets one.

## Part 3: Keycloak in Particular

This section explains how Keycloak fits into the system. It matters because most real hosted failures happen here, not in the MCP handler itself.

### What Keycloak Is Doing Here

In this architecture, Keycloak is the external issuer. That means Keycloak owns:

- login UI
- authorization code flow
- token issuance
- JWKS publication
- client registration
- scope policy
- redirect URI policy

`go-go-mcp` should not duplicate those responsibilities in production.

### Realm Design

At minimum, you need a realm that contains:

- a stable issuer URL
- one or more MCP-related clients
- the scopes you expect MCP clients to request
- a registration policy if clients use DCR

For the `smailnail` production deployment, the issuer is:

```text
https://auth.scapegoat.dev/realms/smailnail
```

That exact issuer must be the one:

- published in discovery
- configured into the MCP service
- used to verify the token's `iss` claim

### Client Models

There are two broad operational choices:

- Pre-provisioned client:
  You create the OAuth client in Keycloak ahead of time and configure the connector to use it.
- Dynamic client registration:
  You let the connector register its own client at runtime.

The pre-provisioned path is usually easier to reason about and safer to audit. DCR is more flexible, but policy mistakes are common and client behavior can vary between vendors.

### Why DCR Fails So Often

DCR usually fails for one of these reasons:

- anonymous registration is disabled
- the redirect URI is not allowed
- requested scopes are not allowed
- the client type is not allowed
- trusted hosts or hostname policy rejects the request
- the connector asks for a scope or capability you did not expect

In our live Claude debugging, the actual Keycloak failure was:

```text
CLIENT_REGISTER_ERROR
error="not_allowed"
Requested scope 'service_account' not trusted
```

That is a Keycloak policy failure, not a `go-go-mcp` `/mcp` route failure.

### Recommended Keycloak Strategy

For production, the safer default is:

1. start with a pre-provisioned dedicated client per external MCP product
2. only enable anonymous DCR if the product truly requires it
3. if you enable DCR, keep the allowed scopes and redirects narrow

That recommendation exists because the security and operability tradeoff is better:

- easier to audit
- easier to document
- fewer moving parts during login
- fewer surprises from vendor-specific DCR payloads

### A Good Dedicated Client Shape

For a Claude-specific client, a reasonable starting point is:

```json
{
  "clientId": "smailnail-claude-mcp",
  "enabled": true,
  "protocol": "openid-connect",
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": false,
  "serviceAccountsEnabled": false,
  "publicClient": true,
  "redirectUris": [
    "https://claude.ai/api/mcp/auth_callback",
    "https://claude.com/api/mcp/auth_callback"
  ]
}
```

Why those settings matter:

- `standardFlowEnabled=true` enables authorization code flow
- `serviceAccountsEnabled=false` avoids unnecessary capability
- explicit callback URLs reduce attack surface
- a dedicated client keeps product-specific behavior separate from other apps

### `kcadm.sh`

`kcadm.sh` is Keycloak's admin CLI. It is just a command-line wrapper around the Keycloak admin REST API. It matters because it gives you a reproducible and scriptable way to inspect or update production config without relying on memory or screenshots from the web UI.

Typical uses:

- authenticate as the admin user
- inspect a realm
- inspect clients
- inspect registration policy components
- create or update clients

Example login command:

```bash
/opt/keycloak/bin/kcadm.sh config credentials \
  --server http://127.0.0.1:8080 \
  --realm master \
  --user "$KC_BOOTSTRAP_ADMIN_USERNAME" \
  --password "$KC_BOOTSTRAP_ADMIN_PASSWORD"
```

Example inspection commands:

```bash
/opt/keycloak/bin/kcadm.sh get realms/smailnail
/opt/keycloak/bin/kcadm.sh get clients -r smailnail -q clientId=smailnail-mcp
/opt/keycloak/bin/kcadm.sh get components -r smailnail
```

### Keycloak Troubleshooting Order

When a connector fails, inspect Keycloak in this order:

1. Is the issuer URL reachable?
2. Does discovery include the registration, authorization, token, and JWKS endpoints?
3. Did the client ever hit discovery?
4. Did the client attempt DCR?
5. Did Keycloak reject DCR by policy?
6. Did the user ever reach `/auth` or `/authorize`?
7. Did token exchange happen?

That order matters because it prevents you from treating "no token ever existed" as "resource server rejected a valid token."

## Part 4: OpenAI and Claude as Real MCP Clients

This section covers the practical client differences. It matters because different vendors often fail at different steps of the same protocol.

### OpenAI

OpenAI tooling can work in two different practical modes:

- a bearer-token control path where you already have a token
- a remote connector path that may use DCR and OAuth on your behalf

That means "it works with OpenAI" is not always proof that the full remote connector OAuth dance is healthy. Sometimes it only proves that the resource server accepts a valid bearer token once one exists.

The strongest OpenAI-specific reference in this repository is:

- `ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/reference/03-openai-keycloak-dcr-debug-guide.md`

### Claude

Claude connector behavior in our real deployment has shown a different shape:

- it probes `/mcp`
- it fetches protected resource metadata
- it attempts Keycloak DCR
- Keycloak rejects the DCR request
- Claude therefore never returns with a bearer token

The strongest Claude-specific reference in this repository is:

- `ttmp/2026/03/18/SMAILNAIL-016-CLAUDE-MCP-LOGIN-ANALYSIS--analyze-claude-mcp-login-failures-against-keycloak-backed-smailnail-mcp/design-doc/01-claude-mcp-oauth-and-keycloak-dynamic-client-registration-guide-for-smailnail.md`

### Why Client-Specific Docs Matter

Two clients can both support "MCP over OAuth" and still differ in:

- whether they insist on DCR
- whether they support pre-provisioned clients
- what default scopes they request
- which callback URLs they use
- how strict they are about metadata shape

That is why a protocol-level success in one client does not automatically mean another client is misbehaving. Sometimes the integration assumptions are genuinely different.

## Part 5: Coolify as the Real Deployment Case

This section turns the abstract protocol into a concrete hosted shape. It matters because public URLs, proxy routing, and health checks are where many otherwise-correct auth setups fail.

### What Coolify Does and Does Not Do

Coolify is the deployment manager and reverse proxy surface. It does not do OAuth for you. It does not make a wrong `resource` URL correct. It does not fix Keycloak policy.

Its actual responsibilities are:

- building and running the container
- setting environment variables
- exposing the public hostname
- probing health checks
- routing HTTPS traffic to the app's internal port

That means Coolify bugs usually show up as:

- wrong public hostnames
- wrong health-check paths
- wrong environment values
- missing or duplicated runtime variables

### Concrete `smailnail` Deployment Shape

The main concrete deployment reference is:

- `smailnail/docs/deployments/smailnaild-merged-coolify.md`

The merged hosted server serves:

- `/`
- `/auth/*`
- `/api/*`
- `/.well-known/oauth-protected-resource`
- `/mcp`
- `/readyz`

The hosted MCP route and the browser session routes share one HTTP listener, but they do not share the same auth mechanism.

The practical split is:

- browser UI uses session cookies after browser OIDC login
- MCP uses bearer tokens on `/mcp`

### The Concrete Environment Model

For the merged hosted deployment, the key environment variables are:

```env
SMAILNAILD_MCP_ENABLED=1
SMAILNAILD_MCP_TRANSPORT=streamable_http
SMAILNAILD_MCP_AUTH_MODE=external_oidc
SMAILNAILD_MCP_AUTH_RESOURCE_URL=https://smailnail.scapegoat.dev/mcp
SMAILNAILD_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail
```

Optional hardening:

```env
SMAILNAILD_MCP_OIDC_AUDIENCE=smailnail-mcp
SMAILNAILD_MCP_OIDC_REQUIRED_SCOPES=mcp:invoke
```

Those values map directly onto the framework's external OIDC auth settings.

### Why Public URL Accuracy Is Critical

The public MCP resource URL must be the actual URL the connector uses. If your hosted app is at:

```text
https://smailnail.scapegoat.dev/mcp
```

then the protected resource metadata must say:

```json
{
  "resource": "https://smailnail.scapegoat.dev/mcp"
}
```

It must not say:

- `https://smailnail.scapegoat.dev`
- `https://smailnail.scapegoat.dev/mcp/`
- `http://smailnail.scapegoat.dev/mcp`
- the container's internal port URL

This is one of the highest-value checks to perform first.

### Why Health Checks Need Care

Coolify health checks should hit a liveness endpoint such as `/readyz`, not `/mcp`.

Reasons:

- `/mcp` is intentionally protected and returns `401` without a token
- `/.well-known/oauth-protected-resource` proves metadata is alive but not the whole app
- `/readyz` is the correct place to say "the process is up"

One practical cleanup from the hosted deployment work was suppressing noisy `/readyz` logs so health probes do not drown out useful auth diagnostics.

### Concrete Validation Sequence

After a deployment, validate in this order:

1. `GET /readyz` returns `200`
2. `GET /.well-known/oauth-protected-resource` returns the expected `resource` and `authorization_servers`
3. anonymous `POST /mcp` returns `401` with the expected `WWW-Authenticate`
4. Keycloak discovery is reachable
5. a known-good token is accepted on `/mcp`
6. only after that, test OpenAI or Claude connector login

This order is efficient because it isolates:

- app liveness
- metadata correctness
- resource-server auth behavior
- external issuer reachability
- end-to-end product-specific connector behavior

## Part 6: A Concrete Failure Analysis Walkthrough

This section shows how to reason about a real failure. It matters because debugging flows is easier when you can map the logs back onto the abstract sequence.

### Example: Claude Reaches MCP but Fails Login

Suppose your logs show:

```text
POST /mcp -> 401
GET /.well-known/oauth-protected-resource -> 200
GET /mcp -> 401
GET /mcp -> 401
```

At first glance, that might look like the MCP service is wrong. But the correct interpretation is:

- the client found the MCP endpoint
- the client understood there was auth
- the client at least fetched protected resource metadata
- the client still never returned with a bearer token

At that point, the next place to inspect is the auth server, not the `/mcp` handler.

### Example: Keycloak Rejects DCR

Suppose Keycloak logs:

```text
CLIENT_REGISTER_ERROR
error="not_allowed"
Requested scope 'service_account' not trusted
```

That means:

- the connector really did reach Keycloak
- the connector tried dynamic client registration
- Keycloak policy rejected the registration request
- the OAuth flow died before token issuance

That is not fixed by changing the MCP route handler. It is fixed by:

- changing DCR policy
- or avoiding DCR via a pre-provisioned client

### The Debugging Rule

Always identify the last successful step in the protocol.

That rule sounds obvious, but it saves hours. For example:

- if the last success is `/.well-known/oauth-protected-resource`, inspect issuer discovery next
- if the last success is Keycloak discovery, inspect DCR or authorize next
- if the last success is `/token`, inspect bearer validation next

## Part 7: Recommended Implementation Strategy

This section gives a practical recommendation rather than only describing options. It matters because an intern needs a default path, not just a menu of possibilities.

### Recommended Default for Production

Use this order:

1. deploy `go-go-mcp` in `external_oidc` mode
2. make the public resource metadata exact
3. create a dedicated pre-provisioned client per major external MCP product
4. get the control path working with a real bearer token
5. only then decide whether anonymous DCR is necessary

This is the preferred path because it minimizes the number of simultaneously moving parts.

### Recommended Default for Local Development

For local demos or framework development:

- use `embedded_dev` if you want a self-contained local auth flow
- use a local Keycloak fixture if you specifically need to test external issuer behavior

Do not confuse the two goals:

- `embedded_dev` is for convenience
- `external_oidc` is the production shape

### Decision Table

| Need | Recommended choice | Why |
| --- | --- | --- |
| Fast local demo | `embedded_dev` | Fewest dependencies |
| Production hosted MCP | `external_oidc` | Correct trust boundary |
| Tightest security posture | Pre-provisioned clients | Easier audit and rollback |
| Maximum connector self-service | DCR | More flexible, but more policy risk |
| First production rollout | Avoid DCR if possible | Fewer hidden failures |

## Part 8: Operator Checklist

This section gives the actual checklist to follow when setting up or reviewing a deployment. It matters because consistency is more valuable than heroics during an incident.

### Resource Server Checklist

- `auth-mode` is `external_oidc`
- `auth-resource-url` equals the real public `/mcp` URL
- issuer URL matches Keycloak's actual issuer exactly
- anonymous `/mcp` returns `401`
- `/.well-known/oauth-protected-resource` is public
- a known-good token is accepted

### Keycloak Checklist

- realm exists and is enabled
- discovery is public and reachable
- JWKS is public and reachable
- callbacks are explicitly allowed
- client uses authorization code flow
- unnecessary grant types are disabled
- DCR policy is either intentionally allowed or intentionally avoided

### Coolify Checklist

- public domain routes to the correct internal port
- health check uses `/readyz`
- runtime environment variables are present only once
- TLS/public host seen by the client matches the metadata you advertise

### Connector Checklist

- connector is pointed at the exact `/mcp` URL
- callback URLs are registered exactly
- if using a pre-provisioned client, the connector is configured to use it
- if using DCR, Keycloak policy is intentionally configured for it

## Troubleshooting

This section lists the most common hosted MCP/OIDC failures and what they usually mean.

| Problem | Cause | Solution |
| --- | --- | --- |
| `/mcp` returns `401` | Could be healthy anonymous challenge behavior | Check whether the client follows metadata and later retries with a bearer token |
| Client fetches protected resource metadata but nothing else happens | Client may reject metadata or fail before auth discovery | Verify `resource` exactness and inspect auth-server logs |
| Keycloak logs `CLIENT_REGISTER_ERROR` | DCR policy rejected the request | Relax the specific DCR policy or use a pre-provisioned client |
| Token works with manual curl but not with remote connector | Connector path differs from manual bearer-token path | Inspect DCR, callback URL, and connector-specific behavior |
| Connector loops on `/mcp` anonymously | OAuth never completed | Check discovery, DCR, authorize, and token endpoints in order |
| Health checks flood the logs | Probe path is noisy or over-logged | Use `/readyz` and suppress per-probe debug noise if necessary |
| Token rejected despite successful login | Issuer, audience, or scope mismatch | Compare token claims to server expectations |

## See Also

- [Embedded OIDC in go-go-mcp](./07-embedded-oidc.md)
- `ttmp/2026/03/09/MCP-003-KEYCLOAK-EXTERNAL-OIDC-DESIGN--support-external-keycloak-oidc-with-embedded-dev-login-for-go-go-mcp/design-doc/01-go-go-mcp-external-keycloak-oidc-and-embedded-dev-login-architecture-guide.md`
- `ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/reference/03-openai-keycloak-dcr-debug-guide.md`
- `ttmp/2026/03/18/SMAILNAIL-016-CLAUDE-MCP-LOGIN-ANALYSIS--analyze-claude-mcp-login-failures-against-keycloak-backed-smailnail-mcp/design-doc/01-claude-mcp-oauth-and-keycloak-dynamic-client-registration-guide-for-smailnail.md`
- `smailnail/docs/deployments/smailnaild-merged-coolify.md`
