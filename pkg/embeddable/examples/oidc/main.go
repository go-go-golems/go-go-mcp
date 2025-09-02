package main

import (
	"context"
	"fmt"
	"log"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/spf13/cobra"
)

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
		embeddable.WithTool("search", searchHandler,
			embeddable.WithDescription("Search a tiny in-memory corpus and return items"),
			embeddable.WithStringArg("query", "Query string", true),
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

// searchHandler implements a simple in-memory search tool
func searchHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return protocol.NewErrorToolResult(protocol.NewTextContent("query is required")), nil
	}

	// Minimal corpus
	corpus := []struct{ id, title, text, url string }{
		{"1", "Welcome", "Welcome to the OIDC MCP example.", "https://example.com/welcome"},
		{"2", "OIDC Notes", "OIDC Authorization Code with PKCE example.", "https://example.com/oidc"},
		{"3", "MCP Docs", "Model Context Protocol reference.", "https://example.com/mcp"},
	}

	var items []map[string]any
	for _, d := range corpus {
		if query == "" || containsFold(d.text, query) || containsFold(d.title, query) {
			items = append(items, map[string]any{
				"id": d.id, "title": d.title, "text": d.text, "url": d.url,
			})
		}
	}

	// Return results as JSON content
	return protocol.NewToolResult(
		protocol.WithJSON(map[string]any{"items": items}),
		protocol.WithText(fmt.Sprintf("%d result(s)", len(items))),
	), nil
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
