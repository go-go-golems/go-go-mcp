## Report: Client Modifications for Query DSL Support

**Date:** 2025-05-09
**Author:** Gemini AI Assistant

This document outlines the necessary modifications to the existing scholarly search clients (`arxiv`, `crossref`, `openalex`) to fully support the capabilities of the newly implemented `querydsl` and the unified `scholarly.Search` function.

### I. Overview of DSL and Unified Search

The `querydsl` package introduces a `Query` struct with the following fields intended for cross-provider searching:

*   `Text` (free-text/phrase)
*   `Author` (family name/full string)
*   `Title` (words/phrase)
*   `Category` (arXiv primary category, e.g., "cs.AI")
*   `WorkType` (Crossref/OpenAlex type, e.g., "journal-article")
*   `FromYear` (inclusive YYYY)
*   `ToYear` (inclusive YYYY)
*   `OpenAccess` (*bool: true for OA only, false for closed only, nil to ignore)
*   `Sort` (SortRelevance, SortNewest, SortOldest)
*   `MaxResults` (integer)

The `scholarly.Search` function translates a `querydsl.Query` into provider-specific parameters for each client's `Search(common.SearchParams)` method. The `common.SearchParams` struct currently has:

*   `Query` (string)
*   `MaxResults` (int)
*   `Filters` (map[string]string)

### II. Required Client Modifications

