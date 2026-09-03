package embeddable

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/spf13/cobra"
)

// ToolHandler is a simplified function signature for tool handlers
// Session information is available via session.GetSessionFromContext(ctx)
type ToolHandler func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error)

// ServerConfig holds the configuration for the embeddable server
type ServerConfig struct {
	Name        string
	Version     string
	Description string

	// Tool registration and decoded-dispatch authorization.
	toolRegistry      *tool_registry.Registry
	toolAuthorization map[string]ToolAuthorizationPolicy

	// Transport options
	defaultTransport           string
	defaultPort                int
	streamableHTTPStateless    bool
	streamableHTTPJSONResponse bool

	// Configuration options
	enableConfig bool
	configFile   string

	// Internal servers
	internalServers []string

	// Advanced options
	sessionStore       session.SessionStore
	middleware         []ToolMiddleware
	hooks              *Hooks
	commandCustomizers []CommandCustomizer

	// Auth options (HTTP transports only)
	authEnabled        bool
	authOptions        AuthOptions
	customAuthVerifier HTTPAuthVerifier
}

// ToolMiddleware is a function that wraps a ToolHandler
type ToolMiddleware func(next ToolHandler) ToolHandler

// Hooks allows customization of server behavior
type Hooks struct {
	OnServerStart  func(ctx context.Context) error
	BeforeToolCall func(ctx context.Context, toolName string, args map[string]interface{}) error
	AfterToolCall  func(ctx context.Context, toolName string, result *protocol.ToolResult, err error)
}

// CommandCustomizer is a function that can customize a cobra.Command
type CommandCustomizer func(*cobra.Command) error

// ServerOption configures the embeddable MCP server
type ServerOption func(*ServerConfig) error

// NewServerConfig creates a new server configuration with defaults
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Name:                       "Embeddable MCP Server",
		Version:                    "1.0.0",
		Description:                "MCP Server",
		toolRegistry:               tool_registry.NewRegistry(),
		toolAuthorization:          map[string]ToolAuthorizationPolicy{},
		defaultTransport:           "stdio",
		defaultPort:                3000,
		streamableHTTPStateless:    true,
		streamableHTTPJSONResponse: false,
		enableConfig:               false,
		sessionStore:               session.NewInMemorySessionStore(),
	}
}

// Core options
func WithName(name string) ServerOption {
	return func(config *ServerConfig) error {
		config.Name = name
		return nil
	}
}

func WithVersion(version string) ServerOption {
	return func(config *ServerConfig) error {
		config.Version = version
		return nil
	}
}

func WithServerDescription(description string) ServerOption {
	return func(config *ServerConfig) error {
		config.Description = description
		return nil
	}
}

// Transport options
func WithDefaultTransport(transport string) ServerOption {
	return func(config *ServerConfig) error {
		config.defaultTransport = transport
		return nil
	}
}

func WithDefaultPort(port int) ServerOption {
	return func(config *ServerConfig) error {
		config.defaultPort = port
		return nil
	}
}

// WithStreamableHTTPStateless selects stateless Streamable HTTP. It defaults
// to true, which is required for MCP 2026-07-28 and later.
func WithStreamableHTTPStateless(stateless bool) ServerOption {
	return func(config *ServerConfig) error {
		config.streamableHTTPStateless = stateless
		return nil
	}
}

// WithStreamableHTTPJSONResponse controls whether request responses are sent
// as application/json instead of SSE. It defaults to false.
func WithStreamableHTTPJSONResponse(jsonResponse bool) ServerOption {
	return func(config *ServerConfig) error {
		config.streamableHTTPJSONResponse = jsonResponse
		return nil
	}
}

// Tool registration options
func WithTool(name string, handler ToolHandler, opts ...ToolOption) ServerOption {
	return func(config *ServerConfig) error {
		// Create tool configuration
		toolConfig := &ToolConfig{
			Description: "",
			Schema:      map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		}

		// Apply tool options
		for _, opt := range opts {
			if err := opt(toolConfig); err != nil {
				return err
			}
		}

		// Create the tool
		tool, err := createToolFromConfig(name, toolConfig)
		if err != nil {
			return err
		}

		// Register the tool with handler
		config.toolRegistry.RegisterToolWithHandler(tool, func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			// Apply middleware
			finalHandler := handler
			for i := len(config.middleware) - 1; i >= 0; i-- {
				finalHandler = config.middleware[i](finalHandler)
			}

			// Apply hooks
			if config.hooks != nil && config.hooks.BeforeToolCall != nil {
				if err := config.hooks.BeforeToolCall(ctx, name, arguments); err != nil {
					return nil, err
				}
			}

			// Call the handler
			result, err := finalHandler(ctx, arguments)

			// Apply hooks
			if config.hooks != nil && config.hooks.AfterToolCall != nil {
				config.hooks.AfterToolCall(ctx, name, result, err)
			}

			return result, err
		})

		return nil
	}
}

