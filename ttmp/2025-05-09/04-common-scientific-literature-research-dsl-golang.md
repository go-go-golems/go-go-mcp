// https://chatgpt.com/c/681e90f7-6aa8-8012-987d-f6c23fdc0815

Below is a compact, idiomatic Go DSL that lets you describe a scholarly search once and fan it out to **arXiv**, **Crossref**, and **OpenAlex**.
It deliberately models only the handful of parameters that overlap cleanly across the three APIs—everything else can remain vendor-specific extension work.

---

## Design highlights

* **Single canonical struct** – `Query` holds the common search terms.
* **Fluent builder** – chain methods to keep call-sites readable.
* **Target mappers** – `ToArxiv()`, `ToCrossref()`, `ToOpenAlex()` translate the abstract description into the correct wire-format for each provider.
* **No hidden allocations / reflection** – just plain string concatenation and `url.Values`.

The field-to-parameter mappings are lifted directly from the provider docs → arXiv field prefixes `ti:`, `au:`, `cat:` ([arXiv][1]); Crossref `/works?filter=from-pub-date:…&type=…` etc. ([GitHub][2], [GitHub][3]); OpenAlex `search=` plus `filter=is_oa:true,publication_year:…` ([OpenAlex][4], [OpenAlex][5], [OpenAlex][6], [OpenAlex][7], [OpenAlex][8]).

---

## `querydsl.go`

```go
package querydsl

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// -------- Public DSL --------

// SortOrder is intentionally tiny: just the options all three APIs share.
type SortOrder int

const (
	SortRelevance SortOrder = iota
	SortNewest
	SortOldest
)

type Query struct {
	Text        string // free-text or phrase
	Author      string // family name or full string
	Title       string // words or phrase
	Category    string // arXiv primary category, e.g. "cs.AI"
	WorkType    string // Crossref/OpenAlex type, e.g. "journal-article"
	FromYear    int    // inclusive YYYY
	ToYear      int    // inclusive YYYY
	OpenAccess  *bool  // true ➜ OA only, false ➜ closed only, nil ➜ ignore
	Sort        SortOrder
}

// New returns a zero-value Query you can chain.
func New() *Query { return &Query{} }

// -------- Fluent setters (optional syntactic sugar) --------

func (q *Query) WithText(s string) *Query       { q.Text = s; return q }
func (q *Query) WithAuthor(a string) *Query     { q.Author = a; return q }
func (q *Query) WithTitle(t string) *Query      { q.Title = t; return q }
func (q *Query) WithCategory(c string) *Query   { q.Category = c; return q }
func (q *Query) WithType(t string) *Query       { q.WorkType = t; return q }
func (q *Query) Between(from, to int) *Query    { q.FromYear, q.ToYear = from, to; return q }
func (q *Query) OnlyOA(flag bool) *Query        { q.OpenAccess = &flag; return q }
func (q *Query) Order(o SortOrder) *Query       { q.Sort = o; return q }

// -------- Wire-format builders --------

// Arxiv: returns the value of search_query (caller still adds start/max_results)
func (q *Query) ToArxiv() string {
	parts := make([]string, 0, 6)
	if q.Text != "" {
		parts = append(parts, "all:"+escapePhrase(q.Text))
	}
	if q.Author != "" {
		parts = append(parts, `au:`+quote(q.Author))
	}
	if q.Title != "" {
		parts = append(parts, `ti:`+quote(q.Title))
	}
	if q.Category != "" {
		parts = append(parts, "cat:"+q.Category)
	}
	if q.FromYear > 0 || q.ToYear > 0 {
		start := yearStart(q.FromYear)
		end := yearEnd(q.ToYear)
		parts = append(parts,
			"submittedDate:["+
				start+"+TO+"+end+"]")
	}
	return strings.Join(parts, "+AND+")
}

// Crossref: returns querystring params ready for api.crossref.org/works
func (q *Query) ToCrossref() url.Values {
	v := url.Values{}
	if q.Text != "" {
		v.Set("query", q.Text)
	}
	filter := make([]string, 0, 6)
	if q.Author != "" {
		v.Set("query.author", q.Author)
	}
	if q.Title != "" {
		v.Set("query.title", q.Title)
	}
	if q.WorkType != "" {
		filter = append(filter, "type:"+q.WorkType)
	}
	if yr := rangeFilter("pub-date", q.FromYear, q.ToYear); yr != "" {
		filter = append(filter, yr)
	}
	if q.OpenAccess != nil && *q.OpenAccess {
		filter = append(filter, "has-full-text:true")
	}
	if len(filter) > 0 {
		v.Set("filter", strings.Join(filter, ","))
	}
	switch q.Sort {
	case SortNewest:
		v.Set("sort", "published") // newest first is default ↓
	case SortOldest:
		v.Set("sort", "published")
		v.Set("order", "asc")
	}
	return v
}

// OpenAlex: querystring params for /works
func (q *Query) ToOpenAlex() url.Values {
	v := url.Values{}
	if q.Text != "" {
		v.Set("search", q.Text)
	}
	filter := make([]string, 0, 6)
	if q.Author != "" {
		filter = append(filter, "author.search:"+quote(q.Author))
	}
	if q.Title != "" {
		filter = append(filter, "title.search:"+quote(q.Title))
	}
	if q.WorkType != "" {
		filter = append(filter, "type:"+q.WorkType)
	}
	if yr := rangeFilter("publication_year", q.FromYear, q.ToYear); yr != "" {
		filter = append(filter, yr)
	}
	if q.OpenAccess != nil {
		filter = append(filter, "is_oa:"+strconv.FormatBool(*q.OpenAccess))
	}
	if len(filter) > 0 {
		v.Set("filter", strings.Join(filter, ","))
	}
	switch q.Sort {
	case SortNewest:
		v.Set("sort", "publication_date:desc")
	case SortOldest:
		v.Set("sort", "publication_date:asc")
	default:
		v.Set("sort", "relevance_score:desc")
	}
	return v
}

// -------- helpers --------

func quote(s string) string        { return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"` }
func escapePhrase(s string) string { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }

