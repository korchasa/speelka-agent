# Implementation Details

## Speelka Agent Backend (Go)

### Core Components

#### Agent
- **Purpose**: Request processing coordinator
- **File**: `internal/agent/agent.go`
- **Features**:
  - Request lifecycle mgmt
  - Component coordination
  - Tool execution
  - State/context mgmt
- **Improvements**:
  - Robust null checking for interface values
  - Safer type assertion pattern
  - Graceful nil value handling

#### Chat
- **Purpose**: Conversation/prompt mgmt
- **File**: `internal/chat/chat.go`
- **Features**:
  - Message history tracking
  - System prompt formatting
  - Tool description building
  - Jinja2 template support

#### Configuration Manager
- **Purpose**: Config mgmt
- **File**: `internal/configuration/manager.go`
- **Features**:
  - JSON config via `CONFIG_JSON` env var
  - Targeted env var overrides
  - Type-safe config access
  - Default handling
  - Validation

#### LLM Service
- **Purpose**: LLM integration
- **File**: `internal/llm_service/llm_service.go`
- **Features**:
  - OpenAI/Anthropic support
  - Request/response handling
  - Retry logic
  - Error categorization

#### MCP Server
- **Purpose**: MCP protocol impl
- **File**: `internal/mcp_server/mcp_server.go`
- **Features**:
  - HTTP/stdio transport
  - Request handling
  - Debug hooks
  - Graceful shutdown
  - SSE for real-time comms

##### HTTP Server
- Uses SSE server from `github.com/mark3labs/mcp-go/server`
- In daemon mode (`agent.Start(true, ctx)`), exposes:
  - `/sse`: SSE connection endpoint
  - `/message`: HTTP POST endpoint for tools

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

#### MCP Connector
- **Purpose**: External tool mgmt
- **File**: `internal/mcp_connector/mcp_connector.go`
- **Features**:
  - Server connection mgmt
  - Tool discovery
  - Tool execution
  - Connection retry logic

### Key Interfaces

#### Configuration
```go
type ConfigurationManagerSpec interface {
    LoadConfiguration(ctx context.Context) error
    GetMCPServerConfig() MCPServerConfig
    GetMCPConnectorConfig() MCPConnectorConfig
    GetLLMConfig() LLMConfig
    GetLogConfig() LogConfig
}
```

#### LLM Service
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

#### MCP Server
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

#### MCP Connector
```go
type MCPConnectorConfig struct {
    Servers []MCPServerConnection
    RetryConfig RetryConfig
}
```

### Error Handling

#### Categories
- **Validation**: Config/input validation errors
- **Transient**: Network/rate limit errors
- **Internal**: Component/runtime errors
- **External**: Tool execution errors

#### Retry Strategy
```go
type RetryConfig struct {
    MaxRetries int
    InitialBackoff float64
    MaxBackoff float64
    BackoffMultiplier float64
}
```

### Request Processing Flow

1. **Reception**: HTTP/stdio transport → validation → context creation
2. **Processing**: Chat history init → tool discovery → LLM interaction
3. **LLM Interaction**: Prompt format → request send → response parse → tool call extract
4. **Tool Execution**: Tool lookup → request forward → result capture → error handle
5. **Response**: Result format → response send → resource cleanup

### Configuration

#### Environment Variables

```bash
# Agent
export AGENT_NAME="architect-speelka-agent"
export AGENT_VERSION="1.0.0"

# Tool
export TOOL_NAME="architect"
export TOOL_DESCRIPTION="Architecture design and assessment tool"
export TOOL_ARGUMENT_NAME="query"
export TOOL_ARGUMENT_DESCRIPTION="Architecture query or task to process"

# LLM
export LLM_PROVIDER="openai"
export LLM_API_KEY="your_api_key_here"
export LLM_MODEL="gpt-4o"
export LLM_MAX_TOKENS=0
export LLM_TEMPERATURE=0.2
export LLM_PROMPT_TEMPLATE="# ROLE
You are a Senior Software Architect...
# User query
{{query}}
# Available tools
{{tools}}"

# LLM Retry Config
export LLM_RETRY_MAX_RETRIES=3
export LLM_RETRY_INITIAL_BACKOFF=1.0
export LLM_RETRY_MAX_BACKOFF=30.0
export LLM_RETRY_BACKOFF_MULTIPLIER=2.0

# MCP Servers (indexed: MCPS_0, MCPS_1, etc.)
export MCPS_0_ID="time"
export MCPS_0_COMMAND="docker"
export MCPS_0_ARGS="run -i --rm mcp/time"

export MCPS_1_ID="mcp-filesystem-server"
export MCPS_1_COMMAND="mcp-filesystem-server"
export MCPS_1_ARGS="."

# MSPS Retry
export MSPS_RETRY_MAX_RETRIES=3
export MSPS_RETRY_INITIAL_BACKOFF=1.0
export MSPS_RETRY_MAX_BACKOFF=30.0
export MSPS_RETRY_BACKOFF_MULTIPLIER=2.0

# Runtime
export RUNTIME_LOG_LEVEL="debug"
export RUNTIME_LOG_OUTPUT="./architect.log"
export RUNTIME_STDIO_ENABLED=true
export RUNTIME_STDIO_BUFFER_SIZE=8192
export RUNTIME_HTTP_ENABLED=false
export RUNTIME_HTTP_HOST="localhost"
export RUNTIME_HTTP_PORT=3000
```

