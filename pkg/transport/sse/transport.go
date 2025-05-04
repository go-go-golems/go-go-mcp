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
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// XXX: Deprecate this in favor of pkg/session.GetSessionFromContext
	// sessionIDKey contextKey = "sessionID"

	// SessionCookieName is the name of the cookie used to store the session ID
	SessionCookieName = "mcp_session_id"
)

// // GetSessionID retrieves the session ID from the context.
// // Returns an empty string and false if not found.
// func GetSessionID(ctx context.Context) (string, bool) {
// 	sessionID, ok := ctx.Value(sessionIDKey).(string)
// 	return sessionID, ok
// }

type SSETransport struct {
	mu           sync.RWMutex
	logger       zerolog.Logger
	clients      map[string]*SSEClient              // Map clientID -> client
	sessions     map[session.SessionID][]*SSEClient // Map sessionID -> list of clients
	server       *http.Server
	port         int
	handler      transport.RequestHandler
	sessionStore session.SessionStore // Added session store
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	standalone   bool // indicates if we manage our own server
}

type SSEClient struct {
	id          string
	sessionID   session.SessionID
	messageChan chan *protocol.Response
	createdAt   time.Time
	remoteAddr  string
	userAgent   string
	logger      zerolog.Logger
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
		logger:     options.Logger.With().Str("component", "sse_transport").Logger(),
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

	// Initialize sessions map
	s.sessions = make(map[session.SessionID][]*SSEClient)

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
	session_, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return fmt.Errorf("session ID not found in context")
	}

	var targetClients []*SSEClient
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

