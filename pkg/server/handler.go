package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
)

// RequestHandler handles incoming MCP protocol requests
type RequestHandler struct {
	server *Server
}

func NewRequestHandler(s *Server) *RequestHandler {
	return &RequestHandler{
		server: s,
	}
}

// HandleRequest processes a request and returns a response
func (h *RequestHandler) HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		return nil, transport.NewInvalidRequestError("invalid JSON-RPC version")
	}

	if strings.HasPrefix(req.Method, "notifications/") {
		err := h.HandleNotification(ctx, &protocol.Notification{
			JSONRPC: req.JSONRPC,
			Method:  req.Method,
			Params:  req.Params,
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, req)
	case "ping":
		return h.handlePing(ctx, req)
	case "prompts/list":
		return h.handlePromptsList(ctx, req)
	case "prompts/get":
		return h.handlePromptsGet(ctx, req)
	case "resources/list":
		return h.handleResourcesList(ctx, req)
	case "resources/read":
		return h.handleResourcesRead(ctx, req)
	case "tools/list":
		return h.handleToolsList(ctx, req)
	case "tools/call":
		return h.handleToolsCall(ctx, req)
	default:
		return nil, transport.NewMethodNotFoundError(fmt.Sprintf("method %s not found", req.Method))
	}
}

// HandleNotification processes a notification (no response expected)
func (h *RequestHandler) HandleNotification(ctx context.Context, notif *protocol.Notification) error {
	switch notif.Method {
	case "notifications/initialized":
		h.server.logger.Info().Msg("Client initialized")
		return nil
	case "notifications/cancelled":
		return h.handleCancellation(ctx, notif)
	default:
		h.server.logger.Warn().Str("method", notif.Method).Msg("Unknown notification method")
		return nil
	}
}

// HandleBatchRequest processes a batch of requests and returns batch responses
func (h *RequestHandler) HandleBatchRequest(ctx context.Context, batch protocol.BatchRequest) (protocol.BatchResponse, error) {
	h.server.logger.Debug().Int("batch_size", len(batch)).Msg("Handling batch request")

	if err := batch.Validate(); err != nil {
		return nil, transport.NewInvalidRequestError(fmt.Sprintf("invalid batch request: %s", err.Error()))
	}

	responses := make(protocol.BatchResponse, 0, len(batch))

	for i, request := range batch {
		reqLogger := h.server.logger.With().
			Int("batch_index", i).
			Str("method", request.Method).
			Logger()

		// Handle individual request
		if !transport.IsNotification(&request) {
			// This is a request that expects a response
			response, err := h.HandleRequest(ctx, &request)
			if err != nil {
				reqLogger.Error().Err(err).Msg("Error handling batch request")
				// Convert error to JSON-RPC error response
				jsonErr := transport.ProcessError(err)
				response = &protocol.Response{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &protocol.Error{
						Code:    jsonErr.Code,
						Message: jsonErr.Message,
						Data:    jsonErr.Data,
					},
				}
			}
			if response != nil {
				responses = append(responses, *response)
			}
		} else {
			// This is a notification - handle it but don't add to responses
			notif := protocol.Notification{
				JSONRPC: request.JSONRPC,
				Method:  request.Method,
				Params:  request.Params,
			}
			if err := h.HandleNotification(ctx, &notif); err != nil {
				reqLogger.Error().Err(err).Msg("Error handling batch notification")
			}
		}
	}

	return responses, nil
}

// handleCancellation processes cancellation notifications
func (h *RequestHandler) handleCancellation(ctx context.Context, notif *protocol.Notification) error {
	var params protocol.CancellationParams
	if err := json.Unmarshal(notif.Params, &params); err != nil {
		h.server.logger.Error().Err(err).Msg("Failed to unmarshal cancellation parameters")
		return nil // Don't return error for notifications
	}

	h.server.logger.Info().
		Str("request_id", params.RequestID).
		Str("reason", params.Reason).
		Msg("Request cancellation received")

	// TODO: Implement actual cancellation logic
	// This would involve:
	// 1. Finding the in-progress request by ID
	// 2. Cancelling its context
	// 3. Cleaning up resources
	// For now, we just log the cancellation request

	return nil
}

// Helper method to create success response
func (h *RequestHandler) newSuccessResponse(id json.RawMessage, result interface{}) (*protocol.Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}, nil
}

