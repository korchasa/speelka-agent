// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
// Features: Uses JSON-based configuration through environment variable
package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/korchasa/speelka-agent-go/internal/types"
	log "github.com/sirupsen/logrus"
)

// Manager implements the types.ConfigurationManagerSpec interface.
// Responsibility: Managing application configuration
// Features: Reads settings from CONFIG_JSON environment variable
type Manager struct {
	logger             *log.Logger
	mcpServerConfig    types.MCPServerConfig
	mcpConnectorConfig types.MCPConnectorConfig
	llmServiceConfig   types.LLMConfig
	logConfig          types.LogConfig
}

// Configuration represents the complete JSON configuration structure
type Configuration struct {
	Agent struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Tool    struct {
			Name                string `json:"name"`
			Description         string `json:"description"`
			ArgumentName        string `json:"argument_name"`
			ArgumentDescription string `json:"argument_description"`
		} `json:"tool"`
		LLM struct {
			Provider       string  `json:"provider"`
			APIKey         string  `json:"api_key"`
			Model          string  `json:"model"`
			MaxTokens      int     `json:"max_tokens"`
			Temperature    float64 `json:"temperature"`
			PromptTemplate string  `json:"prompt_template"`
			Retry          struct {
				MaxRetries        int     `json:"max_retries"`
				InitialBackoff    float64 `json:"initial_backoff"`
				MaxBackoff        float64 `json:"max_backoff"`
				BackoffMultiplier float64 `json:"backoff_multiplier"`
			} `json:"retry"`
		} `json:"llm"`
		Connections struct {
			Servers []struct {
				ID          string            `json:"id"`
				Transport   string            `json:"transport"`
				URL         string            `json:"url,omitempty"`
				APIKey      string            `json:"api_key,omitempty"`
				Command     string            `json:"command,omitempty"`
				Arguments   []string          `json:"arguments,omitempty"`
				Environment map[string]string `json:"environment,omitempty"`
			} `json:"servers"`
			Retry struct {
				MaxRetries        int     `json:"max_retries"`
				InitialBackoff    float64 `json:"initial_backoff"`
				MaxBackoff        float64 `json:"max_backoff"`
				BackoffMultiplier float64 `json:"backoff_multiplier"`
			} `json:"retry"`
		} `json:"connections"`
	} `json:"agent"`
	Runtime struct {
		Log struct {
			Level  string `json:"level"`
			Format string `json:"format"`
			Output string `json:"output"`
		} `json:"log"`
		Transports struct {
			Stdio struct {
				Enabled    bool `json:"enabled"`
				BufferSize int  `json:"buffer_size"`
				AutoDetect bool `json:"auto_detect"`
			} `json:"stdio"`
			HTTP struct {
				Enabled bool   `json:"enabled"`
				Host    string `json:"host"`
				Port    int    `json:"port"`
			} `json:"http,omitempty"`
		} `json:"transports"`
	} `json:"runtime"`
}

