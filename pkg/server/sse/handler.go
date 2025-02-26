package sse

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg/events"
	"github.com/rs/zerolog"
)

const (
	// SSE Headers
	headerContentType   = "Content-Type"
	headerCacheControl  = "Cache-Control"
	headerConnection    = "Connection"
	headerAccessControl = "Access-Control-Allow-Origin"

	// SSE Header Values
	contentTypeSSE      = "text/event-stream"
	cacheControlNoCache = "no-cache"
	connectionKeepAlive = "keep-alive"
	accessControlAll    = "*"
)

// connection represents an active SSE connection
type connection struct {
	pageID string
	w      http.ResponseWriter
	done   chan struct{}
}

// SSEHandler handles Server-Sent Events connections
type SSEHandler struct {
	events events.EventManager
	logger *zerolog.Logger
	mu     sync.RWMutex
	conns  map[string][]*connection
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(events events.EventManager, logger *zerolog.Logger) *SSEHandler {
	return &SSEHandler{
		events: events,
		logger: logger,
		conns:  make(map[string][]*connection),
	}
}

// ServeHTTP implements http.Handler
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set(headerContentType, contentTypeSSE)
	w.Header().Set(headerCacheControl, cacheControlNoCache)
	w.Header().Set(headerConnection, connectionKeepAlive)
	w.Header().Set(headerAccessControl, accessControlAll)

	// Get page ID from query
	pageID := r.URL.Query().Get("page")
	if pageID == "" {
		http.Error(w, "page parameter is required", http.StatusBadRequest)
		return
	}

	// Create connection
	conn := &connection{
		pageID: pageID,
		w:      w,
		done:   make(chan struct{}),
	}

	// Register connection
	h.registerConnection(conn)
	defer h.unregisterConnection(conn)

	// Handle connection
	h.handleConnection(r.Context(), conn)
}

// registerConnection adds a connection to the handler
func (h *SSEHandler) registerConnection(conn *connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.conns[conn.pageID] = append(h.conns[conn.pageID], conn)
	h.logger.Debug().
		Str("pageID", conn.pageID).
		Int("total_connections", len(h.conns[conn.pageID])).
		Msg("New SSE connection registered")
}

// unregisterConnection removes a connection from the handler
func (h *SSEHandler) unregisterConnection(conn *connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	close(conn.done)

	conns := h.conns[conn.pageID]
	for i, c := range conns {
		if c == conn {
			// Remove connection from slice
			h.conns[conn.pageID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	// Remove page entry if no more connections
	if len(h.conns[conn.pageID]) == 0 {
		delete(h.conns, conn.pageID)
	}

	h.logger.Debug().
		Str("pageID", conn.pageID).
		Int("total_connections", len(h.conns[conn.pageID])).
		Msg("SSE connection unregistered")
}

// handleConnection manages a single SSE connection
func (h *SSEHandler) handleConnection(ctx context.Context, conn *connection) {
	events, err := h.events.Subscribe(ctx, conn.pageID)
	if err != nil {
		h.logger.Error().Err(err).
			Str("pageID", conn.pageID).
			Msg("Failed to subscribe to events")
		return
	}

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := h.sendEvent(conn, event); err != nil {
				h.logger.Error().Err(err).
					Str("pageID", conn.pageID).
					Msg("Failed to send event")
				return
			}
		case <-ctx.Done():
			return
		case <-conn.done:
			return
		}
	}
}

// sendEvent sends an event to a connection
func (h *SSEHandler) sendEvent(conn *connection, event events.UIEvent) error {
	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", event.Type, data)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if f, ok := conn.w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}
