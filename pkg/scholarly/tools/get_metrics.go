package tools

import (
	"fmt"
	"strings"

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

	// Use direct works endpoint
	work, err := client.GetWorkByDOI(workID)
	if err != nil {
		// If we have Crossref data, we can still return some metrics
		if strings.Contains(err.Error(), "not found") {
			// Return metrics with just the data we have
			return &common.Metrics{
				CitationCount:  0,
				CitedByCount:   0,
				ReferenceCount: 0,
				IsOA:           false,
				Altmetrics:     make(map[string]int),
			}, nil
		}
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	// Extract metrics from the result
	metrics := &common.Metrics{
		CitationCount:  work.CitationCount,
		CitedByCount:   work.CitationCount, // Same as citation_count in OpenAlex
		ReferenceCount: 0,                  // Not directly available
		IsOA:           work.IsOA,
		Altmetrics:     make(map[string]int),
	}

	return metrics, nil
}
