# Analysis and Design of OIDC Support for go-go-mcp Framework

**Date:** January 9, 2025  
**Author:** Analysis of existing `mcp-oidc-server` and `go-go-mcp` frameworks  
**Goal:** Transform the standalone OIDC server into reusable framework components  

## Executive Summary

This document analyzes the existing `mcp-oidc-server` implementation and the `go-go-mcp` embeddable framework to design a comprehensive integration strategy. The goal is to create reusable OIDC authentication components that can be easily integrated into any MCP application built with the go-go-mcp framework.

## Pragmatic Review and MVP Plan

The original design sketches a rich, highly-pluggable auth system. That’s valuable long-term, but we can ship value sooner by focusing on a minimal, practical path that reuses what already works.

- **MVP focus (ship in days):**
  - **Protect one HTTP transport (SSE or streamable HTTP) with OIDC** using the existing `idsrv` (Fosite) for discovery, login, token issuance.
  - **Introduce a small auth proxy/gateway** that enforces Bearer on `/mcp` and forwards to the existing `mcp-go` SSE/HTTP server running on localhost.
  - **Advertise RFC 9728 resource metadata** and emit proper `WWW-Authenticate` on 401.
  - **Keep all tools protected (coarse-grained)**; no per-tool ACLs in MVP.
  - **Optional dev tokens** for local testing, off by default in non-dev.

- **What we defer (later phases):**
  - Generic `AuthProvider` abstraction, bearer-only providers, API-key providers.
  - Tool-level authorization and scope enforcement in the tool registry.
  - A unified persistence abstraction; reuse `idsrv` + SQLite as-is.
  - Multi-transport auth parity; start with SSE or streamable HTTP only.
  - Rate limiting, CORS hardening, RBAC/ABAC policies.

- **Why this is pragmatic:**
  - Reuses the proven `idsrv` and the `mcp-go` server as-is.
  - Minimizes invasive changes to `go-go-mcp` internals; can be wired via existing embeddable hooks.
  - Keeps a clean migration path to a pluggable provider model later.

### MVP: concrete deliverables

- Add an `auth-proxy` package that:
  - Terminates HTTP, validates `Authorization: Bearer` against `idsrv` introspection (with optional dev-token fallback in dev).
  - Forwards authenticated requests to the `mcp-go` SSE/HTTP server on `127.0.0.1:<port>`.
  - Serves `/.well-known/oauth-protected-resource` and returns `WWW-Authenticate` on 401.
- Add an embeddable option and hook:
  - `WithOIDCProxy(issuer, dbPath, upstreamPort)` or a `WithHooks(OnServerStart)` helper that starts `idsrv` + proxy before the backend.
  - Document that OIDC auth applies only to HTTP transports; stdio is unaffected.
- Provide docs + a short example app using the embeddable API to start an OIDC-protected SSE server.

This MVP makes the OIDC story usable with minimal churn, while the rest of this document outlines the longer-term direction.

## Current State Analysis

### 1. MCP-OIDC-Server Architecture

The current `mcp-oidc-server` is a standalone application with the following key components:

#### Core Components
- **Identity Server (`pkg/idsrv/`)**: Complete OIDC/OAuth2 provider using Fosite
  - RSA key generation and JWKS endpoints
  - Authorization code flow with PKCE support
  - Token introspection and validation
  - Dynamic client registration
  - SQLite persistence for clients, keys, and tokens
- **Application Server (`pkg/server/`)**: HTTP server with MCP integration
  - Bearer token authentication middleware
  - JSON-RPC MCP endpoint implementation
  - RFC 9728 protected resource metadata advertisement
  - Request/response logging and metrics

#### Key Features
- Full OIDC/OAuth2 compliance with discovery endpoints
- SQLite persistence for production deployment
- Development token fallback system
- Comprehensive logging and debugging support
- CLI management tools for tokens and clients

#### Current Limitations
- Monolithic architecture - cannot be reused by other applications
- Hard-coded MCP tools (`search`, `fetch`) 
- No pluggable authentication mechanisms
- Fixed server configuration and endpoints

### 2. Go-Go-MCP Embeddable Framework Architecture

The `go-go-mcp` embeddable framework provides:

#### Core Features
- **Cobra Integration**: Standard MCP subcommands for existing applications
- **Multiple Transports**: stdio, SSE, and streamable HTTP support
- **Tool Registration**: Function-based, struct-based, and reflection-based tools
- **Session Management**: Context-based session access
- **Middleware Support**: Configurable tool call middleware
- **Enhanced APIs**: Type-safe argument handling and property configuration

