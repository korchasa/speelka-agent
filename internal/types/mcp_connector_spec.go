// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPConnectorSpec represents the interface for the MCP connector component.
// Responsibility: Defining the contract for the MCP connector
// Features: Defines methods for connecting to MCP servers and executing tools
type MCPConnectorSpec interface {
	// InitAndConnectToMCPs initializes connections to all configured MCP servers.
	// It returns an error if any connection fails.
	InitAndConnectToMCPs(ctx context.Context) error

	// ConnectServer connects to a specific MCP server.
	// It returns the client for the server and an error if the connection fails.
	ConnectServer(ctx context.Context, serverID string, serverConfig MCPServerConnection) (client.MCPClient, error)

	// GetAllTools returns a list of all tools available on all connected MCP servers.
	// It returns an error if the tool discovery fails.
	GetAllTools(ctx context.Context) ([]mcp.Tool, error)

	// ExecuteTool executes a tool on the appropriate MCP server.
	// It returns the result of the tool execution and an error if the execution fails.
	ExecuteTool(ctx context.Context, call CallToolRequest) (*mcp.CallToolResult, error)

	// Close closes all connections to MCP servers.
	// It returns an error if any connection fails to close.
	Close() error
}
