---
Title: Hosted Keycloak import and apply playbook
Ticket: SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION
Status: active
Topics:
    - smailnail
    - keycloak
    - terraform
    - oidc
    - deployments
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../smailnail/deployments/terraform/keycloak/README.md
      Note: Terraform operator entrypoint referenced by the playbook
    - Path: ../../../../../../../smailnail/deployments/terraform/keycloak/envs/hosted/main.tf
      Note: Hosted Keycloak realm and client settings managed by the playbook
    - Path: ../../../../../../../smailnail/deployments/terraform/keycloak/envs/hosted/terraform.tfvars.example
      Note: Example hosted Terraform inputs referenced by the playbook
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-18T21:37:46.138214157-04:00
WhatFor: Safely import the live hosted smailnail Keycloak realm into Terraform, inspect drift, and apply only deliberate reviewed changes.
WhenToUse: Use when onboarding a new operator to the hosted Terraform workflow, repeating the first import on another machine, reviewing production drift, or preparing a controlled hosted apply.
---


# Hosted Keycloak import and apply playbook

## Goal

Provide a repeatable operator procedure for working with the hosted `smailnail`
Keycloak realm through Terraform without accidentally changing production auth
behavior.

The immediate goal is not "apply changes quickly." The immediate goal is to
reach a trustworthy, reviewable, low-risk Terraform workflow for the hosted
realm.

## Context

The `smailnail` hosted identity setup lives in Keycloak at
`https://auth.scapegoat.dev` and is consumed by:

- the browser-login client `smailnail-web`
- the MCP client `smailnail-mcp`
- future external MCP connectors such as OpenAI and Claude

The Terraform root for hosted Keycloak is:

- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted`

The current hosted configuration is intentionally conservative. It preserves
live production behavior instead of trying to "clean up" the realm during first
adoption. In particular, it currently encodes:

- hosted public MCP hostname `smailnail.mcp.scapegoat.dev`
- `use_refresh_tokens = false`
- `default_signature_algorithm = "RS256"`
- `manage_scope_attachments = false` because the relevant scope-attachment
  helper resources cannot be imported cleanly with the current provider

The important provider limitation discovered during adoption is:

- `keycloak_openid_client` import requires `realm/internal-client-uuid`
- `keycloak_openid_client_default_scopes` and
  `keycloak_openid_client_optional_scopes` do not support import

That means the safe workflow is:

1. import the realm and importable clients
2. reconcile HCL until the plan is a no-op
3. only then consider a deliberate hosted apply

### System sketch

```text
Browser / MCP Client
        |
        v
  smailnail.mcp.scapegoat.dev
        |
        | external_oidc issuer
        v
  auth.scapegoat.dev (Keycloak)
        |
        +-- realm: smailnail
        +-- client: smailnail-web
        +-- client: smailnail-mcp
        +-- registration / scope / realm policy state
