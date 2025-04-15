# Implementation

## Core Components
- **Agent**: Orchestrates request lifecycle, tool exec, state, logs tool calls as `>> Execute tool toolName(args)`
- **Chat**: Manages history, token/cost, compaction; config immutable (constructor only); all state in `chatInfo` struct; supports multiple compaction strategies; token/cost tracked via LLMResponse, fallback estimation if needed
- **TokenCounter**: Approximates tokens (4 chars ≈ 1 token), type-specific, fallback for unknowns
- **Compaction**: Pluggable strategies (DeleteOld, DeleteMiddle, PartialSummary, NoCompaction); preserves system prompt, context; auto-triggers on token limit
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
    compaction_strategy: "delete-old"
    max_llm_iterations: 25
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

## Compaction Strategies
| Strategy        | Preserves         | Removes         | Use Case                |
|-----------------|------------------|-----------------|-------------------------|
| DeleteOld       | System prompt, recent | Oldest first   | General, recent context |
| DeleteMiddle    | System, first/last N | Middle         | Early/recent context    |
| PartialSummary  | System, recent    | Middle (summary)| LLM summary (future)    |
| NoCompaction    | All               | None            | Small context           |

## Token Counting
- 4 chars ≈ 1 token (fallback)
- Type-specific for text/tool calls
- Overhead for message format

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

## Test Coverage
- Chat: All `chatInfo` fields, compaction, token/cost/approximation, edge cases
- LLM Service: All `LLMResponse` fields, mock LLM, logger
- Config: Defaults, overrides, validation, transport
- E2E: Agent, transport, tools, token/cost

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
```