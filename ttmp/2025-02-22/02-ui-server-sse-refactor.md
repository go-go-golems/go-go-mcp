# Technical Plan: SSE/Watermill Refactoring

This plan outlines the refactoring of the SSE/Watermill functionality in the UI server to create a more modular and maintainable structure.

## 1. Create Dedicated Event System Package
### 1.1 Event Types and Interfaces
- [x] Create new package `pkg/events` for event-related functionality
  ```go
  // pkg/events/types.go
  type UIEvent struct {
      Type      string      `json:"type"`      // e.g. "component-update", "page-reload"
      PageID    string      `json:"pageId"`    // Which page is being updated
      Component interface{} `json:"component"` // The updated component data
  }

  type EventManager interface {
      Subscribe(ctx context.Context, pageID string) (<-chan UIEvent, error)
      Publish(pageID string, event UIEvent) error
      Close() error
  }
  ```

### 1.2 Event Creation Helpers
- [x] Add helper functions for common event types
  ```go
  func NewPageReloadEvent(pageID string) UIEvent
  func NewComponentUpdateEvent(pageID string, componentID string, data interface{}) UIEvent
  ```

## 2. Create Watermill Implementation
### 2.1 Event Manager Implementation
- [x] Create Watermill-specific implementation in `pkg/events/watermill.go`
  ```go
  type WatermillEventManager struct {
      publisher  message.Publisher
      subscriber message.Subscriber
      logger     *zerolog.Logger
  }

  func NewWatermillEventManager(logger *zerolog.Logger) (*WatermillEventManager, error) {
      // Initialize Watermill components with proper configuration
  }
  ```

### 2.2 Message Conversion
- [x] Add methods for converting between UIEvent and Watermill messages
  ```go
  func (m *WatermillEventManager) eventToMessage(event UIEvent) (*message.Message, error)
  func (m *WatermillEventManager) messageToEvent(msg *message.Message) (UIEvent, error)
  ```

## 3. Create SSE Handler Package
### 3.1 Handler Structure
- [x] Create `pkg/server/sse` package for SSE-specific functionality
  ```go
  // pkg/server/sse/handler.go
  type SSEHandler struct {
      events events.EventManager
      logger *zerolog.Logger
  }

  func NewSSEHandler(events events.EventManager, logger *zerolog.Logger) *SSEHandler
  func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
  ```

### 3.2 Connection Management
- [x] Add connection tracking and management
  ```go
  type connection struct {
      pageID string
      w      http.ResponseWriter
      done   chan struct{}
  }

  func (h *SSEHandler) handleConnection(conn *connection)
  ```

## 4. Update Server Structure
### 4.1 Server Modifications
- [x] Modify Server struct to use new components
  ```go
  type Server struct {
      dir     string
      pages   map[string]UIDefinition
      routes  map[string]http.HandlerFunc
      watcher *watcher.Watcher
      mux     *http.ServeMux
      mu      sync.RWMutex
      events  events.EventManager
      logger  *zerolog.Logger
  }
  ```

### 4.2 Constructor Updates
- [x] Update NewServer to initialize new components
  ```go
  func NewServer(dir string, logger *zerolog.Logger) (*Server, error) {
      events, err := events.NewWatermillEventManager(logger)
      if err != nil {
          return nil, fmt.Errorf("failed to create event manager: %w", err)
      }
      // ... rest of initialization
  }
  ```

## 5. Implementation Steps

### 5.1 Event System Implementation
- [ ] Create base event types and interfacees
  - [ ] Define UIEvent struct with proper JSON tags
  - [ ] Define EventManager interface with documentation
  - [ ] Add helper methods for event creation
  - [ ] Add validation methods for events

### 5.2 Watermill Integration Implementation
- [ ] Implement WatermillEventManager
  - [ ] Add constructor with proper configuration
  - [ ] Implement Subscribe method with event conversion
  - [ ] Implement Publish method with message creation
  - [ ] Add proper cleanup in Close method
  - [ ] Add error handling and logging

### 5.3 SSE Handler Implementation
- [ ] Create dedicated SSE handler
  - [ ] Define SSE-specific headers as constants
  - [ ] Add connection management with context
  - [ ] Add error handling with proper status codes
  - [ ] Add structured logging
  - [ ] Add optional metrics collection
  - [ ] Add connection cleanup on client disconnect

### 5.4 Server Updates Implementation
- [ ] Update Server implementation
  - [ ] Modify constructor to use new event system
  - [ ] Update file watcher to publish events
  - [ ] Update page handlers to use event system
  - [ ] Add cleanup in shutdown method
  - [ ] Update logging to use structured logger