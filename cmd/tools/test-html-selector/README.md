# HTML Selector Testing Tool

A command-line tool for testing CSS and XPath selectors against HTML documents. It provides match counts and contextual examples to verify selector accuracy.

## Features

- Support for both CSS and XPath selectors
- Configurable sample count and context size
- YAML configuration for selectors
- DOM path visualization for matched elements
- Parent context for each match
- Extract and print all matches for each selector
- HTML simplification options for cleaner output

## Installation

```bash
go install ./cmd/tools/test-html-selector
```

## Usage

1. Create a YAML configuration file:

```yaml
selectors:
  - name: product_titles
    selector: .product-card h2
    type: css
  - name: prices
    selector: //div[@class='price']
    type: xpath
config:
  sample_count: 5
  context_chars: 100
```

2. Run the tool:

```bash
# Basic usage
test-html-selector --config config.yaml --input path/to/input.html

# Override sample count and context size
test-html-selector --config config.yaml --input path/to/input.html --sample-count 10 --context-chars 200

# Extract and print all matches
test-html-selector --config config.yaml --input path/to/input.html --extract

# Use HTML simplification options
test-html-selector --config config.yaml --input path/to/input.html \
  --strip-scripts --strip-css --simplify-text --markdown
```

## Configuration Options

### Command Line Flags

#### Basic Options
- `--config`: Path to YAML config file (required)
- `--input`: Path to HTML input file (required)
- `--extract`: Extract and print all matches for each selector
- `--sample-count`: Maximum number of examples to show (default: 5)
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

- `selectors`: List of selectors to test
  - `name`: Friendly name for the selector
  - `selector`: CSS or XPath selector string
  - `type`: Either "css" or "xpath"
- `config`:
  - `sample_count`: Maximum number of examples to show (can be overridden by --sample-count)
  - `context_chars`: Number of characters of context to include (can be overridden by --context-chars)

## Example Output

```yaml
- name: product_titles
  selector: .product-card h2
  count: 3
  samples:
    - html:
        - tag: h2
          text: "Awesome Product 1"
      context:
        - tag: div.info
          children:
            - tag: h2
              text: "Awesome Product 1"
            - tag: div.price
              text: "$19.99"
      path: "html > body > div > div > div > h2"
    # ... more samples ...
```

The output shows the full HTML structure in a simplified YAML format. Both `html` and `context` fields contain arrays of documents, allowing for multiple elements to be represented in their full structure. When using `--markdown` or `--simplify-text`, the output will be converted to the appropriate format while preserving important elements.