# Technical Plan: Implementing Wait-for-Response in UI Update Endpoint

## Overview

The goal is to enhance the `/api/ui-update` endpoint to wait for a corresponding user action (submit or click) before completing the request. This creates a synchronous flow between UI updates and user interactions. Additionally, we'll refactor the UI handlers into their own struct to improve code organization.

## Key Components

1. A wait registry to track pending requests
2. UUID generation for request tracking
3. Timeout handling for requests without responses
4. A dedicated UI handler struct for better code organization
5. Frontend modifications to pass request IDs with actions

## Detailed Implementation Plan

### 1. Create a UIHandler Struct

- [x] Define a new `UIHandler` struct to encapsulate all UI-related functionality
- [x] Move UI-specific methods from the Server struct to the UIHandler
- [x] Create a clean interface between Server and UIHandler

```go
// UIHandler manages all UI-related functionality
type UIHandler struct {
    waitRegistry *WaitRegistry
    events       events.EventManager
    sseHandler   *sse.SSEHandler
    logger       *zerolog.Logger
    
    // Configuration
    timeout      time.Duration
}

// NewUIHandler creates a new UI handler with the given dependencies
func NewUIHandler(events events.EventManager, sseHandler *sse.SSEHandler, logger *zerolog.Logger) *UIHandler {
    return &UIHandler{
        waitRegistry: NewWaitRegistry(30 * time.Second), // 30 second default timeout
        events:       events,
        sseHandler:   sseHandler,
        logger:       logger,
        timeout:      30 * time.Second,
    }
}

// RegisterHandlers registers all UI-related HTTP handlers with the given mux
func (h *UIHandler) RegisterHandlers(mux *http.ServeMux) {
    mux.Handle("/api/ui-update", h.handleUIUpdate())
    mux.Handle("/api/ui-action", h.handleUIAction())
}
```

### 2. Create a Wait Registry Structure

- [x] Define a `WaitRegistry` struct to track pending requests
- [x] Implement methods to register, resolve, and timeout requests
- [x] Use channels for synchronization between the update endpoint and action handler

```go
// WaitRegistry tracks pending UI update requests waiting for user actions
type WaitRegistry struct {
    pending map[string]chan UIActionResponse
    mu      sync.RWMutex
    timeout time.Duration
}

// UIActionResponse represents the response from a UI action
type UIActionResponse struct {
    Action      string
    ComponentID string
    Data        map[string]interface{}
    Error       error
    Timestamp   time.Time
}

// NewWaitRegistry creates a new registry with the specified timeout
func NewWaitRegistry(timeout time.Duration) *WaitRegistry {
    return &WaitRegistry{
        pending: make(map[string]chan UIActionResponse),
        timeout: timeout,
    }
}

// Register adds a new request to the registry and returns a channel for the response
func (wr *WaitRegistry) Register(requestID string) chan UIActionResponse {
    wr.mu.Lock()
    defer wr.mu.Unlock()
    
    responseChan := make(chan UIActionResponse, 1)
    wr.pending[requestID] = responseChan
    
    return responseChan
}

// Resolve completes a pending request with the given action response
func (wr *WaitRegistry) Resolve(requestID string, response UIActionResponse) bool {
    wr.mu.Lock()
    defer wr.mu.Unlock()
    
    if ch, exists := wr.pending[requestID]; exists {
        ch <- response
        delete(wr.pending, requestID)
        return true
    }
    return false
}

// Cleanup removes a request from the registry
func (wr *WaitRegistry) Cleanup(requestID string) {
    wr.mu.Lock()
    defer wr.mu.Unlock()
    
    if ch, exists := wr.pending[requestID]; exists {
        close(ch)
        delete(wr.pending, requestID)
    }
}

// CleanupStale removes requests that have been in the registry longer than the timeout
func (wr *WaitRegistry) CleanupStale() int {
    wr.mu.Lock()
    defer wr.mu.Unlock()
    
    // This would require tracking timestamps for each request
    // Implementation would depend on how we track request timestamps
    return 0
}
```

