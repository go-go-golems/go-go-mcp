# Scholarly API Documentation

The Scholarly API provides programmatic access to search for academic papers across multiple sources including ArXiv, Crossref, and OpenAlex.

## Base URL

```
http://localhost:8080
```

## Endpoints

### Search Papers

Search for scholarly papers across multiple sources.

```
GET /api/search
```

#### Query Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|--------|
| query | string | Main search query text | |
| sources | string | Comma-separated list of sources to search from (arxiv,crossref,openalex,all) | all |
| limit | integer | Maximum number of results to return | 20 |
| author | string | Author name to search for | |
| title | string | Title words/phrase to search for | |
| category | string | ArXiv category (e.g., cs.AI) | |
| work-type | string | Publication type (e.g., journal-article) | |
| from-year | integer | Starting year (inclusive) | 0 |
| to-year | integer | Ending year (inclusive) | 0 |
| sort | string | Sort order: relevance, newest, oldest | relevance |
| open-access | string | Open access filter: true, false | |
| mailto | string | Email address for OpenAlex polite pool | wesen@ruinwesen.com |
| disable-rerank | boolean | Disable reranking of search results | false |

#### Example Request

```
GET /api/search?query=quantum+computing&sources=arxiv&limit=5&sort=newest
```

#### Response

```json
{
  "results": [
    {
      "id": "http://arxiv.org/abs/2505.02154v1",
      "doi": "10.48550/arXiv.2505.02154",
      "title": "Quantum Computing Applications in Material Science",
      "authors": ["Jane Smith", "John Doe"],
      "year": 2025,
      "is_oa": true,
      "citation_count": 0,
      "abstract": "This paper explores quantum computing applications...",
      "source_name": "arxiv",
      "pdf_url": "http://arxiv.org/pdf/2505.02154v1",
      "reranked": true,
      "reranker_score": 9.45,
      "original_index": 3
    },
    // Additional results...
  ],
  "query": "quantum computing",
  "count": 5,
  "sources": ["arxiv"]
}
```

### List Available Sources

Returns a list of available sources for searching.

```
GET /api/sources
```

#### Response

```json
{
  "sources": ["arxiv", "crossref", "openalex", "all"]
}
```

### Health Check

Check the health status of the API.

```
GET /api/health
```

#### Response

```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2025-05-10T14:30:15Z"
}
```

## Error Responses

In case of errors, the API returns appropriate HTTP status codes along with error messages in JSON format:

```json
{
  "error": "Error message describing what went wrong"
}
```

Common error codes:

- **400 Bad Request**: Invalid parameters
- **500 Internal Server Error**: Server-side error

## Examples

### Example: Searching for quantum computing papers in ArXiv

```
GET /api/search?query=quantum+computing&sources=arxiv&limit=5&sort=newest
```

### Example: Searching for papers by a specific author

```
GET /api/search?author=Hinton&sources=all&limit=10
```

### Example: Searching for open access papers published after 2020

```
GET /api/search?query=machine+learning&from-year=2020&open-access=true
```