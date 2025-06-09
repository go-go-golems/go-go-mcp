# How to Add MCP Client LLM Inference - Guide and Roadmap

## Purpose and Scope

This document provides a comprehensive guide for implementing LLM inference capabilities in the go-go-mcp project using the Model Context Protocol (MCP). The goal is to enable the MCP client to request LLM completions from connected servers, enabling sophisticated agentic behaviors while maintaining security and user control.

## What We've Accomplished So Far

Based on the exploration of the codebase, we have:

1. **Analyzed the current MCP implementation** in `/pkg`:
   - Client implementation with transport support (stdio, SSE)
   - Protocol definitions for base JSON-RPC 2.0 communication
   - Existing sampling protocol structures
   - Provider interfaces for prompts, resources, and tools

2. **Identified existing infrastructure**:
   - `pkg/client/client.go` already has a `CreateMessage` method for sampling
   - `pkg/protocol/sampling.go` defines the sampling message structures
   - Transport layer supports both stdio and SSE connections
   - Capability negotiation system is in place

3. **Researched MCP protocol specifications**:
   - MCP 2025-03-26 specification with sampling capabilities
   - Security and human-in-the-loop design principles
   - Message format and parameter specifications

## Key Technical Insights

### Current State Analysis

The go-go-mcp project already has foundational MCP client infrastructure:

```go
// pkg/client/client.go - Line 361-402
func (c *Client) CreateMessage(ctx context.Context, messages []protocol.Message, 
    modelPreferences protocol.ModelPreferences, systemPrompt string, maxTokens int) (*protocol.Message, error)
```

```go
// pkg/protocol/sampling.go - Complete sampling protocol structures
type Message struct {
    Role    string         `json:"role"`
    Content MessageContent `json:"content"`
    Model   string         `json:"model,omitempty"`
}

type ModelPreferences struct {
    Hints                []ModelHint `json:"hints,omitempty"`
    CostPriority         float64     `json:"costPriority,omitempty"`
    SpeedPriority        float64     `json:"speedPriority,omitempty"`
    IntelligencePriority float64     `json:"intelligencePriority,omitempty"`
}
```

### Architecture Overview

The MCP protocol follows a client-host-server architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MCP Host      │    │   MCP Client    │    │   MCP Server    │
│ (LLM App/IDE)   │◄──►│ (go-go-mcp)     │◄──►│ (Tool Provider) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

For LLM inference via sampling, the flow is:
1. **Server** requests LLM completion via `sampling/createMessage`
2. **Client** (go-go-mcp) processes the request
3. **Client** can modify/validate the request (human-in-the-loop)
4. **Client** forwards to **Host** for actual LLM inference
5. **Host** returns completion to **Client**
6. **Client** returns result to **Server**

## Implementation Roadmap

### Phase 1: Core Sampling Infrastructure Enhancement

#### 1.1 Enhance Sampling Protocol Support
- [ ] **Update protocol structures** to match MCP 2025-03-26 specification
- [ ] **Add missing sampling parameters**:
  ```go
  type CreateMessageRequest struct {
      Messages         []Message        `json:"messages"`
      ModelPreferences ModelPreferences `json:"modelPreferences,omitempty"`
      SystemPrompt     string           `json:"systemPrompt,omitempty"`
      IncludeContext   string           `json:"includeContext,omitempty"` // "none" | "thisServer" | "allServers"
      Temperature      *float64         `json:"temperature,omitempty"`
      MaxTokens        int              `json:"maxTokens,omitempty"`
      StopSequences    []string         `json:"stopSequences,omitempty"`
      Metadata         map[string]interface{} `json:"metadata,omitempty"`
  }
  ```

#### 1.2 Implement Sampling Capability Declaration
- [ ] **Update client capabilities** to properly declare sampling support:
  ```go
  // pkg/protocol/initialization.go
  type SamplingCapability struct {
      // Add specific sampling capabilities
  }
  ```

