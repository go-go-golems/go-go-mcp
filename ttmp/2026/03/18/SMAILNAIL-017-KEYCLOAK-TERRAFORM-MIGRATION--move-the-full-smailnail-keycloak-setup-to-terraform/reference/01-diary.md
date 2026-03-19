---
Title: Diary
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
    - Path: go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh
      Note: Primary evidence for imperative hosted Keycloak setup captured in diary
    - Path: smailnail/deployments/terraform/keycloak
      Note: Implementation scaffold recorded in diary step 2
    - Path: smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json
      Note: Primary evidence for current local realm contents captured in diary
    - Path: smailnail/docs/deployments/smailnaild-merged-coolify.md
      Note: Primary evidence for current hosted OIDC and MCP contract captured in diary
    - Path: smailnail/docs/deployments/smailnaild-oidc-keycloak.md
      Note: Primary evidence for current browser OIDC expectations captured in diary
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-18T17:14:27.604947896-04:00
WhatFor: Record the investigation and documentation work for planning the migration of the full smailnail Keycloak setup to Terraform.
WhenToUse: Use when continuing the Terraform migration work, reviewing how the design conclusions were reached, or reconstructing the evidence behind the ticket.
---



# Diary

## Goal

Capture the investigation path, evidence, and documentation outputs for the Terraform migration ticket so the next engineer can continue from concrete facts instead of memory.

## Step 1: Inventory the current Keycloak setup and turn it into a migration ticket

The point of this step was to stop treating the Keycloak setup as a collection of one-off fixes and start treating it as a system that needs an infrastructure-as-code owner. The user explicitly asked for a new ticket that covers the entire `smailnail` Keycloak setup, not only the Claude MCP client, and asked for a detailed guide suitable for a new intern.

The investigation quickly showed that the setup is split across multiple forms of state: a local realm import JSON, deployment docs, container entrypoint wiring, and hosted `kcadm.sh` scripts. That is enough evidence to justify a Terraform migration as a design problem rather than a small follow-up tweak.

### Prompt Context

**User prompt (verbatim):** "Then create a ticket to move our keycloak setup to terraform, not just the claude MCP client, but the entire smailnail setup. Add a detailed implementatino plan and tasks.

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** The user wants a new docmgr ticket focused on migrating the full Keycloak setup for `smailnail` to Terraform, plus a thorough intern-facing design and implementation guide, detailed tasks, and a reMarkable upload.

**Inferred user intent:** Replace the current scattered Keycloak setup knowledge with a structured migration plan that can be executed by another engineer and reviewed like real engineering work.

**Commit (code):** N/A

### What I did
- Created the ticket workspace `SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION`.
- Added the main design document and the diary document through `docmgr`.
- Searched the repository for existing Keycloak, OIDC, deployment, and Terraform-related artifacts.
- Read the local realm import, local Docker stack, hosted deployment docs, entrypoint wiring, and existing hosted `kcadm.sh` bootstrap script.
- Wrote the main design document with:
  - current-state architecture
  - gap analysis
  - proposed Terraform structure
  - migration strategy
  - testing strategy
  - risks and alternatives
- Replaced the placeholder task list with a phased implementation plan.

### Why
- The current setup already behaves like infrastructure but is not managed like infrastructure.
- The existing knowledge is spread across too many files and too many tickets.
- A Terraform migration without this analysis would likely recreate the current drift problems in a different format.

### What worked
- `docmgr ticket create-ticket` and `docmgr doc add` created the new workspace cleanly.
- Repository search quickly showed there is no current Terraform for Keycloak in `smailnail`.
- The local realm import file provided a strong evidence anchor for the exact current client and realm shape.
- The hosted `kcadm.sh` script provided strong evidence that production provisioning is still imperative.

### What didn't work
- There was no existing single authoritative document that already mapped the whole Keycloak setup end to end. The guide had to be synthesized from several sources.
- There was no obvious in-repo Terraform scaffold to extend, so the guide had to propose a layout instead of anchoring to existing code.

