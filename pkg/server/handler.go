package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
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
	default:
		h.server.logger.Warn().Str("method", notif.Method).Msg("Unknown notification method")
		return nil
	}
}

// Helper method to create success response
func (h *RequestHandler) newSuccessResponse(id json.RawMessage, result interface{}) (*protocol.Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &protocol.Response{
		ID:     id,
		Result: resultJSON,
	}, nil
}

// Individual request handlers
func (h *RequestHandler) handleInitialize(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
	}

	// Validate protocol version
	supportedVersions := []string{"2024-11-05"}
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

	return h.newSuccessResponse(req.ID, result)
}

func (h *RequestHandler) handlePing(_ context.Context, req *protocol.Request) (*protocol.Response, error) {
	return h.newSuccessResponse(req.ID, struct{}{})
}

func (h *RequestHandler) handlePromptsList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
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

	prompt, err := h.server.promptProvider.GetPrompt(ctx, params.Name, params.Arguments)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	return h.newSuccessResponse(req.ID, prompt)
}

func (h *RequestHandler) handleResourcesList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
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
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
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
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, transport.NewInvalidParamsError(err.Error())
	}

	result, err := h.server.toolProvider.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		return nil, transport.NewInternalError(err.Error())
	}

	return h.newSuccessResponse(req.ID, result)
}
