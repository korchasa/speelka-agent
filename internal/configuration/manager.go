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
	"regexp"
	"strings"

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
			McpServers map[string]struct {
				URL         string            `json:"url,omitempty"`
				APIKey      string            `json:"api_key,omitempty"`
				Command     string            `json:"command,omitempty"`
				Args        []string          `json:"args,omitempty"`
				Environment map[string]string `json:"environment,omitempty"`
			} `json:"mcpServers"`
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
	mcpServers := make(map[string]types.MCPServerConnection)
	for serverID, server := range config.Agent.Connections.McpServers {
		// Convert environment map to slice of "KEY=VALUE" strings
		var envVars []string
		for key, value := range server.Environment {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}

		mcpServers[serverID] = types.MCPServerConnection{
			URL:         server.URL,
			APIKey:      server.APIKey,
			Command:     server.Command,
			Args:        server.Args,
			Environment: envVars,
		}
	}

	cm.mcpConnectorConfig = types.MCPConnectorConfig{
		McpServers: mcpServers,
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

	// Validate prompt template has all required placeholders
	err = cm.validatePromptTemplate(config.Agent.LLM.PromptTemplate, config.Agent.Tool.ArgumentName)
	if err != nil {
		cm.logger.Errorf("PROMPT TEMPLATE VALIDATION ERROR: %v", err)
		return fmt.Errorf("prompt template validation failed: %w", err)
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
		Level:  level,
		Output: output,
	}

	return nil
}

// validatePromptTemplate validates that the prompt template contains all required placeholders.
// Responsibility: Ensuring prompt templates have required placeholders
// Features: Checks for argument name and tools placeholders
func (cm *Manager) validatePromptTemplate(template string, argumentName string) error {
	// Check if template is empty
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}

	// Extract all placeholders from the template
	placeholders, err := cm.extractPlaceholders(template)
	if err != nil {
		return fmt.Errorf("failed to extract placeholders: %w", err)
	}

	// Required placeholders
	requiredPlaceholders := []string{argumentName, "tools"}

	// Check that all required placeholders are present
	var missingPlaceholders []string
	for _, required := range requiredPlaceholders {
		found := false
		for _, placeholder := range placeholders {
			if placeholder == required {
				found = true
				break
			}
		}
		if !found {
			missingPlaceholders = append(missingPlaceholders, required)
		}
	}

	if len(missingPlaceholders) > 0 {
		// Create detailed error message with an example
		errMsg := fmt.Sprintf("prompt template is missing required placeholder(s): %s", strings.Join(missingPlaceholders, ", "))

		// If argumentName is missing, provide specific guidance
		if contains(missingPlaceholders, argumentName) {
			errMsg += "\n\nExpected placeholder '{{" + argumentName + "}}' should match the 'argument_name' value in your tool configuration."
			errMsg += "\nCommon mistake: Using a hardcoded placeholder name like '{{query}}' instead of the configured argument name."
		}

		// Always provide an example of a valid template
		errMsg += "\n\nExample of a valid template:\nYou are a helpful assistant.\n\nUser request: {{" + argumentName + "}}\n\nAvailable tools:\n{{tools}}"

		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// contains checks if a string is present in a slice
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// extractPlaceholders extracts all placeholders from a template string.
// Responsibility: Finding all template variables in a string
// Features: Uses regex to match placeholders with format {{placeholder_name}}
func (cm *Manager) extractPlaceholders(template string) ([]string, error) {
	re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\s*}}`)
	matches := re.FindAllStringSubmatch(template, -1)

	if matches == nil {
		return []string{}, nil
	}

	var placeholders []string
	for _, match := range matches {
		if len(match) >= 2 {
			placeholders = append(placeholders, match[1])
		}
	}

	return placeholders, nil
}

// TestValidatePromptTemplate exposes validatePromptTemplate for testing.
// This should only be used in tests.
func (cm *Manager) TestValidatePromptTemplate(template string, argumentName string) error {
	return cm.validatePromptTemplate(template, argumentName)
}

// TestExtractPlaceholders exposes extractPlaceholders for testing.
// This should only be used in tests.
func (cm *Manager) TestExtractPlaceholders(template string) ([]string, error) {
	return cm.extractPlaceholders(template)
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

// GetString returns a string value from the configuration.
func (cm *Manager) GetString(key string) (string, bool) {
	// Implementation to be added if needed
	return "", false
}

// GetInt returns an integer value from the configuration.
func (cm *Manager) GetInt(key string) (int, bool) {
	// Implementation to be added if needed
	return 0, false
}

// GetFloat returns a float value from the configuration.
func (cm *Manager) GetFloat(key string) (float64, bool) {
	// Implementation to be added if needed
	return 0, false
}

// GetBool returns a boolean value from the configuration.
func (cm *Manager) GetBool(key string) (bool, bool) {
	// Implementation to be added if needed
	return false, false
}

// GetStringMap returns a string map value from the configuration.
func (cm *Manager) GetStringMap(key string) (map[string]string, bool) {
	// Implementation to be added if needed
	return nil, false
}
