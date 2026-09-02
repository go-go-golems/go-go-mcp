package embeddable

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

// ToolAuthorizationPolicy is the single source of truth for a tool's OAuth
// advertisement, decoded-dispatch enforcement, and insufficient-scope challenge.
type ToolAuthorizationPolicy struct {
	Public         bool
	RequiredScopes []string
}

// WithToolAuthorization assigns an authorization policy to a registered tool.
// Public tools must not also declare required scopes.
func WithToolAuthorization(toolName string, policy ToolAuthorizationPolicy) ServerOption {
	return func(config *ServerConfig) error {
		toolName = strings.TrimSpace(toolName)
		if toolName == "" {
			return fmt.Errorf("tool authorization requires a tool name")
		}
		normalized, err := normalizeToolAuthorizationPolicy(policy)
		if err != nil {
			return fmt.Errorf("tool %q authorization: %w", toolName, err)
		}
		if config.toolAuthorization == nil {
			config.toolAuthorization = map[string]ToolAuthorizationPolicy{}
		}
		config.toolAuthorization[toolName] = normalized
		return nil
	}
}

func normalizeToolAuthorizationPolicy(policy ToolAuthorizationPolicy) (ToolAuthorizationPolicy, error) {
	seen := map[string]struct{}{}
	scopes := make([]string, 0, len(policy.RequiredScopes))
	for _, raw := range policy.RequiredScopes {
		scope := strings.TrimSpace(raw)
		if scope == "" {
			continue
		}
		if strings.ContainsAny(scope, " \t\r\n\"") {
			return ToolAuthorizationPolicy{}, fmt.Errorf("invalid scope %q", scope)
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	if policy.Public && len(scopes) > 0 {
		return ToolAuthorizationPolicy{}, fmt.Errorf("public policy cannot require scopes")
	}
	if !policy.Public && len(scopes) == 0 {
		return ToolAuthorizationPolicy{}, fmt.Errorf("protected policy requires at least one scope")
	}
	return ToolAuthorizationPolicy{Public: policy.Public, RequiredScopes: scopes}, nil
}

func applyToolSecurityMetadata(tool *protocol.Tool, policy ToolAuthorizationPolicy) {
	if policy.Public {
		return
	}
	if tool.Meta == nil {
		tool.Meta = map[string]any{}
	}
	tool.Meta["securitySchemes"] = []any{map[string]any{
		"type":   "oauth2",
		"scopes": append([]string(nil), policy.RequiredScopes...),
	}}
}

func authorizeToolRequest(tokenInfo *mcpauth.TokenInfo, policy ToolAuthorizationPolicy, resourceMetadataURL string) *protocol.ToolResult {
	if policy.Public {
		return nil
	}
	available := map[string]struct{}{}
	if tokenInfo != nil {
		for _, scope := range tokenInfo.Scopes {
			available[scope] = struct{}{}
		}
	}
	for _, required := range policy.RequiredScopes {
		if _, ok := available[required]; !ok {
			challenge := buildToolBearerChallenge(resourceMetadataURL, policy.RequiredScopes)
			result := protocol.NewErrorToolResult(
				protocol.NewTextContent("authorization required: missing scope " + required),
			)
			result.Meta = map[string]any{"mcp/www_authenticate": []string{challenge}}
			return result
		}
	}
	return nil
}

func buildToolBearerChallenge(resourceMetadataURL string, scopes []string) string {
	parts := make([]string, 0, 3)
	if resourceMetadataURL != "" {
		parts = append(parts, fmt.Sprintf("resource_metadata=%q", resourceMetadataURL))
	}
	parts = append(parts, `error="insufficient_scope"`)
	if len(scopes) > 0 {
		parts = append(parts, fmt.Sprintf("scope=%q", strings.Join(scopes, " ")))
	}
	return "Bearer " + strings.Join(parts, ", ")
}