```

## Quick Reference

### Preconditions

- You are on a clean enough checkout that you can distinguish Terraform changes
  from unrelated local files.
- You have shell access to the hosted Keycloak container or another approved
  way to fetch admin credentials.
- You understand that this environment currently uses local Terraform state
  files, not a remote backend.
- You will not run `terraform apply` against hosted until `terraform plan`
  shows only changes you explicitly intend.

### Files that matter

- Terraform root:
  `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted`
- Hosted config:
  `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted/main.tf`
- Example variables:
  `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted/terraform.tfvars.example`
- Ticket design guide:
  `../design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md`
- Ticket diary:
  `./01-diary.md`

### Hosted resources currently managed

```text
module.realm.keycloak_realm.this
module.browser_client.keycloak_openid_client.this
module.mcp_client.keycloak_openid_client.this
```

### Values you need before import

You need:

- Keycloak base URL
- Keycloak admin realm
- bootstrap admin username/password or another admin credential pair
- the `smailnail-web` client secret
- the internal Keycloak UUIDs for the hosted clients you are importing

Do not hardcode secrets into committed files. Put them in:

- temporary shell environment variables, or
- an untracked `terraform.tfvars`

### Recommended environment variables

```bash
export TF_DIR=/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted
export KC_URL=https://auth.scapegoat.dev
export KC_ADMIN_REALM=master
export KC_CLIENT_ID=admin-cli
export KC_USERNAME='replace-with-bootstrap-admin-username'
export KC_PASSWORD='replace-with-bootstrap-admin-password'
export WEB_CLIENT_SECRET='replace-with-smailnail-web-secret'
```

### Create an untracked `terraform.tfvars`

From `"$TF_DIR"`:

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit it to contain:

```hcl
keycloak_url           = "https://auth.scapegoat.dev"
keycloak_admin_realm   = "master"
keycloak_client_id     = "admin-cli"
keycloak_client_secret = null
keycloak_username      = "replace-with-bootstrap-admin-username"
keycloak_password      = "replace-with-bootstrap-admin-password"
web_client_secret      = "replace-with-smailnail-web-secret"
mcp_client_secret      = null
```

Make sure `terraform.tfvars` is not committed.

### Initialize and validate

```bash
cd "$TF_DIR"
terraform init -backend=false
terraform validate
```

### Discover hosted client UUIDs with `kcadm.sh`

`kcadm.sh` is the Keycloak admin CLI. It is a shell wrapper around the
Keycloak admin REST API. It is useful here because Terraform import for
`keycloak_openid_client` requires the internal Keycloak UUID, not the public
`clientId`.

Example pattern from inside the Keycloak container:

```bash
/opt/keycloak/bin/kcadm.sh config credentials \
  --server http://127.0.0.1:8080 \
  --realm master \
  --user "$KC_BOOTSTRAP_ADMIN_USERNAME" \
  --password "$KC_BOOTSTRAP_ADMIN_PASSWORD"

/opt/keycloak/bin/kcadm.sh get clients -r smailnail -q clientId=smailnail-web
/opt/keycloak/bin/kcadm.sh get clients -r smailnail -q clientId=smailnail-mcp
```

What you are looking for is the internal `id` field, for example:

```json
[
  {
    "id": "internal-uuid-here",
    "clientId": "smailnail-web"
  }
]
```

### Import commands

Run imports from `"$TF_DIR"` after the UUIDs are known:

```bash
terraform import module.realm.keycloak_realm.this smailnail

terraform import \
  module.browser_client.keycloak_openid_client.this \
  smailnail/replace-with-smailnail-web-internal-uuid

terraform import \
  module.mcp_client.keycloak_openid_client.this \
  smailnail/replace-with-smailnail-mcp-internal-uuid
```

### Confirm imported state

```bash
terraform state list
```

Expected shape:

```text
module.browser_client.keycloak_openid_client.this
module.mcp_client.keycloak_openid_client.this
module.realm.keycloak_realm.this
```

### First hosted drift review

Run:

```bash
terraform plan -input=false
```

Interpretation rules:

- If the plan is `No changes. Your infrastructure matches the configuration.`,
  Terraform has adopted the current hosted baseline cleanly.
- If the plan includes hostname changes, redirect URI changes, token-behavior
  changes, or realm-default changes, do not apply. Update the HCL or revisit
  the product decision first.
- If the plan wants to create scope-attachment helper resources, remember that
  these resources are currently non-importable. Do not let first apply take
  those over accidentally.

### Safe hosted apply rule

Only run hosted `apply` when all three conditions are true:

1. `terraform validate` passes
2. `terraform plan -input=false` shows either no changes or a small deliberate
   change set
3. the change set has been reviewed as an auth behavior decision, not merely a
   Terraform normalization

Command:

```bash
terraform apply -input=false
```

### Post-apply verification checklist

After any hosted apply, verify:

```bash
curl -fsS https://auth.scapegoat.dev/realms/smailnail/.well-known/openid-configuration | jq -r '.issuer'
curl -fsS https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource | jq .
curl -i https://smailnail.mcp.scapegoat.dev/mcp
```

Expected outcomes:

- issuer remains `https://auth.scapegoat.dev/realms/smailnail`
- protected-resource metadata still returns the hosted MCP resource URL
- anonymous `/mcp` still returns `401` with the expected Bearer challenge

