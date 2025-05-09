package arxiv

import "encoding/xml"

// AtomFeed represents the top-level structure of the Arxiv API response.
// Based on Atom 1.0 and Arxiv API specifics.
type AtomFeed struct {
	XMLName      xml.Name `xml:"feed"`
	Title        string   `xml:"title"`
	ID           string   `xml:"id"`
	Updated      string   `xml:"updated"`
	Entries      []Entry  `xml:"entry"`
	TotalResults int      `xml:"totalResults"` // OpenSearch extension
	StartIndex   int      `xml:"startIndex"`   // OpenSearch extension
	ItemsPerPage int      `xml:"itemsPerPage"` // OpenSearch extension
}

// Entry represents a single paper in the Arxiv API response.
// Based on Atom 1.0 and Arxiv API specifics.
type Entry struct {
	ID              string     `xml:"id"` // Usually the Arxiv URL for the paper
	Updated         string     `xml:"updated"`
	Published       string     `xml:"published"`
	Title           string     `xml:"title"`
	Summary         string     `xml:"summary"` // Abstract
	Authors         []Author   `xml:"author"`
	DOI             string     `xml:"doi"`         // Arxiv extension
	Comment         string     `xml:"comment"`     // Arxiv extension
	JournalRef      string     `xml:"journal_ref"` // Arxiv extension
	Link            []Link     `xml:"link"`
	PrimaryCategory Category   `xml:"primary_category"` // Arxiv extension
	Categories      []Category `xml:"category"`
}

// Author represents an author of a paper.
// Based on Atom 1.0.
type Author struct {
	Name string `xml:"name"`
}

// Link represents a link related to the paper (e.g., PDF, abstract page).
// Based on Atom 1.0.
type Link struct {
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`
	Title string `xml:"title,attr,omitempty"`
}

// Category represents a subject category of the paper.
// Based on Atom 1.0 and Arxiv extension.
type Category struct {
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
}
