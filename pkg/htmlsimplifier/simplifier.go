package htmlsimplifier

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
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
	return &Simplifier{
		opts:           opts,
		textSimplifier: NewTextSimplifier(opts.Markdown),
		nodeHandler:    NewNodeHandler(opts),
	}
}

// ProcessHTML simplifies the given HTML content according to the configured options
func (s *Simplifier) ProcessHTML(htmlContent string) (Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return Document{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Apply selector-based filtering if config is provided
	if s.opts.FilterConfig != nil {
		log.Debug().Msg("Applying selector-based filtering")
		for _, sel := range s.opts.FilterConfig.Selectors {
			log.Debug().Str("type", sel.Type).Str("selector", sel.Selector).Msg("Applying selector")
			switch sel.Type {
			case "css":
				doc.Find(sel.Selector).Remove()
			case "xpath":
				nodes, err := htmlquery.QueryAll(doc.Get(0), sel.Selector)
				if err != nil {
					return Document{}, fmt.Errorf("failed to execute XPath selector '%s': %w", sel.Selector, err)
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

	result := s.processNode(doc.Get(0))
	return result, nil
}

func (s *Simplifier) processNode(node *html.Node) Document {
	if node == nil {
		return Document{}
	}

	strategy := s.nodeHandler.GetStrategy(node)
	log.Trace().Str("tag", node.Data).Str("strategy", strategy.String()).Msg("Processing node")

	switch strategy {
	case StrategyFilter:
		return Document{}

	case StrategyUnwrap:
		// Process children and combine them
		var children []Document
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if childDoc := s.processNode(child); !childDoc.IsEmpty() {
				children = append(children, childDoc)
			}
		}
		if len(children) == 1 {
			return children[0]
		}
		return Document{Children: children}

	case StrategyTextOnly:
		if s.nodeHandler.IsTextOnly(node) {
			text := s.textSimplifier.ExtractText(node)
			return Document{Text: text}
		}
		// Fall through to default if not all children are text

	case StrategyPreserveWhitespace:
		if node.Type == html.TextNode {
			return Document{Text: node.Data}
		}
		// Fall through to default for element nodes

	case StrategyMarkdown:
		if s.nodeHandler.IsMarkdownable(node) {
			markdown, ok := s.textSimplifier.ConvertToMarkdown(node)
			if ok {
				return Document{Markdown: markdown}
			}
		}
		// Fall through to default if not all children are markdownable
	}

	// Default processing: keep the node and process children
	doc := Document{
		Tag:   node.Data,
		IsSVG: node.Data == "svg" || (node.Parent != nil && node.Parent.Data == "svg"),
	}

	// Process attributes
	var attrs []string
	for _, attr := range node.Attr {
		if s.opts.StripCSS && attr.Key == "style" {
			continue
		}
		if s.opts.CompactSVG && doc.IsSVG && (attr.Key == "d" || attr.Key == "viewBox" || attr.Key == "transform") {
			continue
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", attr.Key, attr.Val))
	}
	if len(attrs) > 0 {
		doc.Attrs = strings.Join(attrs, " ")
	}

	// Process children
	var children []Document
	itemCount := 0
	isList := node.Data == "ul" || node.Data == "ol" || node.Data == "select"
	isTable := node.Data == "table" || node.Data == "tbody"

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childDoc := s.processNode(child)
		if !childDoc.IsEmpty() {
			if (isList || isTable) && s.opts.MaxListItems > 0 {
				itemCount++
				if itemCount > s.opts.MaxListItems {
					if itemCount == s.opts.MaxListItems+1 {
						children = append(children, Document{Text: "..."})
					}
					continue
				}
			}
			children = append(children, childDoc)
		}
	}

	if len(children) > 0 {
		doc.Children = children
	}

	return doc
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
