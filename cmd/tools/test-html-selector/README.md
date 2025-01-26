# HTML Selector Testing Tool

A command-line tool for testing CSS and XPath selectors against HTML documents. It provides match counts and contextual examples to verify selector accuracy.

## Features

- Support for both CSS and XPath selectors
- Configurable sample count and context size
- YAML configuration and output format
- DOM path visualization for matched elements
- Parent context for each match
- Command-line override for input file
- Extract and print all matches for each selector

## Installation

```bash
go install ./cmd/tools/test-html-selector
```

## Usage

1. Create a YAML configuration file:

```yaml
file: path/to/your.html  # Optional if using --input flag
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
# Using config file only
test-html-selector -c config.yaml

# Override input file from command line
test-html-selector -c config.yaml -i path/to/different.html

# Extract and print all matches
test-html-selector -c config.yaml -e
```

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

## Configuration Options

- `file`: Path to the HTML file to analyze (can be overridden by --input flag)
- `selectors`: List of selectors to test
  - `name`: Friendly name for the selector
  - `selector`: CSS or XPath selector string
  - `type`: Either "css" or "xpath"
- `config`:
  - `sample_count`: Maximum number of examples to show
  - `context_chars`: Number of characters of context to include

## Command Line Flags

- `-c, --config`: Path to YAML config file (required)
- `-i, --input`: Path to HTML input file (overrides config file)
- `-e, --extract`: Extract and print all matches for each selector