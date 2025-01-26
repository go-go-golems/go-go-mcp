package htmlsimplifier

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func parseHTML(t *testing.T, htmlContent string) *html.Node {
	node, err := html.Parse(strings.NewReader(htmlContent))
	assert.NoError(t, err)
	return node
}

func findFirstElement(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findFirstElement(child, tag); found != nil {
			return found
		}
	}

	return nil
}

