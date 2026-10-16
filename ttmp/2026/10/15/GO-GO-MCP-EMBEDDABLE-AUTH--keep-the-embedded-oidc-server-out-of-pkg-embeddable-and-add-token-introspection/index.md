---
Title: Keep the embedded OIDC server out of pkg/embeddable and add token introspection
Ticket: GO-GO-MCP-EMBEDDABLE-AUTH
Status: active
Topics:
    - mcp
    - oidc
    - authentication
    - introspection
    - dependencies
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://ttmp/2026/10/15/GO-GO-MCP-EMBEDDABLE-AUTH--keep-the-embedded-oidc-server-out-of-pkg-embeddable-and-add-token-introspection/design-doc/01-embedded-oidc-server-out-of-pkg-embeddable-and-token-introspection-analysis-design-and-implementation-guide.md
      Note: Primary design guide
    - Path: repo://ttmp/2026/10/15/GO-GO-MCP-EMBEDDABLE-AUTH--keep-the-embedded-oidc-server-out-of-pkg-embeddable-and-add-token-introspection/working-note/01-implementation-diary.md
      Note: Implementation diary
ExternalSources: []
Summary: "Link the embedded development OIDC server only on request (pkg/embeddable/embeddeddev), fix embedded_dev expiration, and add RFC 7662 introspection to external_oidc."
LastUpdated: 2026-10-15T19:12:21.617052228-04:00
WhatFor: ""
WhenToUse: ""
---

# Keep the embedded OIDC server out of pkg/embeddable and add token introspection

## Overview

`pkg/embeddable` links go-go-mcp's embedded development OIDC server into every program (318 compiled packages instead of 61), and `external_oidc` cannot verify opaque access tokens. This ticket moves `embedded_dev` behind an explicit option in a new package, fixes `embedded_dev` tokens being rejected for missing expiration, and adds token introspection.

- Design: [design-doc/01](design-doc/01-embedded-oidc-server-out-of-pkg-embeddable-and-token-introspection-analysis-design-and-implementation-guide.md)
- Diary: [working-note/01](working-note/01-implementation-diary.md)

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- mcp
- oidc
- authentication
- introspection
- dependencies

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
