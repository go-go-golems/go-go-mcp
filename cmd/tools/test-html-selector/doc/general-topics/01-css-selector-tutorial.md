---
Title: Extracting information out of HTML files/urls with selectors
Slug: css-selector-tutorial
Short: Learn how to create selector configurations for extracting data from HTML documents using CSS and XPath selectors.
Topics:
  - html
  - selectors
  - css
  - xpath
Commands:
  - test-html-selector
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# HTML Selector Configuration Tutorial

This tutorial will guide you through creating selector configurations for extracting data from HTML documents.
Each example demonstrates different aspects of the configuration format and selector strategies.

## Running the Tool

You can run the tool in several ways:

```bash
# Using a config file with multiple files
html-selector select --config examples/tutorial/01-basic-text.yaml \
  --files examples/tutorial/01-basic-text.html examples/tutorial/02-nested-content.html

# Testing individual selectors with URLs
html-selector select --urls https://example.com https://example.org \
  --select-css "h1" \
  --select-css "p.intro"

# Showing context around matches
html-selector select --config examples/tutorial/01-basic-text.yaml \
  --files examples/tutorial/01-basic-text.html \
  --show-context

# Extracting all matches with a template (ignores sample-count limit)
html-selector select --config examples/tutorial/01-basic-text.yaml \
  --files examples/tutorial/01-basic-text.html \
  --extract --extract-template template.tmpl
```

For more options and details, see the README.md file.

Note: When using `--extract` mode, ALL matches will be included in the output, regardless of the sample-count setting. This is useful when you need to process all matching elements, not just a sample.

## Output Format

The tool supports two main output formats:

### Default Format (without --extract)
```yaml
- name: selector_name
  selector: "h1"
  type: css
  count: 5
  samples:
    - html: [...]
      context: [...]  # Only if --show-context is true
      path: "file.html"  # Only if --show-path is true
```

### Extract Format (with --extract)
```yaml
- source: file1.html  # or URL
  data:
    selector1:
      - "Match 1"
      - "Match 2"
    selector2:
      - "Match 1"
      - "Match 2"
- source: file2.html
  data:
    selector1:
      - "Match 1"
      - "Match 2"
```

### Template Usage

When using templates with --extract-template, you have access to the source information:

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

## Basic Structure

A selector configuration consists of:
- A description of the overall purpose
- A list of selectors with their types (CSS or XPath) and descriptions
- Configuration options for output formatting

## Examples

Each example below demonstrates a specific use case or technique. The examples are available in the `examples/tutorial/` directory.

### 1. Basic Text Extraction

[01-basic-text.html](examples/tutorial/01-basic-text.html) shows how to extract simple text content:
```html
<div class="content">
  <h1>Welcome to Our Site</h1>
  <p class="intro">This is an introduction paragraph.</p>
  <p class="detail">Here are some details.</p>
</div>
```

[01-basic-text.yaml](examples/tutorial/01-basic-text.yaml) demonstrates basic CSS selectors:
```yaml
description: |
  Basic example showing how to extract text content using CSS selectors.
  Demonstrates selecting headings and paragraphs with different classes.

selectors:
  - name: page_title
    selector: h1
    type: css
    description: |
      Extracts the main page title.
      Uses a simple element selector to find the h1 heading.

  - name: intro_text
    selector: p.intro
    type: css
    description: |
      Extracts the introduction paragraph.
      Uses a class selector to find paragraphs with class "intro".

# Template iterates over a list of documents, where each document contains
# the source (file/URL) and its extracted data
template: |
  # Content from {{ .Source }}
  
  ## Page Title
  {{ index .Data.page_title 0 }}
  
  ## Introduction
  {{ index .Data.intro_text 0 }}

config:
  sample_count: 5
  context_chars: 100
```

### 2. Nested Content

[02-nested-content.html](examples/tutorial/02-nested-content.html) shows how to handle nested structures:
```html
<div class="product-list">
  <div class="product">
    <h2>Product Name</h2>
    <div class="details">
      <span class="price">$19.99</span>
      <span class="stock">In Stock</span>
    </div>
  </div>
  <!-- More products... -->
</div>
```

