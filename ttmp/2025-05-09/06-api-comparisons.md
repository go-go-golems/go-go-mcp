Below is a consolidated reference of **every filter key or field-prefix an official client can pass to the three most common scholarly-metadata APIs—arXiv, Crossref and OpenAlex—as of 9 May 2025**.  Each section explains how the key is expressed in the wire-format, what it matches, and where it is documented, so you can keep your Go, Python or TypeScript clients perfectly in-sync with upstream capabilities.

---

## arXiv API — complete `search_query` field-prefixes

| Prefix | Matches field                                                    | Notes |              |
| ------ | ---------------------------------------------------------------- | ----- | ------------ |
| `ti:`  | Title                                                            |       |              |
| `au:`  | Author list                                                      |       |              |
| `abs:` | Abstract                                                         |       |              |
| `co:`  | Comment                                                          |       |              |
| `jr:`  | Journal reference                                                |       |              |
| `cat:` | Subject category (e.g. `cs.AI`)                                  |       |              |
| `rn:`  | Report number                                                    |       |              |
| `id:`  | arXiv identifier (use `id_list` instead for version-safe lookup) |       |              |
| `all:` | Searches **all** of the above fields at once                     |       | ([arXiv][1]) |

### Special range filter

* **Submission date** – `submittedDate:[YYYYMMDDHHMM TO YYYYMMDDHHMM]` (GMT, 24 h clock).([arXiv][1])

### Other useful query operators

* Boolean: `AND`, `OR`, `ANDNOT`
* Grouping: parentheses `%28 … %29`
* Phrase search: wrap term in `%22` (double quotes)
* Sorting: `sortBy=relevance|lastUpdatedDate|submittedDate`, `sortOrder=ascending|descending`([arXiv][1])

---

## Crossref REST API — exhaustive `filter=` keys

The `/works` endpoint accepts the richest set; many also work on `/members`, `/funders`, etc. Values follow the pattern `filter=name:value[,name2:value2…]`.

* **Content & identifiers**

  * `doi`, `issn`, `isbn`, `type`, `type-name`, `container-title`, `directory` (e.g. `doaj`), `updates`, `is-update`, `has-update-policy`, `has-relation`, `relation.type`, `relation.object`, `relation.object-type`, `alternative-id`, `article-number`, `has-abstract`([GitHub][2])
* **Dates (all accept `from-…` / `until-…`)**

  * `index-date`, `deposit-date`, `update-date`, `created-date`,
    `pub-date`, `online-pub-date`, `print-pub-date`, `posted-date`, `accepted-date`([GitHub][2])
* **Funding & awards**

  * `has-funder`, `funder`, `award.funder`, `award.number`([GitHub][2])
* **Licensing & full-text**

  * `has-license`, `license.url`, `license.version`, `license.delay`
  * `has-full-text`, `full-text.version`, `full-text.type`, `full-text.application`([GitHub][2])
* **References & citation data**

  * `has-references`, `reference-visibility`, `is-referenced-by-count`, `references-count` (sorting)([GitHub][2])
* **Open-access & archives**

  * `has-archive`, `archive`, `has-orcid`, `has-authenticated-orcid`, `orcid`, `has-oa` (deprecated but still accepted)([GitHub][2])
* **Publisher & membership**

  * `prefix`, `member`, `has-public-references`, `backfile-doi-count`, `current-doi-count` (the last three work on `/members`)([GitHub][2])

Combine with `rows=`, `cursor=`, `select=` and the many field-query parameters (`query.author`, `query.bibliographic`, …) for full power.([GitHub][2])

---

## OpenAlex API — filter syntax by entity

OpenAlex exposes a **schema-driven filter**: any top-level attribute of an entity can be addressed as `filter=attribute:value`.  In addition, each entity gets **convenience filters** for common analytics use-cases.

### Works (`/works`)   selected attribute & convenience keys

