# Scientific Research Clients in Go-Go-MCP

This document provides a detailed overview of the scientific research APIs available in the go-go-mcp package, focusing on the scholarly module. The scholarly module enables searching across multiple academic databases and repositories.

## Overview

The scholarly package provides a unified interface to search for academic papers and articles across multiple sources:

- **arXiv**: A repository of electronic pre-prints for scientific papers
- **Crossref**: A comprehensive scholarly metadata database
- **OpenAlex**: An open access research database

The system is designed with a layered architecture:
1. **Client Layer**: Specific API clients for each source (arXiv, Crossref, OpenAlex)
2. **Integration Layer**: Common interfaces and models to standardize data across sources
3. **Tool Layer**: MCP tools that expose the functionality to external applications

## Common Search Interface: SearchParams

At the heart of the API transformation is the `common.SearchParams` struct, which serves as a universal interface between the user-facing API and the different client implementations:

```go
// SearchParams contains common search parameters
type SearchParams struct {
    Query      string               // The search query text
    MaxResults int                  // Maximum number of results to return
    Filters    map[string]string    // Map of filter name to filter value
    Sort       string               // Sort order for results
    EmailAddr  string               // Email address for API providers that require it
}
```

This intermediate structure serves several important functions:

1. **Abstraction**: It separates the user-facing API (`SearchWorksRequest`) from the specific API implementations
2. **Standardization**: It provides a consistent interface across all clients
3. **Transformation**: It allows each client to interpret and adapt the common parameters to their specific API needs

### How SearchParams is Used in the Transformation Flow

1. **User Request → SearchParams**: 
   ```go
   // User request
   req := SearchWorksRequest{
       Query: "quantum computing",
       Source: "arxiv",
       Limit: 10,
       Filter: map[string]string{"category": "cs.AI"}
   }

   // Converted to SearchParams
   params := common.SearchParams{
       Query:      req.Query,             // "quantum computing"
       MaxResults: req.Limit,        // 10
       Filters:    req.Filter,       // {"category": "cs.AI"}
       Sort:       "",               // Not specified in this example
       EmailAddr:  "",               // Not specified in this example
   }
   ```

2. **SearchParams → API-Specific Request**:
   
   For arXiv:
   ```go
   // Convert SearchParams to arXiv API parameters
   apiParams := url.Values{}
   apiParams.Add("search_query", params.Query)        // "all:quantum computing AND cat:cs.AI"
   apiParams.Add("max_results", fmt.Sprintf("%d", params.MaxResults))  // "10"
   apiParams.Add("sortBy", "relevance")
   ```

   For OpenAlex:
   ```go
   // Convert SearchParams to OpenAlex API parameters
   apiParams := url.Values{}
   apiParams.Add("search", params.Query)              // "quantum computing"
   apiParams.Add("per_page", fmt.Sprintf("%d", params.MaxResults))  // "10"
   
   if filter, ok := params.Filters["filter"]; ok {
       apiParams.Add("filter", filter)                // Special handling
   }
   ```

3. **API Response → Common SearchResult**:
   
   All API responses are converted to the standardized `SearchResult` type:
   ```go
   type SearchResult struct {
       Title       string
       Authors     []string
       Abstract    string
       Published   string
       DOI         string
       PDFURL      string
       SourceURL   string
       SourceName  string
       OAStatus    string
       License     string
       FileSize    string
       Citations   int
       Type        string
       JournalInfo string
       Metadata    map[string]interface{} // Additional source-specific data
   }
   ```

### SearchParams vs. SearchWorksRequest

It's important to understand the distinction between:

1. **SearchWorksRequest** - The user-facing API call structure (external interface)
2. **SearchParams** - The internal common interface for all clients

While they appear similar, this separation allows the public API to evolve independently of the internal client implementations.

## Available API Endpoints

### 1. SearchWorks

The primary endpoint for searching academic works across different sources.

**Function**: `scholarly.SearchWorks(req SearchWorksRequest) (*SearchWorksResponse, error)`

**Request Parameters**:

```go
type SearchWorksRequest struct {
    Query  string            // The search query text
    Source string            // Source database: "arxiv", "crossref", or "openalex"
    Limit  int               // Maximum number of results to return (default: 20)
    Filter map[string]string // Source-specific filters
}
```

**Response Structure**:

