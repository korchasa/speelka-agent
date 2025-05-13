# Implementation

## Core Components
- **Agent**: Orchestrates LLM loop, tool execution, chat state. No config/server/CLI logic. Exposes interface for app layer.
- **App MCP**: MCP server/daemon wiring. Instantiates agent, provides CLI/server entrypoints.
- **App Direct**: Direct CLI call wiring. Instantiates agent for single-shot mode.
- **Chat**: Manages history, token/cost tracking, enforces request budget. All state in `chatInfo` struct.
- **Config Manager**: Loads/validates config (YAML, JSON, env), type-safe, strict validation.
- **LLM Service**: Integrates LLM providers, returns structured responses, retry/backoff logic.
- **MCP Server**: HTTP/stdio, routes requests, real-time SSE.
- **MCP Connector**: Manages external MCP servers, tool discovery, per-server timeouts.
- **Logger**: Centralized logging (logrus/MCP), level mapping, client notifications.

## Configuration Example (YAML)
```yaml
runtime:
  log:
    default_level: info
    output: mcp
  transports:
    stdio:
      enabled: true
    http:
      enabled: false
      host: localhost
      port: 3000
agent:
  name: "speelka-agent"
  tool:
    name: "process"
    description: "Process tool for user queries"
    argument_name: "input"
    argument_description: "User query"
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
        timeout: 10
      filesystem:
        command: "mcp-filesystem-server"
        args: ["/path/to/directory"]
```

## Token Counting
- 4 chars ≈ 1 token (fallback)
- Type-specific for text/tool calls
- Cumulative for session, never decreases

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
- Retry/backoff per config
- No panics, safe assertions, descriptive errors
- Orphaned tool calls auto-removed and logged

## Test Coverage
- Unit: All core logic, edge cases
- Integration: LLM, config, transport, logger
- E2E: Agent, transport, tools, token/cost
- Orphaned tool_call detection: Simulated and auto-cleaned

## Example Env Vars
```env
SPL_AGENT_NAME="speelka-agent"
SPL_TOOL_NAME="process"
SPL_LLM_PROVIDER="openai"
SPL_LLM_API_KEY="your_api_key_here"
SPL_LLM_MODEL="gpt-4o"
SPL_LLM_MAX_TOKENS=0
SPL_LLM_TEMPERATURE=0.7
SPL_LLM_RETRY_MAX_RETRIES=3
SPL_LLM_RETRY_INITIAL_BACKOFF=1.0
SPL_LLM_RETRY_MAX_BACKOFF=30.0
SPL_LLM_RETRY_BACKOFF_MULTIPLIER=2.0
SPL_CHAT_REQUEST_BUDGET=0.0
```

## Direct Call Mode
- `--call` flag: single-shot agent run, outputs structured JSON to stdout
- All errors mapped to JSON and exit codes (0: success, 1: user/config, 2: internal/tool)
- Use cases: scripting, automation, CI

## Logging
- Centralized logger (logrus/MCP)
- No log duplication
- Dynamic log level via protocol
- No secrets/PII in logs

// See architecture.md for high-level design.