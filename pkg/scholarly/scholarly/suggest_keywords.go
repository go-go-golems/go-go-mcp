package scholarly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

	"github.com/rs/zerolog/log"
)

// OpenAlexTextRequest represents the request body for OpenAlex text-aboutness
type OpenAlexTextRequest struct {
	Title string `json:"title"`
}

// OpenAlexConceptResponse represents the response from OpenAlex text-aboutness
type OpenAlexConceptResponse struct {
	Concepts []struct {
		ID          string  `json:"id"`
		DisplayName string  `json:"display_name"`
		Score       float64 `json:"score"`
	} `json:"concepts"`
}

// SuggestKeywords suggests keywords for the given text using OpenAlex's text-aboutness endpoint
func SuggestKeywords(req SuggestKeywordsRequest) (*SuggestKeywordsResponse, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	if req.MaxKeywords <= 0 {
		req.MaxKeywords = 10 // Default value
	}

	log.Debug().Str("text_sample", truncateString(req.Text, 50)).Int("max_keywords", req.MaxKeywords).Msg("Suggesting keywords")

	// Call OpenAlex text-aboutness endpoint
	keywords, err := getKeywordsFromOpenAlex(req.Text, req.MaxKeywords)
	if err != nil {
		return nil, err
	}

	response := &SuggestKeywordsResponse{
		Keywords: keywords,
	}

	return response, nil
}

// getKeywordsFromOpenAlex calls the OpenAlex text-aboutness endpoint to get keywords
func getKeywordsFromOpenAlex(text string, maxKeywords int) ([]Keyword, error) {
	// OpenAlex text-aboutness endpoint with query parameters
	url := fmt.Sprintf("https://api.openalex.org/text?title=%s", url.QueryEscape(text))

	// Make a GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "arxiv-libgen-cli/0.1")

	// Make the request
	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAlex API returned status code %d", resp.StatusCode)
	}

	// Parse the response
	var openAlexResp OpenAlexConceptResponse
	if err := json.Unmarshal(resp.Body, &openAlexResp); err != nil {
		return nil, fmt.Errorf("error parsing OpenAlex response: %w", err)
	}

	// Convert to our keyword format
	keywords := make([]Keyword, 0, len(openAlexResp.Concepts))
	for _, concept := range openAlexResp.Concepts {
		keywords = append(keywords, Keyword{
			ID:          concept.ID,
			DisplayName: concept.DisplayName,
			Relevance:   concept.Score,
		})

		// Limit to max keywords
		if len(keywords) >= maxKeywords {
			break
		}
	}

	return keywords, nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
