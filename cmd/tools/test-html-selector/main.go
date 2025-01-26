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

	md "github.com/JohannesKaufmann/html-to-markdown"
)

type Config struct {
	File        string     `yaml:"file"`
	Description string     `yaml:"description"`
	Selectors   []Selector `yaml:"selectors"`
	Config      struct {
		SampleCount  int `yaml:"sample_count"`
		ContextChars int `yaml:"context_chars"`
	} `yaml:"config"`
	Template string `yaml:"template"`
}

type Selector struct {
	Name        string `yaml:"name"`
	Selector    string `yaml:"selector"`
	Type        string `yaml:"type"` // "css" or "xpath"
	Description string `yaml:"description"`
}

type SimplifiedSample struct {
	SimplifiedHTML    []htmlsimplifier.Document `yaml:"simplified_html,omitempty"`
	HTML              string                    `yaml:"html,omitempty"`
	SimplifiedContext []htmlsimplifier.Document `yaml:"simplified_context,omitempty"`
	Context           string                    `yaml:"context,omitempty"`
	Markdown          string                    `yaml:"markdown,omitempty"`
	Path              string                    `yaml:"path,omitempty"`
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
	Data            map[string][]interface{} `yaml:"data,omitempty"`
	SelectorResults []SelectorResult         `yaml:"selector_results,omitempty"`
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
	NoTemplate      bool     `glazed.parameter:"no-template"`
	ShowContext     bool     `glazed.parameter:"show-context"`
	ShowPath        bool     `glazed.parameter:"show-path"`
	ShowSimplified  bool     `glazed.parameter:"show-simplified"`
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

func (s *HTMLSelectorSettings) ShouldTemplate() bool {
	if s.NoTemplate {
		return false
	}
	return s.ExtractData || s.ExtractTemplate != "" || (s.ConfigFile != "" && s.ConfigFile != "-")
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
					"no-template",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Do not use templates"),
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
					"show-simplified",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show simplified HTML in output"),
					parameters.WithDefault(false),
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

	var sourceResults []*SourceResult

	// Process files
	for _, file := range s.Files {
		result, err := processSource(ctx, file, selectors, s)
		if err != nil {
			return fmt.Errorf("failed to process file %s: %w", file, err)
		}
		sourceResults = append(sourceResults, result)
	}

	// Process URLs
	for _, url := range s.URLs {
		result, err := processSource(ctx, url, selectors, s)
		if err != nil {
			return fmt.Errorf("failed to process URL %s: %w", url, err)
		}
		sourceResults = append(sourceResults, result)
	}

	if s.ShouldTemplate() {
		// clear the selector results
		for _, sourceResult := range sourceResults {
			sourceResult.SelectorResults = []SelectorResult{}
		}

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
			return executeTemplate(w, tmpl, sourceResults)
		}

		// Then try config file template if extract mode is on
		if config != nil && config.Template != "" {
			// Parse and execute template from config
			tmpl, err := template.New("config").
				Funcs(sprig.TxtFuncMap()).
				Parse(config.Template)
			if err != nil {
				return fmt.Errorf("failed to parse template from config: %w", err)
			}
			return executeTemplate(w, tmpl, sourceResults)
		}

		// Default to YAML output
		return yaml.NewEncoder(w).Encode(sourceResults)
	}

	// Create markdown converter
	converter := md.NewConverter("", true, nil)

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

				markdown, err := converter.ConvertString(selectorSample.HTML)
				if err == nil {
				}

				sample := SimplifiedSample{
					HTML:     selectorSample.HTML,
					Markdown: markdown,
				}

				if s.ShowSimplified {
					sample.SimplifiedHTML = htmlDocs
				}

				if s.ShowPath {
					sample.Path = selectorSample.Path
				}
				if s.ShowContext {
					htmlDocs, err := simplifier.ProcessHTML(selectorSample.Context)
					if err != nil {
						return fmt.Errorf("failed to process HTML: %w", err)
					}
					if s.ShowSimplified {
						sample.SimplifiedContext = htmlDocs
					}
					sample.Context = selectorSample.Context
				}
				newResults[selectorResult.Name].Samples = append(newResults[selectorResult.Name].Samples, sample)
			}

		}
	}

	return yaml.NewEncoder(w).Encode(newResults)
}

func processSource(
	ctx context.Context,
	source string,
	selectors []Selector,
	s *HTMLSelectorSettings,
) (*SourceResult, error) {
	result := &SourceResult{
		Source: source,
	}

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
	if s.ShouldTemplate() {
		sampleCount = 0
	}

	tester, err := NewSelectorTester(&Config{
		File:      source,
		Selectors: selectors,
		Config: struct {
			SampleCount  int `yaml:"sample_count"`
			ContextChars int `yaml:"context_chars"`
		}{
			SampleCount:  sampleCount,
			ContextChars: s.ContextChars,
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
			// Convert sample to markdown if requested
			if s.Markdown {
				// Create markdown converter
				converter := md.NewConverter("", true, nil)
				var markdown string

				// Convert HTML to markdown if present
				if selectorSample.HTML != "" {
					markdown, err = converter.ConvertString(selectorSample.HTML)
					if err == nil {
						matches = append(matches, markdown)
						continue
					}
				}
			}
			matches = append(matches, selectorSample.HTML)
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

// executeTemplate handles template execution and provides a subset of data on error
func executeTemplate(w io.Writer, tmpl *template.Template, sourceResults []*SourceResult) error {
	// First try executing the template with all source results
	err := tmpl.Execute(w, sourceResults)
	if err == nil {
		return nil
	}

	// If that fails, try executing individually for each source
	fmt.Fprintf(os.Stderr, "Error executing combined template: %v\n", err)
	fmt.Fprintf(os.Stderr, "Trying individual execution...\n")

	for i, sr := range sourceResults {
		if i > 0 {
			fmt.Fprintf(w, "\n---\n")
		}
		fmt.Fprintf(w, "# Source: %s\n", sr.Source)

		err := tmpl.Execute(w, sr)
		if err != nil {
			// Create subset of failed source result for error reporting
			subsetResult := &SourceResult{
				Source: sr.Source,
				Data:   make(map[string][]interface{}),
			}

			// Take first 3 samples for each selector
			for name, matches := range sr.Data {
				if len(matches) > 3 {
					subsetResult.Data[name] = matches[:3]
				} else {
					subsetResult.Data[name] = matches
				}
			}

			// Print the error and data subset
			fmt.Fprintf(os.Stderr, "Error executing template for source %s: %v\n", sr.Source, err)
			fmt.Fprintf(os.Stderr, "Here is a subset of the input data:\n")
			enc := yaml.NewEncoder(os.Stderr)
			enc.SetIndent(2)
			if err := enc.Encode(subsetResult); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding data subset: %v\n", err)
			}
			return fmt.Errorf("template execution failed for source %s: %w", sr.Source, err)
		}
	}

	return nil
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
