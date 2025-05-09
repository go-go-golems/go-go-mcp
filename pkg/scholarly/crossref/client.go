package crossref

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

	"github.com/rs/zerolog/log"
)

// Client represents a Crossref API client
type Client struct {
	BaseURL string
	Mailto  string
}

// NewClient creates a new Crossref API client
func NewClient(mailto string) *Client {
	return &Client{
		BaseURL: "https://api.crossref.org/works",
		Mailto:  mailto,
	}
}

// Search searches Crossref for papers matching the given parameters
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error) {
	apiParams := url.Values{}
	apiParams.Add("query", params.Query)
	apiParams.Add("rows", fmt.Sprintf("%d", params.MaxResults))

	if c.Mailto != "" {
		apiParams.Add("mailto", c.Mailto)
	}

	if filter, ok := params.Filters["filter"]; ok {
		apiParams.Add("filter", filter)
	}

	// Add select for common fields to keep response manageable
	apiParams.Add("select", "DOI,title,author,publisher,type,created,issued,URL,abstract,ISSN,ISBN,subject,link")

	apiURL := c.BaseURL + "?" + apiParams.Encode()
	log.Debug().Str("url", apiURL).Msg("Requesting Crossref API URL")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Crossref API request: %w", err)
	}

	// Set a User-Agent with contact information if provided
	if c.Mailto != "" {
		req.Header.Set("User-Agent", fmt.Sprintf("go-go-mcp-scholarly/0.1 (mailto:%s)", c.Mailto))
	} else {
		req.Header.Set("User-Agent", "go-go-mcp-scholarly/0.1 (https://github.com/go-go-golems/go-go-mcp)")
	}

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crossref API request failed with status %d", resp.StatusCode)
	}

	var crossrefResp CrossrefResponse
	if err := json.Unmarshal(resp.Body, &crossrefResp); err != nil {
		log.Error().Err(err).Bytes("body", resp.Body[:min(500, len(resp.Body))]).Msg("Error unmarshalling Crossref JSON response")
		return nil, fmt.Errorf("error parsing Crossref API response: %w", err)
	}

	if crossrefResp.Status != "ok" || len(crossrefResp.Message.Items) == 0 {
		return []common.SearchResult{}, nil
	}

	return convertToSearchResults(crossrefResp.Message.Items), nil
}

// convertToSearchResults converts Crossref items to the common search result format
func convertToSearchResults(items []CrossrefItem) []common.SearchResult {
	results := make([]common.SearchResult, len(items))

	for i, item := range items {
		result := common.SearchResult{
			DOI:        item.DOI,
			SourceURL:  item.URL,
			SourceName: "crossref",
			Type:       item.Type,
			Metadata: map[string]interface{}{
				"publisher": item.Publisher,
				"issn":      item.ISSN,
				"isbn":      item.ISBN,
				"subject":   item.Subject,
			},
		}

		// Add title - use first title if available
		if len(item.Title) > 0 {
			result.Title = item.Title[0]
		}

		// Add authors
		authors := make([]string, len(item.Author))
		for j, author := range item.Author {
			authors[j] = fmt.Sprintf("%s %s", author.Given, author.Family)
		}
		result.Authors = authors

		// Add publication date - prefer issued date over created date
		if len(item.Issued.DateParts) > 0 && len(item.Issued.DateParts[0]) > 0 {
			dateStr := fmt.Sprintf("%d", item.Issued.DateParts[0][0])
			if len(item.Issued.DateParts[0]) > 1 {
				dateStr += fmt.Sprintf("-%02d", item.Issued.DateParts[0][1])
			}
			if len(item.Issued.DateParts[0]) > 2 {
				dateStr += fmt.Sprintf("-%02d", item.Issued.DateParts[0][2])
			}
			result.Published = dateStr
		} else if item.Created.DateTime != "" {
			result.Published = item.Created.DateTime
		}

		// Clean abstract if present
		if item.Abstract != "" {
			// Clean up abstract, remove potential HTML tags (simple replace)
			cleanAbstract := strings.ReplaceAll(item.Abstract, "<jats:p>", "")
			cleanAbstract = strings.ReplaceAll(cleanAbstract, "</jats:p>", "")
			cleanAbstract = strings.ReplaceAll(cleanAbstract, "<title>", "")
			cleanAbstract = strings.ReplaceAll(cleanAbstract, "</title>", "")
			result.Abstract = strings.Join(strings.Fields(cleanAbstract), " ")
		}

		// Look for PDF links
		for _, link := range item.Link {
			if strings.Contains(strings.ToLower(link.ContentType), "pdf") {
				result.PDFURL = link.URL
				break
			}
		}

		results[i] = result
	}

	return results
}
