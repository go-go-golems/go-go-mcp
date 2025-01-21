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

// handleSSE handles new SSE connections
func (s *SSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
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
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, sessionID)
		close(messageChan)
		s.mu.Unlock()
	}()

	// Send initial endpoint event
	endpointMsg := fmt.Sprintf("data: {\"endpoint\": \"/messages?session_id=%s\"}\n\n", sessionID)
	fmt.Fprint(w, endpointMsg)
	w.(http.Flusher).Flush()

	// Keep connection open and send messages
	for {
		select {
		case msg := <-messageChan:
			if msg == nil {
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal message")
				continue
			}

			// Send message event
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()

		case <-r.Context().Done():
			return
		}
	}
}

// handleMessages processes incoming client messages
func (s *SSEServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	messageChan, exists := s.clients[sessionID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	var request protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Process the request based on method
	var response *protocol.Response

	switch request.Method {
	case "initialize":
		var params protocol.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		} else {
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

			response = &protocol.Response{
				JSONRPC: "2.0",
				Result:  mustMarshal(result),
			}
		}

	// Add other method handlers here
	default:
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
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "failed to send response", http.StatusInternalServerError)
	}
}
