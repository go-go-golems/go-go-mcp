package common

// SearchResult represents a standardized search result from any source
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

// SearchParams contains common search parameters
type SearchParams struct {
	Query      string
	MaxResults int
	Filters    map[string]string
	Sort       string
	EmailAddr  string
}
