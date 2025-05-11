package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/libgen"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/clients/openalex"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/rs/zerolog/log"
)

// FindFullText finds the best PDF or HTML URL for a work
func FindFullText(req common.FindFullTextRequest) (*common.FindFullTextResponse, error) {
	if req.DOI == "" && req.Title == "" {
		return nil, fmt.Errorf("either DOI or title must be provided")
	}

	if req.PreferVersion == "" {
		req.PreferVersion = "published" // Default to published version
	}

	log.Debug().Str("doi", req.DOI).Str("title", req.Title).Str("prefer_version", req.PreferVersion).Msg("Finding full text")

	// Try Unpaywall first if DOI is provided (legal OA source)
	if req.DOI != "" {
		response, err := findFullTextUnpaywall(req.DOI, req.PreferVersion)
		if err == nil && response.PDFURL != "" {
			log.Debug().Str("source", "unpaywall").Str("url", response.PDFURL).Msg("Found full text")
			return response, nil
		}
		log.Debug().Err(err).Msg("Unpaywall full text lookup failed, trying next source")
	}

	// Try OpenAlex next if DOI is provided (legal OA source)
	if req.DOI != "" {
		response, err := findFullTextOpenAlex(req.DOI, req.PreferVersion)
		if err == nil && response.PDFURL != "" {
			log.Debug().Str("source", "openalex").Str("url", response.PDFURL).Msg("Found full text")
			return response, nil
		}
		log.Debug().Err(err).Msg("OpenAlex full text lookup failed, trying next source")
	}

	// Finally try LibGen with title (or DOI if no title provided)
	searchTerm := req.Title
	if searchTerm == "" {
		searchTerm = req.DOI
	}

	response, err := findFullTextLibGen(searchTerm)
	if err == nil && response.PDFURL != "" {
		log.Debug().Str("source", "libgen").Str("url", response.PDFURL).Msg("Found full text")
		return response, nil
	}

	// If all methods failed, return error
	return nil, fmt.Errorf("could not find full text using any available source")
}

