---
Title: move the full smailnail keycloak setup to terraform
Ticket: SMAILNAIL-017-KEYCLOAK-TERRAFORM-MIGRATION
Status: active
Topics:
    - smailnail
    - keycloak
    - terraform
    - oidc
    - deployments
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-18T17:14:19.133883955-04:00
WhatFor: Move the full smailnail Keycloak setup from ad-hoc realm imports and shell commands to a reproducible Terraform-managed model across local and hosted environments.
WhenToUse: Use when planning or implementing infrastructure-as-code for smailnail OIDC, Keycloak client setup, realm policy, and connector compatibility.
---

# move the full smailnail keycloak setup to terraform

## Overview

This ticket captures the design and implementation plan for migrating the entire `smailnail` Keycloak setup to Terraform. The target is not only the Claude MCP client. The target is the whole identity surface used by:

- `smailnaild` browser login
- hosted `/mcp` bearer-token auth
- local Keycloak developer fixtures
- future OpenAI and Claude connector compatibility

The central problem is that the current state is spread across three different mechanisms:

- local realm import JSON under `smailnail/dev/keycloak/realm-import/`
- hosted manual or scripted `kcadm.sh` operations
- narrative deployment docs and ticket notes

That split makes drift likely and auditability poor. The migration goal is a single declarative source of truth for the Keycloak realm, clients, scopes, redirect URIs, and registration policy choices.

## Key Links

- Main design document: [design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md](./design-doc/01-intern-guide-to-migrating-the-full-smailnail-keycloak-setup-to-terraform.md)
- Diary: [reference/01-diary.md](./reference/01-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**

## Topics

- smailnail
- keycloak
- terraform
- oidc
- deployments

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design/`
  - primary migration guide and implementation plan
- `reference/`
  - chronological investigation diary
- `playbooks/`
  - reserved for future operator runbooks such as import, plan, apply, and verification
- `scripts/`
  - reserved for migration helper scripts if needed

## Current conclusion

The current recommended direction is:

1. Manage the whole `smailnail` Keycloak realm with Terraform, not just a single client.
2. Keep developer-only users and convenience fixtures clearly separated from production state.
3. Prefer pre-provisioned clients for external MCP products first.
4. Treat anonymous DCR as an explicit policy decision, not an accidental side effect of manual setup.

An initial Terraform scaffold now exists under:

- `/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak`
