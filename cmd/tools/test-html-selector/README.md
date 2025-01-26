# HTML Selector Testing Tool

A command-line tool for testing CSS and XPath selectors against HTML documents. It provides match counts and contextual examples to verify selector accuracy.

## Features

- Support for both CSS and XPath selectors
- Configurable sample count and context size
- YAML configuration for selectors
- DOM path visualization for matched elements
- Parent context for each match
- Extract and print all matches for each selector

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
```

## Configuration Options

### Command Line Flags

- `--config`: Path to YAML config file (required)
- `--input`: Path to HTML input file (required)
- `--extract`: Extract and print all matches for each selector
- `--sample-count`: Maximum number of examples to show (default: 5)
- `--context-chars`: Number of characters of context to include (default: 100)

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
    - html: "<h2>Awesome Product 1</h2>"
      context: "<div class=\"info\"><h2>Awesome Product 1</h2><div class=\"price\">$19.99</div></div>"
      path: "html > body > div > div > div > h2"
    # ... more samples ...
```