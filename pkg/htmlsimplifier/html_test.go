package htmlsimplifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHtmlSimplifier_Structure(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		opts     Options
		expected Document
	}{
		{
			name: "preserve html structure",
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
									{Text: "Test Page"},
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
									{Text: "Hello World"},
								},
							},
						},
					},
				},
			},
		},
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
					{Text: "\n  Line 1\n    Line 2\n      Line 3\n"},
				},
			},
		},
		{
			name: "preserve table structure",
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
									{Tag: "th", Children: []Document{{Text: "Header 1"}}},
									{Tag: "th", Children: []Document{{Text: "Header 2"}}},
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
									{Tag: "td", Children: []Document{{Text: "Cell 1"}}},
									{Tag: "td", Children: []Document{{Text: "Cell 2"}}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "preserve list structure",
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
					{Tag: "li", Children: []Document{{Text: "Item 1"}}},
					{
						Tag: "li",
						Children: []Document{
							{Text: "Item 2"},
							{
								Tag: "ul",
								Children: []Document{
									{Tag: "li", Children: []Document{{Text: "Subitem 1"}}},
									{Tag: "li", Children: []Document{{Text: "Subitem 2"}}},
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
			s := NewSimplifier(tt.opts)
			result, err := s.ProcessHTML(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