#### Architecture Strengths
- Modular design with clear separation of concerns
- Flexible tool registration mechanisms
- Transport abstraction allowing multiple protocols
- Comprehensive configuration options
- Backend abstraction using `mcp-go` library

#### Current Limitations
- No built-in authentication mechanisms
- Limited security features for HTTP transports
- No support for protected resources or authorization

## Integration Design

### 1. Proposed Architecture

We propose creating a layered architecture that extends the embeddable framework with OIDC capabilities:

```
┌─────────────────────────────────────────────────────────┐
│                  Application Layer                      │
│  (User's Go Application with MCP + OIDC Support)       │
├─────────────────────────────────────────────────────────┤
│              OIDC-Enhanced Embeddable                   │
│  ┌─────────────────┐  ┌─────────────────────────────────┤
│  │  Auth Providers │  │     Enhanced Server Config     │
│  │  - OIDC/OAuth2  │  │  - Auth-aware transports       │
│  │  - Bearer Token │  │  - Protected tool registration  │
│  │  - Custom Auth  │  │  - Security middleware         │
│  └─────────────────┘  └─────────────────────────────────┤
├─────────────────────────────────────────────────────────┤
│              Core Embeddable Framework                  │
│  ┌─────────────────┐  ┌─────────────────────────────────┤
│  │  Tool Registry  │  │       Backend & Transport      │
│  │  - Registration │  │    - stdio, SSE, HTTP         │
│  │  - Middleware   │  │    - mcp-go integration        │
│  │  - Execution    │  │    - Session management        │
│  └─────────────────┘  └─────────────────────────────────┤
├─────────────────────────────────────────────────────────┤
│                 Protocol Layer                          │
│         (MCP Protocol Implementation)                   │
└─────────────────────────────────────────────────────────┘
```

### 2. Key Components to Extract and Refactor

Note: The following sections describe the long-term, pluggable design. For the MVP, we’ll implement only the minimal proxy-based integration described above. Inline annotations call out what’s deferred.

#### A. Authentication Provider Abstraction

Create a pluggable authentication system:

Pragmatic note:
- MVP does not need this full interface; we can start by hard-wiring OIDC via `idsrv` + proxy.
- Keep this interface design as a target for later refactor.

```go
// AuthProvider interface for different authentication mechanisms
type AuthProvider interface {
    Name() string
    Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error)
    ValidateToken(ctx context.Context, token string) (*TokenInfo, error)
    Routes(mux *http.ServeMux) // Optional HTTP routes for auth flow
}

// AuthRequest contains authentication details from HTTP request
type AuthRequest struct {
    Headers    map[string]string
    Method     string
    Path       string
    Body       []byte
    RemoteAddr string
}

// AuthResult contains authentication outcome
type AuthResult struct {
    Success    bool
    Subject    string
    ClientID   string
    Scopes     []string
    Claims     map[string]interface{}
    Error      error
}
```

#### B. OIDC Provider Implementation

Extract and refactor the current OIDC server into a reusable provider:

Pragmatic note:
- For MVP, do not refactor `idsrv`. Start it as-is and expose its routes under the proxy. Extraction into `pkg/auth/oidc` is a later step.

```go
// OIDCProvider implements AuthProvider for OAuth2/OIDC authentication
type OIDCProvider struct {
    issuer       string
    server       *idsrv.Server
    config       *OIDCConfig
}

type OIDCConfig struct {
    Issuer          string
    PrivateKey      *rsa.PrivateKey
    SQLiteDB        string
    DevFallback     bool
    ClientDefaults  *ClientConfig
    Scopes          []string
}
```

#### C. Enhanced Server Configuration

Extend the embeddable framework to support authentication:

Pragmatic note:
- For MVP, avoid expanding `ServerConfig` surface area. Add a single helper like `WithOIDCProxy(...)` or use `WithHooks(OnServerStart)` to start `idsrv` + proxy alongside the chosen HTTP transport.

```go
// Enhanced ServerConfig with authentication support
type ServerConfig struct {
    // ... existing fields ...
    
    // Authentication configuration
    authProvider     AuthProvider
    authRequired     bool
    protectedTools   []string
    publicTools      []string
    
    // Security options
    corsEnabled      bool
    corsOrigins      []string
    rateLimiting     *RateLimitConfig
}

// Configuration options
func WithOIDCAuth(config *OIDCConfig) ServerOption
func WithBearerTokenAuth(validator TokenValidator) ServerOption
func WithCustomAuth(provider AuthProvider) ServerOption
func WithProtectedTools(tools ...string) ServerOption
func WithPublicTools(tools ...string) ServerOption
```

