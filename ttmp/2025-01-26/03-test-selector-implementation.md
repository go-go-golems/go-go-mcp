# Implementing the HTML Selector Testing Tool in Go

## Overview
This document describes how to implement the HTML Selector Testing tool in Go, which allows testing CSS and XPath selectors against HTML documents. The tool provides match counts and contextual examples to verify selector accuracy.

## Core Dependencies
```go
import (
    "github.com/PuerkitoBio/goquery"  // For CSS selectors
    "github.com/antchfx/htmlquery"    // For XPath selectors
    "gopkg.in/yaml.v3"                // For YAML input/output
    "context"
    "io"
)
```

## Data Structures

```go
// Configuration represents the input configuration
type Config struct {
    File      string     `yaml:"file"`
    Selectors []Selector `yaml:"selectors"`
    Config    struct {
        SampleCount   int `yaml:"sample_count"`
        ContextChars  int `yaml:"context_chars"`
    } `yaml:"config"`
}

// Selector represents a single selector to test
type Selector struct {
    Name     string `yaml:"name"`
    Selector string `yaml:"selector"`
    Type     string `yaml:"type"` // "css" or "xpath"
}

// Result represents the output for a single selector
type Result struct {
    Name     string    `yaml:"name"`
    Selector string    `yaml:"selector"`
    Count    int       `yaml:"count"`
    Samples  []Sample  `yaml:"samples"`
}

// Sample represents a single matching element with context
type Sample struct {
    HTML    string `yaml:"html"`
    Context string `yaml:"context"`
    Path    string `yaml:"path"`
}
```

## Core Components

### 1. Selector Engine

```go
// SelectorEngine handles both CSS and XPath selectors
type SelectorEngine struct {
    doc *goquery.Document
}

func NewSelectorEngine(r io.Reader) (*SelectorEngine, error) {
    doc, err := goquery.NewDocumentFromReader(r)
    if err != nil {
        return nil, fmt.Errorf("failed to parse HTML: %w", err)
    }
    return &SelectorEngine{doc: doc}, nil
}

func (se *SelectorEngine) FindElements(ctx context.Context, sel Selector) ([]Sample, error) {
    switch sel.Type {
    case "css":
        return se.findWithCSS(ctx, sel.Selector)
    case "xpath":
        return se.findWithXPath(ctx, sel.Selector)
    default:
        return nil, fmt.Errorf("unsupported selector type: %s", sel.Type)
    }
}
```

### 2. Context Extractor

```go
// ContextExtractor extracts surrounding context for matched elements
type ContextExtractor struct {
    contextChars int
}

func (ce *ContextExtractor) ExtractContext(node *goquery.Selection) Sample {
    html, _ := node.Html()
    parent := node.Parent()
    parentHtml, _ := parent.Html()
    
    return Sample{
        HTML:    html,
        Context: ce.truncateContext(parentHtml),
        Path:    ce.generateDOMPath(node),
    }
}

func (ce *ContextExtractor) truncateContext(html string) string {
    if len(html) <= ce.contextChars {
        return html
    }
    // Implement smart truncation that doesn't break HTML tags
    return html[:ce.contextChars] + "..."
}

func (ce *ContextExtractor) generateDOMPath(node *goquery.Selection) string {
    var path []string
    node.Parents().Each(func(i int, s *goquery.Selection) {
        path = append([]string{s.Get(0).Data}, path...)
    })
    return strings.Join(path, " > ")
}
```

### 3. Main Tool Implementation

```go
// SelectorTester is the main tool implementation
type SelectorTester struct {
    engine   *SelectorEngine
    extractor *ContextExtractor
    config   Config
}

func NewSelectorTester(config Config) (*SelectorTester, error) {
    f, err := os.Open(config.File)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer f.Close()

    engine, err := NewSelectorEngine(f)
    if err != nil {
        return nil, err
    }

    return &SelectorTester{
        engine: engine,
        extractor: &ContextExtractor{
            contextChars: config.Config.ContextChars,
        },
        config: config,
    }, nil
}

func (st *SelectorTester) Run(ctx context.Context) ([]Result, error) {
    var results []Result

    for _, sel := range st.config.Selectors {
        samples, err := st.engine.FindElements(ctx, sel)
        if err != nil {
            return nil, fmt.Errorf("selector '%s' failed: %w", sel.Name, err)
        }

        // Limit samples to configured count
        if len(samples) > st.config.Config.SampleCount {
            samples = samples[:st.config.Config.SampleCount]
        }

        results = append(results, Result{
            Name:     sel.Name,
            Selector: sel.Selector,
            Count:    len(samples),
            Samples:  samples,
        })
    }

    return results, nil
}
```

## Usage Example

```go
func main() {
    // Read configuration
    config := Config{
        File: "example.html",
        Selectors: []Selector{
            {
                Name:     "product_titles",
                Selector: ".product-card h2",
                Type:     "css",
            },
        },
        Config: struct {
            SampleCount  int `yaml:"sample_count"`
            ContextChars int `yaml:"context_chars"`
        }{
            SampleCount:  5,
            ContextChars: 100,
        },
    }

    // Create and run tester
    tester, err := NewSelectorTester(config)
    if err != nil {
        log.Fatal(err)
    }

    results, err := tester.Run(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // Output results as YAML
    encoder := yaml.NewEncoder(os.Stdout)
    if err := encoder.Encode(results); err != nil {
        log.Fatal(err)
    }
}
```

## Error Handling and Best Practices

1. Always use context for cancellation support
2. Handle malformed HTML gracefully
3. Implement proper resource cleanup
4. Use streaming for large files when possible
5. Validate selectors before execution
6. Provide clear error messages

## Performance Considerations

1. Use goquery's document caching
2. Implement concurrent selector processing for multiple selectors
3. Stream HTML parsing for large files
4. Optimize context extraction for large documents
5. Use efficient string builders for DOM path generation

## Testing Strategy

1. Unit tests for individual components
2. Integration tests with sample HTML documents
3. Performance benchmarks for large documents
4. Edge case testing for malformed HTML
5. Selector validation tests 