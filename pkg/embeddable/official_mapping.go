package embeddable

import (
	"encoding/base64"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	official "github.com/modelcontextprotocol/go-sdk/mcp"
)

func mapToolToOfficial(tool protocol.Tool) *official.Tool {
	mapped := &official.Tool{
		Name:        tool.Name,
		Title:       tool.Title,
		Description: tool.Description,
		InputSchema: tool.InputSchema,
		Meta:        cloneMeta(tool.Meta),
	}
	if len(tool.OutputSchema) > 0 {
		mapped.OutputSchema = tool.OutputSchema
	}
	if tool.Annotations != nil {
		mapped.Annotations = &official.ToolAnnotations{
			Title:           tool.Annotations.Title,
			ReadOnlyHint:    tool.Annotations.ReadOnlyHint,
			DestructiveHint: tool.Annotations.DestructiveHint,
			IdempotentHint:  tool.Annotations.IdempotentHint,
			OpenWorldHint:   tool.Annotations.OpenWorldHint,
		}
	}
	return mapped
}

func mapToolResultToOfficial(result *protocol.ToolResult) (*official.CallToolResult, error) {
	if result == nil {
		return &official.CallToolResult{}, nil
	}
	mapped := &official.CallToolResult{
		Meta:              cloneMeta(result.Meta),
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
	}
	for _, content := range result.Content {
		switch content.Type {
		case "text":
			mapped.Content = append(mapped.Content, &official.TextContent{Text: content.Text})
		case "image":
			data, err := base64.StdEncoding.DecodeString(content.Data)
			if err != nil {
				return nil, fmt.Errorf("decode image content: %w", err)
			}
			mapped.Content = append(mapped.Content, &official.ImageContent{Data: data, MIMEType: content.MimeType})
		case "resource":
			if content.Resource == nil {
				return nil, fmt.Errorf("resource content is missing resource payload")
			}
			resource := &official.ResourceContents{
				URI:      content.Resource.URI,
				MIMEType: content.Resource.MimeType,
				Text:     content.Resource.Text,
			}
			if content.Resource.Blob != "" {
				blob, err := base64.StdEncoding.DecodeString(content.Resource.Blob)
				if err != nil {
					return nil, fmt.Errorf("decode resource blob: %w", err)
				}
				resource.Blob = blob
			}
			mapped.Content = append(mapped.Content, &official.EmbeddedResource{Resource: resource})
		default:
			return nil, fmt.Errorf("unsupported tool content type %q", content.Type)
		}
	}
	return mapped, nil
}

func cloneMeta(meta map[string]any) official.Meta {
	if len(meta) == 0 {
		return nil
	}
	cloned := make(official.Meta, len(meta))
	for key, value := range meta {
		cloned[key] = value
	}
	return cloned
}
