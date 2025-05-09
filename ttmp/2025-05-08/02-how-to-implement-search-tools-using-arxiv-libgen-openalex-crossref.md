# Implementation Plan for Search Tools

## Overview
This document outlines how to implement the six scholarly search functions using our existing client packages.

## Function Implementation Details

### 1. `search_works`

#### Implementation Approach
- **OpenAlex Path**: `pkg/openalex/client.go`
  - Use existing `SearchWorks()` method, extending to accept filter parameters
  - Map limit to OpenAlex's `per_page` parameter
  - Pass search query directly to the API

- **Crossref Path**: `pkg/crossref/client.go`
  - Enhance `Search()` method to handle the work type specifically
  - Map limit parameter to Crossref's `rows` parameter
  - Convert filter object to Crossref filter syntax

#### Required Client Modifications
- Add response transformation to standardize return format between providers
- Implement common `Work` struct in `pkg/common/types.go`
- Create wrapper function that routes to appropriate client based on source parameter

### 2. `resolve_doi`

#### Implementation Approach
- **Primary Logic**: `pkg/common/doi_resolver.go` (new file)
  - Implement sequential calls to both OpenAlex and Crossref
  - Merge responses with appropriate field precedence

- **OpenAlex**: Use `GetWorkByDOI()` method
- **Crossref**: Extend `GetWork()` method for DOI resolution

#### Required Client Modifications
- Create field mapping utilities in `pkg/common/mapping.go`
- Add response merger that implements the field precedence rules

### 3. `suggest_keywords`

#### Implementation Approach
- **OpenAlex Path**: `pkg/openalex/concepts.go` (new file)
  - Implement `GetConceptsFromText()` method using the text-aboutness endpoint
  - Create response formatter to extract relevant concept information

#### Required Client Modifications
- Add concept-related types to `pkg/openalex/types.go`
- Implement text-aboutness POST request in the OpenAlex client

### 4. `get_metrics`

#### Implementation Approach
- **OpenAlex Path**: `pkg/openalex/metrics.go` (new file)
  - Create dedicated `GetMetrics()` method that extracts metrics from work objects
  - Use existing `GetWork()` method and transform response

#### Required Client Modifications
- Add metrics extraction utilities to parse OpenAlex work responses
- Create metrics-specific response struct

### 5. `get_citations`

#### Implementation Approach
- **OpenAlex Path**: `pkg/openalex/citations.go` (new file)
  - For outgoing citations: Extract from `referenced_works` in work response
  - For incoming citations: Implement filter-based search with `cites:` parameter
  - Handle pagination using cursor tokens

#### Required Client Modifications
- Add pagination support to OpenAlex client
- Create citation-specific response formatter

### 6. `find_full_text`

#### Implementation Approach
- **Orchestration**: `pkg/common/full_text_finder.go` (new file)
  - Implement cascading resolution strategy with fallbacks

- **Resolution Order**:
  1. **OpenAlex**: Extract from work locations, filtering by OA status and version
  2. **Unpaywall**: Add minimal Unpaywall client in `pkg/unpaywall/`
  3. **LibGen**: Use existing `pkg/libgen/client.go` search capabilities

#### Required Client Modifications
- Add location extraction to OpenAlex client
- Create minimal Unpaywall client for DOI resolution
- Enhance LibGen client to better handle title searches

## Common Infrastructure

### Error Handling Strategy
- Implement standardized error types in `pkg/common/errors.go`
- Use structured logging with zerolog for debugging
- Create meaningful user-facing error messages

### Rate Limiting Considerations
- Implement backoff mechanisms for all API clients
- Add configurable rate limits per provider
- Cache responses where appropriate

### Testing Strategy
- Create mock responses for each API
- Implement unit tests for each function
- Add integration tests for end-to-end verification

## Implementation Phases

1. **Phase 1**: Create common types and interfaces
2. **Phase 2**: Enhance existing clients with required methods
3. **Phase 3**: Implement the six core functions
4. **Phase 4**: Add error handling and rate limiting
5. **Phase 5**: Comprehensive testing

## Cobra Command Integration

Each function will be exposed as a Cobra command in the CLI:

```go
// Example for search_works command in cmd/search.go
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search works across providers",
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags and call the search_works function
	},
}

func init() {
	searchCmd.Flags().StringP("source", "s", "openalex", "Data source (openalex, crossref)")
	searchCmd.Flags().IntP("limit", "l", 20, "Maximum results to return")
	rootCmd.AddCommand(searchCmd)
}
```