To fully leverage the DSL, the individual client `Search` methods and their parameter handling logic need to be updated to interpret the fields passed via `common.SearchParams` (which are derived from the DSL's `ToProvider()` methods).

**A. ArXiv Client (`pkg/scholarly/arxiv/client.go`)**

The current `arxivClient.Search` method primarily uses `params.Query` for the `search_query` and `params.MaxResults`.

*   **Sorting:**
    *   The DSL supports `SortRelevance`, `SortNewest`, `SortOldest`.
    *   The `querydsl.Query.ToArxiv()` method does **not** currently translate these into arXiv-specific sort parameters (e.g., `sortBy=submittedDate`, `sortOrder=descending`).
    *   **Action Needed:** Modify `querydsl.Query.ToArxiv()` to append sort parameters to the query string based on `q.Sort`. The `arxivClient.Search` will then inherently use these as they become part of the `search_query`.
        *   `SortRelevance` -> `sortBy=relevance` (default, already handled)
        *   `SortNewest` -> `sortBy=submittedDate&sortOrder=descending`
        *   `SortOldest` -> `sortBy=submittedDate&sortOrder=ascending`
*   **Category:**
    *   The `querydsl.Query.ToArxiv()` correctly adds `cat:<Category>` to the query string.
    *   **Action Needed:** No direct change to `arxivClient.Search` is needed for this, as it's part of the `params.Query`.
*   **Date Range:**
    *   `querydsl.Query.ToArxiv()` correctly adds `submittedDate:[YYYYMMDDHHMM+TO+YYYYMMDDHHMM]`.
    *   **Action Needed:** No direct change to `arxivClient.Search`.
*   **Author/Title/Text:**
    *   `querydsl.Query.ToArxiv()` correctly formats these with prefixes (`au:`, `ti:`, `all:`).
    *   **Action Needed:** No direct change to `arxivClient.Search`.
*   **OpenAccess:**
    *   ArXiv is inherently Open Access. The DSL's `OpenAccess` flag is not directly applicable as a filter for ArXiv in the same way as for other providers.
    *   **Action Needed:** No change required in `arxivClient` for this. The `querydsl` correctly omits OA filtering for ArXiv.
*   **WorkType:**
    *   ArXiv does not have a direct equivalent filter for general "WorkType" like "journal-article". It deals with pre-prints and categories.
    *   **Action Needed:** No change required. The `querydsl` correctly omits this for ArXiv.

**B. Crossref Client (`pkg/scholarly/crossref/client.go`)**

The current `crossrefClient.Search` uses `params.Query` for the main query, `params.MaxResults` for `rows`, and `params.Filters["filter"]` for the filter string.

*   **Sorting:**
    *   `querydsl.Query.ToCrossref()` correctly sets `sort` and `order` parameters in the returned `url.Values`.
    *   The `scholarly.Search` function **does not** currently pass these sort parameters from the `url.Values` into `common.SearchParams.Filters` in a way the `crossrefClient` would pick them up directly. The `crossrefClient` currently has no logic to extract sort parameters from `params.Filters`.
    *   **Action Needed:** Modify `crossrefClient.Search` to extract `sort` and `order` from `params.Filters` if they exist, and add them to `apiParams`. Alternatively, modify `scholarly.Search` to more intelligently populate `common.SearchParams` (e.g., add dedicated fields for sort if it becomes a common pattern, or ensure the `crossrefClient` can parse the `filter` string for these, though the `ToCrossref()` method already separates them).
        *   A cleaner approach would be for `scholarly.Search` to pass `sort` and `order` into `params.Filters` with specific keys (e.g., `params.Filters["crossref_sort"]`, `params.Filters["crossref_order"]`) and have `crossrefClient.Search` look for these.
*   **Author/Title/Text:**
    *   `querydsl.Query.ToCrossref()` sets `query`, `query.author`, and `query.title`.
    *   The `scholarly.Search` function currently only passes `values.Get("query")` (which is `q.Text`) into `params.Query`.
    *   **Action Needed:** Modify `scholarly.Search` when handling `ProviderCrossref` to correctly populate `params.Query` (for main text) and potentially add `query.author` and `query.title` to `params.Filters` if the client is to remain generic. *Alternatively, and preferably for clarity,* modify `crossrefClient.Search` to accept more structured input than just a single query string and a filter bag, or have `scholarly.Search` construct a more complex `params.Query` string if CrossRef supports combining these (e.g., `params.Query = "main_text author:Smith title:Something"` if that's a valid Crossref query pattern). The current `ToCrossref()` puts them as distinct URL parameters which is standard.
    *   The simplest fix aligning with current `ToCrossref()` is to modify `crossrefClient.Search` to check `params.Filters` for keys like `query.author` and `query.title` and add them to `apiParams` if present. `scholarly.Search` would need to be updated to populate these into `params.Filters` from `query.ToCrossref()` values.
*   **WorkType, Date Range, OpenAccess:**
    *   `querydsl.Query.ToCrossref()` correctly adds these to the `filter` string in `url.Values`.
    *   `scholarly.Search` correctly passes this `filter` string into `params.Filters["filter"]`.
    *   The `crossrefClient.Search` correctly uses `params.Filters["filter"]`.
    *   **Action Needed:** No change required for these specific fields' path.

**C. OpenAlex Client (`pkg/scholarly/openalex/client.go`)**

The current `openalexClient.Search` uses `params.Query` for `search`, `params.MaxResults` for `per_page`, `params.Filters["filter"]` for the filter string, and `params.Filters["sort"]` for sorting.

*   **Sorting:**
    *   `querydsl.Query.ToOpenAlex()` correctly sets the `sort` parameter in `url.Values`.
    *   `scholarly.Search` correctly extracts this `sort` value and puts it into `params.Filters["sort"]`.
    *   `openalexClient.Search` correctly uses `params.Filters["sort"]`.
    *   **Action Needed:** No change required for sorting.
*   **Author/Title/Text:**
    *   `querydsl.Query.ToOpenAlex()` sets the `search` parameter for free text and adds `author.search` and `title.search` to the `filter` string.
    *   `scholarly.Search` passes `values.Get("search")` (which is `q.Text`) to `params.Query` and `values.Get("filter")` (which contains author and title filters) to `params.Filters["filter"]`.
    *   `openalexClient.Search` uses `params.Query` for the `search` API parameter and `params.Filters["filter"]` for the `filter` API parameter.
    *   **Action Needed:** No change required. The current mapping appears correct.
*   **WorkType, Date Range, OpenAccess:**
    *   `querydsl.Query.ToOpenAlex()` correctly adds these to the `filter` string in `url.Values`.
    *   `scholarly.Search` correctly passes this `filter` string into `params.Filters["filter"]`.
    *   `openalexClient.Search` correctly uses `params.Filters["filter"]`.
    *   **Action Needed:** No change required for these fields.
*   **Category (e.g. `cs.AI`):**
    *   The `querydsl.Query` has a `Category` field, primarily for ArXiv.
    *   `querydsl.ToOpenAlex()` does **not** map this field, as OpenAlex uses a different concept system (e.g., `concepts.id`, `concepts.wikidata`).
    *   **Action Needed:** No direct change to `openalexClient.Search` for the ArXiv-style `Category`. If fine-grained concept/subject search is desired for OpenAlex via the DSL in the future, the `querydsl.Query` struct and `ToOpenAlex()` method would need expansion, and `openalexClient` would need to process these, likely via the `filter` parameter.

### III. Summary of Key Actions:

1.  **`querydsl/querydsl.go`:**
    *   Enhance `ToArxiv()` to include `sortBy` and `sortOrder` parameters based on `q.Sort`.

2.  **`scholarly/search.go`:**
    *   **Crossref Author/Title:** Ensure `query.author` and `query.title` from `query.ToCrossref()` are passed to `crossrefClient.Search`. This might involve adding them to `params.Filters` (e.g., `params.Filters["crossref_query.author"] = values.Get("query.author")`) or a more structured `SearchParams` if preferred.
    *   **Crossref Sorting:** Ensure `sort` and `order` from `query.ToCrossref()` are passed to `crossrefClient.Search`, likely via `params.Filters` (e.g., `params.Filters["crossref_sort"] = values.Get("sort")`, `params.Filters["crossref_order"] = values.Get("order")`).

3.  **`pkg/scholarly/crossref/client.go`:**
    *   Modify `Search` method to look for and use the author, title, sort, and order parameters from `params.Filters` (e.g., `params.Filters["crossref_query.author"]`, `params.Filters["crossref_sort"]`) and add them to the `apiParams` sent to Crossref.

### IV. General Recommendations:

*   **Consistency in `common.SearchParams`:** While `Filters` map[string]string is flexible, for common parameters like sorting or specific query fields (author, title) that are not part of the main text query, consider if `common.SearchParams` should be augmented with more specific fields if this pattern repeats. However, for now, using uniquely prefixed keys within `Filters` (e.g., `crossref_sort`, `openalex_sort`) is a viable approach for provider-specific needs not covered by the most common denominator.
*   **Client Responsibility:** The individual clients should be responsible for interpreting the `common.SearchParams` (including its `Filters` bag) and constructing the correct API request for their specific service. The `scholarly.Search` function's role is primarily to orchestrate and translate the high-level DSL query into this common structure.

By implementing these changes, the clients will be better equipped to handle the richer queries enabled by the `querydsl`, leading to more precise and powerful federated searching. 