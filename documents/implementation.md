# Implementation

## Core Components
- **Agent (internal/agent)**: Core agent logic only. Orchestrates request lifecycle, tool exec, state, logs tool calls as `>> Execute tool toolName(args)`. No config loading, server, CLI, or direct call JSON types. Exposes a clean interface for use by the app layer.
- **App MCP (internal/app_mcp)**: Application wiring for MCP server/daemon mode. Instantiates and manages the agent, provides CLI entry points. Implements the shared `Application` interface.
- **App Direct (internal/app_direct)**: Application wiring for direct CLI call mode. Instantiates and manages the agent for direct call mode. Implements the shared `Application` interface.
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
        timeout: 10   # Tool call timeout in seconds (optional, default 30s if not set)
      filesystem:
        command: "mcp-filesystem-server"
        args: ["/path/to/directory"]
        # timeout: 60
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

## MCPConnector
- Now supports per-server tool call timeout: each MCP server in config can specify a `timeout` (seconds, float or int). If not set, defaults to 30s.
- Timeout is loaded from YAML/JSON, merged in `Apply`, copied in `GetMCPConnectorConfig`, and enforced in `MCPConnector.ExecuteTool`.
- Manual timeout logic replaces context.WithTimeout for better control and logging. Enhanced logging for tool execution, including timeout/cancellation details.
- Comprehensive tests added for timeout propagation and enforcement.

## Logger
- После загрузки конфигурации фабрика NewLogger выбирает реализацию по runtime.log.output ("mcp" или "stderr").
- MCPLogger: пишет только через MCP-нотификации (notifications/message), stderr не используется.
- StderrLogger: пишет только в stderr, MCP не используется.
- IOWriterLogger: пишет только в stderr, MCP не используется.
- Дублирования логов нет: каждая реализация отвечает только за свой канал.
- По-умолчанию используется MCPLogger (output = "mcp").
- MCPLogger интегрируется с MCPServer через интерфейс MCPServerNotifier (SetMCPServer принимает MCPServerNotifier, а не конкретный тип сервера).
- Все MCP-логи содержат delivered_to_client=true.
- Все тесты покрывают оба варианта, проверяют отсутствие дублирования, работу с уровнями, edge-cases, работу с полями.
- Для тестирования MCPLogger используются моки MCPServerNotifier.
- Logger respects config log default_level (runtime.log.default_level).

## File Removals
- Deleted: `internal/app/direct_app_test.go`, `internal/app/direct_types.go`, `internal/app/util.go`, `site/examples/ai-news-subagent-extractor.yaml` (obsolete, replaced by `text-extractor.yaml`).
- Tests and code referencing these files removed or updated.

## Test Coverage
- Added/updated tests for:
  - Per-server timeout propagation and enforcement (YAML/JSON loader, config manager, MCPConnector).
  - Logger respects config log level.
  - Removal of obsolete files and references.

## Direct Call Mode (CLI)
- **Flag:** `--call` (string, user query)
- **Usage:** `./bin/speelka-agent --config config.yaml --call 'What is the weather?'
- **Behavior:** Runs agent in single-shot mode, outputs structured JSON to stdout. Uses `internal/app.DirectApp` (independent from `App`, wires up agent and dependencies for direct call mode).
- **Output Example:**
  ```json
  {
    "success": true,
    "result": { "answer": "The weather is sunny." },
    "meta": { "tokens": 42, "cost": 0.01, "duration_ms": 1234 },
    "error": { "type": "", "message": "" }
  }
  ```
- **Error Handling:** All errors are mapped to JSON output and exit codes:
  - `0`: success
  - `1`: user/config error
  - `2`: internal/agent/LLM/tool error
- **Implementation:** Uses `DirectApp`, reuses agent core/config/env logic.
- **Use Cases:** Scripting, automation, debugging, CI integration.

## Shell/Integration Test Plan: run Script check Sequence

### Purpose
To verify that the `./run check` sequence:
- Prints a clear error message and exits before the success message if any step fails.
- Prints the final success message only if all steps succeed.

### Test Cases

1. **Simulate Failure in a Step**
   - Temporarily modify one of the steps (e.g., `./run lint`) to return a non-zero exit code.
   - Run `./run check`.
   - **Expected:**
     - The script prints an error message indicating which step failed (e.g., `Lint failed`).
     - The script exits before printing `✓ All checks passed!`.

2. **All Steps Succeed**
   - Ensure all steps (`build`, `lint`, `test`, `call`, `call-multistep`, `test-direct-call`) succeed.
   - Run `./run check`.
   - **Expected:**
     - The script prints all intermediate step outputs.
     - The script prints `✓ All checks passed!` at the end.

### Implementation Notes
- Use explicit error handling in the `check` block for each step.
- For CI, capture and assert on the presence/absence of the success message in the output.
- Document any temporary modifications for failure simulation and revert after testing.

## MCP Protocol Logging (MCP-логирование)

### Server
- Declares `logging` capability in handshake.
- Sends logs to client via `notifications/message` (level, logger, data).
- Handles `logging/setLevel` from client to change log level.
- All logs go through centralized logger (logrus), marked as delivered to client.
- No duplication to stderr.

### Client
- Handles `notifications/message` (parsing, filtering, output in CLI/UI).
- Sends `logging/setLevel` (tool call) to server.
- (Optionally) Duplicates MCP logs to local log.
- Handles `logging/setLevel` from client to change log default_level.
- Client can change log default_level.

### Security
- Secrets/PII must not be logged (responsibility of business logic).

### Testing
- **Server:**
  - Unit/integration tests: sending notifications, structure, filtering by level, negative test for secrets/PII.
- **Client:**
  - Unit tests: parsing notifications, filtering, output, sending `logging/setLevel`.
- All tests pass except negative test for secrets (requires explicit filtering logic).

Все тесты для MCP-логирования реализованы и проходят (см. internal/mcp_server/mcp_server_test.go и cmd/mcp-call/main_test.go). Покрытие: отправка, фильтрация, структура, смена уровня, защита от утечек.

### Definition of Done
- Server and client exchange structured MCP logs.
- Client can change log level.
- All changes covered by tests.
- See `architecture.md` for high-level design.