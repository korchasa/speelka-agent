// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
// Features: Uses environment variables for configuration
package configuration

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/types"
	log "github.com/sirupsen/logrus"
)

// Manager implements the types.ConfigurationManagerSpec interface.
// Responsibility: Managing application configuration
// Features: Reads settings from environment variables
type Manager struct {
	logger             *log.Logger
	mcpServerConfig    types.MCPServerConfig
	mcpConnectorConfig types.MCPConnectorConfig
	llmServiceConfig   types.LLMConfig
	logConfig          types.LogConfig
}

// Configuration represents the complete JSON configuration structure
// Kept for backward compatibility
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

// LoadConfiguration loads configuration from environment variables
// Responsibility: Loading configuration settings
// Features: Loads configuration from environment variables
func (cm *Manager) LoadConfiguration(ctx context.Context) error {
	// Try to load from environment variables
	err := cm.loadFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load configuration from environment variables: %w", err)
	}

	cm.logger.Info("Loaded configuration from environment variables")
	return nil
}

// loadFromEnvironment loads configuration from environment variables.
// Responsibility: Parsing environment variable configuration
// Features: Converts environment variables to configuration structs
func (cm *Manager) loadFromEnvironment() error {
	var validationErrors []string

	// MCP Server Config
	cm.mcpServerConfig = types.MCPServerConfig{
		Name:    getEnvString("AGENT_NAME", ""),
		Version: getEnvString("AGENT_VERSION", "1.0.0"),
		Tool: types.MCPServerToolConfig{
			Name:                getEnvString("TOOL_NAME", ""),
			Description:         getEnvString("TOOL_DESCRIPTION", ""),
			ArgumentName:        getEnvString("TOOL_ARGUMENT_NAME", "query"),
			ArgumentDescription: getEnvString("TOOL_ARGUMENT_DESCRIPTION", ""),
		},
		HTTP: types.HTTPConfig{
			Enabled: getEnvBool("RUNTIME_HTTP_ENABLED", false),
			Host:    getEnvString("RUNTIME_HTTP_HOST", "localhost"),
			Port:    getEnvInt("RUNTIME_HTTP_PORT", 3000),
		},
		Stdio: types.StdioConfig{
			Enabled:    getEnvBool("RUNTIME_STDIO_ENABLED", true),
			BufferSize: getEnvInt("RUNTIME_STDIO_BUFFER_SIZE", 8192),
		},
		Debug: false,
	}

	// Validate required fields for MCP Server Config
	if cm.mcpServerConfig.Name == "" {
		validationErrors = append(validationErrors, "AGENT_NAME environment variable is required")
	}
	if cm.mcpServerConfig.Tool.Name == "" {
		validationErrors = append(validationErrors, "TOOL_NAME environment variable is required")
	}
	if cm.mcpServerConfig.Tool.Description == "" {
		validationErrors = append(validationErrors, "TOOL_DESCRIPTION environment variable is required")
	}

	// MCP Connector Config - Handle MCP Servers
	mcpServers := make(map[string]types.MCPServerConnection)

	// Find all MCPS_n_ID environment variables to determine how many servers there are
	serverIndices := findServerIndices()

	// Process each server
	for _, idx := range serverIndices {
		idxStr := strconv.Itoa(idx)
		serverID := getEnvString(fmt.Sprintf("MCPS_%s_ID", idxStr), "")
		if serverID == "" {
			cm.logger.Warnf("MCP Server at index %s has no ID", idxStr)
			continue
		}

		// Parse args string into slice (if provided as space-separated string)
		argsStr := getEnvString(fmt.Sprintf("MCPS_%s_ARGS", idxStr), "")
		var args []string
		if argsStr != "" {
			args = strings.Fields(argsStr)
		}

		mcpServers[serverID] = types.MCPServerConnection{
			URL:         getEnvString(fmt.Sprintf("MCPS_%s_URL", idxStr), ""),
			APIKey:      getEnvString(fmt.Sprintf("MCPS_%s_API_KEY", idxStr), ""),
			Command:     getEnvString(fmt.Sprintf("MCPS_%s_COMMAND", idxStr), ""),
			Args:        args,
			Environment: []string{}, // Environment variables not loaded yet
		}
	}

	// Set the MCP Connector Config with the servers and retry settings
	cm.mcpConnectorConfig = types.MCPConnectorConfig{
		McpServers: mcpServers,
		RetryConfig: types.RetryConfig{
			MaxRetries:        getEnvInt("MSPS_RETRY_MAX_RETRIES", 3),
			InitialBackoff:    getEnvFloat("MSPS_RETRY_INITIAL_BACKOFF", 1.0),
			MaxBackoff:        getEnvFloat("MSPS_RETRY_MAX_BACKOFF", 30.0),
			BackoffMultiplier: getEnvFloat("MSPS_RETRY_BACKOFF_MULTIPLIER", 2.0),
		},
	}

	// LLM Config
	promptTemplate := getEnvString("LLM_PROMPT_TEMPLATE", "")
	cm.llmServiceConfig = types.LLMConfig{
		Provider:             getEnvString("LLM_PROVIDER", ""),
		Model:                getEnvString("LLM_MODEL", ""),
		APIKey:               getEnvString("LLM_API_KEY", ""),
		MaxTokens:            getEnvInt("LLM_MAX_TOKENS", 0),
		Temperature:          getEnvFloat("LLM_TEMPERATURE", 0.7),
		SystemPromptTemplate: promptTemplate,
		RetryConfig: types.RetryConfig{
			MaxRetries:        getEnvInt("LLM_RETRY_MAX_RETRIES", 3),
			InitialBackoff:    getEnvFloat("LLM_RETRY_INITIAL_BACKOFF", 1.0),
			MaxBackoff:        getEnvFloat("LLM_RETRY_MAX_BACKOFF", 30.0),
			BackoffMultiplier: getEnvFloat("LLM_RETRY_BACKOFF_MULTIPLIER", 2.0),
		},
	}

	// Validate required fields for LLM Config
	if cm.llmServiceConfig.Provider == "" {
		validationErrors = append(validationErrors, "LLM_PROVIDER environment variable is required")
	}
	if cm.llmServiceConfig.Model == "" {
		validationErrors = append(validationErrors, "LLM_MODEL environment variable is required")
	}
	if promptTemplate == "" {
		validationErrors = append(validationErrors, "LLM_PROMPT_TEMPLATE environment variable is required")
	} else {
		// Validate prompt template has all required placeholders
		err := cm.validatePromptTemplate(promptTemplate, cm.mcpServerConfig.Tool.ArgumentName)
		if err != nil {
			cm.logger.Errorf("PROMPT TEMPLATE VALIDATION ERROR: %v", err)
			validationErrors = append(validationErrors, fmt.Sprintf("prompt template validation failed: %v", err))
		}
	}

	// Log Config
	logLevel := getEnvString("RUNTIME_LOG_LEVEL", "info")
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level `%s`: %v", logLevel, err)
	}

	var output io.Writer
	logOutput := getEnvString("RUNTIME_LOG_OUTPUT", "stdout")
	switch logOutput {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		outputFile, err := os.OpenFile(logOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file `%s`: %v", logOutput, err)
		}
		output = outputFile
	}

	cm.logConfig = types.LogConfig{
		Level:  level,
		Output: output,
	}

	// Return validation errors if any were found
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation errors: %s", strings.Join(validationErrors, ", "))
	}

	return nil
}

// Helper function to find server indices from environment variables
func findServerIndices() []int {
	indices := make(map[int]bool)

	// Look for all environment variables starting with MCPS_
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) < 1 {
			continue
		}

		key := parts[0]
		if !strings.HasPrefix(key, "MCPS_") {
			continue
		}

		// Extract the index from the key (e.g., MCPS_0_ID -> 0)
		parts = strings.Split(key, "_")
		if len(parts) < 3 {
			continue
		}

		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		indices[idx] = true
	}

	// Convert to slice
	var result []int
	for idx := range indices {
		result = append(result, idx)
	}

	return result
}

// Helper function to get string from environment variable with default
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get int from environment variable with default
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// Helper function to get float from environment variable with default
func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

// Helper function to get bool from environment variable with default
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
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
