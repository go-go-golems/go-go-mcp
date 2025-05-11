package common

// Work represents a simplified scholarly record for search results
type Work struct {
	ID            string   `json:"id"`
	DOI           string   `json:"doi,omitempty"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors,omitempty"`
	Year          int      `json:"year,omitempty"`
	IsOA          bool     `json:"is_oa,omitempty"`
	CitationCount int      `json:"citation_count,omitempty"`
	Abstract      string   `json:"abstract,omitempty"`
	SourceName    string   `json:"source_name,omitempty"`
	PDFURL        string   `json:"pdf_url,omitempty"`
}

// SearchWorksRequest represents parameters for searching works
type SearchWorksRequest struct {
	Query  string            `json:"query"`
	Source string            `json:"source"` // "openalex", "crossref", "arxiv"
	Limit  int               `json:"limit,omitempty"`
	Filter map[string]string `json:"filter,omitempty"`
}

// SearchWorksResponse represents the response from a search_works function
type SearchWorksResponse struct {
	Works []Work `json:"works"`
}

// ResolveDOIRequest represents parameters for resolving a DOI
type ResolveDOIRequest struct {
	DOI string `json:"doi"`
}

// Keywords represents a list of keywords with relevance scores
type Keyword struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Relevance   float64 `json:"relevance"`
}

// SuggestKeywordsRequest represents parameters for suggesting keywords
type SuggestKeywordsRequest struct {
	Text        string `json:"text"`
	MaxKeywords int    `json:"max_keywords,omitempty"`
}

// SuggestKeywordsResponse represents the response from suggest_keywords
type SuggestKeywordsResponse struct {
	Keywords []Keyword `json:"keywords"`
}

// GetMetricsRequest represents parameters for getting metrics
type GetMetricsRequest struct {
	WorkID string `json:"work_id"`
}

// Metrics represents quantitative metrics for a work
type Metrics struct {
	CitationCount  int            `json:"citation_count"`
	CitedByCount   int            `json:"cited_by_count"`
	ReferenceCount int            `json:"reference_count"`
	IsOA           bool           `json:"is_oa"`
	OAStatus       string         `json:"oa_status,omitempty"`
	Altmetrics     map[string]int `json:"altmetrics,omitempty"`
}

// GetCitationsRequest represents parameters for getting citations
type GetCitationsRequest struct {
	WorkID    string `json:"work_id"`
	Direction string `json:"direction"` // "refs" or "cited_by"
	Limit     int    `json:"limit,omitempty"`
}

// Citation represents a lightweight citation record
type Citation struct {
	ID    string `json:"id"`
	DOI   string `json:"doi,omitempty"`
	Title string `json:"title"`
	Year  int    `json:"year,omitempty"`
}

// GetCitationsResponse represents the response from get_citations
type GetCitationsResponse struct {
	Citations  []Citation `json:"citations"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

// FindFullTextRequest represents parameters for finding full text
type FindFullTextRequest struct {
	DOI           string `json:"doi,omitempty"`
	Title         string `json:"title,omitempty"`
	PreferVersion string `json:"prefer_version,omitempty"`
}

// FindFullTextResponse represents the response from find_full_text
type FindFullTextResponse struct {
	PDFURL   string `json:"pdf_url"` // Can be PDF or HTML URL depending on what's available
	Source   string `json:"source"`
	OAStatus string `json:"oa_status,omitempty"`
	License  string `json:"license,omitempty"`
	MD5      string `json:"md5,omitempty"`    // Only for LibGen
	IsPDF    bool   `json:"is_pdf,omitempty"` // Indicates if the URL is for a PDF (false means HTML)
}

// SearchResult represents a standardized search result from any source
type SearchResult struct {
	Title         string
	Authors       []string
	Abstract      string
	Published     string
	DOI           string
	PDFURL        string
	SourceURL     string
	SourceName    string
	OAStatus      string
	License       string
	FileSize      string
	Citations     int
	Type          string
	JournalInfo   string
	Metadata      map[string]interface{} // Additional source-specific data
	Reranked      bool                   `json:"reranked,omitempty"`       // Whether this result was reranked
	RerankerScore float64                `json:"reranker_score,omitempty"` // Score from the reranker
	OriginalIndex int                    `json:"original_index,omitempty"` // Original position before reranking
}

// SearchParams contains common search parameters
type SearchParams struct {
	Query      string
	MaxResults int
	Filters    map[string]string
	Sort       string
	EmailAddr  string
}
