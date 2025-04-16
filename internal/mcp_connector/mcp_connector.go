// Package mcp_connector provides functionality for connecting to MCP servers.
// Responsibility: Ensuring interaction with external MCP servers
// Features: Supports various transport protocols (HTTP, stdio) and manages connections to multiple servers
package mcp_connector

import (
	"bufio"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/utils"

	"github.com/korchasa/speelka-agent-go/internal/error_handling"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPConnector implements the contracts.MCPConnectorSpec interface
// Responsibility: Managing connections to external MCP servers
// Features: Provides access to tools from all connected servers
type MCPConnector struct {
	config          types.MCPConnectorConfig
	clients         map[string]client.MCPClient
	tools           map[string][]mcp.Tool
	dataLock        sync.RWMutex
	logger          types.LoggerSpec
	toolCallTimeout time.Duration
}

// NewMCPConnector creates a new instance of MCPConnector
// Responsibility: Factory method for creating an MCP connector
// Features: Returns a simple instance without initialization
func NewMCPConnector(config types.MCPConnectorConfig, logger types.LoggerSpec) *MCPConnector {
	return &MCPConnector{
		clients:         make(map[string]client.MCPClient),
		tools:           make(map[string][]mcp.Tool),
		config:          config,
		logger:          logger,
		toolCallTimeout: 30 * time.Second, // Default timeout
	}
}

// InitAndConnectToMCPs connects to all configured MCP servers.
// Responsibility: Establishing connections with all servers specified in the configuration
// Features: Gets and registers tools from each server
func (mc *MCPConnector) InitAndConnectToMCPs(ctx context.Context) error {
	mc.dataLock.Lock()
	defer mc.dataLock.Unlock()
	// Connecting to all configured MCP servers
	for serverID, srvCfg := range mc.config.McpServers {
		mc.logger.Infof("Connecting to MCP server `%s`", serverID)
		mc.logger.Debugf("Details: %s", utils.SDump(srvCfg))
		mcpClient, err := mc.ConnectServer(ctx, serverID, srvCfg)
		if err != nil {
			return error_handling.WrapError(
				err,
				fmt.Sprintf("failed to connect to MCP server %s", serverID),
				error_handling.ErrorCategoryExternal,
			)
		}
		mc.logger.Infof("Connected to MCP server `%s`", serverID)

		toolsResp, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			return error_handling.WrapError(
				err,
				fmt.Sprintf("failed to list tools from MCP server %s", serverID),
				error_handling.ErrorCategoryExternal,
			)
		}
		filteredTools := make([]mcp.Tool, 0)
		for _, tool := range toolsResp.Tools {
			if srvCfg.IsToolAllowed(tool.Name) {
				mc.logger.Infof("`%s:%s` tool added", serverID, tool.Name)
				mc.logger.Debugf("Details: %s", utils.SDump(tool))
				filteredTools = append(filteredTools, tool)
			} else {
				mc.logger.Infof("`%s:%s` tool not allowed", serverID, tool.Name)
			}
		}
		mc.clients[serverID] = mcpClient
		mc.tools[serverID] = filteredTools
		mc.logger.Infof("Connected to MCP server `%s` with %d tools", serverID, len(filteredTools))
	}
	mc.logger.Infof("Connected to %d MCP servers", len(mc.clients))
	return nil
}

