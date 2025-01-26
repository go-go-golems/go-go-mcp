package htmlsimplifier

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type SelectorMode string

const (
	SelectorModeSelect SelectorMode = "select"
	SelectorModeFilter SelectorMode = "filter"
)

type Selector struct {
	Type     string       `yaml:"type"`     // "css" or "xpath"
	Mode     SelectorMode `yaml:"mode"`     // "select" or "filter"
	Selector string       `yaml:"selector"` // The actual selector string
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
	SimplifyText bool
	Markdown     bool // Convert text with important elements to markdown
}

// Simplifier handles HTML simplification with configurable options
type Simplifier struct {
	opts           Options
	textSimplifier *TextSimplifier
	nodeHandler    *NodeHandler
}

// NewSimplifier creates a new HTML simplifier with the given options
func NewSimplifier(opts Options) *Simplifier {
	ret := &Simplifier{
		opts:           opts,
		textSimplifier: NewTextSimplifier(opts.Markdown),
		nodeHandler:    NewNodeHandler(opts),
	}
	// only output markdown if markdown is requested
	if opts.SimplifyText && opts.Markdown {
		opts.SimplifyText = false
	}
	return ret
}

// ProcessHTML simplifies the given HTML content according to the configured options
func (s *Simplifier) ProcessHTML(htmlContent string) ([]Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Apply selector-based filtering if config is provided
	if s.opts.FilterConfig != nil {
		log.Debug().Msg("Applying selector-based filtering")

		// First apply select selectors - mark all matching elements
		hasSelects := false
		for _, sel := range s.opts.FilterConfig.Selectors {
			if sel.Mode == SelectorModeSelect {
				hasSelects = true
				log.Debug().Str("type", sel.Type).Str("selector", sel.Selector).Msg("Marking selected elements")
				switch sel.Type {
				case "css":
					doc.Find(sel.Selector).Each(func(i int, s *goquery.Selection) {
						s.SetAttr("data-simplifier-keep", "true")
						// Mark all parents as well
						s.Parents().Each(func(i int, p *goquery.Selection) {
							p.SetAttr("data-simplifier-keep", "true")
						})
					})
				case "xpath":
					nodes, err := htmlquery.QueryAll(doc.Get(0), sel.Selector)
					if err != nil {
						return nil, fmt.Errorf("failed to execute XPath selector '%s': %w", sel.Selector, err)
					}
					for _, node := range nodes {
						// Add attribute to mark this node
						for n := node; n != nil && n.Type == html.ElementNode; n = n.Parent {
							found := false
							for i := range n.Attr {
								if n.Attr[i].Key == "data-simplifier-keep" {
									found = true
									break
								}
							}
							if !found {
								n.Attr = append(n.Attr, html.Attribute{Key: "data-simplifier-keep", Val: "true"})
							}
						}
					}
				}
			}
		}

		// If we have any select selectors, remove everything that wasn't marked
		if hasSelects {
			doc.Find("*").Each(func(i int, s *goquery.Selection) {
				if _, exists := s.Attr("data-simplifier-keep"); !exists {
					s.Remove()
				}
			})
		}

		// Then apply filter selectors
		for _, sel := range s.opts.FilterConfig.Selectors {
			if sel.Mode == SelectorModeFilter {
				log.Debug().Str("type", sel.Type).Str("selector", sel.Selector).Msg("Applying filter selector")
				switch sel.Type {
				case "css":
					doc.Find(sel.Selector).Remove()
				case "xpath":
					nodes, err := htmlquery.QueryAll(doc.Get(0), sel.Selector)
					if err != nil {
						return nil, fmt.Errorf("failed to execute XPath selector '%s': %w", sel.Selector, err)
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

		// Clean up temporary attributes
		doc.Find("[data-simplifier-keep]").RemoveAttr("data-simplifier-keep")
	}

	return s.processNode(doc.Get(0)), nil
}

func (s *Simplifier) processNode(node *html.Node) []Document {
	if node == nil {
		return nil
	}

	strategy := s.nodeHandler.GetStrategy(node)
	log.Trace().Str("tag", node.Data).Str("strategy", strategy.String()).Msg("Processing node")

	// Process attributes for all nodes
	var attrs []string
	var classes []string
	var id string
	for _, attr := range node.Attr {
		if s.opts.StripCSS && attr.Key == "style" {
			continue
		}
		if s.opts.CompactSVG && (node.Data == "svg" || (node.Parent != nil && node.Parent.Data == "svg")) &&
			(attr.Key == "d" || attr.Key == "viewBox" || attr.Key == "transform") {
			continue
		}
		if attr.Key == "class" {
			classes = strings.Fields(attr.Val)
			continue
		}
		if attr.Key == "id" {
			id = attr.Val
			continue
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", attr.Key, attr.Val))
	}
	attrsStr := strings.Join(attrs, " ")

	// Handle text nodes first
	if node.Type == html.TextNode {
		// Skip pure whitespace nodes unless in preserve whitespace mode
		if strategy != StrategyPreserveWhitespace && strings.TrimSpace(node.Data) == "" {
			return nil
		}
		return []Document{{
			Tag:   "#text",
			Attrs: attrsStr,
			Text:  node.Data,
		}}
	}

	// Build tag name with id and classes
	tagName := node.Data
	if id != "" {
		tagName = fmt.Sprintf("%s#%s", tagName, id)
	}
	for _, class := range classes {
		tagName = fmt.Sprintf("%s.%s", tagName, class)
	}

	switch strategy {
	case StrategyFilter:
		return nil

	case StrategyUnwrap:
		// Process children and combine them
		var result []Document
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			result = append(result, s.processNode(child)...)
		}
		return result

	case StrategyTextOnly:
		if s.opts.Markdown && s.nodeHandler.IsMarkdownable(node) {
			if markdown, ok := s.textSimplifier.ConvertToMarkdown(node); ok {
				return []Document{{
					Tag:      tagName,
					Attrs:    attrsStr,
					Markdown: markdown,
				}}
			}
		}

		if s.opts.SimplifyText && s.nodeHandler.IsTextOnly(node) {
			if text, ok := s.textSimplifier.SimplifyText(node); ok {
				return []Document{{
					Tag:   tagName,
					Attrs: attrsStr,
					Text:  text,
				}}
			}
		}
		// If text simplification fails or is disabled, extract text normally
		text := s.textSimplifier.ExtractText(node)
		if text != "" {
			return []Document{{
				Tag:   tagName,
				Attrs: attrsStr,
				Text:  text,
			}}
		}
		// Fall through to default if text extraction yields nothing

	case StrategyPreserveWhitespace:
		if node.Type == html.TextNode {
			return []Document{{
				Tag:   "#text",
				Attrs: attrsStr,
				Text:  node.Data,
			}}
		}
		// For element nodes with preserved whitespace, process children normally
		// but maintain the original structure

	case StrategyMarkdown:
		if s.opts.Markdown && s.nodeHandler.IsMarkdownable(node) {
			if markdown, ok := s.textSimplifier.ConvertToMarkdown(node); ok {
				return []Document{{
					Tag:      tagName,
					Attrs:    attrsStr,
					Markdown: markdown,
				}}
			}
		}
		// Fall through to default if markdown conversion fails

	case StrategyDefault:
		// Check if all children are markdown-able
		if s.opts.Markdown {
			if docs, ok := s.tryMarkdownConversion(node, tagName, attrsStr); ok {
				return docs
			}
		}

		// Check if all children are text-only
		if s.opts.SimplifyText {
			if docs, ok := s.tryTextSimplification(node, tagName, attrsStr); ok {
				return docs
			}
		}
		// Fall through to default if text simplification fails
	}

	// Default processing: keep the node and process children
	doc := Document{
		Tag:   tagName,
		Attrs: attrsStr,
		IsSVG: node.Data == "svg" || (node.Parent != nil && node.Parent.Data == "svg"),
	}

	// Process children
	var children []Document
	itemCount := 0
	isList := node.Data == "ul" || node.Data == "ol" || node.Data == "select"
	isTable := node.Data == "table" || node.Data == "tbody"

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childDocs := s.processNode(child)
		for _, childDoc := range childDocs {
			if !childDoc.IsEmpty() {
				if (isList || isTable) && s.opts.MaxListItems > 0 {
					itemCount++
					if itemCount > s.opts.MaxListItems {
						if itemCount == s.opts.MaxListItems+1 {
							children = append(children, Document{
								Text: "...",
							})
						}
						continue
					}
				}
				children = append(children, childDoc)
			}
		}
	}

	if len(children) > 0 {
		doc.Children = children
	}

	return []Document{doc}
}

func (s *Simplifier) tryMarkdownConversion(node *html.Node, tagName string, attrsStr string) ([]Document, bool) {
	allMarkdownable := true
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !s.nodeHandler.IsMarkdownable(child) {
			allMarkdownable = false
			break
		}
	}
	if allMarkdownable {
		markdown, ok := s.textSimplifier.ConvertToMarkdown(node)
		if ok {
			return []Document{{
				Tag:      tagName,
				Attrs:    attrsStr,
				Markdown: markdown,
			}}, true
		}
	}
	return nil, false
}

func (s *Simplifier) tryTextSimplification(node *html.Node, tagName string, attrsStr string) ([]Document, bool) {
	allTextable := true
	var textParts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !s.nodeHandler.IsTextOnly(child) {
			allTextable = false
			break
		}
		if text, ok := s.textSimplifier.SimplifyText(child); ok {
			textParts = append(textParts, text)
		}
	}
	if allTextable && len(textParts) > 0 {
		return []Document{{
			Tag:   tagName,
			Attrs: attrsStr,
			Text:  strings.Join(textParts, " "),
		}}, true
	}
	return nil, false
}

// IsEmpty returns true if the document is empty (no content)
func (d Document) IsEmpty() bool {
	return d.Tag == "" && d.Text == "" && d.Markdown == "" && len(d.Children) == 0
}

// String returns a string representation of the strategy
func (s NodeHandlingStrategy) String() string {
	switch s {
	case StrategyDefault:
		return "default"
	case StrategyUnwrap:
		return "unwrap"
	case StrategyFilter:
		return "filter"
	case StrategyTextOnly:
		return "text-only"
	case StrategyMarkdown:
		return "markdown"
	case StrategyPreserveWhitespace:
		return "preserve-whitespace"
	default:
		return "unknown"
	}
}
