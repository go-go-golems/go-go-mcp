package tools

import (
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/openalex"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/rs/zerolog/log"
)

// GetMetrics retrieves quantitative metrics for a work
func GetMetrics(req common.GetMetricsRequest) (*common.Metrics, error) {
	if req.WorkID == "" {
		return nil, fmt.Errorf("work_id cannot be empty")
	}

	log.Debug().Str("work_id", req.WorkID).Msg("Getting metrics")

	// Check if the ID is a DOI
	if isValidDOI(req.WorkID) {
		return getMetricsByDOI(req.WorkID)
	}

	// Otherwise, assume it's an OpenAlex ID
	return getMetricsByOpenAlexID(req.WorkID)
}

// getMetricsByDOI gets metrics for a work identified by DOI
func getMetricsByDOI(doi string) (*common.Metrics, error) {
	// Resolve the DOI to get the OpenAlex ID
	resolveReq := common.ResolveDOIRequest{
		DOI: doi,
	}

	work, err := ResolveDOI(resolveReq)
	if err != nil {
		return nil, fmt.Errorf("error resolving DOI: %w", err)
	}

	// Now get metrics by OpenAlex ID
	return getMetricsByOpenAlexID(work.ID)
}

// getMetricsByOpenAlexID gets metrics for a work identified by OpenAlex ID
func getMetricsByOpenAlexID(workID string) (*common.Metrics, error) {
	client := openalex.NewClient("")

	// TODO: Implement a direct GetWork method in the OpenAlex client
	// For now, simulate with a search query
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

	// Extract metrics from the result
	metrics := &common.Metrics{
		CitationCount:  result.Citations,
		CitedByCount:   result.Citations, // Same as citation_count in OpenAlex
		ReferenceCount: 0,                // Not directly available
		IsOA:           false,
		Altmetrics:     make(map[string]int),
	}

	// Extract OA status
	if oa, ok := result.Metadata["is_oa"].(bool); ok {
		metrics.IsOA = oa
	}

	if result.OAStatus != "" {
		metrics.OAStatus = result.OAStatus
	}

	// Try to extract reference count if available
	if refCount, ok := result.Metadata["referenced_works_count"].(int); ok {
		metrics.ReferenceCount = refCount
	}

	return metrics, nil
}
