# File Structure

## Root
- `README.md`: Project overview
- `go.mod`, `go.sum`: Go modules
- `Dockerfile`: Container build
- `run`: Build/test/check script
- `.gitignore`, `.cursorignore`, `.cursorrules`: Ignore/config rules
- `.github/`: CI/CD workflows (`workflows/`)
- `.junie/`: Project guidelines
- `bin/`: (empty, for built binaries)
- `cmd/`: Entrypoints (see below)
- `internal/`: Core logic (see below)
- `site/`: Web UI, config examples, assets (see below)
- `vendor/`: Vendored Go dependencies
- `documents/`: Project documentation (see below)
- `LICENSE`: MIT

## cmd/
- `server/`: Main server entrypoint
- `mcp-call/`: (subdir for MCP call utilities)

## internal/
- `agent/`: Core agent logic (`agent.go`)
- `app/`: (empty)
- `app_mcp/`: MCP server app wiring (`app.go`, `direct_app_test.go`, `util.go`)
- `app_direct/`: Direct CLI call app wiring (`direct_app.go`, `direct_types.go`)
- `chat/`: Chat/session logic (`chat.go`, `chat_test.go`)
- `configuration/`: Config loading/validation (`manager.go`, loaders, tests)
- `direct_app/`: (empty)
- `error_handling/`: Error handling utilities
- `llm_models/`: LLM token/cost logic, calculators
- `llm_service/`: LLM service abstraction
- `logger/`: Logging (logrus wrapper, formatter, entry)
- `mcp_connector/`: MCP server connection logic
- `mcp_server/`: MCP server implementation
- `types/`: All shared types/specs
- `utils/`: Misc utilities

## site/
- `index.html`: Web UI
- `css/`, `js/`, `img/`: Static assets
- `examples/`: Example agent configs:
  - `minimal.yaml`, `ai-news.yaml`, `infra-news.yaml`, `architect.yaml`, `all-options.yaml`, `text-extractor.yaml`
  - **Per-server timeout:** Each MCP server in `connections.mcpServers` can have a `timeout` parameter (seconds, float or int, default 30s if not set).
  - `ai-news-subagent-extractor.yaml` is obsolete and replaced by `text-extractor.yaml`.
- `sitemap.xml`, `robots.txt`: SEO

## vendor/
- `modules.txt`: Vendored Go modules
- `github.com/`, `golang.org/`, `gopkg.in/`: Vendored dependencies

## documents/
- `architecture.md`: System design
- `file_structure.md`: This file
- `implementation.md`: Implementation details
- `knowledge.md`: Code/protocol refs
- `remote_resources.md`: External links
- `whiteboard.md`: Temp planning (ephemeral)
- `mcp-go.xml`, `model-context-protocol.xml`: Protocol/library docs
- `.gitignore`: Ignore rules for docs

## File Removals (2024-06)
- Deleted: `internal/app/direct_app_test.go`, `internal/app/direct_types.go`, `internal/app/util.go`, `site/examples/ai-news-subagent-extractor.yaml` (obsolete, replaced by `text-extractor.yaml`).
- All references and tests for these files have been removed or updated.

Везде в конфиге используется runtime.log.default_level вместо runtime.log.level.
MCPLogger интегрируется с MCPServer через интерфейс MCPServerNotifier.
