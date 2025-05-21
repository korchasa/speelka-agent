// Package mcp_server provides functionality for the MCP server.
// Responsibility: Implementation of the MCP server for processing client requests
// Features: Supports two operating modes - HTTP (daemon) and stdio
package mcp_server

import (
    "context"
    "fmt"
    "os"
    "sync"

    "github.com/korchasa/speelka-agent-go/internal/utils/log_levels"

    "github.com/korchasa/speelka-agent-go/internal/configuration"
    "github.com/sirupsen/logrus"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

const (
    setLevelToolName = "logging/setLevel"
)

// MCPServer implements an MCP server for handling client requests and managing the lifecycle of tools.
// Thread-safe for public methods. All external dependencies are injected via the constructor (DI).
type MCPServer struct {
    server     *server.MCPServer             // Internal MCP server
    cfg        configuration.MCPServerConfig // Server configuration
    log        *logrus.Logger                // Logger (DI)
    sseServer  *server.SSEServer             // HTTP SSE server (optional)
    isHttpMode bool                          // true if HTTP is enabled, false if Stdio is enabled)
    mu         sync.Mutex                    // Protects the state of server/sseServer
}

// NewMCPServer creates a new instance of MCPServer with the given configuration and logger.
// All dependencies are injected via parameters (Dependency Injection).
func NewMCPServer(cfg configuration.MCPServerConfig, log *logrus.Logger) (*MCPServer, error) {
    var err error
    var opts []server.ServerOption
    if cfg.MCPLogEnabled {
        opts = append(opts, server.WithLogging())
    }
    if cfg.Debug {
        opts = append(opts, server.WithHooks((&MCPServer{cfg: cfg, log: log}).BuildHooks()))
    }

    mcpSrv := server.NewMCPServer(
        cfg.Name,
        cfg.Version,
        opts...,
    )

    mcps := &MCPServer{
        server: mcpSrv,
        cfg:    cfg,
        log:    log,
    }

    log.Infof("MCPServer: server created with config: %+v", cfg)
    mcps.isHttpMode, err = getIsHttpMode(cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to validate config: %w", err)
    }

    return mcps, nil
}

// Serve starts the MCP server in daemon (HTTP SSE) or script (stdio) mode.
// Thread-safe. Releases resources before completion.
func (s *MCPServer) Serve(ctx context.Context, handler server.ToolHandlerFunc) error {
    // Register tools immediately
    for _, tool := range s.buildTools() {
        var h server.ToolHandlerFunc = nil
        if tool.Name == s.cfg.Tool.Name {
            h = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
                s.log.AddHook(&LogHook{server: s, ctx: ctx})
                if handler == nil {
                    return nil, fmt.Errorf("main tool handler is not set for '%s'", tool.Name)
                }
                res, err := handler(ctx, req)
                return res, err
            }
        } else if tool.Name == setLevelToolName {
            h = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
                s.log.AddHook(&LogHook{server: s, ctx: ctx})
                mcpLevel, ok := req.Params.Arguments["level"].(string)
                if !ok {
                    return nil, fmt.Errorf("can't find level argument in %s request", setLevelToolName)
                }
                logrusLevel, err := log_levels.MCPToLogrusLevel(mcpLevel)
                if err != nil {
                    return nil, fmt.Errorf("failed to convert MCP log level '%s' to logrus level: %w", mcpLevel, err)
                }
                s.log.SetLevel(logrusLevel)
                result := &mcp.CallToolResult{}
                return result, nil
            }
        }
        s.server.AddTool(tool, h)
    }

    if s.isHttpMode {
        s.log.Info("Running in daemon mode with HTTP SSE MCP server")
        if err := s.initSSEServer(handler); err != nil {
            return fmt.Errorf("failed to start HTTP MCP server: %w", err)
        }
    } else {
        s.log.Info("Running in script mode with stdio MCP server")
        if err := s.initStdioServer(ctx, handler); err != nil {
            return fmt.Errorf("failed to start Stdio MCP Server: %w", err)
        }
    }
    s.log.Infof("MSP Server: finished")
    return nil
}

// initSSEServer initializes and starts the HTTP SSE MCP server.
func (s *MCPServer) initSSEServer(handler server.ToolHandlerFunc) error {
    if s.server == nil {
        return fmt.Errorf("server is not *server.MCPServer")
    }
    s.log.Info("MCP SSE server initialized successfully")
    addr := fmt.Sprintf("%s:%d", s.cfg.HTTP.Host, s.cfg.HTTP.Port)
    baseUrl := fmt.Sprintf("http://%s:%d", s.cfg.HTTP.Host, s.cfg.HTTP.Port)
    s.sseServer = server.NewSSEServer(s.server, server.WithBaseURL(baseUrl))
    if err := s.sseServer.Start(addr); err != nil {
        return fmt.Errorf("failed to serve SSE MCP server: %w", err)
    }
    return nil
}

// initStdioServer initializes and starts the stdio MCP server with external context support.
func (s *MCPServer) initStdioServer(ctx context.Context, handler server.ToolHandlerFunc) error {
    if s.server == nil {
        return fmt.Errorf("server is not *server.MCPServer")
    }
    s.log.Info("MCP Stdio server initialized successfully")
    return server.NewStdioServer(s.server).Listen(ctx, os.Stdin, os.Stdout)
}

// buildMainTool creates the main tool for the server.
func (s *MCPServer) buildMainTool() mcp.Tool {
    return mcp.NewTool(s.cfg.Tool.Name,
        mcp.WithDescription(s.cfg.Tool.Description),
        mcp.WithString(s.cfg.Tool.ArgumentName,
            mcp.Description(s.cfg.Tool.ArgumentDescription),
            mcp.Required(),
        ),
    )
}

// buildLoggingTool creates a tool for managing logging.
func (s *MCPServer) buildLoggingTool() mcp.Tool {
    return mcp.NewTool(setLevelToolName,
        mcp.WithString("level", mcp.Required(), mcp.Description("Log level to set")),
    )
}

// Stop gracefully shuts down the MCP server and releases all resources.
// Safe for repeated calls and concurrent access.
func (s *MCPServer) Stop(ctx context.Context) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.sseServer != nil {
        if err := s.sseServer.Shutdown(ctx); err != nil {
            s.log.Warnf("Error stopping SSE server: %v", err)
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
        s.log.Infof("[MCP] Before call %s: %+v", message.Params.Name, message)
    })

    hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
        s.log.Infof("[MCP] After call %s result: %+v", message.Params.Name, result)
    })

    hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
        s.log.Errorf("[MCP] Error with method %s: %v | message: %+v", method, err, message)
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
        if s.cfg.MCPLogEnabled {
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

// buildTools returns a list of all tools to register on the server.
func (s *MCPServer) buildTools() []mcp.Tool {
    tools := []mcp.Tool{s.buildMainTool()}
    if s.cfg.MCPLogEnabled {
        tools = append(tools, s.buildLoggingTool())
    }
    return tools
}

func getIsHttpMode(cfg configuration.MCPServerConfig) (bool, error) {

    isHttpEnabled := cfg.HTTP.Enabled
    isStdioEnabled := cfg.Stdio.Enabled
    if isHttpEnabled && isStdioEnabled {
        return false, fmt.Errorf("both HTTP and Stdio modes cannot be enabled at the same time")
    }
    if !isHttpEnabled && !isStdioEnabled {
        return false, fmt.Errorf("either HTTP or Stdio mode must be enabled")
    }
    return isHttpEnabled, nil
}
