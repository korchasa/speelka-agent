# File Structure

## Root
- `README.md`: Project overview and instructions
- `go.mod`, `go.sum`: Go modules
- `Dockerfile`: Container build
- `run`: Build/test/check script
- `.gitignore`, `.cursorignore`, `.cursorrules`: Ignore/config rules
- `.github/`: CI/CD workflows
- `bin/`: Built binaries
- `cmd/`: Entrypoints (server, mcp-call, test-mcp-logging)
- `internal/`: Core logic (see below)
- `site/`: Web UI, config examples, static assets
- `vendor/`: Vendored Go dependencies (see below)
    - `github.com/knadh/koanf/v2`: Core configuration library
    - `github.com/knadh/koanf/providers/file`: File provider
    - `github.com/knadh/koanf/providers/env`: Env provider
    - `github.com/knadh/koanf/providers/confmap`: Confmap provider
    - `github.com/knadh/koanf/providers/structs`: Structs provider
    - `github.com/knadh/koanf/parsers/json`: JSON parser
    - `github.com/knadh/koanf/parsers/yaml`: YAML parser
    - `github.com/knadh/koanf/parsers/toml`: TOML parser
- `documents/`: Project documentation
- `LICENSE`: License

## cmd/
- `server/`: Main MCP server/daemon entrypoint (uses app_mcp)
- `mcp-call/`: Standalone MCP call/test utility (for E2E and protocol tests)
- `test-mcp-logging/`: Standalone test server/client for MCP logging

## internal/
- `agent/`: Core agent logic (protocol-agnostic, no MCP/CLI logic)
- `app_mcp/`: MCP server/daemon app wiring (uses NewAgentServerMode, DispatchMCPCall)
- `app_direct/`: Direct CLI call app wiring (uses NewAgentCLI with real MCP connector to load tools)
    - `app.go`: CLI application entrypoint
    - `types.go`: Types for CLI mode
- `chat/`: Chat/session logic
- `configuration/`: Config loading and validation (koanf-based, no custom loaders; all config structs use koanf tags only)
- `error_handling/`: Error handling utilities
- `llm_models/`: LLM model-specific utilities (e.g., cost calculation)
- `llm_service/`: LLM service abstraction and retry logic
- `logger/`: Logging utilities and spec
- `mcp_connector/`: MCP server connection logic
    - `mcp_connector.go`: ToolConnector implementation, public methods
    - `connection.go`: MCP client connection and initialization logic
    - `logging.go`: Log routing (MCP logs or fallback to stderr)
- `mcp_server/`: MCP server implementation
- `types/`: Type definitions and interfaces
    - `testdata/`: Test data for types
- `utils/`: Utility functions