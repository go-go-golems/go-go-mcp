package libgen

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

// Client represents a LibGen scraping client
type Client struct {
	BaseURL string
}

// NewClient creates a new LibGen client
func NewClient(mirrorURL string) *Client {
	// Ensure URL has protocol
	effectiveMirror := mirrorURL
	if !strings.HasPrefix(effectiveMirror, "http://") && !strings.HasPrefix(effectiveMirror, "https://") {
		effectiveMirror = "https://" + effectiveMirror
	}

	return &Client{
		BaseURL: effectiveMirror,
	}
}

// Search searches LibGen for papers matching the given parameters
func (c *Client) Search(params common.SearchParams) ([]common.SearchResult, error) {
	// Parse base URL
	baseDomainURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid mirror URL parsing for base domain: %w", err)
	}

	// Construct the actual API URL
	apiURLObject := &url.URL{
		Scheme: baseDomainURL.Scheme,
		Host:   baseDomainURL.Host,
		Path:   "/scimag/", // Path for scientific articles on mirrors like libgen.is
	}

	qParams := apiURLObject.Query()
	qParams.Set("q", params.Query)
	apiURLObject.RawQuery = qParams.Encode()

	apiURL := apiURLObject.String()

	log.Debug().Str("url", apiURL).Msg("Requesting LibGen URL")
	time.Sleep(1 * time.Second) // Be nice to the server

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("libgen mirror request failed with status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML response from LibGen: %w", err)
	}
	log.Debug().Msg("LibGen HTML response parsed successfully")

	entries, err := parseLibGenResults(doc, apiURLObject, params.MaxResults)
	if err != nil {
		return nil, err
	}

	return convertToSearchResults(entries), nil
}

// parseLibGenResults parses the LibGen HTML page for search results
func parseLibGenResults(doc *goquery.Document, baseURL *url.URL, maxResults int) ([]LibgenEntry, error) {
	var entries []LibgenEntry
	log.Debug().Msg("Attempting to find result rows with selector: table.catalog tbody tr")
	resultRows := doc.Find("table.catalog tbody tr")
	log.Debug().Int("num_rows_found_by_selector", resultRows.Length()).Msg("Initial rows found by selector")

	resultRows.Each(func(i int, row *goquery.Selection) {
		if len(entries) >= maxResults {
			return
		}
		log.Debug().Int("row_index", i).Msg("Processing result row")

		cells := row.Find("td")
		var entry LibgenEntry

		if cells.Length() == 5 {
			entry.Authors = strings.TrimSpace(cells.Eq(0).Text())

			articleCell := cells.Eq(1)
			titleLink := articleCell.Find("p a").First()
			entry.Title = strings.TrimSpace(titleLink.Text())
			if href, exists := titleLink.Attr("href"); exists {
				parsedHref, _ := url.Parse(href)
				entry.DownloadURL = baseURL.ResolveReference(parsedHref).String()
			}
			doiText := articleCell.Find("p").Eq(1).Text()
			if strings.HasPrefix(strings.ToLower(doiText), "doi:") {
				entry.DOI = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(doiText), "doi:"))
			} else {
				doiRegex := regexp.MustCompile(`10\.\d{4,9}/[-._;()/:A-Z0-9a-z]+`)
				foundDOI := doiRegex.FindString(doiText)
				if foundDOI != "" {
					entry.DOI = foundDOI
				}
			}

			journalCell := cells.Eq(2)
			var journalParts []string
			journalCell.Find("p").Each(func(_ int, p *goquery.Selection) {
				journalParts = append(journalParts, strings.TrimSpace(p.Text()))
			})
			entry.JournalInfo = strings.Join(journalParts, "; ")

			fileCell := cells.Eq(3)
			entry.FileSize = strings.TrimSpace(fileCell.Contents().Not("a").Not("br").Text())
			if editLink, exists := fileCell.Find("a[title=\"edit metadata\"]").Attr("href"); exists {
				parsedEditLink, _ := url.Parse(editLink)
				entry.EditLink = baseURL.ResolveReference(parsedEditLink).String()
			}

			mirrorsCell := cells.Eq(4)
			mirrorsCell.Find("ul.record_mirrors li a").Each(func(_ int, link *goquery.Selection) {
				if href, exists := link.Attr("href"); exists {
					parsedMirrorLink, err := url.Parse(href)
					if err == nil {
						resolvedHref := baseURL.ResolveReference(parsedMirrorLink).String()
						entry.MirrorLinks = append(entry.MirrorLinks, resolvedHref)
					}
				}
			})

			if entry.Title != "" {
				entries = append(entries, entry)
				log.Debug().Str("title", entry.Title).Msg("Parsed entry (libgen.is/scimag structure)")
			} else {
				log.Warn().Int("row_index", i).Msg("Skipping row, could not parse title (libgen.is/scimag structure)")
			}
		} else {
			log.Warn().Int("row_index", i).Int("num_cells", cells.Length()).Msg("Skipping row due to unexpected cell count for libgen.is/scimag structure")
		}
	})

	if len(entries) == 0 {
		return nil, fmt.Errorf("no results found, or failed to parse results from the mirror")
	}

	return entries, nil
}

// convertToSearchResults converts LibGen entries to the common search result format
func convertToSearchResults(entries []LibgenEntry) []common.SearchResult {
	results := make([]common.SearchResult, len(entries))

	for i, entry := range entries {
		result := common.SearchResult{
			Title:       entry.Title,
			Published:   "", // Not available directly
			DOI:         entry.DOI,
			SourceURL:   entry.DownloadURL,
			SourceName:  "libgen",
			JournalInfo: entry.JournalInfo,
			FileSize:    entry.FileSize,
			Metadata: map[string]interface{}{
				"edit_link":    entry.EditLink,
				"mirror_links": entry.MirrorLinks,
			},
		}

		// Split authors string into array
		if entry.Authors != "" {
			result.Authors = []string{entry.Authors} // Ideally we'd parse this better
		}

		// Use first mirror link as PDF URL if available
		if len(entry.MirrorLinks) > 0 {
			result.PDFURL = entry.MirrorLinks[0]
		}

		results[i] = result
	}

	return results
}
