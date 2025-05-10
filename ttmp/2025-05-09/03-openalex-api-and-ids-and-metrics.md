# Understanding OpenAlex APIs, Identifiers, and Metrics

## 1. Introduction to OpenAlex and its API

OpenAlex is a free and open scientific knowledge graph (SKG) that provides access to metadata for a vast collection of scholarly entities. It aims to create a comprehensive, interconnected database of works (journal articles, books, datasets, etc.), authors, institutions, venues (journals, repositories), concepts, and funders. Launched as a successor to the Microsoft Academic Graph (MAG), OpenAlex offers its data via a REST API, a full data dump (snapshots), and a web-based interface.

Its key entities include:
*   **Works:** Scholarly documents like articles, books, datasets, etc.
*   **Authors:** Individuals who have created works.
*   **Venues:** Places where works are hosted (e.g., journals, conference proceedings, repositories).
*   **Institutions:** Organizations affiliated with authors (e.g., universities, research labs).
*   **Concepts:** Topics or fields of study tagged to works.
*   **Funders:** Organizations that have funded the research.

## 2. Key OpenAlex API Endpoints for Works

The primary base URL for the OpenAlex API is `https://api.openalex.org`.

*   **Get a single work:**
    *   This is used for direct lookup when you have a unique identifier for the work.
    *   **By OpenAlex ID:**
        *   Endpoint: `/works/{OPENALEX_ID}`
        *   Example: `https://api.openalex.org/works/W2741809807`
    *   **By DOI:**
        *   Endpoint: `/works/https://doi.org/{YOUR_DOI_HERE}` (note the full DOI URL is used as the identifier)
        *   Example: `https://api.openalex.org/works/https://doi.org/10.7717/peerj.4375`
    *   **By PubMed ID (PMID):**
        *   Endpoint: `/works/pmid:{YOUR_PMID_HERE}`
        *   Example: `https://api.openalex.org/works/pmid:14907713`
    *   **By MAG ID:**
        *   Endpoint: `/works/mag:{YOUR_MAG_ID_HERE}`


*   **Search/Filter works:**
    *   This is more flexible for discovery, finding works based on criteria, or when you don't have a direct unique ID.
    *   Endpoint: `/works`
    *   Key Parameters:
        *   `search={query}`: For keyword-based searching across metadata fields like title, abstract, and full text (if available).
            *   Example: `https://api.openalex.org/works?search=nanotechnology+cancer+therapy`
        *   `filter={attribute}:{value}`: For precise filtering based on specific attributes. Multiple filters can be combined.
            *   Example (works by DOI): `https://api.openalex.org/works?filter=doi:10.1038/nphys1170`
            *   Example (works in a specific year): `https://api.openalex.org/works?filter=publication_year:2020`
            *   Example (Open Access articles): `https://api.openalex.org/works?filter=is_oa:true`
        *   `select={field1},{field2}`: To choose specific fields from the Work object to be returned in the response, reducing payload size.
            *   Example: `https://api.openalex.org/works?filter=doi:10.1038/nphys1170&select=id,doi,title,cited_by_count`
        *   `sort={attribute}:{asc|desc}`: To order the results.
            *   Example: `https://api.openalex.org/works?filter=authorships.institutions.id:I12345678&sort=publication_date:desc`
        *   `per_page={number}` & `page={number}`: For paginating through results. (Default `per_page` is 25, max is 200).
        *   `mailto={your_email}`: Polite way to identify yourself for higher rate limits if you have a registered email.

## 3. Understanding Scholarly Identifiers

*   **DOI (Digital Object Identifier):**
    *   **What it is:** A persistent, globally unique alphanumeric string assigned to a digital object. This can be a journal article, book, dataset, report, etc.
    *   **Managed by:** The International DOI Foundation (IDF) and various Registration Agencies (e.g., Crossref is a major RA for scholarly publications, DataCite for datasets).
    *   **Format:** Typically starts with `10.` followed by a prefix (identifying the registrant) and a suffix (chosen by the registrant). Example: `10.1038/nphys1170`. It can be resolved as a URL: `https://doi.org/10.1038/nphys1170`.
    *   **Purpose:** To provide a persistent link to the content's location on the internet and to uniquely identify the content itself. It's a cornerstone of scholarly communication.