#### 1.3 Add Sampling Request Handler
- [ ] **Create sampling request handler** in client:
  ```go
  // pkg/client/sampling.go
  func (c *Client) HandleSamplingRequest(ctx context.Context, request *protocol.CreateMessageRequest) (*protocol.Message, error)
  ```

### Phase 2: LLM Provider Integration

#### 2.1 Define LLM Provider Interface
- [ ] **Create LLM provider interface**:
  ```go
  // pkg/llm/provider.go
  type LLMProvider interface {
      CreateCompletion(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
      ListModels(ctx context.Context) ([]Model, error)
      GetModelCapabilities(ctx context.Context, modelName string) (*ModelCapabilities, error)
  }
  ```

#### 2.2 Implement OpenAI Provider
- [ ] **OpenAI API integration**:
  ```go
  // pkg/llm/openai/provider.go
  type OpenAIProvider struct {
      client *openai.Client
      config OpenAIConfig
  }
  ```

#### 2.3 Implement Anthropic Provider
- [ ] **Anthropic API integration**:
  ```go
  // pkg/llm/anthropic/provider.go
  type AnthropicProvider struct {
      client *anthropic.Client
      config AnthropicConfig
  }
  ```

#### 2.4 Add Provider Registry
- [ ] **LLM provider registry system**:
  ```go
  // pkg/llm/registry.go
  type ProviderRegistry struct {
      providers map[string]LLMProvider
  }
  ```

### Phase 3: Security and Human-in-the-Loop Controls

#### 3.1 Implement Request Validation
- [ ] **Add sampling request validation**:
  ```go
  // pkg/security/sampling.go
  func ValidateSamplingRequest(request *protocol.CreateMessageRequest) error
  func SanitizeMessages(messages []protocol.Message) []protocol.Message
  ```

#### 3.2 Add User Consent System
- [ ] **User consent interface**:
  ```go
  // pkg/consent/interface.go
  type ConsentManager interface {
      RequestSamplingConsent(ctx context.Context, request *SamplingConsentRequest) (*ConsentResponse, error)
      RequestPromptModification(ctx context.Context, prompt string) (string, bool, error)
  }
  ```

#### 3.3 Implement Rate Limiting
- [ ] **Rate limiting for sampling requests**:
  ```go
  // pkg/ratelimit/sampling.go
  type SamplingRateLimiter struct {
      limiter *rate.Limiter
      config  RateLimitConfig
  }
  ```

### Phase 4: Context Management

#### 4.1 Implement Context Inclusion Logic
- [ ] **Context aggregation system**:
  ```go
  // pkg/context/manager.go
  type ContextManager struct {
      servers map[string]*ServerContext
  }
  
  func (cm *ContextManager) GetContext(includeType string, serverID string) (*ContextData, error)
  ```

#### 4.2 Add Resource Context Integration
- [ ] **Resource-based context inclusion**:
  ```go
  // pkg/context/resources.go
  func (cm *ContextManager) AggregateResourceContext(serverIDs []string) (*ResourceContext, error)
  ```

### Phase 5: Configuration and Management

#### 5.1 Add LLM Configuration
- [ ] **Configuration structure for LLM providers**:
  ```yaml
  # config.yaml
  llm:
    providers:
      openai:
        api_key: "${OPENAI_API_KEY}"
        base_url: "https://api.openai.com/v1"
        default_model: "gpt-4"
      anthropic:
        api_key: "${ANTHROPIC_API_KEY}"
        default_model: "claude-3-sonnet"
    
    sampling:
      default_max_tokens: 1000
      rate_limit:
        requests_per_minute: 60
      security:
        require_user_consent: true
        sanitize_prompts: true
  ```

#### 5.2 Add CLI Commands
- [ ] **CLI commands for LLM management**:
  ```go
  // pkg/cmds/llm.go
  var llmCmd = &cobra.Command{
      Use:   "llm",
      Short: "Manage LLM providers and sampling",
  }
  
  var listProvidersCmd = &cobra.Command{
      Use:   "list-providers",
      Short: "List available LLM providers",
  }
  ```