### What I learned
- The local setup is already declarative in spirit because it uses a committed realm import, but the format is a bootstrap artifact rather than a good long-term management format.
- The hosted setup is much more procedural than the local one, which increases the risk that local and hosted identity behavior drift over time.
- The application-side configuration surface is already stable enough to support a Terraform-managed Keycloak model without major app refactoring.

### What was tricky to build
- The main difficulty was defining the migration boundary correctly. It would have been easy to write a plan that only covered the Claude client because that is the immediate pain point. But the repository evidence shows that the real infrastructure problem is broader: realm settings, browser client, MCP client, local fixtures, and registration policy all belong to the same Keycloak system.
- Another subtle point was separating "Terraform should own identity infrastructure" from "Terraform should own all user data." Those are not the same. The guide therefore recommends keeping production users out of the first Terraform pass while still allowing local fixture users if the team wants reproducible local tests.

### What warrants a second pair of eyes
- The provider-coverage assumption around Keycloak registration policy should be verified before implementation starts.
- The proposed Terraform directory layout should be reviewed against the team's preferred infrastructure organization conventions.
- The plan to import the hosted realm rather than recreate it should be reviewed by whoever owns production rollback and incident response.

### What should be done in the future
- Add a follow-up playbook once the first Terraform files exist.
- Decide the exact provider version and state backend.
- Validate whether anonymous DCR policy can be managed declaratively or should be designed out by using pre-provisioned clients.

### Code review instructions
- Start with the main guide:
  - `go-go-mcp/ttmp/2026/03/18/SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION--move-the-full-smailnail-keycloak-setup-to-terraform/design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md`
- Then review the evidence anchors:
  - `smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json`
  - `smailnail/docker-compose.local.yml`
  - `smailnail/docs/deployments/smailnaild-oidc-keycloak.md`
  - `smailnail/docs/deployments/smailnaild-merged-coolify.md`
  - `go-go-mcp/ttmp/2026/03/16/SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT--deploy-smailnail-mcp-to-coolify-with-keycloak-external-oidc/scripts/create_keycloak_realm_and_mcp_client.sh`
- Validate documentation hygiene with:
  - `docmgr doctor --ticket SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION --stale-after 30`

### Technical details
- Commands used:
```bash
docmgr status --summary-only
docmgr ticket create-ticket --ticket SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION --title "move the full smailnail keycloak setup to terraform" --topics smailnail,keycloak,terraform,oidc,deployments
docmgr doc add --ticket SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION --doc-type design-doc --title "intern guide to migrating the full smailnail keycloak setup to terraform"
docmgr doc add --ticket SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION --doc-type reference --title "Diary"
rg -n --hidden -S "terraform|keycloak|realm-import|oidc|clientId|smailnail-web|smailnail-mcp|clients-registrations|kcadm" smailnail go-go-mcp
```

## Step 2: Land the first Terraform scaffold and validate it

This step turned the design into a real repository structure under `smailnail/deployments/terraform/keycloak`. The goal was not to finish the migration in one pass. The goal was to create a concrete starting point that expresses the current realm, browser client, MCP client, and local fixture boundaries in Terraform and then verify that the real provider accepts the structure.

The main technical correction in this step was provider resolution. The first attempt defined the provider only in the environment roots, which caused Terraform to treat the child modules as if they were using `hashicorp/keycloak`. That failed during `terraform init`. The fix was to add an explicit `required_providers` block with `source = "keycloak/keycloak"` in every module that uses Keycloak resources.

### Prompt Context

**User prompt (verbatim):** "yes"

**Assistant interpretation:** The user wants the next concrete step from the migration guide executed now, which means scaffolding the Terraform layout in the repository.

**Inferred user intent:** Move from planning into implementation so the Terraform migration has a real starting point instead of remaining documentation-only.

**Commit (code):** N/A

