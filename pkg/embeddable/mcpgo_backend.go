package embeddable

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"
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
		return &sseBackend{server: s, port: cfg.defaultPort, cfg: cfg}, nil
	case "streamable_http":
		log.Debug().Str("transport", "streamable_http").Int("port", cfg.defaultPort).Msg("Selected transport")
		return &streamBackend{server: s, port: cfg.defaultPort, cfg: cfg}, nil
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
	cfg    *ServerConfig
}

func (b *sseBackend) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", b.port)

	// Create mcp-go SSE server and mount on mux
	sse := mcpserver.NewSSEServer(b.server, mcpserver.WithStaticBasePath("/mcp"))
	mux := http.NewServeMux()

	var handler http.Handler = sse

	if b.cfg != nil && b.cfg.oidcEnabled {
		// OIDC-enabled: create OIDC server, mount routes, and protect /mcp
		oidcSrv, err := oidc.New(oidc.Config{
			Issuer:          b.cfg.oidcOptions.Issuer,
			DBPath:          b.cfg.oidcOptions.DBPath,
			EnableDevTokens: b.cfg.oidcOptions.EnableDevTokens,
		})
		if err != nil {
			return err
		}
		oidcSrv.Routes(mux)
		mux.HandleFunc("/.well-known/oauth-protected-resource", protectedResourceHandler(b.cfg))
		handler = oidcAuthMiddleware(b.cfg, oidcSrv, handler)
	}

	// Mount SSE under /mcp/ (ServeHTTP routes internally to /mcp/sse and /mcp/message)
	mux.Handle("/mcp/", handler)

	log.Debug().Str("addr", addr).Str("endpoint", "/mcp").Msg("Starting SSE server (single-port)")
	return http.ListenAndServe(addr, mux)
}

// streamable-http backend

type streamBackend struct {
	server *mcpserver.MCPServer
	port   int
	cfg    *ServerConfig
}

func (b *streamBackend) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", b.port)

	// Create mcp-go Streamable HTTP server and mount on mux
	stream := mcpserver.NewStreamableHTTPServer(b.server)
	mux := http.NewServeMux()

	var handler http.Handler = stream

	if b.cfg != nil && b.cfg.oidcEnabled {
		// OIDC-enabled: create OIDC server, mount routes, and protect /mcp
		oidcSrv, err := oidc.New(oidc.Config{
			Issuer:          b.cfg.oidcOptions.Issuer,
			DBPath:          b.cfg.oidcOptions.DBPath,
			EnableDevTokens: b.cfg.oidcOptions.EnableDevTokens,
		})
		if err != nil {
			return err
		}
		oidcSrv.Routes(mux)
		mux.HandleFunc("/.well-known/oauth-protected-resource", protectedResourceHandler(b.cfg))
		handler = oidcAuthMiddleware(b.cfg, oidcSrv, handler)
	}

	// Mount streamable HTTP under /mcp
	mux.Handle("/mcp", handler)
	mux.Handle("/mcp/", handler)

	log.Debug().Str("addr", addr).Str("endpoint", "/mcp").Msg("Starting StreamableHTTP server (single-port)")
	return http.ListenAndServe(addr, mux)
}

// --- OIDC helpers ---

func oidcAuthMiddleware(cfg *ServerConfig, oidcSrv *oidc.Server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/.well-known/openid-configuration" || p == "/.well-known/oauth-authorization-server" || p == "/jwks.json" || p == "/login" || strings.HasPrefix(p, "/oauth2/") || p == "/register" || p == "/.well-known/oauth-protected-resource" {
			next.ServeHTTP(w, r)
			return
		}

		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			advertiseWWWAuthenticate(w, cfg)
			http.Error(w, "missing bearer", http.StatusUnauthorized)
			return
		}
		tok := strings.TrimPrefix(authz, "Bearer ")

		// Accept static AuthKey for testing if configured
		if cfg.oidcOptions.AuthKey != "" && tok == cfg.oidcOptions.AuthKey {
			r2 := r.Clone(r.Context())
			r2.Header.Set("X-MCP-Subject", "static-key-user")
			r2.Header.Set("X-MCP-Client-ID", "static-key-client")
			next.ServeHTTP(w, r2)
			return
		}

		subj, cid, ok, err := oidcSrv.IntrospectAccessToken(r.Context(), tok)
		if err != nil || !ok {
			advertiseWWWAuthenticate(w, cfg)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		r2 := r.Clone(r.Context())
		r2.Header.Set("X-MCP-Subject", subj)
		r2.Header.Set("X-MCP-Client-ID", cid)
		next.ServeHTTP(w, r2)
	})
}

func protectedResourceHandler(cfg *ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		j := map[string]any{
			"authorization_servers": []string{cfg.oidcOptions.Issuer},
			"resource":              cfg.oidcOptions.Issuer + "/mcp",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(j)
	}
}

func advertiseWWWAuthenticate(w http.ResponseWriter, cfg *ServerConfig) {
	asMeta := cfg.oidcOptions.Issuer + "/.well-known/oauth-authorization-server"
	prm := cfg.oidcOptions.Issuer + "/.well-known/oauth-protected-resource"
	hdr := "Bearer realm=\"mcp\", resource=\"" + cfg.oidcOptions.Issuer + "/mcp\"" + ", authorization_uri=\"" + asMeta + "\", resource_metadata=\"" + prm + "\""
	w.Header().Set("WWW-Authenticate", hdr)
}
