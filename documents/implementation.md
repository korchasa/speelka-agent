# Implementation Details

## Core Components

### Agent
- **Purpose**: Central coordinator for request processing
- **Implementation**: `internal/agent/agent.go`
- **Key Features**:
  - Manages request lifecycle
  - Coordinates between components
  - Handles tool execution
  - Manages state and context
- **Recent Improvements**:
  - Added robust null checking for interface value handling in `HandleRequest` method to prevent panics
  - Implemented safer type assertion pattern with descriptive error messages
  - Enhanced error handling to gracefully handle nil values in tool arguments

### Chat
- **Purpose**: Manages conversation history and prompt formatting
- **Implementation**: `internal/chat/chat.go`
- **Key Features**:
  - Maintains message history
  - Formats system prompts
  - Builds tool descriptions
  - Uses Jinja2 templates

### Configuration Manager
- **Purpose**: Manages application configuration
- **Implementation**: `internal/configuration/manager.go`
- **Key Features**:
  - JSON configuration via `CONFIG_JSON` environment variable
  - Specific environment variables (LLM_API_KEY) for targeted overrides
  - Type-safe configuration access
  - Default value handling
  - Configuration validation

### LLM Service
- **Purpose**: Handles LLM provider integration
- **Implementation**: `internal/llm_service/llm_service.go`
- **Key Features**:
  - OpenAI/Anthropic support
  - Request/response handling
  - Retry logic
  - Error categorization

### MCP Server
- **Purpose**: Implements MCP protocol server
- **Implementation**: `internal/mcp_server/mcp_server.go`
- **Key Features**:
  - HTTP/stdio transport
  - Request handling
  - Debug hooks
  - Graceful shutdown
  - SSE (Server-Sent Events) for real-time communication

#### HTTP Server Implementation
- The MCP server uses the SSE (Server-Sent Events) server from `github.com/mark3labs/mcp-go/server`
- When running in daemon mode (`agent.Start(true, ctx)`), the server exposes two main endpoints:
  - `/sse` - For establishing a Server-Sent Events connection to receive real-time updates
  - `/message` - For sending HTTP POST requests to invoke tools

- **Request Format**:
  ```json
  {
    "method": "tools/call",
    "params": {
      "name": "process",
      "arguments": {
        "input": "User query text here"
      }
    }
  }
  ```

- **Response Format**:
  ```json
  {
    "jsonrpc": "2.0",
    "id": 1,
    "result": "Response from the LLM or tool",
    "isError": false
  }
  ```

- The server is started in `MCPServer.ServeDaemon()` which initializes an SSE server and starts it on the configured host and port

### MCP Connector
- **Purpose**: Manages external tool connections
- **Implementation**: `internal/mcp_connector/mcp_connector.go`
- **Key Features**:
  - Server connection management
  - Tool discovery
  - Tool execution
  - Connection retry logic

## Key Interfaces

### Configuration
```go
type ConfigurationManagerSpec interface {
    LoadConfiguration(ctx context.Context) error
    GetMCPServerConfig() MCPServerConfig
    GetMCPConnectorConfig() MCPConnectorConfig
    GetLLMConfig() LLMConfig
    GetLogConfig() LogConfig
}
```

### LLM Service
```go
type LLMConfig struct {
    Provider string
    Model string
    APIKey string
    MaxTokens int
    Temperature float64
    PromptTemplate string
    RetryConfig RetryConfig
}
```

### MCP Server
```go
type MCPServerConfig struct {
    Name string
    Version string
    Tool MCPServerToolConfig
    HTTP HTTPConfig
    Stdio StdioConfig
    Debug bool
}
```

### MCP Connector
```go
type MCPConnectorConfig struct {
    Servers []MCPServerConnection
    RetryConfig RetryConfig
}
```

## Error Handling

### Categories
- **Validation**: Configuration and input validation errors
- **Transient**: Network and rate limit errors
- **Internal**: Component and runtime errors
- **External**: Tool execution errors

### Retry Strategy
```go
type RetryConfig struct {
    MaxRetries int
    InitialBackoff float64
    MaxBackoff float64
    BackoffMultiplier float64
}
```

