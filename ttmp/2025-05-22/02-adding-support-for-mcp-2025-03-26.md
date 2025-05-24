# Adding Support for MCP 2025-03-26 Features

**Document**: ttmp/2025-05-22/02-adding-support-for-mcp-2025-03-26.md  
**Date**: 2025-05-22  
**Author**: Manuel  
**Purpose**: Comprehensive guide for upgrading go-go-mcp to support MCP 2025-03-26 specification  

## Executive Summary

This document outlines the features and changes required to upgrade the go-go-mcp implementation from the 2024-11-05 MCP specification to the 2025-03-26 revision. The major changes include:

1. **Streamable HTTP Transport** - Replacing HTTP+SSE with a more flexible transport
2. **JSON-RPC Batching** - Support for multiple requests in a single call
3. **Tool Annotations** - Rich metadata for describing tool behavior
4. **Audio Content Support** - Adding audio to existing text/image content types
5. **Progress Notifications Enhancement** - Adding descriptive messages
6. **Completions Capability** - Argument autocompletion support

**Note**: OAuth 2.1 Authorization Framework is explicitly excluded from this document and will be handled separately.

## Current State Analysis


### Existing Implementation Overview

The current go-go-mcp implementation (based on 2024-11-05 spec) includes:

- **Transport Layer**: SSE and stdio transports (`pkg/transport/`)
- **Protocol Layer**: JSON-RPC 2.0 base with MCP message types (`pkg/protocol/`)
- **Server Features**: Tools, resources, prompts (`pkg/server/`, `pkg/tools/`, etc.)
- **Client Features**: Basic client implementation (`pkg/client/`)
- **Session Management**: Session handling (`pkg/session/`)

### Current Capabilities

```go
// Current ServerCapabilities from pkg/protocol/initialization.go
type ServerCapabilities struct {
    Prompts      *PromptsCapability     `json:"prompts,omitempty"`
    Resources    *ResourcesCapability   `json:"resources,omitempty"`
    Tools        *ToolsCapability       `json:"tools,omitempty"`
    Logging      *LoggingCapability     `json:"logging,omitempty"`
    Experimental map[string]interface{} `json:"experimental,omitempty"`
}
```

## Required Changes by Feature

## 1. Streamable HTTP Transport

### 1.1 Overview
Replace the current HTTP+SSE transport with a more flexible Streamable HTTP transport that enables real-time, bidirectional data flow with better compatibility.

### 1.2 Current State
- Current implementation: `pkg/transport/sse/transport.go` (571 lines)
- Uses Server-Sent Events (SSE) for server-to-client communication
- HTTP POST for client-to-server messages

### 1.3 Required Changes

#### 1.3.1 New Transport Implementation
- [ ] Create `pkg/transport/streamable_http/` directory
- [ ] Implement `StreamableHTTPTransport` struct
- [ ] Support bidirectional streaming over HTTP
- [ ] Maintain compatibility with existing SSE clients during transition

#### 1.3.2 Transport Interface Updates
```go
// pkg/transport/transport.go - Add new transport type
type TransportInfo struct {
    Type         string            // Add "streamable_http"
    RemoteAddr   string            
    Capabilities map[string]bool   // Add streaming capabilities
    Metadata     map[string]string 
}
```

#### 1.3.3 Configuration Updates
- [ ] Update transport configuration options
- [ ] Add streamable HTTP specific settings
- [ ] Maintain backward compatibility with SSE configuration

#### 1.3.4 Implementation Tasks
- [ ] Design bidirectional streaming protocol over HTTP
- [ ] Implement connection management for persistent streams
- [ ] Add proper error handling and reconnection logic
- [ ] Update documentation and examples

### 1.4 Migration Strategy
1. Implement new transport alongside existing SSE transport
2. Add feature flag for transport selection
3. Gradually migrate clients to new transport
4. Deprecate SSE transport in future release

## 2. JSON-RPC Batching Support

### 2.1 Overview
Add support for JSON-RPC batching to allow clients to send multiple requests in one call, improving efficiency and reducing latency.

### 2.2 Current State
- Current implementation handles single requests only
- `pkg/protocol/base.go` defines individual Request/Response types