// Individual request handlers
func (h *RequestHandler) handleInitialize(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
	}

	// Log session ID if present
	currentSession, sessionFound := session.GetSessionFromContext(ctx)
	if sessionFound {
		h.server.logger.Info().Str("session_id", string(currentSession.ID)).Msg("Handling initialize request for session")
		// Optionally clear session state upon initialize, depending on desired behavior
		// currentSession.State = make(session.SessionState)
		// h.server.sessionStore.Update(currentSession) // Persist the cleared state
	} else {
		h.server.logger.Warn().Msg("Handling initialize request without session context (should not happen with stdio/sse)")
	}

	// Validate protocol version
	supportedVersions := []string{"2024-11-05", "2025-03-26"}
	isSupported := false
	for _, version := range supportedVersions {
		if params.ProtocolVersion == version {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return nil, fmt.Errorf("unsupported protocol version %s, supported versions: %v", params.ProtocolVersion, supportedVersions)
	}

	// Return server capabilities
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
			Name:    h.server.serverName,
			Version: h.server.serverVersion,
		},
	}

	// Add session ID to the result if available
	// XXX there is no session ID in the 2024 spec
	// if sessionFound {
	// 	result.SessionID = currentSession.ID
	// }

	return h.newSuccessResponse(req.ID, result)
}

func (h *RequestHandler) handlePing(_ context.Context, req *protocol.Request) (*protocol.Response, error) {
	return h.newSuccessResponse(req.ID, struct{}{})
}

func (h *RequestHandler) handlePromptsList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Cursor string `json:"cursor,omitempty"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, transport.NewInvalidParamsError(err.Error())
		}
	}

	if h.server.promptProvider == nil {
		return h.newSuccessResponse(req.ID, protocol.ListPromptsResult{
			Prompts:    []protocol.Prompt{},
			NextCursor: "",
		})
	}

	prompts, nextCursor, err := h.server.promptProvider.ListPrompts(ctx, params.Cursor)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	if prompts == nil {
		prompts = []protocol.Prompt{}
	}

	return h.newSuccessResponse(req.ID, protocol.ListPromptsResult{
		Prompts:    prompts,
		NextCursor: nextCursor,
	})
}

func (h *RequestHandler) handlePromptsGet(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
	}

	if h.server.promptProvider == nil {
		return nil, transport.NewInternalError("prompt provider not configured")
	}

	prompt, err := h.server.promptProvider.GetPrompt(ctx, params.Name, params.Arguments)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	return h.newSuccessResponse(req.ID, prompt)
}

func (h *RequestHandler) handleResourcesList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Cursor string `json:"cursor,omitempty"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, transport.NewInvalidParamsError(err.Error())
		}
	}

	if h.server.resourceProvider == nil {
		return h.newSuccessResponse(req.ID, protocol.ListResourcesResult{
			Resources:  []protocol.Resource{},
			NextCursor: "",
		})
	}

	resources, nextCursor, err := h.server.resourceProvider.ListResources(ctx, params.Cursor)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	if resources == nil {
		resources = []protocol.Resource{}
	}

	return h.newSuccessResponse(req.ID, protocol.ListResourcesResult{
		Resources:  resources,
		NextCursor: nextCursor,
	})
}

func (h *RequestHandler) handleResourcesRead(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
	}

	if h.server.resourceProvider == nil {
		return nil, transport.NewInternalError("resource provider not configured")
	}

	contents, err := h.server.resourceProvider.ReadResource(ctx, params.Name)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	return h.newSuccessResponse(req.ID, protocol.ResourceResult{
		Contents: contents,
	})
}

func (h *RequestHandler) handleToolsList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Cursor string `json:"cursor,omitempty"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, transport.NewInvalidParamsError(err.Error())
		}
	}

	if h.server.toolProvider == nil {
		return h.newSuccessResponse(req.ID, protocol.ListToolsResult{
			Tools:      []protocol.Tool{},
			NextCursor: "",
		})
	}

	tools, nextCursor, err := h.server.toolProvider.ListTools(ctx, params.Cursor)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	if tools == nil {
		tools = []protocol.Tool{}
	}

	return h.newSuccessResponse(req.ID, protocol.ListToolsResult{
		Tools:      tools,
		NextCursor: nextCursor,
	})
}

func (h *RequestHandler) handleToolsCall(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	h.server.logger.Info().Str("method", req.Method).Str("params", string(req.Params)).Msg("handleToolsCall")

	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.server.logger.Error().Err(err).Str("params", string(req.Params)).Msg("failed to unmarshal tool call arguments")
		return nil, transport.NewInvalidParamsError(err.Error())
	}

	h.server.logger.Info().Str("name", params.Name).Str("arguments", fmt.Sprintf("%v", params.Arguments)).Msg("calling tool")

	result, err := h.server.toolProvider.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		h.server.logger.Error().Err(err).Str("name", params.Name).Str("arguments", fmt.Sprintf("%v", params.Arguments)).Msg("failed to call tool")
		return nil, transport.NewInternalError(err.Error())
	}

	return h.newSuccessResponse(req.ID, result)
}
