package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/spf13/cobra"
)

// shared in-memory corpus
type doc struct{ id, title, text, url string }

var corpus = []doc{
	{"1", "Welcome", "Welcome to the OIDC MCP example.", "https://example.com/welcome"},
	{"2", "OIDC Notes", "OIDC Authorization Code with PKCE example.", "https://example.com/oidc"},
	{"3", "MCP Docs", "Model Context Protocol reference.", "https://example.com/mcp"},
}

func main() {
	rootCmd := &cobra.Command{Use: "oidc-example"}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return logging.InitLoggerFromViper()
	}

	// Add MCP server capability with OIDC enabled
	err := embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("OIDC MCP Server"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("Example MCP server secured by embedded OIDC"),
		embeddable.WithDefaultTransport("sse"),
		embeddable.WithDefaultPort(3001),
		embeddable.WithOIDC(embeddable.OIDCOptions{
			Issuer:          "http://localhost:3001",
			DBPath:          "/tmp/mcp-oidc.db",
			EnableDevTokens: false,
			AuthKey:         "TEST_AUTH_KEY_123",
		}),
		// search tool: returns exactly one text content with JSON-encoded {"results":[{id,title,url},...]}
		embeddable.WithTool("search", searchHandler,
			embeddable.WithDescription("Return relevant results: {results:[{id,title,url}]} JSON as text content"),
			embeddable.WithStringArg("query", "Query string", true),
		),
		// fetch tool: returns exactly one text content with JSON-encoded {id,title,text,url,metadata}
		embeddable.WithTool("fetch", fetchHandler,
			embeddable.WithDescription("Fetch a document by id: {id,title,text,url,metadata} JSON as text content"),
			embeddable.WithStringArg("id", "Document ID", true),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := clay.InitViper("jesus", rootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize viper: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// searchHandler returns a single text content with JSON-encoded results: {"results":[{id,title,url}, ...]}
func searchHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return protocol.NewErrorToolResult(protocol.NewTextContent("query is required")), nil
	}

	// Build results [{id,title,url}]
	type result struct{ ID, Title, URL string }
	var results []result
	for _, d := range corpus {
		if containsFold(d.text, query) || containsFold(d.title, query) {
			results = append(results, result{ID: d.id, Title: d.title, URL: d.url})
		}
	}
	payload := map[string]any{"results": results}
	b, _ := json.Marshal(payload)
	return protocol.NewToolResult(protocol.WithText(string(b))), nil
}

// fetchHandler returns a single text content with JSON-encoded document {id,title,text,url,metadata}
func fetchHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return protocol.NewErrorToolResult(protocol.NewTextContent("id is required")), nil
	}

	for _, d := range corpus {
		if d.id == id {
			obj := map[string]any{
				"id":       d.id,
				"title":    d.title,
				"text":     d.text,
				"url":      d.url,
				"metadata": map[string]any{"source": "in_memory"},
			}
			b, _ := json.Marshal(obj)
			return protocol.NewToolResult(protocol.WithText(string(b))), nil
		}
	}
	return protocol.NewErrorToolResult(protocol.NewTextContent("not found")), nil
}

func containsFold(haystack, needle string) bool {
	return len(needle) == 0 || (indexFold(haystack, needle) >= 0)
}

// indexFold returns case-insensitive index of needle in s or -1.
func indexFold(s, substr string) int {
	// naive implementation to avoid pulling extra deps
	ls, lsub := len(s), len(substr)
	if lsub == 0 {
		return 0
	}
	for i := 0; i+lsub <= ls; i++ {
		if equalFold(s[i:i+lsub], substr) {
			return i
		}
	}
	return -1
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca|0x20 != cb|0x20 { // ascii fold
			if ca >= 'A' && ca <= 'Z' {
				ca += 'a' - 'A'
			}
			if cb >= 'A' && cb <= 'Z' {
				cb += 'a' - 'A'
			}
			if ca != cb {
				return false
			}
		}
	}
	return true
}