### 2.3 Required Changes

#### 2.3.1 Protocol Types Updates
```go
// pkg/protocol/base.go - Add batch support
type BatchRequest []Request
type BatchResponse []Response

// Add batch validation
func (br BatchRequest) Validate() error {
    if len(br) == 0 {
        return fmt.Errorf("batch request cannot be empty")
    }
    // Additional validation logic
    return nil
}
```

#### 2.3.2 Transport Layer Updates
- [ ] Update `transport.RequestHandler` interface to support batch requests
- [ ] Modify all transport implementations (stdio, sse, streamable_http)
- [ ] Add batch size limits and validation

#### 2.3.3 Server Implementation
```go
// pkg/server/ - Add batch handling
type BatchRequestHandler interface {
    HandleBatchRequest(ctx context.Context, batch BatchRequest) (BatchResponse, error)
}
```

#### 2.3.4 Client Implementation
- [ ] Add batch request builders in client package
- [ ] Implement batch response handling
- [ ] Add convenience methods for common batch operations

#### 2.3.5 Implementation Tasks
- [ ] Define batch request/response types
- [ ] Update JSON-RPC message parsing to handle arrays
- [ ] Implement batch processing in server handlers
- [ ] Add batch support to all transport layers
- [ ] Update client libraries with batch capabilities
- [ ] Add configuration for batch size limits

### 2.4 Error Handling
- Individual request errors within a batch should not fail the entire batch
- Implement partial success handling
- Add proper error aggregation and reporting

## 3. Tool Annotations

### 3.1 Overview
Add comprehensive tool annotations for better describing tool behavior, including whether tools are read-only, destructive, or have other behavioral characteristics.

### 3.2 Current State
```go
// pkg/protocol/tools.go - Current Tool definition
type Tool struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"inputSchema"`
}
```

### 3.3 Required Changes

#### 3.3.1 Enhanced Tool Definition
```go
// pkg/protocol/tools.go - Enhanced Tool with annotations
type Tool struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"inputSchema"`
    Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

type ToolAnnotations struct {
    // Behavioral annotations
    ReadOnly     bool     `json:"readOnly,omitempty"`     // Tool only reads data
    Destructive  bool     `json:"destructive,omitempty"`  // Tool may modify/delete data
    Idempotent   bool     `json:"idempotent,omitempty"`   // Safe to call multiple times
    
    // Performance annotations
    LongRunning  bool     `json:"longRunning,omitempty"`  // May take significant time
    Expensive    bool     `json:"expensive,omitempty"`    // Resource intensive
    
    // Security annotations
    RequiresAuth bool     `json:"requiresAuth,omitempty"` // Requires authentication
    Sensitive    bool     `json:"sensitive,omitempty"`    // Handles sensitive data
    
    // Categorization
    Category     string   `json:"category,omitempty"`     // Tool category
    Tags         []string `json:"tags,omitempty"`         // Searchable tags
    
    // Dependencies
    Dependencies []string `json:"dependencies,omitempty"` // Required tools/resources
    
    // Custom annotations
    Custom       map[string]interface{} `json:"custom,omitempty"`
}
```

#### 3.3.2 Tool Registry Updates
- [ ] Update tool registration to include annotations
- [ ] Add annotation validation
- [ ] Implement annotation-based tool filtering and discovery

#### 3.3.3 Server Implementation Updates
```go
// pkg/tools/ - Update tool registration
type ToolRegistration struct {
    Tool        Tool
    Handler     ToolHandler
    Annotations *ToolAnnotations
}

// Add annotation-based queries
func (r *ToolRegistry) GetToolsByAnnotation(filter AnnotationFilter) []Tool
```

#### 3.3.4 Client Implementation Updates
- [ ] Add annotation-aware tool discovery
- [ ] Implement tool filtering based on annotations
- [ ] Add safety checks based on destructive/sensitive annotations

#### 3.3.5 Implementation Tasks
- [ ] Define comprehensive annotation schema
- [ ] Update tool registration APIs
- [ ] Implement annotation validation
- [ ] Add annotation-based tool discovery
- [ ] Update documentation with annotation guidelines
- [ ] Create annotation best practices guide