func (s *SSETransport) Close(ctx context.Context) error {
	s.mu.Lock()

	if s.server != nil {
		s.logger.Info().Msg("Stopping SSE transport")

		if s.cancel != nil {
			s.cancel()
		}

		for sessionID, clients := range s.sessions {
			s.logger.Debug().Str("sessionId", string(sessionID)).Msg("Closing client connections")
			for _, client := range clients {
				close(client.messageChan)
				delete(s.clients, client.id)
			}
			delete(s.sessions, sessionID)
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
	s.mu.Lock()
	if s.handler == nil {
		s.mu.Unlock()
		s.logger.Error().Msg("Handler not set")
		http.Error(w, "Server not ready", http.StatusInternalServerError)
		return
	}
	if s.sessionStore == nil {
		s.mu.Unlock()
		s.logger.Error().Msg("Session store not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}
	s.mu.Unlock() // Unlock early, lock specific sections below

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Get or create session
	ctx := r.Context()
	currentSession, sessionID := s.getSessionFromRequest(r)
	if currentSession == nil {
		// Should not happen if getSessionFromRequest works correctly, but handle defensively
		currentSession = s.sessionStore.Create()
		sessionID = currentSession.ID
		s.logger.Info().Str("session_id", string(sessionID)).Msg("Created new session (edge case)")
	}

	// Set session cookie
	s.setSessionCookie(w, sessionID)

	// Create a new client for this connection
	clientID := uuid.New().String()
	client := &SSEClient{
		id:          clientID,
		sessionID:   sessionID,
		messageChan: make(chan *protocol.Response, 100),
		createdAt:   time.Now(),
		remoteAddr:  r.RemoteAddr,
		userAgent:   r.UserAgent(),
		logger: s.logger.With().
			Str("client_id", clientID).
			Str("session_id", string(sessionID)).
			Str("remote_addr", r.RemoteAddr).
			Logger(),
	}

	s.mu.Lock()
	s.clients[clientID] = client
	s.sessions[sessionID] = append(s.sessions[sessionID], client)
	s.mu.Unlock()

	client.logger.Info().Msg("Client connected")

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow CORS for debugging

	s.wg.Add(1)
	defer s.wg.Done()
	defer s.removeClient(clientID)
	defer close(client.messageChan)

	for {
		select {
		case <-ctx.Done():
			client.logger.Info().Msg("Client disconnected (context cancelled)")
			return
		case <-r.Context().Done():
			client.logger.Info().Msg("Client disconnected (request context done)")
			return
		case msg, ok := <-client.messageChan:
			if !ok {
				client.logger.Info().Msg("Client message channel closed")
				return // Channel closed
			}

			jsonData, err := json.Marshal(msg)
			if err != nil {
				client.logger.Error().Err(err).Msg("Failed to marshal response")
				continue
			}

			// SSE format: data: <json string>\n\n
			_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
			if err != nil {
				client.logger.Error().Err(err).Msg("Failed to write response")
				continue
			}
			flusher.Flush() // Ensure data is sent immediately
			client.logger.Debug().RawJSON("data", jsonData).Msg("Sent message via SSE")
		}
	}
}

func (s *SSETransport) handleMessages(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	if s.handler == nil {
		s.mu.RUnlock()
		s.logger.Error().Msg("Handler not set")
		http.Error(w, "Server not ready", http.StatusInternalServerError)
		return
	}
	if s.sessionStore == nil {
		s.mu.RUnlock()
		s.logger.Error().Msg("Session store not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}
	s.mu.RUnlock() // Unlock early

	// Handle CORS preflight requests
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Add headers you expect
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow CORS for actual requests

	// Get session and enhance context
	currentSession, sessionID := s.getSessionFromRequest(r)
	if currentSession == nil {
		// This path might occur if a /messages POST happens before /sse GET
		currentSession = s.sessionStore.Create()
		sessionID = currentSession.ID
		s.logger.Info().Str("session_id", string(sessionID)).Msg("Created new session via /messages")
		// Important: Need to set the cookie in the response here too!
		s.setSessionCookie(w, sessionID)
	}
	ctx := session.WithSession(r.Context(), currentSession)

	// Create a scoped logger for this request
	reqLogger := s.logger.With().Str("session_id", string(sessionID)).Logger()

	var request protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		reqLogger.Error().Err(err).Msg("Failed to decode request")
		// Send JSON-RPC error response for parse error
		errResp := transport.NewParseError(fmt.Sprintf("Failed to decode request: %v", err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Use 400 for parse errors
		err = json.NewEncoder(w).Encode(errResp)
		if err != nil {
			reqLogger.Error().Err(err).Msg("Failed to encode error response")
		}
		return
	}

	reqLogger = reqLogger.With().Str("method", request.Method).Logger()
	reqLogger.Info().RawJSON("params", request.Params).Msg("Received message via POST")

	// Handle notifications (no response expected, but acknowledge receipt)
	if transport.IsNotification(&request) {
		notif := &protocol.Notification{
			JSONRPC: request.JSONRPC,
			Method:  request.Method,
			Params:  request.Params,
		}
		err := s.handler.HandleNotification(ctx, notif)
		if err != nil {
			reqLogger.Error().Err(err).Msg("Error handling notification")
			// Do not send error response for notifications per JSON-RPC spec
		}
		w.WriteHeader(http.StatusNoContent) // Acknowledge receipt
		return
	}

	// Handle regular requests
	response, err := s.handler.HandleRequest(ctx, &request)
	if err != nil {
		reqLogger.Error().Err(err).Msg("Error handling request")
		// Convert handler error to JSON-RPC error
		jsonRPCError := transport.ProcessError(err)
		response = &protocol.Response{
			JSONRPC: "2.0",
			ID:      request.ID, // Echo the original request ID
			Error:   jsonRPCError,
		}
	}

	if response != nil {
		w.Header().Set("Content-Type", "application/json")
		if response.Error != nil {
			// Determine appropriate HTTP status code from JSON-RPC error code
			statusCode := transport.ErrorToHTTPStatus(response.Error.Code)
			w.WriteHeader(statusCode)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			reqLogger.Error().Err(err).Msg("Failed to encode response")
			// Don't try to write again if encoding fails
			return
		}
		reqLogger.Debug().
			RawJSON("response", response.Result).
			Interface("error", response.Error).
			Msg("Sent response via POST")
	} else {
		// Should not happen for non-notifications, but handle defensively
		reqLogger.Warn().Msg("Handler returned nil response for non-notification request")
		w.WriteHeader(http.StatusNoContent)
	}
}

// getSessionFromRequest retrieves the session ID from the cookie or creates a new session.
// It ensures the session exists in the store.
func (s *SSETransport) getSessionFromRequest(r *http.Request) (*session.Session, session.SessionID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sessionStore == nil {
		s.logger.Error().Msg("Session store is nil in getSessionFromRequest")
		// Return a transient session? This indicates a severe configuration error.
		// For now, let's create one on the fly, but log heavily.
		panic("Session store not initialized in SSE transport") // Or handle more gracefully
	}

	cookie, err := r.Cookie(SessionCookieName)
	var sessionID string
	var currentSession *session.Session
	foundInStore := false

	if err == nil && cookie.Value != "" {
		sessionID = cookie.Value
		currentSession, foundInStore = s.sessionStore.Get(session.SessionID(sessionID))
		sessionLogger := s.logger.With().Str("session_id", sessionID).Logger()
		if foundInStore {
			sessionLogger.Debug().Msg("Retrieved existing session from store")
		} else {
			sessionLogger.Warn().Msg("Session ID from cookie not found in store, creating new session")
		}
	}

	// If no valid cookie, or session not found in store, create a new one
	if !foundInStore {
		currentSession = s.sessionStore.Create()
		sessionID = string(currentSession.ID)
		s.logger.Info().Str("session_id", sessionID).Msg("Created new session")
	}

	return currentSession, session.SessionID(sessionID)
}

// setSessionCookie sets the session ID cookie in the HTTP response.
func (s *SSETransport) setSessionCookie(w http.ResponseWriter, sessionID session.SessionID) {
	cookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    string(sessionID),
		Path:     "/", // Cookie is valid for all paths
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Consider Strict if applicable
		// Secure: true, // Enable this if served over HTTPS
		// MaxAge: 3600 * 24 * 7, // Example: 1 week expiry
	}
	http.SetCookie(w, &cookie)
}

func (s *SSETransport) removeClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[clientID]
	if !ok {
		return // Client already removed
	}

	sessionID := client.sessionID
	delete(s.clients, clientID)

	// Remove client from session list
	if sessionClients, sessionExists := s.sessions[sessionID]; sessionExists {
		newSessionClients := make([]*SSEClient, 0, len(sessionClients)-1)
		for _, c := range sessionClients {
			if c.id != clientID {
				newSessionClients = append(newSessionClients, c)
			}
		}
		// If the session is now empty, remove it
		if len(newSessionClients) == 0 {
			delete(s.sessions, sessionID)
			client.logger.Info().Msg("Session closed (last client disconnected)")
		} else {
			s.sessions[sessionID] = newSessionClients
		}
	}

	client.logger.Info().Msg("Client removed")
}

// SetSessionStore implements the transport.Transport interface
func (s *SSETransport) SetSessionStore(store session.SessionStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionStore = store
	s.logger.Info().Msg("Session store set for SSE transport")
}
