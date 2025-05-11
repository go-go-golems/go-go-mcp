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

// KeywordsCommand is a glazed command for suggesting keywords from text
type KeywordsCommand struct {
	*cmds.CommandDescription
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &KeywordsCommand{}

// KeywordsSettings holds the parameters for keyword suggestion
type KeywordsSettings struct {
	Text     string `glazed.parameter:"text"`
	MaxCount int    `glazed.parameter:"max_count"`
}

// RunIntoGlazeProcessor executes the keyword suggestion and processes results
func (c *KeywordsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &KeywordsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Validate text
	if s.Text == "" {
		return fmt.Errorf("text cannot be empty")
	}

	log.Debug().
		Str("text_sample", truncateText(s.Text, 30)).
		Int("max_keywords", s.MaxCount).
		Msg("Suggesting keywords")

	req := common.SuggestKeywordsRequest{
		Text:        s.Text,
		MaxKeywords: s.MaxCount,
	}

	// Get keyword suggestions
	response, err := tools.SuggestKeywords(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to suggest keywords")
		return err
	}

	if len(response.Keywords) == 0 {
		log.Info().Msg("No keywords found for the given text")
		return nil
	}

	// Output each keyword as a row
	for _, keyword := range response.Keywords {
		row := types.NewRow()
		row.Set("display_name", keyword.DisplayName)
		row.Set("id", keyword.ID)
		row.Set("relevance", keyword.Relevance)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// truncateText truncates text to a specified length with ellipsis
func truncateText(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}

// NewKeywordsCommand creates a new keywords command
func NewKeywordsCommand() (*KeywordsCommand, error) {
	// Create the Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"keywords",
		cmds.WithShort("Suggest keywords for a text"),
		cmds.WithLong(`Generate suggested keywords for a given text using
OpenAlex's controlled vocabulary concepts.

This command analyzes the provided text and returns relevant
keywords that can be used for further searches.

Example:
  arxiv-libgen-cli keywords --text "Quantum computing uses qubits to perform calculations"
  arxiv-libgen-cli keywords --text "Climate change mitigation strategies" --max-count 5 --output json`),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"text",
				parameters.ParameterTypeString,
				parameters.WithHelp("Text to analyze for keywords (required)"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("t"),
			),
			parameters.NewParameterDefinition(
				"max_count",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of keywords to return"),
				parameters.WithDefault(10),
				parameters.WithShortFlag("m"),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &KeywordsCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// init registers the keywords command
func init() {
	// Create the keywords command
	keywordsCmd, err := NewKeywordsCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create keywords command")
	}

	// Convert to Cobra command
	keywordsCobraCmd, err := cli.BuildCobraCommandFromCommand(keywordsCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build keywords cobra command")
	}

	// Add to root command
	rootCmd.AddCommand(keywordsCobraCmd)
}
