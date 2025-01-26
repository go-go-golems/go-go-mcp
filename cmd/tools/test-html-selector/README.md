# HTML Selector Testing Tool

A command-line tool for testing CSS and XPath selectors against HTML documents. It provides match counts and contextual examples to verify selector accuracy.

## Features

- Support for both CSS and XPath selectors
- Process multiple files and URLs in a single run
- Configurable sample count and context size
- YAML configuration for selectors
- DOM path visualization for matched elements
- Parent context for each match
- Extract and print all matches for each selector
- HTML simplification options for cleaner output
- Template-based output formatting

## Installation

```bash
go install ./cmd/tools/test-html-selector
```

## Usage

1. Create a YAML configuration file:

```yaml
description: |
  Description of what these selectors are trying to match
selectors:
  - name: product_titles
    selector: .product-card h2
    type: css
    description: Extracts product titles from cards
  - name: prices
    selector: //div[@class='price']
    type: xpath
    description: Extracts price elements
config:
  sample_count: 5
  context_chars: 100
  template: |  # Optional Go template for formatting output
    {{- range . }}
    # Results from {{ .Source }}
    {{- range $selector, $matches := .Data }}
    ## {{ $selector }}
    {{- range $matches }}
    - {{ . }}
    {{- end }}
    {{- end }}
    {{- end }}
```

2. Run the tool:

```bash
# Basic usage with config file and multiple sources
test-html-selector --config config.yaml --files file1.html file2.html

# Process multiple URLs
test-html-selector --urls https://example.com https://example.org \
  --select-css ".product-card h2" \
  --select-xpath "//div[@class='price']"

# Extract all matches with template formatting
test-html-selector --config config.yaml \
  --files file1.html file2.html \
  --urls https://example.com \
  --extract --extract-template template.tmpl

# Show context and customize output
test-html-selector --config config.yaml \
  --files input1.html input2.html \
  --show-context --sample-count 10 --context-chars 200
```

## Configuration Options

### Command Line Flags

#### Basic Options
- `--config`: Path to YAML config file
- `--files`: HTML files to process (can be specified multiple times)
- `--urls`: URLs to fetch and process (can be specified multiple times)
- `--select-css`: CSS selectors to test (can be specified multiple times)
- `--select-xpath`: XPath selectors to test (can be specified multiple times)
- `--extract`: Extract all matches into a YAML map of selector name to matches (ignores sample-count limit)
- `--extract-data`: Extract raw data without applying templates
- `--extract-template`: Go template file to render with extracted data
- `--show-context`: Show context around matched elements (default: false)
- `--show-path`: Show path to matched elements (default: true)
- `--sample-count`: Maximum number of examples to show in normal mode (default: 3)
- `--context-chars`: Number of characters of context to include (default: 100)

#### HTML Simplification Options
- `--strip-scripts`: Remove <script> tags (default: true)
- `--strip-css`: Remove <style> tags and style attributes (default: true)
- `--shorten-text`: Shorten <span> and <p> elements longer than 200 characters (default: true)
- `--compact-svg`: Simplify SVG elements in output (default: true)
- `--strip-svg`: Remove all SVG elements (default: true)
- `--simplify-text`: Collapse nodes with only text/br children into a single text field (default: true)
- `--markdown`: Convert text with important elements to markdown format (default: true)
- `--max-list-items`: Maximum number of items to show in lists and select boxes (default: 4, 0 for unlimited)
- `--max-table-rows`: Maximum number of rows to show in tables (default: 4, 0 for unlimited)

### YAML Configuration

```yaml
description: String describing the purpose of these selectors
selectors:
  - name: Friendly name for the selector
    selector: CSS or XPath selector string
    type: "css" or "xpath"
    description: Description of what this selector matches
config:
  sample_count: Maximum number of examples to show
  context_chars: Number of characters of context to include
  template: Optional Go template for formatting extracted data
```

## Example Output

### Default Format (without --extract)
```yaml
- name: product_titles
  selector: .product-card h2
  type: css
  count: 3
  samples:
    - html:
        - tag: h2
          text: "Awesome Product 1"
      context:  # Only shown with --show-context
        - tag: div.info
          children:
            - tag: h2
              text: "Awesome Product 1"
            - tag: div.price
              text: "$19.99"
      path: "html > body > div > div > div > h2"  # Only shown with --show-path
```

### Extract Format (with --extract)
```yaml
- source: file1.html
  data:
    product_titles:
      - "Awesome Product 1"
      - "Awesome Product 2"
    prices:
      - "$19.99"
      - "$29.99"
- source: https://example.com
  data:
    product_titles:
      - "Example Product"
    prices:
      - "$9.99"
```

When using `--extract` with a template, the output format will be determined by your template. The template has access to a list of source results, each containing a map of selector names to their matches. The matches can be text content, markdown, or full document structures depending on your simplification settings.

### Template Example
```go
{{- range . }}
# Results from {{ .Source }}
{{- range $selector, $matches := .Data }}
## {{ $selector }}
{{- range $matches }}
- {{ . }}
{{- end }}
{{- end }}
{{- end }}
```

## Template Functions

The tool includes the full set of [Sprig template functions](http://masterminds.github.io/sprig/) for use in templates, including:

- String manipulation: `trim`, `upper`, `lower`, `replace`
- Math functions: `add`, `sub`, `mul`, `div`
- Date formatting: `now`, `date`, `dateModify`
- And many more...

For more detailed examples and best practices, see the [TUTORIAL.md](TUTORIAL.md) file.