### 3.4 Annotation Categories

#### 3.4.1 Safety Annotations
- `readOnly`: Tool only reads data, never modifies
- `destructive`: Tool may delete or irreversibly modify data
- `idempotent`: Safe to call multiple times with same parameters

#### 3.4.2 Performance Annotations
- `longRunning`: Tool may take significant time to complete
- `expensive`: Tool is resource-intensive (CPU, memory, network)
- `cached`: Tool results can be cached

#### 3.4.3 Security Annotations
- `requiresAuth`: Tool requires authentication
- `sensitive`: Tool handles sensitive data
- `audit`: Tool actions should be audited

## 4. Audio Content Support

### 4.1 Overview
Add support for audio data to join the existing text and image content types in tool results and other content contexts.

### 4.2 Current State
```go
// pkg/protocol/tools.go - Current ToolContent
type ToolContent struct {
    Type     string           `json:"type"`           // "text", "image", or "resource"
    Text     string           `json:"text,omitempty"` 
    Data     string           `json:"data,omitempty"` // Base64 encoded for image
    MimeType string           `json:"mimeType,omitempty"`
    Resource *ResourceContent `json:"resource,omitempty"`
}
```

### 4.3 Required Changes

#### 4.3.1 Enhanced Content Types
```go
// pkg/protocol/tools.go - Add audio support
type ToolContent struct {
    Type     string           `json:"type"`           // "text", "image", "audio", or "resource"
    Text     string           `json:"text,omitempty"` 
    Data     string           `json:"data,omitempty"` // Base64 encoded for image/audio
    MimeType string           `json:"mimeType,omitempty"`
    Resource *ResourceContent `json:"resource,omitempty"`
    
    // Audio-specific fields
    Duration float64          `json:"duration,omitempty"` // Duration in seconds
    SampleRate int            `json:"sampleRate,omitempty"` // Audio sample rate
}
```

#### 4.3.2 Audio Content Helpers
```go
// pkg/protocol/tools.go - Add audio content helpers
func NewAudioContent(base64Data, mimeType string, duration float64, sampleRate int) ToolContent {
    return ToolContent{
        Type:       "audio",
        Data:       base64Data,
        MimeType:   mimeType,
        Duration:   duration,
        SampleRate: sampleRate,
    }
}

func WithAudio(base64Data, mimeType string, duration float64, sampleRate int) ToolResultOption {
    return func(tr *ToolResult) {
        tr.Content = append(tr.Content, NewAudioContent(base64Data, mimeType, duration, sampleRate))
    }
}
```

#### 4.3.3 Content Validation
- [ ] Add MIME type validation for audio formats
- [ ] Implement audio content size limits
- [ ] Add audio format support documentation

#### 4.3.4 Implementation Tasks
- [ ] Update ToolContent struct with audio fields
- [ ] Add audio content creation helpers
- [ ] Implement audio content validation
- [ ] Update content type constants
- [ ] Add audio format documentation
- [ ] Create audio content examples

### 4.4 Supported Audio Formats
- `audio/wav` - WAV format
- `audio/mp3` - MP3 format  
- `audio/ogg` - OGG format
- `audio/flac` - FLAC format
- `audio/aac` - AAC format

## 5. Progress Notifications Enhancement

### 5.1 Overview
Add a `message` field to `ProgressNotification` to provide descriptive status updates beyond just numeric progress.

### 5.2 Current State
- Progress notifications exist but may lack descriptive messaging
- Need to locate current progress notification implementation

### 5.3 Required Changes

#### 5.3.1 Enhanced Progress Notification
```go
// pkg/protocol/ - Enhanced progress notification
type ProgressNotification struct {
    ProgressToken string  `json:"progressToken"`
    Progress      float64 `json:"progress"`      // 0.0 to 1.0
    Message       string  `json:"message,omitempty"` // NEW: Descriptive message
    Total         int64   `json:"total,omitempty"`   // Total units of work
    Completed     int64   `json:"completed,omitempty"` // Completed units
}
```

