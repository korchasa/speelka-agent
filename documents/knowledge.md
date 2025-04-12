# Reference Materials

## Protocols & Libraries

### Model Context Protocol (MCP)
- [MCP GitHub](https://github.com/machine-cognition-protocol/machine-cognition-protocol)
- [MCP-Go Library](https://github.com/mark3labs/mcp-go)

**Key Concepts**:
- **Tool Definition**: Structured tool definitions with typed parameters
- **Request/Response**: Standard tool invocation format
- **Transport Agnostic**: Supports HTTP, WebSockets, stdio

### LangChain Go
- [LangChainGo GitHub](https://github.com/tmc/langchaingo)
- [LLM Integration Docs](https://pkg.go.dev/github.com/tmc/langchaingo/llms)

**Features**:
- Message format standardization
- Tool/function calling support
- Provider-specific clients

## Code Examples

### Tool Definition
```go
// Tool definition example
tool := mcp.NewTool("example_tool",
    mcp.WithDescription("An example tool"),
    mcp.WithString("param1",
        mcp.Required(),
        mcp.Description("First parameter"),
    ),
    mcp.WithNumber("param2",
        mcp.Description("Second parameter"),
    ),
)
```

### LLM Request
```go
// LLM request example
response, err := s.client.GenerateContent(
    ctx,
    messages,
    llms.WithTools(llmTools),
    llms.WithToolChoice("required")
)
```

### Chat History Management
```go
// Adding tool call to history
c.history = append(c.history, llms.MessageContent{
    Role: llms.ChatMessageTypeAI,
    Parts: []llms.ContentPart{
        llms.ToolCall{
            ID:   toolCall.ID,
            Type: toolCall.ToLLM().Type,
            FunctionCall: &llms.FunctionCall{
                Name:      toolCall.ToLLM().FunctionCall.Name,
                Arguments: toolCall.ToLLM().FunctionCall.Arguments,
            },
        },
    },
})
```

### Error Handling Patterns
```go
// Retry with backoff
err = error_handling.RetryWithBackoff(ctx, sendFn, error_handling.RetryConfig{
    MaxRetries:        s.config.RetryConfig.MaxRetries,
    InitialBackoff:    time.Duration(s.config.RetryConfig.InitialBackoff * float64(time.Second)),
    BackoffMultiplier: s.config.RetryConfig.BackoffMultiplier,
    MaxBackoff:        time.Duration(s.config.RetryConfig.MaxBackoff * float64(time.Second)),
})
```

## Environment Variables Reference

| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| **Agent** | `SPL_AGENT_NAME` | Name of the agent | *Required* |
| | `SPL_AGENT_VERSION` | Version of the agent | "1.0.0" |
| **Tool** | `SPL_TOOL_NAME` | Name of the tool | *Required* |
| | `SPL_TOOL_DESCRIPTION` | Description of the tool | *Required* |
| | `SPL_TOOL_ARGUMENT_NAME` | Argument name | *Required* |
| | `SPL_TOOL_ARGUMENT_DESCRIPTION` | Argument description | *Required* |
| **LLM** | `SPL_LLM_PROVIDER` | Provider (openai, anthropic) | *Required* |
| | `SPL_LLM_API_KEY` | API key | *Required* |
| | `SPL_LLM_MODEL` | Model name | *Required* |
| | `SPL_LLM_MAX_TOKENS` | Max output tokens | 0 (no limit) |
| | `SPL_LLM_TEMPERATURE` | Temperature for sampling | 0.7 |
| | `SPL_LLM_PROMPT_TEMPLATE` | System prompt template | *Required* |
| **Retry** | `SPL_LLM_RETRY_MAX_RETRIES` | Max retry attempts | 3 |
| | `SPL_LLM_RETRY_INITIAL_BACKOFF` | Initial backoff (seconds) | 1.0 |
| | `SPL_LLM_RETRY_MAX_BACKOFF` | Max backoff (seconds) | 30.0 |
| | `SPL_LLM_RETRY_BACKOFF_MULTIPLIER` | Backoff multiplier | 2.0 |
| **Runtime** | `SPL_LOG_LEVEL` | Log level | "info" |
| | `SPL_LOG_OUTPUT` | Log destination | "stderr" |
| | `SPL_RUNTIME_STDIO_ENABLED` | Enable stdio transport | true |
| | `SPL_RUNTIME_HTTP_ENABLED` | Enable HTTP transport | false |
| | `SPL_RUNTIME_HTTP_HOST` | HTTP host | "localhost" |
| | `SPL_RUNTIME_HTTP_PORT` | HTTP port | 3000 |

## Command Reference

```bash
# Build agent
./run build

# Run daemon mode
./run start

# Run CLI mode
./run cli

# Run tests
./run test

# Development commands
./run dev         # Run in development mode
./run lint        # Run code linting
./run check       # Run all checks (test, lint, build)

# Test commands
./run call                # Test with simple query
./run complex-call        # Test with complex query
./run call-news           # Test news agent
./run fetch_url <url>     # Fetch URL using MCP

# Inspection
./run inspect     # Inspect project with MCP inspector
```

## System Prompt Template Example
```
You are a useful AI agent who can use tools to accomplish tasks. Your primary goal is to assist the user with their request.

Available tools:
{{ tools }}

User request: {{ input }}
```
