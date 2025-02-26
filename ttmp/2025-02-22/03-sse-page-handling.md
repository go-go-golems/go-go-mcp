# Understanding SSE Page Handling in Go-Go-MCP

This guide explains how Server-Sent Events (SSE) and page reload watching work in our system, including how templates are selected and where page data is stored. This document is intended for new developers to understand the architecture and flow of our real-time UI update system.

## Overview

Our UI server uses Server-Sent Events (SSE) to provide real-time updates to web pages without requiring page refreshes. This creates a responsive, dynamic user interface that feels like a single-page application while maintaining the simplicity of server-rendered HTML.

The system consists of several key components:

1. **File Watcher**: Monitors YAML files for changes
2. **Event System**: Publishes and subscribes to UI events
3. **SSE Handler**: Manages SSE connections and renders HTML updates
4. **Template System**: Renders components and pages using the templ templating engine

## File Watching and Page Loading

When the server starts, it scans a directory for YAML files that define UI pages. Each time a file is added, modified, or removed, the server updates its internal state and notifies connected clients.

### How Page Loading Works

1. The server loads all YAML files at startup:

```go
func (s *Server) loadPages() error {
    // Walk through directory and load each YAML file
    return filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
        if strings.HasSuffix(d.Name(), ".yaml") {
            if err := s.loadPage(path); err != nil {
                return err
            }
        }
        return nil
    })
}
```

2. Each YAML file is parsed into a `UIDefinition` struct:

```go
type UIDefinition struct {
    Components []map[string]interface{} `yaml:"components"`
}
```

3. The parsed definitions are stored in the server's `pages` map:

```go
s.mu.Lock()
s.pages[relPath] = def
s.routes[urlPath] = s.handlePage(relPath)
s.mu.Unlock()
```

4. When a file changes, the watcher triggers a callback:

```go
w := watcher.NewWatcher(
    watcher.WithPaths(dir),
    watcher.WithMask("**/*.yaml"),
    watcher.WithWriteCallback(s.handleFileChange),
    watcher.WithRemoveCallback(s.handleFileRemove),
)
```

5. The callback reloads the page and publishes a page-reload event:

```go
// Publish page reload event with the page definition
event := events.NewPageReloadEvent(relPath, def)
if err := s.events.Publish(relPath, event); err != nil {
    log.Error().Err(err).Str("path", relPath).Msg("Failed to publish page reload event")
}
```

## Event System

The event system uses the Watermill library to implement a publish-subscribe pattern. This decouples the file watcher from the SSE handler, making the system more maintainable and extensible.

### Event Types

Events are represented by the `UIEvent` struct:

```go
type UIEvent struct {
    Type      string      `json:"type"`      // e.g. "component-update", "page-reload"
    PageID    string      `json:"pageId"`    // Which page is being updated
    Component interface{} `json:"component"` // The updated component data
}
```

The system supports several event types:
- `page-reload`: Reloads an entire page

### Creating Events

Helper functions create properly formatted events:

```go
// NewPageReloadEvent creates a new event for reloading a page
func NewPageReloadEvent(pageID string, pageDef interface{}) UIEvent {
    return UIEvent{
        Type:      "page-reload",
        PageID:    pageID,
        Component: map[string]interface{}{"data": pageDef},
    }
}
```

## SSE Handler

The SSE handler manages client connections and sends events to browsers. It's responsible for:

1. Setting up SSE connections with proper headers
2. Tracking active connections
3. Processing events and rendering HTML updates
4. Sending formatted SSE messages to clients

### Connection Management

When a client connects to the `/sse` endpoint, the handler:

1. Sets up SSE headers:
```go
w.Header().Set(headerContentType, contentTypeSSE)
w.Header().Set(headerCacheControl, cacheControlNoCache)
w.Header().Set(headerConnection, connectionKeepAlive)
```

2. Creates and registers a connection:
```go
conn := &connection{
    pageID: pageID,
    w:      w,
    done:   make(chan struct{}),
}
h.registerConnection(conn)
```

3. Subscribes to events for the requested page:
```go
events, err := h.events.Subscribe(ctx, conn.pageID)
```

4. Processes events as they arrive:
```go
for {
    select {
    case event, ok := <-events:
        if !ok {
            return
        }
        if err := h.processEvent(conn, event); err != nil {
            // Handle error
        }
    // Other cases for context cancellation, etc.
    }
}
```

### Event Processing

The handler processes different event types:

```go
switch event.Type {
    return h.handleComponentUpdate(conn, event)
case "page-reload":
    return h.handlePageReload(conn, event)
default:
    // For unknown event types, just send the raw event
    return h.sendEvent(conn, event.Type, event)
}
```

