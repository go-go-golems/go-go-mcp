---
Title: Concrete deployment plan for smailnail-mcp on Coolify with Keycloak
Ticket: SMAILNAIL-010-MCP-COOLIFY-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - coolify
    - keycloak
    - deployments
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Concrete production deployment plan for packaging smailnail-mcp, publishing it on Coolify behind smailnail.mcp.scapegoat.dev, and protecting it with external Keycloak OIDC.
LastUpdated: 2026-03-16T00:10:00-04:00
WhatFor: Turn the existing MCP binary into a deployable service with a public HTTPS endpoint and a matching IMAP test target.
WhenToUse: Use when implementing the deployment slice or reviewing whether the production deployment shape is coherent.
---

# Concrete deployment plan for smailnail-mcp on Coolify with Keycloak

## Executive Summary

The fastest path to a real hosted integration is to deploy `smailnail-imap-mcp` as its own service, separate from `smailnaild`. The production target is:

- MCP base URL: `https://smailnail.mcp.scapegoat.dev/mcp`
- OIDC issuer: `https://auth.scapegoat.dev/realms/smailnail`

The first implementation slice should produce:

- a production Docker image for the MCP binary
- stable runtime defaults for HTTP deployment
- documented Coolify environment/command settings
- a validated external-OIDC configuration against Keycloak
- a separate hosted Dovecot target for remote end-to-end testing

## Problem Statement

`smailnail-imap-mcp` already exists as a local binary and already supports `external_oidc` through `go-go-mcp`, but the deployment path is still missing:

- there is no Dockerfile for the MCP binary
- there is no deployment-specific `.dockerignore`
- there are no Coolify deployment instructions or environment examples
- there is no production hostname wired into the repository docs
- there is no hosted IMAP target analogous to the local Dovecot fixture

Without these pieces, the HTTPS/OIDC path for Claude Desktop and other remote clients remains theoretical.

## Proposed Solution

Implement the hosted slice in two parts.

Part 1: MCP packaging and deployment

- add a production Dockerfile for `cmd/smailnail-imap-mcp`
- add any small runtime-default improvements needed so the service naturally runs on `streamable_http` with the intended HTTP port
- document the exact production start command using:
  - `--auth-mode external_oidc`
  - `--auth-resource-url https://smailnail.mcp.scapegoat.dev/mcp`
  - `--oidc-issuer-url https://auth.scapegoat.dev/realms/smailnail`
- validate the container locally
- deploy it to the Hetzner/Coolify host and verify:
  - `/.well-known/oauth-protected-resource`
  - `401` on unauthenticated `/mcp`

Part 2: Hosted Dovecot target

- stand up a separate Dovecot service on the same infrastructure, with isolated storage
- expose a stable testing configuration that mirrors the local fixture intent
- document hostnames, ports, credentials, and safety notes so hosted MCP tests have a known-good IMAP target

## Design Decisions

Deploy the MCP binary separately from `smailnaild`

This avoids waiting for the broader hosted app work. The MCP binary already exists and already has the OIDC hooks that matter for Claude/Desktop and remote clients.

Use `streamable_http` as the primary production transport

That keeps the public contract simple: a single `/mcp` endpoint behind standard HTTPS. SSE can remain available for debugging, but the deployment target should optimize for the connector path.

Use external Keycloak, not embedded OIDC

The production host already has Keycloak with a real public domain. Reusing it avoids deploying a second authorization server inside the MCP container.

Provide a separate hosted IMAP target

Remote MCP testing is much less convincing if it still depends on a developer laptop or ad hoc third-party IMAP credentials. A dedicated Dovecot target closes that gap.

## Alternatives Considered

Wait for `smailnaild` and deploy everything together

Rejected because it delays the first real remote MCP/OIDC validation without adding value to the MCP packaging work.

Use embedded OIDC in the MCP container

Rejected because a real public Keycloak issuer already exists and better matches the intended production architecture.

Rely only on the local Dovecot fixture

Rejected because it does not give the hosted MCP service a remotely reachable IMAP endpoint.

## Implementation Plan

1. Ticket and documentation setup.
2. Repository implementation:
   - Dockerfile
   - `.dockerignore`
   - runtime default improvements
   - deployment docs
3. Local validation:
   - `go test`
   - `go build`
   - container build
   - local container smoke
4. Keycloak deployment-side review:
   - realm name
   - client config
   - callback/resource expectations
5. Coolify/Hetzner deployment:
   - publish MCP container
   - bind hostname `smailnail.mcp.scapegoat.dev`
   - verify HTTPS and auth behavior
6. Hosted Dovecot deployment and documentation.

## Open Questions

Whether the actual Coolify application creation can be completed headlessly from the server or whether the web UI login will be needed to register the new app cleanly.

Whether the hosted Dovecot target should be exposed directly on standard IMAP ports or fronted through a more explicit testing hostname/routing setup.

## References

- Previous architecture ticket: `../../../../03/08/SMAILNAIL-003-COOLIFY-SSO-MCP-DESIGN--deploy-smailnail-on-coolify-with-github-sso-per-user-imap-config-and-mcp-oidc/index.md`
- Coolify configuration diary: `/home/manuel/code/wesen/2026-03-15--install-coolify/ttmp/2026/03/15/COOLIFY-001--configure-coolify-on-hetzner-server/reference/01-diary.md`
