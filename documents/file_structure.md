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
    - mcp_connector.go — реализация MCPConnector, публичные методы, делегирование
    - connection.go — логика подключения и инициализации MCP клиентов
    - logging.go — маршрутизация логов (MCP-логи или fallback на stderr)
    - mcp_connector_test.go — тесты для проверки маршрутизации логов (MCP и stderr)
    - utils_test.go — вспомогательные функции для тестирования
- `mcp_server/`: MCP server implementation
- `