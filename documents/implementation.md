# Implementation

## Core Components
- **Agent**: Orchestrates LLM loop, tool execution, chat state. No config/server/CLI logic. Exposes interface for app layer.
- **App MCP**: MCP server/daemon wiring. Instantiates agent, provides CLI/server entrypoints.
- **App Direct**: Direct CLI call wiring. Instantiates agent for single-shot mode. Implements NewAgentCLI and dummyToolConnector, does not depend on app_mcp.
    - `app.go`: CLI application, contains NewAgentCLI and dummyToolConnector
    - `types.go`: Types for CLI mode
- **Chat**: Manages history, token/cost tracking, enforces request budget. All state in `chatInfo` struct.
- **Config Manager**: Loads/validates config (YAML, JSON, env), type-safe, strict validation.
- **LLM Service**: Integrates LLM providers, returns structured responses, retry/backoff logic.
- **MCP Server**: HTTP/stdio, routes requests, real-time SSE.
- **MCP Connector**: Manages external MCP servers, tool discovery, per-server timeouts.
- **Logger**: Centralized logging (logrus/MCP), level mapping, client notifications, flexible output and format.

## Example Configuration (YAML)
```yaml
runtime:
  log:
    defaultLevel: info
    output: ':mcp:'
    format: json
  transports:
    stdio:
      enabled: true
      buffer_size: 1024
    http:
      enabled: false
      host: localhost
      port: 3000
agent:
  name: "speelka-agent"
  version: "v1.0.0"
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
    apiKey: "dummy-api-key"
    model: "gpt-4o"
    temperature: 0.7
    promptTemplate: "You are a helpful assistant. {{input}}. Available tools: {{tools}}"
    retry:
      max_retries: 3
      initial_backoff: 1.0
      max_backoff: 30.0
      backoff_multiplier: 2.0
  connections:
    mcpServers:
      time:
        command: "docker"
        args: ["run", "-i", "--rm", "mcp/time"]
        timeout: 10
      filesystem:
        command: "mcp-filesystem-server"
        args: ["/path/to/directory"]
    retry:
      max_retries: 2
      initial_backoff: 1.5
      max_backoff: 10.0
      backoff_multiplier: 2.5
```

## Log Configuration
- **LogConfig**: Centralized structure for log management.
    - `DefaultLevel`: log level string (info, debug, warn, error, etc.)
    - `Output`: output destination (`:stdout:`, `:stderr:`, `:mcp:`, file path)
    - `Format`: log format (`custom`, `json`, `text`, `unknown`)
    - `UseMCPLogs`: flag for MCP logging
    - Constants: `LogOutputStdout`, `LogOutputStderr`, `LogOutputMCP`
- **LoggerSpec**: Extended logger interface with SetFormatter, SetMCPServer (via MCPServerNotifier)
- **MCPServerNotifier**: Interface for sending MCP notifications from the logger

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

## Config Loading (Koanf-based)
- All configuration is now loaded, merged, and validated using the koanf library (core + providers: file, env, confmap, structs; parsers: json, yaml, toml).
- The Manager loads defaults via confmap.Provider, then overlays file (yaml/json/toml) via file.Provider, then overlays env via env.Provider (SPL_ prefix, custom path mapping).
- All config structs use only `koanf` tags; `json`/`yaml` tags are removed.
- No custom loaders remain; all merging and env parsing is handled by koanf best practices.
- See `internal/configuration/manager.go` for implementation.

## Config Loading Hierarchy
1. Defaults (confmap)
2. Config file (yaml/json)
3. Env vars (SPL_ prefix)

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
- **BuildLogConfig**: tests for all output/format/level variants, including invalid values
- **GetAgentConfig, GetLLMConfig, GetMCPServerConfig, GetMCPConnectorConfig**: tests for correct config mapping
- **Golden serialization tests**: compare config structure serialization with golden file
- **Overlay property-based tests**: verify correct config overlay (edge-cases, map merge, zero-value preservation)