### Page Reload Handling

The page reload handler is particularly interesting:

```go
func (h *SSEHandler) handlePageReload(conn *connection, event events.UIEvent) error {
    // First, send a page-reload event notification
    if err := h.sendRawEvent(conn, "page-reload", "{}"); err != nil {
        return fmt.Errorf("failed to send page-reload event: %w", err)
    }

    // Then, check if we have a page renderer to render the full page
    h.mu.RLock()
    pageRenderer, ok := h.renderers["page-content-template"]
    h.mu.RUnlock()

    if !ok {
        return nil
    }

    // Render the page content template with the page ID and event data
    html, err := pageRenderer(conn.pageID, event.Component)
    if err != nil {
        return fmt.Errorf("failed to render page content template: %w", err)
    }

    // Send the rendered page as a component-update event
    return h.sendRawEvent(conn, "component-update", html)
}
```

This function:
1. Notifies the client that a page reload is happening
2. Renders the updated page content using a registered renderer
3. Sends the rendered HTML as a component update

## Template Selection

The system uses the `templ` templating engine to render HTML. Templates are defined in `templates.templ` and registered with the SSE handler.

### Template Structure

We have several key templates:
- `base`: The outer HTML structure (doctype, head, body)
- `pageTemplate`: Renders a complete page with the base template
- `pageContentTemplate`: Renders just the content portion of a page
- `renderComponent`: Renders individual UI components

The separation between `pageTemplate` and `pageContentTemplate` is crucial for efficient updates:

```go
templ pageTemplate(name string, def UIDefinition) {
    @base("UI Server - " + name) {
        @pageContentTemplate(name, def)
    }
}

templ pageContentTemplate(name string, def UIDefinition) {
    <div
        hx-ext="sse"
        sse-connect={ "/sse?page=" + name }
        sse-swap="component-update"
    >
        <!-- Page content here -->
    </div>
}
```

### Renderer Registration

The server registers renderers for different event types:

```go
func (s *Server) registerComponentRenderers() {
    // Register page-content-template renderer
    s.sseHandler.RegisterRenderer("page-content-template", func(pageID string, data interface{}) (string, error) {
        // Get the page definition from event data or server state
        var def UIDefinition
        
        // Fall back to stored definition if needed
        if len(def.Components) == 0 {
            s.mu.RLock()
            storedDef, ok := s.pages[pageID]
            s.mu.RUnlock()
            
            if !ok {
                return "", fmt.Errorf("page not found: %s", pageID)
            }
            
            def = storedDef
        }
        
        // Render the template
        var buf bytes.Buffer
        err := pageContentTemplate(pageID, def).Render(context.Background(), &buf)
        return buf.String(), err
    })
}
```

## Page Data Management

Page definitions are stored in several places:

1. **Server's pages map**: The primary storage for all loaded page definitions
   ```go
   s.pages[relPath] = def
   ```

2. **Event payload**: When a page reload event is published, the page definition is included
   ```go
   event := events.NewPageReloadEvent(relPath, def)
   ```

3. **Renderer state**: The page-content-template renderer tries to use the definition from the event, falling back to the stored definition if needed

This redundancy ensures that even if there's a race condition between updating the server state and sending events, clients always receive the most up-to-date page definition.

## The Complete Flow

Let's walk through the complete flow of a page update:

1. A YAML file is modified on disk
2. The file watcher detects the change and calls `handleFileChange`
3. The server loads and parses the updated YAML into a `UIDefinition`
4. The server updates its internal `pages` map with the new definition
5. The server creates a page-reload event containing the new definition
6. The event is published to the event system
7. The SSE handler receives the event for all connections subscribed to that page
8. For each connection, the handler:
   - Sends a page-reload notification
   - Uses the page-content-template renderer to render the updated page
   - Sends the rendered HTML as a component-update event
9. The browser receives the SSE events and:
   - Processes the page-reload notification (which might trigger client-side logic)
   - Updates the DOM with the new HTML from the component-update event

This architecture provides several benefits:
- **Efficiency**: Only the necessary parts of the page are updated
- **Reliability**: Multiple layers ensure the correct data is used
- **Flexibility**: Different event types can trigger different rendering strategies
- **Maintainability**: Clear separation of concerns between components

## Conclusion

The SSE page handling system provides a powerful way to create dynamic, real-time web interfaces without the complexity of a full JavaScript framework. By understanding how the file watcher, event system, SSE handler, and templates work together, you can effectively maintain and extend the system.

For further exploration, look at:
- The `handleComponentUpdate` method to understand how individual components are updated
- The `yamlString` function to see how YAML is rendered for display
- The htmx attributes in the templates that connect the client-side to the SSE events 