### 3. Update the Server Struct

- [x] Remove UI-specific handlers from the Server struct
- [x] Add the UIHandler as a field in the Server struct
- [x] Update the Server initialization to create and use the UIHandler

```go
type Server struct {
    dir        string
    pages      map[string]UIDefinition
    routes     map[string]http.HandlerFunc
    watcher    *watcher.Watcher
    mux        *http.ServeMux
    mu         sync.RWMutex
    publisher  message.Publisher
    subscriber message.Subscriber
    events     events.EventManager
    sseHandler *sse.SSEHandler
    uiHandler  *UIHandler // New field for UI handler
}

func NewServer(dir string) (*Server, error) {
    // ... existing initialization
    
    // Create SSE handler
    sseHandler := sse.NewSSEHandler(eventManager, &log.Logger)
    
    // Create UI handler
    uiHandler := NewUIHandler(eventManager, sseHandler, &log.Logger)
    
    s := &Server{
        dir:        dir,
        pages:      make(map[string]UIDefinition),
        routes:     make(map[string]http.HandlerFunc),
        mux:        http.NewServeMux(),
        publisher:  publisher,
        subscriber: publisher,
        events:     eventManager,
        sseHandler: sseHandler,
        uiHandler:  uiHandler,
    }
    
    // Register component renderers
    s.registerComponentRenderers()
    
    // ... rest of initialization
    
    // Set up SSE endpoint
    s.mux.Handle("/sse", sseHandler)
    
    // Set up static file handler
    s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    // Register UI handlers
    uiHandler.RegisterHandlers(s.mux)
    
    // ... rest of initialization
    
    return s, nil
}
```

### 4. Implement the UI Update Handler in UIHandler

- [x] Move the UI update handler from Server to UIHandler
- [x] Generate a unique request ID for each update
- [x] Add the request ID to the UI definition sent to the frontend
- [x] Register the request in the wait registry
- [x] Wait for a response or timeout

```go
func (h *UIHandler) handleUIUpdate() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Only accept POST requests
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Parse JSON body
        var jsonData map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
            // Return detailed JSON error response
            h.sendErrorResponse(w, http.StatusBadRequest, "json_parse_error", "Invalid JSON: "+err.Error(), nil)
            return
        }

        // Generate a unique request ID
        requestID := uuid.New().String()
        
        // Add the request ID to the UI definition
        // This will be used by the frontend to associate actions with this request
        jsonData["requestID"] = requestID
        
        // Convert to YAML for storage
        yamlData, err := yaml.Marshal(jsonData)
        if err != nil {
            h.sendErrorResponse(w, http.StatusInternalServerError, "yaml_conversion_error", "Failed to convert to YAML: "+err.Error(), nil)
            return
        }

        // Parse into UIDefinition
        var def UIDefinition
        if err := yaml.Unmarshal(yamlData, &def); err != nil {
            h.sendErrorResponse(w, http.StatusBadRequest, "ui_definition_error", "Invalid UI definition: "+err.Error(), map[string]interface{}{
                "yaml": string(yamlData),
            })
            return
        }

        // Validate the UI definition
        validationErrors := h.validateUIDefinition(def)
        if len(validationErrors) > 0 {
            h.sendErrorResponse(w, http.StatusBadRequest, "ui_validation_error", "UI definition validation failed", map[string]interface{}{
                "details": validationErrors,
            })
            return
        }

        // Register this request in the wait registry
        responseChan := h.waitRegistry.Register(requestID)
        
        // Create and publish refresh-ui event with the request ID
        event := events.UIEvent{
            Type:      "refresh-ui",
            PageID:    "ui-update",
            Component: def,
        }

        // Add the request ID to the event data if possible
        if compData, ok := event.Component.(UIDefinition); ok {
            compData.RequestID = requestID
            event.Component = compData
        }

        if err := h.events.Publish("ui-update", event); err != nil {
            // Clean up the registry entry
            h.waitRegistry.Cleanup(requestID)
            h.sendErrorResponse(w, http.StatusInternalServerError, "event_publish_error", "Failed to publish event: "+err.Error(), nil)
            return
        }

        // Log that we're waiting for a response
        h.logger.Debug().
            Str("requestId", requestID).
            Msg("Waiting for UI action response")

        // Wait for the response or timeout
        select {
        case response := <-responseChan:
            // Request was resolved with a user action
            if response.Error != nil {
                // There was an error processing the action
                h.sendErrorResponse(w, http.StatusInternalServerError, "action_processing_error", response.Error.Error(), nil)
                return
            }
            
            // Log the successful response
            h.logger.Info().
                Str("requestId", requestID).
                Str("action", response.Action).
                Str("componentId", response.ComponentID).
                Msg("UI action received, completing request")
            
            // Return success with the action data
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            _ = json.NewEncoder(w).Encode(map[string]interface{}{
                "status": "success",
                "action": response.Action,
                "componentId": response.ComponentID,
                "data": response.Data,
            })
            
        case <-time.After(h.timeout):
            // Request timed out
            h.waitRegistry.Cleanup(requestID)
            
            h.logger.Warn().
                Str("requestId", requestID).
                Dur("timeout", h.timeout).
                Msg("Request timed out waiting for UI action")
            
            h.sendErrorResponse(w, http.StatusRequestTimeout, "timeout", "No user action received within the timeout period", nil)
        }
    }
}
```

