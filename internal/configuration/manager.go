package configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/utils"

	"github.com/korchasa/speelka-agent-go/internal/types"
)

// Manager implements the types.ConfigurationManagerSpec interface.
// Responsibility: Managing application configuration by coordinating multiple loaders
type Manager struct {
	logger        types.LoggerSpec
	config        *types.Configuration
	defaultLoader LoaderSpec
	envLoader     LoaderSpec
}

// NewConfigurationManager creates a new instance of ConfigurationManagerSpec.
// Responsibility: Factory method for creating a configuration manager
func NewConfigurationManager(logger types.LoggerSpec) *Manager {
	manager := &Manager{
		logger: logger,
	}
	// Initialize loaders
	manager.defaultLoader = NewDefaultLoader()
	manager.envLoader = NewEnvLoader()

	return manager
}

// LoadConfiguration loads configuration using the configured loaders.
// It first loads default values, then from a configuration file if specified,
// and finally applies environment variables which take precedence.
// Responsibility: Coordinating the loading of configuration from multiple sources
func (cm *Manager) LoadConfiguration(ctx context.Context, configFilePath string) error {
	cm.logger.Infof("Loading configuration...")

	// Start with default configuration
	cm.logger.Debugf("Loading default configuration...")
	defaultConfig, err := cm.defaultLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load default configuration: %w", err)
	}
	cm.config = defaultConfig

	// Load from file if specified
	if configFilePath != "" {
		cm.logger.Infof("Loading configuration from file: %s", configFilePath)
		var fileLoader LoaderSpec

		// Choose loader based on file extension
		if strings.HasSuffix(configFilePath, ".yaml") || strings.HasSuffix(configFilePath, ".yml") {
			fileLoader = NewYAMLLoader(configFilePath)
		} else if strings.HasSuffix(configFilePath, ".json") {
			fileLoader = NewJSONLoader(configFilePath)
		} else {
			return fmt.Errorf("unsupported configuration file format: %s", configFilePath)
		}

		// Load configuration from file
		fileConfig, err := fileLoader.LoadConfiguration()
		if err != nil {
			return fmt.Errorf("failed to load configuration from file: %w", err)
		}

		// Apply file configuration to default configuration instead of replacing it
		cm.config.Apply(fileConfig)
	}

	// Load and apply environment variables (highest precedence)
	cm.logger.Debugf("Loading environment variables...")
	envConfig, err := cm.envLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Apply environment variables if they exist
	cm.logger.Debugf("Applying environment configurations...")
	cm.config.Apply(envConfig)

	cm.logger.Infof("Configuration: %s", utils.SDump(cm.config.RedactedCopy()))

	// Validate the final configuration
	cm.logger.Debugf("Validating configuration...")
	err = cm.config.Validate()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.logger.Infof("Configuration loaded successfully")
	return nil
}

// GetMCPServerConfig returns the MCP server configuration.
func (cm *Manager) GetMCPServerConfig() types.MCPServerConfig {
	// Convert from the new configuration format to the MCP server config format
	mcpConfig := types.MCPServerConfig{
		Name: cm.config.Agent.Name,
		// Use version from agent config
		Version: cm.config.Agent.Version,
		Tool: types.MCPServerToolConfig{
			Name:                cm.config.Agent.Tool.Name,
			Description:         cm.config.Agent.Tool.Description,
			ArgumentName:        cm.config.Agent.Tool.ArgumentName,
			ArgumentDescription: cm.config.Agent.Tool.ArgumentDescription,
		},
		// Use HTTP config from RuntimeTransportConfig
		HTTP: types.HTTPConfig{
			Enabled: cm.config.Runtime.Transports.HTTP.Enabled,
			Host:    cm.config.Runtime.Transports.HTTP.Host,
			Port:    cm.config.Runtime.Transports.HTTP.Port,
		},
		// Use Stdio config from RuntimeTransportConfig
		Stdio: types.StdioConfig{
			Enabled:    cm.config.Runtime.Transports.Stdio.Enabled,
			BufferSize: cm.config.Runtime.Transports.Stdio.BufferSize,
		},
	}

	return mcpConfig
}

