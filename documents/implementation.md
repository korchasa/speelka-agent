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

### JSON Configuration

```bash
# Complete configuration in a single environment variable
CONFIG_JSON='{"server":{"name":"speelka-agent","version":"1.0.0","tool":{"name":"process","description":"Process tool for handling user queries with LLM","argument_name":"input","argument_description":"User query to process"},"http":{"enabled":true,"host":"localhost","port":3000},"stdio":{"enabled":true,"buffer_size":8192,"auto_detect":false},"debug":false},"mcp_connector":{"servers":[{"id":"server-id-1","transport":"stdio","command":"docker","arguments":["run","-i","--rm","mcp/time"],"environment":{"NODE_ENV":"production"}}],"retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}},"llm":{"provider":"openai","api_key":"your_api_key_here","model":"gpt-4o","max_tokens":0,"temperature":0.7,"prompt_template":"You are a helpful AI assistant...","retry":{"max_retries":3,"initial_backoff":1.0,"max_backoff":30.0,"backoff_multiplier":2.0}},"log":{"level":"info","format":"text","output":"stdout"}}'
```

### Environment Variables

In addition to the `CONFIG_JSON` environment variable, specific configuration values can be overridden using dedicated environment variables:

```bash
# Override the LLM API key
export LLM_API_KEY=your_actual_api_key

# Run with both configuration settings
export CONFIG_JSON='...'
./speelka-agent
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