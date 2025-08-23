package embeddable

import (
	"context"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// Backend represents a runnable server backend
// that starts the selected transport using an mcp-go MCPServer.
type Backend interface {
	Start(ctx context.Context) error
}

// NewBackend constructs an mcp-go based backend from the provided ServerConfig.
// It builds an MCP server, registers tools via existing configuration, and
// returns a transport-specific backend that can Start(ctx).
func NewBackend(cfg *ServerConfig) (Backend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil server config")
	}

	log.Debug().
		Str("name", cfg.Name).
		Str("version", cfg.Version).
		Str("transport", cfg.defaultTransport).
		Int("port", cfg.defaultPort).
		Msg("Creating mcp-go backend")

	// Build mcp-go server
	s := mcpserver.NewMCPServer(cfg.Name, cfg.Version,
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithLogging(),
	)

	// Register tools from our registry into mcp-go server
	if err := registerToolsFromRegistry(context.Background(), s, cfg.toolRegistry, cfg); err != nil {
		return nil, err
	}

	switch cfg.defaultTransport {
	case "stdio":
		log.Debug().Str("transport", "stdio").Msg("Selected transport")
		return &stdioBackend{server: s}, nil
	case "sse":
		log.Debug().Str("transport", "sse").Int("port", cfg.defaultPort).Msg("Selected transport")
		return &sseBackend{server: s, port: cfg.defaultPort}, nil
	case "streamable_http":
		log.Debug().Str("transport", "streamable_http").Int("port", cfg.defaultPort).Msg("Selected transport")
		return &streamBackend{server: s, port: cfg.defaultPort}, nil
	default:
		return nil, fmt.Errorf("unknown transport: %s", cfg.defaultTransport)
	}
}

func registerToolsFromRegistry(ctx context.Context, s *mcpserver.MCPServer, reg *tool_registry.Registry, cfg *ServerConfig) error {
	if reg == nil {
		log.Debug().Msg("No tool registry set; skipping registration")
		return nil
	}

	tools, _, err := reg.ListTools(ctx, "")
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	log.Debug().Int("count", len(tools)).Msg("Registering tools")

	for _, t := range tools {
		// Map our protocol.Tool to mcp-go Tool with raw schema
		mt := mcp.NewToolWithRawSchema(t.Name, t.Description, t.InputSchema)

		name := t.Name
		log.Debug().
			Str("tool", name).
			Str("description", t.Description).
			Msg("Adding tool to mcp-go server")

		// Build a wrapped handler that applies middleware and hooks around registry.CallTool
		baseHandler := func(callCtx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
			return reg.CallTool(callCtx, name, args)
		}

		// Apply middleware stack (reverse order)
		wrapped := baseHandler
		if len(cfg.middleware) > 0 {
			log.Debug().Str("tool", name).Int("middleware_count", len(cfg.middleware)).Msg("Applying middleware chain")
			for i := len(cfg.middleware) - 1; i >= 0; i-- {
				wrapped = cfg.middleware[i](wrapped)
			}
		}

		// Adapter for mcp-go handler signature
		s.AddTool(mt, func(callCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()

			if cfg.hooks != nil && cfg.hooks.BeforeToolCall != nil {
				if err := cfg.hooks.BeforeToolCall(callCtx, name, args); err != nil {
					return nil, err
				}
			}

			log.Debug().Str("tool", name).Interface("args", args).Msg("Handling tool call")

			res, err := wrapped(callCtx, args)

			mcpRes := mapToolResultToMCP(res)

			if cfg.hooks != nil && cfg.hooks.AfterToolCall != nil {
				cfg.hooks.AfterToolCall(callCtx, name, res, err)
			}

			if err != nil {
				log.Error().Str("tool", name).Err(err).Msg("Tool call errored")
			} else {
				log.Debug().Str("tool", name).Bool("is_error", mcpRes.IsError).Msg("Tool call completed")
			}
			return mcpRes, err
		})
	}

	return nil
}

func mapToolResultToMCP(res *protocol.ToolResult) *mcp.CallToolResult {
	if res == nil {
		return &mcp.CallToolResult{}
	}

	out := &mcp.CallToolResult{
		IsError: res.IsError,
	}

	for _, c := range res.Content {
		switch c.Type {
		case "text":
			out.Content = append(out.Content, mcp.TextContent{Type: "text", Text: c.Text})
		case "image":
			out.Content = append(out.Content, mcp.ImageContent{Type: "image", Data: c.Data, MIMEType: c.MimeType})
		case "resource":
			if c.Resource != nil {
				var rc mcp.ResourceContents
				if c.Resource.Blob != "" {
					rc = mcp.BlobResourceContents{URI: c.Resource.URI, MIMEType: c.Resource.MimeType, Blob: c.Resource.Blob}
				} else {
					rc = mcp.TextResourceContents{URI: c.Resource.URI, MIMEType: c.Resource.MimeType, Text: c.Resource.Text}
				}
				embedded := mcp.NewEmbeddedResource(rc)
				out.Content = append(out.Content, embedded)
			}
		}
	}

	return out
}

// stdio backend

type stdioBackend struct {
	server *mcpserver.MCPServer
}

func (b *stdioBackend) Start(ctx context.Context) error {
	// Use ServeStdio convenience which binds to os.Stdin/os.Stdout
	return mcpserver.ServeStdio(b.server)
}

// sse backend

type sseBackend struct {
	server *mcpserver.MCPServer
	port   int
}

func (b *sseBackend) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", b.port)
	return mcpserver.NewSSEServer(b.server).Start(addr)
}

// streamable-http backend

type streamBackend struct {
	server *mcpserver.MCPServer
	port   int
}

func (b *streamBackend) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", b.port)
	return mcpserver.NewStreamableHTTPServer(b.server).Start(addr)
}
