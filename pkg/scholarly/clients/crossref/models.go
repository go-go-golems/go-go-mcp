package crossref

// CrossrefResponse represents the top-level structure of the Crossref API response.
type CrossrefResponse struct {
	Status         string          `json:"status"`
	MessageType    string          `json:"message-type"`
	MessageVersion string          `json:"message-version"`
	Message        CrossrefMessage `json:"message"`
}

// CrossrefMessage contains the actual search results.
type CrossrefMessage struct {
	TotalResults int            `json:"total-results"`
	Items        []CrossrefItem `json:"items"`
	Query        CrossrefQuery  `json:"query"`
	ItemsPerPage int            `json:"items-per-page"`
}

// CrossrefQuery details the query performed.
type CrossrefQuery struct {
	SearchTerms string `json:"search-terms"`
	StartIndex  int    `json:"start-index"`
}

// CrossrefItem represents a single work item from Crossref.
type CrossrefItem struct {
	DOI       string            `json:"DOI"`
	Title     []string          `json:"title"`
	Author    []CrossrefAuthor  `json:"author,omitempty"`
	Publisher string            `json:"publisher"`
	Type      string            `json:"type"`
	Created   CrossrefDate      `json:"created,omitempty"`
	Issued    CrossrefDateParts `json:"issued,omitempty"` // Contains date-parts
	URL       string            `json:"URL"`
	Abstract  string            `json:"abstract,omitempty"` // May not always be present or may be truncated
	ISSN      []string          `json:"ISSN,omitempty"`
	ISBN      []string          `json:"ISBN,omitempty"`
	Subject   []string          `json:"subject,omitempty"`
	Link      []CrossrefLink    `json:"link,omitempty"`
}

// CrossrefAuthor represents an author.
type CrossrefAuthor struct {
	Given       string                `json:"given,omitempty"`
	Family      string                `json:"family,omitempty"`
	Sequence    string                `json:"sequence,omitempty"`
	Affiliation []CrossrefAffiliation `json:"affiliation,omitempty"`
}

// CrossrefAffiliation represents an author's affiliation.
type CrossrefAffiliation struct {
	Name string `json:"name,omitempty"`
}

// CrossrefDate represents a date object in Crossref responses.
type CrossrefDate struct {
	DateParts [][]int `json:"date-parts,omitempty"`
	DateTime  string  `json:"date-time,omitempty"`
	Timestamp int64   `json:"timestamp,omitempty"`
}

// CrossrefDateParts is specifically for the "issued" field which often has only date-parts.
type CrossrefDateParts struct {
	DateParts [][]int `json:"date-parts,omitempty"`
}

// CrossrefLink represents a link associated with a work.
type CrossrefLink struct {
	URL                 string `json:"URL"`
	ContentType         string `json:"content-type,omitempty"`
	ContentVersion      string `json:"content-version,omitempty"`
	IntendedApplication string `json:"intended-application,omitempty"`
}
