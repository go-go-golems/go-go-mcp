package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/scholarly"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	keywordsText       string
	keywordsMaxCount   int
	keywordsJsonOutput bool
)

// keywordsCmd represents the keywords command
var keywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Suggest keywords for a text",
	Long: `Generate suggested keywords for a given text using
OpenAlex's controlled vocabulary concepts.

This command analyzes the provided text and returns relevant
keywords that can be used for further searches.

Example:
  arxiv-libgen-cli keywords --text "Quantum computing uses qubits to perform calculations"
  arxiv-libgen-cli keywords --text "Climate change mitigation strategies" --max 5`,
	Run: func(cmd *cobra.Command, args []string) {
		if keywordsText == "" {
			fmt.Println("Error: text cannot be empty")
			if err := cmd.Help(); err != nil {
				log.Error().Err(err).Msg("Failed to show help")
			}
			os.Exit(1)
		}

		log.Debug().Str("text_sample", truncateText(keywordsText, 30)).Int("max_keywords", keywordsMaxCount).Msg("Suggesting keywords")

		req := scholarly.SuggestKeywordsRequest{
			Text:        keywordsText,
			MaxKeywords: keywordsMaxCount,
		}

		response, err := scholarly.SuggestKeywords(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to suggest keywords")
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}

		if len(response.Keywords) == 0 {
			fmt.Println("No keywords found for the given text.")
			return
		}

		if keywordsJsonOutput {
			printKeywordsAsJSON(response)
		} else {
			printKeywordsNice(response)
		}
	},
}

func init() {
	rootCmd.AddCommand(keywordsCmd)

	keywordsCmd.Flags().StringVarP(&keywordsText, "text", "t", "", "Text to analyze for keywords (required)")
	keywordsCmd.Flags().IntVarP(&keywordsMaxCount, "max", "m", 10, "Maximum number of keywords to return")
	keywordsCmd.Flags().BoolVarP(&keywordsJsonOutput, "json", "j", false, "Output as JSON")

	if err := keywordsCmd.MarkFlagRequired("text"); err != nil {
		log.Fatal().Err(err).Msg("Failed to mark text flag as required")
	}
}

// truncateText truncates text to a specified length with ellipsis
func truncateText(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}

// printKeywordsAsJSON prints the keywords as JSON
func printKeywordsAsJSON(response *scholarly.SuggestKeywordsResponse) {
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal keywords to JSON")
		fmt.Println("Error formatting keywords as JSON")
		return
	}

	fmt.Println(string(jsonData))
}

// printKeywordsNice prints the keywords in a nice format
func printKeywordsNice(response *scholarly.SuggestKeywordsResponse) {
	fmt.Printf("Found %d keywords:\n", len(response.Keywords))
	fmt.Println("--------------------------------------------------")

	for i, keyword := range response.Keywords {
		fmt.Printf("%d. %s (relevance: %.2f)\n", i+1, keyword.DisplayName, keyword.Relevance)
		fmt.Printf("   ID: %s\n", keyword.ID)
	}

	fmt.Println("--------------------------------------------------")
}
