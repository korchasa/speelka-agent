// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServerSpec represents the interface for the MCP server.
// Responsibility: Defining the contract for the MCP server
// Features: Defines methods for starting, stopping, and managing tools
type MCPServerSpec interface {
	// ServeDaemon initializes and starts the HTTP MCP server.
	// It returns an error if the server fails to start.
	ServeDaemon(handler server.ToolHandlerFunc) error

	// ServeStdio initializes and starts the stdio MCP server.
	// It returns an error if the server fails to start.
	ServeStdio(handler server.ToolHandlerFunc) error

	// Stop gracefully shuts down the MCP server.
	// It returns an error if the server fails to stop.
	Stop(ctx context.Context) error

	// AddTool adds a tool to the MCP server.
	AddTool(tool mcp.Tool, handler server.ToolHandlerFunc)

	// GetAllTools returns all tools registered on the server.
	GetAllTools() []mcp.Tool

	// GetServer returns the underlying server instance.
	GetServer() *server.MCPServer
}

// ParameterSpec represents the specification of a parameter.
type ParameterSpec struct {
	// Type is the data type of the parameter.
	Type string

	// Description is a description of the parameter.
	Description string

	// Required indicates whether the parameter is required.
	Required bool

	// Default is the default value of the parameter if it is not provided.
	Default interface{}

	// Enum is a list of possible values for the parameter.
	Enum []interface{}
}
