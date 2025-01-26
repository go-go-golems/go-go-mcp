package main

import (
	"context"
	"fmt"
	"io"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/pkg/htmlsimplifier"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Config struct {
	File      string     `yaml:"file"`
	Selectors []Selector `yaml:"selectors"`
	Config    struct {
		SampleCount  int `yaml:"sample_count"`
		ContextChars int `yaml:"context_chars"`
	} `yaml:"config"`
}

type Selector struct {
	Name     string `yaml:"name"`
	Selector string `yaml:"selector"`
	Type     string `yaml:"type"` // "css" or "xpath"
}

type SimplifiedSample struct {
	HTML    []htmlsimplifier.Document `yaml:"html"`
	Context []htmlsimplifier.Document `yaml:"context"`
	Path    string                    `yaml:"path"`
}

type SimplifiedResult struct {
	Name     string             `yaml:"name"`
	Selector string             `yaml:"selector"`
	Type     string             `yaml:"type"`
	Count    int                `yaml:"count"`
	Samples  []SimplifiedSample `yaml:"samples"`
}

type TestHTMLSelectorCommand struct {
	*cmds.CommandDescription
}

type TestHTMLSelectorSettings struct {
	ConfigFile   string `glazed.parameter:"config"`
	InputFile    string `glazed.parameter:"input"`
	Extract      bool   `glazed.parameter:"extract"`
	ShowContext  bool   `glazed.parameter:"show-context"`
	ShowPath     bool   `glazed.parameter:"show-path"`
	SampleCount  int    `glazed.parameter:"sample-count"`
	ContextChars int    `glazed.parameter:"context-chars"`
	StripScripts bool   `glazed.parameter:"strip-scripts"`
	StripCSS     bool   `glazed.parameter:"strip-css"`
	ShortenText  bool   `glazed.parameter:"shorten-text"`
	CompactSVG   bool   `glazed.parameter:"compact-svg"`
	StripSVG     bool   `glazed.parameter:"strip-svg"`
	SimplifyText bool   `glazed.parameter:"simplify-text"`
	Markdown     bool   `glazed.parameter:"markdown"`
	MaxListItems int    `glazed.parameter:"max-list-items"`
	MaxTableRows int    `glazed.parameter:"max-table-rows"`
}

func NewTestHTMLSelectorCommand() (*TestHTMLSelectorCommand, error) {
	return &TestHTMLSelectorCommand{
		CommandDescription: cmds.NewCommandDescription(
			"test-html-selector",
			cmds.WithShort("Test HTML/XPath selectors against HTML documents"),
			cmds.WithLong(`A tool for testing CSS and XPath selectors against HTML documents.
It provides match counts and contextual examples to verify selector accuracy.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"config",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML config file containing selectors"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"input",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to HTML input file"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"extract",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Extract and print all matches for each selector"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"show-context",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show context around matched elements"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"show-path",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show path to matched elements"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"sample-count",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of examples to show"),
					parameters.WithDefault(5),
				),
				parameters.NewParameterDefinition(
					"context-chars",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of characters of context to include"),
					parameters.WithDefault(100),
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
			),
		),
	}, nil
}

func (c *TestHTMLSelectorCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &TestHTMLSelectorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	config, err := loadConfig(s.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config settings with command line parameters
	config.Config.SampleCount = s.SampleCount
	config.Config.ContextChars = s.ContextChars

	// Create HTML simplifier
	simplifier := htmlsimplifier.NewSimplifier(htmlsimplifier.Options{
		StripScripts: s.StripScripts,
		StripCSS:     s.StripCSS,
		ShortenText:  s.ShortenText,
		CompactSVG:   s.CompactSVG,
		StripSVG:     s.StripSVG,
		SimplifyText: s.SimplifyText,
		Markdown:     s.Markdown,
		MaxListItems: s.MaxListItems,
		MaxTableRows: s.MaxTableRows,
	})

	tester, err := NewSelectorTester(&Config{
		File:      s.InputFile,
		Selectors: config.Selectors,
		Config: struct {
			SampleCount  int `yaml:"sample_count"`
			ContextChars int `yaml:"context_chars"`
		}{
			SampleCount:  s.SampleCount,
			ContextChars: s.ContextChars,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tester: %w", err)
	}

	results, err := tester.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}

	// Convert results to use Document structure
	var newResults []SimplifiedResult
	for _, result := range results {
		newResult := SimplifiedResult{
			Name:     result.Name,
			Selector: result.Selector,
			Type:     result.Type,
			Count:    result.Count,
		}

		for _, sample := range result.Samples {
			newSample := SimplifiedSample{}

			// Only include path if ShowPath is true
			if s.ShowPath {
				newSample.Path = sample.Path
			}

			// Process HTML content
			htmlDocs, err := simplifier.ProcessHTML(sample.HTML)
			if err == nil {
				newSample.HTML = htmlDocs
			}

			// Process context only if ShowContext is true
			if s.ShowContext {
				contextDocs, err := simplifier.ProcessHTML(sample.Context)
				if err == nil {
					newSample.Context = contextDocs
				}
			}

			newResult.Samples = append(newResult.Samples, newSample)
		}
		newResults = append(newResults, newResult)
	}

	if s.Extract {
		for _, result := range newResults {
			fmt.Fprintf(w, "Selector: %s\n", result.Selector)
			for _, sample := range result.Samples {
				for _, doc := range sample.HTML {
					if doc.Text != "" {
						fmt.Fprintln(w, doc.Text)
					} else if doc.Markdown != "" {
						fmt.Fprintln(w, doc.Markdown)
					}
				}
			}
		}
	} else {
		return yaml.NewEncoder(w).Encode(newResults)
	}

	return nil
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var config Config
	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "test-html-selector",
		Short: "Test HTML/XPath selectors against HTML documents",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// reinitialize the logger because we can now parse --log-level and co
			// from the command line flag
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	err := clay.InitViper("test-html-selector", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	cmd, err := NewTestHTMLSelectorCommand()
	cobra.CheckErr(err)

	cobraCmd, err := cli.BuildCobraCommandFromWriterCommand(cmd)
	cobra.CheckErr(err)

	rootCmd.AddCommand(cobraCmd)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
