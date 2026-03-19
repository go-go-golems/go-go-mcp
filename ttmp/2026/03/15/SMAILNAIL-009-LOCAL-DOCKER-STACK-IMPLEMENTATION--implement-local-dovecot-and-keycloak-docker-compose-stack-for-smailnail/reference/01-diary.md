---
Title: Diary
Ticket: SMAILNAIL-009-LOCAL-DOCKER-STACK-IMPLEMENTATION
Status: complete
Topics:
    - smailnail
    - go
    - sql
    - keycloak
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docker-compose.local.yml
    - /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json
    - /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md
ExternalSources: []
Summary: Chronological implementation diary for the local Dovecot plus Keycloak stack.
LastUpdated: 2026-03-15T19:28:00-04:00
WhatFor: Preserve exact commands, verification steps, and the local environment decisions made during implementation.
WhenToUse: Use when reviewing or reproducing the local stack setup work.
---

# Diary

## Goal

Stand up a repeatable local Docker stack for both IMAP and OIDC development so `smailnail` work can target a real Dovecot server and a real Keycloak realm.

## Context

The repo already had an IMAP-oriented workflow and a new `smailnaild` skeleton, but no checked-in local OIDC provider. The user asked specifically to set up Keycloak in the Docker Compose stack and bring the stack up so Dovecot and Keycloak are both available for development.

## Quick Reference

Start the stack:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml up -d
```

Verify the services:

```bash
docker compose -f docker-compose.local.yml ps
curl -sf http://127.0.0.1:18080/realms/smailnail-dev/.well-known/openid-configuration
bash -lc 'exec 3<>/dev/tcp/127.0.0.1/993 && echo open && exec 3<&- && exec 3>&-'
```

Local defaults:

- Keycloak admin: `http://127.0.0.1:18080/admin`
- Keycloak bootstrap admin: `admin` / `admin`
- Imported realm: `smailnail-dev`
- Issuer: `http://127.0.0.1:18080/realms/smailnail-dev`
- Dovecot IMAPS: `127.0.0.1:993`
- Dovecot users: `a`, `b`, `c`, `d`
- Dovecot password: `pass`

## Usage Examples

Step 1. Create the ticket workspace.

Commands:

```bash
docmgr ticket create-ticket \
  --ticket SMAILNAIL-009-LOCAL-DOCKER-STACK-IMPLEMENTATION \
  --title "Implement local Dovecot and Keycloak docker compose stack for smailnail" \
  --topics smailnail,go,sql,keycloak
docmgr doc add --ticket SMAILNAIL-009-LOCAL-DOCKER-STACK-IMPLEMENTATION --doc-type design-doc --title "Implementation plan for local Dovecot and Keycloak stack"
docmgr doc add --ticket SMAILNAIL-009-LOCAL-DOCKER-STACK-IMPLEMENTATION --doc-type reference --title "Diary"
```

Result:

- ticket workspace created under `go-go-mcp/ttmp/2026/03/15/...`
- design doc and diary scaffolds created

Step 2. Add the local stack files.

Changes:

- added `smailnail/docker-compose.local.yml`
- added `smailnail/dev/keycloak/realm-import/smailnail-dev-realm.json`
- updated `smailnail/README.md`
- updated `smailnail/.gitignore` to ignore local `data/`

Intent:

- keep the stack self-contained in the repo
- make Keycloak startup deterministic by importing a realm instead of relying on manual console setup

Step 3. Validate the compose definition.

Command:

```bash
docker compose -f docker-compose.local.yml config
```

Result:

- compose rendering succeeded with no schema errors

Step 4. Start the stack.

Command:

```bash
docker compose -f docker-compose.local.yml up -d
```

Result:

- images pulled successfully
- named volumes created
- `smailnail-dovecot`, `smailnail-keycloak-postgres`, and `smailnail-keycloak` started

Step 5. Verify Keycloak import and container status.

Commands:

```bash
docker compose -f docker-compose.local.yml ps
docker compose -f docker-compose.local.yml logs --no-color --tail=120 keycloak
curl -sf http://127.0.0.1:18080/realms/smailnail-dev/.well-known/openid-configuration
```

Observed results:

- compose status showed all three services up and healthy where health checks exist
- Keycloak logs showed `Realm 'smailnail-dev' imported`
- OIDC discovery resolved successfully and returned issuer `http://127.0.0.1:18080/realms/smailnail-dev`

Step 6. Verify Dovecot reachability.

Command:

```bash
bash -lc 'exec 3<>/dev/tcp/127.0.0.1/993 && echo open && exec 3<&- && exec 3>&-'
```

Result:

- returned `open`, confirming IMAPS is listening on the host-mapped port

One wrinkle during verification:

- non-escalated localhost network probes were blocked by the sandbox and returned connection or permission failures
- the same probes succeeded once re-run with escalated permissions
- this was an execution-environment restriction, not a problem with the compose stack

## Related

- Ticket index: `../index.md`
- Implementation plan: `../design-doc/01-implementation-plan-for-local-dovecot-and-keycloak-stack.md`