### 5. Implement the UI Action Handler in UIHandler

- [x] Move the UI action handler from Server to UIHandler
- [x] Update to check for request IDs
- [x] Resolve waiting requests when matching actions are received

```go
func (h *UIHandler) handleUIAction() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Only accept POST requests
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Parse JSON body
        var action struct {
            ComponentID string                 `json:"componentId"`
            Action      string                 `json:"action"`
            Data        map[string]interface{} `json:"data"`
            RequestID   string                 `json:"requestId"` // Added field for request ID
        }

        if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
            http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
            return
        }

        // Determine if this is an important event to log at INFO level
        isImportantEvent := false
        switch action.Action {
        case "clicked", "changed", "submitted":
            isImportantEvent = true
        }

        // Log the action at appropriate level
        logger := h.logger.Debug()
        if isImportantEvent {
            logger = h.logger.Info()
        }

        // Create log entry with component and action info
        logger = logger.
            Str("componentId", action.ComponentID).
            Str("action", action.Action)

        // Add request ID to log if present
        if action.RequestID != "" {
            logger = logger.Str("requestId", action.RequestID)
        }

        // Add data to log if it exists and is relevant
        if len(action.Data) > 0 {
            // For form submissions, log the form data in detail
            if action.Action == "submitted" && action.Data["formData"] != nil {
                logger = logger.Interface("formData", action.Data["formData"])
            } else if action.Action == "changed" {
                // For changed events, log the new value
                logger = logger.Interface("data", action.Data)
            } else {
                // For other events, just log that data exists
                logger = logger.Bool("hasData", true)
            }
        }

        // Output the log message
        logger.Msg("UI action received")

        // Check if this action is associated with a waiting request
        if action.RequestID != "" && (action.Action == "clicked" || action.Action == "submitted") {
            // Try to resolve the waiting request
            resolved := h.waitRegistry.Resolve(action.RequestID, UIActionResponse{
                Action:      action.Action,
                ComponentID: action.ComponentID,
                Data:        action.Data,
                Error:       nil,
                Timestamp:   time.Now(),
            })
            
            if resolved {
                logger.Bool("waitingRequestResolved", true).Msg("Resolved waiting request")
            }
        }

        // Return success response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        err := json.NewEncoder(w).Encode(map[string]string{"status": "success"})
        if err != nil {
            http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
        }
    }
}
```

