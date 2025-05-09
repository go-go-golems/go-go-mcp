**Specifying LLM Agent Tools**

The user has requested a specification for a subset of tools related to a research client using libgen, crossref, or OpenAlex. They likely want the spec to include function names, descriptions, parameters/arguments with validation rules, and examples like JSON schema, sample calls/responses, and possible error messages. They also might need details like rate limiting and backend mappings, while preferring in-depth, structured documentation using markdown. I’ll need to focus solely on these core technical details and avoid anything outside of the request.

**Specifying Functions for LLM Agent**

The user wants a concise but thorough specification for six functions, with no extra features or platitudes like error handling suggestions. The spec should include details such as JSON parameter schemas and expected return objects. It should be in markdown, with subsections for each tool. Since Crossref and OpenAlex APIs are likely stable by 2025, we may need to confirm parameter names, but browsing for the latest endpoint information might be beneficial to ensure up-to-date accuracy while avoiding unnecessary complexity.


---

The user wants a well-defined, concise specification for the six functions, without extra features or broad recommendations like error handling tips. However, specifying error responses is okay for completeness. The focus should be on detailing request methods and their parameters in JSON, not code suggestions. Given the user’s request about avoiding error-handling platitudes, I’ll avoid mentioning those, but still ensure necessary details like request/response formats are included for clarity. Let’s focus on the essential, functional points of the spec.

**Specifying Endpoints and Parameters**

The spec should define the JSON parameter schema and the expected return objects for each tool, presented in markdown with subsections for clarity. While Crossref and OpenAlex endpoints are likely stable, it would be a good idea to check for updates, especially for OpenAlex’s newer version. I'll be thorough by including relevant details such as accurate parameter mapping, ensuring the specification works with the latest available knowledge as of May 2025. We may need to browse for specifics but can also rely on existing knowledge.


**Defining API Specifications**

For the spec, we'll cover five functions outlined by the user: `search_works`, `resolve_doi`, `suggest_keywords`, `get_metrics`, and `get_citations`. Each function will have clear parameters and return objects specified, focusing mainly on metadata. For example, `search_works` will return a list of Work objects with attributes like `id`, `doi`, and `title`, while `resolve_doi` will return a single Work object. We'll ensure that everything follows a consistent structure to fulfill the user’s requests.


**API Specification Breakdown**

We need to define the specs for several functions. For example, `search_works` will search Crossref and OpenAlex with parameters like `query`, `source`, `limit`, and `fields`. It will return a list of Work objects, including info such as title, author, and citation count. For `resolve_doi`, it will return a Work object. Similarly, we’ll define specs for other functions like `suggest_keywords` and `get_metrics`, ensuring to provide the necessary metadata and clear return formats.

**Specifying find\_full\_text Tool**

The `find_full_text` function should take either a DOI or title as input, with source order preference: OpenAlex > Unpaywall > LibGen. The expected output should include a JSON object with fields such as `pdf_url`, `source`, `oa_status`, `license`, `version`, and possibly `mirror_url`.

The spec will define the function signature, arguments schema, return schema, and provide example JSON for each. Error conditions will also be included. We’ll reference sources like OpenAlex API docs for accurate parameter mapping, but avoid unnecessary suggestions like caching. The user prefers comprehensive details without extra features.


Thought for 49 seconds


Below is a pragmatic, implementation-ready specification for the six LLM-agent tools you asked about. Each spec is written as a self-contained contract: it defines the JSON argument schema the agent must supply, the JSON it will receive back, and the external API calls that the tool wrapper should perform under the hood.

---

### 1 · `search_works`

