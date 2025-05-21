# Speelka Agent

Universal LLM agent based on Model Context Protocol (MCP) with support for external tools, flexible configuration, and extensible logging.

## Key Features
- **Multi-agent orchestration**: Supports tools from other MCP servers.
- **Flexible configuration**: YAML, JSON, environment variables, overlay, and property-based overlay.
- **Extensible logging**: Centralized LogConfig, output to stdout, stderr, file, MCP protocol, custom/json/text formats.
- **Security**: Key isolation, log protection, tool access control.
- **Testing**: Golden serialization tests, property-based overlay, unit/integration/E2E.
- **Scalability**: HTTP and stdio support, dynamic tool/session management.

## Architecture
- All components are interface-driven, tested, and follow single-responsibility.
- See [documents/architecture.md](documents/architecture.md) for details.

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

## Logging
- Managed via LogConfig: level, format, output (stdout, stderr, file, MCP).
- MCP logs available via protocol or fallback to stderr (for stdio servers).
- Formats: custom, json, text, unknown.
- See [documents/architecture.md](documents/architecture.md) and [documents/implementation.md](documents/implementation.md) for details.

## Testing
- Unit, integration, E2E.
- Property-based overlay tests (edge-cases, map merge, zero-value preservation).
- Test examples: [documents/implementation.md](documents/implementation.md).

## Project Structure
- See [documents/file_structure.md](documents/file_structure.md) for details.
- Key directories: internal/agent, internal/logger, internal/mcp_connector, internal/types.

## Quick Start
1. Clone the repository and build the agent:
   ```bash
   git clone https://github.com/korchasa/speelka-agent-go.git
   cd speelka-agent-go
   go build ./cmd/server
   ```
2. Prepare a config (see example above) or use environment variables (SPL_...).
3. Run the agent:
   - HTTP mode: `./speelka-agent --daemon [--config config.yaml]`
   - CLI/stdio: `./speelka-agent [--config config.yaml]`

## Documentation
- Architecture: [documents/architecture.md](documents/architecture.md)
- Implementation & tests: [documents/implementation.md](documents/implementation.md)
- File structure: [documents/file_structure.md](documents/file_structure.md)
- External resources: [documents/remote_resources.md](documents/remote_resources.md)

---

For overlay, MCP logs, tests, and structure details, see the documentation in the documents/ folder.

```mermaid
flowchart TB
    User["Any MCP Client"] --> |"1.Request"| Agent["Speelka Agent"]
    Agent --> |"2.Format prompt"| LLM["LLM Service"]
    LLM --> |"3.Tool calls"| Agent
    Agent --> |"4.Execute tools"| Tools["External MCP Tools"]
    Tools --> |"5.Return results"| Agent
    Agent --> |"6.Process repeat"| LLM
    Agent --> |"7.Final answer"| User
```

## Use Cases
- Improve accuracy by splitting large, complex instructions into specialized, focused tasks.
- Reduce cost by using different models for different task parts.
- Extend, narrow, or modify third-party MCP server responses.
- Switch between "real" and LLM-based tool implementations easily.
- Restrict capabilities by limiting available tools in an MCP server.
- Orchestrate multi-step workflows across multiple MCP tools in a single session.
- Enforce per-request token and cost budgets for predictable usage.
- Automatic retry and exponential backoff for transient LLM or MCP server errors.
- Seamless provider switching between LLM services (OpenAI, Anthropic) via unified config.

## Key Features
- **Precise Agent Definition**: Define agent behavior via prompt engineering
- **Client-Side Context Optimization**: Reduce context size for efficient token usage
- **LLM Flexibility**: Use different LLM providers on client and agent sides
- **Centralized Tool Management**: Single control point for all tools
- **Multiple Integration Options**: MCP stdio, MCP HTTP, Simple HTTP API
- **Built-in Reliability**: Retry mechanisms for transient failures
- **Extensibility**: Extend system behavior without client changes
- **MCP-Aware Logging**: Structured logging with MCP notifications
- **Token Management**: Automatic token counting
- **Flexible Configuration**: Environment variables, YAML, JSON
- **LLMService.SendRequest** returns an `LLMResponse` struct with:
  - Response text
  - List of tool calls
  - CompletionTokens, PromptTokens, ReasoningTokens, TotalTokens (token usage)