#### Required Env Vars
- `AGENT_NAME`: Agent name
- `TOOL_NAME`: Tool name
- `TOOL_DESCRIPTION`: Tool description
- `TOOL_ARGUMENT_NAME`: Tool argument name
- `TOOL_ARGUMENT_DESCRIPTION`: Tool argument description
- `LLM_PROVIDER`: LLM provider ("openai")
- `LLM_MODEL`: Model name ("gpt-4o")
- `LLM_PROMPT_TEMPLATE`: System prompt template (must include {{query}} and {{tools}})

#### MCP Servers Config Format
```
MCPS_<index>_ID="server-id"
MCPS_<index>_COMMAND="command"
MCPS_<index>_ARGS="arg1 arg2 arg3"
```
- `<index>`: 0-based index for each server
- `ID`: Key in MCP servers map

#### Legacy JSON Config (Deprecated)
Single `CONFIG_JSON` env var with complete JSON configuration (see examples).

## Configuration Website

### Purpose
User-friendly interface for:
1. Configuring Speelka Agent
2. Generating env var configs
3. Viewing docs/examples
4. Testing/validating configs

### Structure
- **HTML**: Main structure (`site/index.html`)
- **CSS**: Styling (`site/css/styles.css`)
- **JS**: Functionality (`site/js/main.js`)
- **Images**: Visual elements (`site/img/`)

### Form Validation

#### Key Components
1. **Validation Functions**:
   - `validateField()`: Central field validation
   - `updateValidationUI()`: UI error display
   - `setupFormValidation()`: Event delegation setup

2. **Event Handling**:
   - Event delegation vs individual listeners
   - Debouncing prevents excessive updates
   - Consolidated event handlers

#### Validation Flow
1. User interaction → event delegation
2. Validation function checks requirements
3. UI updates with validation status
4. Config generated only when validation passes

### Performance Optimizations

1. **Lazy Loading**:
   - Images: Intersection Observer API
   - Non-critical JS: defer loading

2. **DOM Manipulation**:
   - Batch operations to minimize reflows
   - Complete element creation before DOM insertion
   - DocumentFragment for complex elements

3. **Event Handling**:
   - Debouncing for input-heavy operations
   - Event delegation vs individual listeners
   - Minimized redundant handlers

### Configuration System

1. **Env Var Focus**:
   - Primary generation as env vars
   - Removed deprecated JSON config support
   - Clear section headers in output

2. **Improved Usability**:
   - Better visual organization
   - Clear field-to-config connections
   - Consistent error handling

### CSS Improvements

1. **Media Query Consolidation**:
   - Consolidated redundant queries
   - Grouped related styles by breakpoint
   - Improved organization by device size

2. **Animation Optimization**:
   - Essential keyframes only
   - Removed unused animations
   - Simplified transitions

3. **Style Organization**:
   - Related styles grouped
   - Improved selector specificity
   - Reduced redundancy

### Future Website Recommendations

1. **CSS Modularization**:
   - Component-specific CSS files
   - CSS preprocessor (SASS/LESS)
   - CSS modules or CSS-in-JS

2. **JS Modularization**:
   - Functional JS modules
   - Build process for efficient bundling
   - Unit tests for core functionality

3. **Accessibility**:
   - Comprehensive a11y audit
   - Additional ARIA attributes
   - Consistent keyboard navigation

### Website Functionality

The Speelka Agent website provides a simplified interface with the following features:

1. **Core Functionality**: The website uses vanilla JavaScript to handle basic functionality:
   - Lazy loading of images for improved performance
   - Throttled event handling to optimize scrolling
   - Error handling and logging
   - Responsive design with mobile support

2. **Navigation**: Simple navigation through documentation sections

### Performance Optimization

The website implements several performance optimization techniques:

1. **Lazy Loading**: Images are loaded only when they enter the viewport
2. **Throttling**: Event handlers (like scroll) are throttled to reduce unnecessary function calls
3. **Minimal Dependencies**: No external JavaScript libraries are used to keep the bundle size small

### Testing

To test the website functionality:

1. Open the site in different browsers
2. Verify all navigation functions correctly
3. Test responsive design by resizing window
4. Check all images load properly with lazy loading
5. Ensure error handling captures and logs issues appropriately

### Error Handling

The website implements a simple error handling strategy:

1. All JavaScript functions are wrapped in try-catch blocks
2. Errors are logged to console with descriptive messages
3. User-friendly error messages are displayed when appropriate

### Future Improvements

Potential areas for improvement include:

1. Adding dark mode support
2. Implementing better accessibility features
3. Adding more interactive examples
4. Improving documentation with searchable content

## Bug Fixes

### Nil Pointer Dereference in Logger
A nil pointer dereference bug was fixed in the application startup sequence:

**Problem:** The global `log` variable in `cmd/server/main.go` was being used before initialization, causing a nil pointer dereference when attempting to log an error.

**Root cause:** The logger initialization was happening in the `run()` function, but was needed earlier in the `main()` function.

**Solution:**
- Moved basic logger initialization to the beginning of the `main()` function
- Kept the detailed logger configuration in the `run()` function
- This ensures the logger is always initialized before use

**Implementation:**
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
log.SetReportCaller(true)
log.SetFormatter(utils.NewCustomLogFormatter())
```

**Impact:** Prevents application crash on startup when encountering errors early in the initialization process.