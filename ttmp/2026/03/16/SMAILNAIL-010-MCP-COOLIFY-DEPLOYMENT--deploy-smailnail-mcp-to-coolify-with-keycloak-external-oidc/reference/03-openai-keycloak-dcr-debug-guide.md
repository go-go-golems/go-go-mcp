---
Title: OpenAI Connector and Keycloak DCR debug guide
Ticket: SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - keycloak
    - oauth
    - deployments
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../smailnail/docs/deployments/smailnail-imap-mcp-coolify.md
      Note: |-
        Documents the deployed public MCP URL and Keycloak issuer used in this analysis
        Documents the live MCP and issuer URLs analyzed in this guide
    - Path: pkg/embeddable/auth_provider_external.go
      Note: |-
        Confirms the hosted MCP advertises OAuth metadata and validates issuer, JWKS, optional audience, and scopes after token acquisition
        Shows the hosted MCP only validates tokens after DCR succeeds and explains why the current failure is upstream in Keycloak
    - Path: pkg/embeddable/mcpgo_backend.go
      Note: |-
        Confirms the hosted MCP serves /.well-known/oauth-protected-resource and logs unauthenticated /mcp requests
        Shows the hosted MCP serves protected resource metadata and logs anonymous /mcp requests from failed connector attempts
    - Path: ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/smoke_hosted_mcp_oidc.go
      Note: |-
        Provides a known-good non-DCR authenticated MCP client path for comparison against the failing OpenAI connector flow
        Provides a known-good bearer-auth MCP control path that bypasses DCR entirely
ExternalSources:
    - https://codegym.cc/quests/lectures/en.codegym.chatgptapp.next.lecture.level10.lecture03
    - https://kane.mx/posts/2025/deploy-keycloak-aws-mcp-oauth/
Summary: Fresh analysis of why OpenAI connector setup against the hosted smailnail MCP still fails, including live Keycloak/MCP evidence, the misleading RFC 7591 error message, and a concrete debug playbook.
LastUpdated: 2026-03-16T12:05:00-04:00
WhatFor: Explain the real failure sequence between OpenAI, Keycloak, and the hosted MCP deployment and provide a precise operator guide for unblocking Dynamic Client Registration.
WhenToUse: Use when OpenAI or another remote MCP connector reports a generic DCR or OAuth failure against the hosted smailnail MCP service.
---


# OpenAI Connector and Keycloak DCR debug guide

## Executive summary

The current OpenAI error message, "The server `smailnail.mcp.scapegoat.dev` doesn't support RFC 7591 Dynamic Client Registration", is misleading. The live system **does** advertise a valid `registration_endpoint`, and OpenAI **is** making Dynamic Client Registration attempts against Keycloak.

The real problem is that Keycloak is rejecting those registration attempts through **anonymous client registration policies**. During this investigation, the active blockers were:

- `Trusted Hosts`
- `Allowed Client Scopes`

Because registration never completes, OpenAI never acquires a bearer token, and the hosted MCP only sees anonymous metadata and `/mcp` requests.

## Sources reviewed

### CodeGym Keycloak lecture

The CodeGym lecture is directionally correct for MCP/OAuth basics:

- the MCP client should behave as a **public client**
- Authorization Code + **PKCE** is the expected interactive flow
- Keycloak must expose discovery metadata and a usable `registration_endpoint`

The lecture also emphasizes that redirect URIs and scopes must be aligned exactly. That matters here because anonymous DCR policies in Keycloak are effectively rejecting the scopes OpenAI is trying to register for.

### Kane Keycloak + MCP article

The Kane article is more directly relevant to the current problem. The most useful takeaways are:

- Keycloak can work as an MCP auth server, but **anonymous DCR policy configuration matters**
- for MCP clients, the author explicitly updates the **Allowed Client Scopes** registration policy through the Admin REST API, including `openid`
- Keycloak still has an **RFC 8707 / resource indicator gap**, so audience/resource handling may need workarounds even after DCR succeeds

That matches the live system closely:

- our current failure is happening **before** token issuance, in DCR policy evaluation
- after DCR is fixed, **resource/audience** may become the next issue to watch

## Live system evidence

### Hosted MCP is advertising the expected OAuth resource metadata

Protected resource metadata:

```json
{
  "resource": "https://smailnail.mcp.scapegoat.dev/mcp",
  "authorization_servers": [
    "https://auth.scapegoat.dev/realms/smailnail"
  ]
}
```

OIDC discovery:

```json
{
  "issuer": "https://auth.scapegoat.dev/realms/smailnail",
  "registration_endpoint": "https://auth.scapegoat.dev/realms/smailnail/clients-registrations/openid-connect",
  "authorization_endpoint": "https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/auth",
  "token_endpoint": "https://auth.scapegoat.dev/realms/smailnail/protocol/openid-connect/token",
  "code_challenge_methods_supported": ["plain", "S256"],
  "scopes_supported": ["openid", "offline_access", "microprofile-jwt", "organization", "roles", "web-origins", "address", "basic", "service_account", "profile", "phone", "email", "acr"]
}
```

So the OpenAI message is **not** caused by a missing registration endpoint.

### Hosted MCP logs show OpenAI never gets a bearer token

Recent `smailnail-mcp` logs showed requests like:

```text
Unauthorized: missing bearer header method=POST path=/mcp ua="Python/3.12 aiohttp/3.13.3"
served protected resource metadata endpoint=/.well-known/oauth-protected-resource ... ua="Python/3.12 aiohttp/3.13.3"
```

That means the connector is reaching the protected resource and the MCP endpoint, but it is still unauthenticated. That is consistent with a DCR or OAuth failure **upstream** in Keycloak.

### Keycloak logs show multiple real DCR attempts and multiple policy failures

Observed Keycloak failures from live logs:

```text
Policy 'Trusted Hosts' rejected request to client-registration service. Details: Host not trusted.
ipAddress="52.176.139.180"
```

```text
Policy 'Trusted Hosts' rejected request to client-registration service. Details: Host not trusted.
ipAddress="52.176.139.187"
```

```text
Policy 'Trusted Hosts' rejected request to client-registration service. Details: Host not trusted.
ipAddress="52.176.139.178"
```

```text
Policy 'Trusted Hosts' rejected request to client-registration service. Details: Host not trusted.
ipAddress="52.176.139.181"
```

```text
Policy 'Trusted Hosts' rejected request to client-registration service. Details: Host not trusted.
ipAddress="52.176.139.183"
```

```text
Policy 'Trusted Hosts' rejected request to client-registration service. Details: Host not trusted.
ipAddress="52.176.139.190"
```

And also:

```text
Requested scope 'openid' not trusted in the list: [role_list, saml_organization, profile, email, roles, web-origins, acr, basic, offline_access, address, phone, microprofile-jwt, organization]
Policy 'Allowed Client Scopes' rejected request to client-registration service. Details: Not permitted to use specified clientScope
```

Observed from:

- OpenAI-related source IPs such as `52.176.139.180`, `52.176.139.187`, `52.176.139.178`, `52.176.139.181`, `52.176.139.183`, `52.176.139.190`
- direct probe traffic from workstation IPs such as `72.200.191.248`, `160.79.106.122`, `160.79.106.11`

### Direct DCR probe still fails

A direct POST to the advertised `registration_endpoint` with a standard OIDC client payload currently returns:

```text
403
{"error":"insufficient_scope","error_description":"Policy 'Allowed Client Scopes' rejected request to client-registration service. Details: Not permitted to use specified clientScope"}
```

That means the realm is still not usable for anonymous DCR even outside OpenAI.

## What the current Keycloak policy state actually is

Anonymous client registration policy dump from the live realm:

```json
[
  {
    "name": "Max Clients Limit",
    "providerId": "max-clients",
    "subType": "anonymous"
  },
  {
    "name": "Allowed Client Scopes",
    "providerId": "allowed-client-templates",
    "subType": "anonymous",
    "config": {
      "allow-default-scopes": ["true"]
    }
  },
  {
    "name": "Full Scope Disabled",
    "providerId": "scope",
    "subType": "anonymous"
  },
  {
    "name": "Allowed Protocol Mapper Types",
    "providerId": "allowed-protocol-mappers",
    "subType": "anonymous"
  },
  {
    "name": "Trusted Hosts",
    "providerId": "trusted-hosts",
    "subType": "anonymous",
    "config": {
      "host-sending-registration-request-must-match": ["true"],
      "client-uris-must-match": ["true"]
    }
  },
  {
    "name": "Consent Required",
    "providerId": "consent-required",
    "subType": "anonymous"
  }
]
```

The important findings are:

- `Trusted Hosts` is active for anonymous DCR and currently rejects OpenAI’s rotating source IPs
- `Allowed Client Scopes` is active for anonymous DCR, but its config does **not** show an explicit allowed scope list containing `openid`
- the provider is labeled `allowed-client-templates` in the dump even though the UI calls it `Allowed Client Scopes`

Also note:

```json
{
  "realm": "smailnail",
  "registrationAllowed": false
}
```

This field came from the realm representation. It does **not** appear to be the immediate cause of the observed failures, because Keycloak is clearly evaluating anonymous client registration policies already. It is likely unrelated user-registration state rather than the decisive DCR gate in this flow.

## Fresh analysis

## 1. The OpenAI message is generic, not diagnostic

OpenAI is not accurately reporting the precise failure. The live evidence shows:

- OpenAI discovers the auth server
- OpenAI reaches the Keycloak DCR endpoint
- Keycloak rejects DCR policy checks

So the issue is not "RFC 7591 unsupported" in the literal sense. It is "RFC 7591 registration attempt rejected by provider policy."

## 2. `Trusted Hosts` is operationally brittle for OpenAI

This is the clearest blocker. OpenAI’s requests arrived from multiple IPs within a short time window:

- `52.176.139.178`
- `52.176.139.180`
- `52.176.139.181`
- `52.176.139.183`
- `52.176.139.187`
- `52.176.139.190`

That makes IP-based trust impractical unless OpenAI publishes a stable allowlist suitable for Keycloak’s policy model. For this specific use case, `Trusted Hosts` is likely the wrong anonymous-registration policy to keep enabled.

## 3. `Allowed Client Scopes` is still misconfigured even after UI changes

This is the second blocker. The UI dropdown you saw only contained regular Keycloak client scopes like:

- `profile`
- `email`
- `roles`
- `web-origins`
- `basic`

It did **not** contain `openid`, which is exactly what the logs show Keycloak rejecting.

That aligns strongly with the Kane article: the fix may require the **Admin REST API**, not just the UI, to insert an explicit allowed scope list including `openid`. Right now the anonymous policy dump contains only:

```json
{
  "allow-default-scopes": ["true"]
}
```

and no visible `allowed-client-scopes` entry.

## 4. RFC 8707 / audience is a likely next issue, but not the current one

The Kane article is right to call out Keycloak’s lack of native RFC 8707 `resource` support. That could become the next compatibility issue after DCR is fixed.

In the current deployment, though, it is **not** the first failure:

- `smailnail-mcp` does not currently enforce audience because no `--oidc-audience` is configured
- the current flow is dying before token issuance, at anonymous DCR policy evaluation

So audience/resource should be treated as a **second-phase check**, not the main blocker today.

## Likely root cause stack

The likely failure chain is:

1. OpenAI discovers the protected resource and auth server correctly
2. OpenAI attempts anonymous DCR against Keycloak
3. Keycloak rejects registration because:
   - `Trusted Hosts` does not trust OpenAI’s source IP
   - `Allowed Client Scopes` still rejects `openid`
4. OpenAI never gets a client registration or token
5. OpenAI surfaces a generic RFC 7591 "doesn't support DCR" style error
6. `smailnail-mcp` only sees anonymous metadata and unauthenticated `/mcp` requests

## Recommended remediation order

## Option A: fastest debug path

