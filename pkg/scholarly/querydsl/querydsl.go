package querydsl

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SortOrder is intentionally tiny: just the options all three APIs share.
type SortOrder int

const (
	SortRelevance SortOrder = iota
	SortNewest
	SortOldest
)

type Query struct {
	Text       string // free-text or phrase
	Author     string // family name or full string
	Title      string // words or phrase
	Category   string // arXiv primary category, e.g. "cs.AI"
	WorkType   string // Crossref/OpenAlex type, e.g. "journal-article"
	FromYear   int    // inclusive YYYY
	ToYear     int    // inclusive YYYY
	OpenAccess *bool  // true ➜ OA only, false ➜ closed only, nil ➜ ignore
	Sort       SortOrder
	MaxResults int // maximum number of results to return
}

// New returns a zero-value Query you can chain.
func New() *Query { return &Query{MaxResults: 20} }

// -------- Fluent setters (optional syntactic sugar) --------

func (q *Query) WithText(s string) *Query     { q.Text = s; return q }
func (q *Query) WithAuthor(a string) *Query   { q.Author = a; return q }
func (q *Query) WithTitle(t string) *Query    { q.Title = t; return q }
func (q *Query) WithCategory(c string) *Query { q.Category = c; return q }
func (q *Query) WithType(t string) *Query     { q.WorkType = t; return q }
func (q *Query) Between(from, to int) *Query  { q.FromYear, q.ToYear = from, to; return q }
func (q *Query) OnlyOA(flag bool) *Query      { q.OpenAccess = &flag; return q }
func (q *Query) Order(o SortOrder) *Query     { q.Sort = o; return q }
func (q *Query) WithMaxResults(n int) *Query  { q.MaxResults = n; return q }

// -------- Wire-format builders --------

// ToArxiv returns the value of search_query (caller still adds start/max_results)
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

// ToCrossref returns querystring params ready for api.crossref.org/works
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
	v.Set("rows", strconv.Itoa(q.MaxResults))
	return v
}

// ToOpenAlex returns querystring params for /works
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
	v.Set("per_page", strconv.Itoa(q.MaxResults))
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
