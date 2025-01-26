# HTML Analysis Tools Specification

## Overview
Two complementary tools for analyzing and testing HTML documents: an HTML Structure Analyzer and a Selector Testing Tool. The tools are designed to help developers understand HTML document structure and verify CSS/XPath selectors for data extraction.

## Core Design Principles
- Process large HTML files efficiently using streaming where possible
- Provide clear, actionable output in YAML format
- Handle malformed HTML gracefully
- Support both CSS and XPath selectors
- Focus on finding repeating patterns and structures

## Tool 1: HTML Structure Analyzer
### Purpose
Analyzes HTML documents to identify structural patterns, tag usage, and DOM hierarchy. Helps developers understand document organization and identify potential selectors for data extraction.

### Input Parameters
```yaml
file: string           # Path to HTML file
options:
  tag_frequencies: boolean    # Count tag occurrences
  class_frequencies: boolean  # Count class occurrences
  depth_analysis: boolean    # Analyze DOM tree depth
  common_patterns: boolean   # Find repeating structures
```

### Output Format
```yaml
tags:
  div: 145
  span: 67
  a: 234
  # ... other tags
classes:
  product-card: 24
  price: 24
  title: 26
  # ... other classes
depth:
  max: 12
  average: 4.7
  histogram:
    "1": 45
    "2": 156
    # ... depth frequencies
patterns:
  - signature: "div.product(img.thumb+div.info(h3.title+p.price))"
    count: 24
    examples:
      - "<div class='product'><img class='thumb' src='p1.jpg'><div class='info'>...</div></div>"
      # ... more examples
```

### Key Components

#### 1. HTML Parser
```python
class HTMLParser:
    def parse(html_stream: Stream) -> Document:
        # Stream-based parsing to handle large files
        # Returns virtual DOM or similar structure
```

#### 2. Pattern Detector
```python
class PatternDetector:
    def generate_signature(element: Element) -> string:
        # Generate structural signature including:
        tag = element.tagName
        classes = sort(element.classList)
        children = element.children.map(generate_signature)
        
        return format_signature(tag, classes, children)
    
    def find_patterns(doc: Document) -> List[Pattern]:
        signatures = {}
        # Walk DOM tree, generate signatures
        # Group by signature
        # Filter to find repeated patterns
        return patterns
```

#### 3. Statistics Collector
```python
class StatsCollector:
    def collect_all(doc: Document) -> Stats:
        return {
            'tags': count_tags(doc),
            'classes': count_classes(doc),
            'depth': analyze_depth(doc)
        }
```

## Tool 2: Selector Tester
### Purpose
Tests CSS and XPath selectors against HTML documents, providing match counts and contextual examples to verify selector accuracy.

### Input Parameters
```yaml
file: string           # Path to HTML file
selectors:
  - name: string       # Friendly name for selector
    selector: string   # CSS or XPath selector
    type: "css" | "xpath"
config:
  sample_count: number    # Number of examples to show
  context_chars: number   # Characters of context
```

### Output Format
```yaml
results:
  - name: "product_titles"
    selector: ".product-card h2"
    count: 24
    samples:
      - html: "<h2>Product Name</h2>"
        context: "<div class='product-card'>...<h2>Product Name</h2>...</div>"
        path: "html > body > main > div.product-card > h2"
      # ... more samples
```

### Key Components

#### 1. Selector Engine
```python
class SelectorEngine:
    def find_elements(doc: Document, selector: Selector) -> List[Element]:
        if selector.type == "css":
            return doc.querySelectorAll(selector.selector)
        else:
            return doc.evaluate(selector.selector)  # XPath
```

#### 2. Context Extractor
```python
class ContextExtractor:
    def get_element_context(element: Element, chars: number) -> Context:
        parent_html = element.parentElement.outerHTML
        element_pos = find_element_position(element, parent_html)
        
        return {
            'html': element.outerHTML,
            'context': extract_surrounding(parent_html, element_pos, chars),
            'path': generate_dom_path(element)
        }
```

