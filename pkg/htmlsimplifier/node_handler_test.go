package htmlsimplifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeHandler_GetStrategy(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		tag      string
		opts     Options
		expected NodeHandlingStrategy
	}{
		{
			name:     "unwrap html tag",
			html:     "<html><body>test</body></html>",
			tag:      "html",
			opts:     Options{},
			expected: StrategyUnwrap,
		},
		{
			name:     "filter script when enabled",
			html:     "<script>alert('test')</script>",
			tag:      "script",
			opts:     Options{StripScripts: true},
			expected: StrategyFilter,
		},
		{
			name:     "keep script when disabled",
			html:     "<script>alert('test')</script>",
			tag:      "script",
			opts:     Options{StripScripts: false},
			expected: StrategyDefault,
		},
		{
			name:     "preserve whitespace in pre",
			html:     "<pre>  test  </pre>",
			tag:      "pre",
			opts:     Options{},
			expected: StrategyPreserveWhitespace,
		},
		{
			name:     "text only for p",
			html:     "<p>test</p>",
			tag:      "p",
			opts:     Options{},
			expected: StrategyTextOnly,
		},
		{
			name:     "default for div",
			html:     "<div>test</div>",
			tag:      "div",
			opts:     Options{},
			expected: StrategyDefault,
		},
		{
			name:     "markdown for strong when enabled",
			html:     "<strong>test</strong>",
			tag:      "strong",
			opts:     Options{Markdown: true},
			expected: StrategyMarkdown,
		},
		{
			name:     "default for strong when markdown disabled",
			html:     "<strong>test</strong>",
			tag:      "strong",
			opts:     Options{Markdown: false},
			expected: StrategyDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			node := findFirstElement(doc, tt.tag)
			assert.NotNil(t, node, "test node not found")

			handler := NewNodeHandler(tt.opts)
			strategy := handler.GetStrategy(node)
			assert.Equal(t, tt.expected, strategy)
		})
	}
}

func TestNodeHandler_IsTextOnly(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		tag      string
		opts     Options
		expected bool
	}{
		{
			name:     "simple paragraph",
			html:     "<p>test</p>",
			tag:      "p",
			opts:     Options{},
			expected: true,
		},
		{
			name:     "paragraph with span",
			html:     "<p>test <span>more</span></p>",
			tag:      "p",
			opts:     Options{},
			expected: true,
		},
		{
			name:     "paragraph with link",
			html:     "<p>test <a href='#'>link</a></p>",
			tag:      "p",
			opts:     Options{},
			expected: false,
		},
		{
			name:     "pre with whitespace",
			html:     "<pre>  test  \n  more  </pre>",
			tag:      "pre",
			opts:     Options{},
			expected: true,
		},
		{
			name:     "div with mixed content",
			html:     "<div>test <strong>bold</strong></div>",
			tag:      "div",
			opts:     Options{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			node := findFirstElement(doc, tt.tag)
			assert.NotNil(t, node, "test node not found")

			handler := NewNodeHandler(tt.opts)
			result := handler.IsTextOnly(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNodeHandler_IsMarkdownable(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		tag      string
		opts     Options
		expected bool
	}{
		{
			name:     "simple text",
			html:     "<p>test</p>",
			tag:      "p",
			opts:     Options{Markdown: true},
			expected: true,
		},
		{
			name:     "markdown disabled",
			html:     "<p>test</p>",
			tag:      "p",
			opts:     Options{Markdown: false},
			expected: false,
		},
		{
			name:     "text with markdown elements",
			html:     "<p>test <strong>bold</strong> and <em>italic</em></p>",
			tag:      "p",
			opts:     Options{Markdown: true},
			expected: true,
		},
		{
			name:     "text with non-markdown elements",
			html:     "<p>test <div>block</div></p>",
			tag:      "body", // body because the div is a sibling of the p, when parsed
			opts:     Options{Markdown: true},
			expected: false,
		},
		{
			name:     "nested markdown elements",
			html:     "<strong>bold <em>and italic</em></strong>",
			tag:      "strong",
			opts:     Options{Markdown: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			node := findFirstElement(doc, tt.tag)
			assert.NotNil(t, node, "test node not found")

			handler := NewNodeHandler(tt.opts)
			result := handler.IsMarkdownable(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}
