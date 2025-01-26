package htmlsimplifier

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

// TextSimplifier handles text-related simplification operations
type TextSimplifier struct {
	markdownEnabled bool
	nodeHandler     *NodeHandler
}

// NewTextSimplifier creates a new text simplifier
func NewTextSimplifier(markdownEnabled bool) *TextSimplifier {
	opts := Options{Markdown: markdownEnabled}
	return &TextSimplifier{
		markdownEnabled: markdownEnabled,
		nodeHandler:     NewNodeHandler(opts),
	}
}

// MarkdownElements defines HTML elements that can be converted to markdown
var MarkdownElements = map[string]bool{
	"a":      true, // Links (only within p or span)
	"strong": true, // Bold text
	"em":     true, // Italic text
	"b":      true, // Bold text (alternative)
	"i":      true, // Italic text (alternative)
	"code":   true, // Code snippets
}

// MarkdownStart defines the opening markdown syntax for each element type
var MarkdownStart = map[string]string{
	"a":      "[",  // Links
	"strong": "**", // Bold text
	"em":     "*",  // Italic text
	"b":      "**", // Bold text (alternative)
	"i":      "*",  // Italic text (alternative)
	"code":   "`",  // Code snippets
}

// MarkdownEnd defines the closing markdown syntax for each element type
var MarkdownEnd = map[string]string{
	"a":      "](%s)", // Links (format with href)
	"strong": "**",    // Bold text
	"em":     "*",     // Italic text
	"b":      "**",    // Bold text (alternative)
	"i":      "*",     // Italic text (alternative)
	"code":   "`",     // Code snippets
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

	// Check if markdown is enabled for this node
	if !t.markdownEnabled && MarkdownElements[node.Data] {
		log.Trace().Str("node_type", node.Data).Msg("ConvertToMarkdown: markdown disabled for this element")
		return "", false
	}

	log.Trace().Str("node_type", node.Data).Msg("ConvertToMarkdown: processing element node")

	// Process children first
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			text := strings.TrimSpace(child.Data)
			if text != "" {
				parts = append(parts, text)
				log.Trace().Str("text", text).Msg("ConvertToMarkdown: added text node content")
			}
			continue
		}

		switch child.Data {
		case "a":
			href := ""
			for _, attr := range child.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}
			text, ok := t.ConvertToMarkdown(child)
			if !ok || text == "" {
				log.Trace().Msg("ConvertToMarkdown: failed to process link content")
				return "", false
			}
			link := fmt.Sprintf("[%s](%s)", text, href)
			parts = append(parts, link)
			log.Trace().Str("link", link).Msg("ConvertToMarkdown: processed link")
		case "strong", "b":
			text, ok := t.ConvertToMarkdown(child)
			if !ok || text == "" {
				log.Trace().Msg("ConvertToMarkdown: failed to process strong/bold content")
				return "", false
			}
			bold := fmt.Sprintf("**%s**", text)
			parts = append(parts, bold)
			log.Trace().Str("bold", bold).Msg("ConvertToMarkdown: processed strong/bold")
		case "em", "i":
			text, ok := t.ConvertToMarkdown(child)
			if !ok || text == "" {
				log.Trace().Msg("ConvertToMarkdown: failed to process emphasis content")
				return "", false
			}
			em := fmt.Sprintf("*%s*", text)
			parts = append(parts, em)
			log.Trace().Str("emphasis", em).Msg("ConvertToMarkdown: processed emphasis")
		case "code":
			text, ok := t.ConvertToMarkdown(child)
			if !ok || text == "" {
				log.Trace().Msg("ConvertToMarkdown: failed to process code content")
				return "", false
			}
			code := fmt.Sprintf("`%s`", text)
			parts = append(parts, code)
			log.Trace().Str("code", code).Msg("ConvertToMarkdown: processed code")
		case "br":
			parts = append(parts, "\n")
			log.Trace().Msg("ConvertToMarkdown: processed line break")
		default:
			text, ok := t.ConvertToMarkdown(child)
			if !ok {
				log.Trace().Str("node_type", child.Data).Msg("ConvertToMarkdown: failed to process unknown element")
				return "", false
			}
			if text != "" {
				parts = append(parts, text)
				log.Trace().Str("text", text).Msg("ConvertToMarkdown: processed unknown element")
			}
		}
	}

	result := strings.Join(parts, " ")
	if result == "" {
		log.Trace().Msg("ConvertToMarkdown: empty result")
		return "", false
	}

	// replace ' \n ' with '\n'
	result = strings.ReplaceAll(result, " \n ", "\n")

	log.Trace().Str("result", result).Msg("ConvertToMarkdown: final result")
	return result, true
}
