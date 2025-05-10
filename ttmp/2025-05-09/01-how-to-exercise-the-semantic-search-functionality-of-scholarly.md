# How to Exercise the Semantic Search Functionality of Scholarly

This document provides a detailed plan for testing the various search capabilities of the scholarly search tool, using both the command-line interface and the MCP tool calls. The plan is designed to systematically test different filters and sources to ensure comprehensive coverage of the tool's functionality.

## Table of Contents

1. [Introduction](#introduction)
2. [Tool Overview](#tool-overview)
3. [Testing arXiv Source](#testing-arxiv-source)
4. [Testing OpenAlex Source](#testing-openalex-source)
5. [Testing Crossref Source](#testing-crossref-source)
6. [Testing Multi-Source Search](#testing-multi-source-search)
7. [MCP Tool Call Testing](#mcp-tool-call-testing)
8. [Advanced Use Cases](#advanced-use-cases)
9. [Troubleshooting](#troubleshooting)

## Introduction

The scholarly search tool provides a unified interface for searching academic papers across multiple sources, including arXiv, OpenAlex, and Crossref. Each source has its own set of filters and parameters that can be used to refine search results. This guide will demonstrate how to effectively test and utilize these features.

## Tool Overview

The scholarly tool is implemented as a Go application with the following components:

- **Command-line interface**: Located at `./cmd/apps/scholarly`
- **Source-specific clients**: Located in `/pkg/scholarly/[source]` directories
- **Data models**: Each source has its own models for parsing API responses
- **MCP tool call interface**: Allows programmatic access to the search functionality

## Testing arXiv Source

ArXiv allows searching pre-prints and academic papers with specific query syntax and category filtering. Testing this source will focus on the query syntax, category filtering, and author filtering.

### Basic Search

```bash
# Basic keyword search
go run ./cmd/apps/scholarly arxiv -q "quantum computing" -n 5 --output markdown

# Search with specific field syntax
go run ./cmd/apps/scholarly arxiv -q "ti:neural networks AND au:Schmidhuber" -n 3 --output json

# Category-specific search
go run ./cmd/apps/scholarly arxiv -q "reinforcement learning" -n 5 --output markdown
```

### Advanced Filtering

ArXiv uses a specific query syntax where:
- `ti:` searches in titles
- `au:` searches for authors
- `abs:` searches in abstracts
- `all:` searches across all fields
- Boolean operators like `AND`, `OR`, and `NOT` can be used

```bash
# Complex query with multiple conditions
go run ./cmd/apps/scholarly arxiv -q "ti:\"long short-term memory\" AND (au:Schmidhuber OR au:Hochreiter)" -n 3

# Search in specific categories
go run ./cmd/apps/scholarly arxiv -q "cat:cs.AI transfer learning" -n 5
```

Reference: See query field handling in [`/pkg/scholarly/arxiv/client.go`](https://github.com/go-go-golems/go-go-mcp/blob/main/pkg/scholarly/arxiv/client.go).

## Testing OpenAlex Source

OpenAlex is a comprehensive database of academic papers with rich filter options including open access status, publication year, and institutional affiliations.

### Basic Search

```bash
# Basic search with polite pool (add your email)
go run ./cmd/apps/scholarly openalex -q "climate change" -n 5 -m "your.email@example.com"

# Search with output format options
go run ./cmd/apps/scholarly openalex -q "machine learning" -n 3 --output json
```

### Filter Testing

OpenAlex supports a wide range of filters through the `-f` flag:

```bash
# Filter by publication year
go run ./cmd/apps/scholarly openalex -q "neural networks" -f "publication_year:2023" -n 5

# Filter by open access status
go run ./cmd/apps/scholarly openalex -q "genomics" -f "is_oa:true" -n 5

# Filter by publication type
go run ./cmd/apps/scholarly openalex -q "psychology" -f "type:journal-article" -n 5

# Combined filters
go run ./cmd/apps/scholarly openalex -q "artificial intelligence" -f "publication_year:2023,is_oa:true,type:journal-article" -n 5
```

### Sorting

OpenAlex supports custom sorting through the `-s` flag:

```bash
# Sort by citation count (descending)
go run ./cmd/apps/scholarly openalex -q "quantum physics" -s "cited_by_count:desc" -n 5

# Sort by publication date (newest first)
go run ./cmd/apps/scholarly openalex -q "cancer research" -s "publication_date:desc" -n 5
```

Reference: Filter implementation in [`/pkg/scholarly/openalex/client.go`](https://github.com/go-go-golems/go-go-mcp/blob/main/pkg/scholarly/openalex/client.go) lines 25-65.

## Testing Crossref Source

Crossref is a comprehensive database of published academic work with strong support for date-based and publication type filtering.

### Basic Search

```bash
# Basic search with polite pool
go run ./cmd/apps/scholarly crossref -q "machine learning" -n 5 -m "your.email@example.com"

# Output as JSON for processing
go run ./cmd/apps/scholarly crossref -q "deep learning" -n 3 --output json
```

### Filter Testing

Crossref uses a different filter syntax through the `-f` flag:

```bash
# Filter by date range
go run ./cmd/apps/scholarly crossref -q "renewable energy" -f "from-pub-date:2022-01-01,until-pub-date:2022-12-31" -n 5

# Filter by publication type
go run ./cmd/apps/scholarly crossref -q "covid-19" -f "type:journal-article" -n 5

# Filter by publisher
go run ./cmd/apps/scholarly crossref -q "particle physics" -f "publisher:IEEE" -n 5

# Combined filters
go run ./cmd/apps/scholarly crossref -q "immunology" -f "from-pub-date:2023-01-01,type:journal-article" -n 5
```

Reference: Filter handling in [`/pkg/scholarly/crossref/client.go`](https://github.com/go-go-golems/go-go-mcp/blob/main/pkg/scholarly/crossref/client.go) lines 30-40.

## Testing Multi-Source Search

The tool provides a unified search interface through the `search` command, which can be used to search across multiple sources.

```bash
# Search across all sources
go run ./cmd/apps/scholarly search -q "machine learning" -l 5

# Search specific source
go run ./cmd/apps/scholarly search -q "climate change" -s "openalex" -l 5

# Search with filters
go run ./cmd/apps/scholarly search -q "neural networks" -s "crossref" -f "type:journal-article" -l 5

# Output as JSON
go run ./cmd/apps/scholarly search -q "quantum computing" -s "arxiv" -j
```

## MCP Tool Call Testing

The MCP scholarly search function can be used programmatically through the MCP interface. Here are various configurations to test:

### Basic Query Structure

```json
{
  "query": "machine learning",
  "source": "arxiv",
  "limit": 5
}
```

### arXiv-Specific Queries

```json
{
  "query": "ti:\"neural networks\" AND au:Schmidhuber",
  "source": "arxiv",
  "filter": {
    "category": "cs.AI"
  },
  "limit": 5
}
```

### OpenAlex-Specific Queries

```json
{
  "query": "climate change adaptation",
  "source": "openalex",
  "filter": {
    "is_oa": true,
    "publication_year": 2023,
    "type": "journal-article"
  },
  "limit": 5
}
```

### Crossref-Specific Queries

```json
{
  "query": "renewable energy",
  "source": "crossref",
  "filter": {
    "from-pub-date": "2023-01-01",
    "until-pub-date": "2023-12-31",
    "type": "journal-article"
  },
  "limit": 5
}
```

### Multi-Source Batch Query

```json
{
  "queries": [
    {
      "query": "reinforcement learning",
      "source": "arxiv",
      "filter": {
        "category": "cs.LG"
      },
      "limit": 5
    },
    {
      "query": "reinforcement learning",
      "source": "openalex",
      "filter": {
        "is_oa": true,
        "publication_year": 2023
      },
      "limit": 5
    },
    {
      "query": "reinforcement learning",
      "source": "crossref",
      "filter": {
        "type": "journal-article",
        "from-pub-date": "2023-01-01"
      },
      "limit": 5
    }
  ]
}
```

## Advanced Use Cases

### Finding Specific Authors' Works

```bash
# arXiv - Find papers by a specific author
go run ./cmd/apps/scholarly arxiv -q "au:\"Schmidhuber\"" -n 10

# OpenAlex - Find highly cited papers by an author
go run ./cmd/apps/scholarly openalex -q "Hinton" -f "type:journal-article" -s "cited_by_count:desc" -n 5

# Crossref - Find recent papers by an author
go run ./cmd/apps/scholarly crossref -q "LeCun" -f "from-pub-date:2020-01-01" -n 5
```

### Tracking Research Trends

```bash
# Compare publication counts over years
go run ./cmd/apps/scholarly openalex -q "large language models" -f "publication_year:2020" -n 1
go run ./cmd/apps/scholarly openalex -q "large language models" -f "publication_year:2021" -n 1
go run ./cmd/apps/scholarly openalex -q "large language models" -f "publication_year:2022" -n 1
go run ./cmd/apps/scholarly openalex -q "large language models" -f "publication_year:2023" -n 1
```

### Finding Open Access Full-Text

```bash
# Find open access articles
go run ./cmd/apps/scholarly openalex -q "climate change mitigation" -f "is_oa:true" -n 5
```

## Troubleshooting

### Rate Limiting

Most academic APIs employ rate limiting. To avoid issues:

1. Always use the `mailto` parameter (`-m` flag) with your email address
2. Use reasonable request limits
3. Add delays between consecutive requests
4. Check error responses for rate limit indicators

### Query Syntax Issues

Different sources have different query syntax requirements:

- **arXiv**: Uses field prefixes (ti:, au:, etc.) and supports Boolean operators
- **OpenAlex**: Uses natural language queries with separate filter parameters
- **Crossref**: Similar to OpenAlex but with different filter parameter names

Reference: Parse the respective `client.go` files to understand how queries are constructed for each source.

### Common Errors

- **Empty Results**: Check your query syntax and filters
- **HTTP Errors**: May indicate rate limiting or connectivity issues
- **Parse Errors**: Check for any special characters in queries

When encountering issues, use the `-d` debug flag to get more detailed log output:

```bash
go run ./cmd/apps/scholarly arxiv -q "quantum computing" -n 5 -d
```

This will provide details on the API request, response, and any parsing issues.

## Conclusion

This guide provides a comprehensive approach to testing the scholarly search functionality across different sources and with various filters. By systematically working through these examples, you can verify that all aspects of the tool are functioning correctly and understand how to effectively use each feature.

For additional details on implementation, refer to the source code in:
- `/pkg/scholarly/arxiv/`
- `/pkg/scholarly/openalex/`
- `/pkg/scholarly/crossref/`

And the command definitions in:
- `/cmd/apps/scholarly/` 