# Speelka Agent Go

## Overview

Speelka Agent is an MCP (Model-Controller-Presenter) server that implements a "process" tool. It acts as a bridge between user requests and Large Language Models (LLMs), with the ability to utilize tools from other MCP servers.

The agent facilitates communication with LLM models using configurable parameters and enables a chain of tool calls until a final answer is produced.

## Features

- **MCP Server Integration**: Connect to all MCP servers specified in the configuration and extract available tools
- **LLM Integration**: Form requests to LLM models using configurable prompt templates and model parameters
- **Tool Orchestration**: Redirect tool calls from the LLM to appropriate MCP servers and include results in subsequent LLM requests
- **Answer Generation**: Return the final answer when the LLM completes its processing
- **Command-Line Interface**: Run as a command-line tool (script mode)
- **HTTP Interface**: Support for HTTP transport with Server-Sent Events (SSE)
- **Error Handling**: Comprehensive error handling with retry strategies for transient failures

## Architecture

The application follows a modular architecture and is organized around the following core components:

- **Agent**: Central component that coordinates all interactions and workflow
- **Configuration Manager**: Handles application settings loaded from environment variables
- **LLM Service**: Manages communication with language model providers (OpenAI, Anthropic)
- **MCP Server**: Implements the MCP protocol server with stdio transport options
- **MCP Connector**: Handles connections to external MCP servers and tool execution
- **Chat**: Manages conversation history and prompt formatting

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Access to an LLM provider (OpenAI, Anthropic, etc.)
- Access to other MCP servers (optional, but recommended for full functionality)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/korchasa/speelka-agent-go.git
   cd speelka-agent-go
   ```

2. Build the application:
   ```bash
   go build -o speelka-agent
   ```

3. Create a configuration and set it in the environment variables (see example below)

4. Run the application in script mode:
   ```bash
   echo '{"input": "What is the capital of France?"}' | ./speelka-agent
   ```

### Configuration

The Speelka Agent is configured using a single `CONFIG_JSON` environment variable in JSON format.

Example:

```bash
# Set the CONFIG_JSON environment variable
export CONFIG_JSON='{
  "agent": {
    "name": "speelka-agent",
    "version": "1.0.0",
    "tool": {
      "name": "process",
      "description": "Process tool for handling user queries with LLM",
      "argument_name": "input",
      "argument_description": "User query to process"
    },
    "llm": {
      "provider": "openai",
      "api_key": "your_api_key_here",
      "model": "gpt-4o",
      "temperature": 0.7,
      "prompt_template": "You are a helpful AI assistant...",
      "retry": {
        "max_retries": 3,
        "initial_backoff": 1.0,
        "max_backoff": 30.0,
        "backoff_multiplier": 2.0
      }
    }
  },
  "runtime": {
    "transports": {
      "stdio": {
        "enabled": true,
        "buffer_size": 8192,
        "auto_detect": false
      }
    }
  }
}'

# Run the application
./speelka-agent
```

For convenience, you can store your configuration in a JSON file and load it:

```bash
# Using the provided example script
./examples/run-with-json-config.sh

