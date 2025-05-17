// Package mcp_connector provides functionality for connecting to MCP servers.
// Responsibility: Ensuring interaction with external MCP servers
// Features: Supports various transport protocols (HTTP, stdio) and manages connections to multiple servers
package mcp_connector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/utils"

	"github.com/korchasa/speelka-agent-go/internal/error_handling"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

// MCPConnector implements the contracts.ToolConnectorSpec interface
// Responsibility: Managing connections to external MCP servers
// Features: Provides access to tools from all connected servers
type MCPConnector struct {
	config       types.MCPConnectorConfig
	clients      map[string]client.MCPClient
	tools        map[string][]mcp.Tool
	capabilities map[string]mcp.ServerCapabilities // capabilities per server
	dataLock     sync.RWMutex
	logger       types.LoggerSpec
}

// NewMCPConnector creates a new instance of MCPConnector
// Responsibility: Factory method for creating an MCP connector
// Features: Returns a simple instance without initialization
func NewMCPConnector(config types.MCPConnectorConfig, logger types.LoggerSpec) *MCPConnector {
	return &MCPConnector{
		clients:      make(map[string]client.MCPClient),
		tools:        make(map[string][]mcp.Tool),
		capabilities: make(map[string]mcp.ServerCapabilities),
		config:       config,
		logger:       logger,
	}
}

// InitAndConnectToMCPs connects to all configured MCP servers.
// Responsibility: Establishing connections with all servers specified in the configuration
// Features: Gets and registers tools from each server
func (mc *MCPConnector) InitAndConnectToMCPs(ctx context.Context) error {
	for serverID, srvCfg := range mc.config.McpServers {
		mc.logger.Debugf("[MCP-CONNECT] About to connectAndRegisterServer: %s at %s", serverID, time.Now().Format(time.RFC3339Nano))
		if err := mc.connectAndRegisterServer(ctx, serverID, srvCfg); err != nil {
			mc.logger.Errorf("[MCP-CONNECT] [ERROR] connectAndRegisterServer failed for %s: %v", serverID, err)
			return err
		}
		mc.logger.Debugf("[MCP-CONNECT] Finished connectAndRegisterServer: %s at %s", serverID, time.Now().Format(time.RFC3339Nano))
	}
	mc.logger.Infof("Connected to %d MCP servers", len(mc.clients))
	return nil
}

// connectAndRegisterServer handles connection and tool registration for a single server.
func (mc *MCPConnector) connectAndRegisterServer(ctx context.Context, serverID string, srvCfg types.MCPServerConnection) error {
	mc.logger.Infof("[MCP-CONNECT] Server config: %s", utils.SDump(srvCfg))
	mcpClient, err := mc.ConnectServer(ctx, serverID, srvCfg)
	if err != nil {
		return error_handling.WrapError(
			err,
			fmt.Sprintf("failed to connect to MCP server %s", serverID),
			error_handling.ErrorCategoryExternal,
		)
	}

	toolsResp, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return error_handling.WrapError(
			err,
			fmt.Sprintf("failed to list tools from MCP server %s", serverID),
			error_handling.ErrorCategoryExternal,
		)
	}
	filteredTools := mc.filterAllowedTools(serverID, toolsResp.Tools, srvCfg)
	mc.clients[serverID] = mcpClient
	mc.tools[serverID] = filteredTools
	mc.logger.Infof("Connected to MCP server `%s` with %d tools", serverID, len(filteredTools))
	return nil
}

// filterAllowedTools filters tools based on server config.
func (mc *MCPConnector) filterAllowedTools(serverID string, tools []mcp.Tool, srvCfg types.MCPServerConnection) []mcp.Tool {
	filtered := make([]mcp.Tool, 0)
	for _, tool := range tools {
		if srvCfg.IsToolAllowed(tool.Name) {
			mc.logger.Infof("`%s:%s` tool added", serverID, tool.Name)
			mc.logger.Debugf("Details: %s", utils.SDump(tool))
			filtered = append(filtered, tool)
		} else {
			mc.logger.Infof("`%s:%s` tool not allowed", serverID, tool.Name)
		}
	}
	return filtered
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
	mc.logger.Debugf("[MCP-CONNECT] ExecuteTool called for tool: %s at %s", call.ToolName(), time.Now().Format(time.RFC3339Nano))
	mc.dataLock.RLock()
	defer mc.dataLock.RUnlock()

	serverID, mcpClient, err := mc.findServerAndClientForTool(call.Params.Name)
	if err != nil {
		mc.logger.Errorf("[MCP-CONNECT] [ERROR] findServerAndClientForTool failed: %v", err)
		return nil, err
	}

	timeout := mc.getServerTimeout(serverID)
	callTimeout := time.Duration(timeout * float64(time.Second))
	mc.logger.Debugf("[MCP-CONNECT] About to callToolWithTimeout: tool=%s, serverID=%s, timeout=%.2fs, at=%s", call.ToolName(), serverID, timeout, time.Now().Format(time.RFC3339Nano))
	mc.logToolExecutionStart(call, serverID, timeout)

	result, execErr, timedOut := mc.callToolWithTimeout(ctx, mcpClient, call, callTimeout)
	mc.logger.Debugf("[MCP-CONNECT] callToolWithTimeout finished: tool=%s, serverID=%s, timedOut=%v, execErr=%v, at=%s", call.ToolName(), serverID, timedOut, execErr, time.Now().Format(time.RFC3339Nano))
	mc.logger.Infof("<<< Tool execution complete in %s", callTimeout)

	if timedOut {
		mc.logToolTimeout(call, serverID, timeout)
		return nil, error_handling.NewError(
			fmt.Sprintf("tool `%s` execution timed out after %.0f seconds", call.Params.Name, timeout),
			error_handling.ErrorCategoryInternal,
		)
	}

	if execErr != nil {
		mc.logToolError(call, serverID, timeout, execErr)
		return nil, error_handling.WrapError(
			execErr,
			fmt.Sprintf("failed to call tool `%s`", call.Params.Name),
			error_handling.ErrorCategoryInternal,
		)
	}
	return result, nil
}

