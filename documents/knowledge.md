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
| Tool     | SPL_AGENT_TOOL_NAME | Tool name | *req* |
|          | SPL_AGENT_TOOL_DESCRIPTION | Desc | *req* |
|          | SPL_AGENT_TOOL_ARGUMENTNAME | Arg name | *req* |
|          | SPL_AGENT_TOOL_ARGUMENTDESCRIPTION | Arg desc | *req* |
| LLM      | SPL_AGENT_LLM_PROVIDER | Provider | *req* |
|          | SPL_AGENT_LLM_APIKEY | API key | *req* |
|          | SPL_AGENT_LLM_MODEL | Model | *req* |
|          | SPL_AGENT_LLM_MAXTOKENS | Max tokens | 0 |
|          | SPL_AGENT_LLM_TEMPERATURE | Temp | 0.7 |
|          | SPL_AGENT_LLM_PROMPTTEMPLATE | Prompt | *req* |
| Chat     | SPL_AGENT_CHAT_MAXTOKENS | Max hist tokens | 0 |
|          | SPL_AGENT_CHAT_MAXLLMITERATIONS | Max LLM iters | 25 |
|          | SPL_AGENT_CHAT_REQUESTBUDGET | Max cost per request | 0.0 |
| Retry    | SPL_AGENT_LLM_RETRY_MAXRETRIES | Max retries | 3 |
|          | SPL_AGENT_LLM_RETRY_INITIALBACKOFF | Init backoff | 1.0 |
|          | SPL_AGENT_LLM_RETRY_MAXBACKOFF | Max backoff | 30.0 |
|          | SPL_AGENT_LLM_RETRY_BACKOFFMULTIPLIER | Multiplier | 2.0 |
| Runtime  | SPL_RUNTIME_LOG_DEFAULTLEVEL | Log defaultLevel | info |

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
| Tool     | SPL_AGENT_TOOL_NAME | Tool name | *req* |
|          | SPL_AGENT_TOOL_DESCRIPTION | Desc | *req* |
|          | SPL_AGENT_TOOL_ARGUMENTNAME | Arg name | *req* |
|          | SPL_AGENT_TOOL_ARGUMENTDESCRIPTION | Arg desc | *req* |
| LLM      | SPL_AGENT_LLM_PROVIDER | Provider | *req* |
|          | SPL_AGENT_LLM_APIKEY | API key | *req* |
|          | SPL_AGENT_LLM_MODEL | Model | *req* |
|          | SPL_AGENT_LLM_MAXTOKENS | Max tokens | 0 |
|          | SPL_AGENT_LLM_TEMPERATURE | Temp | 0.7 |
|          | SPL_AGENT_LLM_PROMPTTEMPLATE | Prompt | *req* |
| Chat     | SPL_AGENT_CHAT_MAXTOKENS | Max hist tokens | 0 |
|          | SPL_AGENT_CHAT_MAXLLMITERATIONS | Max LLM iters | 25 |
|          | SPL_AGENT_CHAT_REQUESTBUDGET | Max cost per request | 0.0 |
| Retry    | SPL_AGENT_LLM_RETRY_MAXRETRIES | Max retries | 3 |
|          | SPL_AGENT_LLM_RETRY_INITIALBACKOFF | Init backoff | 1.0 |
|          | SPL_AGENT_LLM_RETRY_MAXBACKOFF | Max backoff | 30.0 |
|          | SPL_AGENT_LLM_RETRY_BACKOFFMULTIPLIER | Multiplier | 2.0 |
| Runtime  | SPL_RUNTIME_LOG_DEFAULTLEVEL | Log defaultLevel | info |
