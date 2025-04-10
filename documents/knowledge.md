# Knowledge References

## Model Context Protocol (MCP)
- [MCP GitHub](https://github.com/machine-cognition-protocol/machine-cognition-protocol)
- [MCP-Go Library](https://github.com/mark3labs/mcp-go)

**Key Concepts**:
- **Tool Definition**: Structured tool def with params
- **Request/Response**: Standard tool invocation format
- **Transport Agnostic**: HTTP, WebSockets, stdio

## LangChain Go
- [LangChainGo GitHub](https://github.com/tmc/langchaingo)
- [LLM Integration Docs](https://pkg.go.dev/github.com/tmc/langchaingo/llms)

**Features**:
- Message format standardization
- Tool/function calling support
- Provider-specific clients

## Code Snippets

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

## Configuration Examples

### LLM Configuration
```
LLM_PROVIDER=openai
LLM_MODEL=gpt-4
LLM_API_KEY=<your-api-key>
LLM_MAX_TOKENS=4096
LLM_TEMPERATURE=0.7
LLM_RETRY_MAX=3
LLM_RETRY_INITIAL_BACKOFF=1
LLM_RETRY_MAX_BACKOFF=60
LLM_RETRY_BACKOFF_MULTIPLIER=2
```

### MCP Server Configuration
```
MCP_SERVER_NAME=speelka-agent
MCP_SERVER_VERSION=1.0.0
MCP_SERVER_DEBUG=true
MCP_SERVER_HTTP_ENABLED=true
MCP_SERVER_HTTP_HOST=localhost
MCP_SERVER_HTTP_PORT=3000
MCP_SERVER_STDIO_ENABLED=true
MCP_SERVER_STDIO_BUFFER_SIZE=4096
MCP_SERVER_STDIO_AUTO_DETECT=true
```

### MCP Connector Configuration
```
MCP_CONNECTOR_SERVER_0_ID=playwright
MCP_CONNECTOR_SERVER_0_TRANSPORT=stdio
MCP_CONNECTOR_SERVER_0_COMMAND=mcp-playwright
MCP_CONNECTOR_SERVER_0_ENV_NODE_ENV=production
```

## Error Handling Patterns
```go
// Retry with backoff
err = error_handling.RetryWithBackoff(ctx, sendFn, error_handling.RetryConfig{
    MaxRetries:        s.config.RetryConfig.MaxRetries,
    InitialBackoff:    time.Duration(s.config.RetryConfig.InitialBackoff * float64(time.Second)),
    BackoffMultiplier: s.config.RetryConfig.BackoffMultiplier,
    MaxBackoff:        time.Duration(s.config.RetryConfig.MaxBackoff * float64(time.Second)),
})
```

## Useful Commands
```bash
# Build agent
./run build

# Run daemon mode
./run start

# Run CLI mode
./run cli

# Run tests
./run test
```

## System Prompt Template Example
```
You are a useful AI agent who can use tools to accomplish tasks. Your primary goal is to assist the user with their request.

Available tools:
{{ tools }}

User request: {{ input }}
```