## Request Processing Flow

1. **Request Reception**
   - HTTP or stdio transport
   - Request validation
   - Context creation

2. **Agent Processing**
   - Chat history initialization
   - Tool discovery
   - LLM interaction

3. **LLM Interaction**
   - Prompt formatting
   - Request sending
   - Response parsing
   - Tool call extraction

4. **Tool Execution**
   - Tool lookup
   - Request forwarding
   - Result capture
   - Error handling

5. **Response Generation**
   - Result formatting
   - Response sending
   - Resource cleanup

## Configuration Example

### Environment Variables Configuration

The Speelka Agent is now configured using environment variables directly instead of a single `CONFIG_JSON` environment variable. This makes the configuration more transparent and easier to manage in container environments.

```bash
# Agent
export AGENT_NAME="architect-speelka-agent"
export AGENT_VERSION="1.0.0"

# Tool
export TOOL_NAME="architect"
export TOOL_DESCRIPTION="Architecture design and assessment tool for software systems"
export TOOL_ARGUMENT_NAME="query"
export TOOL_ARGUMENT_DESCRIPTION="Architecture query or task to process"

# LLM
export LLM_PROVIDER="openai"
export LLM_API_KEY="your_api_key_here"
export LLM_MODEL="gpt-4o"
export LLM_MAX_TOKENS=0
export LLM_TEMPERATURE=0.2
export LLM_PROMPT_TEMPLATE="# ROLE
You are a Senior Software Architect with extensive expertise in design
patterns, system architecture, performance optimization, and security best
practices.

# GOAL
Analyze and enhance the architecture of the system according to the user
query below.

# WORKFLOW
1. First, carefully analyze the current architecture described in the query or existing documentation.
2. Generate a detailed analysis of strengths and weaknesses.
3. Analyse current state of project, located in ./
4. Propose architectural improvements with clear justifications.
5. Use diagrams when helpful to illustrate complex concepts.
6. Provide implementation recommendations with relevant examples.

# User query
{{query}}

# Available tools
NOTE: Try to minimize call count!
{{tools}}"
export LLM_RETRY_MAX_RETRIES=3
export LLM_RETRY_INITIAL_BACKOFF=1.0
export LLM_RETRY_MAX_BACKOFF=30.0
export LLM_RETRY_BACKOFF_MULTIPLIER=2.0

# MCP Servers are defined using indexed variables (MCPS_0, MCPS_1, etc.)
export MCPS_0_ID="time"
export MCPS_0_COMMAND="docker"
export MCPS_0_ARGS="run -i --rm mcp/time"

export MCPS_1_ID="mcp-filesystem-server"
export MCPS_1_COMMAND="mcp-filesystem-server"
export MCPS_1_ARGS="."

# MSPS Retry Configuration
export MSPS_RETRY_MAX_RETRIES=3
export MSPS_RETRY_INITIAL_BACKOFF=1.0
export MSPS_RETRY_MAX_BACKOFF=30.0
export MSPS_RETRY_BACKOFF_MULTIPLIER=2.0

# Runtime Configuration
export RUNTIME_LOG_LEVEL="debug"
export RUNTIME_LOG_OUTPUT="./architect.log"

export RUNTIME_STDIO_ENABLED=true
export RUNTIME_STDIO_BUFFER_SIZE=8192
export RUNTIME_HTTP_ENABLED=false
export RUNTIME_HTTP_HOST="localhost"
export RUNTIME_HTTP_PORT=3000
```

#### Required Environment Variables

The following environment variables are required for proper operation:

* `AGENT_NAME`: The name of the agent
* `TOOL_NAME`: The name of the tool provided by the agent
* `TOOL_DESCRIPTION`: Description of the tool functionality
* `LLM_PROVIDER`: The provider of the LLM service (e.g., "openai")
* `LLM_MODEL`: The specific model to use (e.g., "gpt-4o")
* `LLM_PROMPT_TEMPLATE`: The template to use for system prompts (must include {{query}} and {{tools}} placeholders)

#### MCP Servers Configuration

