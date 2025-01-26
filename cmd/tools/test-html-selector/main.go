package main

import (
	"fmt"
	"log"
	"os"

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

var (
	configFile string
	inputFile  string
	extract    bool
)

var rootCmd = &cobra.Command{
	Use:   "test-html-selector",
	Short: "Test HTML/XPath selectors against HTML documents",
	Long: `A tool for testing CSS and XPath selectors against HTML documents.
It provides match counts and contextual examples to verify selector accuracy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override file from config if input file is provided
		if inputFile != "" {
			config.File = inputFile
		}

		if config.File == "" {
			return fmt.Errorf("HTML input file is required (either in config or via --input flag)")
		}

		tester, err := NewSelectorTester(config)
		if err != nil {
			return fmt.Errorf("failed to create tester: %w", err)
		}

		results, err := tester.Run(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to run tests: %w", err)
		}

		if extract {
			for _, result := range results {
				fmt.Printf("Selector: %s\n", result.Selector)
				for _, sample := range result.Samples {
					fmt.Println(sample.HTML)
				}
			}
		} else {
			return yaml.NewEncoder(os.Stdout).Encode(results)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to YAML config file")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Path to HTML input file (overrides config file)")
	rootCmd.PersistentFlags().BoolVarP(&extract, "extract", "e", false, "Extract and print all matches for each selector")
	rootCmd.MarkPersistentFlagRequired("config")
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
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
