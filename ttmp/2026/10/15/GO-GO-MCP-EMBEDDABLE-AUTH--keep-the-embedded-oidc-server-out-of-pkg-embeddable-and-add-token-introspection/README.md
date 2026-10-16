# Keep the embedded OIDC server out of pkg/embeddable and add token introspection

This is the document workspace for ticket GO-GO-MCP-EMBEDDABLE-AUTH.

## Structure

- **design/**: Design documents and architecture notes
- **reference/**: Reference documentation and API contracts
- **playbooks/**: Operational playbooks and procedures
- **scripts/**: Utility scripts and automation
- **sources/**: External sources and imported documents
- **various/**: Scratch or meeting notes, working notes
- **archive/**: Optional space for deprecated or reference-only artifacts

## Getting Started

Use docmgr commands to manage this workspace:

- Add documents: `docmgr doc add --ticket GO-GO-MCP-EMBEDDABLE-AUTH --doc-type design-doc --title "My Design"`
- Import sources: `docmgr import file --ticket GO-GO-MCP-EMBEDDABLE-AUTH --file /path/to/doc.md`
- Update metadata: `docmgr meta update --ticket GO-GO-MCP-EMBEDDABLE-AUTH --field Status --value review`