# Or manually
export CONFIG_JSON=$(cat examples/config.json)
./speelka-agent
```

For your convenience, the project includes an `examples/config.json` file that you can copy and customize.

### Environment Variables

The Speelka Agent also supports configuration through environment variables:

- `LLM_API_KEY` - Sets or overrides the API key for the LLM provider. This takes precedence over the value in CONFIG_JSON.

### Configuration Categories

Main configuration categories:

1. **Agent Settings**:
   - `agent.name` - Name of the agent
   - `agent.version` - Version of the agent
   - `agent.tool.name` - Name of the tool provided by the agent
   - `agent.tool.description` - Description of the tool provided by the agent
   - `agent.tool.argument_name` - Name of the tool argument
   - `agent.tool.argument_description` - Description of the tool argument

2. **LLM Settings**:
   - `agent.llm.provider` - LLM provider (openai, anthropic)
   - `agent.llm.api_key` - API key for LLM provider (can also be set via LLM_API_KEY environment variable)
   - `agent.llm.model` - LLM model to use
   - `agent.llm.temperature` - Temperature parameter
   - `agent.llm.prompt_template` - Template for system messages, using Jinja2 format
   - `agent.llm.retry.*` - Retry configuration for LLM requests

3. **Connection Settings**:
   - `agent.connections.servers` - Array of MCP server configurations
   - `agent.connections.servers[].id` - ID of the MCP server
   - `agent.connections.servers[].transport` - Transport type (stdio)
   - `agent.connections.servers[].command` - Command for stdio transport
   - `agent.connections.servers[].arguments` - Arguments for stdio transport command
   - `agent.connections.servers[].environment` - Environment variables for stdio transport
   - `agent.connections.retry.*` - Retry configuration for connection attempts

4. **Runtime Settings**:
   - `runtime.log.level` - Logging level (debug, info, warn, error, fatal, panic)
   - `runtime.log.output` - Log output destination (stdout, stderr, or file path)
   - `runtime.transports.stdio.enabled` - Enable stdio server
   - `runtime.transports.stdio.buffer_size` - Buffer size for stdio transport
   - `runtime.transports.stdio.auto_detect` - Auto-detect stdio mode
   - `runtime.transports.http.enabled` - Enable HTTP server
   - `runtime.transports.http.host` - HTTP server host
   - `runtime.transports.http.port` - HTTP server port

## CONFIG_JSON Structure

Below is a comprehensive overview of the CONFIG_JSON structure:

```json
{
  "agent": {
    "name": "string",          // Name of the agent
    "version": "string",       // Version of the agent
    "tool": {
      "name": "string",        // Name of the tool provided by the agent
      "description": "string", // Description of the tool provided by the agent
      "argument_name": "string", // Name of the tool argument
      "argument_description": "string" // Description of the tool argument
    },
    "llm": {
      "provider": "string",     // LLM provider ("openai", "anthropic")
      "api_key": "string",      // API key for the LLM provider
      "model": "string",        // LLM model name
      "max_tokens": number,     // Maximum tokens to generate (default: 0)
      "temperature": number,    // Temperature parameter (default: 0.7)
      "prompt_template": "string", // System prompt template with Jinja2 format
      "retry": {
        "max_retries": number,  // Maximum number of retry attempts (default: 3)
        "initial_backoff": number, // Initial backoff in seconds (default: 1.0)
        "max_backoff": number,  // Maximum backoff in seconds (default: 30.0)
        "backoff_multiplier": number // Backoff multiplier (default: 2.0)
      }
    },
    "connections": {
      "servers": [
        {
          "id": "string",       // Unique ID of the MCP server
          "transport": "string", // Transport type ("stdio")
          "command": "string",  // Command for stdio transport
          "arguments": ["string"], // Arguments for stdio transport command
          "environment": {      // Environment variables for stdio transport
            "key": "value"
          }
        }
      ],
      "retry": {
        "max_retries": number,  // Maximum number of retry attempts (default: 3)
        "initial_backoff": number, // Initial backoff in seconds (default: 1.0)
        "max_backoff": number,  // Maximum backoff in seconds (default: 30.0)
        "backoff_multiplier": number // Backoff multiplier (default: 2.0)
      }
    }
  },
  "runtime": {
    "log": {
      "level": "string",        // Log level (debug, info, warn, error, fatal, panic)
      "output": "string"        // Output destination (stdout, stderr, file path)
    },
    "transports": {
      "stdio": {
        "enabled": boolean,     // Enable stdin/stdout server (default: true)
        "buffer_size": number,  // Buffer size for reading/writing (default: 8192)
        "auto_detect": boolean  // Auto-detect stdio mode (default: false)
      },
      "http": {
        "enabled": boolean,     // Enable HTTP server (default: false)
        "host": "string",       // HTTP server host (default: "localhost")
        "port": number         // HTTP server port (default: 3000)
      }
    }
  }
}
```

## Usage

### Script Mode

In script mode, the Speelka Agent operates as a command-line tool that reads from stdin and writes to stdout:

```bash
echo '{"input": "What is the capital of France?"}' | ./speelka-agent
```

You can also use it in shell pipelines:

```bash
cat query.json | ./speelka-agent | jq '.output'
```

Or as a subprocess from other applications:

```python
import subprocess