*   **OpenAlex ID:**
    *   **What it is:** A unique identifier assigned by OpenAlex to each entity within its database (Works, Authors, Institutions, Venues, Concepts, Funders).
    *   **Format:** Starts with a letter indicating the entity type, followed by a sequence of numbers.
        *   `W` for Works (e.g., `W2741809807`)
        *   `A` for Authors (e.g., `A1969205032`)
        *   `I` for Institutions (e.g., `I13670809`)
        *   `V` for Venues (e.g., `V198377359`)
        *   `C` for Concepts (e.g., `C2778711962`)
    *   **Purpose:** To uniquely identify and disambiguate entities within the OpenAlex knowledge graph. This is crucial for linking entities together (e.g., connecting a work to its authors, venue, and concepts). It is the most direct way to retrieve a specific entity from OpenAlex if known.

*   **"Work ID" (in the context of `go-go-mcp`):**
    *   **Definition:** This is an application-specific term used within `go-go-mcp` (e.g., in the `metrics` command). It refers to an input string that the application expects to be *either* a DOI *or* an OpenAlex Work ID.
    *   **Usage:** The application's logic must parse this "Work ID" to determine its type.
        *   If it resembles a DOI (e.g., starts with "10." or is a `doi.org` URL), it should be treated as a DOI.
        *   If it resembles an OpenAlex Work ID (e.g., starts with "W" followed by numbers), it should be treated as such.
    This determination then dictates how the application queries OpenAlex (or other services like Crossref).

## 4. When to Use Which Identifier with OpenAlex

*   **Using a DOI with OpenAlex:**
    *   **Direct Lookup (preferred for single DOI):**
        *   Method: Use the specific work endpoint with the full DOI URL.
        *   API Call: `GET /works/https://doi.org/{YOUR_DOI}`
        *   Example: `https://api.openalex.org/works/https://doi.org/10.1038/nphys1170`
    *   **Filtering (useful for batch or combining with other filters):**
        *   Method: Use the `/works` endpoint with the `filter` parameter. The DOI value should be *just the DOI string*, not the full URL.
        *   API Call: `GET /works?filter=doi:{YOUR_DOI_STRING}`
        *   Example: `https://api.openalex.org/works?filter=doi:10.1038/nphys1170`
    *   **Best for:** When you have a DOI and want to find its corresponding entry and metadata in OpenAlex.

*   **Using an OpenAlex ID with OpenAlex:**
    *   **Direct Lookup (most efficient):**
        *   Method: Use the specific work endpoint with the OpenAlex Work ID.
        *   API Call: `GET /works/{OPENALEX_WORK_ID}`
        *   Example: `https://api.openalex.org/works/W2741809807`
    *   **Filtering (less common for single ID but possible):**
        *   Method: Use the `/works` endpoint with the `filter` parameter.
        *   API Call: `GET /works?filter=openalex:{OPENALEX_WORK_ID}` or `filter=ids.openalex:{OPENALEX_WORK_ID}`
        *   Example: `https://api.openalex.org/works?filter=openalex:W2741809807`
    *   **Best for:** When you already have the OpenAlex Work ID. This is the most direct and unambiguous way to retrieve a work from OpenAlex.

*   **Using the generic "Work ID" (application-level in `go-go-mcp`):**
    1.  **Parse the input:** Determine if the string is likely a DOI or an OpenAlex ID based on its format.
    2.  **If identified as a DOI:**
        *   Query OpenAlex using `GET /works/https://doi.org/{DOI}` or `GET /works?filter=doi:{DOI_STRING}`.
        *   Optionally, if more comprehensive metadata or cross-validation is needed, query other services like Crossref API (`https://api.crossref.org/works/{DOI}`).
    3.  **If identified as an OpenAlex Work ID:**
        *   Query OpenAlex using `GET /works/{OPENALEX_WORK_ID}`.

## 5. Getting Metrics for a Scholarly Work via OpenAlex

OpenAlex provides a wealth of metadata that directly includes or can be used to derive various metrics for a scholarly work.

