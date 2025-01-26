package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/server/dispatcher"
	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// SSEServer handles SSE transport for MCP protocol
type SSEServer struct {
	mu           sync.RWMutex
	logger       zerolog.Logger
	registry     *pkg.ProviderRegistry
	clients      map[string]*SSEClient
	server       *http.Server
	port         int
	dispatcher   *dispatcher.Dispatcher
	nextClientID int
	wg           sync.WaitGroup
	cancel       context.CancelFunc
}

type SSEClient struct {
	id          string
	sessionID   string
	messageChan chan *protocol.Response
	createdAt   time.Time
	remoteAddr  string
	userAgent   string
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

// NewSSEServer creates a new SSE server instance
func NewSSEServer(logger zerolog.Logger, ps services.PromptService, rs services.ResourceService, ts services.ToolService, is services.InitializeService, port int) *SSEServer {
	return &SSEServer{
		logger:     logger,
		clients:    make(map[string]*SSEClient),
		port:       port,
		dispatcher: dispatcher.NewDispatcher(logger, ps, rs, ts, is),
	}
}

// Start begins the SSE server
func (s *SSEServer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	r := mux.NewRouter()

	// SSE endpoint for clients to establish connection
	r.HandleFunc("/sse", s.handleSSE).Methods("GET")

	// POST endpoint for receiving client messages
	r.HandleFunc("/messages", s.handleMessages).Methods("POST")

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: r,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// Create a channel to capture server errors
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info().Int("port", s.port).Msg("Starting SSE server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop(context.Background())
	}
}

// Stop gracefully stops the SSE server
func (s *SSEServer) Stop(ctx context.Context) error {
	s.mu.Lock()

	if s.server != nil {
		s.logger.Info().Msg("Stopping SSE server")

		// Cancel all client goroutines
		if s.cancel != nil {
			s.cancel()
		}

		// Close all client connections
		for sessionID, client := range s.clients {
			s.logger.Debug().Str("sessionId", sessionID).Msg("Closing client connection")
			close(client.messageChan)
			delete(s.clients, sessionID)
		}

		s.mu.Unlock()

		// Wait for all client goroutines to finish with a timeout
		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			s.logger.Debug().Msg("All client goroutines finished")
		case <-ctx.Done():
			s.logger.Warn().Msg("Timeout waiting for client goroutines")
		}

		// Shutdown the HTTP server
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("error shutting down server: %w", err)
		}

		return nil
	}

	s.mu.Unlock()
	return nil
}

// marshalJSON marshals data to JSON and returns any error
func (s *SSEServer) marshalJSON(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		s.logger.Error().Err(err).Interface("value", v).Msg("Failed to marshal JSON")
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// withSessionID adds a session ID to the context
func withSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// handleSSE handles new SSE connections
func (s *SSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set SSE headers according to protocol spec
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Create unique session ID
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s", uuid.New())
	}

	s.mu.Lock()
	s.nextClientID++
	clientID := fmt.Sprintf("client-%d", s.nextClientID)
	client := &SSEClient{
		id:          clientID,
		sessionID:   sessionID,
		messageChan: make(chan *protocol.Response, 100),
		createdAt:   time.Now(),
		remoteAddr:  r.RemoteAddr,
		userAgent:   r.UserAgent(),
	}
	s.clients[clientID] = client
	clientCount := len(s.clients)
	s.mu.Unlock()

	s.logger.Debug().
		Str("client_id", clientID).
		Str("sessionId", sessionID).
		Str("remote_addr", r.RemoteAddr).
		Str("user_agent", r.UserAgent()).
		Int("total_clients", clientCount).
		Msg("New SSE connection")

	// Send initial endpoint event with session ID
	endpoint := fmt.Sprintf("%s?sessionId=%s", "/messages", sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpoint)
	w.(http.Flusher).Flush()

	// Add to waitgroup before starting goroutine
	s.wg.Add(1)
	defer s.wg.Done()

	defer func() {
		s.mu.Lock()
		if c, exists := s.clients[clientID]; exists {
			close(c.messageChan)
			delete(s.clients, clientID)
			s.logger.Debug().
				Str("client_id", clientID).
				Str("sessionId", sessionID).
				Int("total_clients", len(s.clients)).
				Dur("connection_duration", time.Since(c.createdAt)).
				Msg("Client disconnected")
		}
		s.mu.Unlock()
	}()

	// Keep connection open and send messages
	for {
		select {
		case msg := <-client.messageChan:
			if msg == nil {
				s.logger.Debug().
					Str("client_id", clientID).
					Str("sessionId", sessionID).
					Msg("Received nil message, closing connection")
				return
			}

			data, err := s.marshalJSON(msg)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("client_id", clientID).
					Str("sessionId", sessionID).
					Interface("message", msg).
					Msg("Failed to marshal message")
				continue
			}

			s.logger.Debug().
				Str("client_id", clientID).
				Str("sessionId", sessionID).
				RawJSON("message", data).
				Msg("Sending message to client")

			// Send message event according to protocol spec
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()

		case <-ctx.Done():
			s.logger.Debug().
				Str("client_id", clientID).
				Str("sessionId", sessionID).
				Msg("Context done, closing connection")
			return
		}
	}
}

// handleMessages processes incoming client messages
func (s *SSEServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := r.URL.Query().Get("sessionId")
	s.logger.Debug().
		Str("url", r.URL.String()).
		Str("remote_addr", r.RemoteAddr).
		Str("sessionId", sessionID).
		Msg("Received message request")

	// Use default session if none provided
	if sessionID == "" {
		sessionID = "default"
		s.logger.Debug().Msg("Using default session")
	}

	// Add sessionId to context
	ctx = dispatcher.WithSessionID(ctx, sessionID)

	// Find all clients for this session
	s.mu.RLock()
	var sessionClients []*SSEClient
	for _, client := range s.clients {
		if client.sessionID == sessionID {
			sessionClients = append(sessionClients, client)
		}
	}
	s.mu.RUnlock()

	if len(sessionClients) == 0 {
		s.logger.Error().
			Str("sessionId", sessionID).
			Msg("No active clients found for session")
		http.Error(w, "No active clients found for session", http.StatusBadRequest)
		return
	}

	var request protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response, _ := dispatcher.NewErrorResponse(nil, -32700, "Parse error", err)
		// Send error to all session clients
		for _, client := range sessionClients {
			select {
			case client.messageChan <- response:
			default:
				s.logger.Error().
					Str("client_id", client.id).
					Str("sessionId", sessionID).
					Msg("Failed to send error response to client")
			}
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// Use the dispatcher to handle the request
	response, err := s.dispatcher.Dispatch(ctx, request)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error dispatching request")
		response, _ = dispatcher.NewErrorResponse(request.ID, -32603, "Internal error", err)
	}

	// Send response to all session clients if it's not nil (notifications don't have responses)
	if response != nil {
		for _, client := range sessionClients {
			select {
			case client.messageChan <- response:
				s.logger.Debug().
					Str("client_id", client.id).
					Str("sessionId", sessionID).
					Interface("response", response).
					Msg("Response sent to client")
			default:
				s.logger.Error().
					Str("client_id", client.id).
					Str("sessionId", sessionID).
					Msg("Failed to send response to client")
			}
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