```go
type Work struct {
    ID            string   // Identifier in the source database
    DOI           string   // Digital Object Identifier
    Title         string   // Paper title
    Authors       []string // List of author names
    Year          int      // Publication year
    IsOA          bool     // Is open access
    CitationCount int      // Number of citations
    Abstract      string   // Paper abstract
    SourceName    string   // Source database name
    PDFURL        string   // URL to PDF if available
}
```

## Client Implementations and API Transformations

Each client implements a common interface but handles the specific details of each API. Here's how the Go functions transform into API calls:

### 1. arXiv Client

**Client Function Signature**:
```go
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error)
```

**Parameters Transformation**:
```go
// Example transformation from SearchWorksRequest to arXiv API call
func searchArxiv(req SearchWorksRequest) (*SearchWorksResponse, error) {
    client := arxiv.NewClient()

    params := common.SearchParams{
        Query:      req.Query,
        MaxResults: req.Limit,
    }

    // Convert filter map to appropriate arxiv query parameters
    if len(req.Filter) > 0 {
        filterQueries := []string{}
        for k, v := range req.Filter {
            switch k {
            case "category":
                filterQueries = append(filterQueries, fmt.Sprintf("cat:%s", v))
            case "author":
                filterQueries = append(filterQueries, fmt.Sprintf("au:\"%s\"", v))
            case "title":
                filterQueries = append(filterQueries, fmt.Sprintf("ti:\"%s\"", v))
            }
        }

        // Combine with main query using AND
        if len(filterQueries) > 0 {
            if !strings.Contains(params.Query, ":") {
                params.Query = "all:" + params.Query
            }
            params.Query = fmt.Sprintf("%s AND %s", params.Query, strings.Join(filterQueries, " AND "))
        }
    }

    results, err := client.Search(params)
    // Process results...
}
```

**HTTP Request Formation**:
```go
// Inside arXiv client:
apiParams := url.Values{}
apiParams.Add("search_query", params.Query)
apiParams.Add("max_results", fmt.Sprintf("%d", params.MaxResults))
apiParams.Add("sortBy", "relevance")

apiURL := c.BaseURL + "?" + apiParams.Encode()
req, err := http.NewRequest("GET", apiURL, nil)
```

### 2. Crossref Client

**Client Function Signature**:
```go
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error)
```

**Parameters Transformation**:
```go
// Example transformation from SearchWorksRequest to Crossref API call
func searchCrossref(req SearchWorksRequest) (*SearchWorksResponse, error) {
    client := crossref.NewClient("") // No mailto for now

    params := common.SearchParams{
        Query:      req.Query,
        MaxResults: req.Limit,
    }

    // Apply filters if present
    if len(req.Filter) > 0 {
        params.Filters = req.Filter
    }

    results, err := client.Search(params)
    // Process results...
}
```

**HTTP Request Formation**:
```go
// Inside Crossref client:
apiParams := url.Values{}
apiParams.Add("query", params.Query)
apiParams.Add("rows", fmt.Sprintf("%d", params.MaxResults))

if c.Mailto != "" {
    apiParams.Add("mailto", c.Mailto)
}

if filter, ok := params.Filters["filter"]; ok {
    apiParams.Add("filter", filter)
}

apiParams.Add("select", "DOI,title,author,publisher,type,created,issued,URL,abstract,ISSN,ISBN,subject,link")

apiURL := c.BaseURL + "?" + apiParams.Encode()
req, err := http.NewRequest("GET", apiURL, nil)
```

### 3. OpenAlex Client

**Client Function Signature**:
```go
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error)
```

**Parameters Transformation**:
```go
// Example transformation from SearchWorksRequest to OpenAlex API call
func searchOpenAlex(req SearchWorksRequest) (*SearchWorksResponse, error) {
    client := openalex.NewClient("") // No mailto for now

    params := common.SearchParams{
        Query:      req.Query,
        MaxResults: req.Limit,
    }

    // Apply filters if present
    if len(req.Filter) > 0 {
        params.Filters = req.Filter
    }

    results, err := client.Search(params)
    // Process results...
}
```

