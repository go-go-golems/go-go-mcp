# Changelog

## 2026-03-18

- Initial workspace created
- Added a detailed intern-facing design and implementation guide for migrating the full `smailnail` Keycloak setup to Terraform.
- Added a structured task list covering provider choice, realm modeling, connector policy, migration mechanics, verification, and documentation handoff.
- Added a diary entry capturing the evidence gathered from the current local realm import, deployment docs, and hosted `kcadm.sh` setup.
- Added an initial Terraform scaffold under `smailnail/deployments/terraform/keycloak` with shared modules plus `local` and `hosted` environments.
- Verified the scaffold with the real `keycloak/keycloak` provider using `terraform init -backend=false` and `terraform validate`.
- Proved a real local apply against sandbox realm `smailnail-dev-tf`, verified OIDC discovery for that realm, and reached a clean no-op follow-up plan.

## 2026-03-18

Document the current smailnail Keycloak system and propose a Terraform migration plan that covers the realm, browser client, MCP client, environment separation, and connector policy.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/18/SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION--move-the-full-smailnail-keycloak-setup-to-terraform/design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md — Primary design document
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/go-go-mcp/ttmp/2026/03/18/SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION--move-the-full-smailnail-keycloak-setup-to-terraform/reference/01-diary.md — Chronological investigation record
