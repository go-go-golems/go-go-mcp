package streamable_http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// SessionCookieName is the name of the cookie used to store the session ID
	SessionCookieName = "mcp_session_id"

	// Default timeouts
	writeTimeout = 10 * time.Second
	readTimeout  = 60 * time.Second
	pingInterval = 30 * time.Second
)

// StreamableHTTPTransport implements bidirectional streaming over HTTP using WebSockets
type StreamableHTTPTransport struct {
	mu           sync.RWMutex
	logger       zerolog.Logger
	clients      map[string]*StreamClient              // Map clientID -> client
	sessions     map[session.SessionID][]*StreamClient // Map sessionID -> list of clients
	server       *http.Server
	port         int
	handler      transport.RequestHandler
	sessionStore session.SessionStore
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	standalone   bool // indicates if we manage our own server
	upgrader     websocket.Upgrader
}

// StreamClient represents a connected streaming client
type StreamClient struct {
	id          string
	sessionID   session.SessionID
	conn        *websocket.Conn
	messageChan chan *protocol.Response
	createdAt   time.Time
	remoteAddr  string
	userAgent   string
	logger      zerolog.Logger
	cancel      context.CancelFunc
}

// StreamableHTTPHandlers contains the HTTP handlers for streamable HTTP endpoints
type StreamableHTTPHandlers struct {
	StreamHandler  http.HandlerFunc
	MessageHandler http.HandlerFunc
}

func NewStreamableHTTPTransport(opts ...transport.TransportOption) (*StreamableHTTPTransport, error) {
	pid := os.Getpid()
	options := &transport.TransportOptions{
		MaxMessageSize: 1024 * 1024, // 1MB default
		Logger:         log.Logger.With().Int("pid", pid).Logger(),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.StreamableHTTP == nil {
		return nil, fmt.Errorf("StreamableHTTP options are required")
	}

	s := &StreamableHTTPTransport{
		logger:     options.Logger.With().Str("component", "streamable_http_transport").Logger(),
		clients:    make(map[string]*StreamClient),
		sessions:   make(map[session.SessionID][]*StreamClient),
		port:       8080, // Default port
		standalone: true, // Default to standalone mode
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking for security
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	// If middleware is provided, we assume we're not standalone
	if len(options.StreamableHTTP.Middleware) > 0 {
		s.standalone = false
	}

	// Parse the port from the address if provided
	if options.StreamableHTTP != nil && options.StreamableHTTP.Addr != "" {
		_, portStr, err := net.SplitHostPort(options.StreamableHTTP.Addr)
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

// GetHandlers returns the HTTP handlers for streamable HTTP endpoints
func (s *StreamableHTTPTransport) GetHandlers() *StreamableHTTPHandlers {
	return &StreamableHTTPHandlers{
		StreamHandler:  s.handleStream,
		MessageHandler: s.handleMessages,
	}
}

// RegisterHandlers registers the streamable HTTP handlers with the provided router
func (s *StreamableHTTPTransport) RegisterHandlers(r *mux.Router) {
	r.HandleFunc("/stream", s.handleStream).Methods("GET")
	r.HandleFunc("/messages", s.handleMessages).Methods("POST", "OPTIONS")
}

func (s *StreamableHTTPTransport) Listen(ctx context.Context, handler transport.RequestHandler) error {
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
			s.logger.Info().Int("port", s.port).Msg("Starting standalone streamable HTTP transport")
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
		s.logger.Info().Msg("Starting integrated streamable HTTP transport")
		<-ctx.Done()
		return s.Close(context.Background())
	}
}

func (s *StreamableHTTPTransport) Send(ctx context.Context, response *protocol.Response) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find client by session ID from context
	session_, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return fmt.Errorf("session ID not found in context")
	}

	var targetClients []*StreamClient
	for _, client := range s.clients {
		if client.sessionID == session_.ID {
			targetClients = append(targetClients, client)
		}
	}

	if len(targetClients) == 0 {
		return fmt.Errorf("no clients found for session %s", session_.ID)
	}

	// Send to all clients in the session
	for _, client := range targetClients {
		select {
		case client.messageChan <- response:
			client.logger.Debug().
				Str("sessionId", string(session_.ID)).
				Interface("response", response).
				Msg("Response sent to client")
		default:
			client.logger.Error().
				Str("sessionId", string(session_.ID)).
				Msg("Failed to send response to client")
		}
	}

	return nil
}

func (s *StreamableHTTPTransport) Close(ctx context.Context) error {
	s.logger.Info().Msg("Closing streamable HTTP transport")

	if s.cancel != nil {
		s.cancel()
	}

	// Close all client connections
	s.mu.Lock()
	for _, client := range s.clients {
		client.cancel()
		if err := client.conn.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Error closing client connection")
		}
	}
	s.clients = make(map[string]*StreamClient)
	s.sessions = make(map[session.SessionID][]*StreamClient)
	s.mu.Unlock()

	// Close the HTTP server if we manage it
	if s.server != nil && s.standalone {
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error().Err(err).Msg("Error shutting down server")
			return err
		}
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info().Msg("All goroutines finished")
	case <-time.After(10 * time.Second):
		s.logger.Warn().Msg("Timeout waiting for goroutines to finish")
	}

	return nil
}