## Example Env Vars
```env
SPL_AGENT_NAME="speelka-agent"
SPL_TOOL_NAME="process"
SPL_LLM_PROVIDER="openai"
SPL_LLM_APIKEY="your_api_key_here"
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
- Dynamic log level and format via config
- Output: stdout, stderr, file, or MCP protocol
- No log duplication
- No secrets/PII in logs
- **LoggerSpec**: extended logger interface with SetFormatter, SetMCPServer (via MCPServerNotifier)
- **MCPServerNotifier**: interface for sending MCP notifications from the logger

## Direct-call MCP logging

In direct-call (CLI) mode, all MCP logs (notifications/message) are routed to stderr using a stub implementation of MCPServerNotifier (`mcpLogStub`). This stub is set in `app_direct.NewDirectApp` and prints logs in the format `[MCP level] message` for the user. This ensures:

- No conditional logic in main.go for logging mode.
- No empty `mcp` file is created.
- All MCP logs are visible to CLI users.

The logger is always created according to the configuration, and the stub is injected only for direct-call mode inside the application layer.

// See architecture.md for high-level design.

# MCPConnector Fallback Logging Implementation

## Functionality
- MCPConnector determines MCP logging support via capabilities after initialize.
- If logging is supported — subscribes to notifications/message.
- If not — fallback: reads stderr of the child process (only for stdio servers).
- All log routes are managed via LogConfig and support dynamic format and level changes.

## Test Examples
- MCP log routing test:
  - capabilities.Logging != nil
  - Simulate MCP log (info/debug/error) — logger receives message with prefix [MCP ...]
- Fallback to stderr test:
  - capabilities.Logging == nil
  - Simulate line in stderr — logger receives message with prefix stderr
- Helper functions for testing are in internal/mcp_connector/utils_test.go
- BuildLogConfig tests: all output/format/level variants, including invalid values
- Golden serialization tests: compare config structure serialization with golden file
- Overlay property-based tests: edge-cases, map merge, zero-value preservation

## Test Environment
- All tests run via ./run test
- For linter/build check: ./run check
- Mocks: mockLogger, fakeMCPClient (see internal/mcp_connector/mcp_connector_test.go)

## Important Details
- For HTTP servers, fallback is not possible (no access to stderr).
- For stdio servers, fallback is implemented via a separate goroutine and bufio.Scanner.
- All changes are covered by unit tests.

## Overlay and Config Compatibility

### Property-based overlay tests
- Goal: ensure correct overlay for any value combinations.
- Uses `testing/quick` to generate random config pairs.
- Checks:
  - overlay does not overwrite default values with zero-value fields;
  - correctly merges maps;
  - does not lose values;
  - edge-cases: empty strings, zeros, nil maps, partially filled structs.
- Test: `TestConfiguration_Overlay_PropertyBased` (`internal/types/configuration_test.go`).

### Golden compatibility tests
- Goal: control serialization compatibility of `types.Configuration` structure.
- Test serializes default config and compares with golden file.
- On structure change, test signals incompatibility.
- Test: `TestConfiguration_Serialization_Golden` (`internal/types/configuration_test.go`).

## New MCPServer Test Cases
- Thread safety check: concurrent Stop and Serve calls do not cause races or panics.
- Error logging check: BroadcastNotification logs an error if sending a notification fails (uses mock interface notificationBroadcasterWithError).
- Tool consistency check: the set of tools from GetAllTools matches those actually registered on the server.
- Log filtering and secret passing check: logger does not filter PII/secrets, responsibility is on business logic.

## Mock Interfaces for Testing
- notificationBroadcaster: allows substituting the internal MCP server to check notification sending.
- notificationBroadcasterWithError: extends notificationBroadcaster to simulate errors when sending notifications.
- errorCatchingLogger: records the fact of error logging to check BroadcastNotification behavior.

## Example of Using Mocks
In the BroadcastNotification_LogsError test, MCPServer replaces the server field with mockServerWithError and the logger with errorCatchingLogger. This allows checking that the error is actually logged without affecting production code.
// exitTool is now built based on MCPServerConfig.Tool (name, description, argument, argument description), not hardcoded.