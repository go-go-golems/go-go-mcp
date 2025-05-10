package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/querydsl"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	servePort   int
	serveHost   string
	corsOrigins []string
)

// ScholarlySearchResponse represents the API response structure
type ScholarlySearchResponse struct {
	Results []common.SearchResult `json:"results"`
	Query   string               `json:"query"`
	Count   int                  `json:"count"`
	Sources []string             `json:"sources"`
}

// ServeCommand starts an HTTP server that provides the scholarly search API
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start an API server for scholarly search",
	Long: `Start an HTTP server that provides a REST API for scholarly search functionality.

The server exposes endpoints to search for academic papers across multiple sources
including Arxiv, Crossref, and OpenAlex.

Examples:
  scholarly serve --port 8080 --host localhost
  scholarly serve --port 8080 --cors-origins "http://localhost:3000,http://localhost:5173"
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create router
		r := mux.NewRouter()

		// Setup routes
		r.HandleFunc("/api/search", handleSearch).Methods("GET")
		r.HandleFunc("/api/sources", handleSources).Methods("GET")
		r.HandleFunc("/api/health", handleHealth).Methods("GET")

		// Serve static files for the frontend
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("web/scholarly/dist")))

		// Setup CORS
		c := cors.New(cors.Options{
			AllowedOrigins:   corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Accept", "Authorization"},
			AllowCredentials: true,
		})

		// Create server
		server := &http.Server{
			Addr:         fmt.Sprintf("%s:%d", serveHost, servePort),
			Handler:      c.Handler(r),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		}

		// Start server
		log.Info().Msgf("Starting scholarly API server on %s:%d", serveHost, servePort)
		log.Info().Msgf("CORS allowed origins: %v", corsOrigins)
		log.Fatal().Err(server.ListenAndServe()).Msg("Server stopped")
	},
}

// handleSearch handles the /api/search endpoint
func handleSearch(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling search request")
	
	// Parse query parameters
	queryParams := r.URL.Query()
	queryText := queryParams.Get("query")
	sourcesStr := queryParams.Get("sources")
	limitStr := queryParams.Get("limit")
	author := queryParams.Get("author")
	title := queryParams.Get("title")
	category := queryParams.Get("category")
	workType := queryParams.Get("work-type")
	fromYearStr := queryParams.Get("from-year")
	toYearStr := queryParams.Get("to-year")
	sortStr := queryParams.Get("sort")
	openAccess := queryParams.Get("open-access")
	mailto := queryParams.Get("mailto")
	disableRerankStr := queryParams.Get("disable-rerank")
	
	log.Debug().Str("query", queryText).Str("sources", sourcesStr).Str("author", author).Msg("Search parameters")

	// Set default values
	limit := 20
	fromYear := 0
	toYear := 0
	sources := []string{"all"}
	disableRerank := false

	// Parse parameters
	if sourcesStr != "" {
		sources = strings.Split(sourcesStr, ",")
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if fromYearStr != "" {
		if y, err := strconv.Atoi(fromYearStr); err == nil {
			fromYear = y
		}
	}

	if toYearStr != "" {
		if y, err := strconv.Atoi(toYearStr); err == nil {
			toYear = y
		}
	}

	if disableRerankStr != "" {
		if dr, err := strconv.ParseBool(disableRerankStr); err == nil {
			disableRerank = dr
		}
	}

	// Create query using the DSL
	query := querydsl.New()

	// Set basic text fields
	if queryText != "" {
		query.WithText(queryText)
	}
	if author != "" {
		query.WithAuthor(author)
	}
	if title != "" {
		query.WithTitle(title)
	}
	if category != "" {
		query.WithCategory(category)
	}
	if workType != "" {
		query.WithType(workType)
	}

	// Set year range if provided
	if fromYear > 0 || toYear > 0 {
		query.Between(fromYear, toYear)
	}

	// Set max results
	query.WithMaxResults(limit)

	// Set sort order if provided
	if sortStr != "" {
		var sortOrder querydsl.SortOrder
		switch strings.ToLower(sortStr) {
		case "newest":
			sortOrder = querydsl.SortNewest
		case "oldest":
			sortOrder = querydsl.SortOldest
		default:
			sortOrder = querydsl.SortRelevance
		}
		query.Order(sortOrder)
	}

	// Set OpenAccess filter if provided
	if openAccess != "" {
		isOA := strings.ToLower(openAccess) == "true"
		query.OnlyOA(isOA)
	}

	// Determine which providers to search
	providers := getProvidersFromSources(sources)

	// Create options
	opts := tools.SearchOptions{
		Providers:   providers,
		Mailto:      mailto,
		UseReranker: !disableRerank,
	}

	// Perform the search
	ctx := context.Background()
	log.Debug().Interface("query", query).Interface("options", opts).Msg("Executing search")
	results, err := tools.Search(ctx, query, opts)
	if err != nil {
		log.Error().Err(err).Msg("Search failed")
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}
	log.Debug().Int("resultCount", len(results)).Msg("Search completed")

	// Create the response
	response := ScholarlySearchResponse{
		Results: results,
		Query:   queryText,
		Count:   len(results),
		Sources: sources,
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Debug log the response
	respBytes, _ := json.MarshalIndent(response, "", "  ")
	log.Debug().RawJSON("response", respBytes).Msg("Preparing response")
	
	// Encode and write the response
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
	log.Debug().Msg("Response sent successfully")
}

// handleSources handles the /api/sources endpoint
func handleSources(w http.ResponseWriter, r *http.Request) {
	// Define available sources
	sources := []string{"arxiv", "crossref", "openalex", "all"}

	// Create response
	response := map[string]interface{}{
		"sources": sources,
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and write the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleHealth handles the /api/health endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Create response
	response := map[string]interface{}{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and write the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func init() {
	// Add flags to the serve command
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to run the server on")
	serveCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "Host to run the server on")
	serveCmd.Flags().StringSliceVar(&corsOrigins, "cors-origins", []string{"*"}, "Allowed CORS origins")

	// Add the serve command to the root command
	rootCmd.AddCommand(serveCmd)
}
