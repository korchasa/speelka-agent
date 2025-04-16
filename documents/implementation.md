# Implementation

## Core Components
- **Agent**: Orchestrates request lifecycle, tool exec, state, logs tool calls as `>> Execute tool toolName(args)`
- **Chat**: Manages history, token/cost; config immutable (constructor only); all state in `chatInfo` struct; token/cost tracked via LLMResponse, fallback estimation if needed. **TotalTokens** and **TotalCost** are cumulative (monotonically increasing) and never decrease. Chat history is not compacted or compressed.
- **TokenCounter**: Approximates tokens (4 chars ≈ 1 token), type-specific, fallback for unknowns
- **Config Manager**: Loads/validates config (YAML, JSON, env), type-safe, strict validation, only `Apply` parses log/output
- **LLM Service**: Integrates OpenAI/Anthropic, returns `LLMResponse` (text, tool calls, token/cost, duration), retry/backoff logic
- **MCP Server**: HTTP/stdio, routes requests, SSE for real-time
- **MCP Connector**: Manages external MCP servers, tool discovery, retry
- **Logger**: Wraps logrus, MCP protocol, level mapping, client notifications

## Config Structure (YAML)
```yaml
runtime:
  log:
    level: debug
    output: ./simple.log
  transports:
    stdio:
      enabled: true
      buffer_size: 8192
    http:
      enabled: false
      host: localhost
      port: 3000
agent:
  name: "simple-speelka-agent"
  tool:
    name: "process"
    description: "Process tool for handling user queries with LLM"
    argument_name: "input"
    argument_description: "The user query to process"
  chat:
    max_tokens: 0
    max_llm_iterations: 25
    request_budget: 0.0
  llm:
    provider: "openai"
    api_key: "dummy-api-key"
    model: "gpt-4o"
    temperature: 0.7
    prompt_template: "You are a helpful assistant. {{input}}. Available tools: {{tools}}"
  connections:
    mcpServers:
      time:
        command: "docker"
        args: ["run", "-i", "--rm", "mcp/time"]
      filesystem:
        command: "mcp-filesystem-server"
        args: ["/path/to/directory"]
```

## Token Counting
- 4 chars ≈ 1 token (fallback)
- Type-specific for text/tool calls
- Overhead for message format
- **TotalTokens** and **TotalCost** are cumulative for the session and never decrease.

## Request Processing
1. Receive (HTTP/stdio)
2. Validate, create context
3. Init chat, discover tools
4. LLM prompt, parse response, extract tool calls
5. Tool exec, capture result
6. Format/send response

## Config Loading Hierarchy
1. CLI args
2. Env vars (SPL_ prefix)
3. Config file
4. Defaults

## Error Handling
- Categories: Validation, Transient, Internal, External
- Retry: Backoff config
- Safe assertions, descriptive errors, no panics
- **Orphaned tool_call auto-cleanup:** If a tool_call is found in the message stack without a matching tool result (e.g., due to error or interruption), it is now automatically removed from the stack and a warning is logged. This prevents protocol errors and improves robustness.

## Test Coverage
- Chat: All `chatInfo` fields, token/cost/approximation, edge cases
- LLM Service: All `LLMResponse` fields, mock LLM, logger
- Config: Defaults, overrides, validation, transport
- E2E: Agent, transport, tools, token/cost
- **Orphaned tool_call detection:** Tests simulate a tool_call without a result and verify that the system detects and auto-cleans it, logging a warning.

## Example Env Vars
```env
SPL_AGENT_NAME="architect-speelka-agent"
SPL_TOOL_NAME="architect"
SPL_LLM_PROVIDER="openai"
SPL_LLM_API_KEY="your_api_key_here"
SPL_LLM_MODEL="gpt-4o"
SPL_LLM_MAX_TOKENS=0
SPL_LLM_TEMPERATURE=0.2
SPL_LLM_RETRY_MAX_RETRIES=3
SPL_LLM_RETRY_INITIAL_BACKOFF=1.0
SPL_LLM_RETRY_MAX_BACKOFF=30.0
SPL_LLM_RETRY_BACKOFF_MULTIPLIER=2.0
SPL_CHAT_REQUEST_BUDGET=0.0
```

## MCPServerConnection Tool Filtering

- The fields `IncludeTools` and `ExcludeTools` in `MCPServerConnection` allow fine-grained control over which tools are available from each MCP server.
- These fields can be set in YAML/JSON configuration files, and now also via environment variables:
  - `SPL_MCPS_<N>_INCLUDE_TOOLS` (comma- or space-separated list)
  - `SPL_MCPS_<N>_EXCLUDE_TOOLS` (comma- or space-separated list)
- The configuration overlay logic (Apply) merges these fields as follows:
  - If the overlay config provides a non-nil value for `IncludeTools` or `ExcludeTools`, it replaces the previous value.
  - If the overlay config provides nil, the previous value is preserved.
- Comprehensive tests cover all edge cases for loading and merging these fields.