// ConnectServer connects to an MCP server using HTTP or stdio transport.
// Responsibility: Establishing a connection with a specific MCP server
// Features: Selects the appropriate transport based on configuration and uses a retry strategy
func (mc *MCPConnector) ConnectServer(ctx context.Context, serverID string, serverConfig types.MCPServerConnection) (client.MCPClient, error) {
	// Define a function that attempts to connect
	var mcpClient client.MCPClient
	connectFn := func() error {
		var err error

		// Determine transport type based on available fields
		if serverConfig.Command != "" {
			// Use stdio client for command-based servers
			mcpClient, err = client.NewStdioMCPClient(
				serverConfig.Command,
				serverConfig.Environment,
				serverConfig.Args...,
			)
			if err != nil {
				return error_handling.WrapError(
					err,
					"failed to create stdio MCP client",
					error_handling.ErrorCategoryExternal,
				)
			}

			// Capture stderr output and log it with warning level
			if stdioClient, ok := mcpClient.(*client.StdioMCPClient); ok {
				stderrReader := stdioClient.Stderr()
				go func() {
					reader := bufio.NewReader(stderrReader)
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							return
						}
						mc.logger.Infof("`%s` stderr: %s", serverID, line)
					}
				}()
			}
		} else if serverConfig.URL != "" {
			// Use HTTP client with SSE
			// Set up headers
			headers := make(map[string]string)
			if serverConfig.APIKey != "" {
				headers["Authorization"] = "Bearer " + serverConfig.APIKey
			}

			// Create HTTP client
			mcpClient, err = client.NewSSEMCPClient(
				serverConfig.URL,
				client.WithHeaders(headers),
			)
			if err != nil {
				return error_handling.WrapError(
					err,
					"failed to create HTTP MCP client",
					error_handling.ErrorCategoryExternal,
				)
			}
		} else {
			return error_handling.NewError(
				"neither command nor URL is specified for MCP server connection",
				error_handling.ErrorCategoryValidation,
			)
		}

		// Initialize the client with timeout
		initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "speelka-agent",
			Version: "1.0.0",
		}

		_, err = mcpClient.Initialize(initCtx, initRequest)
		if err != nil {
			return error_handling.WrapError(
				err,
				"failed to initialize MCP client",
				error_handling.ErrorCategoryInternal,
			)
		}

		return nil
	}

	// Use retry with backoff for transient errors
	return mcpClient, error_handling.RetryWithBackoff(ctx, connectFn, error_handling.RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxBackoff:        5 * time.Second,
	})
}

func (mc *MCPConnector) GetAllTools(ctx context.Context) ([]mcp.Tool, error) {
	mc.dataLock.RLock()
	defer mc.dataLock.RUnlock()

	allTools := make([]mcp.Tool, 0)
	for _, tools := range mc.tools {
		allTools = append(allTools, tools...)
	}
	return allTools, nil
}

// ExecuteTool executes a tool on an MCP server.
func (mc *MCPConnector) ExecuteTool(ctx context.Context, call types.CallToolRequest) (*mcp.CallToolResult, error) {
	mc.dataLock.RLock()
	defer mc.dataLock.RUnlock()

	foundServerID := ""
	for serverID, serverTools := range mc.tools {
		for _, tool := range serverTools {
			if tool.Name == call.Params.Name {
				foundServerID = serverID
				break
			}
		}
	}

	if foundServerID == "" {
		return nil, error_handling.NewError(
			fmt.Sprintf("tool `%s` not found", call.Params.Name),
			error_handling.ErrorCategoryValidation,
		)
	}

	mcpClient, exists := mc.clients[foundServerID]
	if !exists {
		return nil, error_handling.NewError(
			fmt.Sprintf("not connected to server: %s", foundServerID),
			error_handling.ErrorCategoryValidation,
		)
	}

	// ToolCall the tool with timeout
	callCtx, cancel := context.WithTimeout(ctx, mc.toolCallTimeout)
	defer cancel()

	result, err := mcpClient.CallTool(callCtx, call.CallToolRequest)
	if err != nil {
		return nil, error_handling.WrapError(
			err,
			fmt.Sprintf("failed to call tool `%s`", call.Params.Name),
			error_handling.ErrorCategoryInternal,
		)
	}
	// Process and return the result
	return result, nil
}

// Close closes all client connections.
func (mc *MCPConnector) Close() error {
	mc.dataLock.Lock()
	defer mc.dataLock.Unlock()

	for id, cl := range mc.clients {
		if err := cl.Close(); err != nil {
			mc.logger.WithFields(map[string]interface{}{
				"server_id": id,
				"error":     err.Error(),
			}).Error("Failed to close MCP client")
		}
	}

	return nil
}