**HTTP Request Formation**:
```go
// Inside OpenAlex client:
apiParams := url.Values{}

if params.Query != "" {
    apiParams.Add("search", params.Query)
}

apiParams.Add("per_page", fmt.Sprintf("%d", params.MaxResults))

if c.Mailto != "" {
    apiParams.Add("mailto", c.Mailto)
}

if filter, ok := params.Filters["filter"]; ok {
    apiParams.Add("filter", filter)
}

if sort, ok := params.Filters["sort"]; ok {
    apiParams.Add("sort", sort)
} else {
    apiParams.Add("sort", "relevance_score:desc")
}

apiParams.Add("select", "id,doi,title,display_name,publication_year,publication_date,cited_by_count,authorships,primary_location,open_access,type,concepts,abstract_inverted_index,relevance_score,referenced_works,related_works")

apiURL := c.BaseURL + "?" + apiParams.Encode()
req, err := http.NewRequest("GET", apiURL, nil)
```

## Common Data Transformation Flow

The overall transformation flow from user request to API response follows this pattern:

1. **User Request**: The user passes a `SearchWorksRequest` to `scholarly.SearchWorks()`
2. **Source Routing**: Based on the `Source` field, the request is routed to the appropriate client function (`searchArxiv`, `searchCrossref`, or `searchOpenAlex`)
3. **Parameter Transformation**: The source-specific function transforms the common request into appropriate parameters for that API
4. **API Call**: The client makes the HTTP request to the external API
5. **Response Transformation**: The source-specific response is converted to a common `SearchWorksResponse` format
6. **Result Return**: The standardized response is returned to the user

This abstraction allows clients to work with a consistent interface while handling the nuances of each external API behind the scenes.

## Source-Specific Query Parameters

### 1. arXiv Client

**Base Endpoint**: `http://export.arxiv.org/api/query`

**Supported Filters**:
- `category`: Search by arXiv category (e.g., "cs.AI", "physics.gen-ph")
- `author`: Filter by author name
- `title`: Search within paper titles

**Example Filter Usage**:
```
"category": "cs.AI"
"author": "John Smith"
"title": "quantum computing"
```

**Implementation Details**:
- Uses the Atom feed format
- Provides full abstracts and PDF links
- All content is open access

### 2. Crossref Client

**Base Endpoint**: `https://api.crossref.org/works`

**Supported Filters**:
- `filter`: Raw Crossref filter query string
- `type`: Filter by publication type (e.g., "journal-article")
- `from-pub-date`: Filter by publication date (YYYY-MM-DD)
- `until-pub-date`: Filter by publication date (YYYY-MM-DD)

**Implementation Details**:
- Returns detailed publication metadata
- Provides DOIs for most entries
- May include citation counts

### 3. OpenAlex Client

**Base Endpoint**: `https://api.openalex.org/works`

**Supported Filters**:
- `filter`: Raw OpenAlex filter query string
- `sort`: Sort order (default: "relevance_score:desc")
- `is_oa`: Filter for open access works (true/false)
- `publication_year`: Filter by publication year
- `has_doi`: Filter for works with DOIs (true/false)

**Implementation Details**:
- Provides rich metadata including citation counts
- Includes open access status and links
- Returns abstracts for most papers
- Has citation and reference data

## Example Usage

### Searching for Papers on arXiv

```go
req := scholarly.SearchWorksRequest{
    Query:  "quantum computing",
    Source: "arxiv",
    Limit:  10,
    Filter: map[string]string{
        "category": "cs.AI",
    },
}

response, err := scholarly.SearchWorks(req)
if err != nil {
    // Handle error
}

// Process results
for _, work := range response.Works {
    fmt.Printf("Title: %s\n", work.Title)
    fmt.Printf("Authors: %s\n", strings.Join(work.Authors, ", "))
    fmt.Printf("PDF URL: %s\n", work.PDFURL)
}
```

### Searching for Papers on OpenAlex

```go
req := scholarly.SearchWorksRequest{
    Query:  "climate change",
    Source: "openalex",
    Limit:  20,
    Filter: map[string]string{
        "is_oa": "true",
        "publication_year": "2023",
    },
}

response, err := scholarly.SearchWorks(req)
// Process results...
```

## Integration with MCP

The scholarly module is integrated with the Model Context Protocol through the function:

```go
RegisterScholarlyTools(registry *tool_registry.Registry) error
```

This makes the scholarly search functionality available as an MCP tool that can be called by AI models or other applications.

## Detailed Filter Support by Client

Each client interprets filters differently, based on the capabilities and conventions of the underlying API. Below is a comprehensive guide to the filters supported by each client.

