package embeddable

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

const toolDescriptionPreviewEdge = 80

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
			Str("description_preview", previewDescription(t.Description, toolDescriptionPreviewEdge)).
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

func previewDescription(desc string, edge int) string {
	desc = strings.TrimSpace(desc)
	if edge <= 0 {
		return desc
	}

	runes := []rune(desc)
	if len(runes) <= edge*2 {
		return desc
	}

	start := string(runes[:edge])
	end := string(runes[len(runes)-edge:])
	return fmt.Sprintf("%s...%s", start, end)
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

	if b.cfg != nil && b.cfg.authEnabled {
		provider, err := newHTTPAuthProvider(b.cfg)
		if err != nil {
			return err
		}
		provider.MountRoutes(mux)
		mux.HandleFunc("/.well-known/oauth-protected-resource", protectedResourceHandler(provider))
		handler = authMiddleware(provider, handler)
	}

	// Mount SSE under /mcp/ (ServeHTTP routes internally to /mcp/sse and /mcp/message)
	mux.Handle("/mcp/", withRequestLogging(handler))

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		log.Info().Str("addr", addr).Msg("Shutting down SSE HTTP server")
		_ = server.Shutdown(context.Background())
	}()

	log.Debug().Str("addr", addr).Str("endpoint", "/mcp").Msg("Starting SSE server (single-port)")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
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

	if b.cfg != nil && b.cfg.authEnabled {
		provider, err := newHTTPAuthProvider(b.cfg)
		if err != nil {
			return err
		}
		provider.MountRoutes(mux)
		mux.HandleFunc("/.well-known/oauth-protected-resource", protectedResourceHandler(provider))
		handler = authMiddleware(provider, handler)
	}

	// Mount streamable HTTP under /mcp
	mux.Handle("/mcp", withRequestLogging(handler))
	mux.Handle("/mcp/", withRequestLogging(handler))

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		log.Info().Str("addr", addr).Msg("Shutting down Streamable HTTP server")
		_ = server.Shutdown(context.Background())
	}()

	log.Debug().Str("addr", addr).Str("endpoint", "/mcp").Msg("Starting StreamableHTTP server (single-port)")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// --- Auth helpers ---

func authMiddleware(provider HTTPAuthProvider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if len(authz) < len("Bearer ") || authz[:len("Bearer ")] != "Bearer " {
			advertiseWWWAuthenticate(w, provider)
			log.Warn().Str("path", r.URL.Path).Str("method", r.Method).Str("ua", r.UserAgent()).Str("remote", r.RemoteAddr).Msg("Unauthorized: missing bearer header")
			http.Error(w, "missing bearer", http.StatusUnauthorized)
			return
		}
		tok := authz[len("Bearer "):]

		principal, err := provider.ValidateBearerToken(r.Context(), tok)
		if err != nil {
			advertiseWWWAuthenticate(w, provider)
			log.Warn().Str("path", r.URL.Path).Str("method", r.Method).Str("ua", r.UserAgent()).Str("remote", r.RemoteAddr).Err(err).Msg("Unauthorized: token validation failed")
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		r2 := r.Clone(r.Context())
		r2 = r2.WithContext(WithAuthPrincipal(r2.Context(), principal))
		r2.Header.Set("X-MCP-Subject", principal.Subject)
		r2.Header.Set("X-MCP-Client-ID", principal.ClientID)
		log.Info().Str("path", r.URL.Path).Str("method", r.Method).Str("ua", r.UserAgent()).Str("remote", r.RemoteAddr).Str("subject", principal.Subject).Str("client_id", principal.ClientID).Bool("authorized", true).Msg("Authorized request")
		next.ServeHTTP(w, r2)
	})
}

func protectedResourceHandler(provider HTTPAuthProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		j := provider.ProtectedResourceMetadata()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(j)
		log.Info().Str("endpoint", "/.well-known/oauth-protected-resource").Str("ua", r.UserAgent()).Str("remote", r.RemoteAddr).Interface("response", j).Msg("served protected resource metadata")
	}
}

func advertiseWWWAuthenticate(w http.ResponseWriter, provider HTTPAuthProvider) {
	hdr := provider.WWWAuthenticateHeader()
	w.Header().Set("WWW-Authenticate", hdr)
	log.Debug().Str("header", hdr).Msg("set WWW-Authenticate")
}

// withRequestLogging logs request summaries for debugging while censoring sensitive data.
func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Censor Authorization header presence
		hasAuth := r.Header.Get("Authorization") != ""
		// Censor known sensitive query params
		q := r.URL.Query()
		if q.Has("code") {
			q.Set("code", "***")
		}
		if q.Has("code_verifier") {
			q.Set("code_verifier", "***")
		}
		if q.Has("refresh_token") {
			q.Set("refresh_token", "***")
		}
		if q.Has("token") {
			q.Set("token", "***")
		}

		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("ua", r.UserAgent()).
			Str("remote", r.RemoteAddr).
			Str("accept", r.Header.Get("Accept")).
			Str("content_type", r.Header.Get("Content-Type")).
			Str("x_forwarded_for", r.Header.Get("X-Forwarded-For")).
			Bool("has_authz", hasAuth).
			Str("query", q.Encode()).
			Msg("http request")

		next.ServeHTTP(w, r)
	})
}