#### D. Authentication Middleware Integration

Integrate authentication into the transport layer:

Pragmatic note:
- Instead of modifying all backends, place the auth proxy in front of the existing SSE/streamable HTTP server. This isolates auth concerns and avoids touching `mcp-go` internals.

```go
// AuthMiddleware wraps tool calls with authentication
type AuthMiddleware struct {
    provider AuthProvider
    config   *AuthConfig
}

func (m *AuthMiddleware) Wrap(next ToolHandler) ToolHandler {
    return func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
        // Extract authentication from context
        auth, ok := GetAuthFromContext(ctx)
        if !ok || !auth.Success {
            return protocol.NewErrorToolResult(
                protocol.NewTextContent("Authentication required")), nil
        }
        
        // Proceed with authenticated context
        return next(ctx, args)
    }
}
```

### 3. Transport Layer Integration

#### A. HTTP Transport Enhancement

Pragmatic approach (MVP):
- Do not modify existing backends. Run the `mcp-go` SSE/HTTP server on localhost, and front it with an `auth-proxy` that terminates HTTP, validates Bearer tokens via `idsrv`, and forwards to the upstream.
- This preserves the current embeddable backend code and keeps auth changes isolated.

Longer-term, we can fold this proxy behavior into a first-class `AuthenticatedHTTPBackend` once the design stabilizes.

Enhance the HTTP transports (SSE, streamable HTTP) with authentication (future direction):

```go
// AuthenticatedHTTPBackend wraps existing backends with auth
type AuthenticatedHTTPBackend struct {
    backend  Backend
    auth     AuthProvider
    config   *AuthConfig
}

func (b *AuthenticatedHTTPBackend) Start(ctx context.Context) error {
    // Set up HTTP server with auth middleware
    mux := http.NewServeMux()
    
    // Add auth provider routes (e.g., OIDC discovery, JWKS)
    if httpAuth, ok := b.auth.(HTTPAuthProvider); ok {
        httpAuth.Routes(mux)
    }
    
    // Add MCP endpoints with auth middleware
    mux.Handle("/mcp", b.authMiddleware(b.mcpHandler))
    
    return b.backend.StartWithMux(ctx, mux)
}
```

#### B. Protected Resource Metadata

Implement RFC 9728 protected resource metadata:

```go
// ProtectedResourceMetadata advertises authentication requirements
type ProtectedResourceMetadata struct {
    AuthorizationServers []string `json:"authorization_servers"`
    Resource            string   `json:"resource"`
    Scopes              []string `json:"scopes,omitempty"`
}

func (b *AuthenticatedHTTPBackend) protectedResourceHandler(w http.ResponseWriter, r *http.Request) {
    metadata := &ProtectedResourceMetadata{
        AuthorizationServers: []string{b.config.Issuer},
        Resource:            b.config.ResourceURI,
    }
    json.NewEncoder(w).Encode(metadata)
}
```

### 4. Database and Persistence Integration

#### A. Generic Persistence Interface

Create a generic interface for authentication persistence:

```go
type AuthPersistence interface {
    StoreClient(client *Client) error
    GetClient(clientID string) (*Client, error)
    ListClients() ([]*Client, error)
    
    StoreToken(token *Token) error
    GetToken(tokenValue string) (*Token, error)
    ValidateToken(tokenValue string) (*TokenInfo, error)
    
    StoreKey(keyID string, key []byte) error
    GetKey(keyID string) ([]byte, error)
    
    LogToolCall(entry *ToolCallLog) error
}
```

#### B. SQLite Implementation

Refactor the existing SQLite persistence:

```go
type SQLitePersistence struct {
    db   *sql.DB
    path string
}

func NewSQLitePersistence(dbPath string) (*SQLitePersistence, error) {
    // Initialize database and create tables
    // Reuse existing table schemas from idsrv
}
```

## Implementation Strategy

### Phase 1: MVP – OIDC Proxy in Front of HTTP Backend (1–3 days)

**Goals:**
- Start `idsrv` (as-is) and a small auth proxy that protects `/mcp` and forwards to the existing SSE/HTTP backend.
- Advertise RFC 9728 and proper `WWW-Authenticate`.
- Optional dev-token fallback under a `--dev` flag; default off.

