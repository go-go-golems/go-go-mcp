---
Title: Implementation diary for merging smailnaild and MCP into one hosted server
Ticket: SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT
Status: active
Topics:
    - smailnail
    - mcp
    - oidc
    - keycloak
    - coolify
    - deployments
    - diary
DocType: reference
Intent: long-term
Summary: Chronological record of the server-merge and deployment work, including failed attempts, validation commands, and rollout notes.
---

# Implementation diary for merging smailnaild and MCP into one hosted server

## Kickoff

### Prompt context

The user wants to stop treating the hosted web app and the hosted MCP as separate production services and instead serve them from the same binary and the same `http.Server`. The immediate request is to create a new ticket and make it actionable enough to drive the actual refactor and deployment work.

### What I did

I created ticket `SMAILNAIL-015-MERGED-HOSTED-SERVER-DEPLOYMENT` and added:

- a detailed implementation plan
- a granular task breakdown
- this diary as the running execution log

### Why

The merge is large enough that it needs a fresh execution record rather than being buried inside older identity or deployment tickets. The older tickets explain why the current split exists; this new ticket should explain how to remove that split safely.

### Initial assumptions

- the single-server design is desirable
- the merge should keep browser-session auth and bearer-token auth distinct
- the deployment cutover should happen only after a local proof and a hosted smoke

### Immediate next step

Start with the code-shape refactor:

- extract a mountable MCP handler from the standalone server package
- then mount it into `smailnaild`
