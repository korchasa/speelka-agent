// Package types defines the interfaces for the MCP server components.
// Responsibility: Defining types and interfaces for all system components
// Features: Serves as the foundation for component implementation and ensures loose coupling between them
package types

import (
    "context"
    "fmt"
    "github.com/sirupsen/logrus"
    "io"
    "strings"
)

// ConfigurationManagerSpec represents the interface for managing configuration.
// Responsibility: Providing unified access to system configuration
// Features: Supports configuration loading from files and strings,
// provides access to various types of configuration parameters
type ConfigurationManagerSpec interface {
    // LoadConfiguration loads configuration from a file.
    // It returns an error if the loading fails.
    LoadConfiguration(ctx context.Context) error

    // GetMCPServerConfig returns the MCP server configuration.
    GetMCPServerConfig() MCPServerConfig

    // GetMCPConnectorConfig returns the MCP connector configuration.
    GetMCPConnectorConfig() MCPConnectorConfig

    // GetLLMConfig returns the LLM configuration.
    GetLLMConfig() LLMConfig

    // GetLogConfig returns the logging configuration.
    GetLogConfig() LogConfig

    // GetString returns a string configuration value.
    // It returns the value and a boolean indicating whether the value was found.
    GetString(key string) (string, bool)

    // GetInt returns an integer configuration value.
    // It returns the value and a boolean indicating whether the value was found.
    GetInt(key string) (int, bool)

    // GetFloat returns a float configuration value.
    // It returns the value and a boolean indicating whether the value was found.
    GetFloat(key string) (float64, bool)

    // GetBool returns a boolean configuration value.
    // It returns the value and a boolean indicating whether the value was found.
    GetBool(key string) (bool, bool)

    // GetStringMap returns a string map configuration value.
    // It returns the value and a boolean indicating whether the value was found.
    GetStringMap(key string) (map[string]string, bool)
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
// Features: Defines settings for working with stdin/stdout,
// including automatic mode detection
type StdioConfig struct {
    // Enabled determines if stdio transport is active.
    Enabled bool

    // BufferSize is the size of the read/write buffers.
    BufferSize int

    // AutoDetect determines if the application should automatically
    // detect and use stdio mode based on environment variables.
    AutoDetect bool
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
// Features: Contains a list of servers to connect to and parameters
// for the connection retry strategy
type MCPConnectorConfig struct {
    // Servers is a list of MCP servers to connect to.
    Servers []MCPServerConnection

    // RetryConfig is the configuration for retrying failed connections.
    RetryConfig RetryConfig
}

// MCPServerConnection represents a connection to an MCP server.
// Responsibility: Storing parameters for establishing a connection to a specific MCP server
// Features: Supports various transport types (HTTP, stdio) and
// contains all necessary information for authentication and connection configuration
type MCPServerConnection struct {
    // ID is a unique identifier for this server connection.
    ID string

    // Transport is the transport type to use ("http" or "stdio").
    Transport string

    // URL is the URL of the MCP server (for HTTP transport).
    URL string

    // APIKey is the API key for authenticating with the server (for HTTP transport).
    APIKey string

    // Command is the command to execute for stdio transport.
    Command string

    // Arguments are the arguments to pass to the command for stdio transport.
    Arguments []string

    // Environment is a list of environment variables to set for the stdio transport command in the format "KEY=VALUE".
    Environment []string
}

func (s *MCPServerConnection) GetDSN() string {
    if s.Transport == "http" {
        return fmt.Sprintf("http://%s@%s", s.ID, s.URL)
    } else {
        return fmt.Sprintf("stdio: id='%s' command='%s %s' envs='%s'", s.ID, s.Command, strings.Join(s.Arguments, " "), strings.Join(s.Environment, ","))
    }
}

// LogConfig represents the configuration for logging.
// Responsibility: Storing logging system settings
// Features: Defines the level, format, and output location for logs
type LogConfig struct {
    // Level is the log level.
    Level logrus.Level

    // Formatter is the log format.
    Formatter logrus.Formatter

    // Output is the log output.
    Output io.Writer
}