// findFullTextOpenAlex tries to find full text via OpenAlex
func findFullTextOpenAlex(doi string, preferVersion string) (*common.FindFullTextResponse, error) {
	client := openalex.NewClient("")

	// OpenAlex expects DOIs with a URL prefix
	doiURL := doi
	if !strings.HasPrefix(doi, "https://doi.org/") {
		doiURL = "https://doi.org/" + doi
	}

	// Search for the DOI
	params := common.SearchParams{
		Query:      doiURL,
		MaxResults: 1,
	}

	results, err := client.Search(params)
	if err != nil {
		return nil, fmt.Errorf("OpenAlex error: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("DOI not found in OpenAlex")
	}

	result := results[0]

	// Check if the work has a PDF URL
	if result.PDFURL != "" {
		// Check if it's open access
		if result.OAStatus != "" {
			return &common.FindFullTextResponse{
				PDFURL:   result.PDFURL,
				Source:   "openalex",
				OAStatus: result.OAStatus,
				License:  result.License,
			}, nil
		}
	}

	// If not found, check locations from metadata
	locations, ok := result.Metadata["locations"].([]interface{})
	if !ok || len(locations) == 0 {
		return nil, fmt.Errorf("no locations found in OpenAlex")
	}

	// Look for OA locations with the preferred version
	for _, loc := range locations {
		location, ok := loc.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if location is OA
		isOA, ok := location["is_oa"].(bool)
		if !ok || !isOA {
			continue
		}

		// Check version preference
		version, ok := location["version"].(string)
		if !ok {
			continue
		}

		// Version matching based on preference
		var versionMatch bool
		switch preferVersion {
		case "published":
			versionMatch = (version == "publishedVersion")
		case "accepted":
			versionMatch = (version == "acceptedVersion")
		case "submitted":
			versionMatch = (version == "submittedVersion")
		default:
			versionMatch = true // Accept any version if preference unknown
		}

		if !versionMatch {
			continue
		}

		// Get PDF URL and license
		pdfURL, _ := location["pdf_url"].(string)
		htmlURL, _ := location["url"].(string)
		license, _ := location["license"].(string)
		oaStatus, _ := location["oa_status"].(string)

		// Prefer PDF URL if available, otherwise use HTML URL
		url := pdfURL
		isPDF := true
		if url == "" {
			url = htmlURL
			isPDF = false
		}

		if url != "" {
			return &common.FindFullTextResponse{
				PDFURL:   url,
				Source:   "openalex",
				OAStatus: oaStatus,
				License:  license,
				IsPDF:    isPDF,
			}, nil
		}
	}

	return nil, fmt.Errorf("no suitable full text found in OpenAlex")
}

// UnpaywallResponse represents the Unpaywall API response
type UnpaywallResponse struct {
	BestOALocation struct {
		URL         string `json:"url"`
		URLForPDF   string `json:"url_for_pdf,omitempty"`
		VersionType string `json:"version"`
		License     string `json:"license"`
		HostType    string `json:"host_type"`
		IsPDF       bool   `json:"is_pdf"`
	} `json:"best_oa_location"`
	IsOA        bool   `json:"is_oa"`
	OAStatus    string `json:"oa_status"`
	OALocations []struct {
		URL         string `json:"url"`
		URLForPDF   string `json:"url_for_pdf,omitempty"`
		VersionType string `json:"version"`
		License     string `json:"license"`
		HostType    string `json:"host_type"`
		IsPDF       bool   `json:"is_pdf"`
	} `json:"oa_locations"`
}

// findFullTextUnpaywall tries to find full text via Unpaywall API
func findFullTextUnpaywall(doi string, preferVersion string) (*common.FindFullTextResponse, error) {
	// Clean DOI format
	cleanDOI := strings.TrimPrefix(doi, "https://doi.org/")
	cleanDOI = strings.TrimPrefix(cleanDOI, "doi.org/")

	// Unpaywall API endpoint - use manuel's email address
	apiURL := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=manuel@bl0rg.net", url.PathEscape(cleanDOI))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Unpaywall request: %w", err)
	}

	req.Header.Set("User-Agent", "arxiv-libgen-cli/0.1")

	resp := common.MakeHTTPRequest(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unpaywall API request failed with status %d", resp.StatusCode)
	}

	var unpaywall UnpaywallResponse
	if err := json.Unmarshal(resp.Body, &unpaywall); err != nil {
		return nil, fmt.Errorf("error parsing Unpaywall response: %w", err)
	}

	// Check if it's OA
	if !unpaywall.IsOA {
		return nil, fmt.Errorf("no open access location found in Unpaywall")
	}

	// First try to find a location matching the preferred version
	for _, location := range unpaywall.OALocations {
		// Check version preference
		version := location.VersionType
		var versionMatch bool
		switch preferVersion {
		case "published":
			versionMatch = (version == "publishedVersion")
		case "accepted":
			versionMatch = (version == "acceptedVersion")
		case "submitted":
			versionMatch = (version == "submittedVersion")
		default:
			versionMatch = true // Accept any version if preference unknown
		}

		if !versionMatch {
			continue
		}

		// Prefer PDF URL if available, otherwise use regular URL
		url := location.URLForPDF
		isPDF := true

		if url == "" {
			url = location.URL
			isPDF = location.IsPDF // Use the IsPDF flag from Unpaywall
		}

		if url != "" {
			return &common.FindFullTextResponse{
				PDFURL:   url,
				Source:   "unpaywall",
				OAStatus: unpaywall.OAStatus,
				License:  location.License,
				IsPDF:    isPDF,
			}, nil
		}
	}

	// If no version-matched location found, use the best OA location
	if unpaywall.BestOALocation.URL != "" {
		// Prefer PDF URL if available
		pdfURL := unpaywall.BestOALocation.URLForPDF
		isPDF := true

		if pdfURL == "" {
			pdfURL = unpaywall.BestOALocation.URL
			isPDF = unpaywall.BestOALocation.IsPDF
		}

		return &common.FindFullTextResponse{
			PDFURL:   pdfURL,
			Source:   "unpaywall",
			OAStatus: unpaywall.OAStatus,
			License:  unpaywall.BestOALocation.License,
			IsPDF:    isPDF,
		}, nil
	}

	return nil, fmt.Errorf("no suitable full text found in Unpaywall")
}

// findFullTextLibGen tries to find full text via LibGen
func findFullTextLibGen(searchTerm string) (*common.FindFullTextResponse, error) {
	client := libgen.NewClient("https://libgen.rs")

	// Search LibGen
	params := common.SearchParams{
		Query:      searchTerm,
		MaxResults: 5, // Get a few results to find the best match
	}

	results, err := client.Search(params)
	if err != nil {
		return nil, fmt.Errorf("LibGen search error: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found in LibGen")
	}

	// Find the best match (prefer first result that has a download URL)
	for _, result := range results {
		// Make sure we have a download URL
		if result.PDFURL != "" {
			// Extract MD5 from result metadata if available
			md5 := ""
			if md5Val, ok := result.Metadata["md5"].(string); ok {
				md5 = md5Val
			}

			return &common.FindFullTextResponse{
				PDFURL:   result.PDFURL,
				Source:   "libgen",
				OAStatus: "closed", // LibGen content is not officially OA
				MD5:      md5,
				IsPDF:    true, // LibGen typically provides PDF downloads
			}, nil
		}
	}

	return nil, fmt.Errorf("no download URL found in LibGen results")
}
