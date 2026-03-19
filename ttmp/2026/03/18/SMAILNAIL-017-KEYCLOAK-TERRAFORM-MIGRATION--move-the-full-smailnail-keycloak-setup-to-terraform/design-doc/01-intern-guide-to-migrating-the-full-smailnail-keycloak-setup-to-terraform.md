---
Title: intern guide to migrating the full smailnail keycloak setup to terraform
Ticket: SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION
Status: active
Topics:
    - smailnail
    - keycloak
    - terraform
    - oidc
    - deployments
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh
      Note: Evidence that hosted Keycloak provisioning is still imperative today
    - Path: smailnail/deployments/terraform/keycloak
      Note: Initial Terraform scaffold landed from this ticket
    - Path: smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json
      Note: Current local Keycloak realm definition to be translated into Terraform
    - Path: smailnail/docker-compose.local.yml
      Note: Shows how local Keycloak is bootstrapped today via import-realm
    - Path: smailnail/docs/deployments/smailnaild-merged-coolify.md
      Note: Documents the current hosted merged deployment contract and MCP issuer/resource URLs
    - Path: smailnail/docs/deployments/smailnaild-oidc-keycloak.md
      Note: Documents the current browser OIDC client expectations
    - Path: smailnail/pkg/smailnaild/auth/config.go
      Note: Defines the application-side OIDC configuration boundary
    - Path: smailnail/scripts/docker-entrypoint.smailnaild.sh
      Note: Shows how runtime env vars map to OIDC and MCP flags today
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-18T17:14:27.604381918-04:00
WhatFor: Explain the full current smailnail Keycloak system, identify what must move into Terraform, and provide a concrete implementation plan for an intern to execute safely.
WhenToUse: Use when onboarding an engineer to the current smailnail identity stack, planning Terraform ownership boundaries, or implementing the migration from manual Keycloak setup to infrastructure as code.
---



# intern guide to migrating the full smailnail keycloak setup to terraform

## Executive Summary

This document explains how the `smailnail` identity system currently works and how to migrate the whole Keycloak setup to Terraform. The goal is not to automate one client in isolation. The goal is to make the Keycloak realm itself a reproducible, reviewable, version-controlled part of the deployment.

Today, the Keycloak configuration lives in multiple places at once:

- the local development realm is defined by a JSON import file in [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json#L1)
- the local Docker stack mounts that JSON into the Keycloak container in [docker-compose.local.yml](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml#L33)
- the hosted production setup has been created or adjusted with `kcadm.sh` and one-off scripts such as [create_keycloak_realm_and_mcp_client.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh#L1)
- the expected runtime contract is described narratively in deployment docs such as [smailnaild-oidc-keycloak.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-oidc-keycloak.md#L1) and [smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md#L1)

That split is workable for prototyping, but it is a poor long-term operating model. It makes drift likely, review difficult, and production debugging slower because the team does not have one authoritative definition of the realm.

The recommended migration is to introduce Terraform as the source of truth for:

- realm configuration
- browser-login clients
- MCP clients
- redirect URIs and web origins
- client scopes and protocol mappers
- selected client-registration policy decisions

The recommended implementation path is incremental:

1. inventory the current realm and document the exact resource boundary
2. create Terraform for the existing local and hosted shape
3. import or recreate the hosted realm safely
4. add validation and drift review
5. only then tighten or expand connector support such as Claude or OpenAI-specific clients

An initial implementation scaffold now exists in:

- [smailnail/deployments/terraform/keycloak](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak)

## Problem Statement

The current `smailnail` Keycloak setup is not managed from a single declarative system. Instead, different parts of the identity surface are encoded in different formats for different environments.

The local environment is bootstrapped by a realm import file. That file currently hardcodes:

- realm name
- base realm settings
- the `smailnail-web` client
- the `smailnail-mcp` client
- a seeded local test user

You can see that directly in [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json#L2), where the realm is `smailnail-dev`, and in [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json#L11), where the clients array begins.

The hosted environment, however, is not defined in that file. Instead, it has been created procedurally with admin CLI operations. The clearest example is [create_keycloak_realm_and_mcp_client.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh#L7), which writes a JSON payload to `/tmp`, copies it to the remote host, and then shells into the Keycloak container to create the realm and client.

At the same time, the applications consuming Keycloak configuration are reading their expected values from runtime flags and environment variables. Browser OIDC for `smailnaild` is defined by the Glazed auth section in [config.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/config.go#L12), and the production environment contract for the merged hosted deployment is described in [smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md#L69).

This causes four concrete problems:

1. Drift:
   The local realm import, hosted Keycloak state, and deployment docs can diverge silently.
2. Reviewability:
   Infrastructure changes are not always visible in a normal code review.
3. Reproducibility:
   Recreating the same hosted realm requires memory, shell history, and ticket archaeology.
4. Operational risk:
   Connector failures become harder to debug because it is not obvious which Keycloak settings are authoritative.

## Scope

This document covers:

- how the current local and hosted Keycloak setup works
- which pieces should move to Terraform
- how Terraform should be structured
- how to migrate safely without breaking browser login or MCP auth
- how to think about OpenAI and Claude-specific client behavior during the migration

This document does not cover:

- a full primer on Terraform syntax
- non-Keycloak Coolify resources such as the `smailnaild` app itself
- a final code implementation of the Terraform modules in this ticket
- a guarantee that every Keycloak registration policy surface is already exposed cleanly by the provider

## Goals

- Make the Keycloak realm declarative and reviewable.
- Remove dependence on manual admin-console edits and ad-hoc `kcadm.sh` commands.
- Keep local development and hosted production aligned at the conceptual level.
- Preserve the current runtime contracts for browser login and hosted MCP auth.
- Make connector compatibility work easier to reason about later.

## Non-Goals

- Managing end-user identities in production with Terraform.
- Using Terraform to rotate every secret automatically on day one.
- Redesigning the `smailnail` app's identity model.
- Making anonymous DCR looser than necessary for convenience.

## Current-State Architecture

This section explains what exists today. It matters because the Terraform migration should encode the current intended system, not an imagined replacement.

### High-level system picture

```text
                   +-------------------------------+
                   |         Coolify / HTTPS       |
                   | public hosts and healthchecks |
                   +---------------+---------------+
                                   |
                                   v
                  +------------------------------------+
                  |           smailnaild               |
                  |  /auth/*   /api/*   /mcp   /readyz |
                  +----------------+-------------------+
                                   |
            +----------------------+----------------------+
            |                                             |
            v                                             v
+----------------------------+            +------------------------------+
| browser OIDC login         |            | MCP bearer token validation  |
| client_id = smailnail-web  |            | issuer = Keycloak realm      |
+----------------------------+            +------------------------------+
            \                                             /
             \                                           /
              \                                         /
               v                                       v
               +---------------------------------------+
               |        Keycloak realm: smailnail      |
               | users, clients, scopes, DCR policies  |
               +---------------------------------------+
```

### Local development Keycloak

The local Keycloak stack is started by [docker-compose.local.yml](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml#L17), which defines:

- a Postgres container for Keycloak
- a Keycloak container
- `start-dev --import-realm`
- a bind mount from `./dev/keycloak/realm-import` into `/opt/keycloak/data/import`

That means local Keycloak state is effectively bootstrapped from version-controlled JSON. This is good in spirit because it is declarative. The problem is that JSON import is a one-shot bootstrap format, not a nice long-term infrastructure-management format. It is difficult to diff semantically and awkward to evolve in a modular way.

### What the local realm currently contains

The local realm import in [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json#L1) includes:

- realm settings:
  - `realm = smailnail-dev`
  - `enabled = true`
  - login and password-reset toggles
- the `smailnail-web` client:
  - confidential
  - fixed secret `smailnail-web-secret`
  - loopback redirect URIs and origins
- the `smailnail-mcp` client:
  - public client
  - standard flow enabled
  - loopback redirect URIs
- a seeded local test user:
  - `alice`
  - password `secret`

That single file currently mixes several different concerns:

- realm infrastructure
- application client definitions
- developer fixture users
- local-only callback URLs

Terraform should separate those concerns much more clearly.

### Hosted Keycloak expectations from the application side

The hosted browser-login path expects a confidential client with a client secret. [smailnaild-oidc-keycloak.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-oidc-keycloak.md#L26) recommends:

- realm `smailnail`
- client ID `smailnail-web`
- confidential client
- standard flow enabled
- direct access grants disabled
- service accounts disabled
- scopes `openid`, `profile`, `email`

The hosted merged deployment adds a second auth surface for MCP. [smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md#L80) documents:

- `SMAILNAILD_OIDC_CLIENT_ID=smailnail-web`
- `SMAILNAILD_MCP_AUTH_MODE=external_oidc`
- `SMAILNAILD_MCP_AUTH_RESOURCE_URL=https://smailnail.scapegoat.dev/mcp`
- `SMAILNAILD_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`

That means one Keycloak realm is already serving two distinct application consumers:

- browser login for the web app
- bearer-token validation expectations for `/mcp`

### How the app consumes this config

Browser OIDC settings are defined structurally in [config.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/config.go#L20), which includes:

- issuer URL
- client ID
- client secret
- redirect URL
- scopes

The Docker entrypoint translates environment variables into those flags in [docker-entrypoint.smailnaild.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnaild.sh#L43) and [docker-entrypoint.smailnaild.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnaild.sh#L81). That means the application contract is already stable enough to serve as a consumer boundary for Terraform-managed Keycloak resources.

### How hosted Keycloak is currently provisioned

The strongest evidence of current hosted provisioning is [create_keycloak_realm_and_mcp_client.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh#L1). It:

1. writes a client JSON file locally
2. copies it to the remote host
3. runs `kcadm.sh config credentials`
4. creates the `smailnail` realm if needed
5. creates the `smailnail-mcp` client if needed

This script is useful as evidence because it shows which configuration is actually important in practice. But it is also evidence of the problem:

- imperative
- not strongly typed
- not centrally composed
- hard to review
- hard to compare between local and hosted

## Gap Analysis

The target state is a single version-controlled definition of the Keycloak setup. The current state falls short in several specific ways.

### Gap 1: No single source of truth

Current truth is split across:

- realm import JSON
- shell scripts
- manual hosted state
- deployment docs

Terraform should become the single declared authority for realm configuration.

### Gap 2: No clear environment layering

Local and hosted environments differ today, but the differences are not expressed cleanly. Instead they are encoded by:

- separate realm names
- different redirect URIs
- different provisioning methods

Terraform should express:

- shared base configuration
- local-only overrides
- hosted-only overrides

### Gap 3: Connector policy is not modeled clearly

OpenAI and Claude introduce real policy questions:

- do we use pre-provisioned clients or DCR
- which redirect URIs are allowed
- which scopes are allowed
- how strict should anonymous client-registration policy be

Those are identity-policy choices. They should not remain hidden in production container state.

### Gap 4: Local test users are mixed with infrastructure

The `alice` user in the realm import is fine as a local fixture, but it should not be confused with infrastructure ownership. Terraform can model local fixture users if the team wants reproducible local smoke tests, but those users should be clearly classified as local-only test resources.

## Proposed Solution

The proposed solution is to manage Keycloak for `smailnail` using Terraform with a modular structure that separates:

- shared realm configuration
- environment-specific values
- local-only fixture users
- product-specific MCP clients
- DCR policy decisions

### Proposed repository layout

This is the recommended layout inside the `smailnail` repo:

```text
smailnail/
  deployments/
    terraform/
      keycloak/
        versions.tf
        providers.tf
        variables.tf
        outputs.tf
        modules/
          realm-base/
          browser-client/
          mcp-client/
          local-fixtures/
          registration-policy/
        envs/
          local/
            main.tf
            terraform.tfvars.example
          hosted/
            main.tf
            terraform.tfvars.example
```

Why this layout is recommended:

- it matches the repository's existing `deployments/` naming
- it keeps Keycloak deployment code near other deployment artifacts
- it allows shared modules with thin environment wrappers
- it makes it obvious where future CI and playbooks should live

### Proposed ownership boundary

Terraform should manage these resources:

- realm settings
- browser-login client definitions
- MCP client definitions
- client scopes
- protocol mappers
- redirect URIs
- web origins
- client-registration policy components if the provider supports them cleanly enough

Terraform should probably not manage these production resources in the first iteration:

- real human users
- production user passwords
- short-lived incident-response toggles

Terraform may manage these local-only resources:

- test user `alice`
- fixed local-only client secrets

### Proposed environment model

The environment model should be:

```text
base module
  -> shared realm name conventions
  -> shared scopes
  -> shared client defaults

local environment
  -> realm = smailnail-dev
  -> loopback redirects
  -> seeded local test user
  -> relaxed settings for local smoke tests where intentional

hosted environment
  -> realm = smailnail
  -> public HTTPS redirects
  -> production secrets injected from CI or operator environment
  -> stricter DCR / callback / origin policy
```

That approach preserves conceptual sameness while still allowing the environments to differ where they genuinely need to.

## Detailed Resource Inventory

This section lists what needs to be translated from the current system into Terraform.

### Realm-level settings

From the local import and hosted docs, the realm model includes:

- name
- enabled flag
- email/login policy
- duplicate email policy
- reset-password policy
- SSL requirement

These currently live in [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json#L2).

### `smailnail-web` client

This client exists to support browser login for `smailnaild`.

Important characteristics from the current docs:

- confidential
- standard flow enabled
- direct grants disabled
- service accounts disabled
- redirects to `/auth/callback`
- claims such as `email` and `preferred_username` are used only as profile metadata

See [smailnaild-oidc-keycloak.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-oidc-keycloak.md#L26).

### `smailnail-mcp` client

This client exists to support MCP-related auth experiments and hosted bearer-token flows.

Important characteristics from the local import and deployment scripts:

- public client today
- standard flow enabled
- service accounts disabled
- redirect URIs for loopback locally
- redirect URIs for Claude callbacks and hosted MCP domains in hosted provisioning scripts

See [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json#L39) and [create_keycloak_realm_and_mcp_client.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh#L7).

### Connector-specific clients

The migration should plan for dedicated external product clients rather than treating every connector as the same thing forever.

Recommended likely future clients:

- `smailnail-openai-mcp`
- `smailnail-claude-mcp`

This is not because the products are special in principle. It is because their callback URLs and DCR behavior differ enough to justify separation.

### Client scopes and protocol mappers

At minimum, the browser-login path depends on:

- `openid`
- `profile`
- `email`

The broader identity guidance in `smailnail` has also already documented useful claims such as:

- `email`
- `email_verified`
- `preferred_username`
- `name`
- `picture`

These are referenced in [smailnaild-oidc-keycloak.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-oidc-keycloak.md#L53).

### Registration policy surfaces

The current live troubleshooting for OpenAI and Claude has shown that client-registration policy is a first-class operational concern, not an edge case. The Terraform design therefore needs an explicit policy stance:

- either manage anonymous DCR policy declaratively
- or intentionally avoid anonymous DCR and use pre-provisioned clients

The right design default is to prefer pre-provisioned clients unless DCR is truly required by the product integration.

## Design Decisions

### Decision 1: Terraform should own the whole realm, not only the Claude client

Rationale:

- managing one client without the realm around it still leaves drift everywhere else
- redirects, scopes, and DCR policies are realm-level concerns
- the browser client and MCP client are already linked conceptually

### Decision 2: Local and hosted should share modules, not share one variable file

Rationale:

- the environments are similar but not identical
- local test users and loopback redirects should not leak into production
- hosted secrets should not be hardcoded

### Decision 3: Prefer pre-provisioned connector clients first

Rationale:

- easier to audit
- fewer hidden DCR policy failures
- clearer ownership over redirect URIs and scopes

### Decision 4: Keep production users out of Terraform in the first pass

Rationale:

- identity infrastructure and user lifecycle are different domains
- managing production user data as Terraform resources is usually the wrong abstraction here

### Decision 5: Keep local test fixtures explicit and optional

Rationale:

- local smoke tests still need stable credentials
- but those fixtures should be clearly marked as local-only

## Proposed Terraform Flow

### First implementation slice

The first slice should aim for parity, not invention.

Pseudocode:

```text
module "realm_base" {
  realm_name = var.realm_name
  login_policies = ...
}

module "browser_client" {
  realm_name = module.realm_base.name
  client_id = "smailnail-web"
  confidential = true
  redirect_uris = var.web_redirect_uris
  web_origins = var.web_origins
  scopes = ["openid", "profile", "email"]
}

module "mcp_client" {
  realm_name = module.realm_base.name
  client_id = "smailnail-mcp"
  public = true
  redirect_uris = var.mcp_redirect_uris
  web_origins = var.mcp_web_origins
}

module "local_fixtures" {
  count = var.enable_local_fixtures ? 1 : 0
  create_test_user = true
  test_user_username = "alice"
}
```

The job of the first slice is:

- recreate current known-good intent
- make `terraform plan` meaningful
- avoid mixing migration work with auth redesign

### Second implementation slice

After parity is reached:

- add dedicated connector clients if needed
- decide whether anonymous DCR remains necessary
- tighten audience and scope policy
- add import and drift workflows for hosted

## Migration Strategy

### Strategy options

There are two realistic strategies for hosted Keycloak:

1. Import existing hosted resources into Terraform state.
2. Recreate selected resources under Terraform control.

### Recommended initial approach

Prefer importing the current hosted realm and clients if the live realm is already serving production traffic.

Why:

- lower blast radius
- easier rollback
- avoids cutting over both auth behavior and management model at once

### Safe migration sequence

```text
1. Snapshot current live realm settings
2. Encode intended configuration in Terraform
3. Run terraform plan against a local or staging Keycloak first
4. Import live realm and client resources
5. Run plan against hosted realm in read-only review mode
6. Resolve drift intentionally
7. Apply only small, reviewed changes
8. Re-run browser and MCP verification
```

### What must be verified after each change

- browser login still reaches Keycloak
- `/auth/callback` still works
- `/.well-known/oauth-protected-resource` still advertises the right resource
- `/mcp` still returns the expected `401` challenge when anonymous
- a known-good token still works
- connector-specific callback URIs remain allowed

## Testing and Validation Strategy

### Local validation

Use the existing local Docker stack from [docker-compose.local.yml](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml#L1) as the fast feedback loop.

Suggested workflow:

```text
terraform fmt
terraform validate
terraform plan for local
start local keycloak
apply local
run smailnail browser login
run local MCP smoke
```

### Hosted validation

Use the hosted deployment checks already documented in [smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md#L179):

1. `GET /readyz`
2. `GET /.well-known/oauth-protected-resource`
3. anonymous `POST /mcp` returns `401`
4. browser login through `/auth/login`
5. `GET /api/me`
6. obtain token and call `/mcp`

### Connector validation

OpenAI and Claude should be validated explicitly and separately.

Why:

- they may use different callback URLs
- they may differ on DCR behavior
- they may ask for different scopes

## Risks

### Risk 1: Terraform provider coverage gaps

Some Keycloak surfaces, especially client-registration policy details, may not map cleanly onto Terraform resources. If that happens, the team must decide whether to:

- supplement Terraform with scripted admin API calls
- or reduce reliance on that Keycloak feature, for example by preferring pre-provisioned clients

### Risk 2: Importing bad live state

If the live realm already contains accidental drift, importing it blindly can fossilize mistakes. The import process should include a deliberate review of:

- redirect URIs
- web origins
- enabled grant types
- DCR-related settings

### Risk 3: Secret handling

The browser client currently uses a confidential secret. Terraform must not hardcode production secrets in plain repo files. The plan should use:

- CI-injected sensitive variables
- operator-managed secret input
- or state protections appropriate for sensitive data

### Risk 4: Local and hosted policy divergence

If local is too lax and hosted is too strict, engineers may repeatedly "pass local, fail prod." The Terraform design should therefore share defaults where it can and isolate the truly environment-specific differences clearly.

## Alternatives Considered

### Alternative 1: Keep using realm import JSON

Rejected because:

- poor modularity
- awkward diffing
- not a good fit for hosted lifecycle management
- weak story for incremental updates

### Alternative 2: Keep using `kcadm.sh` scripts

Rejected as the primary long-term model because:

- imperative rather than declarative
- easy to drift
- harder to review
- poor reuse across environments

### Alternative 3: Terraform only the Claude client

Rejected because:

- it solves the immediate symptom but not the underlying infrastructure-management problem
- browser and MCP clients are still related system state
- DCR and redirect policy remain unmanaged

## Implementation Plan

### Phase A: Establish Terraform foundation

- choose the provider and version to standardize on
- create the `deployments/terraform/keycloak` layout
- decide backend and state handling for hosted
- define provider authentication for local and hosted Keycloak

### Phase B: Recreate the current realm declaratively

- model realm settings
- model `smailnail-web`
- model `smailnail-mcp`
- model existing scopes and mappers
- model local fixture user if keeping that path

### Phase C: Validate against local Keycloak

- destroy and recreate local Keycloak realm from Terraform
- prove browser login still works
- prove local MCP token path still works
- compare outputs against existing docs and tests

### Phase D: Import hosted realm and review drift

- capture live realm state
- import into Terraform state
- run plan without applying
- resolve differences deliberately

### Phase E: Add connector policy decisions

- decide DCR vs pre-provisioned clients
- if needed, add dedicated OpenAI and Claude clients
- tighten callback URLs and scopes

### Phase F: Update docs and handoff

- update deployment docs to point at Terraform
- add operator playbooks for plan/apply/import
- make the old shell scripts clearly deprecated or archive-only

## Open Questions

- Which exact Keycloak registration-policy surfaces are cleanly supported by the chosen provider version?
- Should local fixture users remain Terraform-managed, or should they move to a separate test bootstrap step?
- Should `smailnail-mcp` remain a shared public client, or should production move immediately to per-product clients?
- What should the long-term production stance on anonymous DCR be?

## Official API and Documentation References

- Keycloak client registration service: https://www.keycloak.org/securing-apps/client-registration
- Keycloak server administration guide: https://www.keycloak.org/docs/latest/server_admin/
- Keycloak bootstrapping and admin recovery: https://www.keycloak.org/server/bootstrap-admin-recovery
- Keycloak server configuration guide: https://www.keycloak.org/server/configuration
- Keycloak Terraform provider repository: https://github.com/keycloak/terraform-provider-keycloak

## References

- [smailnail/deployments/terraform/keycloak](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak)
- [smailnail-dev-realm.json](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json)
- [docker-compose.local.yml](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml)
- [smailnaild-oidc-keycloak.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-oidc-keycloak.md)
- [shared-oidc-playbook.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/shared-oidc-playbook.md)
- [smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md)
- [docker-entrypoint.smailnaild.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnaild.sh)
- [config.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/config.go)
- [create_keycloak_realm_and_mcp_client.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh)

## Problem Statement

<!-- Describe the problem this design addresses -->

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
