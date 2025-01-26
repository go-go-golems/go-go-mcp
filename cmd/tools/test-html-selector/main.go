package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
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
	File        string     `yaml:"file"`
	Description string     `yaml:"description"`
	Selectors   []Selector `yaml:"selectors"`
	Config      struct {
		SampleCount  int    `yaml:"sample_count"`
		ContextChars int    `yaml:"context_chars"`
		Template     string `yaml:"template"`
	} `yaml:"config"`
}

type Selector struct {
	Name        string `yaml:"name"`
	Selector    string `yaml:"selector"`
	Type        string `yaml:"type"` // "css" or "xpath"
	Description string `yaml:"description"`
}

type SimplifiedSample struct {
	HTML    []htmlsimplifier.Document `yaml:"html,omitempty"`
	Context []htmlsimplifier.Document `yaml:"context,omitempty"`
	Path    string                    `yaml:"path,omitempty"`
}

type SimplifiedResult struct {
	Name     string             `yaml:"name"`
	Selector string             `yaml:"selector"`
	Type     string             `yaml:"type"`
	Count    int                `yaml:"count"`
	Samples  []SimplifiedSample `yaml:"samples"`
}

type SourceResult struct {
	Source          string                   `yaml:"source"`
	Data            map[string][]interface{} `yaml:"data"`
	SelectorResults []SelectorResult         `yaml:"selector_results"`
}

type HTMLSelectorCommand struct {
	*cmds.CommandDescription
}

type HTMLSelectorSettings struct {
	ConfigFile      string   `glazed.parameter:"config"`
	SelectCSS       []string `glazed.parameter:"select-css"`
	SelectXPath     []string `glazed.parameter:"select-xpath"`
	Files           []string `glazed.parameter:"files"`
	URLs            []string `glazed.parameter:"urls"`
	Extract         bool     `glazed.parameter:"extract"`
	ExtractData     bool     `glazed.parameter:"extract-data"`
	ExtractTemplate string   `glazed.parameter:"extract-template"`
	ShowContext     bool     `glazed.parameter:"show-context"`
	ShowPath        bool     `glazed.parameter:"show-path"`
	SampleCount     int      `glazed.parameter:"sample-count"`
	ContextChars    int      `glazed.parameter:"context-chars"`
	StripScripts    bool     `glazed.parameter:"strip-scripts"`
	StripCSS        bool     `glazed.parameter:"strip-css"`
	ShortenText     bool     `glazed.parameter:"shorten-text"`
	CompactSVG      bool     `glazed.parameter:"compact-svg"`
	StripSVG        bool     `glazed.parameter:"strip-svg"`
	SimplifyText    bool     `glazed.parameter:"simplify-text"`
	Markdown        bool     `glazed.parameter:"markdown"`
	MaxListItems    int      `glazed.parameter:"max-list-items"`
	MaxTableRows    int      `glazed.parameter:"max-table-rows"`
}

