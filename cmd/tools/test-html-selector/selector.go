package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type SelectorSample struct {
	HTML    string `yaml:"html"`
	Context string `yaml:"context"`
	Path    string `yaml:"path"`
}

type SelectorResult struct {
	Name     string           `yaml:"name"`
	Selector string           `yaml:"selector"`
	Type     string           `yaml:"type"`
	Count    int              `yaml:"count"`
	Samples  []SelectorSample `yaml:"samples"`
}

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

func (se *SelectorEngine) findWithCSS(ctx context.Context, selector string) ([]SelectorSample, error) {
	var samples []SelectorSample
	se.doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		parent := s.Parent()
		parentHtml, _ := parent.Html()

		samples = append(samples, SelectorSample{
			HTML:    html,
			Context: parentHtml,
			Path:    generateDOMPath(s),
		})
	})
	return samples, nil
}

func (se *SelectorEngine) findWithXPath(ctx context.Context, selector string) ([]SelectorSample, error) {
	nodes, err := htmlquery.QueryAll(se.doc.Get(0), selector)
	if err != nil {
		return nil, fmt.Errorf("failed to execute XPath query: %w", err)
	}

	var samples []SelectorSample
	for _, node := range nodes {
		html := htmlquery.OutputHTML(node, true)
		parent := node.Parent
		parentHtml := ""
		if parent != nil {
			parentHtml = htmlquery.OutputHTML(parent, true)
		}

		samples = append(samples, SelectorSample{
			HTML:    html,
			Context: parentHtml,
			Path:    generateXPathDOMPath(node),
		})
	}
	return samples, nil
}

func (se *SelectorEngine) FindElements(ctx context.Context, sel Selector) ([]SelectorSample, error) {
	if sel.Type == "" {
		sel.Type = "css"
	}

	switch sel.Type {
	case "css":
		return se.findWithCSS(ctx, sel.Selector)
	case "xpath":
		return se.findWithXPath(ctx, sel.Selector)
	default:
		return nil, fmt.Errorf("unsupported selector type: %s", sel.Type)
	}
}

func generateDOMPath(s *goquery.Selection) string {
	var path []string
	s.Parents().Each(func(i int, p *goquery.Selection) {
		path = append([]string{elementDescriptor(p)}, path...)
	})
	path = append(path, elementDescriptor(s))
	return strings.Join(path, " > ")
}

func elementDescriptor(s *goquery.Selection) string {
	elem := s.Get(0)
	if elem == nil {
		return ""
	}
	id, _ := s.Attr("id")
	class, _ := s.Attr("class")
	descriptor := elem.Data
	if id != "" {
		descriptor += "#" + id
	}
	if class != "" {
		descriptor += "." + strings.ReplaceAll(class, " ", ".")
	}
	return descriptor
}

func generateXPathDOMPath(n *html.Node) string {
	var path []string
	current := n
	for current != nil && current.Type == html.ElementNode {
		descriptor := current.Data
		for _, attr := range current.Attr {
			switch attr.Key {
			case "id":
				descriptor += "#" + attr.Val
			case "class":
				descriptor += "." + strings.ReplaceAll(attr.Val, " ", ".")
			}
		}
		path = append([]string{descriptor}, path...)
		current = current.Parent
	}
	return strings.Join(path, " > ")
}

type SelectorTester struct {
	engine *SelectorEngine
	config *Config
}

func NewSelectorTester(config *Config, f io.Reader) (*SelectorTester, error) {
	engine, err := NewSelectorEngine(f)
	if err != nil {
		return nil, err
	}

	return &SelectorTester{
		engine: engine,
		config: config,
	}, nil
}

func (st *SelectorTester) Run(ctx context.Context) ([]SelectorResult, error) {
	var results []SelectorResult

	for _, sel := range st.config.Selectors {
		samples, err := st.engine.FindElements(ctx, sel)
		if err != nil {
			return nil, fmt.Errorf("selector '%s' failed: %w", sel.Name, err)
		}

		totalCount := len(samples)

		// Limit samples to configured count
		if st.config.Config.SampleCount > 0 && len(samples) > st.config.Config.SampleCount {
			samples = samples[:st.config.Config.SampleCount]
		}

		results = append(results, SelectorResult{
			Name:     sel.Name,
			Selector: sel.Selector,
			Type:     sel.Type,
			Count:    totalCount,
			Samples:  samples,
		})
	}

	return results, nil
}
