# HTML Simplification Tool

A command-line tool to simplify and minimize HTML documents by removing unnecessary elements, shortening content, and providing a clean YAML representation of the document structure.

## Features

- Strip script and style tags
- Remove SVG elements
- Shorten long text content
- Limit list items and table rows
- Filter elements using CSS and XPath selectors
- Simplify text-only nodes
- Compact attribute representation

## Installation

```bash
go install ./cmd/tools/simplify-html
```

## Usage

Basic usage:
```bash
simplify-html input.html > output.yaml
```

With configuration file:
```bash
simplify-html --config filters.yaml input.html > output.yaml
```

## Options

- `--strip-scripts` (default: true): Remove `<script>` tags
- `--strip-css` (default: true): Remove `<style>` tags and style attributes
- `--strip-svg` (default: true): Remove SVG elements
- `--shorten-text` (default: true): Shorten text content longer than 200 characters
- `--simplify-text` (default: true): Collapse nodes with only text/br children into a single text field
- `--compact-svg` (default: true): Simplify SVG elements by removing detailed attributes
- `--max-list-items` (default: 4): Maximum number of items to show in lists and select boxes (0 for unlimited)
- `--max-table-rows` (default: 4): Maximum number of rows to show in tables (0 for unlimited)
- `--config`: Path to YAML configuration file containing selectors to filter out

## Configuration File Format

The configuration file uses YAML format and supports both CSS and XPath selectors:

```yaml
selectors:
  # CSS selectors
  - type: css
    selector: ".advertisement"
  - type: css
    selector: "#sidebar"

  # XPath selectors
  - type: xpath
    selector: "//*[@data-analytics]"
  - type: xpath
    selector: "//div[contains(@class, 'social-media')]"
```

## Output Format

The tool outputs a YAML representation of the HTML document structure:

```yaml
tag: div
attrs: class=content
text: Simple text content  # For text-only nodes
children:                  # For nodes with children
  - tag: p
    text: First paragraph
  - tag: ul
    children:
      - tag: li
        text: List item 1
      - tag: li
        text: List item 2
      - tag: li
        text: ...         # Truncation indicator
```

## Examples

The `examples/` directory contains sample HTML files demonstrating different features:

- `simple.html`: Basic text and inline elements
- `lists.html`: Various types of lists and nesting
- `table.html`: Tables with simple and complex content

Try them out:
```bash
# Basic simplification
simplify-html examples/simple.html

# Limit list items
simplify-html --max-list-items=2 examples/lists.html

# Complex table handling
simplify-html --max-table-rows=3 examples/table.html
```

## Text Simplification

The `--simplify-text` option collapses nodes that contain only text and `<br>` elements into a single text field. This helps reduce the complexity of the output while preserving the content and line breaks.

For example, this HTML:
```html
<div class="content">
    First line<br>
    Second line<br>
    Third line
</div>
```

Becomes:
```yaml
tag: div
attrs: class=content
text: "First line\nSecond line\nThird line"
```

Note: Text simplification is only applied when a node contains exclusively text nodes and `<br>` elements. If a node contains any other elements (like links or formatting), it will preserve the full structure. 