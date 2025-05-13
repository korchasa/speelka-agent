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
- `mcp_server/`: MCP server implementation
- `types/`: Shared types/specs
- `utils/`: Misc utilities

## site/
- `index.html`: Web UI
- `css/`, `js/`, `img/`: Static assets
- `examples/`: Example agent configs
- `sitemap.xml`, `robots.txt`: SEO

## vendor/
- Vendored Go modules and dependencies

## documents/
- `architecture.md`: System design
- `file_structure.md`: This file
- `implementation.md`: Implementation details
- `knowledge.md`: Code/protocol refs
- `remote_resources.md`: External links
- `whiteboard.md`: Temp planning (ephemeral)
- `mcp-go.xml`, `model-context-protocol.xml`: Protocol/library docs

// All obsolete files and references have been removed.

Везде в конфиге используется runtime.log.default_level вместо runtime.log.level.
MCPLogger интегрируется с MCPServer через интерфейс MCPServerNotifier.