[02-nested-content.yaml](examples/tutorial/02-nested-content.yaml) demonstrates descendant selectors:
```yaml
description: |
  Example showing how to extract data from nested structures.
  Demonstrates using descendant selectors and multiple selectors per element.

selectors:
  - name: products
    selector: div.product
    type: css
    description: |
      Extracts complete product blocks.
      Useful when you need the entire product context.

  - name: product_names
    selector: div.product h2
    type: css
    description: |
      Extracts product names using a descendant selector.
      The space in "div.product h2" means "find h2 inside div.product".

  - name: prices
    selector: .product .price
    type: css
    description: |
      Extracts prices from product blocks.
      Shows how to target deeply nested elements.

# Template iterates over a list of documents, where each document contains
# the source (file/URL) and its extracted data
template: |
  # Products from {{ .Source }}
  
  {{ $ := . }}
  {{- range $index, $name := .Data.product_names }}
  ## Product {{ add $index 1 }}
  - Name: {{ $name }}
  - Price: {{ index $.Data.prices $index }}
  {{- end }}

config:
  sample_count: 5
  context_chars: 100
```

### 3. Tables and Lists

[03-tables-lists.html](examples/tutorial/03-tables-lists.html) demonstrates handling structured data:
```html
<table class="data-table">
  <tr>
    <th>Name</th>
    <th>Age</th>
  </tr>
  <tr>
    <td>John</td>
    <td>25</td>
  </tr>
</table>

<ul class="features">
  <li>Feature 1</li>
  <li>Feature 2</li>
</ul>
```

[03-tables-lists.yaml](examples/tutorial/03-tables-lists.yaml) shows how to extract structured data:
```yaml
description: |
  Example showing how to extract data from tables and lists.
  Demonstrates different strategies for handling structured data.

selectors:
  - name: table_rows
    selector: .data-table tr
    type: css
    description: |
      Extracts complete table rows.
      Includes both header and data rows.

  - name: table_cells
    selector: .data-table td
    type: css
    description: |
      Extracts just the data cells.
      Excludes header cells (th elements).

  - name: list_items
    selector: .features li
    type: css
    description: |
      Extracts items from the features list.
      Simple example of list extraction.

# Template iterates over a list of documents, where each document contains
# the source (file/URL) and its extracted data
template: |
  # Data from {{ .Source }}
  
  ## Table Data
  | Row | Content |
  |-----|---------|
  {{- range .Data.table_cells }}
  | {{ . }} |
  {{- end }}
  
  ## Features
  {{- range .Data.list_items }}
  - {{ . }}
  {{- end }}

config:
  sample_count: 10
  context_chars: 100
```

### 4. XPath Examples

[04-xpath.html](examples/tutorial/04-xpath.html) shows cases where XPath is useful:
```html
<div class="article">
  <p>First paragraph</p>
  <p>Second paragraph</p>
  <p>Third paragraph</p>
  <div class="comments">
    <div class="comment">Comment 1</div>
    <div class="comment">Comment 2</div>
  </div>
</div>
```

[04-xpath.yaml](examples/tutorial/04-xpath.yaml) demonstrates XPath selectors:
```yaml
description: |
  Example showing how to use XPath selectors for complex queries.
  Demonstrates XPath's power for specific selections.

selectors:
  - name: second_paragraph
    selector: //div[@class='article']/p[2]
    type: xpath
    description: |
      Extracts the second paragraph specifically.
      Shows how to use XPath position predicates.

  - name: last_comment
    selector: //div[@class='comments']/div[last()]
    type: xpath
    description: |
      Extracts the last comment using XPath's last() function.
      Demonstrates dynamic position selection.

  - name: paragraphs_before_comments
    selector: //div[@class='comments']/preceding-sibling::p
    type: xpath
    description: |
      Extracts all paragraphs that come before the comments section.
      Shows XPath's powerful axis navigation.

# Template iterates over a list of documents, where each document contains
# the source (file/URL) and its extracted data
template: |
  # Content Analysis from {{ .Source }}
  
  ## Second Paragraph
  {{ index .Data.second_paragraph 0 }}
  
  ## Last Comment
  {{ index .Data.last_comment 0 }}
  
  ## Paragraphs Before Comments
  {{- range .Data.paragraphs_before_comments }}
  - {{ . }}
  {{- end }}

config:
  sample_count: 5
  context_chars: 100
```

### 5. Template Output