result = subprocess.run(
    ["./speelka-agent"],
    input='{"input": "What is the capital of France?"}',
    text=True,
    capture_output=True
)
print(result.stdout)
```

## Development

### Project Structure

```
speelka-agent-go/
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── internal/             # Private application code
│   ├── agent/            # Central agent handling requests
│   ├── chat/             # Chat history management
│   ├── configuration/    # Configuration handling
│   ├── error_handling/   # Error handling utilities
│   ├── llm_service/      # LLM service implementation
│   ├── mcp_connector/    # MCP client connections
│   ├── mcp_server/       # MCP server implementation
│   ├── types/            # Shared types and interfaces
│   └── utils/            # Utility functions
├── test/                 # Test files
│   ├── integration/      # Integration tests
│   └── unit/             # Unit tests
├── docs/                 # Documentation
└── configs/              # Configuration files
```

### Running Tests

To run tests with coverage information:

```bash
./run test
```

This will run all tests and display package-level coverage information in a compact format. An HTML coverage report will be automatically generated in the `.coverage/coverage.html` file, providing a visual representation of code coverage that can be viewed in a web browser.

For more detailed function-level coverage reporting:

```bash
./run test --details
```

### Running Modes

Speelka Agent supports script mode:

**Script Mode** - Runs as a stdio MCP server that processes a single task and exits:
```bash
echo '{"input": "What is the capital of France?"}' | ./speelka-agent
```
In this mode, the agent reads from stdin, processes the request using the LLM, and writes the response to stdout.

### Building Docker Image

```bash
docker build -t speelka-agent:latest .
```

### Docker Usage

#### Available Docker Images

The Speelka Agent Docker images are available from GitHub Container Registry:

```
ghcr.io/korchasa/speelka-agent:latest    # Latest version from main branch
ghcr.io/korchasa/speelka-agent:v1.0.0    # Specific version (example)
```

#### Using Environment Variables for Configuration

You can provide the configuration as an environment variable:

```bash
# Using a config file
docker run \
  -e CONFIG_JSON="$(cat config.json)" \
  -e LLM_API_KEY="your_api_key" \
  ghcr.io/korchasa/speelka-agent:latest
```

#### Using Volumes

You can mount a configuration file from your host:

```bash
docker run \
  -v $(pwd)/examples:/app/examples \
  -e CONFIG_JSON="$(cat examples/config.json)" \
  ghcr.io/korchasa/speelka-agent:latest
```

#### Running in Script Mode (STDIO)

To run the agent in script mode (processing input from stdin):

```bash
echo '{"input": "What is the capital of France?"}' | \
  docker run -i --rm \
  -e CONFIG_JSON="$(cat examples/simple.json)" \
  ghcr.io/korchasa/speelka-agent:latest
```

#### Using Docker Compose

Here's an example `docker-compose.yml` file:

```yaml
version: '3'

services:
  speelka-agent:
    image: ghcr.io/korchasa/speelka-agent:latest
    environment:
      - CONFIG_JSON={"server":{"name":"speelka-agent","version":"1.0.0","tool":{"name":"process","description":"Process tool for handling user queries with LLM","argument_name":"input","argument_description":"User query to process"},"stdio":{"enabled":true,"buffer_size":8192,"auto_detect":false}},"llm":{"provider":"openai","api_key":"your_api_key_here","model":"gpt-4o","temperature":0.7}}
      - LLM_API_KEY=your_api_key_here
    restart: unless-stopped
```

To use this compose file:

```bash
docker-compose up -d
```

#### Security Notes

- Never hardcode API keys in your Docker images or Dockerfiles.
- Use environment variables or mounted secrets for sensitive values.
- Consider using Docker secrets or a vault service for production deployments.

## Roadmap

- [x] Short format for logs on INFO level
- [x] Change config format
- [ ] Configuration page: Добавь в режим демона страницу /config, которая будет отображать html-страну с текушим конфигом, возможностью его менять и подсказкой
- [ ] Testing
  - [ ] Agent
  - [ ] LLM Service
  - [ ] MCP Connector
  - [ ] MCP Server
  - [ ] Chat
- [ ] MCP logging support
- [ ] Refactoring: Application, error handling, interfaces
- [ ] Return MCP errors to LLM
- [ ] MCP Notifications about calls
- [ ] Thoughts and goals of tool usage
- [ ] MCP capabilities cache

## License

This project is licensed under the MIT License - see the LICENSE file for details.