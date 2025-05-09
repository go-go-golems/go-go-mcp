# Scholarly CLI

`scholarly` is a command-line tool integrated into go-go-mcp that allows users to search for scientific papers across multiple academic databases and repositories including:

- ArXiv
- Library Genesis (LibGen)
- Crossref
- OpenAlex

## Features

- Search multiple academic APIs with a consistent interface
- Download papers and documents when available
- Get metrics and citation information for papers
- Extract keywords from text
- Resolve DOIs to full metadata
- Find fulltext versions of papers

## Usage

The tool provides specific subcommands for each platform, allowing targeted searches with various filters and options.

```
scholarly arxiv -q "all:electron" -n 5
scholarly libgen -q "artificial intelligence" -m "https://libgen.is"
scholarly crossref -q "climate change mitigation"
scholarly openalex -q "machine learning applications"
```

### Available Commands

- `arxiv`: Search ArXiv for scientific papers
- `crossref`: Search Crossref for scientific papers
- `libgen`: Search Library Genesis for books and papers
- `openalex`: Search OpenAlex for scientific papers
- `citations`: Get citation information for a paper
- `doi`: Resolve a DOI to get paper metadata
- `fulltext`: Find fulltext versions of a paper
- `keywords`: Extract keywords from text
- `metrics`: Get metrics for a paper
- `search`: Search across multiple sources with one query
- `mcp`: Start an MCP server exposing scholarly tools

## MCP Integration

Scholarly also functions as an MCP tool provider, exposing its functionality to AI models and other clients via the Model Control Protocol (MCP). This allows models like Claude to directly interact with academic databases.

### Running as an MCP Server

Start the MCP server on port 8080:

```
scholarly mcp --port 8080 --host 0.0.0.0
```

### Available MCP Tools

The following tools are exposed via MCP:

1. **mcp_scholarly_search_works** - Search across academic databases
   - Parameters: `query`, `source` (arxiv/openalex/crossref), `limit`, `filter`

2. **mcp_scholarly_resolve_doi** - Retrieve detailed metadata for a DOI
   - Parameters: `doi`

3. **mcp_scholarly_suggest_keywords** - Extract academic concepts from text
   - Parameters: `text`, `max_keywords`

4. **mcp_scholarly_get_metrics** - Get citation metrics for a work
   - Parameters: `work_id`

5. **mcp_scholarly_get_citations** - Get citation relationships
   - Parameters: `work_id`, `direction` (cited_by/references), `limit`

6. **mcp_scholarly_find_full_text** - Find fulltext PDF/HTML links
   - Parameters: `doi` and/or `title`, `prefer_version`

### Using as a Library

You can also embed the scholarly MCP tools in your own Go applications:

```go
import (
    "github.com/go-go-golems/go-go-mcp/pkg/scholarly"
    "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
)

// Create a registry and register the scholarly tools
registry := tool_registry.NewRegistry()
err := scholarly.RegisterScholarlyTools(registry)
```

## Integration Notes

This CLI tool was merged from the standalone `arxiv-libgen-cli` project into the go-go-mcp ecosystem. The code was reorganized as follows:

- All package code moved to `pkg/scholarly/`
- Command files moved to `cmd/apps/scholarly/cmd/`
- Main entry point at `cmd/apps/scholarly/main.go`

Import paths were updated to use the go-go-mcp module path. 