### What I did
- Added the scaffold root:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak`
- Added shared modules:
  - `modules/realm-base`
  - `modules/browser-client`
  - `modules/mcp-client`
  - `modules/local-fixtures`
- Added two environment roots:
  - `envs/local`
  - `envs/hosted`
- Added `.gitignore` for Terraform working artifacts.
- Ran:
  - `terraform fmt -recursive smailnail/deployments/terraform/keycloak`
  - `terraform init -backend=false && terraform validate` in both `envs/local` and `envs/hosted`

### Why
- A migration guide without an implementation starting point leaves too much invention for the next engineer.
- A validated scaffold forces the provider assumptions to become real.

### What worked
- Terraform 1.12.2 was already installed locally.
- The official `keycloak/keycloak` provider installed successfully.
- After adding explicit `required_providers` blocks to the modules, both environment roots initialized and validated successfully.

### What didn't work
- The first `terraform init` failed because Terraform tried to resolve `hashicorp/keycloak` for child modules.

Exact error:

```text
Could not retrieve the list of available versions for provider hashicorp/keycloak
Did you intend to use keycloak/keycloak? If so, you must specify that source address in each module which requires that provider.
```

### What I learned
- The module-level `required_providers` declarations are not optional here if we want reliable initialization with the Keycloak provider source.
- The basic resource surface used for the scaffold is accepted by provider `v5.7.0`.

### What was tricky to build
- The tricky part was choosing a scaffold that was real enough to validate but conservative enough not to over-encode unknown policy surfaces. The solution was to implement the realm and client baseline first, and leave advanced registration-policy modeling for a later phase after provider coverage is reviewed more carefully.

### What warrants a second pair of eyes
- The `mcp-client` module currently keeps the baseline hosted `smailnail-mcp` client shape. The later decision about dedicated OpenAI and Claude clients still needs review.
- The scaffold does not yet model advanced protocol mappers or DCR policy resources.

### What should be done in the future
- Add provider-backed modeling for scopes and mappers beyond the defaults.
- Decide whether DCR policy should be modeled in Terraform or designed around by preferring pre-provisioned clients.
- Add import instructions for the live hosted realm.

### Code review instructions
- Start at:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/README.md`
- Then inspect:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/modules/realm-base/main.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/modules/browser-client/main.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/modules/mcp-client/main.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/local/main.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted/main.tf`
- Re-run validation with:
```bash
terraform init -backend=false && terraform validate
```

### Technical details
- Commands used:
```bash
terraform version
terraform fmt -recursive smailnail/deployments/terraform/keycloak
terraform init -backend=false && terraform validate
```

## Step 3: Prove a real local Terraform apply against a sandbox Keycloak realm

This step was about moving from "the config validates" to "the provider can actually create and then reconcile real Keycloak objects." To avoid colliding with the existing imported `smailnail-dev` realm, I used a sandbox realm name, `smailnail-dev-tf`, while keeping the same local Keycloak container and the same general browser/MCP client shapes.

The first useful discovery here was that the provider supports `username` and `password`, so local Terraform does not need a dedicated admin client immediately. The second useful discovery was a drift quirk around default scopes: including `openid` in the managed default-scope lists caused a permanent follow-up diff because Keycloak did not persist it as a normal default client scope for these clients. The fix was to align the Terraform defaults with the actual persisted Keycloak behavior.

### Prompt Context

**User prompt (verbatim):** "ok, do 1"

**Assistant interpretation:** The user wants the safer next implementation step from the previous recommendation: make the local Terraform path real and prove it against local Keycloak before attempting hosted import work.

**Inferred user intent:** Reduce risk by validating the Terraform approach against a real local Keycloak before touching hosted state.

**Commit (code):** N/A

### What I did
- Updated the local provider config to support:
  - `username`
  - `password`
  - `client_id = admin-cli` by default
- Added `realm_name` and `realm_display_name` variables so local applies can target a sandbox realm.
- Ran a real local `terraform plan` against:
  - `keycloak_url=http://127.0.0.1:18080`
  - `realm_name=smailnail-dev-tf`
  - `realm_display_name=smailnail-dev-tf`
  - `web_client_secret=smailnail-web-secret`
- Ran a real `terraform apply -auto-approve` with the same inputs.
- Verified OIDC discovery at:
  - `http://127.0.0.1:18080/realms/smailnail-dev-tf/.well-known/openid-configuration`
