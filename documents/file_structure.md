# File Structure

## Root
- `README.md`: Project overview and instructions
- `go.mod`, `go.sum`: Go modules
- `Dockerfile`: Container build
- `run`: Build/test/check script
- `.gitignore`, `.cursorignore`, `.cursorrules`: Ignore/config rules
- `.github/`: CI/CD workflows
- `bin/`: Built binaries
- `cmd/`: Entrypoints (server, mcp-call)
- `internal/`: Core logic (see below)
- `site/`: Web UI, config examples, static assets
- `vendor/`: Vendored Go dependencies
- `documents/`: Project documentation
- `LICENSE`: License

## cmd/
- `server/`: Main server entrypoint
- `mcp-call/`: MCP call utilities

## internal/
- `agent/`: Core agent logic
- `app_mcp/`: MCP server app wiring
- `app_direct/`: Direct CLI call app wiring
- `chat/`: Chat/session logic
- `configuration/`: Config loading/validation
- `error_handling/`: Error handling utilities
- `llm_models/`: LLM token/cost logic
- `llm_service/`: LLM service abstraction
- `logger/`: Logging
- `mcp_connector/`: MCP server connection logic
    - mcp_connector.go — MCPConnector implementation, public methods, delegation
    - connection.go — MCP client connection and initialization logic
    - logging.go — log routing (MCP logs or fallback to stderr)
    - mcp_connector_test.go — tests for log routing (MCP and stderr)
    - utils_test.go — helper functions for testing
- `mcp_server/`: MCP server implementation
- `types/`:
    - logger_spec.go — LogConfig, LoggerSpec, MCPServerNotifier interfaces
    - configuration_test.go — golden serialization and overlay property-based tests

## Test Data
- `internal/types/testdata/configuration_golden.json`: Golden file for config serialization tests