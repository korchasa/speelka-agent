# System Architecture

## Overview
- Universal LLM agent (MCP-based)
- Modular, extensible, clean architecture

## Principles
- Single-responsibility components
- Dependency injection
- Interface-driven, testable
- Structured error handling
- Centralized config

## High-Level Flow
```mermaid
flowchart TB
    User["MCP Client"] -->|"Request"| Agent["Speelka Agent"]
    Agent -->|"Prompt"| LLM["LLM Service"]
    LLM -->|"Tool calls"| Agent
    Agent -->|"Exec tools"| Tools["MCP Tools"]
    Tools -->|"Results"| Agent
    Agent -->|"Repeat"| LLM
    Agent -->|"Final answer"| User
```

## Components
- **Agent**: Orchestrates flow, manages state, LLM loop, tool exec, chat state (token/cost tracked via LLMResponse, fallback estimation if needed)
- **Config Manager**: Loads/validates config (env, YAML, JSON), provides typed access, matches `types.Configuration` structure
- **LLM Service**: Handles LLM requests, retry logic, returns `LLMResponse` (text, tool calls, token/cost)
- **MCP Server**: Exposes agent (HTTP, stdio), manages tools, processes requests
- **MCP Connector**: Connects to external MCP servers, routes tool calls, manages connections
- **Chat**: Manages history, formatting, compaction, token/cost, context, all state in `chatInfo` struct, immutable config
- **Logger**: Wraps logrus, MCP protocol logging, client notifications

## Data Flow
1. User → MCP Server
2. Agent → Chat session
3. LLM Service (prompt + tools)
4. LLM → text/tool calls
5. MCP Connector → tool exec
6. Tool results → Chat
7. Token check/compaction
8. Repeat until answer
9. Response → User

## Error Handling
- Categories: Validation, Transient, Internal, External
- Retry: Per error type
- Context-rich, sanitized messages
- Graceful degradation
- Principles: Always check nil, safe assertions, descriptive errors, no panics

## Security
- API keys: env/secure storage
- Sanitized logs/errors
- HTTP transport security
- Tool access control

## Multi-Transport
- Daemon: HTTP server
- CLI: stdio

## Dependencies
- `mcp-go`: MCP impl
- `langchaingo`: LLM client
- `logrus`: Logging

## Config System
- Flexible: YAML, JSON, env
- Type-safe, validated, defaults
- Secure: API keys via env
- Only `Apply` parses log level/output
- Load order: default → file → env

## Testing
- Unit: 75%+ coverage, mocks
- Config: defaults, overrides, validation, transport
- Integration: component, config, API
- E2E: agent, transport, tools, token/cost/approximation

## Diagrams
### Request Flow
```mermaid
graph TD
    A[User] --> B[MCP Server]
    B --> C[Agent]
    C --> D[Chat]
    D --> DA[Token Counter]
    DA --> DB[Compaction]
    DB --> D
    D --> E[LLM Service]
    E --> F[LLM Provider]
    F --> E
    E --> D
    D --> G[MCP Connector]
    G --> H[Tools]
    H --> G
    G --> D
    D --> C
    C --> B
    B --> I[User]
```

### Config Structure
```mermaid
graph TD
    A[Env/Files] --> B[Config Manager]
    B --> C[Agent Config]
    B --> D[Runtime Config]
    D --> E[Log]
    D --> T[Transports]
    T --> TS[Stdio]
    T --> TH[HTTP]
    C --> CV[Version]
    C --> N[Name]
    C --> I[Tool]
    C --> J[Chat]
    C --> K[LLM]
    C --> L[Connections]
    I --> M[Tool Name/Desc]
    J --> O[Max Tokens]
    J --> P[Compaction]
    J --> Q[Max LLM Iter]
    K --> KS[Provider]
    K --> KM[Model]
    K --> KP[Prompt]
    K --> KR[Retry]
    L --> LS[Servers]
    L --> LR[Retry]
    KR --> KRM[Max Retries]
    KR --> KRI[Init Backoff]
    KR --> KRMX[Max Backoff]
    KR --> KRBM[Multiplier]
    LR --> LRM[Max Retries]
    LR --> LRI[Init Backoff]
    LR --> LRMX[Max Backoff]
    LR --> LRBM[Multiplier]
```