### 6. Move UI Update Page Handler to UIHandler

- [x] Move the UI update page handler from Server to UIHandler


### 7. Move UI Definition Validation to UIHandler

- [x] Move the validation logic from Server to UIHandler

```go
// validateUIDefinition checks a UI definition for common errors
func (h *UIHandler) validateUIDefinition(def UIDefinition) []map[string]interface{} {
    var errors []map[string]interface{}

    // Check if components exist
    if len(def.Components) == 0 {
        errors = append(errors, map[string]interface{}{
            "path":    "components",
            "message": "No components defined",
        })
        return errors
    }

    // Validate each component
    for i, comp := range def.Components {
        for typ, props := range comp {
            propsMap, ok := props.(map[string]interface{})
            if !ok {
                errors = append(errors, map[string]interface{}{
                    "path":    fmt.Sprintf("components[%d].%s", i, typ),
                    "message": "Component properties must be a map",
                })
                continue
            }

            // Check for required ID
            if _, hasID := propsMap["id"]; !hasID && h.requiresID(typ) {
                errors = append(errors, map[string]interface{}{
                    "path":    fmt.Sprintf("components[%d].%s", i, typ),
                    "message": "Component is missing required 'id' property",
                })
            }

            // Validate nested components in forms
            if typ == "form" {
                // ... existing validation logic for forms
            }

            // Validate list items
            if typ == "list" {
                // ... existing validation logic for lists
            }
        }
    }

    return errors
}

// requiresID returns true if the component type requires an ID
func (h *UIHandler) requiresID(componentType string) bool {
    switch componentType {
    case "text", "title":
        // These can optionally have IDs
        return false
    default:
        // All other components require IDs
        return true
    }
}
```

### 8. Update the UIEvent Structure

- [x] Add RequestID field to the UIEvent struct

```go
// In pkg/events/types.go
type UIEvent struct {
    Type      string      `json:"type"`                // e.g. "component-update", "page-reload"
    PageID    string      `json:"pageId"`              // Which page is being updated
    Component interface{} `json:"component"`           // The updated component data
    RequestID string      `json:"requestId,omitempty"` // Optional request ID for tracking user actions
}
```

### 9. Add Cleanup Mechanism for Orphaned Requests

- [x] Implement a background goroutine to clean up stale requests

```go
// Add to UIHandler initialization
func NewUIHandler(events events.EventManager, sseHandler *sse.SSEHandler, logger *zerolog.Logger) *UIHandler {
    h := &UIHandler{
        waitRegistry: NewWaitRegistry(30 * time.Second),
        events:       events,
        sseHandler:   sseHandler,
        logger:       logger,
        timeout:      30 * time.Second,
    }
    
    // Start background cleanup for orphaned requests
    go h.cleanupOrphanedRequests(context.Background())
    
    return h
}

// Add cleanup function
func (h *UIHandler) cleanupOrphanedRequests(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Perform cleanup
            count := h.waitRegistry.CleanupStale()
            if count > 0 {
                h.logger.Debug().Int("count", count).Msg("Cleaned up orphaned requests")
            }
            
        case <-ctx.Done():
            return
        }
    }
}
```

### 10. Modify the Frontend JavaScript to Include Request ID

- [x] Update the client-side code to extract and include the request ID in action events

```javascript
// Add to the page template or UI update JavaScript
function sendUIAction(componentId, action, data = {}) {
    // Get the request ID from the page if it exists
    const requestId = document.body.getAttribute('data-request-id') || 
                      document.querySelector('[data-request-id]')?.getAttribute('data-request-id') || 
                      '';
    
    logToConsole(`Component ${componentId} ${action}${requestId ? ' (request: ' + requestId + ')' : ''}`);
    
    // Rest of the existing function...
    
    // Include the request ID in the action data
    fetch('/api/ui-action', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            componentId: componentId,
            action: action,
            data: data,
            requestId: requestId // Include the request ID
        })
    })
    // Rest of the existing function...
}
```

