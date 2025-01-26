package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/htmlsimplifier"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configFile string

func setupLogging(debug bool, logLevel string) {
	// Default to no logging unless debug is enabled
	if !debug {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		return
	}

	// Configure console writer with caller info
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	}
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()

	// Set log level based on flag
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Msgf("Invalid log level %q, defaulting to debug", logLevel)
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
}

func main() {
	var opts htmlsimplifier.Options
	var debug bool
	var logLevel string

	var rootCmd = &cobra.Command{
		Use:   "simplify-html [file]",
		Short: "Simplify and minimize HTML documents",
		Long: `A tool to simplify and minimize HTML documents by removing unnecessary elements 
and attributes, and shortening overly long text content. The output is provided 
in a structured YAML format.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setupLogging(debug, logLevel)
			log.Debug().Msg("Starting HTML simplification")

			if configFile != "" {
				config, err := loadFilterConfig(configFile)
				if err != nil {
					return err
				}
				opts.FilterConfig = config
			}

			return processFile(args[0], opts)
		},
	}

	rootCmd.Flags().BoolVar(&opts.StripScripts, "strip-scripts", true, "Remove <script> tags")
	rootCmd.Flags().BoolVar(&opts.StripCSS, "strip-css", true, "Remove <style> tags and style attributes")
	rootCmd.Flags().BoolVar(&opts.ShortenText, "shorten-text", true, "Shorten <span> and <p> elements longer than 200 characters")
	rootCmd.Flags().BoolVar(&opts.CompactSVG, "compact-svg", true, "Simplify SVG elements in output")
	rootCmd.Flags().BoolVar(&opts.StripSVG, "strip-svg", true, "Remove all SVG elements")
	rootCmd.Flags().BoolVar(&opts.SimplifyText, "simplify-text", true, "Collapse nodes with only text/br children into a single text field")
	rootCmd.Flags().IntVar(&opts.MaxListItems, "max-list-items", 4, "Maximum number of items to show in lists and select boxes (0 for unlimited)")
	rootCmd.Flags().IntVar(&opts.MaxTableRows, "max-table-rows", 4, "Maximum number of rows to show in tables (0 for unlimited)")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Path to YAML config file containing selectors")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "debug", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&opts.Markdown, "markdown", true, "Convert text with important elements to markdown format")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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

	// Validate selector types
	for _, sel := range config.Selectors {
		if sel.Type != "css" && sel.Type != "xpath" {
			return nil, fmt.Errorf("invalid selector type '%s': must be 'css' or 'xpath'", sel.Type)
		}
	}

	log.Debug().Int("selector_count", len(config.Selectors)).Msg("Loaded filter config")
	return &config, nil
}

func processFile(filename string, opts htmlsimplifier.Options) error {
	log.Debug().Str("filename", filename).Msg("Processing HTML file")

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	s := htmlsimplifier.NewSimplifier(opts)
	result, err := s.ProcessHTML(string(data))
	if err != nil {
		return fmt.Errorf("failed to process HTML: %w", err)
	}

	yamlData, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	log.Debug().Int("yaml_bytes", len(yamlData)).Msg("Generated YAML output")
	fmt.Println(string(yamlData))
	return nil
}
