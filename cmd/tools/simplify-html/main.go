package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/pkg/htmlsimplifier"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type SimplifyHTMLCommand struct {
	*cmds.CommandDescription
}

type SimplifyHTMLSettings struct {
	StripScripts bool     `glazed.parameter:"strip-scripts"`
	StripCSS     bool     `glazed.parameter:"strip-css"`
	ShortenText  bool     `glazed.parameter:"shorten-text"`
	CompactSVG   bool     `glazed.parameter:"compact-svg"`
	StripSVG     bool     `glazed.parameter:"strip-svg"`
	SimplifyText bool     `glazed.parameter:"simplify-text"`
	Markdown     bool     `glazed.parameter:"markdown"`
	MaxListItems int      `glazed.parameter:"max-list-items"`
	MaxTableRows int      `glazed.parameter:"max-table-rows"`
	ConfigFile   string   `glazed.parameter:"config"`
	Debug        bool     `glazed.parameter:"debug"`
	LogLevel     string   `glazed.parameter:"log-level"`
	Files        []string `glazed.parameter:"files"`
	URLs         []string `glazed.parameter:"urls"`
}

func NewSimplifyHTMLCommand() (*SimplifyHTMLCommand, error) {
	return &SimplifyHTMLCommand{
		CommandDescription: cmds.NewCommandDescription(
			"simplify-html",
			cmds.WithShort("Simplify and minimize HTML documents"),
			cmds.WithLong(`A tool to simplify and minimize HTML documents by removing unnecessary elements 
and attributes, and shortening overly long text content. The output is provided 
in a structured YAML format.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("HTML files to process (use - for stdin)"),
					parameters.WithRequired(false),
				),
				parameters.NewParameterDefinition(
					"urls",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("URLs to fetch and process"),
					parameters.WithRequired(false),
				),
				parameters.NewParameterDefinition(
					"strip-scripts",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Remove <script> tags"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"strip-css",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Remove <style> tags and style attributes"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"shorten-text",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Shorten <span> and <p> elements longer than 200 characters"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"compact-svg",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Simplify SVG elements in output"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"strip-svg",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Remove all SVG elements"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"simplify-text",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Collapse nodes with only text/br children into a single text field"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"markdown",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Convert text with important elements to markdown format"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"max-list-items",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of items to show in lists and select boxes (0 for unlimited)"),
					parameters.WithDefault(4),
				),
				parameters.NewParameterDefinition(
					"max-table-rows",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of rows to show in tables (0 for unlimited)"),
					parameters.WithDefault(4),
				),
				parameters.NewParameterDefinition(
					"config",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML config file containing selectors"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"debug",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable debug logging"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"log-level",
					parameters.ParameterTypeString,
					parameters.WithHelp("Log level (debug, info, warn, error)"),
					parameters.WithDefault("debug"),
				),
			),
		),
	}, nil
}

func setupLogging(debug bool, logLevel string) {
	if !debug {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		return
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	}
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Msgf("Invalid log level %q, defaulting to debug", logLevel)
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
}

func (c *SimplifyHTMLCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &SimplifyHTMLSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	if len(s.Files) == 0 && len(s.URLs) == 0 {
		return fmt.Errorf("no input sources provided: use either --files or --urls")
	}

	setupLogging(s.Debug, s.LogLevel)
	log.Debug().Msg("Starting HTML simplification")

	opts := htmlsimplifier.Options{
		StripScripts: s.StripScripts,
		StripCSS:     s.StripCSS,
		ShortenText:  s.ShortenText,
		CompactSVG:   s.CompactSVG,
		StripSVG:     s.StripSVG,
		SimplifyText: s.SimplifyText,
		Markdown:     s.Markdown,
		MaxListItems: s.MaxListItems,
		MaxTableRows: s.MaxTableRows,
	}

	if s.ConfigFile != "" {
		config, err := loadFilterConfig(s.ConfigFile)
		if err != nil {
			return err
		}
		opts.FilterConfig = config
	}

	simplifier := htmlsimplifier.NewSimplifier(opts)
	var results []map[string]interface{}

	// Process files
	for _, file := range s.Files {
		var fileData []byte
		var err error

		if file == "-" {
			fileData, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		} else {
			fileData, err = os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", file, err)
			}
		}

		result, err := simplifier.ProcessHTML(string(fileData))
		if err != nil {
			return fmt.Errorf("failed to process HTML from %s: %w", file, err)
		}

		results = append(results, map[string]interface{}{
			"source": file,
			"data":   result,
		})
	}

	// Process URLs
	for _, url := range s.URLs {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch URL %s: %w", url, err)
		}
		defer resp.Body.Close()

		fileData, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response from %s: %w", url, err)
		}

		result, err := simplifier.ProcessHTML(string(fileData))
		if err != nil {
			return fmt.Errorf("failed to process HTML from %s: %w", url, err)
		}

		results = append(results, map[string]interface{}{
			"source": url,
			"data":   result,
		})
	}

	yamlData, err := yaml.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	log.Debug().Int("yaml_bytes", len(yamlData)).Msg("Generated YAML output")
	_, err = w.Write(yamlData)
	return err
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "simplify-html",
		Short: "Simplify and minimize HTML documents",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// reinitialize the logger because we can now parse --log-level and co
			// from the command line flag
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	err := clay.InitViper("simplify-html", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	cmd, err := NewSimplifyHTMLCommand()
	cobra.CheckErr(err)

	cobraCmd, err := cli.BuildCobraCommandFromWriterCommand(cmd)
	cobra.CheckErr(err)

	rootCmd.AddCommand(cobraCmd)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}

func loadFilterConfig(filename string) (*htmlsimplifier.FilterConfig, error) {
	if filename == "" {
		log.Debug().Msg("No config file specified")
		return nil, nil
	}

	log.Debug().Str("filename", filename).Msg("Loading filter config")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config htmlsimplifier.FilterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate selector types and modes
	for _, sel := range config.Selectors {
		if sel.Type != "css" && sel.Type != "xpath" {
			return nil, fmt.Errorf("invalid selector type '%s': must be 'css' or 'xpath'", sel.Type)
		}
		if sel.Mode != htmlsimplifier.SelectorModeSelect && sel.Mode != htmlsimplifier.SelectorModeFilter {
			return nil, fmt.Errorf("invalid selector mode '%s': must be 'select' or 'filter'", sel.Mode)
		}
	}

	log.Debug().Int("selector_count", len(config.Selectors)).Msg("Loaded filter config")
	return &config, nil
}
