// Package mcp_server provides functionality for the MCP server.
// Responsibility: Implementation of the MCP server for processing client requests
// Features: Supports two operating modes - HTTP (daemon) and stdio
package mcp_server

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer implements an MCP server for handling client requests and managing the lifecycle of tools.
// Thread-safe for public methods. All external dependencies are injected via the constructor (DI).
type MCPServer struct {
	server    *server.MCPServer     // Internal MCP server
	config    types.MCPServerConfig // Server configuration
	logger    types.LoggerSpec      // Logger (DI)
	sseServer *server.SSEServer     // HTTP SSE server (optional)

	mainToolHandler server.ToolHandlerFunc // handler for the main tool
	mu              sync.Mutex             // Protects the state of server/sseServer
}

// NewMCPServer creates a new instance of MCPServer with the given configuration and logger.
// All dependencies are injected via parameters (Dependency Injection).
func NewMCPServer(config types.MCPServerConfig, logger types.LoggerSpec) *MCPServer {
	var opts []server.ServerOption
	if config.MCPLogEnabled {
		opts = append(opts, server.WithLogging())
	}
	if config.Debug {
		opts = append(opts, server.WithHooks((&MCPServer{config: config, logger: logger}).BuildHooks()))
	}

	mcpSrv := server.NewMCPServer(
		config.Name,
		config.Version,
		opts...,
	)

	mcps := &MCPServer{
		server: mcpSrv,
		config: config,
		logger: logger,
	}

	// Register tools immediately
	for _, tool := range mcps.buildTools() {
		fmt.Fprintf(os.Stderr, "[MCPServer] Registering tool: %s\n", tool.Name)
		var h server.ToolHandlerFunc = nil
		if tool.Name == config.Tool.Name {
			h = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				fmt.Fprintf(os.Stderr, "[MCPServer] Handler called for %s\n", tool.Name)
				mcps.mu.Lock()
				fmt.Fprintf(os.Stderr, "[MCPServer] handler: mutex acquired\n")
				handler := mcps.mainToolHandler
				mcps.mu.Unlock()
				fmt.Fprintf(os.Stderr, "[MCPServer] handler: mutex released\n")
				if handler == nil {
					fmt.Fprintf(os.Stderr, "[MCPServer] mainToolHandler not set for %s\n", tool.Name)
					return nil, fmt.Errorf("main tool handler is not set for '%s'", tool.Name)
				}
				fmt.Fprintf(os.Stderr, "[MCPServer] handler: calling mainToolHandler for %s\n", tool.Name)
				res, err := handler(ctx, req)
				fmt.Fprintf(os.Stderr, "[MCPServer] handler: mainToolHandler finished for %s, err=%v\n", tool.Name, err)
				return res, err
			}
		} else if tool.Name == "logging/setLevel" {
			h = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				fmt.Fprintf(os.Stderr, "[MCPServer] Handler called for logging/setLevel\n")
				res, err := logger.HandleMCPSetLevel(ctx, req)
				if err != nil {
					return nil, err
				}
				result, ok := res.(*mcp.CallToolResult)
				if !ok {
					return nil, fmt.Errorf("unexpected result type from HandleMCPSetLevel")
				}
				return result, nil
			}
		}
		mcps.server.AddTool(tool, h)
	}

	return mcps
}

// Serve starts the MCP server in daemon (HTTP SSE) or script (stdio) mode.
// Thread-safe. Releases resources before completion.
func (s *MCPServer) Serve(ctx context.Context, daemonMode bool, handler server.ToolHandlerFunc) error {
	fmt.Fprintf(os.Stderr, "[MCPServer] Serve: entry, daemonMode=%v\n", daemonMode)
	if daemonMode {
		s.logger.Info("Running in daemon mode with HTTP SSE MCP server")
		if err := s.initSSEServer(handler); err != nil {
			return fmt.Errorf("failed to start HTTP MCP server: %w", err)
		}
	} else {
		s.logger.Info("Running in script mode with stdio MCP server")
		if err := s.initStdioServer(handler, ctx); err != nil {
			return fmt.Errorf("failed to start Stdio MCP Server: %w", err)
		}
	}
	fmt.Fprintf(os.Stderr, "[MCPServer] Serve: finished\n")
	return nil
}

// ServeStdioWithContext starts the stdio MCP server with external context support.
// Used for integration and testing.
func ServeStdioWithContext(mcpSrv *server.MCPServer, logger types.LoggerSpec, ctx context.Context) error {
	if mcpSrv != nil {
		return server.NewStdioServer(mcpSrv).Listen(ctx, os.Stdin, os.Stdout)
	}
	return fmt.Errorf("mcpSrv is not *server.MCPServer")
}

// --- Приватные orchestration-функции ---

