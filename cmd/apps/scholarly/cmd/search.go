package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/scholarly"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	searchQuery  string
	searchSource string
	searchLimit  int
	searchFilter string
	jsonOutput   bool
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for scholarly works across multiple sources",
	Long: `Search for scholarly works across Arxiv, Crossref, and OpenAlex.

Examples:
  arxiv-libgen-cli search --query "quantum computing" --source arxiv
  arxiv-libgen-cli search --query "climate change" --source openalex --limit 5
  arxiv-libgen-cli search --query "machine learning" --source crossref --filter "type:journal-article"
  arxiv-libgen-cli search --query "neural networks" --source arxiv --json`,
	Run: func(cmd *cobra.Command, args []string) {
		if searchQuery == "" {
			fmt.Println("Error: query cannot be empty")
			cmd.Help()
			os.Exit(1)
		}

		if !isValidSource(searchSource) {
			fmt.Printf("Error: invalid source '%s'. Must be one of: arxiv, crossref, openalex\n", searchSource)
			os.Exit(1)
		}

		log.Debug().Str("query", searchQuery).Str("source", searchSource).Int("limit", searchLimit).Str("filter", searchFilter).Msg("Search initiated")

		filterMap := parseFilterString(searchFilter)

		req := scholarly.SearchWorksRequest{
			Query:  searchQuery,
			Source: searchSource,
			Limit:  searchLimit,
			Filter: filterMap,
		}

		response, err := scholarly.SearchWorks(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to search works")
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}

		if len(response.Works) == 0 {
			fmt.Println("No results found.")
			return
		}

		if jsonOutput {
			printJSONResults(response)
		} else {
			printFormattedResults(response)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Search query (required)")
	searchCmd.Flags().StringVarP(&searchSource, "source", "s", "arxiv", "Source to search (arxiv, crossref, openalex)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results to return")
	searchCmd.Flags().StringVarP(&searchFilter, "filter", "f", "", "Filter string (format: key1:value1,key2:value2)")
	searchCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results as JSON")

	searchCmd.MarkFlagRequired("query")
}

// isValidSource checks if the provided source is valid
func isValidSource(source string) bool {
	validSources := []string{"arxiv", "crossref", "openalex"}
	source = strings.ToLower(source)

	for _, valid := range validSources {
		if source == valid {
			return true
		}
	}
	return false
}

// parseFilterString parses a comma-separated list of key:value pairs into a map
func parseFilterString(filterStr string) map[string]string {
	filterMap := make(map[string]string)
	if filterStr == "" {
		return filterMap
	}

	filters := strings.Split(filterStr, ",")
	for _, filter := range filters {
		parts := strings.SplitN(filter, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			filterMap[key] = value
		}
	}

	return filterMap
}

// printJSONResults prints the search results as JSON
func printJSONResults(response *scholarly.SearchWorksResponse) {
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal JSON results")
		fmt.Println("Error formatting results as JSON")
		return
	}

	fmt.Println(string(jsonData))
}

// printFormattedResults prints the search results in a human-readable format
func printFormattedResults(response *scholarly.SearchWorksResponse) {
	fmt.Printf("Found %d results:\n", len(response.Works))
	fmt.Println("--------------------------------------------------")

	for i, work := range response.Works {
		fmt.Printf("Result %d:\n", i+1)
		fmt.Printf("  Title: %s\n", work.Title)

		if len(work.Authors) > 0 {
			fmt.Printf("  Authors: %s\n", strings.Join(work.Authors, ", "))
		}

		if work.Year > 0 {
			fmt.Printf("  Year: %d\n", work.Year)
		}

		if work.DOI != "" {
			fmt.Printf("  DOI: %s\n", work.DOI)
		}

		if work.CitationCount > 0 {
			fmt.Printf("  Citations: %d\n", work.CitationCount)
		}

		if work.IsOA {
			fmt.Printf("  Open Access: Yes\n")
		}

		if work.PDFURL != "" {
			fmt.Printf("  PDF URL: %s\n", work.PDFURL)
		}

		if work.Abstract != "" {
			// Truncate abstract if it's too long
			abstract := work.Abstract
			if len(abstract) > 200 {
				abstract = abstract[:200] + "..."
			}
			fmt.Printf("  Abstract: %s\n", abstract)
		}

		fmt.Printf("  Source: %s\n", work.SourceName)
		fmt.Println("--------------------------------------------------")
	}
}