- **Interface**: `SendRequest(ctx, messages, tools) (LLMResponse, error)`

## Getting Started

### Prerequisites
- Go 1.19 or higher
- LLM API credentials (OpenAI or Anthropic)
- External MCP tools (optional)

### Installation
```bash
git clone https://github.com/korchasa/speelka-agent-go.git
cd speelka-agent-go
go build ./cmd/server
```

### Configuration
Configuration can be provided using YAML, JSON, or environment variables.

> **Note:** The `./examples` directory is deprecated. Use examples in `./site/examples` instead.

Example configuration files are in `site/examples`:
- `site/examples/minimal.yaml`: Basic agent config (YAML)
- `site/examples/ai-news.yaml`: AI news agent config (YAML)
- `site/examples/architect.yaml`: Architect agent config (YAML)

Simple YAML config example:

```yaml
agent:
  name: "simple-speelka-agent"
  version: "1.0.0"
  tool:
    name: "process"
    description: "Process tool for handling user queries with LLM"
    argument_name: "input"
    argument_description: "The user query to process"
  llm:
    provider: "openai"
    apiKey: ""  # Set via environment variable for security
    model: "gpt-4o"
    temperature: 0.7
    promptTemplate: "You are a helpful AI assistant. Respond to the following request: {{input}}. Provide a detailed and helpful response. Available tools: {{tools}}"
  chat:
    max_tokens: 0
    max_llm_iterations: 25
    request_budget: 0.0
  connections:
    mcpServers:
      time:
        command: "docker"
        args: ["run", "-i", "--rm", "mcp/time"]
        includeTools:
          - now
          - utc
      filesystem:
        command: "mcp-filesystem-server"
        args: ["/path/to/directory"]
        excludeTools:
          - delete
runtime:
  log:
    level: "info"
  transports:
    stdio:
      enabled: true
```

#### Using Environment Variables

All environment variables are prefixed with `SPL_`:

| Environment Variable                | Default Value | Description                                                                                                        |
|-------------------------------------|---------------|--------------------------------------------------------------------------------------------------------------------|
| **Agent Configuration**             |               |                                                                                                                    |
| `SPL_AGENT_NAME`                    | *Required*    | Name of the agent                                                                                                  |
| `SPL_AGENT_VERSION`                 | "1.0.0"       | Version of the agent                                                                                               |
| **Tool Configuration**              |               |                                                                                                                    |
| `SPL_AGENT_TOOL_NAME`                     | *Required*    | Name of the tool provided by the agent                                                                             |
| `SPL_AGENT_TOOL_DESCRIPTION`              | *Required*    | Description of the tool functionality                                                                              |
| `SPL_AGENT_TOOL_ARGUMENT_NAME`            | *Required*    | Name of the argument for the tool                                                                                  |
| `SPL_AGENT_TOOL_ARGUMENT_DESCRIPTION`     | *Required*    | Description of the argument for the tool                                                                           |
| **LLM Configuration**               |               |                                                                                                                    |
| `SPL_AGENT_LLM_PROVIDER`                  | *Required*    | Provider of LLM service (e.g., "openai", "anthropic")                                                              |
| `SPL_AGENT_LLM_APIKEY`                   | *Required*    | API key for the LLM provider                                                                                       |
| `SPL_AGENT_LLM_MODEL`                     | *Required*    | Model name (e.g., "gpt-4o", "claude-3-opus-20240229")                                                              |
| `SPL_AGENT_LLM_MAX_TOKENS`                | 0             | Maximum tokens to generate (0 means no limit)                                                                      |
| `SPL_AGENT_LLM_TEMPERATURE`               | 0.7           | Temperature parameter for randomness in generation                                                                 |
| `SPL_AGENT_LLM_PROMPTTEMPLATE`           | *Required*    | Template for system prompts (must include placeholder matching the `SPL_AGENT_TOOL_ARGUMENTNAME` value and `{{tools}}`) |
| **Chat Configuration**              |               |                                                                                                                    |
| `SPL_AGENT_CHAT_MAX_LLM_ITERATIONS`           | 100           | Maximum number of LLM iterations                                                                                   |
| `SPL_AGENT_CHAT_MAX_TOKENS`               | 0             | Maximum tokens in chat history (0 means based on model)                                                            |
| `SPL_AGENT_CHAT_REQUEST_BUDGET`           | 1.0           | Maximum cost (USD or token-equivalent) per request (0 = unlimited)                                                 |
| **LLM Retry Configuration**         |               |                                                                                                                    |
| `SPL_AGENT_LLM_RETRY_MAX_RETRIES`         | 3             | Maximum number of retry attempts for LLM API calls                                                                 |
| `SPL_AGENT_LLM_RETRY_INITIAL_BACKOFF`     | 1.0           | Initial backoff time in seconds                                                                                    |
| `SPL_AGENT_LLM_RETRY_MAX_BACKOFF`         | 30.0          | Maximum backoff time in seconds                                                                                    |
| `SPL_AGENT_LLM_RETRY_BACKOFF_MULTIPLIER`  | 2.0           | Multiplier for increasing backoff time                                                                             |
| **MCP Servers Configuration**       |               |                                                                                                                    |
| `SPL_AGENT_CONNECTIONS_MCPSERVERS_0_ID`                     | ""            | Identifier for the first MCP server                                                                                |
| `SPL_AGENT_CONNECTIONS_MCPSERVERS_0_COMMAND`                | ""            | Command to execute for the first server                                                                            |
| `SPL_AGENT_CONNECTIONS_MCPSERVERS_0_ARGS`                   | ""            | Command arguments as space-separated string                                                                        |
| `SPL_AGENT_CONNECTIONS_MCPSERVERS_0_ENV_*`                  | ""            | Environment variables for the server (prefix with `SPL_AGENT_CONNECTIONS_MCPSERVERS_0_ENV_`)                                               |
| `SPL_AGENT_CONNECTIONS_MCPSERVERS_1_ID`, etc.               | ""            | Configuration for additional servers (increment index)                                                             |
| **MCP Retry Configuration**         |               |                                                                                                                    |
| `SPL_AGENT_CONNECTIONS_RETRY_MAX_RETRIES`        | 3             | Maximum number of retry attempts for MCP server connections                                                        |
| `SPL_AGENT_CONNECTIONS_RETRY_INITIAL_BACKOFF`    | 1.0           | Initial backoff time in seconds                                                                                    |
| `SPL_AGENT_CONNECTIONS_RETRY_MAX_BACKOFF`        | 30.0          | Maximum backoff time in seconds                                                                                    |
| `SPL_AGENT_CONNECTIONS_RETRY_BACKOFF_MULTIPLIER` | 2.0           | Multiplier for increasing backoff time                                                                             |
| **Runtime Configuration**           |               |                                                                                                                    |
| `SPL_RUNTIME_LOG_DEFAULTLEVEL`              | "info"        | Log defaultLevel (debug, info, warn, error)                                                                            |
| `SPL_RUNTIME_LOG_OUTPUT`                    | ":stderr:"      | Log output destination (:stdout:, :stderr:, :mcp:, file path)                                                                 |
| `SPL_RUNTIME_STDIO_ENABLED`         | true          | Enable stdin/stdout transport                                                                                      |
| `SPL_RUNTIME_STDIO_BUFFER_SIZE`     | 8192          | Buffer size for stdio transport                                                                                    |
| `SPL_RUNTIME_HTTP_ENABLED`          | false         | Enable HTTP transport                                                                                              |
| `SPL_RUNTIME_HTTP_HOST`             | "localhost"   | Host for HTTP server                                                                                               |
| `SPL_RUNTIME_HTTP_PORT`             | 3000          | Port for HTTP server                                                                                               |