**Deliverables:**
1. `pkg/authproxy/` with a minimal reverse proxy and Bearer validation via `idsrv`.
2. Embeddable helper: `WithOIDCProxy(issuer, dbPath, upstreamPort)` or `WithHooks(OnServerStart)` usage docs.
3. Example app demonstrating OIDC-protected SSE server.

**Implementation Steps:**
1. Reuse `idsrv` directly; start it on `:issuerPort`.
2. Implement reverse proxy with middleware that:
   - Validates `Authorization: Bearer` via `idsrv.Provider.IntrospectToken`.
   - Falls back to `idsrv.GetToken` only when `--dev` is set.
   - Sets `WWW-Authenticate` header on 401 with RFC 9728 metadata.
3. Start `mcp-go` SSE/HTTP server on `127.0.0.1:<upstreamPort>`.
4. Wire the proxy to forward `/mcp` to the upstream.
5. Document how to run and verify end-to-end.

### Phase 2: OIDC Provider Extraction (Week 2–3)

**Goals:**
- Refactor existing OIDC server into reusable provider
- Implement SQLite persistence interface
- Create OIDC configuration options

**Deliverables:**
1. `pkg/auth/oidc/` package with extracted OIDC provider
2. Generic persistence interfaces
3. SQLite persistence implementation
4. OIDC configuration options for embeddable framework

**Implementation Steps:**
1. Extract `idsrv` package into `pkg/auth/oidc/provider.go`
2. Implement `AuthProvider` interface for OIDC
3. Create generic persistence interfaces
4. Refactor SQLite code into reusable persistence layer
5. Add OIDC configuration options (`WithOIDCAuth`)

### Phase 3: Transport Integration (Week 3–4)

**Goals:**
- Integrate authentication into HTTP transports
- Implement protected resource metadata
- Add authentication middleware to tool calls

**Deliverables:**
1. Enhanced HTTP backends with authentication
2. RFC 9728 protected resource metadata support
3. Authentication middleware integration
4. Enhanced backend factory with auth support

**Implementation Steps:**
1. Create `AuthenticatedHTTPBackend` wrapper
2. Implement protected resource metadata endpoints
3. Add authentication middleware to tool call pipeline
4. Update backend factory to handle auth providers
5. Test end-to-end authentication flows

### Phase 4: Developer Experience & Documentation (Week 4–5)

**Goals:**
- Create comprehensive examples and documentation
- Implement CLI tools for auth management
- Add testing utilities and helpers

**Deliverables:**
1. Complete example applications
2. Developer documentation and tutorials
3. CLI extensions for auth management
4. Testing utilities and mock providers

**Implementation Steps:**
1. Create example applications showing different auth patterns
2. Write comprehensive documentation
3. Add CLI commands for auth management (`mcp auth`)
4. Create testing utilities and mock auth providers
5. Performance testing and optimization

### Phase 5: Advanced Features (Week 5–6)

**Goals:**
- Implement advanced security features
- Add support for custom claims and scopes
- Create production deployment guides

**Deliverables:**
1. Advanced security features (rate limiting, CORS)
2. Custom claims and scope validation
3. Production deployment documentation
4. Performance benchmarks and optimization

## API Design Examples

### 1. Basic Usage - Bearer Token Authentication

```go
func main() {
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My MCP application with bearer token auth",
    }

    // Simple bearer token authentication
    tokenValidator := auth.NewSimpleTokenValidator(map[string]*auth.TokenInfo{
        "secret-token": {Subject: "user1", Scopes: []string{"mcp:tools"}},
    })

    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("MyApp MCP Server"),
        embeddable.WithBearerTokenAuth(tokenValidator),
        embeddable.WithProtectedTools("sensitive-operation"),
        embeddable.WithPublicTools("health-check"),
        embeddable.WithTool("health-check", healthCheckHandler),
        embeddable.WithTool("sensitive-operation", sensitiveHandler),
    )
    if err != nil {
        log.Fatal(err)
    }

    rootCmd.Execute()
}
```

### 2. OIDC Authentication

```go
func main() {
    rootCmd := &cobra.Command{
        Use:   "enterprise-app",
        Short: "Enterprise MCP app with OIDC",
    }

    oidcConfig := &auth.OIDCConfig{
        Issuer:      "https://myapp.example.com",
        SQLiteDB:    "/data/auth.db",
        DevFallback: false, // Disable in production
        ClientDefaults: &auth.ClientConfig{
            GrantTypes:    []string{"authorization_code", "refresh_token"},
            ResponseTypes: []string{"code"},
            Scopes:        []string{"openid", "profile", "mcp:tools"},
        },
    }

    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("Enterprise MCP Server"),
        embeddable.WithOIDCAuth(oidcConfig),
        embeddable.WithDefaultTransport("sse"), // OIDC requires HTTP transport
        embeddable.WithDefaultPort(8443),
        embeddable.WithProtectedTools("*"), // All tools require auth
        embeddable.WithSessionStore(session.NewRedisSessionStore("localhost:6379")),
        // ... tool registration
    )

    rootCmd.Execute()
}
```

