package openalex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

	"github.com/rs/zerolog/log"
)

// Client represents an OpenAlex API client
type Client struct {
	BaseURL string
	Mailto  string
}

// NewClient creates a new OpenAlex API client
func NewClient(mailto string) *Client {
	return &Client{
		BaseURL: "https://api.openalex.org/works",
		Mailto:  mailto,
	}
}

// Search searches OpenAlex for papers matching the given parameters
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error) {
	apiParams := url.Values{}

	// Add parameters
	if params.Query != "" {
		apiParams.Add("search", params.Query)
	}

	apiParams.Add("per_page", fmt.Sprintf("%d", params.MaxResults))

	if c.Mailto != "" {
		apiParams.Add("mailto", c.Mailto)
	}

	if filter, ok := params.Filters["filter"]; ok {
		apiParams.Add("filter", filter)
	}

	if sort, ok := params.Filters["sort"]; ok {
		apiParams.Add("sort", sort)
	} else {
		apiParams.Add("sort", "relevance_score:desc")
	}

	// Ensure select fields are valid according to OpenAlex documentation
	apiParams.Add("select", "id,doi,title,display_name,publication_year,publication_date,cited_by_count,authorships,primary_location,open_access,type,concepts,abstract_inverted_index,relevance_score,referenced_works,related_works")

	apiURL := c.BaseURL + "?" + apiParams.Encode()
	log.Debug().Str("url", apiURL).Msg("Requesting OpenAlex API URL")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenAlex API request: %w", err)
	}

	if c.Mailto != "" {
		req.Header.Set("User-Agent", fmt.Sprintf("arxiv-libgen-cli/0.1 (mailto:%s)", c.Mailto))
	} else {
		req.Header.Set("User-Agent", "arxiv-libgen-cli/0.1 (https://github.com/user/repo - please update with actual repo if public)")
	}

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAlex API request failed with status %d", resp.StatusCode)
	}

	var oaResp OpenAlexResponse
	if err := json.Unmarshal(resp.Body, &oaResp); err != nil {
		log.Error().Err(err).Str("body_snippet", string(resp.Body[:min(500, len(resp.Body))])).Msg("Error unmarshalling OpenAlex JSON response")
		return nil, fmt.Errorf("error parsing OpenAlex API response: %w", err)
	}

	if len(oaResp.Results) == 0 {
		return []common.SearchResult{}, nil
	}

	// Process abstracts for all works
	for i := range oaResp.Results {
		oaResp.Results[i].Abstract = reconstructAbstract(oaResp.Results[i].AbstractInvertedIndex)
	}

	return convertToSearchResults(oaResp.Results), nil
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// reconstructAbstract reconstructs the abstract from the inverted index
func reconstructAbstract(invertedIndex map[string][]int) string {
	if invertedIndex == nil || len(invertedIndex) == 0 {
		return ""
	}
	maxLength := 0
	for _, positions := range invertedIndex {
		for _, pos := range positions {
			if pos+1 > maxLength {
				maxLength = pos + 1
			}
		}
	}
	if maxLength == 0 {
		return ""
	}
	orderedWords := make([]string, maxLength)
	for word, positions := range invertedIndex {
		for _, pos := range positions {
			if pos < maxLength {
				orderedWords[pos] = word
			}
		}
	}
	finalWords := []string{}
	for _, w := range orderedWords {
		if w != "" {
			finalWords = append(finalWords, w)
		}
	}
	return strings.Join(finalWords, " ")
}

// convertToSearchResults converts OpenAlex works to the common search result format
func convertToSearchResults(works []OpenAlexWork) []common.SearchResult {
	results := make([]common.SearchResult, len(works))

	for i, work := range works {
		result := common.SearchResult{
			Title:      work.DisplayName,
			DOI:        work.DOI,
			Published:  work.PublicationDate,
			SourceURL:  work.ID,
			Abstract:   work.Abstract,
			Type:       work.Type,
			Citations:  work.CitedByCount,
			SourceName: "openalex",
			Metadata: map[string]interface{}{
				"publication_year": work.PublicationYear,
				"relevance_score":  work.RelevanceScore,
			},
		}

		// Extract authors
		var authors []string
		for _, authorship := range work.Authorships {
			authors = append(authors, authorship.Author.DisplayName)
		}
		result.Authors = authors

		// Get primary location info
		if work.PrimaryLocation != nil {
			result.PDFURL = work.PrimaryLocation.PdfURL
			result.License = work.PrimaryLocation.License

			if work.PrimaryLocation.Source != nil {
				result.JournalInfo = work.PrimaryLocation.Source.DisplayName
			}
		}

		// Set open access status if available
		if work.OpenAccess != nil {
			result.OAStatus = work.OpenAccess.OAStatus

			// If PDF URL not set but OA URL is available, use that
			if result.PDFURL == "" && work.OpenAccess.OAURL != "" {
				result.PDFURL = work.OpenAccess.OAURL
			}
		}

		results[i] = result
	}

	return results
}