MCP servers are configured using indexed environment variables in the format:

```
MCPS_<index>_ID="server-id"
MCPS_<index>_COMMAND="command"
MCPS_<index>_ARGS="arg1 arg2 arg3"
```

The `<index>` should start from 0 and increment for each server. The `ID` field is used as the key in the map of MCP servers.

### Legacy JSON Configuration (Deprecated)

The previous JSON-based configuration method using the `CONFIG_JSON` environment variable is now deprecated:

```bash
# Complete configuration in a single environment variable
CONFIG_JSON='{"agent":{"name":"speelka-agent","version":"1.0.0","tool":{"name":"process","description":"Process tool for handling user queries with LLM","argument_name":"input","argument_description":"User query to process"},"llm":{"provider":"openai","api_key":"your_api_key_here","model":"gpt-4o","max_tokens":0,"temperature":0.7,"prompt_template":"You are a helpful AI assistant...","retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}},"connections":{"mcpServers":{"time":{"command":"docker","args":["run","-i","--rm","mcp/time"]}},"retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}}},"runtime":{"log":{"level":"info","output":"stdout"},"transports":{"stdio":{"enabled":true,"buffer_size":8192},"http":{"enabled":true,"host":"localhost","port":3000}}}}'
```

## Core Agent Loop

The main agent loop in `internal/agent/agent.go` follows this pattern:

1. Receive user request via MCP server
2. Initialize Chat with system prompt and available tools
3. Enter the LLM interaction loop (max 20 iterations):
   - Send current conversation to LLM
   - Process LLM response (text + tool calls)
   - If "exit" tool is called, return final response
   - For each tool call:
     - Execute tool via MCP connector
     - Add result to conversation history
   - Continue loop if more tools are called
4. Return error if maximum iterations reached

## LLM Service Implementation

The LLM service (`internal/llm_service/llm_service.go`) supports multiple providers:

| Provider | Models | Implementation |
|----------|--------|----------------|
| OpenAI   | GPT-3.5, GPT-4 | Via langchaingo/llms/openai |
| Anthropic | Claude models | Via langchaingo/llms/anthropic |

Key features:
- Provider-specific client initialization
- Consistent error handling across providers
- Token usage tracking
- Configurable retry with exponential backoff
- Tool-enabled conversations

## MCP Protocol Integration

### Server Implementation

The MCP server (`internal/mcp_server/mcp_server.go`) provides:

- HTTP transport with Server-Sent Events (SSE)
- Stdio transport for CLI applications
- Single tool registration for the main agent functionality
- Debug hooks for request/response logging
- Graceful shutdown handling

### Connector Implementation

The MCP connector (`internal/mcp_connector/mcp_connector.go`) provides:

- Connection to multiple external MCP servers
- Support for both HTTP and stdio transports
- Tool discovery and registration
- Concurrent tool execution
- Connection retries with exponential backoff
- Server selection based on available tools

## Chat History Management

The Chat component (`internal/chat/chat.go`) manages:

- Conversation history using langchaingo message formats
- System prompt template with Jinja2 syntax
- Tool description formatting
- Tool call and result tracking
- AI and user message handling

## Error Handling Strategy

Error handling is implemented using the `internal/error_handling` package:

```go
// Example error creation
return error_handling.NewError(
    "provider is required",
    error_handling.ErrorCategoryValidation,
)

// Example error wrapping
return error_handling.WrapError(
    err,
    "failed to initialize OpenAI client",
    error_handling.ErrorCategoryInternal,
)
```

Error categories:
- `ErrorCategoryValidation`: Input or configuration errors
- `ErrorCategoryTransient`: Temporary errors that can be retried
- `ErrorCategoryInternal`: Internal system errors
- `ErrorCategoryExternal`: Errors from external systems
- `ErrorCategoryUnknown`: Unclassified errors

## Configuration Management

Configuration is loaded from environment variables or files and accessed through the `ConfigurationManagerSpec` interface:

```go
// Example configuration access
llmConfig := configManager.GetLLMConfig()
mcpServerConfig := configManager.GetMCPServerConfig()
mcpConnectorConfig := configManager.GetMCPConnectorConfig()
```

