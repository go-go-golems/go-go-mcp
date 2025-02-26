package sse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// ComponentRenderer is a function that renders a component to HTML
type ComponentRenderer func(componentID string, data interface{}) (string, error)

// connection represents an active SSE connection
type connection struct {
	pageID string
	w      http.ResponseWriter
	done   chan struct{}
}

// SSEHandler handles Server-Sent Events connections
type SSEHandler struct {
	events    events.EventManager
	logger    *zerolog.Logger
	mu        sync.RWMutex
	conns     map[string][]*connection
	renderers map[string]ComponentRenderer // Map of event types to renderer functions
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(events events.EventManager, logger *zerolog.Logger) *SSEHandler {
	return &SSEHandler{
		events:    events,
		logger:    logger,
		conns:     make(map[string][]*connection),
		renderers: make(map[string]ComponentRenderer),
	}
}

// RegisterRenderer registers a component renderer for a specific event type
func (h *SSEHandler) RegisterRenderer(eventType string, renderer ComponentRenderer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.renderers[eventType] = renderer
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

	// Send initial ping to establish connection
	if err := h.sendPing(conn); err != nil {
		h.logger.Error().Err(err).
			Str("pageID", conn.pageID).
			Msg("Failed to send initial ping")
		return
	}

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := h.processEvent(conn, event); err != nil {
				h.logger.Error().Err(err).
					Str("pageID", conn.pageID).
					Msg("Failed to process event")
				return
			}
		case <-ctx.Done():
			h.logger.Debug().
				Str("pageID", conn.pageID).
				Msg("Context done, closing connection")
			return
		case <-conn.done:
			h.logger.Debug().
				Str("pageID", conn.pageID).
				Msg("Connection done, closing")
			return
		}
	}
}

// sendPing sends a ping event to keep the connection alive
func (h *SSEHandler) sendPing(conn *connection) error {
	_, err := fmt.Fprintf(conn.w, "event: ping\ndata: {}\n\n")
	if err != nil {
		return fmt.Errorf("failed to write ping: %w", err)
	}

	if f, ok := conn.w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// processEvent processes a UIEvent and sends it to the client
func (h *SSEHandler) processEvent(conn *connection, event events.UIEvent) error {
	h.logger.Debug().
		Str("pageID", conn.pageID).
		Str("eventType", event.Type).
		Msg("Processing event")

	switch event.Type {
	case "page-reload":
		return h.handlePageReload(conn, event)
	case "refresh-ui":
		return h.handleRefreshUI(conn, event)
	default:
		// For unknown event types, just send the raw event
		return h.sendEvent(conn, event.Type, event)
	}
}

// handlePageReload handles a page reload event
func (h *SSEHandler) handlePageReload(conn *connection, event events.UIEvent) error {
	// First, send a page-reload event notification
	if err := h.sendRawEvent(conn, "page-reload", "{}"); err != nil {
		return fmt.Errorf("failed to send page-reload event: %w", err)
	}

	// Then, check if we have a page renderer to render the full page
	h.mu.RLock()
	pageRenderer, ok := h.renderers["page-content-template"]
	h.mu.RUnlock()

	if !ok {
		h.logger.Debug().
			Str("eventType", "page-content-template").
			Msg("No page renderer registered, skipping full page update")
		return nil
	}

	// Render the page content template with the page ID
	// Pass the event data which might contain updated page definition
	html, err := pageRenderer(conn.pageID, event.Component)
	if err != nil {
		return fmt.Errorf("failed to render page content template: %w", err)
	}

	// Send the rendered page as a component-update event
	return h.sendRawEvent(conn, "page-content-update", html)
}

// handleRefreshUI handles a refresh-ui event
func (h *SSEHandler) handleRefreshUI(conn *connection, event events.UIEvent) error {
	// First, send a refresh-ui event notification
	if err := h.sendRawEvent(conn, "refresh-ui", "{}"); err != nil {
		return fmt.Errorf("failed to send refresh-ui event: %w", err)
	}

	// Then, check if we have a page renderer to render the full page
	h.mu.RLock()
	pageRenderer, ok := h.renderers["page-content-template"]
	h.mu.RUnlock()

	if !ok {
		h.logger.Debug().
			Str("eventType", "page-content-template").
			Msg("No page renderer registered, skipping UI update")
		return nil
	}

	// Render the page content template with the page ID and event data
	html, err := pageRenderer(conn.pageID, event.Component)
	if err != nil {
		return fmt.Errorf("failed to render page content template: %w", err)
	}

	// Send the rendered page as a ui-update event
	return h.sendRawEvent(conn, "ui-update", html)
}

// sendEvent sends an event to a connection with JSON data
func (h *SSEHandler) sendEvent(conn *connection, eventName string, data interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err := fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", eventName, buf.String())
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if f, ok := conn.w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// sendRawEvent sends an event with raw data (no JSON encoding)
func (h *SSEHandler) sendRawEvent(conn *connection, eventName string, data string) error {
	// Replace newlines with HTML entities to prevent breaking SSE protocol
	// This is safe because we're sending HTML content that will be rendered in the browser
	safeData := strings.ReplaceAll(data, "\n", "&#10;")

	_, err := fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", eventName, safeData)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if f, ok := conn.w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}