func (s *StreamableHTTPTransport) Info() transport.TransportInfo {
	return transport.TransportInfo{
		Type:       "streamable_http",
		RemoteAddr: fmt.Sprintf(":%d", s.port),
		Capabilities: map[string]bool{
			"bidirectional": true,
			"streaming":     true,
			"websocket":     true,
		},
		Metadata: map[string]string{
			"protocol": "websocket",
			"version":  "mcp-2025-03-26",
		},
	}
}

func (s *StreamableHTTPTransport) SetSessionStore(store session.SessionStore) {
	s.sessionStore = store
}

// handleStream handles WebSocket connections for bidirectional streaming
func (s *StreamableHTTPTransport) handleStream(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Error closing WebSocket connection")
		}
	}()

	clientID := uuid.New().String()
	sessionID := s.getOrCreateSession(r)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	client := &StreamClient{
		id:          clientID,
		sessionID:   sessionID,
		conn:        conn,
		messageChan: make(chan *protocol.Response, 100),
		createdAt:   time.Now(),
		remoteAddr:  r.RemoteAddr,
		userAgent:   r.UserAgent(),
		logger:      s.logger.With().Str("clientId", clientID).Str("sessionId", string(sessionID)).Logger(),
		cancel:      cancel,
	}

	// Register the client
	s.mu.Lock()
	s.clients[clientID] = client
	s.sessions[sessionID] = append(s.sessions[sessionID], client)
	s.mu.Unlock()

	client.logger.Info().Msg("StreamableHTTP client connected")

	// Start goroutines for handling the connection
	s.wg.Add(2)
	go s.handleClientReader(ctx, client)
	go s.handleClientWriter(ctx, client)

	// Wait for context cancellation
	<-ctx.Done()

	// Cleanup
	s.mu.Lock()
	delete(s.clients, clientID)
	// Remove client from session
	sessionClients := s.sessions[sessionID]
	for i, c := range sessionClients {
		if c.id == clientID {
			s.sessions[sessionID] = append(sessionClients[:i], sessionClients[i+1:]...)
			break
		}
	}
	// Remove empty session
	if len(s.sessions[sessionID]) == 0 {
		delete(s.sessions, sessionID)
	}
	s.mu.Unlock()

	client.logger.Info().Msg("StreamableHTTP client disconnected")
}

