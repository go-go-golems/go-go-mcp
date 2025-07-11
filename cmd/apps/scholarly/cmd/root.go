package cmd

import (
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	helpCmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var debugMode bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scholarly",
	Short: "A CLI tool to search Arxiv, LibGen, Crossref, and OpenAlex for scientific papers.",
	Long: `scholarly is a command-line tool that allows users to search for scientific papers across multiple academic databases and repositories including Arxiv, Library Genesis, Crossref, and OpenAlex.

It provides specific subcommands for each platform, allowing targeted searches with various filters and options.

Examples:
  scholarly arxiv -q "all:electron" -n 5
  scholarly libgen -q "artificial intelligence" -m "https://libgen.is"
  scholarly crossref -q "climate change mitigation"
  scholarly openalex -q "machine learning applications"`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := logging.InitLoggerFromViper()
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	helpSystem := help.NewHelpSystem()
	helpCmd.SetupCobraRootCommand(helpSystem, rootCmd)

	err := clay.InitViper("mcp", rootCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize viper")
	}

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
	// Remove the default toggle flag if it exists from the initial cobra init
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
