# Tasks

## Phase 0: Discovery and provider choice

- [x] Capture the current Keycloak setup spread across realm-import JSON, deployment docs, and `kcadm.sh` scripts.
- [x] Document the desired migration scope as "the full smailnail Keycloak setup", not only the Claude MCP client.
- [x] Confirm the exact Terraform provider and minimum version to standardize on for the repo.
- [ ] Confirm which Keycloak policy surfaces are first-class Terraform resources versus surfaces that may still need supplemental automation.

## Phase 1: Define the target Terraform boundary

- [x] Decide where Terraform should live in the repo.
- [x] Define environment boundaries for `local` and `hosted`.
- [ ] Decide which Keycloak resources are owned by Terraform and which must remain out of scope.
- [ ] Define how sensitive values such as confidential client secrets are injected.
- [ ] Define how Terraform state will be stored and locked for the hosted environment.

## Phase 2: Model the current smailnail realm declaratively

- [x] Create Terraform for the realm-level settings currently encoded in `smailnail-dev-realm.json`.
- [x] Create Terraform for the `smailnail-web` browser-login client.
- [x] Create Terraform for the `smailnail-mcp` MCP client.
- [ ] Create Terraform for explicit client scopes and protocol mappers needed by the current application behavior.
- [ ] Represent callback URLs and web origins for local and hosted environments.
- [ ] Represent the current browser-login claim contract used by `smailnaild`.

## Phase 3: Model hosted MCP connector policy

- [ ] Decide whether external MCP products use pre-provisioned clients, DCR, or a hybrid model.
- [ ] If DCR remains required, model the Keycloak client-registration policy surface declaratively or document the exact gap.
- [ ] Add a dedicated client shape for Claude if pre-provisioned-client mode is chosen.
- [ ] Add a dedicated client shape for OpenAI if pre-provisioned-client mode is chosen.
- [ ] Define the production default for audience and scope enforcement once connector behavior is stable.

## Phase 4: Migration mechanics

- [ ] Decide whether to import existing hosted Keycloak resources into Terraform state or recreate them.
- [ ] Write an operator playbook for safe first import and first plan against the live realm.
- [ ] Add a drift-review workflow before any apply to production.
- [ ] Create rollback guidance for accidental realm or client misconfiguration.

## Phase 5: Verification and CI

- [x] Add `terraform fmt` and `terraform validate` to the implementation workflow.
- [x] Add a local verification workflow against the local Keycloak container.
- [ ] Add a hosted verification workflow that checks discovery, browser login, and MCP metadata.
- [ ] Add a connector-focused verification checklist for OpenAI and Claude.

## Phase 6: Documentation and handover

- [x] Create the intern-facing analysis and implementation guide.
- [x] Create the ticket diary.
- [ ] Add a follow-up operator playbook once the Terraform layout is implemented.
- [ ] Update the human-facing deployment docs to point at Terraform as the source of truth.
