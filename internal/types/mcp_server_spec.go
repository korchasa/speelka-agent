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
	// Serve initializes and starts the MCP server.
	// It returns an error if the server fails to start.
	Serve(ctx context.Context, daemonMode bool, handler server.ToolHandlerFunc) error

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

// HTTPConfig represents the configuration for HTTP transport.
// Responsibility: Storing HTTP transport configuration parameters
// Features: Contains all necessary parameters for configuring an HTTP server,
// including TLS settings
type HTTPConfig struct {
	// Enabled determines if HTTP transport is active.
	Enabled bool

	// Host is the host to listen on.
	Host string

	// Port is the port to listen on.
	Port int
}

// StdioConfig represents the configuration for stdio transport.
// Responsibility: Storing configuration parameters for stdio transport
// Features: Defines settings for working with stdin/stdout
type StdioConfig struct {
	// Enabled determines if stdio transport is active.
	Enabled bool

	// BufferSize is the size of the read/write buffers.
	BufferSize int
}

// MCPServerConfig represents the configuration for the MCP server.
// Responsibility: Storing the complete MCP server configuration
// Features: Combines configurations for various transport protocols
// and contains general server settings
type MCPServerConfig struct {
	// ID is a unique identifier for this server.
	Name string

	// Version is the version string of the server.
	Version string

	// HTTP contains configuration for HTTP transport.
	HTTP HTTPConfig

	// Stdio contains configuration for stdio transport.
	Stdio StdioConfig

	// Tools is a list of tools available on this server.
	Tool MCPServerToolConfig

	// Debug determines if debug mode is enabled.
	Debug bool
}

type MCPServerToolConfig struct {
	// Name is the name of the tool.
	Name string
	// Description is the description of the tool.
	Description string
	// ArgumentName is the name of the argument for the tool.
	ArgumentName string
	// ArgumentDescription is the description of the argument for the tool.
	ArgumentDescription string
}

// MCPConnectorConfig represents the configuration for the MCP connector.
// Responsibility: Storing parameters for connecting to MCP servers
// Features: Contains a map of servers to connect to and parameters
// for the connection retry strategy
type MCPConnectorConfig struct {
	// McpServers is a map of MCP servers to connect to, with the key being the server ID.
	McpServers map[string]MCPServerConnection

	// RetryConfig is the configuration for retrying failed connections.
	RetryConfig RetryConfig
}

// MCPServerConnection represents a connection to an MCP server.
// Responsibility: Storing parameters for establishing a connection to a specific MCP server
// Features: Contains all necessary information for connection configuration
type MCPServerConnection struct {
	// URL is the URL of the MCP server (for HTTP transport).
	URL string `json:"url" yaml:"url"`

	// APIKey is the API key for authenticating with the server (for HTTP transport).
	APIKey string `json:"api_key" yaml:"api_key"`

	// Command is the command to execute for stdio transport.
	Command string `json:"command" yaml:"command"`

	// Args are the arguments to pass to the command for stdio transport.
	Args []string `json:"args" yaml:"args"`

	// Environment is a list of environment variables to set for the stdio transport command in the format "KEY=VALUE".
	Environment []string `json:"environment" yaml:"environment"`
}
