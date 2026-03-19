# Changelog

## 2026-03-18

- Initial workspace created
- Added a detailed intern-facing design and implementation guide for migrating the full `smailnail` Keycloak setup to Terraform.
- Added a structured task list covering provider choice, realm modeling, connector policy, migration mechanics, verification, and documentation handoff.
- Added a diary entry capturing the evidence gathered from the current local realm import, deployment docs, and hosted `kcadm.sh` setup.
- Added an initial Terraform scaffold under `smailnail/deployments/terraform/keycloak` with shared modules plus `local` and `hosted` environments.
- Verified the scaffold with the real `keycloak/keycloak` provider using `terraform init -backend=false` and `terraform validate`.
- Proved a real local apply against sandbox realm `smailnail-dev-tf`, verified OIDC discovery for that realm, and reached a clean no-op follow-up plan.
- Added hosted Terraform auth support via bootstrap admin username/password and imported the live hosted realm plus the `smailnail-web` and `smailnail-mcp` clients into hosted Terraform state.
- Confirmed two important provider behaviors during hosted import: `keycloak_openid_client` import requires the internal Keycloak client UUID in the import ID, and `keycloak_openid_client_default_scopes` / `optional_scopes` do not support import.
- Aligned hosted Terraform with the decision to keep `smailnail.mcp.scapegoat.dev` as the canonical public hostname, which removed the old-vs-new hostname drift from the hosted plan.
- Reconciled the remaining hosted Terraform drift by preserving live production behavior for `use_refresh_tokens`, `RS256`, and the unmanaged scope-attachment surfaces, and by dropping the optional hosted browser client display name so the imported live state now plans cleanly with `No changes`.
- Pushed the Terraform and documentation branches after pre-push validation passed in both repositories.
- Added a hosted operator playbook covering safe import, no-op drift review, deliberate apply rules, rollback guidance, and OpenAI/Claude verification checks.
- Added a Terraform-managed hosted remediation step for the anonymous Keycloak DCR allowed-scope policy, applied it to production, and restored a clean hosted no-op plan afterward.

## 2026-03-18

Document the current smailnail Keycloak system and propose a Terraform migration plan that covers the realm, browser client, MCP client, environment separation, and connector policy.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/18/SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION--move-the-full-smailnail-keycloak-setup-to-terraform/design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md — Primary design document
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/18/SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION--move-the-full-smailnail-keycloak-setup-to-terraform/reference/01-diary.md — Chronological investigation record