### Phase 6: Testing and Validation

#### 6.1 Unit Tests
- [ ] **Comprehensive unit tests**:
  ```go
  // pkg/client/sampling_test.go
  func TestSamplingRequestHandling(t *testing.T)
  func TestContextInclusion(t *testing.T)
  func TestSecurityValidation(t *testing.T)
  ```

#### 6.2 Integration Tests
- [ ] **End-to-end integration tests**:
  ```go
  // tests/integration/sampling_test.go
  func TestMCPSamplingWorkflow(t *testing.T)
  func TestMultiProviderSampling(t *testing.T)
  ```

#### 6.3 Security Tests
- [ ] **Security validation tests**:
  ```go
  // tests/security/sampling_test.go
  func TestPromptInjectionPrevention(t *testing.T)
  func TestRateLimitEnforcement(t *testing.T)
  ```

## Implementation Details

### Core Sampling Flow Implementation

```go
// pkg/client/sampling.go
func (c *Client) HandleSamplingRequest(ctx context.Context, request *protocol.CreateMessageRequest) (*protocol.Message, error) {
    // 1. Validate request
    if err := c.validateSamplingRequest(request); err != nil {
        return nil, fmt.Errorf("invalid sampling request: %w", err)
    }
    
    // 2. Check user consent
    if c.config.RequireUserConsent {
        consent, err := c.consentManager.RequestSamplingConsent(ctx, request)
        if err != nil || !consent.Approved {
            return nil, fmt.Errorf("user consent required")
        }
    }
    
    // 3. Apply rate limiting
    if err := c.rateLimiter.Allow(); err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }
    
    // 4. Include context if requested
    if request.IncludeContext != "none" {
        context, err := c.contextManager.GetContext(request.IncludeContext, c.serverID)
        if err != nil {
            return nil, fmt.Errorf("failed to get context: %w", err)
        }
        request.Messages = c.mergeContext(request.Messages, context)
    }
    
    // 5. Select appropriate LLM provider
    provider, err := c.selectProvider(request.ModelPreferences)
    if err != nil {
        return nil, fmt.Errorf("failed to select provider: %w", err)
    }
    
    // 6. Make LLM request
    response, err := provider.CreateCompletion(ctx, &llm.CompletionRequest{
        Messages:      request.Messages,
        SystemPrompt:  request.SystemPrompt,
        Temperature:   request.Temperature,
        MaxTokens:     request.MaxTokens,
        StopSequences: request.StopSequences,
    })
    if err != nil {
        return nil, fmt.Errorf("LLM completion failed: %w", err)
    }
    
    // 7. Convert response to MCP format
    return &protocol.Message{
        Role:    "assistant",
        Content: protocol.MessageContent{
            Type: "text",
            Text: response.Content,
        },
        Model: response.Model,
    }, nil
}
```

### LLM Provider Implementation Example

```go
// pkg/llm/openai/provider.go
type OpenAIProvider struct {
    client *openai.Client
    config OpenAIConfig
}

func (p *OpenAIProvider) CreateCompletion(ctx context.Context, request *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    messages := make([]openai.ChatCompletionMessage, len(request.Messages))
    for i, msg := range request.Messages {
        messages[i] = openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content.Text,
        }
    }
    
    req := openai.ChatCompletionRequest{
        Model:       p.selectModel(request.ModelPreferences),
        Messages:    messages,
        Temperature: request.Temperature,
        MaxTokens:   request.MaxTokens,
        Stop:        request.StopSequences,
    }
    
    if request.SystemPrompt != "" {
        req.Messages = append([]openai.ChatCompletionMessage{{
            Role:    "system",
            Content: request.SystemPrompt,
        }}, req.Messages...)
    }
    
    resp, err := p.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return &llm.CompletionResponse{
        Content:    resp.Choices[0].Message.Content,
        Model:      resp.Model,
        StopReason: string(resp.Choices[0].FinishReason),
    }, nil
}
```

