// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import (
	"github.com/korchasa/speelka-agent-go/internal/types"
)

// DefaultLoader implements the LoaderSpec interface for loading default configuration values.
type DefaultLoader struct{}

// NewDefaultLoader creates a new DefaultLoader.
func NewDefaultLoader() *DefaultLoader {
	return &DefaultLoader{}
}

// LoadConfiguration returns a Config with default values set.
func (l *DefaultLoader) LoadConfiguration() (*types.Configuration, error) {
	config := types.NewConfiguration()

	// Set default values for RuntimeConfig
	config.Runtime.Log.RawLevel = "info"
	config.Runtime.Log.RawOutput = "stderr"

	// Set default values for Runtime Transport
	config.Runtime.Transports.Stdio.Enabled = true
	config.Runtime.Transports.Stdio.BufferSize = 8192
	config.Runtime.Transports.HTTP.Enabled = false
	config.Runtime.Transports.HTTP.Host = "localhost"
	config.Runtime.Transports.HTTP.Port = 3000

	// Set default values for AgentConfig
	config.Agent.Name = "speelka-agent"
	config.Agent.Version = "1.0.0"

	// Set default values for Tool configuration
	config.Agent.Tool.Name = "process"
	config.Agent.Tool.Description = "Process user queries with LLM"
	config.Agent.Tool.ArgumentName = "input"
	config.Agent.Tool.ArgumentDescription = "The user query to process"

	// Set default values for Chat configuration
	config.Agent.Chat.MaxTokens = 8192
	config.Agent.Chat.CompactionStrategy = "delete-old"
	config.Agent.Chat.MaxLLMIterations = 100
	config.Agent.Chat.RequestBudget = 1.0

	// Set default values for LLM configuration
	config.Agent.LLM.Provider = "openai"
	config.Agent.LLM.Model = "gpt-4"
	config.Agent.LLM.PromptTemplate = "You are a helpful assistant. Respond to the following request: {{input}}. Available tools: {{tools}}"
	config.Agent.LLM.Temperature = 0.7

	// Set default values for LLM retry configuration
	config.Agent.LLM.Retry.MaxRetries = 3
	config.Agent.LLM.Retry.InitialBackoff = 1.0
	config.Agent.LLM.Retry.MaxBackoff = 30.0
	config.Agent.LLM.Retry.BackoffMultiplier = 2.0

	// Set default values for Connection retry configuration
	config.Agent.Connections.Retry.MaxRetries = 3
	config.Agent.Connections.Retry.InitialBackoff = 1.0
	config.Agent.Connections.Retry.MaxBackoff = 30.0
	config.Agent.Connections.Retry.BackoffMultiplier = 2.0

	// Initialize empty server connections map
	config.Agent.Connections.McpServers = make(map[string]types.MCPServerConnection)

	return config, nil
}