### 11. Update the Page Template to Include Request ID

- [x] Modify the page template to include the request ID as a data attribute

```go
// In pageTemplate or pageContentTemplate
func pageContentTemplate(pageID string, def UIDefinition) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // Extract request ID if present
        requestID := extractRequestID(def)
        
        // Include the request ID as a data attribute on the container
        if requestID != "" {
            fmt.Fprintf(w, `<div data-request-id="%s" class="page-content">`, requestID)
        } else {
            fmt.Fprintf(w, `<div class="page-content">`)
        }
        
        // Rest of the existing template...
        
        fmt.Fprintf(w, `</div>`)
        return nil
    })
}

// Helper function to extract request ID from UI definition
func extractRequestID(def UIDefinition) string {
    // Try to extract from the component data
    for _, comp := range def.Components {
        for _, props := range comp {
            if propsMap, ok := props.(map[string]interface{}); ok {
                if rid, ok := propsMap["requestID"].(string); ok {
                    return rid
                }
            }
        }
    }
    return ""
}
```

## Configuration Options

- [ ] Make timeout configurable via environment variables or config file
- [ ] Add option to disable wait-for-response behavior for specific endpoints
- [ ] Configure logging verbosity for UI actions

```go
type UIHandlerConfig struct {
    Timeout           time.Duration
    CleanupInterval   time.Duration
    WaitForResponse   bool
    LogFormData       bool
    LogDebugEvents    bool
}

func DefaultUIHandlerConfig() UIHandlerConfig {
    return UIHandlerConfig{
        Timeout:           30 * time.Second,
        CleanupInterval:   5 * time.Minute,
        WaitForResponse:   true,
        LogFormData:       true,
        LogDebugEvents:    false,
    }
}
```

## Testing Plan

1. Test the basic flow:
   - [ ] Send a UI update request
   - [ ] Verify the UI is updated with the request ID
   - [ ] Trigger a button click or form submission
   - [ ] Verify the original request completes with the action data

2. Test timeout handling:
   - [ ] Send a UI update request
   - [ ] Don't trigger any user action
   - [ ] Verify the request times out after the configured duration

3. Test error handling:
   - [ ] Test with invalid UI definitions
   - [ ] Test with network errors during event publishing
   - [ ] Test with concurrent requests for the same component

4. Test the refactored code structure:
   - [ ] Verify all UI handlers are properly registered
   - [ ] Verify the Server struct no longer has UI-specific logic
   - [ ] Verify the UIHandler properly encapsulates all UI functionality

## Implementation Considerations

1. **Performance**: The wait registry should be efficient for concurrent requests
2. **Memory Management**: Ensure proper cleanup of channels and registry entries
3. **Error Handling**: Provide clear error messages for all failure scenarios
4. **Timeout Configuration**: Make the timeout configurable via server options
5. **Logging**: Add detailed logging for debugging and monitoring
6. **Code Organization**: Keep the UIHandler focused on UI concerns only
7. **Testability**: Design the code to be easily testable with mocks

## Benefits of the Refactoring

1. **Separation of Concerns**: UI handling logic is separated from the core server functionality
2. **Improved Maintainability**: Smaller, focused structs are easier to understand and maintain
3. **Better Testability**: UI handlers can be tested independently from the server
4. **Reduced Complexity**: The Server struct is simplified by removing UI-specific logic
5. **Extensibility**: New UI features can be added to the UIHandler without modifying the Server

## Conclusion

This implementation will create a synchronous flow between UI updates and user actions, allowing the server to wait for user interaction before completing the request. The use of channels and a wait registry provides a clean way to handle the asynchronous nature of user interactions while maintaining a synchronous API for the caller.

Additionally, the refactoring to move UI handlers into their own struct will significantly improve the code organization and maintainability of the codebase. 