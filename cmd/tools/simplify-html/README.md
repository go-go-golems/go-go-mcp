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

The configuration file uses YAML format and supports both CSS and XPath selectors with two modes:
- `select`: Keep only these elements and their parents
- `filter`: Remove these elements from the document

The selectors are applied in order:
1. First, all `select` selectors are applied - only elements matching these selectors (and their parents) are kept
2. Then, all `filter` selectors are applied to remove unwanted elements from the selected ones
3. If no `select` selectors are provided, the entire document is processed with just the `filter` selectors

```yaml
selectors:
  # Select only these elements and their parents
  - type: css
    mode: select
    selector: "main, article"
  - type: css
    mode: select
    selector: "h1, h2, h3"

  # Then filter out these elements from the selected ones
  - type: css
    mode: filter
    selector: ".advertisement"
  - type: xpath
    mode: filter
    selector: "//*[@data-analytics]"
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