## Security Considerations

### 1. Input Validation and Sanitization
- Validate all incoming sampling requests
- Sanitize user prompts to prevent injection attacks
- Limit message content size and complexity

### 2. User Consent and Control
- Implement explicit user consent for all sampling requests
- Allow users to modify or reject prompts before processing
- Provide clear visibility into what data is being sent to LLMs

### 3. Rate Limiting and Resource Management
- Implement per-server and per-user rate limiting
- Monitor token usage and costs
- Set reasonable limits on context size and generation length

### 4. Data Privacy
- Ensure sensitive data is not inadvertently included in context
- Implement data masking for PII
- Provide audit logs for all sampling requests

## Best Practices

### 1. Error Handling
- Implement comprehensive error handling for all LLM provider failures
- Provide meaningful error messages to users
- Implement retry logic with exponential backoff

### 2. Monitoring and Observability
- Log all sampling requests and responses
- Monitor LLM provider performance and costs
- Track user consent patterns and security events

### 3. Configuration Management
- Use environment variables for sensitive configuration
- Provide sensible defaults for all settings
- Support hot-reloading of configuration changes

### 4. Testing Strategy
- Unit tests for all core components
- Integration tests with mock LLM providers
- Security tests for injection and abuse scenarios
- Performance tests for high-load scenarios

## Next Steps for Implementation

### Immediate Actions (Week 1-2)
1. **Enhance sampling protocol structures** to match latest MCP specification
2. **Implement basic LLM provider interface** and OpenAI provider
3. **Add sampling request validation** and basic security measures
4. **Create configuration structure** for LLM providers

### Short-term Goals (Week 3-4)
1. **Implement context inclusion logic** for "thisServer" and "allServers"
2. **Add user consent system** with CLI prompts
3. **Implement rate limiting** for sampling requests
4. **Add comprehensive error handling** and logging

### Medium-term Goals (Month 2)
1. **Add additional LLM providers** (Anthropic, local models)
2. **Implement advanced security features** (prompt sanitization, PII detection)
3. **Add monitoring and observability** features
4. **Create comprehensive test suite**

### Long-term Goals (Month 3+)
1. **Add web UI** for sampling management and monitoring
2. **Implement advanced context management** with caching
3. **Add support for streaming responses**
4. **Implement cost tracking and budgeting** features

## Key Resources

### Official Documentation
- [MCP Specification 2025-03-26](https://modelcontextprotocol.io/specification/2025-03-26/)
- [MCP Sampling Documentation](https://modelcontextprotocol.io/docs/concepts/sampling)
- [Anthropic MCP Documentation](https://docs.anthropic.com/en/docs/agents-and-tools/mcp)

### Implementation Examples
- [LangChain MCP Integration](https://github.com/langchain-ai/langchain-mcp-adapters)
- [FastMCP Python Library](https://github.com/jlowin/fastmcp)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)

### Current Codebase Files
- `pkg/client/client.go` - Main MCP client implementation
- `pkg/protocol/sampling.go` - Sampling protocol structures
- `pkg/protocol/initialization.go` - Capability negotiation
- `pkg/providers.go` - Provider interfaces

## Conclusion

The go-go-mcp project already has a solid foundation for implementing LLM inference via MCP sampling. The existing client infrastructure, protocol definitions, and transport layer provide a good starting point. The main work involves:

1. **Enhancing the sampling protocol** to match the latest specification
2. **Implementing LLM provider integrations** with proper abstraction
3. **Adding security and user control** mechanisms
4. **Implementing context management** for sophisticated agentic behaviors

By following this roadmap, the project can provide a robust, secure, and user-friendly MCP client with full LLM inference capabilities while maintaining the human-in-the-loop design principles that make MCP secure and trustworthy.

---

**Note**: Save all future research and implementation notes in `ttmp/YYYY-MM-DD/0X-XXX.md` format to maintain organized documentation. 