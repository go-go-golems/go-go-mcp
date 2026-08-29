package embeddable

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	official "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/rs/zerolog/log"
)

// Backend represents a runnable server backend that starts the selected
// transport using the official Model Context Protocol Go SDK.
type Backend interface {
	Start(ctx context.Context) error
}

const toolDescriptionPreviewEdge = 80
const shutdownTimeout = 10 * time.Second

// NewBackend constructs an official-SDK backend from the provided ServerConfig.
// It builds an MCP server, registers tools through the existing registry, and
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
		Msg("Creating official MCP SDK backend")

	s := official.NewServer(&official.Implementation{Name: cfg.Name, Version: cfg.Version}, nil)

	// Register tools from our registry into the official SDK server
	if err := registerToolsFromRegistry(context.Background(), s, cfg.toolRegistry); err != nil {
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

// MountHTTPHandlers mounts HTTP-based MCP routes into an existing mux.
// It is intended for applications that already own an http.Server and want to
// expose MCP on the same listener.
func MountHTTPHandlers(mux *http.ServeMux, cfg *ServerConfig) error {
	if mux == nil {
		return fmt.Errorf("nil mux")
	}
	if cfg == nil {
		return fmt.Errorf("nil server config")
	}

	log.Debug().
		Str("name", cfg.Name).
		Str("version", cfg.Version).
		Str("transport", cfg.defaultTransport).
		Msg("Mounting MCP HTTP handlers")

	s := official.NewServer(&official.Implementation{Name: cfg.Name, Version: cfg.Version}, nil)
	if err := registerToolsFromRegistry(context.Background(), s, cfg.toolRegistry); err != nil {
		return err
	}

	switch cfg.defaultTransport {
	case "sse":
		return mountSSEHandlers(mux, s, cfg)
	case "streamable_http":
		return mountStreamableHTTPHandlers(mux, s, cfg)
	case "stdio":
		return fmt.Errorf("stdio transport cannot be mounted into an existing HTTP mux")
	default:
		return fmt.Errorf("unknown transport: %s", cfg.defaultTransport)
	}
}

func registerToolsFromRegistry(ctx context.Context, s *official.Server, reg *tool_registry.Registry) error {
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
		mappedTool := mapToolToOfficial(t)
		name := t.Name
		log.Debug().
			Str("tool", name).
			Str("description_preview", previewDescription(t.Description, toolDescriptionPreviewEdge)).
			Msg("Adding tool to official MCP SDK server")

		// The registry handler owns middleware and hook execution. Applying them
		// again in the SDK adapter would invoke each layer twice for every call.
		s.AddTool(mappedTool, func(callCtx context.Context, req *official.CallToolRequest) (*official.CallToolResult, error) {
			args := map[string]any{}
			if req.Params != nil && len(req.Params.Arguments) > 0 {
				if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
					return nil, fmt.Errorf("decode arguments for tool %q: %w", name, err)
				}
			}
			log.Debug().Str("tool", name).Interface("args", args).Msg("Handling tool call")

			res, err := reg.CallTool(callCtx, name, args)
			if err != nil {
				log.Error().Str("tool", name).Err(err).Msg("Tool call errored")
				return nil, err
			}
			mappedResult, err := mapToolResultToOfficial(res)
			if err != nil {
				return nil, fmt.Errorf("map result for tool %q: %w", name, err)
			}
			log.Debug().Str("tool", name).Bool("is_error", mappedResult.IsError).Msg("Tool call completed")
			return mappedResult, nil
		})
	}

	return nil
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
	server *official.Server
}

func (b *stdioBackend) Start(ctx context.Context) error {
	return b.server.Run(ctx, &official.StdioTransport{})
}

// sse backend

type sseBackend struct {
	server *official.Server
	port   int
	cfg    *ServerConfig
}