### arXiv Client Filters

The arXiv client transforms filter parameters into special arXiv-specific query syntax.

**Supported Filter Keys**:

| Filter Key | Description | Example Value | Transformation |
|------------|-------------|---------------|----------------|
| `category` | arXiv subject category | `"cs.AI"` | `cat:cs.AI` |
| `author` | Author name | `"John Smith"` | `au:"John Smith"` |
| `title` | Words in title | `"neural networks"` | `ti:"neural networks"` |

**Filter Transformation Example**:
```go
// Original SearchParams.Filters:
// {"category": "cs.AI", "author": "Hinton"}

// Transformed into arXiv query syntax:
// "all:quantum computing AND cat:cs.AI AND au:\"Hinton\""
```

**Important Notes**:
1. Multiple filters are combined with `AND` operators
2. arXiv prefixes query terms with field identifiers:
   - `cat:` - for category
   - `au:` - for author
   - `ti:` - for title
   - `all:` - for all fields (applied to the main query)
3. The main Query parameter is prefixed with `all:` if it doesn't already have a field prefix

### Crossref Client Filters

The Crossref client passes most filters directly to the Crossref API.

**Supported Filter Keys**:

| Filter Key | Description | Example Value | API Parameter |
|------------|-------------|---------------|---------------|
| `filter` | Raw Crossref filter string | `"type:journal-article,from-pub-date:2020-01-01"` | `filter` |
| `type` | Work type filter | `"journal-article"` | Added to `filter` param |
| `from-pub-date` | Publication date lower bound | `"2020-01-01"` | Added to `filter` param |
| `until-pub-date` | Publication date upper bound | `"2022-12-31"` | Added to `filter` param |

**Filter Transformation Example**:
```go
// Original SearchParams.Filters:
// {"type": "journal-article", "from-pub-date": "2020-01-01"}

// Transformed and sent as Crossref API parameter:
// ?filter=type:journal-article,from-pub-date:2020-01-01
```

**Important Notes**:
1. Crossref has a rich filtering syntax that can be passed directly via the `filter` key
2. The client extracts certain common filters and formats them into Crossref's filter syntax
3. Special parameters like `select` are added by the client to limit response size

### OpenAlex Client Filters

The OpenAlex client handles filters through both dedicated parameters and a versatile filter syntax.

**Supported Filter Keys**:

| Filter Key | Description | Example Value | API Parameter |
|------------|-------------|---------------|---------------|
| `filter` | Raw OpenAlex filter string | `"publication_year:2022"` | `filter` |
| `sort` | Sort order | `"relevance_score:desc"` | `sort` |
| `is_oa` | Open Access filter | `"true"` | Added to `filter` param |
| `publication_year` | Publication year | `"2022"` | Added to `filter` param |
| `has_doi` | Filter by DOI presence | `"true"` | Added to `filter` param |

**Filter Transformation Example**:
```go
// Original SearchParams.Filters:
// {"is_oa": "true", "publication_year": "2022", "sort": "cited_by_count:desc"}

// Transformed and sent as OpenAlex API parameters:
// ?filter=is_oa:true,publication_year:2022&sort=cited_by_count:desc
```

**Important Notes**:
1. OpenAlex provides both a `filter` parameter for complex filtering and dedicated parameters for common filters
2. The `sort` parameter works with various fields and directions (`:asc` or `:desc`)
3. The client has special handling for certain filter keys like `sort` that are not added to the filter string

### Filter Best Practices

When working with the scholarly search clients, consider these best practices:

1. **Source-Specific Filters**: Use source-specific filters when targeting a particular database
   ```go
   // For arXiv
   Filter: map[string]string{"category": "cs.AI"}
   
   // For OpenAlex
   Filter: map[string]string{"is_oa": "true", "publication_year": "2022"}
   ```

2. **Generic Searches**: For more generic searches, use simpler queries without filters
   ```go
   // Works well across all sources
   req := scholarly.SearchWorksRequest{
       Query: "quantum computing applications",
       Source: "arxiv",  // or "crossref" or "openalex"
       Limit: 20,
   }
   ```

3. **Filter Combinations**: Combine multiple filters to narrow results
   ```go
   // Find recent open access papers on climate change
   Filter: map[string]string{
       "is_oa": "true",
       "publication_year": "2023",
       "sort": "relevance_score:desc",
   }
   ``` 