### Connector verification checklist

For OpenAI:

- confirm the connector can complete OAuth or use the configured credential path
- confirm it reaches `/mcp` with `Authorization: Bearer ...`
- confirm tool listing still works

For Claude:

- confirm it fetches `/.well-known/oauth-protected-resource`
- confirm it reaches the Keycloak issuer / authorize flow or the configured
  pre-provisioned client path
- confirm it eventually retries `/mcp` with `Authorization: Bearer ...`
- if it stops during DCR, inspect Keycloak events for
  `CLIENT_REGISTER_ERROR`

### Rollback rules

If you only imported state and did not apply:

- delete the local hosted state files and re-initialize
- do not change production manually just to match a bad local state file

Typical local cleanup:

```bash
rm -f terraform.tfstate terraform.tfstate.backup
rm -rf .terraform .terraform.lock.hcl
git checkout -- .terraform.lock.hcl
```

Use the last line only if the lock file was already committed and changed
locally. Do not discard intentional HCL edits.

If you already applied an unwanted change:

1. stop making additional changes
2. inspect the exact Terraform diff that was applied
3. restore the Keycloak setting through either:
   - a corrective Terraform change reviewed before apply, or
   - a temporary admin-console / `kcadm.sh` repair if the system is impaired
4. re-run `terraform plan` until the configuration and live state converge

### Known pitfalls

- Import IDs for clients are not `clientId`; they are `realm/internal-uuid`.
- Scope attachment resources are not currently importable.
- The first production apply should not be used to change refresh-token
  behavior or hostname policy.
- Terraform adoption is not the same as Terraform ownership. Start with
  behavior-preserving import, then tighten ownership gradually.

## Usage Examples

### Example 1: Repeat the hosted import on a fresh machine

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted
cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars
terraform init -backend=false
terraform validate
terraform import module.realm.keycloak_realm.this smailnail
terraform import module.browser_client.keycloak_openid_client.this smailnail/replace-with-web-uuid
terraform import module.mcp_client.keycloak_openid_client.this smailnail/replace-with-mcp-uuid
terraform state list
terraform plan -input=false
```

The expected endpoint of this example is a no-op plan, not an immediate apply.

### Example 2: Routine production drift review

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted
terraform validate
terraform plan -input=false
```

If the plan is non-empty, classify each diff before acting:

- intended product change
- live drift that should be preserved
- live drift that should be corrected
- provider artifact / non-importable surface

### Example 3: Deliberate hosted client change

Suppose the team decides to add a new redirect URI to `smailnail-mcp`.

Procedure:

1. edit `envs/hosted/main.tf`
2. run `terraform validate`
3. run `terraform plan -input=false`
4. review the diff with an auth mindset
5. apply only after the redirect change is explicitly approved
6. verify discovery, metadata, and connector login behavior

### Pseudocode for operator judgment

```text
if terraform_validate_fails:
    stop_and_fix_hcl()

plan = terraform_plan()

if plan.is_noop():
    record_success()
elif plan.contains_unreviewed_auth_behavior_change():
    stop_and_reconcile_hcl()
elif plan.contains_only_intended_reviewed_changes():
    apply_and_verify()
else:
    inspect_provider_limits_and_state_model()
```

## Related

- Main design guide:
  [../design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md](../design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md)
- Diary:
  [./01-diary.md](./01-diary.md)
- Ticket index:
  [../index.md](../index.md)
- Hosted Terraform root:
  `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted`