// GetMCPConnectorConfig returns the MCP connector configuration.
func (cm *Manager) GetMCPConnectorConfig() types.MCPConnectorConfig {
	// Convert from the new configuration format to the MCP connector config format
	mcpConnectorConfig := types.MCPConnectorConfig{
		McpServers: make(map[string]types.MCPServerConnection),
		RetryConfig: types.RetryConfig{
			MaxRetries:        cm.config.Agent.Connections.Retry.MaxRetries,
			InitialBackoff:    cm.config.Agent.Connections.Retry.InitialBackoff,
			MaxBackoff:        cm.config.Agent.Connections.Retry.MaxBackoff,
			BackoffMultiplier: cm.config.Agent.Connections.Retry.BackoffMultiplier,
		},
	}

	// Set default values for retry config if not specified
	if mcpConnectorConfig.RetryConfig.MaxRetries == 0 {
		mcpConnectorConfig.RetryConfig.MaxRetries = 3
	}
	if mcpConnectorConfig.RetryConfig.InitialBackoff == 0 {
		mcpConnectorConfig.RetryConfig.InitialBackoff = 1.0
	}
	if mcpConnectorConfig.RetryConfig.MaxBackoff == 0 {
		mcpConnectorConfig.RetryConfig.MaxBackoff = 60.0
	}
	if mcpConnectorConfig.RetryConfig.BackoffMultiplier == 0 {
		mcpConnectorConfig.RetryConfig.BackoffMultiplier = 2.0
	}

	// Convert MCP server connections
	for id, conn := range cm.config.Agent.Connections.McpServers {
		mcpConnectorConfig.McpServers[id] = types.MCPServerConnection{
			URL:          conn.URL,
			APIKey:       conn.APIKey,
			Command:      conn.Command,
			Args:         conn.Args,
			Environment:  conn.Environment,
			IncludeTools: conn.IncludeTools,
			ExcludeTools: conn.ExcludeTools,
			Timeout:      conn.Timeout,
		}
	}

	return mcpConnectorConfig
}

// GetLLMConfig returns the LLM configuration.
func (cm *Manager) GetLLMConfig() types.LLMConfig {
	// Create a new LLMConfig based on the configuration
	llmConfig := types.LLMConfig{
		Provider:             cm.config.Agent.LLM.Provider,
		APIKey:               cm.config.Agent.LLM.APIKey,
		Model:                cm.config.Agent.LLM.Model,
		MaxTokens:            cm.config.Agent.LLM.MaxTokens,
		SystemPromptTemplate: cm.config.Agent.LLM.PromptTemplate,
	}

	// Set MaxTokens if explicitly provided, otherwise leave at default
	if cm.config.Agent.LLM.IsMaxTokensSet {
		llmConfig.MaxTokens = cm.config.Agent.LLM.MaxTokens
		llmConfig.IsMaxTokensSet = true
	}

	// Set Temperature if provided
	llmConfig.Temperature = cm.config.Agent.LLM.Temperature
	llmConfig.IsTemperatureSet = true

	// Set RetryConfig with default values if not specified
	llmConfig.RetryConfig = types.RetryConfig{
		MaxRetries:        cm.config.Agent.LLM.Retry.MaxRetries,
		InitialBackoff:    cm.config.Agent.LLM.Retry.InitialBackoff,
		MaxBackoff:        cm.config.Agent.LLM.Retry.MaxBackoff,
		BackoffMultiplier: cm.config.Agent.LLM.Retry.BackoffMultiplier,
	}

	// Set default values for retry config if not specified
	if llmConfig.RetryConfig.MaxRetries == 0 {
		llmConfig.RetryConfig.MaxRetries = 3
	}
	if llmConfig.RetryConfig.InitialBackoff == 0 {
		llmConfig.RetryConfig.InitialBackoff = 1.0
	}
	if llmConfig.RetryConfig.MaxBackoff == 0 {
		llmConfig.RetryConfig.MaxBackoff = 60.0
	}
	if llmConfig.RetryConfig.BackoffMultiplier == 0 {
		llmConfig.RetryConfig.BackoffMultiplier = 2.0
	}

	return llmConfig
}

// GetLogConfig returns the logging configuration.
func (cm *Manager) GetLogConfig() types.LogConfig {
	// Convert from the new configuration format to the Log config format
	logConfig := types.LogConfig{
		RawLevel:  cm.config.Runtime.Log.RawLevel,
		RawOutput: cm.config.Runtime.Log.RawOutput,
		Level:     cm.config.Runtime.Log.LogLevel,
		Output:    cm.config.Runtime.Log.Output,
	}
	return logConfig
}

// GetAgentConfig returns the agent configuration.
func (cm *Manager) GetAgentConfig() types.AgentConfig {
	// Convert from the new configuration format to the Agent config format
	agentConfig := types.AgentConfig{
		Tool: types.MCPServerToolConfig{
			Name:                cm.config.Agent.Tool.Name,
			Description:         cm.config.Agent.Tool.Description,
			ArgumentName:        cm.config.Agent.Tool.ArgumentName,
			ArgumentDescription: cm.config.Agent.Tool.ArgumentDescription,
		},
		Model:                cm.config.Agent.LLM.Model,
		SystemPromptTemplate: cm.config.Agent.LLM.PromptTemplate,
		MaxTokens:            cm.config.Agent.Chat.MaxTokens,
		MaxLLMIterations:     cm.config.Agent.Chat.MaxLLMIterations,
	}

	return agentConfig
}
