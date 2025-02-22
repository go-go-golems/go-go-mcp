# Technical Plan: Implementing SSE with Watermill and HTMX SSE Extension

## 1. Dependencies
```go
// go.mod additions
require (
    github.com/ThreeDotsLabs/watermill v1.3.5
)
```

## 2. Event Types
```go
type UIUpdateEvent struct {
    Type      string      `json:"type"`      // e.g. "component-update", "page-reload"
    PageID    string      `json:"pageId"`    // Which page is being updated
    Component interface{} `json:"component"` // The updated component data
}
```

## 3. Server Structure
```go
type Server struct {
    // ... existing fields ...
    publisher message.Publisher
    subscriber message.Subscriber
    router *message.Router
}

func NewServer(dir string) *Server {
    // ... existing initialization ...
    
    // Initialize watermill router
    router := message.NewRouter(message.RouterConfig{})
    
    // Create publisher/subscriber (using Pub/Sub or AMQP)
    publisher := gochannel.NewGoChannel(
        gochannel.Config{},
        watermill.NewStdLogger(false, false),
    )
    subscriber := gochannel.NewGoChannel(
        gochannel.Config{},
        watermill.NewStdLogger(false, false),
    )
    
    return &Server{
        // ... existing fields ...
        publisher: publisher,
        subscriber: subscriber,
        router: router,
    }
}
```

## 4. SSE Handler
```go
func (s *Server) handleSSE() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Set SSE headers
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        
        // Get page ID from query
        pageID := r.URL.Query().Get("page")
        
        // Subscribe to page-specific topic
        messages, err := s.subscriber.Subscribe(r.Context(), "ui-updates."+pageID)
        if err != nil {
            http.Error(w, "Failed to subscribe to events", http.StatusInternalServerError)
            return
        }
        
        // Stream messages
        for msg := range messages {
            // Format SSE message
            fmt.Fprintf(w, "event: %s\n", msg.Metadata["event-type"])
            fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
            w.(http.Flusher).Flush()
        }
    }
}
```

## 5. Template Modifications
### 5.1 Base Template
```go
templ base(title string) {
    <!DOCTYPE html>
    <html lang="en">
        <head>
            // ... existing head content ...
            <script src="https://unpkg.com/htmx.org@1.9.10"></script>
            <script src="https://unpkg.com/htmx-ext-sse@2.2.2"></script>
        </head>
        <body hx-ext="sse">
            // ... existing content ...
        </body>
    </html>
}
```

### 5.2 Page Template
```go
templ pageTemplate(name string, def UIDefinition) {
    @base("UI Server - " + name) {
        <div 
            class="row"
            hx-ext="sse" 
            sse-connect={"/sse?page=" + name}
        >
            <div class="col-md-6" sse-swap="component-update">
                // ... existing component rendering ...
            </div>
            <div class="col-md-6" sse-swap="yaml-update">
                // ... existing YAML source ...
            </div>
        </div>
    }
}
```

## 6. Component Updates
### 6.1 Component Rendering
```go
templ renderComponent(typ string, props map[string]interface{}) {
    <div 
        id={fmt.Sprintf("component-%s", props["id"])}
        sse-swap="component-update"
    >
        // ... existing component rendering ...
    </div>
}
```

## 7. Event Publishing
### 7.1 File Watcher Events
```go
func (s *Server) handleFileChange(path string) error {
    // ... existing file loading code ...
    
    // Create and publish event
    event := UIUpdateEvent{
        Type:   "page-reload",
        PageID: relPath,
    }
    
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    msg := message.NewMessage(
        watermill.NewUUID(),
        payload,
    )
    msg.Metadata = map[string]string{
        "event-type": "page-reload",
    }
    
    return s.publisher.Publish("ui-updates."+relPath, msg)
}
```

### 7.2 Component Updates
```go
func (s *Server) publishComponentUpdate(pageID string, componentID string, newProps map[string]interface{}) error {
    event := UIUpdateEvent{
        Type:      "component-update",
        PageID:    pageID,
        Component: newProps,
    }
    
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    msg := message.NewMessage(
        watermill.NewUUID(),
        payload,
    )
    msg.Metadata = map[string]string{
        "event-type": "component-update",
    }
    
    return s.publisher.Publish("ui-updates."+pageID, msg)
}
```

## 8. Implementation Plan

### 8.1 Core Infrastructure
- [ ] Add Watermill dependencies and initialize publisher/subscriber
  - [ ] Add dependencies to go.mod
  - [ ] Initialize router, publisher, and subscriber in NewServer
  - [ ] Add graceful shutdown handling

### 8.2 SSE Implementation
- [ ] Create SSE endpoint with Watermill subscription
  - [ ] Implement handleSSE function
  - [ ] Add route to server mux
  - [ ] Add error handling and connection management

### 8.3 Frontend Integration
- [ ] Update templates to use the new HTMX SSE extension
  - [ ] Add SSE extension script to base template
  - [ ] Update page template with SSE attributes
  - [ ] Add event handling for component updates

### 8.4 Component System
- [ ] Modify component rendering for partial updates
  - [ ] Add SSE swap attributes to components
  - [ ] Implement partial update rendering
  - [ ] Add event type handling

### 8.5 Event System
- [ ] Update file watcher to publish through Watermill
  - [ ] Modify handleFileChange to publish events
  - [ ] Add error handling for publishing
  - [ ] Test file change propagation

### 8.6 API Layer
- [ ] Add component update publishing mechanism
  - [ ] Implement publishComponentUpdate
  - [ ] Add API endpoint for manual updates
  - [ ] Add validation and error handling

### 8.7 Testing & Validation
- [ ] Testing
  - [ ] Test file modifications and SSE updates
  - [ ] Test component updates through API
  - [ ] Verify multiple client connections
  - [ ] Test reconnection scenarios
  - [ ] Load test with multiple concurrent updates

## 9. Next Steps
- [ ] Review and prioritize tasks
- [ ] Set up development environment with Watermill
- [ ] Begin implementation of core functionality