// findServerAndClientForTool searches for the server and client by tool name.
func (mc *MCPConnector) findServerAndClientForTool(toolName string) (string, client.MCPClient, error) {
	for serverID, serverTools := range mc.tools {
		for _, tool := range serverTools {
			if tool.Name == toolName {
				mcpClient, exists := mc.clients[serverID]
				if !exists {
					return "", nil, error_handling.NewError(
						fmt.Sprintf("not connected to server: %s", serverID),
						error_handling.ErrorCategoryValidation,
					)
				}
				return serverID, mcpClient, nil
			}
		}
	}
	return "", nil, error_handling.NewError(
		fmt.Sprintf("tool `%s` not found", toolName),
		error_handling.ErrorCategoryValidation,
	)
}

// getServerTimeout returns the timeout for the server.
func (mc *MCPConnector) getServerTimeout(serverID string) float64 {
	timeout := 30.0
	if srvCfg, ok := mc.config.McpServers[serverID]; ok && srvCfg.Timeout > 0 {
		timeout = srvCfg.Timeout
	}
	return timeout
}

// logToolExecutionStart logs the start of tool execution.
func (mc *MCPConnector) logToolExecutionStart(call types.CallToolRequest, serverID string, timeout float64) {
	mc.logger.Infof(
		">>> Execute tool `%s` (server_id=%s, timeout_sec=%.0f, arguments=%v)",
		call.ToolName(), serverID, timeout, call.Params.Arguments,
	)
	mc.logger.Debugf(">>> Details: %s", call.Params.Arguments)
}

// callToolWithTimeout calls the tool with a timeout.
func (mc *MCPConnector) callToolWithTimeout(ctx context.Context, mcpClient client.MCPClient, call types.CallToolRequest, callTimeout time.Duration) (*mcp.CallToolResult, error, bool) {
	mc.logger.Debugf("[MCP-CONNECT] callToolWithTimeout: tool=%s, timeout=%s, at=%s", call.ToolName(), callTimeout, time.Now().Format(time.RFC3339Nano))
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)

	go func() {
		mc.logger.Debugf("[MCP-CONNECT] goroutine started for tool=%s at %s", call.ToolName(), time.Now().Format(time.RFC3339Nano))
		result, err := mcpClient.CallTool(ctxWithCancel, call.CallToolRequest)
		if err != nil {
			mc.logger.Warnf("[MCP-CONNECT] goroutine: error for tool=%s: %v at %s", call.ToolName(), err, time.Now().Format(time.RFC3339Nano))
			errCh <- err
			return
		}
		mc.logger.Debugf("[MCP-CONNECT] goroutine: result for tool=%s at %s", call.ToolName(), time.Now().Format(time.RFC3339Nano))
		resultCh <- result
	}()

	timer := time.NewTimer(callTimeout)
	defer timer.Stop()

	select {
	case result := <-resultCh:
		mc.logger.Debugf("[MCP-CONNECT] callToolWithTimeout: result received for tool=%s at %s", call.ToolName(), time.Now().Format(time.RFC3339Nano))
		return result, nil, false
	case err := <-errCh:
		mc.logger.Warnf("[MCP-CONNECT] callToolWithTimeout: error received for tool=%s: %v at %s", call.ToolName(), err, time.Now().Format(time.RFC3339Nano))
		return nil, err, false
	case <-timer.C:
		mc.logger.Warnf("[MCP-CONNECT] callToolWithTimeout: timeout for tool=%s at %s", call.ToolName(), time.Now().Format(time.RFC3339Nano))
		cancel()
		return nil, nil, true
	}
}

// logToolTimeout logs the tool timeout.
func (mc *MCPConnector) logToolTimeout(call types.CallToolRequest, serverID string, timeout float64) {
	mc.logger.WithFields(map[string]interface{}{
		"tool":        call.ToolName(),
		"arguments":   call.Params.Arguments,
		"server_id":   serverID,
		"timeout_sec": timeout,
	}).Warnf("Tool execution timed out after %.0f seconds", timeout)
}

// logToolError logs the tool execution error.
func (mc *MCPConnector) logToolError(call types.CallToolRequest, serverID string, timeout float64, err error) {
	fields := map[string]interface{}{
		"tool":        call.ToolName(),
		"arguments":   call.Params.Arguments,
		"server_id":   serverID,
		"timeout_sec": timeout,
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		fields["context_err"] = err.Error()
		mc.logger.WithFields(fields).Warnf("Tool execution canceled due to context error: %T", err)
	} else {
		fields["error"] = err.Error()
		mc.logger.WithFields(fields).Errorf("Failed to execute tool")
	}
}

// Close closes all client connections.
func (mc *MCPConnector) Close() error {
	mc.logger.Debugf("[MCP-CONNECT] Close: acquiring dataLock at %s", time.Now().Format(time.RFC3339Nano))
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