[05-template.html](examples/tutorial/05-template.html) shows data suitable for template formatting:
```html
<div class="user-profile">
  <h1 class="name">John Doe</h1>
  <div class="details">
    <span class="email">john@example.com</span>
    <span class="location">New York</span>
  </div>
  <ul class="skills">
    <li>Python</li>
    <li>JavaScript</li>
    <li>Go</li>
  </ul>
</div>
```

[05-template.yaml](examples/tutorial/05-template.yaml) demonstrates template usage:
```yaml
description: |
  Example showing how to use templates to format extracted data.
  Demonstrates combining multiple selectors into a formatted output.

selectors:
  - name: user_name
    selector: .name
    type: css
    description: |
      Extracts the user's name.

  - name: user_email
    selector: .email
    type: css
    description: |
      Extracts the user's email.

  - name: user_location
    selector: .location
    type: css
    description: |
      Extracts the user's location.

  - name: user_skills
    selector: .skills li
    type: css
    description: |
      Extracts the user's skills.

template: |
  # Profile from {{ .Source }}

  **Name**: {{ index .Data.user_name 0 }}
  **Email**: {{ index .Data.user_email 0 }}
  **Location**: {{ index .Data.user_location 0 }}

  ## Skills
  {{- range .Data.user_skills }}
  - {{ . }}
  {{- end }}

config:
  sample_count: 5
  context_chars: 100
```

The data structure passed to the template engine is a list of source results, where each source result has this structure:

```yaml
- source: "file.html"  # or URL
  data:
    selector_name:  # matches the name in your selector config
      - "First match as markdown"
      - "Second match as markdown"
      - "..."
    another_selector:
      - "First match"
      - "Second match"
- source: "another-file.html"
  data:
    selector_name:
      - "Matches from second file"
      - "..."
```

You can access this data in your templates using:
- `.Source` - the source file/URL
- `.Data.$selector_name` - list of matches for a given selector
- `index .Data.$selector_name 0` - first match for a selector
- `range .Data.$selector_name` - iterate over all matches

The template has access to all [Sprig template functions](http://masterminds.github.io/sprig/) for string manipulation, date formatting, etc.

## Best Practices

1. **Clear Descriptions**
   - Always provide clear descriptions for both the overall configuration and individual selectors
   - Explain what each selector targets and why it's useful

2. **Selector Naming**
   - Use descriptive names that indicate the content being extracted
   - Use consistent naming conventions (e.g., snake_case)
   - Group related selectors with common prefixes

3. **Selector Types**
   - Use CSS selectors for simple queries and class/id based selection
   - Use XPath for complex queries involving position or relationships
   - Choose the simplest selector that gets the job done

4. **Testing**
   - Start with small samples to verify selector accuracy
   - Use the --show-context flag to see surrounding content
   - Use the --show-path flag to understand element hierarchy

5. **Templates**
   - Use templates for formatting when the default YAML output isn't suitable
   - Take advantage of Sprig functions for data manipulation
   - Consider creating reusable template snippets
   - Make sure you are iterating over {.Source, .Data}[] as you could have multiple sources

## Common Patterns

1. **Extracting Lists**
   ```yaml
   - name: items
     selector: ul.list li
     type: css
   ```

2. **Finding Specific Elements**
   ```yaml
   - name: second_item
     selector: //ul[@class='list']/li[2]
     type: xpath
   ```

3. **Combining Related Data**
   ```yaml
   - name: article_titles
     selector: article h1
     type: css
   - name: article_dates
     selector: article .date
     type: css
   ```

4. **Contextual Selection**
   ```yaml
   - name: active_prices
     selector: .product:not(.sold-out) .price
     type: css
   ```

## Advanced Topics

1. **Using Both CSS and XPath**
   - Mix selector types based on what's most appropriate
   - Use CSS for simple structural queries
   - Use XPath for complex conditions or relationships

2. **Template Functions**
   - Use Sprig functions for data manipulation
   - Common functions: `trim`, `upper`, `lower`, `replace`
   - Date formatting: `now`, `date`, `dateModify`

3. **Context and Paths**
   - Use --show-context to debug selections
   - Use --show-path to understand DOM structure
   - Adjust context_chars based on needs

4. **Performance**
   - Keep selectors as specific as possible
   - Use class and ID selectors when available
   - Avoid overly complex XPath expressions 