package cmd

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/tools"
	"github.com/rs/zerolog/log"
)

// MetricsCommand is a glazed command for getting metrics for a scholarly work
type MetricsCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &MetricsCommand{}

// MetricsSettings holds the parameters for metrics retrieval
type MetricsSettings struct {
	WorkID string `glazed.parameter:"work_id"`
}

// RunIntoGlazeProcessor executes the metrics retrieval and processes results
func (c *MetricsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &MetricsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate work ID
	if s.WorkID == "" {
		return fmt.Errorf("work_id cannot be empty")
	}

	log.Debug().Str("work_id", s.WorkID).Msg("Getting metrics")

	req := common.GetMetricsRequest{
		WorkID: s.WorkID,
	}

	// Get metrics
	metrics, err := tools.GetMetrics(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get metrics")
		return err
	}

	// Create a row from the metrics
	row := types.NewRow()

	// Add all available fields to the row
	row.Set("citation_count", metrics.CitationCount)
	row.Set("cited_by_count", metrics.CitedByCount)
	row.Set("reference_count", metrics.ReferenceCount)
	row.Set("is_open_access", metrics.IsOA)
	row.Set("open_access_status", metrics.OAStatus)

	// Add altmetrics if available
	for metric, value := range metrics.Altmetrics {
		row.Set("altmetric_"+metric, value)
	}

	return gp.AddRow(ctx, row)
}

// NewMetricsCommand creates a new metrics command
func NewMetricsCommand() (*MetricsCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"metrics",
		cmds.WithShort("Get metrics for a scholarly work"),
		cmds.WithLong(`Get quantitative metrics for a scholarly work, including
citation counts, reference counts, and open access status.

The work_id can be either a DOI or an OpenAlex Work ID.

Example:
  arxiv-libgen-cli metrics --work-id "10.1038/nphys1170"
  arxiv-libgen-cli metrics --work-id "W2741809809" --output json`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"work_id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Work ID (DOI or OpenAlex ID) (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("i"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &MetricsCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the metrics command
func init() {
	// Create the metrics command
	metricsCmd, err := NewMetricsCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create metrics command")
	}

	// Convert to Cobra command
	metricsCobraCmd, err := cli.BuildCobraCommandFromCommand(metricsCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build metrics cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(metricsCobraCmd)
}