Temporarily remove or disable these **anonymous** client registration policies:

- `Trusted Hosts`
- `Allowed Client Scopes`

Then retry OpenAI immediately.

Why:

- it proves whether Keycloak policy is the only blocker
- it avoids the OpenAI rotating-IP problem
- it avoids the `openid` UI/config mismatch

If OpenAI gets past DCR after that, reintroduce restrictions one at a time.

## Option B: keep DCR but configure it properly

If you want to keep anonymous DCR guarded:

1. Rework `Trusted Hosts`
   - likely disable it for OpenAI-based DCR
   - or replace it with a model that does not depend on unstable egress IPs

2. Update `Allowed Client Scopes` through the Admin REST API
   - explicitly include `openid`
   - likely also include any additional scopes OpenAI requests or that you want defaulted
   - if you follow the Kane pattern for MCP, include a dedicated custom scope such as `mcp:run`

This second step likely cannot be completed confidently through the UI alone.

## Option C: avoid anonymous DCR entirely

If OpenAI supports a pre-created client model for this connector flow, that would avoid the Keycloak anonymous DCR policy surface altogether. That may be more stable than trying to make Keycloak’s anonymous registration policies happy with a third-party platform using rotating IPs.

This option depends on OpenAI connector behavior and should be evaluated against current OpenAI connector docs.

## Immediate operator playbook

## 1. Watch the logs live while retrying OpenAI

Keycloak:

```bash
ssh root@89.167.52.236 \
  'docker logs -f keycloak-k12lm4blpo13louovn3pfsgs 2>&1 | grep -E "CLIENT_REGISTER|Trusted Hosts|Allowed Client Scopes|openid|error="'
```

Hosted MCP:

```bash
ssh root@89.167.52.236 \
  'docker logs -f fhp3mxqlfftdxdib3vxz89l3-044752750295 2>&1 | grep -E "Unauthorized|protected resource|/mcp|aiohttp|Claude|OpenAI|Python/"'
```

## 2. Confirm DCR is still blocked from outside OpenAI

```bash
curl -s -o /tmp/reg.out -w '%{http_code}\n' \
  -H 'Content-Type: application/json' \
  -d '{"client_name":"probe","redirect_uris":["https://example.com/cb"],"grant_types":["authorization_code"],"response_types":["code"],"token_endpoint_auth_method":"none","scope":"openid profile email"}' \
  https://auth.scapegoat.dev/realms/smailnail/clients-registrations/openid-connect

cat /tmp/reg.out
```

If this still returns a 4xx policy error, OpenAI has no chance of succeeding.

## 3. Use the non-DCR smoke as a control

The ticket script:

- `scripts/smoke_hosted_mcp_oidc.sh`

already proves the hosted MCP, token validation, and IMAP execution path work when a bearer token exists. That is important because it isolates the problem to **registration/onboarding**, not the MCP tool runtime.

## 4. After DCR is fixed, test for the next likely issue

If OpenAI gets past registration but still fails:

- inspect whether Keycloak-issued tokens include the right `aud`
- inspect whether OpenAI is sending `resource=` and expecting RFC 8707 semantics
- compare that to the Kane article’s audience-mapper workaround

## Suggested next change

For the next debug cycle, the highest-signal change is:

1. disable anonymous `Trusted Hosts`
2. disable anonymous `Allowed Client Scopes`
3. retry OpenAI immediately while tailing Keycloak and MCP logs

If that succeeds, reintroduce restrictions in this order:

1. restore a working allowed-scope model
2. decide whether `Trusted Hosts` is worth keeping at all for OpenAI
3. only then consider audience/resource hardening

## Bottom line

The hosted `smailnail-mcp` service is not missing RFC 7591 support. The auth server is advertising DCR correctly, and OpenAI is reaching it. The actual failure is Keycloak anonymous DCR policy rejection, currently split across:

- `Trusted Hosts`
- `Allowed Client Scopes`

Fix those first. Only after registration succeeds should attention move to Keycloak’s RFC 8707/audience limitations.
