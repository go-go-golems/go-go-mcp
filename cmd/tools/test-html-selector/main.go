package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
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

type TestHTMLSelectorCommand struct {
	*cmds.CommandDescription
}

type TestHTMLSelectorSettings struct {
	ConfigFile   string `glazed.parameter:"config"`
	InputFile    string `glazed.parameter:"input"`
	Extract      bool   `glazed.parameter:"extract"`
	SampleCount  int    `glazed.parameter:"sample-count"`
	ContextChars int    `glazed.parameter:"context-chars"`
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

	if s.Extract {
		for _, result := range results {
			fmt.Fprintf(w, "Selector: %s\n", result.Selector)
			for _, sample := range result.Samples {
				fmt.Fprintln(w, sample.HTML)
			}
		}
	} else {
		return yaml.NewEncoder(w).Encode(results)
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
	}

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
