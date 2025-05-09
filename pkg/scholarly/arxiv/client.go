package arxiv

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

	"github.com/rs/zerolog/log"
)

// Client represents an Arxiv API client
type Client struct {
	BaseURL string
}

// NewClient creates a new Arxiv API client
func NewClient() *Client {
	return &Client{
		BaseURL: "http://export.arxiv.org/api/query",
	}
}

// Search searches Arxiv for papers matching the given parameters
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error) {
	apiParams := url.Values{}
	apiParams.Add("search_query", params.Query)
	apiParams.Add("max_results", fmt.Sprintf("%d", params.MaxResults))
	apiParams.Add("sortBy", "relevance") // Default sort order

	apiURL := c.BaseURL + "?" + apiParams.Encode()
	log.Debug().Str("url", apiURL).Msg("Requesting Arxiv API URL")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Arxiv API request: %w", err)
	}

	// Set a User-Agent
	req.Header.Set("User-Agent", "arxiv-libgen-cli/0.1 (https://github.com/user/repo - please update with actual repo if public)")

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arxiv API request failed with status %d", resp.StatusCode)
	}

	var feed AtomFeed
	if err := xml.Unmarshal(resp.Body, &feed); err != nil {
		log.Error().Err(err).Str("body_prefix", string(resp.Body[:min(500, len(resp.Body))])).Msg("Error unmarshalling Arxiv XML response") // Log prefix of body
		return nil, fmt.Errorf("error parsing Arxiv API response: %w", err)
	}

	if len(feed.Entries) == 0 {
		return []common.SearchResult{}, nil
	}

	log.Debug().Int("total_results_api", feed.TotalResults).Int("entries_in_page", len(feed.Entries)).Msg("Arxiv results parsed")

	return convertToSearchResults(feed.Entries), nil
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// convertToSearchResults converts Arxiv entries to the common search result format
func convertToSearchResults(entries []Entry) []common.SearchResult {
	results := make([]common.SearchResult, len(entries))

	for i, entry := range entries {
		result := common.SearchResult{
			Title:      strings.Join(strings.Fields(entry.Title), " "), // Clean up whitespace
			Abstract:   strings.Join(strings.Fields(entry.Summary), " "), // Clean up whitespace
			Published:  entry.Published,
			DOI:        entry.DOI,
			SourceURL:  entry.ID,
			SourceName: "arxiv",
			OAStatus:   "green", // Arxiv is considered green OA
			License:    "", // Arxiv doesn't provide license info via API
			JournalInfo: entry.JournalRef,
			Metadata: map[string]interface{}{
				"updated": entry.Updated,
				"comment": entry.Comment,
				"primary_category": entry.PrimaryCategory.Term,
			},
		}

		// Add authors
		authors := make([]string, len(entry.Authors))
		for j, author := range entry.Authors {
			authors[j] = author.Name
		}
		result.Authors = authors

		// Find PDF link
		for _, link := range entry.Link {
			if link.Title == "pdf" {
				result.PDFURL = link.Href
				break
			}
		}

		results[i] = result
	}

	return results
}