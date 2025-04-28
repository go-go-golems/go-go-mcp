package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	sessionIDKey contextKey = "sessionID"
)

// GetSessionID retrieves the session ID from the context.
// Returns an empty string and false if not found.
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

type SSETransport struct {
	mu         sync.RWMutex
	logger     zerolog.Logger
	clients    map[string]*SSEClient
	server     *http.Server
	port       int
	handler    transport.RequestHandler
	nextID     int
	wg         sync.WaitGroup
	cancel     context.CancelFunc
	standalone bool // indicates if we manage our own server
}

type SSEClient struct {
	id          string
	sessionID   string
	messageChan chan *protocol.Response
	createdAt   time.Time
	remoteAddr  string
	userAgent   string
}

// SSEHandlers contains the HTTP handlers for SSE endpoints
type SSEHandlers struct {
	SSEHandler     http.HandlerFunc
	MessageHandler http.HandlerFunc
}

func NewSSETransport(opts ...transport.TransportOption) (*SSETransport, error) {
	pid := os.Getpid()
	options := &transport.TransportOptions{
		MaxMessageSize: 1024 * 1024, // 1MB default
		Logger:         log.Logger.With().Int("pid", pid).Logger(),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.SSE == nil {
		return nil, fmt.Errorf("SSE options are required")
	}

	s := &SSETransport{
		logger:     options.Logger,
		clients:    make(map[string]*SSEClient),
		port:       8080, // Default port
		standalone: true, // Default to standalone mode
	}

	// If middleware is provided, we assume we're not standalone
	if len(options.SSE.Middleware) > 0 {
		s.standalone = false
	}

	// Parse the port from the address if provided
	if options.SSE != nil && options.SSE.Addr != "" {
		_, portStr, err := net.SplitHostPort(options.SSE.Addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address format: %w", err)
		}
		port, err := net.LookupPort("tcp", portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
		s.port = port
	}

	return s, nil
}

// GetHandlers returns the HTTP handlers for SSE endpoints
func (s *SSETransport) GetHandlers() *SSEHandlers {
	return &SSEHandlers{
		SSEHandler:     s.handleSSE,
		MessageHandler: s.handleMessages,
	}
}

// RegisterHandlers registers the SSE handlers with the provided router
func (s *SSETransport) RegisterHandlers(r *mux.Router) {
	r.HandleFunc("/sse", s.handleSSE).Methods("GET")
	r.HandleFunc("/messages", s.handleMessages).Methods("POST", "OPTIONS")
}

func (s *SSETransport) Listen(ctx context.Context, handler transport.RequestHandler) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.handler = handler

	if s.standalone {
		r := mux.NewRouter()
		s.RegisterHandlers(r)

		s.server = &http.Server{
			Addr:              fmt.Sprintf(":%d", s.port),
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		}

		errChan := make(chan error, 1)
		go func() {
			s.logger.Info().Int("port", s.port).Msg("Starting standalone SSE transport")
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
			close(errChan)
		}()

		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return s.Close(context.Background())
		}
	} else {
		// In non-standalone mode, we just wait for context cancellation
		s.logger.Info().Msg("Starting integrated SSE transport")
		<-ctx.Done()
		return s.Close(context.Background())
	}
}

func (s *SSETransport) Send(ctx context.Context, response *protocol.Response) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find client by session ID from context
	sessionID, ok := GetSessionID(ctx)
	if !ok {
		return fmt.Errorf("session ID not found in context")
	}

	var targetClients []*SSEClient
	for _, client := range s.clients {
		if client.sessionID == sessionID {
			targetClients = append(targetClients, client)
		}
	}

	if len(targetClients) == 0 {
		return fmt.Errorf("no clients found for session %s", sessionID)
	}

	// Send to all clients in the session
	for _, client := range targetClients {
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

	return nil
}

func (s *SSETransport) Close(ctx context.Context) error {
	s.mu.Lock()

	if s.server != nil {
		s.logger.Info().Msg("Stopping SSE transport")

		if s.cancel != nil {
			s.cancel()
		}

		for sessionID, client := range s.clients {
			s.logger.Debug().Str("sessionId", sessionID).Msg("Closing client connection")
			close(client.messageChan)
			delete(s.clients, sessionID)
		}

		s.mu.Unlock()

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

		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("error shutting down server: %w", err)
		}

		return nil
	}

	s.mu.Unlock()
	return nil
}

func (s *SSETransport) Info() transport.TransportInfo {
	return transport.TransportInfo{
		Type: "sse",
		Capabilities: map[string]bool{
			"bidirectional": true,
			"persistent":    true,
		},
		Metadata: map[string]string{
			"port": fmt.Sprintf("%d", s.port),
		},
	}
}

func (s *SSETransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	s.mu.Lock()
	s.nextID++
	clientID := fmt.Sprintf("client-%d", s.nextID)
	client := &SSEClient{
		id:          clientID,
		sessionID:   sessionID,
		messageChan: make(chan *protocol.Response, 100),
		createdAt:   time.Now(),
		remoteAddr:  r.RemoteAddr,
		userAgent:   r.UserAgent(),
	}
	log.Info().Str("client_id", clientID).
		Str("session_id", sessionID).
		Msg("New client connected")
	s.clients[clientID] = client
	s.mu.Unlock()

	s.wg.Add(1)
	defer s.wg.Done()
	defer func() {
		s.mu.Lock()
		if c, exists := s.clients[clientID]; exists {
			close(c.messageChan)
			delete(s.clients, clientID)
		}
		s.mu.Unlock()
	}()

	// Send initial endpoint event
	endpoint := fmt.Sprintf("%s?sessionId=%s", "/messages", sessionID)
	if _, err := fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpoint); err != nil {
		s.logger.Error().Err(err).Msg("Failed to write endpoint event")
		return
	}
	w.(http.Flusher).Flush()

	for {
		select {
		case msg := <-client.messageChan:
			if msg == nil {
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				s.logger.Error().Err(err).Interface("message", msg).Msg("Failed to marshal message")
				continue
			}

			if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", data); err != nil {
				s.logger.Error().Err(err).Msg("Failed to write message event")
				return
			}
			w.(http.Flusher).Flush()

		case <-ctx.Done():
			return
		}
	}
}

func (s *SSETransport) handleMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = "default"
	}

	ctx := context.WithValue(r.Context(), sessionIDKey, sessionID)

	var request protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.logger.Error().Err(err).Msg("Failed to decode request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := s.handler.HandleRequest(ctx, &request)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error handling request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if response != nil {
		if err := s.Send(ctx, response); err != nil {
			s.logger.Error().Err(err).Msg("Error sending response")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
