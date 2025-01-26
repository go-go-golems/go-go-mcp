package htmlsimplifier

import (
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

// TextSimplifier handles text-related simplification operations
type TextSimplifier struct {
	markdownEnabled bool
	nodeHandler     *NodeHandler
	mdConverter     *md.Converter
}

// NewTextSimplifier creates a new text simplifier
func NewTextSimplifier(markdownEnabled bool) *TextSimplifier {
	opts := Options{Markdown: markdownEnabled}
	return &TextSimplifier{
		markdownEnabled: markdownEnabled,
		nodeHandler:     NewNodeHandler(opts),
		mdConverter:     md.NewConverter("", true, nil),
	}
}

// SimplifyText attempts to convert a node and its children to a single text string
func (t *TextSimplifier) SimplifyText(node *html.Node) (string, bool) {
	if node == nil {
		log.Trace().Msg("SimplifyText: node is nil")
		return "", false
	}

	// For text nodes, just return the text
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		log.Trace().Str("text", text).Msg("SimplifyText: processing text node")
		return text, true
	}

	// For br nodes, return newline
	if node.Data == "br" {
		log.Trace().Msg("SimplifyText: processing br node")
		return "\n", true
	}

	// If markdown is enabled and this is a markdown-compatible element
	if t.markdownEnabled && (node.Data == "p" || node.Data == "span") {
		log.Trace().Str("node_type", node.Data).Msg("SimplifyText: attempting markdown conversion")
		text, ok := t.ConvertToMarkdown(node)
		if ok {
			log.Trace().Str("text", text).Msg("SimplifyText: markdown conversion successful")
			return strings.TrimSpace(text), true
		}
		log.Trace().Msg("SimplifyText: markdown conversion failed")
	}

	// Special case for root node (html/body) or text-only nodes
	if node.Type == html.DocumentNode || node.Data == "html" || node.Data == "body" || t.nodeHandler.IsTextOnly(node) {
		// For element nodes, try to combine all child text
		var parts []string
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			text, ok := t.SimplifyText(child)
			if ok && text != "" {
				parts = append(parts, text)
			} else if !ok && !t.nodeHandler.IsTextOnly(node) {
				log.Trace().Str("node_type", node.Data).Msg("SimplifyText: failed to process child node")
				return "", false
			}
		}

		result := strings.Join(parts, "")
		log.Trace().Str("node_type", node.Data).Str("result", result).Msg("SimplifyText: processed element node")
		return result, true
	}

	log.Trace().Str("node_type", node.Data).Msg("SimplifyText: node cannot be converted to text")
	return "", false
}

// ExtractText extracts text from a node and its children, preserving whitespace if needed
func (t *TextSimplifier) ExtractText(node *html.Node) string {
	if node == nil {
		return ""
	}

	// For text nodes, return the text as is
	if node.Type == html.TextNode {
		strategy := t.nodeHandler.GetStrategy(node)
		if strategy == StrategyPreserveWhitespace {
			return node.Data
		}
		return strings.TrimSpace(node.Data)
	}

	// For element nodes, combine all child text
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text := t.ExtractText(child)
		if text != "" {
			parts = append(parts, text)
		}
	}

	// Add appropriate spacing based on the node type and strategy
	strategy := t.nodeHandler.GetStrategy(node)
	switch {
	case strategy == StrategyPreserveWhitespace:
		return strings.Join(parts, "")
	case node.Data == "br":
		return "\n"
	case node.Data == "p", node.Data == "div":
		return strings.Join(parts, "\n")
	case node.Data == "li":
		return "- " + strings.Join(parts, " ")
	default:
		return strings.Join(parts, " ")
	}
}

// ConvertToMarkdown converts a node and its children to markdown format
func (t *TextSimplifier) ConvertToMarkdown(node *html.Node) (string, bool) {
	if node == nil {
		log.Trace().Msg("ConvertToMarkdown: node is nil")
		return "", false
	}

	// For text nodes, return the text as is
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		log.Trace().Str("text", text).Msg("ConvertToMarkdown: processing text node")
		return text, true
	}

	// Convert the node to HTML string
	var buf strings.Builder
	err := html.Render(&buf, node)
	if err != nil {
		log.Error().Err(err).Msg("ConvertToMarkdown: failed to render HTML")
		return "", false
	}

	// Convert to markdown using html-to-markdown
	markdown, err := t.mdConverter.ConvertString(buf.String())
	if err != nil {
		log.Error().Err(err).Msg("ConvertToMarkdown: failed to convert to markdown")
		return "", false
	}

	if markdown == "" {
		log.Trace().Msg("ConvertToMarkdown: empty result")
		return "", false
	}

	// replace ' \n ' with '\n'
	markdown = strings.ReplaceAll(markdown, " \n ", "\n")

	log.Trace().Str("result", markdown).Msg("ConvertToMarkdown: final result")
	return markdown, true
}
