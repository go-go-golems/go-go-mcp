package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// SSEServer handles SSE transport for MCP protocol
type SSEServer struct {
	mu       sync.RWMutex
	logger   zerolog.Logger
	registry *pkg.ProviderRegistry
	clients  map[string]chan *protocol.Response
}

// NewSSEServer creates a new SSE server instance
func NewSSEServer(logger zerolog.Logger, registry *pkg.ProviderRegistry) *SSEServer {
	return &SSEServer{
		logger:   logger,
		registry: registry,
		clients:  make(map[string]chan *protocol.Response),
	}
}

// Start begins the SSE server on the specified port
func (s *SSEServer) Start(port int) error {
	r := mux.NewRouter()

	// SSE endpoint for clients to establish connection
	r.HandleFunc("/sse", s.handleSSE).Methods("GET")

	// POST endpoint for receiving client messages
	r.HandleFunc("/messages", s.handleMessages).Methods("POST")

	s.logger.Info().Int("port", port).Msg("Starting SSE server")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
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

// handleSSE handles new SSE connections
func (s *SSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug().
		Str("remote_addr", r.RemoteAddr).
		Str("user_agent", r.UserAgent()).
		Msg("New SSE connection")

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create unique session ID
	sessionID := uuid.New().String()
	messageChan := make(chan *protocol.Response, 100)

	// Register client
	s.mu.Lock()
	s.clients[sessionID] = messageChan
	clientCount := len(s.clients)
	s.mu.Unlock()

	s.logger.Debug().
		Str("session_id", sessionID).
		Int("total_clients", clientCount).
		Msg("Client registered")

	defer func() {
		s.mu.Lock()
		delete(s.clients, sessionID)
		close(messageChan)
		s.logger.Debug().
			Str("session_id", sessionID).
			Int("total_clients", len(s.clients)).
			Msg("Client disconnected")
		s.mu.Unlock()
	}()

	// Send initial endpoint event
	endpointData, err := s.marshalJSON(map[string]string{
		"endpoint": fmt.Sprintf("/messages?session_id=%s", sessionID),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal endpoint data")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	endpointMsg := fmt.Sprintf("data: %s\n\n", endpointData)
	s.logger.Debug().
		Str("session_id", sessionID).
		Str("endpoint", fmt.Sprintf("/messages?session_id=%s", sessionID)).
		Msg("Sending endpoint information")
	fmt.Fprint(w, endpointMsg)
	w.(http.Flusher).Flush()

	// Keep connection open and send messages
	for {
		select {
		case msg := <-messageChan:
			if msg == nil {
				s.logger.Debug().
					Str("session_id", sessionID).
					Msg("Received nil message, closing connection")
				return
			}

			data, err := s.marshalJSON(msg)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("session_id", sessionID).
					Interface("message", msg).
					Msg("Failed to marshal message")
				continue
			}

			s.logger.Debug().
				Str("session_id", sessionID).
				RawJSON("message", data).
				Msg("Sending message to client")

			// Send message event
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()

		case <-r.Context().Done():
			s.logger.Debug().
				Str("session_id", sessionID).
				Msg("Client context done, closing connection")
			return
		}
	}
}

// handleMessages processes incoming client messages
func (s *SSEServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	s.logger.Debug().
		Str("session_id", sessionID).
		Str("remote_addr", r.RemoteAddr).
		Msg("Received message request")

	if sessionID == "" {
		s.logger.Error().Msg("Missing session_id in request")
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	messageChan, exists := s.clients[sessionID]
	s.mu.RUnlock()

	if !exists {
		s.logger.Error().
			Str("session_id", sessionID).
			Msg("Session not found")
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	var request protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	s.logger.Debug().
		Str("session_id", sessionID).
		Str("method", request.Method).
		Interface("params", request.Params).
		Msg("Processing request")

	// Process the request based on method
	var response *protocol.Response

	switch request.Method {
	case "initialize":
		var params protocol.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			s.logger.Error().
				Err(err).
				Str("session_id", sessionID).
				RawJSON("params", request.Params).
				Msg("Invalid initialize params")
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		} else {
			s.logger.Debug().
				Str("session_id", sessionID).
				Str("protocol_version", params.ProtocolVersion).
				Interface("capabilities", params.Capabilities).
				Msg("Processing initialize request")

			// Handle initialization
			result := protocol.InitializeResult{
				ProtocolVersion: params.ProtocolVersion,
				Capabilities: protocol.ServerCapabilities{
					Logging: &protocol.LoggingCapability{},
					Prompts: &protocol.PromptsCapability{
						ListChanged: true,
					},
					Resources: &protocol.ResourcesCapability{
						Subscribe:   true,
						ListChanged: true,
					},
					Tools: &protocol.ToolsCapability{
						ListChanged: true,
					},
				},
				ServerInfo: protocol.ServerInfo{
					Name:    "go-mcp-server",
					Version: "1.0.0",
				},
			}

			resultJSON, err := s.marshalJSON(result)
			if err != nil {
				response = &protocol.Response{
					JSONRPC: "2.0",
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Internal error",
					},
				}
			} else {
				response = &protocol.Response{
					JSONRPC: "2.0",
					Result:  resultJSON,
				}
			}
		}

	default:
		s.logger.Warn().
			Str("session_id", sessionID).
			Str("method", request.Method).
			Msg("Method not found")
		response = &protocol.Response{
			JSONRPC: "2.0",
			Error: &protocol.Error{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}

	if request.ID != nil {
		response.ID = request.ID
	}

	// Send response through the client's message channel
	select {
	case messageChan <- response:
		s.logger.Debug().
			Str("session_id", sessionID).
			Interface("response", response).
			Msg("Response sent to client")
		w.WriteHeader(http.StatusAccepted)
	default:
		s.logger.Error().
			Str("session_id", sessionID).
			Msg("Failed to send response to client")
		http.Error(w, "failed to send response", http.StatusInternalServerError)
	}
}
