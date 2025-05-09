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

## Integration Notes

This CLI tool was merged from the standalone `arxiv-libgen-cli` project into the go-go-mcp ecosystem. The code was reorganized as follows:

- All package code moved to `pkg/scholarly/`
- Command files moved to `cmd/apps/scholarly/cmd/`
- Main entry point at `cmd/apps/scholarly/main.go`

Import paths were updated to use the go-go-mcp module path. 