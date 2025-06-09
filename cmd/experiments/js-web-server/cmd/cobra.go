package cmd

import (
	"fmt"
	"os"

	embeddings_config "github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings/claude"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings/openai"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/pinocchio/pkg/cmds/cmdlayers"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// BuildCobraCommandWithServeMiddlewares builds a Cobra command with custom js-web-server middlewares
// that include profile support specifically for the js-web-server application.
func BuildCobraCommandWithServeMiddlewares(
	cmd cmds.Command,
	options ...cli.CobraParserOption,
) (*cobra.Command, error) {
	options_ := append([]cli.CobraParserOption{
		cli.WithCobraMiddlewaresFunc(GetServeCommandMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug, cmdlayers.GeppettoHelpersSlug),
	}, options...)

	return cli.BuildCobraCommandFromCommand(cmd, options_...)
}

// GetServeCommandMiddlewares provides the middleware chain for js-web-server commands
// with proper profile support, configuration loading, and parameter precedence.
func GetServeCommandMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	commandSettings := &cli.CommandSettings{}
	err := parsedCommandLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

	profileSettings := &cli.ProfileSettings{}
	err = parsedCommandLayers.InitializeStruct(cli.ProfileSettingsSlug, profileSettings)
	if err != nil {
		return nil, err
	}

	middlewares_ := []middlewares.Middleware{
		middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}

	if commandSettings.LoadParametersFromFile != "" {
		middlewares_ = append(middlewares_,
			middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	// Profile support with layered configuration: pinocchio first, then js-web-server overrides
	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		log.Warn().Err(err).Msg("Could not get user config directory, using current directory")
		xdgConfigPath = "."
	}

	// Set up profile files: pinocchio as base, js-web-server as override
	pinocchioProfileFile := fmt.Sprintf("%s/pinocchio/profiles.yaml", xdgConfigPath)
	jsWebServerProfileFile := fmt.Sprintf("%s/js-web-server/profiles.yaml", xdgConfigPath)

	// Use specified profile file or default to js-web-server
	targetProfileFile := profileSettings.ProfileFile
	if targetProfileFile == "" {
		targetProfileFile = jsWebServerProfileFile
	}

	// Default to development profile for js-web-server
	profile := profileSettings.Profile
	if profile == "" {
		profile = "default"
	}

	middlewares_ = append(middlewares_,
		middlewares.GatherFlagsFromCustomProfiles(
			profile,
			middlewares.WithProfileFile(targetProfileFile),
			middlewares.WithProfileRequired(false), // Don't fail if profile doesn't exist
			middlewares.WithProfileParseOptions(
				parameters.WithParseStepSource("js-web-server-profiles"),
				parameters.WithParseStepMetadata(map[string]interface{}{
					"profileFile": targetProfileFile,
					"profile":     profile,
					"layer":       "override",
				}),
			),
		),
		// Pinocchio profiles as base configuration
		middlewares.GatherFlagsFromProfiles(
			pinocchioProfileFile,
			pinocchioProfileFile,
			profile,
			parameters.WithParseStepSource("pinocchio-profiles"),
			parameters.WithParseStepMetadata(map[string]interface{}{
				"profileFile": pinocchioProfileFile,
				"profile":     profile,
				"layer":       "base",
			}),
		),

		middlewares.WrapWithWhitelistedLayers(
			[]string{
				settings.AiChatSlug,
				settings.AiClientSlug,
				openai.OpenAiChatSlug,
				claude.ClaudeChatSlug,
				cmdlayers.GeppettoHelpersSlug,
				embeddings_config.EmbeddingsSlug,
				cli.ProfileSettingsSlug,
			},
			middlewares.GatherFlagsFromCustomViper(
				middlewares.WithAppName("pinocchio"),
				middlewares.WithParseOptions(parameters.WithParseStepSource("pinocchio-viper")),
			),
			middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("js-web-server-viper")),
			middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
		),
		// JS web server viper config
		middlewares.WrapWithWhitelistedLayers(
			[]string{
				layers.DefaultSlug, // Include the default layer for js-web-server settings
			},
			middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("js-web-server-viper")),
			middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
		),
	)

	return middlewares_, nil
}