For more details, see [Environment Variables Reference](documents/knowledge.md#environment-variables-reference).

### Running the Agent

#### Daemon Mode (HTTP Server)

```bash
./speelka-agent --daemon [--config config.yaml]
```

#### CLI Mode (Standard Input/Output)

```bash
./speelka-agent [--config config.yaml]
```

## Usage Examples

### HTTP API

When running in daemon mode, the agent exposes HTTP endpoints:

```bash
# Send a request to the agent
curl -X POST http://localhost:3000/message -H "Content-Type: application/json" -d '{
  "method": "tools/call",
  "params": {
    "name": "process",
    "arguments": {
      "input": "Your query here"
    }
  }
}'
```

### External Tool Integration

Connect to external tools using the MCP protocol in your YAML configuration:

```yaml
agent:
  # ... other agent configuration ...
  connections:
    mcpServers:
      # MCP server for Playwright browser automation
      playwright:
        command: "mcp-playwright"
        args: []

      # MCP server for filesystem operations
      filesystem:
        command: "mcp-filesystem-server"
        args: ["."]
```

Or using environment variables:

```bash
# MCP server for Playwright browser automation
export SPL_AGENT_CONNECTIONS_MCPSERVERS_0_ID="playwright"
export SPL_AGENT_CONNECTIONS_MCPSERVERS_0_COMMAND="mcp-playwright"
export SPL_AGENT_CONNECTIONS_MCPSERVERS_0_ARGS=""

# MCP server for filesystem operations
export SPL_AGENT_CONNECTIONS_MCPSERVERS_1_ID="filesystem"
export SPL_AGENT_CONNECTIONS_MCPSERVERS_1_COMMAND="mcp-filesystem-server"
export SPL_AGENT_CONNECTIONS_MCPSERVERS_1_ARGS="."
```

## Supported LLM Providers

- **OpenAI**: GPT-3.5, GPT-4, GPT-4o
- **Anthropic**: Claude models

## Documentation

For more details, see:
- [System Architecture](documents/architecture.md)
- [Implementation Details](documents/implementation.md)
- [Project File Structure](documents/file_structure.md)
- [Reference Materials](documents/knowledge.md)
- [External Resources](documents/remote_resources.md)

## Development

### Running Tests

```bash
go test ./...
```

### Helper Commands

The `run` script provides commands for common operations:

```bash
# Development
./run build        # Build the project
./run test         # Run tests with coverage
./run check        # Run all checks
./run lint         # Run linter

# Interaction
./run call         # Test with simple query
./run call-multistep # Test with multi-step query
./run call-news    # Test news agent
./run fetch_url    # Fetch a URL using MCP

# Inspection
./run inspect      # Run with MCP inspector
```

See [Command Reference](documents/knowledge.md#command-reference) for more options.

## License

[MIT License](LICENSE)

### MCP Server Tool Filtering

You can control which tools are exported from each MCP server using the following options in the `mcpServers` section:

- `includeTools`: (optional) List of tool names to include. Only these tools will be available from the server.
- `excludeTools`: (optional) List of tool names to exclude. These tools will not be available from the server.
- If both are set, `includeTools` is applied first, then `excludeTools`.
- Tool names are case-sensitive.

Example:
```yaml
connections:
  mcpServers:
    time:
      command: "docker"
      args: ["run", "-i", "--rm", "mcp/time"]
      includeTools:
        - now
        - utc
    filesystem:
      command: "mcp-filesystem-server"
      args: ["/path/to/directory"]
      excludeTools:
        - delete
```

## Direct Call Mode

You can run the agent in direct call mode to process a single query and output a JSON result. This is useful for scripting, automation, or integration with other tools.

Example:

```sh
./bin/speelka-agent --config site/examples/minimal.yaml --call 'What is 2+2?'
```

- The agent will process the query and print a single JSON result to stdout.
- All logs and debug output are sent to stderr.
- The output JSON will always include the fields: `success`, `result`, `meta`, and `error`.

**Tip:**
- You can use this mode in scripts and pipe the output to `jq` or other tools for further processing.