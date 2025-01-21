package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

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
	clients           map[string]chan *protocol.Response
	server            *http.Server
	port              int
	promptService     services.PromptService
	resourceService   services.ResourceService
	toolService       services.ToolService
	initializeService services.InitializeService
}

// NewSSEServer creates a new SSE server instance
func NewSSEServer(logger zerolog.Logger, ps services.PromptService, rs services.ResourceService, ts services.ToolService, is services.InitializeService, port int) *SSEServer {
	return &SSEServer{
		logger:            logger,
		clients:           make(map[string]chan *protocol.Response),
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
			s.logger.Debug().Str("session_id", sessionID).Msg("Closing client connection")
			close(ch)
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

		case <-ctx.Done():
			s.logger.Debug().
				Str("session_id", sessionID).
				Msg("Context done, closing connection")
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
	ctx := r.Context()

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

			result, err := s.initializeService.Initialize(ctx, params)
			if err != nil {
				response = &protocol.Response{
					JSONRPC: "2.0",
					Error: &protocol.Error{
						Code:    -32603,
						Message: "Initialize failed",
					},
				}
			} else {
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
		}

	case "ping":
		response = &protocol.Response{
			JSONRPC: "2.0",
			Result:  json.RawMessage("{}"),
		}

	case "prompts/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				response = &protocol.Response{
					JSONRPC: "2.0",
					Error: &protocol.Error{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
				break
			}
			cursor = params.Cursor
		}

		prompts, nextCursor, err := s.promptService.ListPrompts(ctx, cursor)
		if err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
				},
			}
		} else {
			result := map[string]interface{}{
				"prompts":    prompts,
				"nextCursor": nextCursor,
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

	case "prompts/get":
		var params struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
			break
		}

		message, err := s.promptService.GetPrompt(ctx, params.Name, params.Arguments)
		if err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Prompt not found",
				},
			}
		} else {
			result := map[string]interface{}{
				"description": "Prompt from provider",
				"messages":    []protocol.PromptMessage{*message},
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

	case "resources/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				response = &protocol.Response{
					JSONRPC: "2.0",
					Error: &protocol.Error{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
				break
			}
			cursor = params.Cursor
		}

		resources, nextCursor, err := s.resourceService.ListResources(ctx, cursor)
		if err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
				},
			}
		} else {
			result := map[string]interface{}{
				"resources":  resources,
				"nextCursor": nextCursor,
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

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
			break
		}

		content, err := s.resourceService.ReadResource(ctx, params.URI)
		if err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32002,
					Message: "Resource not found",
				},
			}
		} else {
			result := map[string]interface{}{
				"contents": []protocol.ResourceContent{*content},
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

	case "tools/list":
		var cursor string
		if request.Params != nil {
			var params struct {
				Cursor string `json:"cursor"`
			}
			if err := json.Unmarshal(request.Params, &params); err != nil {
				response = &protocol.Response{
					JSONRPC: "2.0",
					Error: &protocol.Error{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
				break
			}
			cursor = params.Cursor
		}

		tools, nextCursor, err := s.toolService.ListTools(ctx, cursor)
		if err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32603,
					Message: "Internal error",
				},
			}
		} else {
			result := map[string]interface{}{
				"tools":      tools,
				"nextCursor": nextCursor,
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

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
			break
		}

		result, err := s.toolService.CallTool(ctx, params.Name, params.Arguments)
		if err != nil {
			response = &protocol.Response{
				JSONRPC: "2.0",
				Error: &protocol.Error{
					Code:    -32602,
					Message: "Tool not found",
				},
			}
		} else {
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
