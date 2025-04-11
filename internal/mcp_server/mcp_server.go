// Package mcp_server provides functionality for the MCP server.
// Responsibility: Implementation of the MCP server for processing client requests
// Features: Supports two operating modes - HTTP (daemon) and stdio
package mcp_server

import (
    "context"
    "fmt"

    "github.com/korchasa/speelka-agent-go/internal/logger"
    "github.com/korchasa/speelka-agent-go/internal/types"
    "github.com/korchasa/speelka-agent-go/internal/utils"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

// MCPServer implements the contracts.MCPServerSpec interface
// Responsibility: Managing the lifecycle of the MCP server and processing requests
// Features: Stores server state and provides access to the tool registry
type MCPServer struct {
    server    *server.MCPServer
    config    types.MCPServerConfig
    logger    logger.Spec
    sseServer *server.SSEServer
}

// NewMCPServer creates a new MCPServer instance
// Responsibility: Factory method for creating an MCP server
// Features: Initializes the data structure with the given parameters
func NewMCPServer(config types.MCPServerConfig, logger logger.Spec) *MCPServer {
    return &MCPServer{
        config: config,
        logger: logger,
    }
}

// ServeDaemon initializes and starts the HTTP MCP server
// Responsibility: Starting the server in daemon mode with HTTP interface
// Features: Sets the launch flag and logs configuration information
func (s *MCPServer) ServeDaemon(handler server.ToolHandlerFunc) error {
    var err error
    if err = s.createAndInitMCPServer(handler); err != nil {
        return fmt.Errorf("failed to create and initialize MCP server: %w", err)
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

// ServeStdio initializes and starts the stdio MCP server
// Responsibility: Starting the server in input-output mode through standard streams
// Features: Sets the launch flag and prepares stdin/stdout handling
func (s *MCPServer) ServeStdio(handler server.ToolHandlerFunc) error {
    var err error
    if err = s.createAndInitMCPServer(handler); err != nil {
        return fmt.Errorf("failed to create and initialize MCP server: %w", err)
    }
    s.logger.Info("MCP Stdio server initialized successfully")

    if err := server.ServeStdio(s.server); err != nil {
        return fmt.Errorf("failed to serve stdio MCP server: %w", err)
    }
    return nil
}

func (s *MCPServer) createAndInitMCPServer(handler server.ToolHandlerFunc) error {
    var opts []server.ServerOption
    opts = append(opts, server.WithLogging())
    if s.config.Debug {
        opts = append(opts, server.WithHooks(s.BuildHooks()))
    }

    s.server = server.NewMCPServer(
        s.config.Name,
        s.config.Version,
        opts...,
    )

    s.logger.Infof("MCP server initialized with config: %s", utils.SDump(s.config))

    tool := mcp.NewTool(s.config.Tool.Name,
        mcp.WithDescription(s.config.Tool.Description),
        mcp.WithString(s.config.Tool.ArgumentName,
            mcp.Required(),
            mcp.Description(s.config.Tool.ArgumentDescription),
        ),
    )

    s.server.AddTool(tool, handler)

    return nil
}

// Stop gracefully terminates the MCP server
// Responsibility: Stopping the server and releasing resources
// Features: Resets the launch flag and performs necessary cleanup
func (s *MCPServer) Stop(ctx context.Context) error {
    if s.sseServer != nil {
        err := s.sseServer.Shutdown(ctx)
        if err != nil {
            return fmt.Errorf("failed to shutdown SSE server: %w", err)
        }
    }
    return nil
}

func (s *MCPServer) BuildHooks() *server.Hooks {
    hooks := &server.Hooks{}
    hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
        s.logger.WithField("id", id).Infof("MCP server hook onSuccess")
        s.logger.Debugf("Details: %s", utils.SDump(map[string]any{
            "message": message,
            "result":  result,
        }))
    })
    hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
        s.logger.WithField("id", id).Infof("MCP server hook onError")
        s.logger.Debugf("Details: %s", utils.SDump(map[string]any{
            "message": message,
            "error":   err,
        }))
    })
    hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
        s.logger.WithField("id", id).Infof("MCP server hook beforeInitialize for tool `%s`", message.Method)
        s.logger.Debugf("Details: %s", utils.SDump(map[string]any{
            "message": message,
        }))
    })
    hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
        s.logger.WithField("id", id).Infof("MCP server hook afterInitialize for tool `%s`", message.Method)
        s.logger.Debugf("Details: %s", utils.SDump(map[string]any{
            "message": message,
            "result":  result,
        }))
    })
    hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
        s.logger.WithField("id", id).Infof("MCP server hook beforeCallTool for tool `%s`", message.Method)
        s.logger.Debugf("Details: %s", utils.SDump(map[string]any{
            "message": message,
        }))
    })
    hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
        s.logger.WithField("id", id).Infof("MCP server hook afterCallTool for tool `%s`", message.Method)
        s.logger.Debugf("Details: %s", utils.SDump(map[string]any{
            "message": message,
            "result":  result,
        }))
    })
    return hooks
}

func (s *MCPServer) AttachLogger(spec logger.Spec) {
    spec.SetMCPServer(s.server)
}
