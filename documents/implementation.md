# Implementation Details

## Core Components

### Agent
- **Purpose**: Central request processing coordinator
- **File**: `internal/agent/agent.go`
- **Implementation Features**:
  - Request lifecycle management
  - Component coordination
  - Tool execution orchestration
  - State and context management
- **Technical Improvements**:
  - Robust null checking for interface values
  - Safer type assertion pattern
  - Graceful nil value handling

### Chat
- **Purpose**: Conversation and prompt management
- **File**: `internal/chat/chat.go`
- **Implementation Features**:
  - Message history tracking
  - System prompt formatting
  - Tool description building
  - Jinja2-style template support

### Configuration Manager
- **Purpose**: Configuration management
- **File**: `internal/configuration/manager.go`
- **Implementation Features**:
  - JSON configuration via `CONFIG_JSON` env var (legacy support)
  - Type-safe configuration access
  - Default value handling
  - Configuration validation
  - Log file path handling (values other than stdout/stderr are treated as file paths)

### LLM Service
- **Purpose**: LLM provider integration
- **File**: `internal/llm_service/llm_service.go`
- **Implementation Features**:
  - OpenAI and Anthropic support
  - Request and response handling
  - Retry logic with configurable backoff
  - Error categorization

### MCP Server
- **Purpose**: MCP protocol implementation
- **File**: `internal/mcp_server/mcp_server.go`
- **Implementation Features**:
  - HTTP and stdio transport support
  - Request routing and handling
  - Debug hooks
  - Graceful shutdown
  - Server-Sent Events for real-time communication

#### HTTP Server Implementation
- Uses SSE server from `github.com/mark3labs/mcp-go/server`
- In daemon mode (`agent.Start(true, ctx)`), exposes:
  - `/sse`: Server-Sent Events connection endpoint
  - `/message`: HTTP POST endpoint for tool calls

**Request Format**:
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