#### 5.3.2 Progress Helper Functions
```go
// pkg/protocol/ - Progress notification helpers
func NewProgressNotification(token string, progress float64, message string) *ProgressNotification {
    return &ProgressNotification{
        ProgressToken: token,
        Progress:      progress,
        Message:       message,
    }
}

func NewDetailedProgressNotification(token string, completed, total int64, message string) *ProgressNotification {
    progress := float64(completed) / float64(total)
    return &ProgressNotification{
        ProgressToken: token,
        Progress:      progress,
        Message:       message,
        Total:         total,
        Completed:     completed,
    }
}
```

#### 5.3.3 Implementation Tasks
- [ ] Locate current progress notification implementation
- [ ] Add message field to progress notification struct
- [ ] Update progress notification senders to include messages
- [ ] Add progress message helpers
- [ ] Update documentation with progress message examples

### 5.4 Progress Message Examples
- "Initializing connection..."
- "Processing file 3 of 10..."
- "Downloading data from remote server..."
- "Analyzing results..."
- "Operation completed successfully"

## 6. Completions Capability

### 6.1 Overview
Add `completions` capability to explicitly indicate support for argument autocompletion suggestions.

### 6.2 Current State
```go
// pkg/protocol/initialization.go - Current capabilities
type ServerCapabilities struct {
    Prompts      *PromptsCapability     `json:"prompts,omitempty"`
    Resources    *ResourcesCapability   `json:"resources,omitempty"`
    Tools        *ToolsCapability       `json:"tools,omitempty"`
    Logging      *LoggingCapability     `json:"logging,omitempty"`
    Experimental map[string]interface{} `json:"experimental,omitempty"`
}
```

### 6.3 Required Changes

#### 6.3.1 New Completions Capability
```go
// pkg/protocol/initialization.go - Add completions capability
type ServerCapabilities struct {
    Prompts      *PromptsCapability     `json:"prompts,omitempty"`
    Resources    *ResourcesCapability   `json:"resources,omitempty"`
    Tools        *ToolsCapability       `json:"tools,omitempty"`
    Logging      *LoggingCapability     `json:"logging,omitempty"`
    Completions  *CompletionsCapability `json:"completions,omitempty"` // NEW
    Experimental map[string]interface{} `json:"experimental,omitempty"`
}

type CompletionsCapability struct {
    // Indicates server supports argument autocompletion
}
```

#### 6.3.2 Completions Protocol Messages
```go
// pkg/protocol/ - New completions protocol
type CompletionRequest struct {
    Ref      CompletionRef `json:"ref"`      // What to complete
    Argument string        `json:"argument"` // Argument name to complete
    Value    string        `json:"value"`    // Partial value to complete
}

type CompletionRef struct {
    Type string `json:"type"` // "tools", "prompts", "resources"
    Name string `json:"name"` // Name of the tool/prompt/resource
}

type CompletionResult struct {
    Completions []Completion `json:"completions"`
}

type Completion struct {
    Label       string                 `json:"label"`       // Display text
    Value       string                 `json:"value"`       // Completion value
    Description string                 `json:"description,omitempty"` // Optional description
    Detail      string                 `json:"detail,omitempty"`      // Additional detail
    Kind        CompletionKind         `json:"kind,omitempty"`        // Type of completion
    Data        map[string]interface{} `json:"data,omitempty"`        // Custom data
}

type CompletionKind string

const (
    CompletionKindText     CompletionKind = "text"
    CompletionKindKeyword  CompletionKind = "keyword"
    CompletionKindFunction CompletionKind = "function"
    CompletionKindVariable CompletionKind = "variable"
    CompletionKindFile     CompletionKind = "file"
    CompletionKindFolder   CompletionKind = "folder"
)
```

#### 6.3.3 Server Implementation
```go
// pkg/server/ - Add completions handler
type CompletionsHandler interface {
    HandleCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResult, error)
}

// Tool-specific completion provider
type ToolCompletionProvider interface {
    GetCompletions(ctx context.Context, argument string, value string) ([]Completion, error)
}
```

#### 6.3.4 Implementation Tasks
- [ ] Define completions capability and protocol messages
- [ ] Implement completions request/response handling
- [ ] Add completion providers for tools, prompts, resources
- [ ] Integrate with existing tool/prompt/resource systems
- [ ] Add client-side completion support
- [ ] Create completion examples and documentation

