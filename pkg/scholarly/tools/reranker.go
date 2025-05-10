package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/pkg/errors"
)

const (
	// DefaultRerankerURL is the default URL for the reranker service
	DefaultRerankerURL = "http://localhost:8000/rerank"
	// DefaultRerankerTimeout is the default timeout for reranker requests
	DefaultRerankerTimeout = 10 * time.Second
)

// RerankerClient provides methods to interact with the reranker service
type RerankerClient struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewRerankerClient creates a new client for the reranker service
func NewRerankerClient(url string, timeout time.Duration) *RerankerClient {
	if url == "" {
		url = DefaultRerankerURL
	}
	if timeout == 0 {
		timeout = DefaultRerankerTimeout
	}
	return &RerankerClient{
		URL:     url,
		Timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// RerankerRequest represents a request to the reranker service
type RerankerRequest struct {
	Query   string       `json:"query"`
	Results []ArxivPaper `json:"results"`
	TopN    int          `json:"top_n,omitempty"`
}

// ArxivPaper represents a paper in the reranker request
type ArxivPaper struct {
	Title       string                 `json:"Title"`
	Authors     []string               `json:"Authors"`
	Abstract    string                 `json:"Abstract"`
	Published   string                 `json:"Published"`
	DOI         string                 `json:"DOI,omitempty"`
	PDFURL      string                 `json:"PDFURL,omitempty"`
	SourceURL   string                 `json:"SourceURL,omitempty"`
	SourceName  string                 `json:"SourceName,omitempty"`
	OAStatus    string                 `json:"OAStatus,omitempty"`
	License     string                 `json:"License,omitempty"`
	FileSize    string                 `json:"FileSize,omitempty"`
	Citations   int                    `json:"Citations,omitempty"`
	Type        string                 `json:"Type,omitempty"`
	JournalInfo string                 `json:"JournalInfo,omitempty"`
	Metadata    map[string]interface{} `json:"Metadata,omitempty"`
}

// ScoredPaper represents a paper with a reranker score
type ScoredPaper struct {
	ArxivPaper
	Score float64 `json:"score"`
}

// RerankerResponse represents a response from the reranker service
type RerankerResponse struct {
	Query           string        `json:"query"`
	RerankedResults []ScoredPaper `json:"reranked_results"`
}

// Rerank sends a reranking request to the reranker service
func (c *RerankerClient) Rerank(ctx context.Context, query string, results []common.SearchResult, topN int) ([]common.SearchResult, error) {
	// Convert SearchResults to ArxivPapers
	papers := make([]ArxivPaper, len(results))
	for i, result := range results {
		papers[i] = ArxivPaper{
			Title:       result.Title,
			Authors:     result.Authors,
			Abstract:    result.Abstract,
			Published:   result.Published,
			DOI:         result.DOI,
			PDFURL:      result.PDFURL,
			SourceURL:   result.SourceURL,
			SourceName:  result.SourceName,
			OAStatus:    result.OAStatus,
			License:     result.License,
			FileSize:    result.FileSize,
			Citations:   result.Citations,
			Type:        result.Type,
			JournalInfo: result.JournalInfo,
			Metadata:    result.Metadata,
		}
	}

	// Create the request payload
	reqPayload := RerankerRequest{
		Query:   query,
		Results: papers,
		TopN:    topN,
	}

	// Marshal the request payload
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal reranker request")
	}
	
	// Log the request payload for debugging
	fmt.Printf("Reranker request payload: %s\n", string(reqBody))

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reranker request")
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send reranker request")
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reranker returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var rerankerResp RerankerResponse
	if err := json.NewDecoder(resp.Body).Decode(&rerankerResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode reranker response")
	}

	// Create a map to look up original results by title and abstract
	originalResultsMap := make(map[string]common.SearchResult)
	for _, result := range results {
		// Create a compound key using title and abstract as they should be unique together
		key := result.Title + "::" + result.Abstract
		originalResultsMap[key] = result
	}

	// Convert ScoredPapers back to SearchResults with reranking information
	rerankedResults := make([]common.SearchResult, len(rerankerResp.RerankedResults))
	for i, scoredPaper := range rerankerResp.RerankedResults {
		// Find the original result and its index
		key := scoredPaper.Title + "::" + scoredPaper.Abstract
		originalResult, ok := originalResultsMap[key]
		if !ok {
			return nil, fmt.Errorf("could not find original result for reranked paper: %s", scoredPaper.Title)
		}

		// Find the original index
		originalIndex := -1
		for j, result := range results {
			if result.Title == scoredPaper.Title && result.Abstract == scoredPaper.Abstract {
				originalIndex = j
				break
			}
		}

		// Create a new search result with reranking information
		rerankedResults[i] = originalResult
		rerankedResults[i].Reranked = true
		rerankedResults[i].RerankerScore = scoredPaper.Score
		rerankedResults[i].OriginalIndex = originalIndex
	}

	return rerankedResults, nil
}

// IsRerankerAvailable checks if the reranker service is available
func (c *RerankerClient) IsRerankerAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8000/models", nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