- Investigated the follow-up drift on client default scopes with `kcadm.sh`.
- Adjusted Terraform defaults so the follow-up `plan` became a true no-op.
- Documented the working local verification commands in:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/README.md`

### Why
- A successful `terraform validate` is not enough. The provider and the target system still need to agree on real runtime behavior.
- Using a sandbox realm preserved the existing local imported realm while still proving the full create/reconcile loop.

### What worked
- The local Keycloak container was already running and healthy.
- Provider authentication via `admin/admin` and `admin-cli` worked.
- Terraform successfully created:
  - the sandbox realm
  - the browser client
  - the MCP client
  - the local `alice` fixture user
- OIDC discovery for the new realm returned the expected issuer.
- After adjusting default scope expectations, the same plan converged to `No changes`.

### What didn't work
- The first real plan failed because `keycloak_url` had not been provided.
- The first post-apply plan showed perpetual drift on default scopes involving `openid`.

Exact first variable error:

```text
Error: No value for required variable
The root module input variable "keycloak_url" is not set
```

Observed drift symptom:

```text
module.browser_client.keycloak_openid_client_default_scopes.this will be updated in-place
default_scopes = ["email", "profile"] -> add "openid"
```

### What I learned
- The local provider path does not need a dedicated admin client to get started because bootstrap admin auth is sufficient for local proof work.
- The Keycloak provider/resource model around default scopes is slightly different from the intuitive "just include `openid` everywhere" assumption.
- A real apply-and-replan loop is essential; otherwise provider quirks stay hidden.

### What was tricky to build
- The subtle part was not authentication. It was convergence. It is easy to declare success after the first apply, but infrastructure work is only healthy if the next plan is stable. The default-scope drift forced one more round of investigation against live Keycloak state with `kcadm.sh`, which is exactly the kind of issue that would have been painful to discover for the first time in hosted production.

### What warrants a second pair of eyes
- The decision to omit `openid` from managed default scopes should be reviewed against broader Keycloak conventions before expanding the module set.
- The local sandbox realm now exists and should either remain an intentional test fixture or be cleaned up later through Terraform.

### What should be done in the future
- Add import instructions for adopting the existing hosted realm into Terraform state.
- Decide whether a dedicated Terraform admin client is still desirable for local and hosted parity.
- Expand the scaffold to explicit mappers and any registration-policy resources we decide to manage.

### Code review instructions
- Review the local provider/auth changes:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/local/providers.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/local/variables.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/local/main.tf`
- Review the scope-alignment fix:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/modules/browser-client/variables.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/modules/mcp-client/variables.tf`
- Review the local runbook:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/README.md`

### Technical details
- Commands used:
```bash
terraform plan -var='keycloak_url=http://127.0.0.1:18080' -var='realm_name=smailnail-dev-tf' -var='realm_display_name=smailnail-dev-tf' -var='web_client_secret=smailnail-web-secret'
terraform apply -auto-approve -var='keycloak_url=http://127.0.0.1:18080' -var='realm_name=smailnail-dev-tf' -var='realm_display_name=smailnail-dev-tf' -var='web_client_secret=smailnail-web-secret'
curl -fsS http://127.0.0.1:18080/realms/smailnail-dev-tf/.well-known/openid-configuration | jq -r '.issuer'
docker exec smailnail-keycloak /opt/keycloak/bin/kcadm.sh get clients/.../default-client-scopes -r smailnail-dev-tf
```

## Step 4: Import the live hosted realm and discover the provider import limits

This step moved from local proof into hosted adoption. The goal was not to apply Terraform to production yet. The goal was to get the live hosted Keycloak realm under Terraform state inspection so future work can be based on a real `terraform plan` instead of guesswork.

To make that possible, I first extended the hosted Terraform provider config to support bootstrap admin `username` and `password`, matching the local environment. Then I used the live Keycloak bootstrap admin credentials from the running container to import the hosted `smailnail` realm and the two currently existing hosted clients, `smailnail-web` and `smailnail-mcp`.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** After the local Terraform path was proven, the next requested action was to continue into hosted import work.

