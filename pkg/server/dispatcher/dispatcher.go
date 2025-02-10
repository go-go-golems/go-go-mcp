package dispatcher

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/rs/zerolog"
)

// Dispatcher handles the MCP protocol methods and dispatches them to appropriate services
type Dispatcher struct {
	logger            zerolog.Logger
	promptService     services.PromptService
	resourceService   services.ResourceService
	toolService       services.ToolService
	initializeService services.InitializeService
}

// NewDispatcher creates a new dispatcher instance
func NewDispatcher(
	logger zerolog.Logger,
	ps services.PromptService,
	rs services.ResourceService,
	ts services.ToolService,
	is services.InitializeService,
) *Dispatcher {
	return &Dispatcher{
		logger:            logger,
		promptService:     ps,
		resourceService:   rs,
		toolService:       ts,
		initializeService: is,
	}
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey struct{}

var (
	// sessionIDKey is the key used to store the session ID in context
	sessionIDKey = contextKey{}
)

// GetSessionID retrieves the session ID from the context
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// MustGetSessionID retrieves the session ID from the context, panicking if not found
func MustGetSessionID(ctx context.Context) string {
	sessionID, ok := GetSessionID(ctx)
	if !ok {
		panic("sessionId not found in context")
	}
	return sessionID
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}
