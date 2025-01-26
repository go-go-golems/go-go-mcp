package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

type Document struct {
	Tag        string            `yaml:"tag,omitempty"`
	Attributes map[string]string `yaml:"attributes,omitempty"`
	Content    string            `yaml:"content,omitempty"`
	Children   []Document        `yaml:"children,omitempty"`
}

type Options struct {
	StripScripts bool
	StripCSS     bool
	ShortenText  bool
}

func main() {
	var opts Options
	var rootCmd = &cobra.Command{
		Use:   "simplify-html [file]",
		Short: "Simplify and minimize HTML documents",
		Long: `A tool to simplify and minimize HTML documents by removing unnecessary elements 
and attributes, and shortening overly long text content. The output is provided 
in a structured YAML format.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return processHTML(args[0], opts)
		},
	}

	rootCmd.Flags().BoolVar(&opts.StripScripts, "strip-scripts", false, "Remove <script> tags")
	rootCmd.Flags().BoolVar(&opts.StripCSS, "strip-css", false, "Remove <style> tags and style attributes")
	rootCmd.Flags().BoolVar(&opts.ShortenText, "shorten-text", false, "Shorten <span> and <p> elements longer than 200 characters")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processHTML(filename string, opts Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := processNode(doc.Selection.Nodes[0], opts)

	yamlData, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	fmt.Println(string(yamlData))
	return nil
}

func processNode(node *html.Node, opts Options) Document {
	if node == nil {
		return Document{}
	}

	// Skip script tags if strip-scripts is enabled
	if opts.StripScripts && node.Data == "script" {
		return Document{}
	}

	// Skip style tags if strip-css is enabled
	if opts.StripCSS && node.Data == "style" {
		return Document{}
	}

	doc := Document{
		Tag:        node.Data,
		Attributes: make(map[string]string),
	}

	// Process attributes
	for _, attr := range node.Attr {
		if opts.StripCSS && attr.Key == "style" {
			continue
		}
		doc.Attributes[attr.Key] = attr.Val
	}

	// Process text content
	if node.Type == html.TextNode {
		content := strings.TrimSpace(node.Data)
		if content != "" {
			if opts.ShortenText && len(content) > 200 {
				content = content[:200] + "..."
			}
			doc.Content = content
		}
	}

	// Process children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.TextNode || strings.TrimSpace(child.Data) != "" {
			childDoc := processNode(child, opts)
			if childDoc.Tag != "" || childDoc.Content != "" {
				doc.Children = append(doc.Children, childDoc)
			}
		}
	}

	return doc
}