**Inferred user intent:** Start bringing the real hosted Keycloak realm under Terraform state without yet making configuration changes to production.

**Commit (code):** N/A

### What I did
- Updated hosted Terraform provider config to support:
  - `keycloak_username`
  - `keycloak_password`
  - default `client_id = admin-cli`
- Queried the live hosted Keycloak container to find:
  - bootstrap admin username
  - bootstrap admin password
  - `smailnail-web` client UUID
  - `smailnail-mcp` client UUID
  - current `smailnail-web` secret
- Imported into hosted Terraform state:
  - `module.realm.keycloak_realm.this`
  - `module.browser_client.keycloak_openid_client.this`
  - `module.mcp_client.keycloak_openid_client.this`
- Ran hosted `terraform plan` afterward to inspect remaining drift.

### Why
- Importing the live realm is the safest first hosted step because it does not change production objects.
- A post-import plan reveals the actual differences between live production and the current HCL.

### What worked
- Hosted provider authentication with bootstrap admin credentials worked.
- Import of the hosted realm succeeded.
- Import of both hosted clients succeeded once the correct import ID format was used.
- Hosted state now contains:
  - `module.realm.keycloak_realm.this`
  - `module.browser_client.keycloak_openid_client.this`
  - `module.mcp_client.keycloak_openid_client.this`
- The follow-up plan exposed meaningful drift rather than "everything will be created".

### What didn't work
- The first attempt to import `keycloak_openid_client` using `smailnail/smailnail-web` failed.
- Attempting to import `keycloak_openid_client_default_scopes` failed because the resource does not support import.

Exact client import failure:

```text
error sending GET request to /admin/realms/smailnail/clients/smailnail-web: 404 Not Found
```

Exact scope-import failure:

```text
resource keycloak_openid_client_default_scopes doesn't support import
```

### What I learned
- For this provider, `keycloak_openid_client` import IDs use:
  - `realm_id/internal_client_uuid`
  not the public `clientId` string.
- The client-scope attachment helper resources are create/update-only from Terraform’s point of view and cannot be imported into state.
- The current live hosted web and MCP clients still point at the older `smailnail.mcp.scapegoat.dev` host in places where the current HCL expects `smailnail.scapegoat.dev`.

### What was tricky to build
- The import format ambiguity was the main trap. The provider error message says the supported format is `{{realmId}}/{{openidClientId}}`, but the actual operational value for the second segment was the internal client UUID, not the public `clientId`. The only reliable way to resolve that was to test it against the live provider behavior.
- Another subtle issue was distinguishing "import support" from "resource exists." The client-scope attachment resources clearly exist conceptually, but the provider does not implement import for them, which means adoption planning must account for create/update-only helper resources.

### What warrants a second pair of eyes
- The hosted plan still shows real drift:
  - old vs new public callback/origin hostnames
  - `use_refresh_tokens` differences
  - a realm-level `default_signature_algorithm` field
- Those should be reviewed before any hosted apply.

### What should be done in the future
- Write an explicit hosted import playbook with the exact import IDs and credential requirements.
- Decide whether the old `smailnail.mcp.scapegoat.dev` host should remain or whether Terraform should intentionally converge the clients to `smailnail.scapegoat.dev`.
- Decide how to handle non-importable client-scope attachment resources in the adoption workflow.

### Code review instructions
- Review hosted provider/auth changes:
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted/providers.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted/variables.tf`
  - `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted/terraform.tfvars.example`
- Review current hosted Terraform state with:
```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/hosted
terraform state list
```
- Review drift with:
```bash
terraform plan ...
```

### Technical details
- Commands used:
```bash
terraform import module.realm.keycloak_realm.this smailnail
terraform import module.browser_client.keycloak_openid_client.this smailnail/4a6dfa44-5dd3-48ae-90ae-2581ae3953b9
terraform import module.mcp_client.keycloak_openid_client.this smailnail/7cdece72-a92a-4715-96ba-19c326b22f27
terraform state list
terraform plan ...
```

## Related

- Main design guide: [../design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md](../design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md)
- Ticket index: [../index.md](../index.md)