func NewHTMLSelectorCommand() (*HTMLSelectorCommand, error) {
	return &HTMLSelectorCommand{
		CommandDescription: cmds.NewCommandDescription(
			"select",
			cmds.WithShort("Test HTML/XPath selectors against HTML documents"),
			cmds.WithLong(`A tool for testing CSS and XPath selectors against HTML documents.
It provides match counts and contextual examples to verify selector accuracy.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"config",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML config file containing selectors"),
					parameters.WithRequired(false),
				),
				parameters.NewParameterDefinition(
					"select-css",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("CSS selectors to test (can be specified multiple times)"),
				),
				parameters.NewParameterDefinition(
					"select-xpath",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("XPath selectors to test (can be specified multiple times)"),
				),
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("HTML files to process (can be specified multiple times)"),
				),
				parameters.NewParameterDefinition(
					"urls",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("URLs to fetch and process (can be specified multiple times)"),
				),
				parameters.NewParameterDefinition(
					"extract",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Extract all matches into a YAML map of selector name to matches"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"extract-data",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Extract raw data without applying any templates"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"extract-template",
					parameters.ParameterTypeString,
					parameters.WithHelp("Go template file to render with extracted data"),
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
					parameters.WithDefault(3),
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

func (c *HTMLSelectorCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &HTMLSelectorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	var selectors []Selector

	// Load selectors from config file if provided
	var config *Config
	var err error
	if s.ConfigFile != "" {
		config, err = loadConfig(s.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		selectors = config.Selectors
	}

	// Add CSS selectors from command line
	for i, css := range s.SelectCSS {
		selectors = append(selectors, Selector{
			Name:     fmt.Sprintf("css_%d", i+1),
			Selector: css,
			Type:     "css",
		})
	}

	// Add XPath selectors from command line
	for i, xpath := range s.SelectXPath {
		selectors = append(selectors, Selector{
			Name:     fmt.Sprintf("xpath_%d", i+1),
			Selector: xpath,
			Type:     "xpath",
		})
	}

	// Ensure at least one selector is provided
	if len(selectors) == 0 {
		return fmt.Errorf("no selectors provided: use either --config or --select-css/--select-xpath")
	}

	// Ensure at least one source is provided
	if len(s.Files) == 0 && len(s.URLs) == 0 {
		return fmt.Errorf("no input sources provided: use either --files or --urls")
	}

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

	var sourceResults []SourceResult

	// Process files
	for _, file := range s.Files {
		result, err := processSource(ctx, file, selectors, s, simplifier)
		if err != nil {
			return fmt.Errorf("failed to process file %s: %w", file, err)
		}
		sourceResults = append(sourceResults, result)
	}

	// Process URLs
	for _, url := range s.URLs {
		result, err := processSource(ctx, url, selectors, s, simplifier)
		if err != nil {
			return fmt.Errorf("failed to process URL %s: %w", url, err)
		}
		sourceResults = append(sourceResults, result)
	}

	// If using extract or extract-template, process all matches without sample limit
	if s.Extract || s.ExtractTemplate != "" {
		// If extract-data is true, output raw data regardless of templates
		if s.ExtractData {
			return yaml.NewEncoder(w).Encode(sourceResults)
		}

		// First try command line template
		if s.ExtractTemplate != "" {
			// Load and execute template
			tmpl, err := template.New(s.ExtractTemplate).
				Funcs(sprig.TxtFuncMap()).
				ParseFiles(s.ExtractTemplate)
			if err != nil {
				return fmt.Errorf("failed to parse template file: %w", err)
			}
			return tmpl.Execute(w, sourceResults)
		}

		// Then try config file template if extract mode is on
		if config != nil && config.Config.Template != "" {
			// Parse and execute template from config
			tmpl, err := template.New("config").
				Funcs(sprig.TxtFuncMap()).
				Parse(config.Config.Template)
			if err != nil {
				return fmt.Errorf("failed to parse template from config: %w", err)
			}
			return tmpl.Execute(w, sourceResults)
		}

		// Default to YAML output
		return yaml.NewEncoder(w).Encode(sourceResults)
	}

	// Convert results to use Document structure for normal output
	newResults := make(map[string]*SimplifiedResult)
	for _, sourceResult := range sourceResults {
		for _, selectorResult := range sourceResult.SelectorResults {
			if _, ok := newResults[selectorResult.Name]; !ok {
				newResults[selectorResult.Name] = &SimplifiedResult{
					Name:     selectorResult.Name,
					Selector: selectorResult.Selector,
					Type:     selectorResult.Type,
					Count:    selectorResult.Count,
					Samples:  []SimplifiedSample{},
				}
			}

			for _, selectorSample := range selectorResult.Samples {
				htmlDocs, err := simplifier.ProcessHTML(selectorSample.HTML)
				if err != nil {
					return fmt.Errorf("failed to process HTML: %w", err)
				}

				sample := SimplifiedSample{
					HTML: htmlDocs,
				}

				if s.ShowPath {
					sample.Path = selectorSample.Path
				}
				if s.ShowContext {
					htmlDocs, err := simplifier.ProcessHTML(selectorSample.Context)
					if err != nil {
						return fmt.Errorf("failed to process HTML: %w", err)
					}
					sample.Context = htmlDocs
				}
				newResults[selectorResult.Name].Samples = append(newResults[selectorResult.Name].Samples, sample)
			}

		}
	}

	return yaml.NewEncoder(w).Encode(newResults)
}

func findSelectorByName(selectors []Selector, name string) Selector {
	for _, s := range selectors {
		if s.Name == name {
			return s
		}
	}
	return Selector{}
}

func processSource(
	ctx context.Context,
	source string,
	selectors []Selector,
	s *HTMLSelectorSettings,
	simplifier *htmlsimplifier.Simplifier,
) (SourceResult, error) {
	var result SourceResult
	result.Source = source

	var f io.ReadCloser
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return result, fmt.Errorf("failed to fetch URL: %w", err)
		}
		defer resp.Body.Close()
		f = resp.Body
	} else {
		f, err = os.Open(source)
		if err != nil {
			return result, fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
	}

	sampleCount := s.SampleCount
	if s.Extract || s.ExtractTemplate != "" {
		sampleCount = 0
	}

	tester, err := NewSelectorTester(&Config{
		File:      source,
		Selectors: selectors,
		Config: struct {
			SampleCount  int    `yaml:"sample_count"`
			ContextChars int    `yaml:"context_chars"`
			Template     string `yaml:"template"`
		}{
			SampleCount:  sampleCount,
			ContextChars: s.ContextChars,
			Template:     "",
		},
	}, f)
	if err != nil {
		return result, fmt.Errorf("failed to create tester: %w", err)
	}

	results, err := tester.Run(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to run tests: %w", err)
	}

	result.Data = make(map[string][]interface{})
	result.SelectorResults = results
	for _, r := range results {
		var matches []interface{}
		for _, selectorSample := range r.Samples {
			// Process HTML content
			htmlDocs, err := simplifier.ProcessHTML(selectorSample.HTML)
			if err == nil {
				for _, doc := range htmlDocs {
					if doc.Text != "" {
						matches = append(matches, doc.Text)
					} else if doc.Markdown != "" {
						matches = append(matches, doc.Markdown)
					} else {
						matches = append(matches, doc)
					}
				}
			}
		}
		result.Data[r.Name] = matches
	}

	return result, nil
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
		Use:   "html-selector",
		Short: "Run HTML/XPath selectors against HTML documents",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// reinitialize the logger because we can now parse --log-level and co
			// from the command line flag
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	err := clay.InitViper("html-selector", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)
	AddDocToHelpSystem(helpSystem)

	cmd, err := NewHTMLSelectorCommand()
	cobra.CheckErr(err)

	cobraCmd, err := cli.BuildCobraCommandFromWriterCommand(cmd)
	cobra.CheckErr(err)

	rootCmd.AddCommand(cobraCmd)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
