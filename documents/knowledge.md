# Reference

## Protocols & Libraries
- **MCP**: [Spec](https://github.com/machine-cognition-protocol/machine-cognition-protocol), [Go](https://github.com/mark3labs/mcp-go)
  - Tool: typed params, std req/resp, HTTP/WebSocket/stdio
- **LangChainGo**: [Repo](https://github.com/tmc/langchaingo), [Docs](https://pkg.go.dev/github.com/tmc/langchaingo/llms)
  - Standardized messages, tool/function calls, provider clients

## Code Patterns
### Tool Definition
```go
mcp.NewTool("example_tool",
  mcp.WithDescription("desc"),
  mcp.WithString("param1", mcp.Required(), mcp.Description("desc")),
  mcp.WithNumber("param2", mcp.Description("desc")),
)
```
### LLM Request
```go
resp, err := s.client.GenerateContent(ctx, messages, llms.WithTools(llmTools), llms.WithToolChoice("required"))
```
### Chat History
```go
c.history = append(c.history, llms.MessageContent{
  Role: llms.ChatMessageTypeAI,
  Parts: []llms.ContentPart{
    llms.ToolCall{ID: toolCall.ID, Type: toolCall.ToLLM().Type, FunctionCall: &llms.FunctionCall{Name: toolCall.ToLLM().FunctionCall.Name, Arguments: toolCall.ToLLM().FunctionCall.Arguments}},
  },
})
```
### Token Counting
```go
tokenCounter := utils.NewTokenCounter(logger, "")
tokens := tokenCounter.EstimateTokenCount(message)
```
### Error Handling
```go
err = error_handling.RetryWithBackoff(ctx, sendFn, error_handling.RetryConfig{...})
```

## Env Vars Reference
| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| Agent    | SPL_AGENT_NAME | Agent name | *req* |
|          | SPL_AGENT_VERSION | Version | 1.0.0 |
| Tool     | SPL_TOOL_NAME | Tool name | *req* |
|          | SPL_TOOL_DESCRIPTION | Desc | *req* |
|          | SPL_TOOL_ARGUMENT_NAME | Arg name | *req* |
|          | SPL_TOOL_ARGUMENT_DESCRIPTION | Arg desc | *req* |
| LLM      | SPL_LLM_PROVIDER | Provider | *req* |
|          | SPL_LLM_API_KEY | API key | *req* |
|          | SPL_LLM_MODEL | Model | *req* |
|          | SPL_LLM_MAX_TOKENS | Max tokens | 0 |
|          | SPL_LLM_TEMPERATURE | Temp | 0.7 |
|          | SPL_LLM_PROMPT_TEMPLATE | Prompt | *req* |
| Chat     | SPL_CHAT_MAX_TOKENS | Max hist tokens | 0 |
|          | SPL_CHAT_MAX_ITERATIONS | Max LLM iters | 25 |
|          | SPL_CHAT_REQUEST_BUDGET | Max cost per request | 0.0 |
| Retry    | SPL_LLM_RETRY_MAX_RETRIES | Max retries | 3 |
|          | SPL_LLM_RETRY_INITIAL_BACKOFF | Init backoff | 1.0 |
|          | SPL_LLM_RETRY_MAX_BACKOFF | Max backoff | 30.0 |
|          | SPL_LLM_RETRY_BACKOFF_MULTIPLIER | Multiplier | 2.0 |
| Runtime  | SPL_LOG_DEFAULT_LEVEL | Log default_level | info |

# Reference Patterns

## Go Patterns
### Tool Definition
```go
mcp.NewTool("example_tool",
  mcp.WithDescription("desc"),
  mcp.WithString("param1", mcp.Required(), mcp.Description("desc")),
  mcp.WithNumber("param2", mcp.Description("desc")),
)
```

### LLM Request
```go
resp, err := s.client.GenerateContent(ctx, messages, llms.WithTools(llmTools), llms.WithToolChoice("required"))
```

### Chat History
```go
c.history = append(c.history, llms.MessageContent{
  Role: llms.ChatMessageTypeAI,
  Parts: []llms.ContentPart{
    llms.ToolCall{ID: toolCall.ID, Type: toolCall.ToLLM().Type, FunctionCall: &llms.FunctionCall{Name: toolCall.ToLLM().FunctionCall.Name, Arguments: toolCall.ToLLM().FunctionCall.Arguments}},
  },
})
```

### Token Counting
```go
tokenCounter := utils.NewTokenCounter(logger, "")
tokens := tokenCounter.EstimateTokenCount(message)
```

### Error Handling
```go
err = error_handling.RetryWithBackoff(ctx, sendFn, error_handling.RetryConfig{...})
```

## Env Vars Reference
| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| Agent    | SPL_AGENT_NAME | Agent name | *req* |
|          | SPL_AGENT_VERSION | Version | 1.0.0 |
| Tool     | SPL_TOOL_NAME | Tool name | *req* |
|          | SPL_TOOL_DESCRIPTION | Desc | *req* |
|          | SPL_TOOL_ARGUMENT_NAME | Arg name | *req* |
|          | SPL_TOOL_ARGUMENT_DESCRIPTION | Arg desc | *req* |
| LLM      | SPL_LLM_PROVIDER | Provider | *req* |
|          | SPL_LLM_API_KEY | API key | *req* |
|          | SPL_LLM_MODEL | Model | *req* |
|          | SPL_LLM_MAX_TOKENS | Max tokens | 0 |
|          | SPL_LLM_TEMPERATURE | Temp | 0.7 |
|          | SPL_LLM_PROMPT_TEMPLATE | Prompt | *req* |
| Chat     | SPL_CHAT_MAX_TOKENS | Max hist tokens | 0 |
|          | SPL_CHAT_MAX_ITERATIONS | Max LLM iters | 25 |
|          | SPL_CHAT_REQUEST_BUDGET | Max cost per request | 0.0 |
| Retry    | SPL_LLM_RETRY_MAX_RETRIES | Max retries | 3 |
|          | SPL_LLM_RETRY_INITIAL_BACKOFF | Init backoff | 1.0 |
|          | SPL_LLM_RETRY_MAX_BACKOFF | Max backoff | 30.0 |
|          | SPL_LLM_RETRY_BACKOFF_MULTIPLIER | Multiplier | 2.0 |
| Runtime  | SPL_LOG_DEFAULT_LEVEL | Log default_level | info |
