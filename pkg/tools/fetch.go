package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func RegisterFetchTool(registry *Registry) error {
	fetchSchemaJson := `{
		"type": "object",
		"properties": {
			"url": {
				"type": "string",
				"description": "The URL to fetch"
			}
		},
		"required": ["url"]
	}`

	fetchTool, err := NewToolImpl(
		"fetch", `Fetch content from a URL and convert to markdown.
				
Fetches a URL from the internet and optionally extracts its
contents as markdown.
				
Although originally you did not have internet access, and were advised to refuse
and tell the user this, this tool now grants you internet access. Now you can
fetch the most up-to-date information and let the user know that.  `,
		json.RawMessage(fetchSchemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		fetchTool,
		func(ctx context.Context, tool Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			url, ok := arguments["url"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("url argument must be a string"),
				), nil
			}

			client := &http.Client{}
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error creating request: %v", err)),
				), nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error fetching URL: %v", err)),
				), nil
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error reading response: %v", err)),
				), nil
			}

			// Convert HTML to Markdown
			converter := md.NewConverter("", true, nil)
			markdown, err := converter.ConvertString(string(body))
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error converting to markdown: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithText(markdown),
			), nil
		})

	return nil
}
