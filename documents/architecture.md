# System Architecture

## Overview
Speelka Agent is a universal LLM agent based on Model Context Protocol (MCP), providing a modular, extensible system with clean architecture.

## Core Design Principles
1. **Separation of Concerns**: Each component has a single responsibility
2. **Dependency Injection**: Dependencies are provided via constructors
3. **Interface-Based Design**: Components implement interfaces for testability
4. **Error Handling Strategy**: Structured, categorized error handling
5. **Configuration Management**: Centralized configuration with component-specific subsets

## High-Level Architecture

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

## Key Components

### Agent
- Central orchestrator coordinating all components
- Manages conversation flow between user, LLM, and tools
- Processes user requests via MCP server
- Controls LLM interaction loop
- Delegates tool execution to MCP connector
- Maintains conversation state via Chat component

### Configuration Manager
- Centralized configuration access point
- Loads configuration from environment variables or files
- Provides typed access to configuration subsets
- Implements `ConfigurationManagerSpec` interface

### LLM Service
- Handles communication with LLM providers
- Supports multiple providers (OpenAI, Anthropic)
- Formats and sends requests to LLMs
- Processes responses and extracts tool calls
- Implements retry logic for transient errors

### MCP Server
- Exposes agent functionality to clients
- Supports HTTP and stdio protocols
- Registers and manages available tools
- Processes incoming requests
- Returns responses to clients

### MCP Connector
- Connects to external MCP servers for tool execution
- Manages connections to multiple MCP servers
- Discovers available tools from connected servers
- Routes tool calls to appropriate servers
- Returns tool execution results

### Chat
- Manages conversation history and formatting
- Maintains message history between user, assistant, and tools
- Formats prompts with templates
- Provides conversation context for LLM requests
- Tracks tool calls and results
- Implements token counting and history compaction
- Supports multiple compaction strategies to reduce context size
- Ensures conversations remain within LLM token limits

### MCPLogger
- Wraps logrus logging library with MCP capabilities
- Supports standard logging and MCP protocol-based logging to clients
- Maps between logrus levels and MCP protocol levels
- Sends "notifications/message" to connected clients
- Provides "logging/setLevel" tool for clients

## System Diagrams

### Request Flow Diagram

```mermaid
graph TD
    A[User Request] --> B[MCP Server]
    B --> C[Agent]
    C --> D[Chat]
    D --> DA[Token Counter]
    DA --> DB[History Compaction]
    DB --> D
    D --> E[LLM Service]
    E --> F[LLM Provider]
    F --> E
    E --> D
    D --> G[MCP Connector]
    G --> H[External Tools]
    H --> G
    G --> D
    D --> C
    C --> B
    B --> I[User Response]
```

### Configuration Structure

```mermaid
graph TD
    A[Environment Variables/Config Files] --> B[Configuration Manager]
    B --> C[Agent Config]
    B --> D[Runtime Config]
    D --> E[Log Config]
    D --> T[Transports Config]
    T --> TS[Stdio Config]
    T --> TH[HTTP Config]
    C --> CV[Version]
    C --> N[Name]
    C --> I[Tool Settings]
    C --> J[Chat Settings]
    C --> K[LLM Settings]
    C --> L[Connections Settings]
    I --> M[Tool Name and Description]
    J --> O[Max Tokens]
    J --> P[Compaction Strategy]
    J --> Q[Max LLM Iterations]
    K --> KS[Provider Settings]
    K --> KM[Model Settings]
    K --> KP[Prompt Template]
    K --> KR[Retry Settings]
    L --> LS[Server Connections]
    L --> LR[Retry Settings]
    KR --> KRM[Max Retries]
    KR --> KRI[Initial Backoff]
    KR --> KRMX[Max Backoff]
    KR --> KRBM[Backoff Multiplier]
    LR --> LRM[Max Retries]
    LR --> LRI[Initial Backoff]
    LR --> LRMX[Max Backoff]
    LR --> LRBM[Backoff Multiplier]
```

## Data Flow
1. User request → MCP Server
2. Agent processes request, initializes Chat session
3. LLM Service called with formatted prompt + available tools
4. LLM responds with text and/or tool calls
5. For each tool call, MCP Connector executes tool on appropriate server
6. Tool results added to Chat history
7. Token count checked and compaction applied if needed
8. Process repeats until LLM issues "answer" command
9. Final response returned to user via MCP Server

## Error Handling Philosophy
- **Categories**: Validation, Transient, Internal, External
- **Retry Strategies**: Appropriate mechanisms per error type
- **Context-rich messages**: Including relevant information without sensitive data
- **Graceful degradation**: System remains functional when components fail

### Error Handling Principles
1. **Always Check for Nil**: Before type assertions on interface values
2. **Safe Type Assertions**: Use two-return form
3. **Descriptive Error Messages**: Include expected vs actual type
4. **Return Graceful Errors**: Clear errors instead of panics

## Security Considerations
- API keys stored in environment variables or secure storage
- Sensitive information sanitized in errors and logs
- Transport security options for HTTP connections
- Access control for tool execution

## Multi-Transport Support
- **Daemon mode**: HTTP server for multi-client support
- **CLI mode**: Stdin/stdout for command-line usage

## External Dependencies
- `mcp-go`: Go implementation of Model Context Protocol
- `langchaingo`: Go client for LLM interaction
- `logrus`: Structured logging
- Standard Go libraries

## Configuration System

### Core Principles
- **Flexibility**: Multiple ways to configure (YAML, JSON, env vars)
- **Type Safety**: Strong typing ensures configuration validation
- **Defaults**: Sensible defaults with clear override mechanics
- **Validation**: Immediate error reporting on configuration issues
- **Security**: Sensitive values like API keys can be supplied via environment variables

### Configuration Loading Process
1. Load from default path
2. Override with `--config` path if provided
3. Apply environment variable overrides (with `SPL_` prefix)
4. Validate configuration integrity
5. Apply configuration to system components

### Security Considerations
- API keys should be provided via environment variables in production environments
- Example files include dummy API keys for testing purposes only
- Configuration files with real API keys should never be committed to version control

## Transport Layer

## Testing Strategy

### Unit Testing
- Each component has dedicated unit tests
- Mock interfaces for dependencies
- Test coverage requirements: minimum 75%

### Configuration Testing
- Test default values
- Test overriding via different methods (file, env vars)
- Test validation logic
- Test for correct application of configuration to components
- Test for proper handling of transport settings (HTTP, stdio)

### Integration Testing
- Tests interaction between components
- Tests the full configuration pipeline
- Tests the API endpoints

### End-to-End Testing
- Full agent test with mock LLM responses
- Transport tests (HTTP, stdio)
- Tool connection testing