### 3. Custom Authentication Provider

```go
// Custom authentication provider for API keys
type APIKeyProvider struct {
    keys map[string]*UserInfo
}

func (p *APIKeyProvider) Authenticate(ctx context.Context, req *auth.AuthRequest) (*auth.AuthResult, error) {
    apiKey := req.Headers["X-API-Key"]
    if apiKey == "" {
        return &auth.AuthResult{Success: false, Error: errors.New("missing API key")}, nil
    }

    user, exists := p.keys[apiKey]
    if !exists {
        return &auth.AuthResult{Success: false, Error: errors.New("invalid API key")}, nil
    }

    return &auth.AuthResult{
        Success:  true,
        Subject:  user.ID,
        ClientID: user.ClientID,
        Scopes:   user.Scopes,
    }, nil
}

// Usage
func main() {
    apiKeyProvider := &APIKeyProvider{
        keys: map[string]*UserInfo{
            "key123": {ID: "user1", ClientID: "app1", Scopes: []string{"mcp:tools"}},
        },
    }

    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithCustomAuth(apiKeyProvider),
        // ... other options
    )
}
```

## Migration Path

### For Existing Applications

1. **No Auth → Bearer Token**: Simple migration path for basic security
2. **Bearer Token → OIDC**: Upgrade path for enterprise requirements  
3. **Custom Auth Integration**: Flexible integration with existing auth systems

### Backward Compatibility

- All auth features are opt-in via configuration options
- Existing embeddable applications continue to work unchanged
- New auth features only activate when explicitly configured

## Testing Strategy

### Unit Tests
- Authentication provider interfaces
- Token validation logic
- Middleware integration
- Configuration handling

### Integration Tests
- End-to-end OIDC flows
- Multi-transport authentication
- Error handling and edge cases
- Performance under load

### Example Test Cases
```go
func TestOIDCAuthentication(t *testing.T) {
    // Test full OIDC flow with embeddable framework
}

func TestBearerTokenMiddleware(t *testing.T) {
    // Test bearer token middleware integration
}

func TestProtectedTools(t *testing.T) {
    // Test tool protection and access control
}
```

## Security Considerations

### 1. Token Security
- Secure token storage and transmission
- Token expiration and refresh handling
- Token introspection and validation

### 2. Transport Security
- HTTPS enforcement for OIDC flows
- CORS configuration for web clients
- Rate limiting and DDoS protection

### 3. Authorization
- Scope-based access control
- Tool-level authorization
- Audit logging and monitoring

## Performance Considerations

### 1. Authentication Caching
- Token validation result caching
- Key rotation handling
- Session state management

### 2. Database Performance
- Connection pooling for SQLite
- Index optimization for token lookups
- Batch operations for logging

### 3. Scalability
- Horizontal scaling considerations
- Session store clustering
- Load balancer integration

## Future Enhancements

### 1. Advanced Authentication
- Multi-factor authentication support
- Certificate-based authentication
- SAML integration

### 2. Enhanced Authorization
- Role-based access control (RBAC)
- Attribute-based access control (ABAC)
- Dynamic policy evaluation

### 3. Monitoring and Observability
- Metrics collection and export
- Distributed tracing support
- Real-time monitoring dashboards

## Conclusion

The proposed integration strategy transforms the standalone `mcp-oidc-server` into a comprehensive, reusable authentication framework for the go-go-mcp ecosystem. This approach:

1. **Preserves Existing Investment**: Reuses proven OIDC implementation
2. **Enhances Framework Value**: Adds enterprise-grade security features
3. **Maintains Flexibility**: Supports multiple authentication mechanisms
4. **Ensures Scalability**: Designed for production deployment
5. **Improves Developer Experience**: Simple configuration and integration

The phased implementation approach allows for incremental delivery while maintaining backward compatibility and providing clear upgrade paths for existing applications.

This design positions go-go-mcp as a complete framework for building secure, production-ready MCP applications with enterprise-grade authentication and authorization capabilities.
