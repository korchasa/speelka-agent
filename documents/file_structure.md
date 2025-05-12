# File Structure

## Root
- `README.md`: Project overview
- `go.mod`, `go.sum`: Go modules
- `Makefile`, `Dockerfile`, `run`: Build, run, CI
- `LICENSE`: MIT
- `.github/`: CI/CD workflows
- `cmd/server/`: Main entry
- `internal/`: Core logic
- `site/`: Web UI, config examples
- `vendor/`: Vendored deps
- `documents/`: Docs

## Key Directories
- `cmd/server/`: Entrypoint (`main.go`)
- `configs/`: Templates
- `documents/`: Docs (see below)
- `internal/`: All core packages
- `site/`: Web, config examples, assets
- `vendor/`: Vendored Go deps

## Internal Structure
```mermaid
flowchart TD
    cmd-->internal/app
    internal/app-->internal/agent
    internal/app-->internal/chat
    internal/app-->internal/llm_service
    internal/app-->internal/mcp_connector
    internal/app-->internal/mcp_server
    internal/app-->internal/types
    internal/app-->internal/utils
    internal/agent-->internal/chat
    internal/agent-->internal/llm_service
    internal/agent-->internal/mcp_connector
    internal/agent-->internal/types
    internal/agent-->internal/utils
    internal/chat-->internal/types
    internal/chat-->internal/utils
    internal/llm_service-->internal/error_handling
    internal/llm_service-->internal/types
    internal/llm_service-->internal/utils
    internal/mcp_connector-->internal/error_handling
    internal/mcp_connector-->internal/types
    internal/mcp_connector-->internal/utils
    internal/mcp_server-->internal/types
    internal/mcp_server-->internal/utils
    internal/logger-->internal/types
    internal/logger-->external_mcp
    internal/configuration-->internal/types
    internal/types-->external_mcp[github.com/mark3labs/mcp-go]
    internal/types-->external_llm[github.com/tmc/langchaingo]
    style cmd fill:#f9f,stroke:#333
    style external_mcp fill:#eee,stroke:#333,stroke-dasharray: 5 5
    style external_llm fill:#eee,stroke:#333,stroke-dasharray: 5 5
```

### Key Packages
- `internal/agent`: Core agent logic (LLM, tool orchestration, session management). No config loading, server, CLI, or direct call JSON types.
- `internal/app`: Application wiring, orchestration, lifecycle, CLI. Instantiates and manages the agent, provides CLI entry points. Includes `App` (server/daemon mode) and `DirectApp` (CLI direct-call mode, independent from `App`). Shared stateless utilities for config loading, agent instantiation, etc.

## Examples (site/examples/)
- `minimal.yaml`, `ai-news.yaml`, `architect.yaml`: Agent configs (YAML, preferred), include `agent.chat.request_budget` (limit on total cost per request)

## Dependencies
| Package | Use |
|---------|-----|
| github.com/mark3labs/mcp-go | MCP impl |
| github.com/tmc/langchaingo | LLM client |
| github.com/sirupsen/logrus | Logging |
| github.com/pkoukk/tiktoken-go | Token count |

## Documents
- `architecture.md`: System design
- `file_structure.md`: This file
- `implementation.md`: Implementation
- `knowledge.md`: Code/protocol refs
- `remote_resources.md`: External links
- `whiteboard.md`: Temp planning (ephemeral)