// handleClientReader reads messages from the WebSocket connection
func (s *StreamableHTTPTransport) handleClientReader(ctx context.Context, client *StreamClient) {
	defer s.wg.Done()
	defer client.cancel()

	if err := client.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		client.logger.Error().Err(err).Msg("Error setting read deadline")
	}
	client.conn.SetPongHandler(func(string) error {
		if err := client.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			client.logger.Error().Err(err).Msg("Error setting read deadline in pong handler")
		}
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				client.logger.Error().Err(err).Msg("WebSocket read error")
			}
			return
		}

		// Parse the JSON-RPC message
		var rawMessage json.RawMessage
		if err := json.Unmarshal(message, &rawMessage); err != nil {
			client.logger.Error().Err(err).Msg("Failed to parse message")
			continue
		}

		// Handle the message in the session context
		sessionCtx := session.WithSession(ctx, &session.Session{ID: client.sessionID})

		// Check if it's a request or notification
		var baseMessage struct {
			ID     json.RawMessage `json:"id,omitempty"`
			Method string          `json:"method"`
		}

		if err := json.Unmarshal(message, &baseMessage); err != nil {
			client.logger.Error().Err(err).Msg("Failed to parse base message")
			continue
		}

		if baseMessage.ID != nil {
			// It's a request
			var req protocol.Request
			if err := json.Unmarshal(message, &req); err != nil {
				client.logger.Error().Err(err).Msg("Failed to parse request")
				continue
			}

			go func() {
				resp, err := s.handler.HandleRequest(sessionCtx, &req)
				if err != nil {
					client.logger.Error().Err(err).Msg("Error handling request")
					// Send error response
					errorResp := &protocol.Response{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error: &protocol.Error{
							Code:    -32603,
							Message: err.Error(),
						},
					}
					select {
					case client.messageChan <- errorResp:
					default:
						client.logger.Error().Msg("Failed to send error response")
					}
					return
				}

				if resp != nil {
					select {
					case client.messageChan <- resp:
					default:
						client.logger.Error().Msg("Failed to send response")
					}
				}
			}()
		} else {
			// It's a notification
			var notif protocol.Notification
			if err := json.Unmarshal(message, &notif); err != nil {
				client.logger.Error().Err(err).Msg("Failed to parse notification")
				continue
			}

			go func() {
				if err := s.handler.HandleNotification(sessionCtx, &notif); err != nil {
					client.logger.Error().Err(err).Msg("Error handling notification")
				}
			}()
		}
	}
}

// handleClientWriter writes messages to the WebSocket connection
func (s *StreamableHTTPTransport) handleClientWriter(ctx context.Context, client *StreamClient) {
	defer s.wg.Done()
	defer client.cancel()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case response := <-client.messageChan:
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				client.logger.Error().Err(err).Msg("Error setting write deadline")
			}

			responseBytes, err := json.Marshal(response)
			if err != nil {
				client.logger.Error().Err(err).Msg("Failed to marshal response")
				continue
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				client.logger.Error().Err(err).Msg("Failed to write message")
				return
			}

		case <-ticker.C:
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				client.logger.Error().Err(err).Msg("Error setting write deadline for ping")
			}
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				client.logger.Error().Err(err).Msg("Failed to send ping")
				return
			}
		}
	}
}

// handleMessages handles HTTP POST requests (fallback for non-WebSocket clients)
func (s *StreamableHTTPTransport) handleMessages(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	sessionID := s.getOrCreateSession(r)
	sessionCtx := session.WithSession(r.Context(), &session.Session{ID: sessionID})

	// Parse and handle the message similar to WebSocket handler
	var baseMessage struct {
		ID     json.RawMessage `json:"id,omitempty"`
		Method string          `json:"method"`
	}

	if err := json.Unmarshal(body, &baseMessage); err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse message")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if baseMessage.ID != nil {
		// It's a request
		var req protocol.Request
		if err := json.Unmarshal(body, &req); err != nil {
			s.logger.Error().Err(err).Msg("Failed to parse request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		resp, err := s.handler.HandleRequest(sessionCtx, &req)
		if err != nil {
			s.logger.Error().Err(err).Msg("Error handling request")
			errorResp := &protocol.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &protocol.Error{
					Code:    -32603,
					Message: err.Error(),
				},
			}
			if err := json.NewEncoder(w).Encode(errorResp); err != nil {
				s.logger.Error().Err(err).Msg("Error encoding error response")
			}
			return
		}

		if resp != nil {
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				s.logger.Error().Err(err).Msg("Error encoding response")
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		// It's a notification
		var notif protocol.Notification
		if err := json.Unmarshal(body, &notif); err != nil {
			s.logger.Error().Err(err).Msg("Failed to parse notification")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := s.handler.HandleNotification(sessionCtx, &notif); err != nil {
			s.logger.Error().Err(err).Msg("Error handling notification")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *StreamableHTTPTransport) getOrCreateSession(r *http.Request) session.SessionID {
	// Try to get session from cookie
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if cookie.Value != "" {
			return session.SessionID(cookie.Value)
		}
	}

	// Try to get session from header
	if sessionHeader := r.Header.Get("X-MCP-Session-ID"); sessionHeader != "" {
		return session.SessionID(sessionHeader)
	}

	// Create new session
	newSessionID := session.SessionID(uuid.New().String())
	s.logger.Debug().Str("sessionId", string(newSessionID)).Msg("Created new session")
	return newSessionID
}
