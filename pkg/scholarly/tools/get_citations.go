package tools

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/openalex"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/rs/zerolog/log"
)

// XXX - fix the bug with openalex client

// error getting citations: work not found in OpenAlex

//   "work_id": "10.1162/neco.1992.4.2.234",
//   "direction": "refs",
//   "limit": 50,
//   "original_query_intent": "Finding what works Schmidhuber cited in his 1992 history compression paper"
// }

// GetCitations retrieves one hop of the citation graph
func GetCitations(req common.GetCitationsRequest) (*common.GetCitationsResponse, error) {
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

	work, err := client.GetWorkByDOI(doi)
	if err != nil {
		return "", fmt.Errorf("OpenAlex error: %w", err)
	}

	return work.ID, nil
}

// getReferencedWorks gets the outgoing references (works cited by this work)
func getReferencedWorks(workID string, limit int) (*common.GetCitationsResponse, error) {
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
		return &common.GetCitationsResponse{Citations: []common.Citation{}}, nil
	}

	// Limit the number of references to process
	if len(referencedWorks) > limit {
		referencedWorks = referencedWorks[:limit]
	}

	// Get details for each referenced work
	citations := make([]common.Citation, 0, len(referencedWorks))
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

		citations = append(citations, common.Citation{
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

	return &common.GetCitationsResponse{Citations: citations}, nil
}

// getCitedByWorks gets the incoming citations (works that cite this work)
func getCitedByWorks(workID string, limit int) (*common.GetCitationsResponse, error) {
	client := openalex.NewClient("")

	// Get works that cite this work
	results, err := client.GetCitedWorks(workID, limit)
	if err != nil {
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	citations := make([]common.Citation, 0, len(results))
	for _, result := range results {
		citations = append(citations, common.Citation{
			ID:    result.ID,
			DOI:   result.DOI,
			Title: result.DisplayName,
			Year:  result.PublicationYear,
		})
	}

	return &common.GetCitationsResponse{Citations: citations}, nil
}