*   **What metrics are available directly from the OpenAlex `Work` object?**
    *   `id`: The OpenAlex ID of the work.
    *   `doi`: The DOI of the work.
    *   `title`: Title of the work.
    *   `publication_year`: Year of publication.
    *   `publication_date`: Full date of publication.
    *   `cited_by_count`: The number of times this work has been cited by other works *within the OpenAlex database*. This is a key citation metric.
    *   `referenced_works`: A list of OpenAlex IDs of works that this work cites. The length of this list gives the reference count.
    *   `related_works`: A list of OpenAlex IDs of works that OpenAlex has identified as related.
    *   `authorships`: An array detailing authors, their affiliations, and countries.
    *   `counts_by_year`: An array showing how many citations the work received in specific years: `[{year: YYYY, cited_by_count: N}, ...]`.
    *   `open_access`: An object detailing OA status:
        *   `is_oa`: Boolean, true if any OA version is known.
        *   `oa_status`: e.g., "gold", "green", "hybrid", "bronze", "closed".
        *   `oa_url`: A URL to an OA version, if available.
        *   `any_repository_has_fulltext`: Boolean.
    *   `primary_location`: Information about the primary place the work is hosted (e.g., journal, repository), including its source ID, type, license, and version.
    *   `type`: The type of work (e.g., "journal-article", "book-chapter", "dataset").
    *   `concepts`: A list of concepts associated with the work, with scores.
    *   `grants`: A list of grants associated with the work.

*   **How to fetch metrics using OpenAlex API:**
    1.  **Identify the Work:** Obtain the OpenAlex `Work` object using one of the methods described in Section 4 (using its DOI or OpenAlex ID). It's good practice to use the `select` parameter to fetch only the fields relevant to your metrics to optimize the API call.
        *   Example API call (fetching a work by DOI and selecting specific metric-related fields):
            ```
            https://api.openalex.org/works/https://doi.org/10.1038/nphys1170?select=id,doi,title,cited_by_count,referenced_works,open_access,publication_year,type,counts_by_year
            ```
    2.  **Extract Metrics from the Response:** Parse the JSON response from OpenAlex. The `Work` object will contain the fields listed above.
        *   **Citation Count:** Directly from `cited_by_count`.
        *   **Reference Count:** Calculate `len(referenced_works)`.
        *   **Open Access Status:** From the `open_access` object (`is_oa`, `oa_status`).
        *   **Citations Per Year:** From the `counts_by_year` array.

*   **Considerations for the `go-go-mcp` `metrics` command:**
    *   The `scholarly.GetMetrics` function (as used in `cmd/apps/scholarly/cmd/metrics.go`) is responsible for:
        1.  Accepting a "Work ID" (which can be a DOI or OpenAlex ID).
        2.  Resolving this ID to an OpenAlex `Work` object (and potentially data from other sources like Crossref, as indicated by existing logs).
        3.  Populating a `scholarly.Metrics` struct. This struct would map fields like:
            *   `CitationCount` <- `cited_by_count` from OpenAlex.
            *   `CitedByCount` <- also `cited_by_count` (OpenAlex's `cited_by_count` is generally the primary incoming citation measure).
            *   `ReferenceCount` <- `len(referenced_works)`.
            *   `IsOA` <- `open_access.is_oa`.
            *   `OAStatus` <- `open_access.oa_status`.
    *   **Altmetrics:** The `Altmetrics` map in the `scholarly.Metrics` struct (seen in `cmd/apps/scholarly/cmd/metrics.go`) is not directly populated by standard OpenAlex fields. OpenAlex focuses on scholarly metadata and citation links. Altmetrics (e.g., social media mentions, news coverage) typically come from specialized services like Altmetric.com or PlumX. If this field is to be populated, `go-go-mcp` would need to integrate with such third-party altmetric data providers. Currently, it is likely to remain empty if only OpenAlex and Crossref are consulted.

## 6. Conclusion

Understanding the distinctions between DOIs and OpenAlex IDs, and how to correctly use them with the OpenAlex API, is crucial for accurately retrieving scholarly information. For applications like `go-go-mcp` that aim to provide metrics, leveraging the rich metadata from OpenAlex involves:
1.  Correctly identifying the type of input "Work ID".
2.  Using the appropriate OpenAlex API endpoint and parameters (direct lookup or filtering).
3.  Extracting relevant fields from the OpenAlex `Work` object to populate application-specific metrics structures.

By following these practices, we can build robust tools that effectively harness the power of the OpenAlex knowledge graph. 