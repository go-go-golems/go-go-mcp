package htmlsimplifier

import (
	"golang.org/x/net/html"
)

// NodeHandlingStrategy defines how a node should be processed
type NodeHandlingStrategy int

const (
	// StrategyDefault processes the node normally, keeping its tag and attributes
	StrategyDefault NodeHandlingStrategy = iota

	// StrategyUnwrap removes the node but keeps its children
	StrategyUnwrap

	// StrategyFilter removes the node and all its children
	StrategyFilter

	// StrategyTextOnly converts the node and its children to text if possible
	StrategyTextOnly

	// StrategyMarkdown converts the node and its children to markdown if possible
	StrategyMarkdown

	// StrategyPreserveWhitespace keeps all whitespace in text nodes
	StrategyPreserveWhitespace
)

// NodeHandler determines how to process different types of nodes
type NodeHandler struct {
	// Map of tag names to their handling strategies
	tagStrategies map[string]NodeHandlingStrategy

	// Default strategy for unknown tags
	defaultStrategy NodeHandlingStrategy

	// Whether markdown conversion is enabled
	markdownEnabled bool
}

// NewNodeHandler creates a new NodeHandler with the given options
func NewNodeHandler(opts Options) *NodeHandler {
	h := &NodeHandler{
		tagStrategies:   make(map[string]NodeHandlingStrategy),
		defaultStrategy: StrategyDefault,
		markdownEnabled: opts.Markdown,
	}

	// Configure default strategies
	h.tagStrategies["html"] = StrategyUnwrap
	h.tagStrategies["head"] = StrategyFilter
	h.tagStrategies["body"] = StrategyDefault

	if opts.StripScripts {
		h.tagStrategies["script"] = StrategyFilter
	}
	if opts.StripCSS {
		h.tagStrategies["style"] = StrategyFilter
	}
	if opts.StripSVG {
		h.tagStrategies["svg"] = StrategyFilter
	}

	// Text-focused elements
	h.tagStrategies["p"] = StrategyTextOnly
	h.tagStrategies["span"] = StrategyTextOnly
	h.tagStrategies["br"] = StrategyTextOnly
	h.tagStrategies["title"] = StrategyTextOnly

	// Pre-formatted text
	h.tagStrategies["pre"] = StrategyPreserveWhitespace

	// Markdown-capable elements (if enabled)
	if opts.Markdown {
		h.tagStrategies["strong"] = StrategyMarkdown
		h.tagStrategies["b"] = StrategyMarkdown
		h.tagStrategies["em"] = StrategyMarkdown
		h.tagStrategies["i"] = StrategyMarkdown
		h.tagStrategies["a"] = StrategyMarkdown
		h.tagStrategies["code"] = StrategyMarkdown
	}

	return h
}

// GetStrategy returns the handling strategy for a given node
func (h *NodeHandler) GetStrategy(node *html.Node) NodeHandlingStrategy {
	if node == nil {
		return StrategyFilter
	}

	switch node.Type {
	case html.TextNode:
		// Check if we're in a pre tag
		for parent := node.Parent; parent != nil; parent = parent.Parent {
			if parent.Type == html.ElementNode {
				if h.tagStrategies[parent.Data] == StrategyPreserveWhitespace {
					return StrategyPreserveWhitespace
				}
			}
		}
		return StrategyTextOnly

	case html.ElementNode:
		if strategy, ok := h.tagStrategies[node.Data]; ok {
			return strategy
		}
		return h.defaultStrategy

	case html.DocumentNode:
		return StrategyUnwrap

	default:
		return StrategyFilter
	}
}

// IsTextOnly returns true if all children of the node can be converted to text
func (h *NodeHandler) IsTextOnly(node *html.Node) bool {
	if node == nil {
		return false
	}

	// Text nodes are always text-only
	if node.Type == html.TextNode {
		return true
	}

	// Check if this node's strategy allows text conversion
	strategy := h.GetStrategy(node)
	if strategy != StrategyTextOnly && strategy != StrategyPreserveWhitespace {
		return false
	}

	// Nodes with class or id attributes that are text-only strategy are not text-only
	for _, attr := range node.Attr {
		if attr.Key == "class" || attr.Key == "id" {
			return false
		}
	}

	// Check all children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !h.IsTextOnly(child) {
			return false
		}
	}

	return true
}

// IsMarkdownable returns true if the node and all its children can be converted to markdown
func (h *NodeHandler) IsMarkdownable(node *html.Node) bool {
	if !h.markdownEnabled {
		return false
	}

	if node == nil {
		return false
	}

	// Text nodes are always markdownable
	if node.Type == html.TextNode {
		return true
	}

	// Check if this node's strategy allows markdown conversion
	strategy := h.GetStrategy(node)
	if strategy != StrategyMarkdown && strategy != StrategyTextOnly && strategy != StrategyPreserveWhitespace {
		return false
	}

	// For non-markdown elements that are text-only, we need to check if they contain any non-markdown elements
	if strategy == StrategyTextOnly {
		// Nodes with class or id attributes that are text-only strategy are not markdownable
		for _, attr := range node.Attr {
			if attr.Key == "class" || attr.Key == "id" {
				return false
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				childStrategy := h.GetStrategy(child)
				if childStrategy != StrategyMarkdown && childStrategy != StrategyTextOnly {
					return false
				}
			}
		}
	}

	// Check all children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !h.IsMarkdownable(child) {
			return false
		}
	}

	return true
}
