package dispatcher

import (
	"context"
	"encoding/json"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// Dispatch handles an incoming request and returns a response
func (d *Dispatcher) Dispatch(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	d.logger.Debug().
		Str("method", request.Method).
		RawJSON("params", request.Params).
		Msg("Dispatching request")

	// Validate JSON-RPC version
	if request.JSONRPC != "2.0" {
		return NewErrorResponse(request.ID, -32600, "Invalid Request", "invalid JSON-RPC version")
	}

	// Handle requests vs notifications based on ID presence
	if len(request.ID) == 0 {
		return d.handleNotification(ctx, request)
	}

	switch request.Method {
	case "initialize":
		return d.handleInitialize(ctx, request)
	case "ping":
		return d.handlePing(ctx, request)
	case "prompts/list":
		return d.handlePromptsList(ctx, request)
	case "prompts/get":
		return d.handlePromptsGet(ctx, request)
	case "resources/list":
		return d.handleResourcesList(ctx, request)
	case "resources/read":
		return d.handleResourcesRead(ctx, request)
	case "tools/list":
		return d.handleToolsList(ctx, request)
	case "tools/call":
		return d.handleToolsCall(ctx, request)
	default:
		return NewErrorResponse(request.ID, -32601, "Method not found", nil)
	}
}

func (d *Dispatcher) handleNotification(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	switch request.Method {
	case "notifications/initialized":
		d.logger.Info().Msg("Client initialized")
		return nil, nil
	default:
		d.logger.Warn().Str("method", request.Method).Msg("Unknown notification method")
		return nil, nil
	}
}

func (d *Dispatcher) handleInitialize(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var params protocol.InitializeParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return NewErrorResponse(request.ID, -32602, "Invalid params", err)
	}

	result, err := d.initializeService.Initialize(ctx, params)
	if err != nil {
		return NewErrorResponse(request.ID, -32603, "Initialize failed", err)
	}

	return NewSuccessResponse(request.ID, result)
}

func (d *Dispatcher) handlePing(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	return NewSuccessResponse(request.ID, struct{}{})
}

func (d *Dispatcher) handlePromptsList(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var cursor string
	if request.Params != nil {
		var params struct {
			Cursor string `json:"cursor"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return NewErrorResponse(request.ID, -32602, "Invalid params", err)
		}
		cursor = params.Cursor
	}

	prompts, nextCursor, err := d.promptService.ListPrompts(ctx, cursor)
	if err != nil {
		return NewErrorResponse(request.ID, -32603, "Internal error", err)
	}
	if prompts == nil {
		prompts = []protocol.Prompt{}
	}

	return NewSuccessResponse(request.ID, ListPromptsResult{
		Prompts:    prompts,
		NextCursor: nextCursor,
	})
}

func (d *Dispatcher) handlePromptsGet(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return NewErrorResponse(request.ID, -32602, "Invalid params", err)
	}

	message, err := d.promptService.GetPrompt(ctx, params.Name, params.Arguments)
	if err != nil {
		return NewErrorResponse(request.ID, -32602, "Prompt not found", err)
	}

	return NewSuccessResponse(request.ID, PromptResult{
		Description: "Prompt from provider",
		Messages:    []protocol.PromptMessage{*message},
	})
}

func (d *Dispatcher) handleResourcesList(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var cursor string
	if request.Params != nil {
		var params struct {
			Cursor string `json:"cursor"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return NewErrorResponse(request.ID, -32602, "Invalid params", err)
		}
		cursor = params.Cursor
	}

	resources, nextCursor, err := d.resourceService.ListResources(ctx, cursor)
	if err != nil {
		return NewErrorResponse(request.ID, -32603, "Internal error", err)
	}
	if resources == nil {
		resources = []protocol.Resource{}
	}

	return NewSuccessResponse(request.ID, ListResourcesResult{
		Resources:  resources,
		NextCursor: nextCursor,
	})
}

func (d *Dispatcher) handleResourcesRead(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return NewErrorResponse(request.ID, -32602, "Invalid params", err)
	}

	content, err := d.resourceService.ReadResource(ctx, params.URI)
	if err != nil {
		return NewErrorResponse(request.ID, -32002, "Resource not found", err)
	}

	return NewSuccessResponse(request.ID, ResourceResult{
		Contents: []protocol.ResourceContent{*content},
	})
}

func (d *Dispatcher) handleToolsList(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var cursor string
	if request.Params != nil {
		var params struct {
			Cursor string `json:"cursor"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return NewErrorResponse(request.ID, -32602, "Invalid params", err)
		}
		cursor = params.Cursor
	}

	tools, nextCursor, err := d.toolService.ListTools(ctx, cursor)
	if err != nil {
		return NewErrorResponse(request.ID, -32603, "Internal error", err)
	}
	if tools == nil {
		tools = []protocol.Tool{}
	}

	return NewSuccessResponse(request.ID, ListToolsResult{
		Tools:      tools,
		NextCursor: nextCursor,
	})
}

func (d *Dispatcher) handleToolsCall(ctx context.Context, request protocol.Request) (*protocol.Response, error) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return NewErrorResponse(request.ID, -32602, "Invalid params", err)
	}

	result, err := d.toolService.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		return NewErrorResponse(request.ID, -32602, "Tool not found", err)
	}

	return NewSuccessResponse(request.ID, result)
}
