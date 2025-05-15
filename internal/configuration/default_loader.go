// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import "github.com/korchasa/speelka-agent-go/internal/types"

// DefaultLoader implements the LoaderSpec interface for loading default configuration values.
type DefaultLoader struct{}

// NewDefaultLoader creates a new DefaultLoader.
func NewDefaultLoader() *DefaultLoader {
	return &DefaultLoader{}
}

// LoadConfiguration returns a Config with default values set.
func (l *DefaultLoader) LoadConfiguration() (*types.Configuration, error) {
	cfg := types.NewConfiguration()
	// Runtime defaults
	cfg.Runtime.Log.Output = types.LogOutputMCP
	cfg.Runtime.Log.Format = "text"
	cfg.Runtime.Log.DefaultLevel = "info"
	cfg.Runtime.Transports.Stdio.Enabled = true
	cfg.Runtime.Transports.Stdio.BufferSize = 8192
	cfg.Runtime.Transports.HTTP.Enabled = false
	cfg.Runtime.Transports.HTTP.Host = "localhost"
	cfg.Runtime.Transports.HTTP.Port = 3000
	// Agent defaults
	cfg.Agent.Name = "speelka-agent"
	cfg.Agent.Version = "1.0.0"
	cfg.Agent.Tool.Name = "process"
	cfg.Agent.Tool.Description = "Process user queries with LLM"
	cfg.Agent.Tool.ArgumentName = "input"
	cfg.Agent.Tool.ArgumentDescription = "The user query to process"
	cfg.Agent.Chat.MaxTokens = 8192
	cfg.Agent.Chat.MaxLLMIterations = 100
	cfg.Agent.Chat.RequestBudget = 1.0
	cfg.Agent.LLM.Provider = "openai"
	cfg.Agent.LLM.Model = "gpt-4"
	cfg.Agent.LLM.PromptTemplate = "You are a helpful assistant. Respond to the following request: {{input}}. Available tools: {{tools}}"
	cfg.Agent.LLM.Temperature = 0.7
	cfg.Agent.LLM.APIKey = ""
	cfg.Agent.LLM.Retry.MaxRetries = 3
	cfg.Agent.LLM.Retry.InitialBackoff = 1.0
	cfg.Agent.LLM.Retry.MaxBackoff = 30.0
	cfg.Agent.LLM.Retry.BackoffMultiplier = 2.0
	cfg.Agent.Connections.Retry.MaxRetries = 3
	cfg.Agent.Connections.Retry.InitialBackoff = 1.0
	cfg.Agent.Connections.Retry.MaxBackoff = 30.0
	cfg.Agent.Connections.Retry.BackoffMultiplier = 2.0
	cfg.Agent.Connections.McpServers = make(map[string]types.MCPServerConnection)
	return cfg, nil
}