**Response Format**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "Response from the LLM or tool",
  "isError": false
}
```

### MCP Connector
- **Purpose**: External tool management
- **File**: `internal/mcp_connector/mcp_connector.go`
- **Implementation Features**:
  - Server connection management
  - Tool discovery
  - Tool execution
  - Connection retry logic

### MCPLogger
- **Purpose**: MCP-compatible logging
- **File**: `internal/mcplogger/mcplogger.go`
- **Implementation Features**:
  - Wraps logrus with MCP capabilities
  - Level mapping between logrus and MCP
  - Structured data support
  - `logging/setLevel` tool support

#### Level Mapping

| Logrus Level | MCP Level |
|--------------|-----------|
| TraceLevel   | debug     |
| DebugLevel   | debug     |
| InfoLevel    | info      |
| WarnLevel    | warning   |
| ErrorLevel   | error     |
| FatalLevel   | critical  |
| PanicLevel   | alert     |

#### Notification Format
```json
{
  "level": "info",
  "message": "Log message text",
  "data": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

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
    TemperatureIsSet bool
    MaxTokensIsSet bool
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

## Error Handling Implementation

### Error Categories
- **Validation Errors**: Configuration and input validation failures
- **Transient Errors**: Network, rate limit, and temporary service errors
- **Internal Errors**: Component and runtime errors
- **External Errors**: Tool execution and external service errors

### Retry Strategy Implementation
```go
type RetryConfig struct {
    MaxRetries int
    InitialBackoff float64
    MaxBackoff float64
    BackoffMultiplier float64
}
```

Implementation:
```go
// Retry with backoff
err = error_handling.RetryWithBackoff(ctx, sendFn, error_handling.RetryConfig{
    MaxRetries:        s.config.RetryConfig.MaxRetries,
    InitialBackoff:    time.Duration(s.config.RetryConfig.InitialBackoff * float64(time.Second)),
    BackoffMultiplier: s.config.RetryConfig.BackoffMultiplier,
    MaxBackoff:        time.Duration(s.config.RetryConfig.MaxBackoff * float64(time.Second)),
})
```

## Request Processing Implementation

1. **Reception**:
   - Transport layer receives request (HTTP/stdio)
   - Request validation
   - Context creation with timeout

2. **Processing**:
   - Chat history initialization
   - Tool discovery from MCP servers
   - LLM interaction loop setup

3. **LLM Interaction**:
   - Prompt formatting with templates
   - Request sending with tools
   - Response parsing
   - Tool call extraction

4. **Tool Execution**:
   - Tool lookup in available tools
   - Request forwarding to appropriate server
   - Result capture
   - Error handling

5. **Response**:
   - Result formatting
   - Response sending
   - Resource cleanup

## Configuration Implementation

The system primarily uses environment variables for configuration, with a common `SPL_` prefix. For backward compatibility, the system also accepts environment variables without the prefix.

Example environment variable configuration:

```bash
# Agent
export SPL_AGENT_NAME="architect-speelka-agent"
export SPL_AGENT_VERSION="1.0.0"

# Tool
export SPL_TOOL_NAME="architect"
export SPL_TOOL_DESCRIPTION="Architecture design and assessment tool"
export SPL_TOOL_ARGUMENT_NAME="query"
export SPL_TOOL_ARGUMENT_DESCRIPTION="Architecture query or task to process"

# LLM
export SPL_LLM_PROVIDER="openai"
export SPL_LLM_API_KEY="your_api_key_here"
export SPL_LLM_MODEL="gpt-4o"
export SPL_LLM_MAX_TOKENS=0
export SPL_LLM_TEMPERATURE=0.2

# Retry configuration
export SPL_LLM_RETRY_MAX_RETRIES=3
export SPL_LLM_RETRY_INITIAL_BACKOFF=1.0
export SPL_LLM_RETRY_MAX_BACKOFF=30.0
export SPL_LLM_RETRY_BACKOFF_MULTIPLIER=2.0
```

### MCP Servers Configuration Format
```bash
SPL_MCPS_<index>_ID="server-id"
SPL_MCPS_<index>_COMMAND="command"
SPL_MCPS_<index>_ARGS="arg1 arg2 arg3"
```

Where:
- `<index>`: 0-based index for each server
- `ID`: Key in MCP servers map
- `COMMAND`: Command to execute to start the server
- `ARGS`: Space-separated arguments for the command

## Bug Fixes and Improvements

### Nil Pointer Dereference in Logger

**Problem:** The global `log` variable in `cmd/server/main.go` was being used before initialization, causing a nil pointer dereference when attempting to log an error.

**Solution:**
```go
// In main() function - early initialization
log = logrus.New()
log.SetLevel(logrus.InfoLevel)
log.SetOutput(os.Stderr)

// In run() function - detailed configuration
if *daemonMode {
    log.SetLevel(logrus.DebugLevel)
    log.SetOutput(os.Stdout)
} else {
    log.SetLevel(logrus.DebugLevel)
    log.SetOutput(os.Stderr)
}
```

### Conditional LLM Parameters

**Problem:** The LLM service was always including `temperature` and `maxTokens` parameters in requests regardless of whether they were explicitly configured.

**Solution:**
- Added tracking flags in `LLMConfig` to record whether parameters were explicitly set
- Modified configuration loading to set these flags when environment variables are present
- Updated request creation to conditionally include parameters only when explicitly configured

```go
// Configuration flags
type LLMConfig struct {
    // ... existing fields ...
    MaxTokens int
    Temperature float64
    TemperatureIsSet bool
    MaxTokensIsSet bool
    // ... existing fields ...
}

// In request creation
if s.config.MaxTokensIsSet {
    requestBody["max_tokens"] = s.config.MaxTokens
}
if s.config.TemperatureIsSet {
    requestBody["temperature"] = s.config.Temperature
}
```

## Run Script Commands

The `run` script provides a unified interface for common operations:

### Development Commands
- `./run dev`: Run in development mode
- `./run build`: Build the project
- `./run test`: Run tests with coverage
- `./run lint`: Run linting
- `./run check`: Run all checks (test, lint, build)

### Interaction Commands
- `./run call`: Test with simple "What time is it now?" request
- `./run complex-call`: Test with complex file-finding request
- `./run call-news`: Test news agent
- `./run fetch_url <url>`: Fetch URL using MCP

### Inspection Command
- `./run inspect`: Run with MCP inspector
  - Collects environment variables with `SPL_` prefix
  - Passes them to the inspector with proper handling