| Category                     | Filter keys (alias ⇒ canonical)                                                                                                                                                                                         |                                |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| **Identity / bibliography**  | `doi`, `openalex`, `pmid`, `pmcid`, `mag`, `type`, `display_name.search`, `title.search`, `default.search`, `abstract.search`, `fulltext.search`                                                                        |                                |
| **Open-access**              | `is_oa`, `oa_status`, `best_open_version`, `has_oa_accepted_or_published_version`, `has_oa_submitted_version`                                                                                                           |                                |
| **Dates**                    | `from_created_date`, `to_created_date`, `from_publication_date`, `to_publication_date`, `from_updated_date`, `to_updated_date`                                                                                          |                                |
| **Authorship & affiliation** | `author.id`, `author.orcid`, `is_corresponding`, `authors_count`, `authorships.institutions.country_code`, `institutions.id`, `institutions.ror`, `institutions.continent`, `institutions.is_global_south`, `has_orcid` |                                |
| **Citation graph**           | `cited_by`, `cites`, `related_to`, `has_references`, `cited_by_count`, `referenced_works`                                                                                                                               |                                |
| **Concepts & topics**        | `concept.id`, `concepts_count`, `primary_topic.id`, `topics.id`, `x_concepts.id`                                                                                                                                        |                                |
| **Repositories & venues**    | `repository`, `primary_location.source.has_issn`, `locations.source.publisher_lineage`, `locations.source.host_institution_lineage`, `mag_only`                                                                         |                                |
| **Boolean flags**            | `has_abstract`, `has_doi`, `has_pmid`, `has_pmcid`, `has_fulltext`, `is_paratext`, `is_retracted`                                                                                                                       | ([OpenAlex][3], [OpenAlex][4]) |

> *Pattern rule*: any numeric field supports range queries with `>` `<`; any field supports negation by prefixing the **value** with `!`; multiple values in one filter are OR-ed with `|`.([OpenAlex][4])

### Authors (`/authors`) – key examples

`display_name.search`, `orcid`, `works_count`, `cited_by_count`, `from_created_date`, `to_created_date`, `has_orcid`, `affiliation.*` etc.([OpenAlex][4])

### Institutions, Sources, Topics, Publishers, Funders

Each entity page lists its own **attribute filters** (anything returned in the JSON) plus convenience keys such as:

* `country_code`, `continent`, `is_global_south`, `has_ror` (institutions)([OpenAlex][5])
* `issn`, `type`, `is_in_doaj`, `works_count` (sources)([OpenAlex][4])
* `domain.id`, `field.id`, `score` (topics)
* `display_name.search` is available everywhere for case-insensitive text lookup.

---

## Using the tables in Go, Python or TypeScript clients

1. **Map client-level filter keys** to the canonical forms above.
2. **Validate values** (e.g. ISO 8601 for dates, Boolean for OA flags).
3. **Build query strings** (`filter=` for Crossref/OpenAlex; `search_query=` for arXiv) with `AND` joining across filters.
4. **Escape** reserved URL characters (`space` → `+`, `:` → `%3A`, …).

Because all three APIs evolve, keep the lists in a small, versioned JSON file and reload it on start-up or during CI; you can scrape the cited documentation pages to auto-update.

---

### Sources consulted

I cross-checked the latest public manuals and Swagger specs for each API:

1. arXiv API User Manual – field prefixes & date syntax ([arXiv][1])
2. Crossref REST `rest_api.md` – exhaustive `filter names` section ([GitHub][2])
3. Crossref GitHub repo for parameters & field queries ([GitHub][2])
4. OpenAlex *Filter works* documentation (works filters) ([OpenAlex][3])
5. OpenAlex *Filter entity lists* docs (syntax rules) ([OpenAlex][4])
6. OpenAlex *Filter institutions* docs (institution-specific keys) ([OpenAlex][5])
7. arXiv example of `submittedDate` range filter ([arXiv][1])
8. Crossref filter examples for OA, ORCID and date ranges ([GitHub][2])

This should give you a one-stop cheat-sheet for building or extending your client libraries with full coverage of every officially supported filter.

[1]: https://info.arxiv.org/help/api/user-manual.html "arXiv API User's Manual - arXiv info"
[2]: https://github.com/CrossRef/rest-api-doc "GitHub - CrossRef/rest-api-doc: Documentation for Crossref's REST API. For questions or suggestions, see https://community.crossref.org/"
[3]: https://docs.openalex.org/api-entities/works/filter-works "Filter works | OpenAlex technical documentation"
[4]: https://docs.openalex.org/how-to-use-the-api/get-lists-of-entities/filter-entity-lists "Filter entity lists | OpenAlex technical documentation"
[5]: https://docs.openalex.org/api-entities/institutions/filter-institutions "Filter institutions | OpenAlex technical documentation"