// initSSEServer инициализирует и запускает HTTP SSE MCP сервер.
func (s *MCPServer) initSSEServer(handler server.ToolHandlerFunc) error {
	if s.server == nil {
		return fmt.Errorf("server is not *server.MCPServer")
	}
	s.logger.Info("MCP SSE server initialized successfully")
	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
	baseUrl := fmt.Sprintf("http://%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
	s.sseServer = server.NewSSEServer(s.server, server.WithBaseURL(baseUrl))
	if err := s.sseServer.Start(addr); err != nil {
		return fmt.Errorf("failed to serve SSE MCP server: %w", err)
	}
	return nil
}

// initStdioServer инициализирует и запускает stdio MCP сервер с поддержкой внешнего контекста.
func (s *MCPServer) initStdioServer(handler server.ToolHandlerFunc, ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "[MCPServer] initStdioServer: entry\n")
	if s.server == nil {
		fmt.Fprintf(os.Stderr, "[MCPServer] initStdioServer: server == nil\n")
		return fmt.Errorf("server is not *server.MCPServer")
	}
	s.logger.Info("MCP Stdio server initialized successfully")
	fmt.Fprintf(os.Stderr, "[MCPServer] initStdioServer: starting ServeStdioWithContext\n")
	return ServeStdioWithContext(s.server, s.logger, ctx)
}

// buildMainTool создаёт основной инструмент сервера.
func (s *MCPServer) buildMainTool() mcp.Tool {
	return mcp.NewTool(s.config.Tool.Name,
		mcp.WithDescription(s.config.Tool.Description),
		mcp.WithString(s.config.Tool.ArgumentName,
			mcp.Description(s.config.Tool.ArgumentDescription),
			mcp.Required(),
		),
	)
}

// buildLoggingTool создаёт инструмент для управления логированием.
func (s *MCPServer) buildLoggingTool() mcp.Tool {
	return mcp.NewTool("logging/setLevel",
		mcp.WithString("level", mcp.Required(), mcp.Description("Log level to set")),
	)
}

// buildTools возвращает список всех инструментов для регистрации на сервере.
func (s *MCPServer) buildTools() []mcp.Tool {
	tools := []mcp.Tool{s.buildMainTool()}
	if s.config.MCPLogEnabled {
		tools = append(tools, s.buildLoggingTool())
	}
	return tools
}

// Stop gracefully shuts down the MCP server and releases all resources.
// Safe for repeated calls and concurrent access.
func (s *MCPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sseServer != nil {
		if err := s.sseServer.Shutdown(ctx); err != nil {
			s.logger.Warnf("Error stopping SSE server: %v", err)
		}
		s.sseServer = nil
	}
	s.server = nil
	return nil
}

// BuildHooks creates a set of hooks for logging MCP events.
// Used for debugging and extending server behavior.
func (s *MCPServer) BuildHooks() *server.Hooks {
	hooks := &server.Hooks{}

	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		s.logger.Infof("[MCP] Before call %s: %+v", message.Params.Name, message)
	})

	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		s.logger.Infof("[MCP] After call %s result: %+v", message.Params.Name, result)
	})

	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		s.logger.Errorf("[MCP] Error with method %s: %v | message: %+v", method, err, message)
	})

	return hooks
}

// GetAllTools returns all tools registered on the server.
// Used for testing and integration.
func (s *MCPServer) GetAllTools() []mcp.Tool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buildTools()
}

// SendNotificationToClient sends a notification to a single client via MCP.
// Used for logger integration and tests.
func (s *MCPServer) SendNotificationToClient(ctx context.Context, method string, data map[string]interface{}) error {
	if s.server == nil {
		return fmt.Errorf("MCPServer: underlying server is not initialized")
	}
	err := s.server.SendNotificationToClient(ctx, method, data)
	if err != nil {
		return fmt.Errorf("MCPServer: failed to send notification to client: %w", err)
	}
	return nil
}

// GetServerCapabilities returns ServerCapabilities for tests and integration.
func (s *MCPServer) GetServerCapabilities() mcp.ServerCapabilities {
	caps := mcp.ServerCapabilities{}
	if s.server != nil {
		if s.config.MCPLogEnabled {
			caps.Logging = &struct{}{}
		}
	}
	return caps
}

// GetServer returns the internal *server.MCPServer instance (for tests and integration).
// Returns nil if the server is not initialized.
func (s *MCPServer) GetServer() *server.MCPServer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.server
}

// SetMainToolHandler allows setting the handler for the main tool (process)
func (s *MCPServer) SetMainToolHandler(handler server.ToolHandlerFunc) {
	fmt.Fprintf(os.Stderr, "[MCPServer] SetMainToolHandler: setting handler\n")
	s.mu.Lock()
	s.mainToolHandler = handler
	s.mu.Unlock()
	fmt.Fprintf(os.Stderr, "[MCPServer] SetMainToolHandler: handler set\n")
}