|                                     |                                                                                                                                                                                                                                                                                                                                                                             |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Purpose**                         | Text-search scholarly records (title ∕ abstract ∕ full-text) across Crossref or OpenAlex and return lightweight metadata for ranking.                                                                                                                                                                                                                                       |
| **Target APIs**                     | *OpenAlex* `GET /works?search={query}&per_page={limit}`  ([OpenAlex Overview][1])  <br> *Crossref* `GET /works?query={query}&rows={limit}`  ([www.crossref.org][2])                                                                                                                                                                                                         |
| **Arguments (JSON Schema 2020-12)** | `json { "type": "object", "properties": { "query": { "type": "string" }, "source": { "type": "string", "enum": ["openalex", "crossref"] }, "limit": { "type": "integer", "minimum": 1, "maximum": 100, "default": 20 }, "filter": { "type": "object", "description": "Optional backend-specific filter string (e.g., has_oa:true)" } }, "required": ["query", "source"] } ` |
| **Return**                          | `json { "works": [ { "id": "W2741809809", "doi": "10.1038/nphys1170", "title": "The rise of quantum biology", "authors": ["E. K. Irish", "N. Lambert"], "year": 2008, "is_oa": true, "citation_count": 312 } … ] } `                                                                                                                                                        |
| **Notes**                           | *id* is the provider’s opaque work-id (`OpenAlex: W…`, `Crossref: DOI`).  Return objects are sorted by provider relevance score; no extra ranking logic inside the tool.                                                                                                                                                                                                    |

---

### 2 · `resolve_doi`

|                     |                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Purpose**         | Fetch complete metadata for one DOI, collapsing Crossref and OpenAlex into a single rich record.                                                                                                                                                                                                                                                                                                                                   |
| **Target APIs**     | *Crossref* `GET /works/{doi}` (good for publisher & license) ([www.crossref.org][2]) <br>*OpenAlex* `GET /works/https://doi.org/{doi}` (citations + concepts) ([OpenAlex Overview][3])                                                                                                                                                                                                                                             |
| **Arguments**       | `json { "type": "object", "properties": { "doi": { "type": "string", "pattern": "^10\\..+/.+" } }, "required": ["doi"] } `                                                                                                                                                                                                                                                                                                         |
| **Return (merged)** | `json { "id": "https://openalex.org/W2741809809", "doi": "10.1038/nphys1170", "title": "The rise of quantum biology", "authors": [ … ], "year": 2008, "journal": "Nature Physics", "is_oa": true, "oa_status": "green", "license": "CC-BY-4.0", "citation_count": 312, "referenced_works": [ "W201...", … ], "cited_by_count": 911, "concepts": [ { "id": "C121332964", "display_name": "Quantum biology", "score": 0.86 } … ] } ` |
| **Notes**           | Field precedence: OpenAlex wins for *citations*, Crossref wins for *license* if conflicting.                                                                                                                                                                                                                                                                                                                                       |

---

### 3 · `suggest_keywords`

|                |                                                                                                                                                                             |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Purpose**    | Return controlled-vocabulary concepts for a work title (or arbitrary text) to seed further searches.                                                                        |
| **Target API** | *OpenAlex text-aboutness endpoint* `POST /text` (supports free text ➜ concept IDs) ([OpenAlex Overview][4])                                                                 |
| **Arguments**  | `json { "type":"object", "properties": { "text": { "type":"string" }, "max_keywords": { "type":"integer","default":10,"minimum":1,"maximum":50 } }, "required":["text"] } ` |
| **Return**     | `json { "keywords": [ { "id":"C41008148", "display_name":"Transformer (machine learning)", "relevance":0.92 }, … ] } `                                                      |

---

### 4 · `get_metrics`

|                |                                                                                                                                                                |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Purpose**    | Pull headline quantitative metrics for quick triage of a work.                                                                                                 |
| **Target API** | *OpenAlex* `GET /works/{work_id}` (provides all stats in one call) ([OpenAlex Overview][3])                                                                    |
| **Arguments**  | `json { "type":"object", "properties":{ "work_id":{ "type":"string" } }, "required":["work_id"] } `                                                            |
| **Return**     | `json { "citation_count":312, "cited_by_count":911, "reference_count":42, "is_oa":true, "oa_status":"green", "altmetrics":{ "views":1234,"downloads":456 } } ` |
| **Notes**      | The wrapper simply surfaces numeric fields already present in the OpenAlex Work object—no extra math.                                                          |

