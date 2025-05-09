package scholarly

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/crossref"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/openalex"

	"github.com/rs/zerolog/log"
)

// ResolveDOI fetches complete metadata for a DOI from both Crossref and OpenAlex
// and merges them into a single rich record
func ResolveDOI(req ResolveDOIRequest) (*Work, error) {
	if !isValidDOI(req.DOI) {
		return nil, fmt.Errorf("invalid DOI format: %s", req.DOI)
	}

	log.Debug().Str("doi", req.DOI).Msg("Resolving DOI")

	// Get data from OpenAlex
	openAlexData, oaErr := getOpenAlexData(req.DOI)

	// Get data from Crossref
	crossrefData, crErr := getCrossrefData(req.DOI)

	// If both failed, return error
	if oaErr != nil && crErr != nil {
		return nil, fmt.Errorf("failed to resolve DOI from any source: %s, OpenAlex error: %v, Crossref error: %v",
			req.DOI, oaErr, crErr)
	}

	// Merge the data with precedence rules
	mergedWork := mergeWorkData(crossrefData, openAlexData)
	return mergedWork, nil
}

// isValidDOI checks if a string matches the basic DOI format
func isValidDOI(doi string) bool {
	return strings.HasPrefix(doi, "10.") && strings.Contains(doi, "/")
}

// getOpenAlexData fetches metadata from OpenAlex for a given DOI
func getOpenAlexData(doi string) (*Work, error) {
	client := openalex.NewClient("")

	// OpenAlex expects DOIs with a URL prefix
	doiURL := doi
	if !strings.HasPrefix(doi, "https://doi.org/") {
		doiURL = "https://doi.org/" + doi
	}

	// TODO: Implement a direct GetWorkByDOI method in the OpenAlex client
	// For now, simulate with a search query
	params := common.SearchParams{
		Query:      doiURL,
		MaxResults: 1,
	}

	results, err := client.Search(params)
	if err != nil {
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("DOI not found in OpenAlex")
	}

	result := results[0]

	year := 0
	if y, ok := result.Metadata["publication_year"].(int); ok {
		year = y
	}

	isOA := false
	if oa, ok := result.Metadata["is_oa"].(bool); ok {
		isOA = oa
	}

	work := &Work{
		ID:            result.SourceURL,
		DOI:           doi,
		Title:         result.Title,
		Authors:       result.Authors,
		Year:          year,
		IsOA:          isOA,
		CitationCount: result.Citations,
		Abstract:      result.Abstract,
		SourceName:    "openalex",
		PDFURL:        result.PDFURL,
	}

	return work, nil
}

// getCrossrefData fetches metadata from Crossref for a given DOI
func getCrossrefData(doi string) (*Work, error) {
	client := crossref.NewClient("")

	// TODO: Implement a direct GetWorkByDOI method in the Crossref client
	// For now, simulate with a search query
	params := common.SearchParams{
		Query:      "doi:" + doi,
		MaxResults: 1,
	}

	results, err := client.Search(params)
	if err != nil {
		return nil, fmt.Errorf("Crossref error: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("DOI not found in Crossref")
	}

	result := results[0]

	year := 0
	if y, ok := result.Metadata["year"].(int); ok {
		year = y
	}

	citationCount := 0
	if c, ok := result.Metadata["is-referenced-by-count"].(int); ok {
		citationCount = c
	}

	work := &Work{
		ID:            doi,
		DOI:           doi,
		Title:         result.Title,
		Authors:       result.Authors,
		Year:          year,
		CitationCount: citationCount,
		Abstract:      result.Abstract,
		SourceName:    "crossref",
		PDFURL:        result.PDFURL,
	}

	return work, nil
}

// mergeWorkData merges data from Crossref and OpenAlex with precedence rules
func mergeWorkData(crossref, openalex *Work) *Work {
	if crossref == nil && openalex == nil {
		return nil
	}

	if crossref == nil {
		return openalex
	}

	if openalex == nil {
		return crossref
	}

	// Start with Crossref data as base
	merged := *crossref

	// OpenAlex wins for citations
	merged.CitationCount = openalex.CitationCount

	// OpenAlex may have better abstract
	if merged.Abstract == "" && openalex.Abstract != "" {
		merged.Abstract = openalex.Abstract
	}

	// OpenAlex may have PDF URL
	if merged.PDFURL == "" && openalex.PDFURL != "" {
		merged.PDFURL = openalex.PDFURL
	}

	// OpenAlex has OA status info
	merged.IsOA = openalex.IsOA

	// Set combined source
	merged.SourceName = "combined"

	return &merged
}
