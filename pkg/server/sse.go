package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/services"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// SSEServer handles SSE transport for MCP protocol
type SSEServer struct {
	mu                sync.RWMutex
	logger            zerolog.Logger
	registry          *pkg.ProviderRegistry
	clients           map[string]*SSEClient
	server            *http.Server
	port              int
	promptService     services.PromptService
	resourceService   services.ResourceService
	toolService       services.ToolService
	initializeService services.InitializeService
	nextClientID      int
}

type SSEClient struct {
	id          string
	sessionID   string
	messageChan chan *protocol.Response
	createdAt   time.Time
	remoteAddr  string
	userAgent   string
}

// NewSSEServer creates a new SSE server instance
func NewSSEServer(logger zerolog.Logger, ps services.PromptService, rs services.ResourceService, ts services.ToolService, is services.InitializeService, port int) *SSEServer {
	return &SSEServer{
		logger:            logger,
		clients:           make(map[string]*SSEClient),
		port:              port,
		promptService:     ps,
		resourceService:   rs,
		toolService:       ts,
		initializeService: is,
	}
}

// Start begins the SSE server
func (s *SSEServer) Start(ctx context.Context) error {
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
		return s.Stop(ctx)
	}
}

// Stop gracefully stops the SSE server
func (s *SSEServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.logger.Info().Msg("Stopping SSE server")
		// Close all client connections
		for sessionID, ch := range s.clients {
			s.logger.Debug().Str("sessionId", sessionID).Msg("Closing client connection")
			close(ch.messageChan)
			delete(s.clients, sessionID)
		}
		return s.server.Shutdown(ctx)
	}
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
		s.logger.Error().
			Err(err).
			Str("sessionId", sessionID).
			Msg("Invalid request body")
		response := &protocol.Response{
			JSONRPC: "2.0",
			Error: &protocol.Error{
				Code:    -32700,
				Message: "Parse error",
			},
		}
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

	if request.JSONRPC != "2.0" {
		data, _ := json.Marshal("Invalid JSON-RPC version")
		response := &protocol.Response{
			JSONRPC: "2.0",
			Error: &protocol.Error{
				Code:    -32600,
				Message: "Invalid Request",
				Data:    json.RawMessage(data),
			},
		}
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

	s.logger.Debug().
		Str("sessionId", sessionID).
		Str("method", request.Method).
		Interface("params", request.Params).
		Msg("Processing request")

	// Process the request based on method
	var response *protocol.Response

	switch request.Method {
	case "initialize":
		var params protocol.InitializeParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		} else {
			result, err := s.initializeService.Initialize(ctx, params)
			if err != nil {
				data, _ := json.Marshal(err.Error())
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Internal error",
						Data:    json.RawMessage(data),
					},
				}
			} else {
				resultJSON, err := s.marshalJSON(result)
				if err != nil {
					data, _ := json.Marshal(err.Error())
					response = &protocol.Response{
						JSONRPC: "2.0",
						ID:      request.ID,
						Error: &protocol.Error{
							Code:    -32603,
							Message: "Internal error",
							Data:    json.RawMessage(data),
						},
					}
				} else {
					response = &protocol.Response{
						JSONRPC: "2.0",
						ID:      request.ID,
						Result:  resultJSON,
					}
				}
			}
		}

	case "ping":
		response = &protocol.Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result:  json.RawMessage("{}"),
		}

	case "prompts/list", "resources/list", "tools/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
				break
			}
			cursor = params.Cursor
		}

		var result interface{}
		var nextCursor string
		var err error

		switch request.Method {
		case "prompts/list":
			result, nextCursor, err = s.promptService.ListPrompts(ctx, cursor)
		case "resources/list":
			result, nextCursor, err = s.resourceService.ListResources(ctx, cursor)
		case "tools/list":
			result, nextCursor, err = s.toolService.ListTools(ctx, cursor)
		}

		if err != nil {
			data, _ := json.Marshal(err.Error())
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
					Data:    json.RawMessage(data),
				},
			}
		} else {
			// Ensure we have an empty array instead of null for empty results
			if result == nil {
				switch request.Method {
				case "prompts/list":
					result = []interface{}{}
				case "resources/list":
					result = []interface{}{}
				case "tools/list":
					result = []interface{}{}
				}
			}

			resultMap := map[string]interface{}{
				strings.TrimSuffix(request.Method, "/list"): result,
				"nextCursor": nextCursor,
			}
			resultJSON, err := s.marshalJSON(resultMap)
			if err != nil {
				data, _ := json.Marshal(err.Error())
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Internal error",
						Data:    json.RawMessage(data),
					},
				}
			} else {
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Result:  resultJSON,
				}
			}
		}

	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
			break
		}

		message, err := s.promptService.GetPrompt(ctx, params.Name, params.Arguments)
		if err != nil {
			data, _ := json.Marshal(err.Error())
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
					Data:    json.RawMessage(data),
				},
			}
		} else {
			result := map[string]interface{}{
				"description": "Prompt from provider",
				"messages":    []protocol.PromptMessage{*message},
			}
			resultJSON, err := s.marshalJSON(result)
			if err != nil {
				data, _ := json.Marshal(err.Error())
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Internal error",
						Data:    json.RawMessage(data),
					},
				}
			} else {
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Result:  resultJSON,
				}
			}
		}

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
			break
		}

		content, err := s.resourceService.ReadResource(ctx, params.URI)
		if err != nil {
			data, _ := json.Marshal(err.Error())
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
					Data:    json.RawMessage(data),
				},
			}
		} else {
			result := map[string]interface{}{
				"contents": []protocol.ResourceContent{*content},
			}
			resultJSON, err := s.marshalJSON(result)
			if err != nil {
				data, _ := json.Marshal(err.Error())
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Internal error",
						Data:    json.RawMessage(data),
					},
				}
			} else {
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Result:  resultJSON,
				}
			}
		}

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
			break
		}

		result, err := s.toolService.CallTool(ctx, params.Name, params.Arguments)
		if err != nil {
			data, _ := json.Marshal(err.Error())
			response = &protocol.Response{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
					Data:    json.RawMessage(data),
				},
			}
		} else {
			resultJSON, err := s.marshalJSON(result)
			if err != nil {
				data, _ := json.Marshal(err.Error())
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Internal error",
						Data:    json.RawMessage(data),
					},
				}
			} else {
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Result:  resultJSON,
				}
			}
		}

	default:
		s.logger.Warn().
			Str("sessionId", sessionID).
			Str("method", request.Method).
			Msg("Method not found")
		response = &protocol.Response{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &protocol.Error{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}

	// Send response to all session clients
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

	w.WriteHeader(http.StatusAccepted)
}
