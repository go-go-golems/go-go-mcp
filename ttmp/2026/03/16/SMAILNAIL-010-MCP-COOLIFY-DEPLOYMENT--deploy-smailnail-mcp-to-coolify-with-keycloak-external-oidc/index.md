---
Title: Deploy smailnail-mcp to Coolify with Keycloak external OIDC
Ticket: SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - coolify
    - keycloak
    - deployments
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Implementation ticket for packaging, deploying, and validating smailnail-mcp on Coolify with Keycloak-backed OIDC and a companion hosted Dovecot test target.
LastUpdated: 2026-03-16T05:07:00-04:00
WhatFor: Package the MCP server for production, deploy it behind HTTPS on Coolify, and provision a hosted IMAP target for remote end-to-end testing.
WhenToUse: Use when implementing or reviewing the production deployment path for smailnail-mcp and its hosted IMAP test environment.
---

# Deploy smailnail-mcp to Coolify with Keycloak external OIDC

## Overview

This ticket covers the first deployable hosted slice for `smailnail`:

- package `smailnail-imap-mcp` into a production container image
- deploy it to the Coolify host behind `https://smailnail.mcp.scapegoat.dev`
- protect the HTTP MCP endpoint with external Keycloak OIDC from `https://auth.scapegoat.dev`
- provision a separate hosted Dovecot target on the same Coolify machine so the hosted MCP can be tested against a real IMAP server, not only the local fixture

The work starts from the current state where the MCP binary exists and supports `external_oidc`, but the repository has no production Docker packaging or Coolify deployment instructions yet.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- smailnail
- mcp
- coolify
- keycloak
- deployments

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
