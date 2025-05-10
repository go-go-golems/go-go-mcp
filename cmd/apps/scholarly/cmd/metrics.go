package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	metricsWorkID     string
	metricsJsonOutput bool
)

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Get metrics for a scholarly work",
	Long: `Get quantitative metrics for a scholarly work, including
citation counts, reference counts, and open access status.

The work_id can be either a DOI or an OpenAlex Work ID.

Example:
  arxiv-libgen-cli metrics --id "10.1038/nphys1170"
  arxiv-libgen-cli metrics --id "W2741809809" --json`,
	Run: func(cmd *cobra.Command, args []string) {
		if metricsWorkID == "" {
			fmt.Println("Error: work_id cannot be empty")
			if err := cmd.Help(); err != nil {
				log.Error().Err(err).Msg("Failed to show help")
			}
			os.Exit(1)
		}

		log.Debug().Str("work_id", metricsWorkID).Msg("Getting metrics")

		req := common.GetMetricsRequest{
			WorkID: metricsWorkID,
		}

		metrics, err := tools.GetMetrics(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get metrics")
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}

		if metricsJsonOutput {
			printMetricsAsJSON(metrics)
		} else {
			printMetricsNice(metrics)
		}
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)

	metricsCmd.Flags().StringVarP(&metricsWorkID, "id", "i", "", "Work ID (DOI or OpenAlex ID) (required)")
	metricsCmd.Flags().BoolVarP(&metricsJsonOutput, "json", "j", false, "Output as JSON")

	if err := metricsCmd.MarkFlagRequired("id"); err != nil {
		log.Fatal().Err(err).Msg("Failed to mark id flag as required")
	}
}

// printMetricsAsJSON prints the metrics as JSON
func printMetricsAsJSON(metrics *common.Metrics) {
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal metrics to JSON")
		fmt.Println("Error formatting metrics as JSON")
		return
	}

	fmt.Println(string(jsonData))
}

// printMetricsNice prints the metrics in a nice format
func printMetricsNice(metrics *common.Metrics) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("Work Metrics:")
	fmt.Printf("  Citation Count: %d\n", metrics.CitationCount)
	fmt.Printf("  Cited By Count: %d\n", metrics.CitedByCount)
	fmt.Printf("  Reference Count: %d\n", metrics.ReferenceCount)

	if metrics.IsOA {
		fmt.Printf("  Open Access: Yes")
		if metrics.OAStatus != "" {
			fmt.Printf(" (%s)\n", metrics.OAStatus)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Println("  Open Access: No")
	}

	if len(metrics.Altmetrics) > 0 {
		fmt.Println("  Altmetrics:")
		for metric, value := range metrics.Altmetrics {
			fmt.Printf("    %s: %d\n", metric, value)
		}
	}
	fmt.Println("--------------------------------------------------")
}