### 6.4 Completion Use Cases
- Tool argument values (file paths, enum values, etc.)
- Prompt parameter suggestions
- Resource URI completions
- Dynamic value suggestions based on context

## Implementation Plan

### Phase 1: Foundation (Weeks 1-2)
- [ ] **1.1** Implement JSON-RPC batching support
  - Update protocol types
  - Modify transport layers
  - Add batch processing logic
- [ ] **1.2** Add audio content support
  - Update content types
  - Add audio helpers
  - Implement validation

### Phase 2: Enhanced Features (Weeks 3-4)
- [ ] **2.1** Implement tool annotations
  - Design annotation schema
  - Update tool registration
  - Add annotation-based discovery
- [ ] **2.2** Enhance progress notifications
  - Add message field
  - Update notification senders
  - Add helper functions

### Phase 3: Advanced Features (Weeks 5-6)
- [ ] **3.1** Implement completions capability
  - Define completion protocol
  - Add completion providers
  - Integrate with existing systems
- [ ] **3.2** Implement streamable HTTP transport
  - Design new transport
  - Implement bidirectional streaming
  - Add migration path from SSE

### Phase 4: Integration & Testing (Weeks 7-8)
- [ ] **4.1** Integration testing
  - Test all new features together
  - Verify backward compatibility
  - Performance testing
- [ ] **4.2** Documentation and examples
  - Update API documentation
  - Create feature examples
  - Migration guides

## Testing Strategy

### Unit Tests
- [ ] Test all new protocol message types
- [ ] Test batch request/response handling
- [ ] Test audio content validation
- [ ] Test tool annotation filtering
- [ ] Test completion providers

### Integration Tests
- [ ] Test new transport with existing clients
- [ ] Test batch operations end-to-end
- [ ] Test audio content transmission
- [ ] Test completion workflows

### Backward Compatibility Tests
- [ ] Ensure existing clients continue to work
- [ ] Test graceful degradation of new features
- [ ] Verify protocol version negotiation

## Migration Considerations

### Backward Compatibility
- All new features should be optional and backward compatible
- Existing clients should continue to work without modification
- New capabilities should be discoverable through capability negotiation

### Configuration Updates
- Add feature flags for new capabilities
- Update configuration schemas
- Provide migration tools for existing configurations

### Documentation Updates
- Update API documentation for all new features
- Create migration guides
- Add examples for each new feature

## Risk Assessment

### High Risk
- **Streamable HTTP Transport**: Complex implementation, potential compatibility issues
- **JSON-RPC Batching**: Changes to core message handling, potential performance impact

### Medium Risk
- **Tool Annotations**: Schema design complexity, potential breaking changes to tool APIs
- **Completions Capability**: New protocol surface area, integration complexity

### Low Risk
- **Audio Content Support**: Additive change, minimal impact on existing functionality
- **Progress Notifications Enhancement**: Simple additive change

## Success Criteria

### Functional Requirements
- [ ] All new MCP 2025-03-26 features implemented (except OAuth)
- [ ] Backward compatibility maintained with 2024-11-05 clients
- [ ] Performance impact < 10% for existing operations
- [ ] Comprehensive test coverage (>90%) for new features

### Quality Requirements
- [ ] All new code follows existing code style and patterns
- [ ] Documentation updated for all new features
- [ ] Migration guides provided
- [ ] Examples created for each new feature

### Performance Requirements
- [ ] Batch operations show measurable performance improvement
- [ ] New transport performs at least as well as existing SSE transport
- [ ] Audio content handling doesn't impact text/image performance

## Conclusion

This document provides a comprehensive roadmap for upgrading go-go-mcp to support the MCP 2025-03-26 specification. The implementation should be done in phases to manage complexity and ensure stability. Each feature should be thoroughly tested and documented before moving to the next phase.

The most complex changes are the Streamable HTTP Transport and JSON-RPC Batching, which should be prioritized and given extra attention during implementation. The other features are more straightforward additive changes that can be implemented with lower risk.

Regular testing and validation against the official MCP specification should be performed throughout the implementation process to ensure compliance and interoperability. 