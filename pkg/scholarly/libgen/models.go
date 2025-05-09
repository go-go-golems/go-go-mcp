package libgen

// LibgenEntry represents a single search result from LibGen (scraped)
// Fields are based on common data found on LibGen search result pages for scientific articles.
type LibgenEntry struct {
	Authors     string
	Title       string
	JournalInfo string // Combined Journal, Volume, Issue, Year
	DOI         string
	FileSize    string
	DownloadURL string   // Primary link from title, if any
	MirrorLinks []string // Links to download pages on various mirrors
	EditLink    string   // Link to edit metadata
}