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
		BaseURL: "https://api.openalex.org",
		Mailto:  mailto,
	}
}

// searchAuthor searches for an author by name and returns their OpenAlex ID
func (c *Client) searchAuthor(name string) (string, error) {
	apiParams := url.Values{}
	apiParams.Add("search", name)
	apiParams.Add("per_page", "1") // We just need the top match

	if c.Mailto != "" {
		apiParams.Add("mailto", c.Mailto)
	}

	apiURL := fmt.Sprintf("%s/authors?%s", c.BaseURL, apiParams.Encode())
	log.Debug().Str("url", apiURL).Msg("Requesting OpenAlex API URL for author search")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating OpenAlex API request: %w", err)
	}

	if c.Mailto != "" {
		req.Header.Set("User-Agent", fmt.Sprintf("go-go-mcp/0.1 (mailto:%s)", c.Mailto))
	} else {
		req.Header.Set("User-Agent", "go-go-mcp/0.1 (https://github.com/go-go-golems/go-go-mcp)")
	}

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return "", resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openAlex API request failed with status %d", resp.StatusCode)
	}

	var authorResp struct {
		Results []struct {
			ID string `json:"id"`
		} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body, &authorResp); err != nil {
		return "", fmt.Errorf("error parsing OpenAlex API response: %w", err)
	}

	if len(authorResp.Results) == 0 {
		return "", nil
	}

	return authorResp.Results[0].ID, nil
}

// Search searches OpenAlex for papers matching the given parameters
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error) {
	apiParams := url.Values{}

	// Add parameters
	if params.Query != "" {
		apiParams.Add("search", params.Query)
	}

	apiParams.Add("per_page", fmt.Sprintf("%d", params.MaxResults))

	// Build filter string
	var filters []string

	// Handle author filter - first try to get the author ID
	if author, ok := params.Filters["author"]; ok {
		authorID, err := c.searchAuthor(author)
		if err != nil {
			log.Warn().Err(err).Str("author", author).Msg("Failed to get author ID, falling back to display_name.search")
			filters = append(filters, fmt.Sprintf("display_name.search:%s", url.QueryEscape(author)))
		} else if authorID != "" {
			log.Debug().Str("author", author).Str("id", authorID).Msg("Found author ID")
			filters = append(filters, fmt.Sprintf("author.id:%s", strings.TrimPrefix(authorID, "https://openalex.org/")))
		} else {
			log.Debug().Str("author", author).Msg("No author ID found, falling back to display_name.search")
			filters = append(filters, fmt.Sprintf("display_name.search:%s", url.QueryEscape(author)))
		}
	}

	// Handle publication year filter
	if fromYear, ok := params.Filters["from-publication_year"]; ok {
		filters = append(filters, fmt.Sprintf("publication_year:%s", fromYear))
	}

	// Add any other filters from the params
	if filter, ok := params.Filters["filter"]; ok {
		filters = append(filters, filter)
	}

	// Combine all filters with proper URL encoding
	if len(filters) > 0 {
		apiParams.Set("filter", strings.Join(filters, ","))
	}

	// Determine sort order
	sortOrder := "relevance_score:desc" // Default assumption
	if userSpecifiedSort, ok := params.Filters["sort"]; ok {
		sortOrder = userSpecifiedSort
	}

	// If there is no search query, and the sort is by relevance_score,
	// OpenAlex will return an error.
	// In this case, change sort to 'cited_by_count:desc' as a sensible default,
	// which is also OpenAlex's default for non-search queries.
	if params.Query == "" && strings.HasPrefix(sortOrder, "relevance_score") {
		sortOrder = "cited_by_count:desc"
		log.Debug().Msg("No search query and sort was relevance_score; changed sort to cited_by_count:desc")
	}
	apiParams.Add("sort", sortOrder)

	// Add mailto parameter for polite pool
	if c.Mailto != "" {
		apiParams.Add("mailto", c.Mailto)
	}

	// Ensure select fields are valid according to OpenAlex documentation
	apiParams.Add("select", "id,doi,title,display_name,publication_year,publication_date,cited_by_count,authorships,primary_location,open_access,type,concepts,abstract_inverted_index,relevance_score,referenced_works,related_works")

	// Build the URL with proper encoding
	apiURL := fmt.Sprintf("%s/works?%s", c.BaseURL, apiParams.Encode())
	log.Debug().Str("url", apiURL).Msg("Requesting OpenAlex API URL")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenAlex API request: %w", err)
	}

	if c.Mailto != "" {
		req.Header.Set("User-Agent", fmt.Sprintf("go-go-mcp/0.1 (mailto:%s)", c.Mailto))
	} else {
		req.Header.Set("User-Agent", "go-go-mcp/0.1 (https://github.com/go-go-golems/go-go-mcp)")
	}

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAlex API request failed with status %d", resp.StatusCode)
	}

	// Log response body for debugging
	log.Debug().Str("response_body", string(resp.Body[:min(2000, len(resp.Body))])).Msg("OpenAlex API raw response")

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

// reconstructAbstract reconstructs the abstract from the inverted index
func reconstructAbstract(invertedIndex map[string][]int) string {
	if len(invertedIndex) == 0 {
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
	// Log the number of results and a representative sample
	if len(works) > 0 {
		sampleWork := works[0]
		log.Debug().Int("total_works", len(works)).Str("first_work_title", sampleWork.DisplayName).Str("first_work_id", sampleWork.ID).Int("first_work_citations", sampleWork.CitedByCount).Msg("OpenAlex parsed works sample")
	}
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
