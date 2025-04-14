// Package types defines the interfaces for the MCP server components.
// Responsibility: Defining types and interfaces for all system components
// Features: Serves as the foundation for component implementation and ensures loose coupling between them
package types

import (
	"context"
)

// ConfigurationManagerSpec represents the interface for managing configuration.
// Responsibility: Providing unified access to system configuration
// Features: Supports configuration loading from files and strings,
// provides access to various types of configuration parameters
type ConfigurationManagerSpec interface {
	// LoadConfiguration loads configuration from various sources based on context.
	// It first tries to load from a configuration file if specified,
	// then applies environment variables (which take precedence).
	// Finally, it validates the loaded configuration.
	// It returns an error if the loading or validation fails.
	LoadConfiguration(ctx context.Context, configFilePath string) error

	// GetMCPServerConfig returns the MCP server configuration.
	GetMCPServerConfig() MCPServerConfig

	// GetMCPConnectorConfig returns the MCP connector configuration.
	GetMCPConnectorConfig() MCPConnectorConfig

	// GetLLMConfig returns the LLM configuration.
	GetLLMConfig() LLMConfig

	// GetLogConfig returns the logging configuration.
	GetLogConfig() LogConfig

	// GetAgentConfig returns the agent configuration.
	GetAgentConfig() AgentConfig
}