func (b *sseBackend) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", b.port)
	mux := http.NewServeMux()
	if err := mountSSEHandlers(mux, b.server, b.cfg); err != nil {
		return err
	}

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		log.Info().Str("addr", addr).Msg("Shutting down SSE HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Debug().Str("addr", addr).Str("endpoint", "/mcp").Msg("Starting SSE server (single-port)")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// streamable-http backend

type streamBackend struct {
	server *official.Server
	port   int
	cfg    *ServerConfig
}

func (b *streamBackend) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", b.port)
	mux := http.NewServeMux()
	if err := mountStreamableHTTPHandlers(mux, b.server, b.cfg); err != nil {
		return err
	}

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		log.Info().Str("addr", addr).Msg("Shutting down Streamable HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Debug().Str("addr", addr).Str("endpoint", "/mcp").Msg("Starting StreamableHTTP server (single-port)")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func mountSSEHandlers(mux *http.ServeMux, server *official.Server, cfg *ServerConfig) error {
	sse := official.NewSSEHandler(func(*http.Request) *official.Server { return server }, nil)

	handler, err := wrapHTTPAuthentication(mux, cfg, sse)
	if err != nil {
		return err
	}

	mux.Handle("/mcp/", withRequestLogging(handler))
	return nil
}

func mountStreamableHTTPHandlers(mux *http.ServeMux, server *official.Server, cfg *ServerConfig) error {
	stream := official.NewStreamableHTTPHandler(
		func(*http.Request) *official.Server { return server },
		&official.StreamableHTTPOptions{Stateless: false, JSONResponse: false},
	)

	handler, err := wrapHTTPAuthentication(mux, cfg, stream)
	if err != nil {
		return err
	}

	mux.Handle("/mcp", withRequestLogging(handler))
	mux.Handle("/mcp/", withRequestLogging(handler))
	return nil
}

// --- Auth helpers ---

const officialPrincipalExtraKey = "go-go-mcp/auth-principal"

func wrapHTTPAuthentication(mux *http.ServeMux, cfg *ServerConfig, next http.Handler) (http.Handler, error) {
	if cfg == nil || !cfg.authEnabled {
		return next, nil
	}
	provider, err := newHTTPAuthProvider(cfg)
	if err != nil {
		return nil, err
	}
	provider.MountRoutes(mux)

	if external, ok := provider.(*externalOIDCAuthProvider); ok {
		metadata := &oauthex.ProtectedResourceMetadata{
			Resource:             external.resourceURL,
			AuthorizationServers: []string{external.discovery.Issuer},
			ScopesSupported:      append([]string(nil), external.opts.RequiredScopes...),
		}
		mux.Handle("/.well-known/oauth-protected-resource", mcpauth.ProtectedResourceMetadataHandler(metadata))
		return officialExternalOIDCMiddleware(external, next), nil
	}

	mux.HandleFunc("/.well-known/oauth-protected-resource", protectedResourceHandler(provider))
	return authMiddleware(provider, next), nil
}

func officialExternalOIDCMiddleware(provider *externalOIDCAuthProvider, next http.Handler) http.Handler {
	verifier := func(ctx context.Context, token string, _ *http.Request) (*mcpauth.TokenInfo, error) {
		principal, err := provider.validateBearerToken(ctx, token, false)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", mcpauth.ErrInvalidToken, err)
		}
		return &mcpauth.TokenInfo{
			UserID:     principal.Subject,
			Scopes:     append([]string(nil), principal.Scopes...),
			Expiration: principal.Expiration,
			Extra:      map[string]any{officialPrincipalExtraKey: principal},
		}, nil
	}

	injectPrincipal := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenInfo := mcpauth.TokenInfoFromContext(r.Context())
		if tokenInfo == nil {
			http.Error(w, "verified token information missing", http.StatusInternalServerError)
			return
		}
		principal, ok := tokenInfo.Extra[officialPrincipalExtraKey].(AuthPrincipal)
		if !ok {
			http.Error(w, "verified principal missing", http.StatusInternalServerError)
			return
		}
		ctx := WithAuthPrincipal(r.Context(), principal)
		r2 := r.Clone(ctx)
		r2.Header.Set("X-MCP-Subject", principal.Subject)
		r2.Header.Set("X-MCP-Client-ID", principal.ClientID)
		next.ServeHTTP(w, r2)
	})

	return mcpauth.RequireBearerToken(verifier, &mcpauth.RequireBearerTokenOptions{
		Scopes:              append([]string(nil), provider.opts.RequiredScopes...),
		ResourceMetadataURL: protectedResourceMetadataURL(provider.resourceURL),
	})(injectPrincipal)
}

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
