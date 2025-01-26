package htmlsimplifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name     string
	html     string
	opts     Options
	expected Document
}

func runTestCase(t *testing.T, tc testCase) {
	t.Helper()
	s := NewSimplifier(tc.opts)
	result, err := s.ProcessHTML(tc.html)
	require.NoError(t, err)
	assert.Equal(t, tc.expected, result)
}

func TestSimpleElements(t *testing.T) {
	tests := []testCase{
		{
			name: "single text node",
			html: "Hello World",
			opts: Options{},
			expected: Document{
				Tag:  "#text",
				Text: "Hello World",
			},
		},
		{
			name: "single element with text",
			html: "<p>Hello World</p>",
			opts: Options{},
			expected: Document{
				Tag: "p",
				Children: []Document{
					{Tag: "#text", Text: "Hello World"},
				},
			},
		},
		{
			name: "element with attributes",
			html: `<p class="test" id="p1">Hello World</p>`,
			opts: Options{},
			expected: Document{
				Tag:   "p",
				Attrs: `class=test id=p1`,
				Children: []Document{
					{Tag: "#text", Text: "Hello World"},
				},
			},
		},
		{
			name: "very simple title single element",
			html: "<title>Test Page</title>",
			opts: Options{},
			expected: Document{
				Tag: "title",
				Children: []Document{
					{Tag: "#text", Text: "Test Page"},
				},
			},
		},
		{
			name: "head title",
			html: "<head><title>Test Page</title></head>",
			opts: Options{},
			expected: Document{
				Tag: "title",
				Children: []Document{
					{Tag: "#text", Text: "Test Page"},
				},
			},
		},
		{
			name: "html head title",
			html: "<html><head><title>Test Page</title></head></html>",
			opts: Options{},
			expected: Document{
				Tag: "title",
				Children: []Document{
					{Tag: "#text", Text: "Test Page"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCase(t, tt)
		})
	}
}

func TestPreserveWhitespace(t *testing.T) {
	tests := []testCase{
		{
			name: "preserve whitespace in pre",
			html: `<pre>
  Line 1
    Line 2
      Line 3
</pre>`,
			opts: Options{},
			expected: Document{
				Tag: "pre",
				Children: []Document{
					{Tag: "#text", Text: "\n  Line 1\n    Line 2\n      Line 3\n"},
				},
			},
		},
		{
			name: "preserve whitespace in code",
			html: `<code>
  func main() {
    fmt.Println("Hello")
  }
</code>`,
			opts: Options{},
			expected: Document{
				Tag: "code",
				Children: []Document{
					{Tag: "#text", Text: "\n  func main() {\n    fmt.Println(\"Hello\")\n  }\n"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCase(t, tt)
		})
	}
}

func TestTableStructure(t *testing.T) {
	tests := []testCase{
		{
			name: "simple table",
			html: "<table><tr><td>Cell 1</td></tr></table>",
			opts: Options{},
			expected: Document{
				Tag: "table",
				Children: []Document{
					{
						Tag: "tr",
						Children: []Document{
							{
								Tag: "td",
								Children: []Document{
									{Tag: "#text", Text: "Cell 1"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "table with header and body",
			html: `<table>
  <thead>
    <tr><th>Header 1</th><th>Header 2</th></tr>
  </thead>
  <tbody>
    <tr><td>Cell 1</td><td>Cell 2</td></tr>
  </tbody>
</table>`,
			opts: Options{},
			expected: Document{
				Tag: "table",
				Children: []Document{
					{
						Tag: "thead",
						Children: []Document{
							{
								Tag: "tr",
								Children: []Document{
									{Tag: "th", Children: []Document{{Tag: "#text", Text: "Header 1"}}},
									{Tag: "th", Children: []Document{{Tag: "#text", Text: "Header 2"}}},
								},
							},
						},
					},
					{
						Tag: "tbody",
						Children: []Document{
							{
								Tag: "tr",
								Children: []Document{
									{Tag: "td", Children: []Document{{Tag: "#text", Text: "Cell 1"}}},
									{Tag: "td", Children: []Document{{Tag: "#text", Text: "Cell 2"}}},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCase(t, tt)
		})
	}
}

func TestListStructure(t *testing.T) {
	tests := []testCase{
		{
			name: "simple unordered list",
			html: "<ul><li>Item 1</li><li>Item 2</li></ul>",
			opts: Options{},
			expected: Document{
				Tag: "ul",
				Children: []Document{
					{Tag: "li", Children: []Document{{Tag: "#text", Text: "Item 1"}}},
					{Tag: "li", Children: []Document{{Tag: "#text", Text: "Item 2"}}},
				},
			},
		},
		{
			name: "nested list",
			html: `<ul>
  <li>Item 1</li>
  <li>Item 2
    <ul>
      <li>Subitem 1</li>
      <li>Subitem 2</li>
    </ul>
  </li>
</ul>`,
			opts: Options{},
			expected: Document{
				Tag: "ul",
				Children: []Document{
					{Tag: "li", Children: []Document{{Tag: "#text", Text: "Item 1"}}},
					{
						Tag: "li",
						Children: []Document{
							{Tag: "#text", Text: "Item 2"},
							{
								Tag: "ul",
								Children: []Document{
									{Tag: "li", Children: []Document{{Tag: "#text", Text: "Subitem 1"}}},
									{Tag: "li", Children: []Document{{Tag: "#text", Text: "Subitem 2"}}},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCase(t, tt)
		})
	}
}

func TestCompleteDocumentStructure(t *testing.T) {
	tests := []testCase{
		{
			name: "complete html document",
			html: `<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <p>Hello World</p>
</body>
</html>`,
			opts: Options{},
			expected: Document{
				Tag: "html",
				Children: []Document{
					{
						Tag: "head",
						Children: []Document{
							{
								Tag: "title",
								Children: []Document{
									{Tag: "#text", Text: "Test Page"},
								},
							},
						},
					},
					{
						Tag: "body",
						Children: []Document{
							{
								Tag: "p",
								Children: []Document{
									{Tag: "#text", Text: "Hello World"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCase(t, tt)
		})
	}
}
