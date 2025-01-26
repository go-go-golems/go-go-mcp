package htmlsimplifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestTextSimplifier_SimplifyText(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		markdown     bool
		expectedText string
		canSimplify  bool
	}{
		// {
		// 	name:         "simple text node",
		// 	html:         "Hello World",
		// 	markdown:     false,
		// 	expectedText: "Hello World",
		// 	canSimplify:  true,
		// },
		// {
		// 	name:         "paragraph with text",
		// 	html:         "<p>Hello World</p>",
		// 	markdown:     false,
		// 	expectedText: "Hello World",
		// 	canSimplify:  true,
		// },
		{
			name:         "text with line break",
			html:         "<p>Hello<br>World</p>",
			markdown:     false,
			expectedText: "Hello\nWorld",
			canSimplify:  true,
		},
		{
			name:         "link in div without markdown",
			html:         `<div><a href="https://example.com">Click here</a></div>`,
			markdown:     false,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "link in div with markdown",
			html:         `<div><a href="https://example.com">Click here</a></div>`,
			markdown:     true,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "link in paragraph with markdown",
			html:         `<p><a href="https://example.com">Click here</a></p>`,
			markdown:     true,
			expectedText: "[Click here](https://example.com)",
			canSimplify:  true,
		},
		{
			name:         "link in paragraph without markdown",
			html:         `<p><a href="https://example.com">Click here</a></p>`,
			markdown:     false,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "bold text in div without markdown",
			html:         "<div><strong>Important</strong></div>",
			markdown:     false,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "bold text in paragraph with markdown",
			html:         "<p><strong>Important</strong></p>",
			markdown:     true,
			expectedText: "**Important**",
			canSimplify:  true,
		},
		{
			name:         "mixed formatting in paragraph with markdown",
			html:         `<p><strong>Note:</strong> Please read our <a href="https://example.com">terms</a></p>`,
			markdown:     true,
			expectedText: "**Note:** Please read our [terms](https://example.com)",
			canSimplify:  true,
		},
		{
			name:         "mixed formatting in div without markdown",
			html:         `<div><strong>Note:</strong> Please read our <a href="https://example.com">terms</a></div>`,
			markdown:     false,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "footer with links should not be simplified",
			html:         `<footer><a href="https://example.com">Terms</a> | <a href="https://example.com">Privacy</a></footer>`,
			markdown:     true,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "navigation with links should not be simplified",
			html:         `<nav><ul><li><a href="https://example.com">Home</a></li></ul></nav>`,
			markdown:     true,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "span with link with markdown",
			html:         `<span><a href="https://example.com">Click here</a></span>`,
			markdown:     true,
			expectedText: "[Click here](https://example.com)",
			canSimplify:  true,
		},
		{
			name:         "multiple paragraphs with markdown elements",
			html:         "<div><p>First <strong>paragraph</strong></p><p>Second <em>paragraph</em></p></div>",
			markdown:     true,
			expectedText: "",
			canSimplify:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			// Find the first non-html/head/body element
			var testNode *html.Node
			if doc.FirstChild != nil && doc.FirstChild.FirstChild != nil {
				testNode = doc.FirstChild.FirstChild.NextSibling // body node
				if testNode.FirstChild != nil {
					testNode = testNode.FirstChild
				}
			}
			assert.NotNil(t, testNode, "Failed to find test node")

			ts := NewTextSimplifier(tt.markdown)
			text, canSimplify := ts.SimplifyText(testNode)
			assert.Equal(t, tt.expectedText, text)
			assert.Equal(t, tt.canSimplify, canSimplify)
		})
	}
}

func TestTextSimplifier_ProcessMarkdownElement(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		elementTag   string
		markdown     bool
		expectedText string
		canSimplify  bool
	}{
		{
			name:         "link in paragraph",
			html:         `<p><a href="https://example.com">Click here</a></p>`,
			elementTag:   "p",
			markdown:     true,
			expectedText: "[Click here](https://example.com)",
			canSimplify:  true,
		},
		{
			name:         "strong in paragraph",
			html:         "<p><strong>Important text</strong></p>",
			elementTag:   "p",
			markdown:     true,
			expectedText: "**Important text**",
			canSimplify:  true,
		},
		{
			name:         "emphasis in span",
			html:         "<span><em>Emphasized text</em></span>",
			elementTag:   "span",
			markdown:     true,
			expectedText: "*Emphasized text*",
			canSimplify:  true,
		},
		{
			name:         "code in paragraph",
			html:         "<p><code>print('hello')</code></p>",
			elementTag:   "p",
			markdown:     true,
			expectedText: "`print('hello')`",
			canSimplify:  true,
		},
		{
			name:         "markdown element without markdown enabled",
			html:         "<p><strong>Important text</strong></p>",
			elementTag:   "p",
			markdown:     false,
			expectedText: "",
			canSimplify:  false,
		},
		{
			name:         "empty markdown element",
			html:         "<p><strong></strong></p>",
			elementTag:   "p",
			markdown:     true,
			expectedText: "",
			canSimplify:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			element := findFirstElement(doc, tt.elementTag)
			assert.NotNil(t, element, "Failed to find element")

			ts := NewTextSimplifier(tt.markdown)
			text, ok := ts.ConvertToMarkdown(element)
			if tt.canSimplify {
				assert.True(t, ok)
			} else {
				assert.False(t, ok)
			}
			assert.Equal(t, tt.expectedText, text)
		})
	}
}
