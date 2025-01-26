package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

type Selector struct {
	Type     string `yaml:"type"`     // "css" or "xpath"
	Selector string `yaml:"selector"` // The actual selector string
}

type FilterConfig struct {
	Selectors []Selector `yaml:"selectors"`
}

type Document struct {
	Tag      string     `yaml:"tag,omitempty"`
	Attrs    string     `yaml:"attrs,omitempty"`    // Simplified attributes as space-separated key=value pairs
	Text     string     `yaml:"text,omitempty"`     // For text-only nodes
	Markdown string     `yaml:"markdown,omitempty"` // For markdown-converted content
	IsSVG    bool       `yaml:"svg,omitempty"`      // Mark SVG elements to potentially skip details
	Children []Document `yaml:"children,omitempty"`
}

type Options struct {
	StripScripts bool
	StripCSS     bool
	ShortenText  bool
	CompactSVG   bool
	StripSVG     bool
	MaxListItems int
	MaxTableRows int
	FilterConfig *FilterConfig
	ConfigFile   string
	SimplifyText bool
	Debug        bool
	LogLevel     string
	Markdown     bool // Convert text with important elements to markdown
}

func loadFilterConfig(filename string) (*FilterConfig, error) {
	if filename == "" {
		log.Debug().Msg("No config file specified")
		return nil, nil
	}

	log.Debug().Str("filename", filename).Msg("Loading filter config")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FilterConfig
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

func setupLogging(opts Options) {
	// Default to no logging unless debug is enabled
	if !opts.Debug {
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
	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Warn().Msgf("Invalid log level %q, defaulting to debug", opts.LogLevel)
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
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
			setupLogging(opts)
			log.Debug().Msg("Starting HTML simplification")

			if opts.ConfigFile != "" {
				config, err := loadFilterConfig(opts.ConfigFile)
				if err != nil {
					return err
				}
				opts.FilterConfig = config
			}
			return processHTML(args[0], opts)
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
	rootCmd.Flags().StringVar(&opts.ConfigFile, "config", "", "Path to YAML config file containing selectors to filter out")
	rootCmd.Flags().BoolVar(&opts.Debug, "debug", false, "Enable debug logging")
	rootCmd.Flags().StringVar(&opts.LogLevel, "log-level", "debug", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&opts.Markdown, "markdown", true, "Convert text with important elements to markdown format")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processHTML(filename string, opts Options) error {
	log.Debug().Str("filename", filename).Msg("Processing HTML file")

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Apply selector-based filtering if config is provided
	if opts.FilterConfig != nil {
		log.Debug().Msg("Applying selector-based filtering")
		for _, sel := range opts.FilterConfig.Selectors {
			log.Debug().Str("type", sel.Type).Str("selector", sel.Selector).Msg("Applying selector")
			switch sel.Type {
			case "css":
				doc.Find(sel.Selector).Remove()
			case "xpath":
				nodes, err := htmlquery.QueryAll(doc.Get(0), sel.Selector)
				if err != nil {
					return fmt.Errorf("failed to execute XPath selector '%s': %w", sel.Selector, err)
				}
				log.Debug().Int("removed_nodes", len(nodes)).Msg("Removed nodes by XPath selector")
				for _, node := range nodes {
					if node.Parent != nil {
						node.Parent.RemoveChild(node)
					}
				}
			}
		}
	}

	result := processNode(doc.Selection.Nodes[0], opts)

	yamlData, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	log.Debug().Int("yaml_bytes", len(yamlData)).Msg("Generated YAML output")
	fmt.Println(string(yamlData))
	return nil
}

func processNode(node *html.Node, opts Options) Document {
	if node == nil {
		return Document{}
	}

	// Skip script tags if strip-scripts is enabled
	if opts.StripScripts && node.Data == "script" {
		log.Trace().Msg("Skipping script tag")
		return Document{}
	}

	// Skip style tags if strip-css is enabled
	if opts.StripCSS && node.Data == "style" {
		log.Trace().Msg("Skipping style tag")
		return Document{}
	}

	// Skip SVG tags if strip-svg is enabled
	if opts.StripSVG && (node.Data == "svg" || (node.Parent != nil && node.Parent.Data == "svg")) {
		log.Trace().Msg("Skipping SVG tag")
		return Document{}
	}

	// For text nodes, return just the content
	if node.Type == html.TextNode {
		content := strings.TrimSpace(node.Data)
		if content != "" {
			if opts.ShortenText && len(content) > 200 {
				log.Trace().Int("original_length", len(content)).Msg("Shortening text content")
				content = content[:200] + "..."
			}
			return Document{Text: content}
		}
		return Document{}
	}

	doc := Document{
		Tag:   node.Data,
		IsSVG: node.Data == "svg" || (node.Parent != nil && node.Parent.Data == "svg"),
	}

	log.Trace().Str("tag", doc.Tag).Bool("is_svg", doc.IsSVG).Msg("Processing node")

	// Process attributes into a compact string
	var attrs []string
	for _, attr := range node.Attr {
		if opts.StripCSS && attr.Key == "style" {
			continue
		}
		// Skip detailed SVG attributes if compact mode is enabled
		if opts.CompactSVG && doc.IsSVG && (attr.Key == "d" || attr.Key == "viewBox" || attr.Key == "transform") {
			continue
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", attr.Key, attr.Val))
	}
	if len(attrs) > 0 {
		doc.Attrs = strings.Join(attrs, " ")
		log.Trace().Int("attr_count", len(attrs)).Msg("Processed attributes")
	}

	// Process children
	var children []Document
	var textParts []string
	var markdownParts []string
	hasOnlyTextAndBr := true
	hasImportantElements := false
	itemCount := 0
	isList := node.Data == "ul" || node.Data == "ol" || node.Data == "select"
	isTable := node.Data == "table" || node.Data == "tbody"

	// Important elements that should never be collapsed
	importantElements := map[string]bool{
		"a":      true, // Links
		"strong": true, // Bold text
		"em":     true, // Italic text
		"b":      true, // Bold text (alternative)
		"i":      true, // Italic text (alternative)
		"code":   true, // Code snippets
		"span":   true, // Spans with attributes
		"img":    true, // Images
		"sub":    true, // Subscript
		"sup":    true, // Superscript
		"mark":   true, // Highlighted text
		"time":   true, // Time elements
		"abbr":   true, // Abbreviations
	}

	// Markdown conversion mappings
	markdownStart := map[string]string{
		"a":      "[",  // Links
		"strong": "**", // Bold text
		"em":     "*",  // Italic text
		"b":      "**", // Bold text (alternative)
		"i":      "*",  // Italic text (alternative)
		"code":   "`",  // Code snippets
		"mark":   "==", // Highlighted text (some markdown flavors)
		"sub":    "~",  // Subscript (some markdown flavors)
		"sup":    "^",  // Superscript (some markdown flavors)
	}
	markdownEnd := map[string]string{
		"a":      "](%s)", // Links (format with href)
		"strong": "**",    // Bold text
		"em":     "*",     // Italic text
		"b":      "**",    // Bold text (alternative)
		"i":      "*",     // Italic text (alternative)
		"code":   "`",     // Code snippets
		"mark":   "==",    // Highlighted text
		"sub":    "~",     // Subscript
		"sup":    "^",     // Superscript
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childDoc := processNode(child, opts)

		// Check if this child is not text or br
		if child.Type != html.TextNode && child.Data != "br" {
			hasOnlyTextAndBr = false
			log.Trace().
				Int("node_type", int(child.Type)).
				Str("node_data", child.Data).
				Str("parent_tag", node.Data).
				Msg("Found non-text/br child, will not simplify")
		}

		// Check if this is an important element that should be preserved
		if _, isImportant := importantElements[child.Data]; isImportant {
			hasImportantElements = true
			log.Trace().
				Str("element_type", child.Data).
				Str("parent_tag", node.Data).
				Msg("Found important element that should be preserved")
		}

		if childDoc.Tag != "" || childDoc.Text != "" {
			// Handle markdown conversion for important elements
			if opts.Markdown && childDoc.Text != "" {
				if start, ok := markdownStart[child.Data]; ok {
					text := childDoc.Text
					if child.Data == "a" {
						// Find href attribute for links
						for _, attr := range child.Attr {
							if attr.Key == "href" {
								text = text + fmt.Sprintf(markdownEnd[child.Data], attr.Val)
								break
							}
						}
					} else if end, ok := markdownEnd[child.Data]; ok {
						text = start + text + end
					}
					childDoc.Markdown = text
					log.Trace().
						Str("element", child.Data).
						Str("markdown", text).
						Msg("Converted element to markdown")
				}
			}

			// Never try to simplify content with important elements unless we're using markdown
			if opts.SimplifyText && (!hasImportantElements || opts.Markdown) && (childDoc.Text != "" || childDoc.Tag == "br") {
				log.Trace().
					Str("parent_tag", node.Data).
					Str("child_tag", childDoc.Tag).
					Str("child_text", childDoc.Text).
					Bool("is_br", childDoc.Tag == "br").
					Msg("Collecting text for simplification")

				if childDoc.Tag == "br" {
					textParts = append(textParts, "\n")
					markdownParts = append(markdownParts, "\n")
				} else {
					textParts = append(textParts, childDoc.Text)
					if childDoc.Markdown != "" {
						markdownParts = append(markdownParts, childDoc.Markdown)
					} else {
						markdownParts = append(markdownParts, childDoc.Text)
					}
				}
				continue
			}

			// For list items and options, check if we've reached the limit
			if opts.MaxListItems > 0 && isList &&
				(child.Data == "li" || child.Data == "option") {
				itemCount++
				if itemCount > opts.MaxListItems {
					// Add an ellipsis item if this is the first item we're skipping
					if itemCount == opts.MaxListItems+1 {
						log.Debug().Int("max_items", opts.MaxListItems).Msg("List item limit reached")
						children = append(children, Document{
							Tag:  child.Data,
							Text: "...",
						})
					}
					continue
				}
			}

			// For table rows, check if we've reached the limit
			if opts.MaxTableRows > 0 && isTable && child.Data == "tr" {
				itemCount++
				if itemCount > opts.MaxTableRows {
					// Add an ellipsis row if this is the first row we're skipping
					if itemCount == opts.MaxTableRows+1 {
						log.Debug().Int("max_rows", opts.MaxTableRows).Msg("Table row limit reached")
						children = append(children, Document{
							Tag: "tr",
							Children: []Document{{
								Tag:  "td",
								Text: "...",
							}},
						})
					}
					continue
				}
			}

			children = append(children, childDoc)
		}
	}

	// If we have only text and br nodes and simplification is enabled, use text field
	// But only if there are no important elements to preserve (or we're using markdown)
	if opts.SimplifyText && (!hasImportantElements || opts.Markdown) && hasOnlyTextAndBr && len(textParts) > 0 {
		log.Trace().
			Str("tag", node.Data).
			Bool("has_only_text_br", hasOnlyTextAndBr).
			Bool("has_important_elements", hasImportantElements).
			Bool("using_markdown", opts.Markdown).
			Int("text_parts", len(textParts)).
			Strs("parts", textParts).
			Msg("Attempting text simplification")

		doc.Text = strings.TrimSpace(strings.Join(textParts, ""))
		if opts.Markdown && len(markdownParts) > 0 {
			doc.Markdown = strings.TrimSpace(strings.Join(markdownParts, ""))
		}
		if opts.ShortenText && len(doc.Text) > 200 {
			log.Trace().Int("original_length", len(doc.Text)).Msg("Shortening simplified text")
			doc.Text = doc.Text[:200] + "..."
			if doc.Markdown != "" {
				doc.Markdown = doc.Markdown[:200] + "..."
			}
		}
		log.Trace().
			Str("tag", node.Data).
			Str("final_text", doc.Text).
			Str("final_markdown", doc.Markdown).
			Msg("Simplified text node")
	} else {
		if opts.SimplifyText {
			log.Trace().
				Str("tag", node.Data).
				Bool("has_only_text_br", hasOnlyTextAndBr).
				Bool("has_important_elements", hasImportantElements).
				Bool("using_markdown", opts.Markdown).
				Int("text_parts", len(textParts)).
				Int("children", len(children)).
				Msg("Not simplifying node")
		}
		doc.Children = children
		log.Trace().Int("child_count", len(children)).Msg("Added children to node")
	}

	return doc
}
