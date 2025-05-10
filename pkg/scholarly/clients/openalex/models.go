package openalex

// OpenAlexResponse represents the top-level structure of the OpenAlex API response for works.
type OpenAlexResponse struct {
	Meta    OpenAlexMeta   `json:"meta"`
	Results []OpenAlexWork `json:"results"`
	GroupBy []interface{}  `json:"group_by"`
}

// OpenAlexMeta contains metadata about the response.
type OpenAlexMeta struct {
	Count            int    `json:"count"`
	DBResponseTimeMs int    `json:"db_response_time_ms"`
	Page             int    `json:"page"`
	PerPage          int    `json:"per_page"`
	NextCursor       string `json:"next_cursor,omitempty"`
}

// OpenAlexAuthorship links authors to works.
type OpenAlexAuthorship struct {
	AuthorPosition       string                `json:"author_position"`
	Author               OpenAlexAuthor        `json:"author"`
	Institutions         []OpenAlexInstitution `json:"institutions,omitempty"`
	RawAffiliationString string                `json:"raw_affiliation_string,omitempty"`
}

// OpenAlexAuthor represents an author.
type OpenAlexAuthor struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Orcid       string `json:"orcid,omitempty"`
}

// OpenAlexInstitution represents an institution.
type OpenAlexInstitution struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Ror         string `json:"ror,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Type        string `json:"type,omitempty"`
}

// OpenAlexOpenAccess contains OA details.
type OpenAlexOpenAccess struct {
	IsOA                     bool   `json:"is_oa"`
	OAStatus                 string `json:"oa_status"`
	OAURL                    string `json:"oa_url,omitempty"`
	AnyRepositoryHasFulltext bool   `json:"any_repository_has_fulltext,omitempty"`
}

// OpenAlexSourceInfo is for the "source" object nested within primary_location.
type OpenAlexSourceInfo struct {
	ID                   string   `json:"id,omitempty"`
	DisplayName          string   `json:"display_name,omitempty"`
	IssnL                string   `json:"issn_l,omitempty"`
	Issn                 []string `json:"issn,omitempty"`
	IsOA                 bool     `json:"is_oa,omitempty"`
	IsInDoaj             bool     `json:"is_in_doaj,omitempty"`
	HostOrganizationName string   `json:"host_organization_name,omitempty"`
	Type                 string   `json:"type,omitempty"`
	Publisher            string   `json:"publisher,omitempty"`
}

// OpenAlexPrimaryLocation represents the primary_location object.
type OpenAlexPrimaryLocation struct {
	IsOA           bool                `json:"is_oa"`
	LandingPageURL string              `json:"landing_page_url,omitempty"`
	PdfURL         string              `json:"pdf_url,omitempty"`
	Source         *OpenAlexSourceInfo `json:"source,omitempty"`
	License        string              `json:"license,omitempty"`
	Version        string              `json:"version,omitempty"`
}

// OpenAlexConcept represents a concept associated with a work.
type OpenAlexConcept struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Level       int     `json:"level"`
	Score       float32 `json:"score"`
}

// OpenAlexWork represents a single scholarly work.
type OpenAlexWork struct {
	ID                    string                   `json:"id"`
	DOI                   string                   `json:"doi"`
	Title                 string                   `json:"title"`
	DisplayName           string                   `json:"display_name"`
	PublicationYear       int                      `json:"publication_year"`
	PublicationDate       string                   `json:"publication_date"`
	CitedByCount          int                      `json:"cited_by_count"`
	Authorships           []OpenAlexAuthorship     `json:"authorships"`
	PrimaryLocation       *OpenAlexPrimaryLocation `json:"primary_location,omitempty"`
	OpenAccess            *OpenAlexOpenAccess      `json:"open_access,omitempty"`
	Type                  string                   `json:"type"`
	AbstractInvertedIndex map[string][]int         `json:"abstract_inverted_index,omitempty"`
	Abstract              string                   // This is reconstructed, not directly from JSON
	Concepts              []OpenAlexConcept        `json:"concepts,omitempty"`
	ReferencedWorks       []string                 `json:"referenced_works,omitempty"`
	RelatedWorks          []string                 `json:"related_works,omitempty"`
	RelevanceScore        float64                  `json:"relevance_score,omitempty"`
}
