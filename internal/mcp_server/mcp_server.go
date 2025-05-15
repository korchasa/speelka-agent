// Package mcp_server provides functionality for the MCP server.
// Responsibility: Implementation of the MCP server for processing client requests
// Features: Supports two operating modes - HTTP (daemon) and stdio
package mcp_server

import (
	"context"
	"fmt"
	"os"

	"github.com/korchasa/speelka-agent-go/internal/utils"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer implements the contracts.MCPServerSpec interface
// Responsibility: Managing the lifecycle of the MCP server and processing requests
// Features: Stores server state and provides access to the tool registry
type MCPServer struct {
	server    *server.MCPServer
	config    types.MCPServerConfig
	logger    types.LoggerSpec
	sseServer *server.SSEServer
}

// NewMCPServer creates a new MCPServer instance
// Responsibility: Factory method for creating an MCP server
// Features: Initializes the data structure with the given parameters
func NewMCPServer(config types.MCPServerConfig, logger types.LoggerSpec) *MCPServer {
	return &MCPServer{
		config: config,
		logger: logger,
	}
}

func (s *MCPServer) Serve(ctx context.Context, daemonMode bool, handler server.ToolHandlerFunc) error {
	if daemonMode {
		s.logger.Info("Running in daemon mode with HTTP SSE MCP server")
		if err := s.serveDaemon(handler); err != nil {
			return fmt.Errorf("failed to start HTTP MCP server: %w", err)
		}
	} else {
		s.logger.Info("Running in script mode with stdio MCP server")
		if err := s.serveStdioWithContext(handler, ctx); err != nil {
			return fmt.Errorf("failed to start Stdio MCP Server: %w", err)
		}
	}
	return nil
}

// serveDaemon initializes and starts the HTTP MCP server
// Responsibility: Starting the server in daemon mode with HTTP interface
// Features: Sets the launch flag and logs configuration information
func (s *MCPServer) serveDaemon(handler server.ToolHandlerFunc) error {
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

// serveStdioWithContext initializes and starts the stdio MCP server with external context support
// ServeStdioWithContext starts the stdio MCP server with external context support and without internal signal handling
// SendNotificationToClient implements types.MCPServerNotifier for logger integration
// GetServerCapabilities returns ServerCapabilities for tests and integration
// Check if logging capability is enabled via internal field
// (In mark3labs/mcp-go library, logging capability == true => Logging != nil)
// No public API, so use InitializeResult if extension is needed
// Here, do a simple check via reinitialization of the structure
// (or add an exported method in a forked library)
// For the test: if WithLogging() was called, capability is present
// Check via reinitialization
// Easier to check via handleInitialize, but hard to mock
// Therefore, use knowledge: if LogRawOutput == ":mcp:", capability is present
// Now all nested structures are inline (anonymous)
func (s *MCPServer) serveStdioWithContext(handler server.ToolHandlerFunc, ctx context.Context) error {
	var err error
	if err = s.createAndInitMCPServer(handler); err != nil {
		return fmt.Errorf("failed to create and initialize MCP server: %w", err)
	}
	s.logger.Info("MCP Stdio server initialized successfully")
	return ServeStdioWithContext(s.server, s.logger, ctx)
}

// ServeStdioWithContext starts the stdio MCP server with external context support and without internal signal handling
// Check if logging capability is enabled via internal field
// (In mark3labs/mcp-go library, logging capability == true => Logging != nil)
// No public API, so use InitializeResult if extension is needed
// Here, do a simple check via reinitialization of the structure
// (or add an exported method in a forked library)
// For the test: if WithLogging() was called, capability is present
// Check via reinitialization
// Easier to check via handleInitialize, but hard to mock
// Therefore, use knowledge: if LogRawOutput == ":mcp:", capability is present
func ServeStdioWithContext(mcpSrv *server.MCPServer, logger types.LoggerSpec, ctx context.Context) error {
	return server.NewStdioServer(mcpSrv).Listen(ctx, os.Stdin, os.Stdout)
}

func (s *MCPServer) createAndInitMCPServer(handler server.ToolHandlerFunc) error {
	var opts []server.ServerOption
	if s.config.LogRawOutput == types.LogOutputMCP {
		opts = append(opts, server.WithLogging())
	}
	if s.config.Debug {
		opts = append(opts, server.WithHooks(s.BuildHooks()))
	}

	s.server = server.NewMCPServer(
		s.config.Name,
		s.config.Version,
		opts...,
	)

	s.logger.Debugf("MCP server initialized with config: %s", utils.SDump(s.config))

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
		if err := s.sseServer.Shutdown(ctx); err != nil {
			s.logger.Warnf("Error stopping SSE server: %v", err)
		}
		s.sseServer = nil
	}
	s.server = nil
	return nil
}

// BuildHooks creates hook functions for the MCP server
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

// AttachLogger attaches a logger to the MCP server
func (s *MCPServer) AttachLogger(logger types.LoggerSpec) {
	logger.SetMCPServer(s)
}

// GetServer returns the underlying server instance
func (s *MCPServer) GetServer() *server.MCPServer {
	return s.server
}

// AddTool adds a tool to the MCP server
// Responsibility: Adding a tool to the server
// Features: Delegates to the underlying server's AddTool method
func (s *MCPServer) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	if s.server != nil {
		s.server.AddTool(tool, handler)
	} else {
		s.logger.Warn("Cannot add tool: server not initialized")
	}
}

// GetAllTools returns all tools registered on the server
// Responsibility: Providing access to all available tools
// Features: Collects and returns all tools from the server
func (s *MCPServer) GetAllTools() []mcp.Tool {
	if s.server == nil {
		s.logger.Warn("Cannot get tools: server not initialized")
		return []mcp.Tool{}
	}

	// Since we can't directly access the tools in the server,
	// we'll need to implement this differently or just return a partial list.
	// For now, return just the tool we know exists
	return []mcp.Tool{
		mcp.NewTool(s.config.Tool.Name,
			mcp.WithDescription(s.config.Tool.Description),
			mcp.WithString(s.config.Tool.ArgumentName,
				mcp.Description(s.config.Tool.ArgumentDescription),
				mcp.Required(),
			),
		),
		ExitTool,
	}
}

// ExitTool is used to signal that the conversation should end
var ExitTool = mcp.NewTool("answer",
	mcp.WithDescription("Send response to the user"),
	mcp.WithString("text",
		mcp.Required(),
		mcp.Description("Text to send to the user"),
	),
)

// SendNotificationToClient implements types.MCPServerNotifier for logger integration
func (s *MCPServer) SendNotificationToClient(ctx context.Context, method string, data map[string]interface{}) error {
	if s.server == nil {
		return fmt.Errorf("MCPServer: underlying server is not initialized")
	}
	return s.server.SendNotificationToClient(ctx, method, data)
}

// GetServerCapabilities returns ServerCapabilities for tests and integration
func (s *MCPServer) GetServerCapabilities() mcp.ServerCapabilities {
	caps := mcp.ServerCapabilities{}
	if s.server != nil {
		// Check if logging capability is enabled via internal field
		// (In mark3labs/mcp-go library, logging capability == true => Logging != nil)
		// No public API, so use InitializeResult if extension is needed
		// Here, do a simple check via reinitialization of the structure
		// (or add an exported method in a forked library)
		// For the test: if WithLogging() was called, capability is present
		// Check via reinitialization
		// Easier to check via handleInitialize, but hard to mock
		// Therefore, use knowledge: if LogRawOutput == ":mcp:", capability is present
		if s.config.LogRawOutput == ":mcp:" {
			caps.Logging = &struct{}{}
		}
	}
	return caps
}
