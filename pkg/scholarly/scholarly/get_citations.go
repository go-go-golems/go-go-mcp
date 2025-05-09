package scholarly

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/openalex"

	"github.com/rs/zerolog/log"
)

// GetCitations retrieves one hop of the citation graph
func GetCitations(req GetCitationsRequest) (*GetCitationsResponse, error) {
	if req.WorkID == "" {
		return nil, fmt.Errorf("work_id cannot be empty")
	}

	if req.Direction != "refs" && req.Direction != "cited_by" {
		return nil, fmt.Errorf("direction must be either 'refs' or 'cited_by'")
	}

	if req.Limit <= 0 {
		req.Limit = 100 // Default value as per spec
	} else if req.Limit > 200 {
		req.Limit = 200 // Maximum value as per spec
	}

	log.Debug().Str("work_id", req.WorkID).Str("direction", req.Direction).Int("limit", req.Limit).Msg("Getting citations")

	// If the work ID is a DOI, convert it to an OpenAlex ID first
	workID := req.WorkID
	if isValidDOI(workID) {
		oaID, err := getOpenAlexIDFromDOI(workID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve DOI to OpenAlex ID: %w", err)
		}
		workID = oaID
	}

	// Get citations based on direction
	if req.Direction == "refs" {
		return getReferencedWorks(workID, req.Limit)
	} else {
		return getCitedByWorks(workID, req.Limit)
	}
}

// getOpenAlexIDFromDOI resolves a DOI to an OpenAlex ID
func getOpenAlexIDFromDOI(doi string) (string, error) {
	client := openalex.NewClient("")

	// OpenAlex expects DOIs with a URL prefix
	doiURL := doi
	if !strings.HasPrefix(doi, "https://doi.org/") {
		doiURL = "https://doi.org/" + doi
	}

	// Search for the DOI
	params := common.SearchParams{
		Query:      doiURL,
		MaxResults: 1,
	}

	results, err := client.Search(params)
	if err != nil {
		return "", fmt.Errorf("OpenAlex error: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("DOI not found in OpenAlex")
	}

	return results[0].SourceURL, nil
}

// getReferencedWorks gets the outgoing references (works cited by this work)
func getReferencedWorks(workID string, limit int) (*GetCitationsResponse, error) {
	// First, get the work itself to access its referenced_works
	client := openalex.NewClient("")

	params := common.SearchParams{
		Query:      fmt.Sprintf("id:%s", workID),
		MaxResults: 1,
	}

	results, err := client.Search(params)
	if err != nil {
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("work not found in OpenAlex")
	}

	result := results[0]

	// Extract referenced works from metadata
	referencedWorks, ok := result.Metadata["referenced_works"].([]interface{})
	if !ok || len(referencedWorks) == 0 {
		return &GetCitationsResponse{Citations: []Citation{}}, nil
	}

	// Limit the number of references to process
	if len(referencedWorks) > limit {
		referencedWorks = referencedWorks[:limit]
	}

	// Get details for each referenced work
	citations := make([]Citation, 0, len(referencedWorks))
	for _, refID := range referencedWorks {
		refIDStr, ok := refID.(string)
		if !ok {
			continue
		}

		// Get basic info about the referenced work
		refParam := common.SearchParams{
			Query:      fmt.Sprintf("id:%s", refIDStr),
			MaxResults: 1,
		}

		refResults, err := client.Search(refParam)
		if err != nil || len(refResults) == 0 {
			// Skip works we can't resolve
			continue
		}

		refResult := refResults[0]

		year := 0
		if y, ok := refResult.Metadata["publication_year"].(int); ok {
			year = y
		}

		citations = append(citations, Citation{
			ID:    refResult.SourceURL,
			DOI:   refResult.DOI,
			Title: refResult.Title,
			Year:  year,
		})

		// Respect the limit
		if len(citations) >= limit {
			break
		}
	}

	return &GetCitationsResponse{Citations: citations}, nil
}

// getCitedByWorks gets the incoming citations (works that cite this work)
func getCitedByWorks(workID string, limit int) (*GetCitationsResponse, error) {
	client := openalex.NewClient("")

	// Use filter to get works that cite the given work
	params := common.SearchParams{
		Query:      "*", // Using wildcard query since empty query with sort is not allowed
		MaxResults: limit,
		Filters: map[string]string{
			"filter": fmt.Sprintf("cites:%s", workID),
			"sort":   "cited_by_count:desc", // Sort by citation count instead of relevance
		},
	}

	results, err := client.Search(params)
	if err != nil {
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	citations := make([]Citation, 0, len(results))
	for _, result := range results {
		year := 0
		if y, ok := result.Metadata["publication_year"].(int); ok {
			year = y
		}

		citations = append(citations, Citation{
			ID:    result.SourceURL,
			DOI:   result.DOI,
			Title: result.Title,
			Year:  year,
		})
	}

	// TODO: Implement cursor-based pagination if more results exist
	// This would require modifying the OpenAlex client to return cursor tokens

	return &GetCitationsResponse{Citations: citations}, nil
}
