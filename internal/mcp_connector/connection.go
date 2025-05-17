// Package mcp_connector: connection and initialization logic for MCP clients
package mcp_connector

import (
	"context"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/error_handling"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

// ConnectServer connects to an MCP server using HTTP or stdio transport.
// Delegates to transport-specific methods.
func (mc *MCPConnector) ConnectServer(ctx context.Context, serverID string, serverConfig types.MCPServerConnection) (client.MCPClient, error) {
	startTime := time.Now()

	var mcpClient client.MCPClient
	var err error

	switch {
	case serverConfig.Command != "":
		mcpClient, err = mc.connectStdioServer(ctx, serverID, serverConfig)
		if err != nil {
			return nil, error_handling.WrapError(
				err,
				"failed to connect to MCP server by Stdio",
				error_handling.ErrorCategoryExternal,
			)
		}
	case serverConfig.URL != "":
		mcpClient, err = mc.connectHTTPServer(ctx, serverID, serverConfig)
		if err != nil {
			return nil, error_handling.WrapError(
				err,
				"failed to connect to MCP server by HTTP",
				error_handling.ErrorCategoryExternal,
			)
		}
	default:
		return nil, error_handling.NewError(
			"neither command nor URL is specified for MCP server connection",
			error_handling.ErrorCategoryValidation,
		)
	}

	endTime := time.Now()
	mc.logger.Infof("[MCP-CONNECT] ConnectServer: finished for server '%s' at %s (duration: %s)", serverID, endTime.Format(time.RFC3339Nano), endTime.Sub(startTime))
	return mcpClient, nil
}

// connectStdioServer creates and initializes a stdio MCP client, saves capabilities, and sets up logging.
func (mc *MCPConnector) connectStdioServer(ctx context.Context, serverID string, serverConfig types.MCPServerConnection) (client.MCPClient, error) {
	mc.logger.Debugf("[MCP-CONNECT] connectStdioServer: serverID='%s', command='%s', args=%v, env=%v", serverID, serverConfig.Command, serverConfig.Args, serverConfig.Environment)
	mcpClient, err := client.NewStdioMCPClient(
		serverConfig.Command,
		serverConfig.Environment,
		serverConfig.Args...,
	)
	if err != nil {
		mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectStdioServer: failed to create client for server '%s': %v", serverID, err)
		return nil, error_handling.WrapError(
			err,
			"failed to create stdio MCP client",
			error_handling.ErrorCategoryExternal,
		)
	}
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "speelka-agent",
		Version: "1.0.0",
	}
	initStart := time.Now()
	mc.logger.Debugf("[MCP-CONNECT] connectStdioServer: initializing client for server '%s' at %s", serverID, initStart.Format(time.RFC3339Nano))
	initResult, err := mcpClient.Initialize(initCtx, initRequest)
	initDuration := time.Since(initStart)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectStdioServer: initialization for server '%s' timed out after %s", serverID, initDuration)
		} else {
			mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectStdioServer: initialization for server '%s' failed after %s: %v", serverID, initDuration, err)
		}
		return nil, error_handling.WrapError(
			err,
			"failed to initialize MCP client",
			error_handling.ErrorCategoryInternal,
		)
	}
	mc.logger.Debugf("[MCP-CONNECT] connectStdioServer: initialization for server '%s' completed in %s", serverID, initDuration)
	mc.logger.Debugf("[MCP-CONNECT] connectStdioServer: acquiring dataLock for server '%s' at %s", serverID, time.Now().Format(time.RFC3339Nano))
	mc.dataLock.Lock()
	mc.capabilities[serverID] = initResult.Capabilities
	mc.dataLock.Unlock()
	mc.logger.Debugf("[MCP-CONNECT] connectStdioServer: capabilities for server '%s' set to %v", serverID, initResult.Capabilities)
	mc.setupLoggingRoute(serverID, mcpClient, initResult.Capabilities, serverConfig)
	return mcpClient, nil
}

// connectHTTPServer creates and initializes an HTTP/SSE MCP client, saves capabilities, and sets up logging.
func (mc *MCPConnector) connectHTTPServer(ctx context.Context, serverID string, serverConfig types.MCPServerConnection) (client.MCPClient, error) {
	headers := make(map[string]string)
	if serverConfig.APIKey != "" {
		headers["Authorization"] = "Bearer " + serverConfig.APIKey
	}
	mc.logger.Debugf("[MCP-CONNECT] connectHTTPServer: serverID='%s', url='%s', headers=%v", serverID, serverConfig.URL, headers)
	mcpClient, err := client.NewSSEMCPClient(
		serverConfig.URL,
		client.WithHeaders(headers),
	)
	if err != nil {
		mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectHTTPServer: failed to create client for server '%s': %v", serverID, err)
		return nil, error_handling.WrapError(
			err,
			"failed to create HTTP MCP client",
			error_handling.ErrorCategoryExternal,
		)
	}
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "speelka-agent",
		Version: "1.0.0",
	}
	initStart := time.Now()
	mc.logger.Debugf("[MCP-CONNECT] connectHTTPServer: initializing client for server '%s' at %s", serverID, initStart.Format(time.RFC3339Nano))
	initResult, err := mcpClient.Initialize(initCtx, initRequest)
	initDuration := time.Since(initStart)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectHTTPServer: initialization for server '%s' timed out after %s", serverID, initDuration)
		} else {
			mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectHTTPServer: initialization for server '%s' failed after %s: %v", serverID, initDuration, err)
		}
		return nil, error_handling.WrapError(
			err,
			"failed to initialize MCP client",
			error_handling.ErrorCategoryInternal,
		)
	}
	mc.logger.Debugf("[MCP-CONNECT] connectHTTPServer: initialization for server '%s' completed in %s", serverID, initDuration)
	mc.logger.Debugf("[MCP-CONNECT] connectHTTPServer: acquiring dataLock for server '%s' at %s", serverID, time.Now().Format(time.RFC3339Nano))
	mc.dataLock.Lock()
	mc.capabilities[serverID] = initResult.Capabilities
	mc.dataLock.Unlock()
	mc.setupLoggingRoute(serverID, mcpClient, initResult.Capabilities, serverConfig)
	return mcpClient, nil
}
