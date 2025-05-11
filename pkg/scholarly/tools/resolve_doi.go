package tools

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/crossref"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/openalex"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/rs/zerolog/log"
)

// ResolveDOI fetches complete metadata for a DOI from both Crossref and OpenAlex
// and merges them into a single rich record
func ResolveDOI(req common.ResolveDOIRequest) (*common.Work, error) {
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
func getOpenAlexData(doi string) (*common.Work, error) {
	client := openalex.NewClient("")

	log.Debug().Str("doi", doi).Msg("Requesting work from OpenAlex")

	work, err := client.GetWorkByDOI(doi)
	if err != nil {
		log.Error().Err(err).Str("doi", doi).Msg("Failed to get data from OpenAlex")
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	log.Debug().Str("doi", doi).Str("title", work.Title).Msg("Successfully found work in OpenAlex")
	return work, nil
}

// getCrossrefData fetches metadata from Crossref for a given DOI
func getCrossrefData(doi string) (*common.Work, error) {
	client := crossref.NewClient("")

	log.Debug().Str("doi", doi).Msg("Requesting work from Crossref")

	// TODO: Implement a direct GetWorkByDOI method in the Crossref client
	// For now, simulate with a search query
	params := common.SearchParams{
		Query:      "doi:" + doi,
		MaxResults: 1,
	}

	results, err := client.Search(params)
	if err != nil {
		log.Error().Err(err).Str("doi", doi).Msg("Failed to get data from Crossref")
		return nil, fmt.Errorf("crossref error: %w", err)
	}

	if len(results) == 0 {
		log.Warn().Str("doi", doi).Msg("DOI not found in Crossref")
		return nil, fmt.Errorf("DOI not found in Crossref")
	}

	log.Debug().Str("doi", doi).Str("title", results[0].Title).Msg("Successfully found work in Crossref")
	result := results[0]

	year := 0
	if y, ok := result.Metadata["year"].(int); ok {
		year = y
	}

	citationCount := 0
	if c, ok := result.Metadata["is-referenced-by-count"].(int); ok {
		citationCount = c
	}

	work := &common.Work{
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
func mergeWorkData(crossref, openalex *common.Work) *common.Work {
	log.Debug().Bool("has_crossref", crossref != nil).Bool("has_openalex", openalex != nil).Msg("Merging work data from sources")

	if crossref == nil && openalex == nil {
		log.Warn().Msg("No data available from any source to merge")
		return nil
	}

	if crossref == nil {
		log.Debug().Msg("Using OpenAlex data only (Crossref data missing)")
		return openalex
	}

	if openalex == nil {
		log.Debug().Msg("Using Crossref data only (OpenAlex data missing)")
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

	// Log the merged result details
	log.Debug().Str("doi", merged.DOI).Str("title", merged.Title).Int("citation_count", merged.CitationCount).
		Bool("has_abstract", merged.Abstract != "").Bool("has_pdf", merged.PDFURL != "").Bool("is_oa", merged.IsOA).
		Msg("Successfully merged work data")

	return &merged
}
