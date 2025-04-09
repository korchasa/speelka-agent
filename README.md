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

Configuration is provided through a single `CONFIG_JSON` environment variable containing a complete JSON structure:

```bash
CONFIG_JSON='{"agent":{"name":"speelka-agent","version":"1.0.0","tool":{"name":"process","description":"Process tool for handling user queries with LLM","argument_name":"input","argument_description":"User query to process"},"llm":{"provider":"openai","api_key":"your_api_key_here","model":"gpt-4o","max_tokens":0,"temperature":0.7,"prompt_template":"You are a helpful AI assistant...","retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}},"connections":{"servers":[{"id":"server-id-1","transport":"stdio","command":"docker","arguments":["run","-i","--rm","mcp/time"],"environment":{"NODE_ENV":"production"}}],"retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}}},"runtime":{"log":{"level":"info","output":"stdout"},"transports":{"stdio":{"enabled":true,"buffer_size":8192,"auto_detect":false},"http":{"enabled":true,"host":"localhost","port":3000}}}}'
```

You can also override specific configuration values using environment variables:

```bash
# Override the LLM API key
export LLM_API_KEY=your_actual_api_key
```

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

The agent can connect to external tools using the MCP protocol:

```json
"connections": {
  "servers": [
    {
      "id": "playwright",
      "transport": "stdio",
      "command": "mcp-playwright",
      "environment": {
        "NODE_ENV": "production"
      }
    }
  ]
}
```

## Supported LLM Providers

- **OpenAI**: GPT-3.5, GPT-4, GPT-4o
- **Anthropic**: Claude models

## Development

### Project Structure

- `/cmd`: Command-line application entry points
- `/internal`: Core application code
- `/docs`: Project documentation

### Running Tests

```bash
go test ./...
```

## License

[MIT License](LICENSE)