---

### 5 · `get_citations`

|                |                                                                                                                                                                                                                                             |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Purpose**    | Retrieve one hop of the citation graph—either the works this paper cites (*outgoing*) or the works that cite it (*incoming*).                                                                                                               |
| **Target API** | *OpenAlex* <br>  • Outgoing:  returned in `referenced_works` array of the work itself. <br>  • Incoming: `GET /works?filter=cites:{work_id}` (with cursor for paging) ([OpenAlex Overview][5])                                              |
| **Arguments**  | `json { "type":"object", "properties":{ "work_id":{"type":"string"}, "direction":{"type":"string","enum":["refs","cited_by"]}, "limit":{"type":"integer","default":100,"minimum":1,"maximum":200} }, "required":["work_id","direction"] } ` |
| **Return**     | `json { "citations":[ { "id":"W201...", "doi":"10.1103/PhysRevLett.116.061102", "title":"Observation of gravitational waves...", "year":2016 }, … ], "next_cursor":"AoJ0..." } `                                                            |

---

### 6 · `find_full_text`

|                                          |                                                                                                                                                                                                                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Purpose**                              | Hand back the best PDF (or HTML) URL for a work, trying legal OA sources first and falling back to LibGen.                                                                                                                                                             |
| **Resolution order (stop at first hit)** | 1 OpenAlex `locations` object → pick `is_oa = true` & `version = "publishedVersion"` if present. ([OpenAlex Overview][6]) <br>2 Unpaywall (`GET https://api.unpaywall.org/v2/{doi}`) <br>3 LibGen search via *libgen-api* (`GET /json.php?&title={title}`) ([PyPI][7]) |
| **Arguments**                            | `json { "type":"object", "properties":{ "doi":{"type":"string"}, "title":{"type":"string"}, "prefer_version":{"type":"string","enum":["published","accepted","submitted"],"default":"published"} }, "anyOf":[ {"required":["doi"]}, {"required":["title"]} ] } `       |
| **Return**                               | `json { "pdf_url":"https://journals.aps.org/prl/pdf/10.1103/PhysRevLett.116.061102", "source":"openalex", "oa_status":"gold", "license":"CC-BY-4.0" } `                                                                                                                |
| **Notes**                                | • The tool never persists files; it only returns a resolvable URL. <br>• If LibGen is used, `source` is `"libgen"` and an extra field `"md5"` is included for provenance.                                                                                              |

---

#### Common Conventions

* All tools accept and return **UTF-8 JSON**.
* IDs are provider identifiers (`openalex: W…`) unless explicitly a DOI.
* `limit` arguments cap remote `per_page` or `rows` parameters to avoid silent truncation.
* Pagination cursors follow OpenAlex’s opaque `cursor` tokens verbatim.

These six well-scoped primitives let an agent **discover → enrich → acquire → analyze** scholarly literature without embedding any provider-specific quirks into prompt logic.

[1]: https://docs.openalex.org/api-entities/works/search-works?utm_source=chatgpt.com "Search works | OpenAlex technical documentation"
[2]: https://www.crossref.org/documentation/retrieve-metadata/rest-api/?utm_source=chatgpt.com "REST API - Crossref"
[3]: https://docs.openalex.org/api-entities/works/work-object?utm_source=chatgpt.com "Work object - OpenAlex technical documentation"
[4]: https://docs.openalex.org/how-to-use-the-api/api-overview?utm_source=chatgpt.com "API Overview | OpenAlex technical documentation"
[5]: https://docs.openalex.org/api-entities/works/filter-works?utm_source=chatgpt.com "Filter works - OpenAlex technical documentation"
[6]: https://docs.openalex.org/api-entities/works/work-object/location-object?utm_source=chatgpt.com "Location object | OpenAlex technical documentation"
[7]: https://pypi.org/project/libgen-api/?utm_source=chatgpt.com "libgen-api - PyPI"