Key configuration structures:
- `LLMConfig`: LLM provider settings
- `MCPServerConfig`: Server transport and tool settings
- `MCPConnectorConfig`: External server connections
- `LogConfig`: Logging settings

## Threading and Concurrency

- Main agent loop runs serially for predictable behavior
- MCP connector uses mutex locks for thread safety
- Context propagation for cancellation and timeouts
- Goroutines used for background processing where appropriate

## Testing Approach

Example tests from the codebase:

```go
// From convert_tool_to_llm_test.go
func TestConvertToolsToLLM(t *testing.T) {
    tools := []mcp.Tool{
        mcp.NewTool("test_tool",
            mcp.WithDescription("Test tool description"),
            mcp.WithString("arg1",
                mcp.Description("Argument 1"),
                mcp.Required(),
            ),
        ),
    }

    llmTools, err := ConvertToolsToLLM(tools)
    assert.NoError(t, err)
    assert.Equal(t, 1, len(llmTools))
    assert.Equal(t, "test_tool", llmTools[0].Name)
}
```

- Unit tests for core functionality
- Mocking of external dependencies
- Table-driven test cases
- Integration tests for component interactions

## Run Script Commands

The project includes a versatile `run` script that provides various commands for development, testing, and integration:

| Command | Description | Usage |
|---------|-------------|-------|
| `test` | Run all tests with coverage | `./run test` |
| `lint` | Run code linting | `./run lint` |
| `build` | Build the project | `./run build` |
| `dev` | Run in development mode | `./run dev` |
| `call` | Test with simple query | `./run call` |
| `complex-call` | Test with complex query | `./run complex-call` |
| `http-call` | Call process over HTTP | `./run http-call [url]` |
| `fetch_url` | Fetch URL via MCP | `./run fetch_url <url>` |
| `check` | Run all project checks | `./run check` |

### HTTP Call Testing

The `http-call` command provides a way to test HTTP connections and make calls using the MCP CLI:

```bash
# Test default connection (localhost:3000)
./run http-call

# Test specific endpoint
./run http-call http://localhost:3001
```

Features:
- Validates HTTP server availability
- Automatically starts server if not detected
- Configures appropriate port based on URL
- Uses MCP CLI to test actual integration
- Provides detailed status feedback
- Cleans up resources after testing

Implementation details:
1. Checks if server is running at specified URL
2. If not running, starts server using architect.json config
3. Makes test call using MCP CLI
4. Reports success/failure status

## Prompt Template Validation

The prompt template validation system ensures that all required placeholders are present in the template. This is crucial for the proper functioning of the agent, as missing placeholders would result in incomplete or incorrect prompts being sent to the LLM.

### Key Features

1. **Required Placeholder Validation**: Checks that all required placeholders (tool argument name and `tools`) are present in the template.
2. **Detailed Error Messages**: When validation fails, the system provides comprehensive error messages that:
   - List all missing placeholders
   - Explain that placeholders should match the configuration
   - Identify common mistakes users make (e.g., hardcoded placeholder names)
   - Provide an example of a correct template

3. **Early Validation**: Template validation occurs during configuration loading, ensuring issues are caught before runtime.

### Implementation

The validation logic is implemented in the `validatePromptTemplate` method of the `Manager` struct in the `configuration` package. It:

1. Extracts all placeholders from the template using a regex pattern
2. Checks if all required placeholders are present
3. Generates detailed error messages for missing placeholders

Error messages are specifically designed to guide users toward proper configuration, explaining the relationship between placeholder names and the tool configuration.

### Example Error Message

When a template is missing required placeholders, an error like this is generated:

```
prompt template is missing required placeholder(s): input

Expected placeholder '{{input}}' should match the 'argument_name' value in your tool configuration.
Common mistake: Using a hardcoded placeholder name like '{{query}}' instead of the configured argument name.

Example of a valid template:
You are a helpful assistant.

User request: {{input}}

Available tools:
{{tools}}
```

This clear error messaging significantly improves the user experience when configuring prompt templates.