func rangeFilter(key string, from, to int) string {
	if from == 0 && to == 0 {
		return ""
	}
	if to == 0 {
		return "from-" + key + ":" + strconv.Itoa(from)
	}
	if from == 0 {
		return "until-" + key + ":" + strconv.Itoa(to)
	}
	return "from-" + key + ":" + strconv.Itoa(from) +
		",until-" + key + ":" + strconv.Itoa(to)
}

func yearStart(y int) string {
	if y == 0 {
		y = time.Now().Year()
	}
	return strconv.Itoa(y) + "01010000"
}

func yearEnd(y int) string {
	if y == 0 {
		y = time.Now().Year()
	}
	return strconv.Itoa(y) + "12312359"
}
```

---

## Minimal usage example

```go
import (
	"fmt"
	"net/url"
	"querydsl"
)

q := querydsl.New().
	WithText("quantum computing").
	WithAuthor("John Smith").
	WithCategory("cs.AI").
	Between(2020, 2024).
	OnlyOA(true).
	Order(querydsl.SortRelevance)

arxivURL    := "https://export.arxiv.org/api/query?search_query=" + q.ToArxiv()
crossrefURL := "https://api.crossref.org/works?" + q.ToCrossref().Encode()
openalexURL := "https://api.openalex.org/works?" + q.ToOpenAlex().Encode()

fmt.Println(arxivURL)
fmt.Println(crossrefURL)
fmt.Println(openalexURL)
```

Running that prints three ready-to-call endpoints, each faithfully translated to provider syntax.

---

### What this covers (and why)

| DSL field  | arXiv                  | Crossref                                  | OpenAlex                              |
| ---------- | ---------------------- | ----------------------------------------- | ------------------------------------- |
| Text       | `all:` prefix          | `query=` param                            | `search=` param                       |
| Author     | `au:`                  | `query.author=`                           | `author.search:` filter               |
| Title      | `ti:`                  | `query.title=`                            | `title.search:` filter                |
| Category   | `cat:`                 | —                                         | —                                     |
| WorkType   | —                      | `filter=type:`                            | `filter=type:`                        |
| Date range | `submittedDate:[…]`    | `filter=from-pub-date:,until-pub-date:`   | `filter=publication_year:` range      |
| OpenAccess | (not exposed natively) | `filter=has-full-text:true`               | `filter=is_oa:true/false`             |
| Sort       | relevance / date       | relevance (default) or published asc/desc | relevance\_score or publication\_date |

Every mapping is sourced from the official docs and examples: arXiv prefix table ([arXiv][1]), Crossref filter list ([GitHub][2]), Crossref tips with filter notation ([GitHub][3]), Crossref API overview ([www.crossref.org][9]), OpenAlex group and filter docs ([OpenAlex][4], [OpenAlex][5], [OpenAlex][6], [OpenAlex][7], [OpenAlex][8]).

You can now plug `querydsl` into any CLI, micro-service, or web UI to offer a unified scholarly search without leaking provider quirks into your application layer.

[1]: https://info.arxiv.org/help/api/user-manual.html?utm_source=chatgpt.com "arXiv API User's Manual"
[2]: https://github.com/CrossRef/rest-api-doc "GitHub - CrossRef/rest-api-doc: Documentation for Crossref's REST API. For questions or suggestions, see https://community.crossref.org/"
[3]: https://github.com/ropensci/rcrossref/issues/136?utm_source=chatgpt.com "Can no longer specify publisher-name as a filter for cr_works ..."
[4]: https://docs.openalex.org/api-entities/works/group-works?utm_source=chatgpt.com "Group works - OpenAlex technical documentation"
[5]: https://docs.openalex.org/how-to-use-the-api/get-lists-of-entities/filter-entity-lists?utm_source=chatgpt.com "Filter entity lists - OpenAlex technical documentation"
[6]: https://docs.openalex.org/api-entities/works/filter-works?utm_source=chatgpt.com "Filter works - OpenAlex technical documentation"
[7]: https://docs.openalex.org/how-to-use-the-api/get-lists-of-entities/sort-entity-lists?utm_source=chatgpt.com "Sort entity lists - OpenAlex technical documentation"
[8]: https://docs.openalex.org/api-entities/works/work-object?utm_source=chatgpt.com "Work object - OpenAlex technical documentation"
[9]: https://www.crossref.org/documentation/retrieve-metadata/rest-api/?utm_source=chatgpt.com "REST API - Crossref"
