# Speelka Agent

Speelka Agent is a universal LLM agent based on the Model Context Protocol (MCP), providing tool execution capabilities through a Go-based implementation.

```mermaid
flowchart TB
    User["Any MCP Client"] --> |"1. Request"| Agent["Speelka Agent"]
    Agent --> |"2. Format prompt"| LLM["LLM Service"]
    LLM --> |"3. Tool calls"| Agent
    Agent --> |"4. Execute tools"| Tools["External MCP Tools"]
    Tools --> |"5. Return results"| Agent
    Agent --> |"6. Process repeat"| LLM
    Agent --> |"7. Final answer"| User
```

## Key Advantages

- **Precise Agent Definition**: Enables detailed agent behavior definition through prompt engineering
- **Client-Side Context Optimization**: Reduces context size requirements on the client side, resulting in more efficient token usage and cost savings
- **LLM Flexibility**: Allows different LLM providers and configurations between client and agent sides, optimizing for performance and cost
- **Centralized Tool Management**: Provides a single point of control for all available tools
- **Integration Options**: Supports multiple integration methods including MCP stdio, MCP HTTP* and Simple HTTP API* (*planned)
- **Reliability**: Includes built-in retry mechanisms for handling transient failures
- **Extensibility**: Supports system behavior extensions without requiring client-side changes
- **MCP-Aware Logging**: Integrates structured logging with MCP notifications for improved client monitoring

## Architecture

Speelka Agent uses a clean architecture approach with the following key components:

- **Agent**: Central orchestrator that coordinates all other components
- **Configuration Manager**: Provides centralized access to all configuration settings
- **LLM Service**: Handles communication with Language Model providers
- **MCP Server**: Exposes the agent functionality to clients
- **MCP Connector**: Connects to external MCP servers to execute tools
- **Chat**: Manages the conversation history and formatting
- **MCPLogger**: Provides a wrapper around logrus that implements MCP logging protocol while maintaining familiar logging interface

## Getting Started

### Prerequisites

- Go 1.19 or higher
- LLM API credentials (OpenAI or Anthropic)
- External MCP tools (optional)

### Installation

```bash
git clone https://github.com/korchasa/speelka-agent.git
cd speelka-agent
go build ./cmd/speelka-agent
```

### Configuration

Configuration is provided through environment variables. All environment variables are prefixed with `SPL_`:

| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| **Agent Configuration** | | |
| `SPL_AGENT_NAME` | *Required* | Name of the agent |
| `SPL_AGENT_VERSION` | "1.0.0" | Version of the agent |
| **Tool Configuration** | | |
| `SPL_TOOL_NAME` | *Required* | Name of the tool provided by the agent |
| `SPL_TOOL_DESCRIPTION` | *Required* | Description of the tool functionality |
| `SPL_TOOL_ARGUMENT_NAME` | *Required* | Name of the argument for the tool |
| `SPL_TOOL_ARGUMENT_DESCRIPTION` | *Required* | Description of the argument for the tool |
| **LLM Configuration** | | |
| `SPL_LLM_PROVIDER` | *Required* | Provider of LLM service (e.g., "openai", "anthropic") |
| `SPL_LLM_API_KEY` | *Required* | API key for the LLM provider |
| `SPL_LLM_MODEL` | *Required* | Model name (e.g., "gpt-4o", "claude-3-opus-20240229") |
| `SPL_LLM_MAX_TOKENS` | 0 | Maximum tokens to generate (0 means no limit) |
| `SPL_LLM_TEMPERATURE` | 0.7 | Temperature parameter for randomness in generation |
| `SPL_LLM_PROMPT_TEMPLATE` | *Required* | Template for system prompts (must include placeholder matching the `SPL_TOOL_ARGUMENT_NAME` value and `{{tools}}`) |
| **LLM Retry Configuration** | | |
| `SPL_LLM_RETRY_MAX_RETRIES` | 3 | Maximum number of retry attempts for LLM API calls |
| `SPL_LLM_RETRY_INITIAL_BACKOFF` | 1.0 | Initial backoff time in seconds |
| `SPL_LLM_RETRY_MAX_BACKOFF` | 30.0 | Maximum backoff time in seconds |
| `SPL_LLM_RETRY_BACKOFF_MULTIPLIER` | 2.0 | Multiplier for increasing backoff time |
| **MCP Servers Configuration** | | |
| `SPL_MCPS_0_ID` | "" | Identifier for the first MCP server |
| `SPL_MCPS_0_COMMAND` | "" | Command to execute for the first server |
| `SPL_MCPS_0_ARGS` | "" | Command arguments as space-separated string |
| `SPL_MCPS_0_ENV_*` | "" | Environment variables for the server (prefix with `SPL_MCPS_0_ENV_`) |
| `SPL_MCPS_1_ID`, etc. | "" | Configuration for additional servers (increment index) |
| **MCP Retry Configuration** | | |
| `SPL_MSPS_RETRY_MAX_RETRIES` | 3 | Maximum number of retry attempts for MCP server connections |
| `SPL_MSPS_RETRY_INITIAL_BACKOFF` | 1.0 | Initial backoff time in seconds |
| `SPL_MSPS_RETRY_MAX_BACKOFF` | 30.0 | Maximum backoff time in seconds |
| `SPL_MSPS_RETRY_BACKOFF_MULTIPLIER` | 2.0 | Multiplier for increasing backoff time |
| **Runtime Configuration** | | |
| `SPL_RUNTIME_LOG_LEVEL` | "info" | Log level (debug, info, warn, error) |
| `SPL_RUNTIME_LOG_OUTPUT` | "stderr" | Log output destination (stdout, stderr, file path) |
| `SPL_RUNTIME_STDIO_ENABLED` | true | Enable stdin/stdout transport |
| `SPL_RUNTIME_STDIO_BUFFER_SIZE` | 8192 | Buffer size for stdio transport |
| `SPL_RUNTIME_HTTP_ENABLED` | false | Enable HTTP transport |
| `SPL_RUNTIME_HTTP_HOST` | "localhost" | Host for HTTP server |
| `SPL_RUNTIME_HTTP_PORT` | 3000 | Port for HTTP server |

> **Note**: For backward compatibility, the system also accepts environment variables without the `SPL_` prefix, but this behavior may be removed in future versions.

Example configuration files are available in the `examples` directory:
- `examples/simple.env`: Basic agent configuration
- `examples/architect.env`: Software architecture analysis agent

### Running the Agent

#### Daemon Mode (HTTP Server)

```bash
./speelka-agent --daemon
```

#### CLI Mode (Standard Input/Output)

```bash
./speelka-agent
```

## Usage

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

### Integration with External Tools

The agent can connect to external tools using the MCP protocol by configuring environment variables:

```bash
# MCP server for Playwright
export SPL_MCPS_0_ID="playwright"
export SPL_MCPS_0_COMMAND="mcp-playwright"
export SPL_MCPS_0_ARGS=""
export SPL_MCPS_0_ENV_NODE_ENV="production"
```

## Supported LLM Providers

- **OpenAI**: GPT-3.5, GPT-4, GPT-4o
- **Anthropic**: Claude models

## Development

### Project Structure

- `/cmd`: Command-line application entry points
- `/internal`: Core application code
- `/docs`: Project documentation
- `/examples`: Example configuration files
- `/scripts`: Utility scripts for development and configuration conversion

### Running Tests

```bash
go test ./...
```

### Useful links

- [JSON Escape Online Tool](https://www.devtoolsdaily.com/json/escape/)

## License

[MIT License](LICENSE)