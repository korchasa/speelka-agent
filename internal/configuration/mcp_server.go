package configuration

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

	// MCPLogEnabled determines if MCP logging is enabled.
	MCPLogEnabled bool
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
// Features: Contains all necessary information for connection configuration, including tool filtering options
//
// Optional fields:
//   - IncludeTools: If set, only these tool names will be exported from this server.
//   - ExcludeTools: If set, these tool names will be excluded from export from this server.
type MCPServerConnection struct {
	// URL is the URL of the MCP server (for HTTP transport).
	URL string `json:"url" yaml:"url"`

	// APIKey is the API key for authenticating with the server (for HTTP transport).
	APIKey string `json:"apiKey" yaml:"apiKey"`

	// Command is the command to execute for stdio transport.
	Command string `json:"command" yaml:"command"`

	// Args are the arguments to pass to the command for stdio transport.
	Args []string `json:"args" yaml:"args"`

	// Environment is a list of environment variables to set for the stdio transport command in the format "KEY=VALUE".
	Environment []string `json:"environment" yaml:"environment"`

	// IncludeTools is an optional whitelist of tool names to export from this server. If set, only these tools will be available.
	IncludeTools []string `json:"include_tools,omitempty" yaml:"include_tools,omitempty"`

	// ExcludeTools is an optional blacklist of tool names to exclude from this server. If set, these tools will not be available.
	ExcludeTools []string `json:"exclude_tools,omitempty" yaml:"exclude_tools,omitempty"`

	// Timeout is the tool call timeout for this server, in seconds. If zero, the default is used.
	Timeout float64 `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// IsToolAllowed determines if a tool is allowed based on IncludeTools and ExcludeTools.
//
// Logic:
//   - If IncludeTools is non-empty, only tools in this list are allowed.
//   - If ExcludeTools is non-empty, tools in this list are disallowed, even if present in IncludeTools.
//   - If both lists are empty, all tools are allowed.
//   - Returns true if allowed, false if not.
func (c *MCPServerConnection) IsToolAllowed(name string) bool {
	// Check ExcludeTools first (blacklist has priority)
	for _, ex := range c.ExcludeTools {
		if ex == name {
			return false
		}
	}
	// If IncludeTools is set, only allow if present
	if len(c.IncludeTools) > 0 {
		for _, inc := range c.IncludeTools {
			if inc == name {
				return true
			}
		}
		return false
	}
	// If neither list is set, allow all
	return true
}

// MCPServerConfigForTest returns a sample configuration for MCPServer used in tests.
func MCPServerConfigForTest() MCPServerConfig {
	return MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		MCPLogEnabled: true,
		HTTP:          HTTPConfig{Host: "127.0.0.1", Port: 12345, Enabled: false},
		Stdio:         StdioConfig{Enabled: true},
	}
}
