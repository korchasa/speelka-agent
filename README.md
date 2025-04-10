# Speelka Agent

Speelka Agent is a universal LLM agent based on the Model Context Protocol (MCP), providing tool execution capabilities through a Go-based implementation.

## Key Advantages

- **Precise Agent Definition**: Enables detailed agent behavior definition through prompt engineering
- **Client-Side Context Optimization**: Reduces context size requirements on the client side, resulting in more efficient token usage and cost savings
- **LLM Flexibility**: Allows different LLM providers and configurations between client and agent sides, optimizing for performance and cost
- **Centralized Tool Management**: Provides a single point of control for all available tools
- **Integration Options**: Supports multiple integration methods including MCP stdio, MCP HTTP* and Simple HTTP API* (*planned)
- **Reliability**: Includes built-in retry mechanisms for handling transient failures
- **Extensibility**: Supports system behavior extensions without requiring client-side changes

## Architecture

Speelka Agent uses a clean architecture approach with the following key components:

- **Agent**: Central orchestrator that coordinates all other components
- **Configuration Manager**: Provides centralized access to all configuration settings
- **LLM Service**: Handles communication with Language Model providers
- **MCP Server**: Exposes the agent functionality to clients
- **MCP Connector**: Connects to external MCP servers to execute tools
- **Chat**: Manages the conversation history and formatting

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

Configuration is provided through environment variables:

| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| **Agent Configuration** | | |
| `AGENT_NAME` | *Required* | Name of the agent |
| `AGENT_VERSION` | "1.0.0" | Version of the agent |
| **Tool Configuration** | | |
| `TOOL_NAME` | *Required* | Name of the tool provided by the agent |
| `TOOL_DESCRIPTION` | *Required* | Description of the tool functionality |
| `TOOL_ARGUMENT_NAME` | *Required* | Name of the argument for the tool |
| `TOOL_ARGUMENT_DESCRIPTION` | *Required* | Description of the argument for the tool |
| **LLM Configuration** | | |
| `LLM_PROVIDER` | *Required* | Provider of LLM service (e.g., "openai", "anthropic") |
| `LLM_API_KEY` | *Required* | API key for the LLM provider |
| `LLM_MODEL` | *Required* | Model name (e.g., "gpt-4o", "claude-3-opus-20240229") |
| `LLM_MAX_TOKENS` | 0 | Maximum tokens to generate (0 means no limit) |
| `LLM_TEMPERATURE` | 0.7 | Temperature parameter for randomness in generation |
| `LLM_PROMPT_TEMPLATE` | *Required* | Template for system prompts (must include placeholder matching the `TOOL_ARGUMENT_NAME` value and `{{tools}}`) |
| **LLM Retry Configuration** | | |
| `LLM_RETRY_MAX_RETRIES` | 3 | Maximum number of retry attempts for LLM API calls |
| `LLM_RETRY_INITIAL_BACKOFF` | 1.0 | Initial backoff time in seconds |
| `LLM_RETRY_MAX_BACKOFF` | 30.0 | Maximum backoff time in seconds |
| `LLM_RETRY_BACKOFF_MULTIPLIER` | 2.0 | Multiplier for increasing backoff time |
| **MCP Servers Configuration** | | |
| `MCPS_0_ID` | "" | Identifier for the first MCP server |
| `MCPS_0_COMMAND` | "" | Command to execute for the first server |
| `MCPS_0_ARGS` | "" | Command arguments as space-separated string |
| `MCPS_0_ENV_*` | "" | Environment variables for the server (prefix with `MCPS_0_ENV_`) |
| `MCPS_1_ID`, etc. | "" | Configuration for additional servers (increment index) |
| **MCP Retry Configuration** | | |
| `MSPS_RETRY_MAX_RETRIES` | 3 | Maximum number of retry attempts for MCP server connections |
| `MSPS_RETRY_INITIAL_BACKOFF` | 1.0 | Initial backoff time in seconds |
| `MSPS_RETRY_MAX_BACKOFF` | 30.0 | Maximum backoff time in seconds |
| `MSPS_RETRY_BACKOFF_MULTIPLIER` | 2.0 | Multiplier for increasing backoff time |
| **Runtime Configuration** | | |
| `RUNTIME_LOG_LEVEL` | "info" | Log level (debug, info, warn, error) |
| `RUNTIME_LOG_OUTPUT` | "stdout" | Log output destination (stdout, file path) |
| `RUNTIME_STDIO_ENABLED` | true | Enable stdin/stdout transport |
| `RUNTIME_STDIO_BUFFER_SIZE` | 8192 | Buffer size for stdio transport |
| `RUNTIME_HTTP_ENABLED` | false | Enable HTTP transport |
| `RUNTIME_HTTP_HOST` | "localhost" | Host for HTTP server |
| `RUNTIME_HTTP_PORT` | 3000 | Port for HTTP server |

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
export MCPS_0_ID="playwright"
export MCPS_0_COMMAND="mcp-playwright"
export MCPS_0_ARGS=""
export MCPS_0_ENV_NODE_ENV="production"
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

## License

[MIT License](LICENSE)