func WithToolRegistry(registry *tool_registry.Registry) ServerOption {
	return func(config *ServerConfig) error {
		if registry != nil {
			config.toolRegistry = registry
		}
		return nil
	}
}

// Advanced options
func WithSessionStore(store session.SessionStore) ServerOption {
	return func(config *ServerConfig) error {
		if store != nil {
			config.sessionStore = store
		}
		return nil
	}
}

func WithMiddleware(middleware ...ToolMiddleware) ServerOption {
	return func(config *ServerConfig) error {
		config.middleware = append(config.middleware, middleware...)
		return nil
	}
}

func WithHooks(hooks *Hooks) ServerOption {
	return func(config *ServerConfig) error {
		config.hooks = hooks
		return nil
	}
}

func WithConfigEnabled(enabled bool) ServerOption {
	return func(config *ServerConfig) error {
		config.enableConfig = enabled
		return nil
	}
}

func WithConfigFile(file string) ServerOption {
	return func(config *ServerConfig) error {
		config.configFile = file
		return nil
	}
}

func WithInternalServers(servers ...string) ServerOption {
	return func(config *ServerConfig) error {
		config.internalServers = append(config.internalServers, servers...)
		return nil
	}
}

type AuthMode string

const (
	AuthModeNone         AuthMode = "none"
	AuthModeEmbeddedDev  AuthMode = "embedded_dev"
	AuthModeExternalOIDC AuthMode = "external_oidc"
)

// OIDCOptions is kept as a compatibility alias for the legacy embedded-only API.
type OIDCOptions = EmbeddedOIDCOptions

type EmbeddedOIDCOptions struct {
	Issuer          string
	DBPath          string
	EnableDevTokens bool
	// Optional static auth key for direct bearer access (testing/dev only)
	AuthKey string
	// Optional static user/password for the embedded login form (testing/dev only)
	User string
	Pass string
}

type ExternalOIDCOptions struct {
	IssuerURL      string
	DiscoveryURL   string
	Audience       string
	RequiredScopes []string
}

type AuthOptions struct {
	Mode        AuthMode
	ResourceURL string
	Embedded    EmbeddedOIDCOptions
	External    ExternalOIDCOptions
}

func (o AuthOptions) Enabled() bool {
	return o.Mode != "" && o.Mode != AuthModeNone
}

func (o AuthOptions) EffectiveResourceURL() string {
	if strings.TrimSpace(o.ResourceURL) != "" {
		return strings.TrimSpace(o.ResourceURL)
	}

	switch o.Mode {
	case AuthModeNone:
		return ""
	case AuthModeEmbeddedDev:
		return strings.TrimRight(strings.TrimSpace(o.Embedded.Issuer), "/") + "/mcp"
	case AuthModeExternalOIDC:
		return ""
	default:
		return ""
	}
}

func WithAuth(opts AuthOptions) ServerOption {
	return func(config *ServerConfig) error {
		config.authEnabled = opts.Enabled()
		config.authOptions = opts
		config.customAuthVerifier = nil
		return nil
	}
}

// WithHTTPAuthVerifier installs an application-owned MCP resource-server
// verifier. The application mounts any authorization-server routes separately;
// this verifier owns only bearer validation, protected-resource metadata, and
// challenge formatting. A later WithAuth call replaces it.
func WithHTTPAuthVerifier(verifier HTTPAuthVerifier) ServerOption {
	return func(config *ServerConfig) error {
		if verifier == nil {
			return fmt.Errorf("HTTP auth verifier is required")
		}
		config.authEnabled = true
		config.authOptions = AuthOptions{}
		config.customAuthVerifier = verifier
		return nil
	}
}

func WithOIDC(opts OIDCOptions) ServerOption {
	return WithAuth(AuthOptions{
		Mode:     AuthModeEmbeddedDev,
		Embedded: opts,
	})
}

func WithCommandCustomizer(customizer CommandCustomizer) ServerOption {
	return func(config *ServerConfig) error {
		config.commandCustomizers = append(config.commandCustomizers, customizer)
		return nil
	}
}

// GetCommandFlags retrieves command flags from context
func GetCommandFlags(ctx context.Context) (map[string]interface{}, bool) {
	flags, ok := ctx.Value(CommandFlagsKey).(map[string]interface{})
	return flags, ok
}

// GetToolProvider returns the tool provider for the server config
func (c *ServerConfig) GetToolProvider() pkg.ToolProvider {
	return c.toolRegistry
}

// createToolFromConfig creates a tool from a ToolConfig
func createToolFromConfig(name string, config *ToolConfig) (tools.Tool, error) {
	// Convert schema to JSON
	var schemaBytes []byte
	var err error

	switch s := config.Schema.(type) {
	case json.RawMessage:
		schemaBytes = s
	case string:
		schemaBytes = []byte(s)
	default:
		schemaBytes, err = json.Marshal(s)
		if err != nil {
			return nil, err
		}
	}

	return tools.NewToolImpl(name, config.Description, json.RawMessage(schemaBytes))
}