// NewConfigurationManager creates a new instance of ConfigurationManagerSpec.
// Responsibility: Factory method for creating a configuration manager
// Features: Returns a simple instance without initialization
func NewConfigurationManager(logger *log.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// LoadConfiguration loads configuration from CONFIG_JSON environment variable
// Responsibility: Loading configuration settings
// Features: Loads and parses configuration from JSON
func (cm *Manager) LoadConfiguration(ctx context.Context) error {
	jsonConfig := os.Getenv("CONFIG_JSON")
	if jsonConfig == "" {
		return fmt.Errorf("CONFIG_JSON environment variable is not set")
	}

	err := cm.loadFromJSON(jsonConfig)
	if err != nil {
		return fmt.Errorf("failed to load configuration from CONFIG_JSON: %w", err)
	}

	cm.logger.Info("Loaded configuration from CONFIG_JSON environment variable")
	return nil
}

// loadFromJSON loads configuration from a JSON string.
// Responsibility: Parsing JSON configuration
// Features: Converts JSON to configuration structs
func (cm *Manager) loadFromJSON(jsonConfig string) error {
	var config Configuration
	err := json.Unmarshal([]byte(jsonConfig), &config)
	if err != nil {
		return fmt.Errorf("error parsing CONFIG_JSON: %w", err)
	}

	// Convert the JSON configuration to the internal configuration types
	// MCP Server Config
	cm.mcpServerConfig = types.MCPServerConfig{
		Name:    config.Agent.Name,
		Version: config.Agent.Version,
		Tool: types.MCPServerToolConfig{
			Name:                config.Agent.Tool.Name,
			Description:         config.Agent.Tool.Description,
			ArgumentName:        config.Agent.Tool.ArgumentName,
			ArgumentDescription: config.Agent.Tool.ArgumentDescription,
		},
		HTTP: types.HTTPConfig{
			Enabled: config.Runtime.Transports.HTTP.Enabled,
			Host:    config.Runtime.Transports.HTTP.Host,
			Port:    config.Runtime.Transports.HTTP.Port,
		},
		Stdio: types.StdioConfig{
			Enabled:    config.Runtime.Transports.Stdio.Enabled,
			BufferSize: config.Runtime.Transports.Stdio.BufferSize,
			AutoDetect: config.Runtime.Transports.Stdio.AutoDetect,
		},
		Debug: false, // Debug flag is removed in the new structure
	}

	// MCP Connector Config
	var servers []types.MCPServerConnection
	for _, server := range config.Agent.Connections.Servers {
		// Convert environment map to slice of "KEY=VALUE" strings
		var envVars []string
		for key, value := range server.Environment {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}

		servers = append(servers, types.MCPServerConnection{
			ID:          server.ID,
			Transport:   server.Transport,
			URL:         server.URL,
			APIKey:      server.APIKey,
			Command:     server.Command,
			Arguments:   server.Arguments,
			Environment: envVars,
		})
	}

	cm.mcpConnectorConfig = types.MCPConnectorConfig{
		Servers: servers,
		RetryConfig: types.RetryConfig{
			MaxRetries:        config.Agent.Connections.Retry.MaxRetries,
			InitialBackoff:    config.Agent.Connections.Retry.InitialBackoff,
			MaxBackoff:        config.Agent.Connections.Retry.MaxBackoff,
			BackoffMultiplier: config.Agent.Connections.Retry.BackoffMultiplier,
		},
	}

	// LLM Config
	cm.llmServiceConfig = types.LLMConfig{
		Provider:             config.Agent.LLM.Provider,
		Model:                config.Agent.LLM.Model,
		APIKey:               config.Agent.LLM.APIKey,
		MaxTokens:            config.Agent.LLM.MaxTokens,
		Temperature:          config.Agent.LLM.Temperature,
		SystemPromptTemplate: config.Agent.LLM.PromptTemplate,
		RetryConfig: types.RetryConfig{
			MaxRetries:        config.Agent.LLM.Retry.MaxRetries,
			InitialBackoff:    config.Agent.LLM.Retry.InitialBackoff,
			MaxBackoff:        config.Agent.LLM.Retry.MaxBackoff,
			BackoffMultiplier: config.Agent.LLM.Retry.BackoffMultiplier,
		},
	}

	// Check for LLM_API_KEY environment variable and override config if present
	if envAPIKey := os.Getenv("LLM_API_KEY"); envAPIKey != "" {
		cm.llmServiceConfig.APIKey = envAPIKey
		cm.logger.Info("Using LLM API key from LLM_API_KEY environment variable")
	}

	// Log Config
	level, err := log.ParseLevel(config.Runtime.Log.Level)
	if err != nil {
		return fmt.Errorf("invalid log level `%s`: %v", config.Runtime.Log.Level, err)
	}

	var formatter log.Formatter
	if config.Runtime.Log.Format == "json" {
		formatter = &log.JSONFormatter{}
	} else {
		formatter = &log.TextFormatter{}
	}

	var output io.Writer
	switch config.Runtime.Log.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		outputFile, err := os.OpenFile(config.Runtime.Log.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file `%s`: %v", config.Runtime.Log.Output, err)
		}
		output = outputFile
	}

	cm.logConfig = types.LogConfig{
		Level:     level,
		Formatter: formatter,
		Output:    output,
	}

	return nil
}

// GetMCPServerConfig returns the MCP server configuration.
func (cm *Manager) GetMCPServerConfig() types.MCPServerConfig {
	return cm.mcpServerConfig
}

// GetMCPConnectorConfig returns the MCP connector configuration.
func (cm *Manager) GetMCPConnectorConfig() types.MCPConnectorConfig {
	return cm.mcpConnectorConfig
}

// GetLLMConfig returns the LLM configuration.
func (cm *Manager) GetLLMConfig() types.LLMConfig {
	return cm.llmServiceConfig
}

// GetLogConfig returns the logging configuration.
func (cm *Manager) GetLogConfig() types.LogConfig {
	return cm.logConfig
}

// GetString returns a string value from configuration by key.
// This method is maintained for interface compatibility but always returns false
// as all configuration is now handled through CONFIG_JSON.
func (cm *Manager) GetString(key string) (string, bool) {
	return "", false
}

// GetInt returns an integer value from configuration by key.
// This method is maintained for interface compatibility but always returns false
// as all configuration is now handled through CONFIG_JSON.
func (cm *Manager) GetInt(key string) (int, bool) {
	return 0, false
}

// GetFloat returns a floating point value from configuration by key.
// This method is maintained for interface compatibility but always returns false
// as all configuration is now handled through CONFIG_JSON.
func (cm *Manager) GetFloat(key string) (float64, bool) {
	return 0, false
}

// GetBool returns a boolean value from configuration by key.
// This method is maintained for interface compatibility but always returns false
// as all configuration is now handled through CONFIG_JSON.
func (cm *Manager) GetBool(key string) (bool, bool) {
	return false, false
}

// GetStringMap returns a string-string map from configuration by key.
// This method is maintained for interface compatibility but always returns false
// as all configuration is now handled through CONFIG_JSON.
func (cm *Manager) GetStringMap(key string) (map[string]string, bool) {
	return nil, false
}
