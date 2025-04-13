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
	"github.com/sirupsen/logrus"
)

// Manager implements the types.ConfigurationManagerSpec interface.
// Responsibility: Managing application configuration
// Features: Reads settings from environment variables
type Manager struct {
	logger             types.LoggerSpec
	mcpServerConfig    types.MCPServerConfig
	mcpConnectorConfig types.MCPConnectorConfig
	llmServiceConfig   types.LLMConfig
	logConfig          types.LogConfig
	agentConfig        types.AgentConfig
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
		Chat struct {
			MaxTokens          int    `json:"max_tokens"`
			CompactionStrategy string `json:"compaction_strategy"`
		} `json:"chat"`
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
func NewConfigurationManager(logger types.LoggerSpec) *Manager {
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
	return nil
}

// loadFromEnvironment loads configuration from environment variables.
// Responsibility: Parsing environment variable configuration
// Features: Converts environment variables to configuration structs
func (cm *Manager) loadFromEnvironment() error {
	var validationErrors []string

	// MCP Server Config
	cm.mcpServerConfig = types.MCPServerConfig{
		Name:    getEnvString("SPL_AGENT_NAME", ""),
		Version: getEnvString("SPL_AGENT_VERSION", "1.0.0"),
		Tool: types.MCPServerToolConfig{
			Name:                getEnvString("SPL_TOOL_NAME", ""),
			Description:         getEnvString("SPL_TOOL_DESCRIPTION", ""),
			ArgumentName:        getEnvString("SPL_TOOL_ARGUMENT_NAME", ""),
			ArgumentDescription: getEnvString("SPL_TOOL_ARGUMENT_DESCRIPTION", ""),
		},
		HTTP: types.HTTPConfig{
			Enabled: getEnvBool("SPL_RUNTIME_HTTP_ENABLED", false),
			Host:    getEnvString("SPL_RUNTIME_HTTP_HOST", "localhost"),
			Port:    getEnvInt("SPL_RUNTIME_HTTP_PORT", 3000),
		},
		Stdio: types.StdioConfig{
			Enabled:    getEnvBool("SPL_RUNTIME_STDIO_ENABLED", true),
			BufferSize: getEnvInt("SPL_RUNTIME_STDIO_BUFFER_SIZE", 8192),
		},
		Debug: false,
	}

	// Validate required fields for MCP Server Config
	if cm.mcpServerConfig.Name == "" {
		validationErrors = append(validationErrors, "SPL_AGENT_NAME environment variable is required")
	}
	if cm.mcpServerConfig.Tool.Name == "" {
		validationErrors = append(validationErrors, "SPL_TOOL_NAME environment variable is required")
	}
	if cm.mcpServerConfig.Tool.Description == "" {
		validationErrors = append(validationErrors, "SPL_TOOL_DESCRIPTION environment variable is required")
	}
	if cm.mcpServerConfig.Tool.ArgumentName == "" {
		validationErrors = append(validationErrors, "SPL_TOOL_ARGUMENT_NAME environment variable is required")
	}
	if cm.mcpServerConfig.Tool.ArgumentDescription == "" {
		validationErrors = append(validationErrors, "SPL_TOOL_ARGUMENT_DESCRIPTION environment variable is required")
	}

	// MCP Connector Config - Handle MCP Servers
	mcpServers := make(map[string]types.MCPServerConnection)

	// Find all MCPS_n_ID environment variables to determine how many servers there are
	serverIndices := findServerIndices()

	// Process each server
	for _, idx := range serverIndices {
		idxStr := strconv.Itoa(idx)

		// Only use SPL_ prefixed variables
		serverID := getEnvString(fmt.Sprintf("SPL_MCPS_%s_ID", idxStr), "")
		if serverID == "" {
			cm.logger.Warnf("MCP Server at index %s has no ID", idxStr)
			continue
		}

		// Parse args string into slice (if provided as space-separated string)
		argsStr := getEnvString(fmt.Sprintf("SPL_MCPS_%s_ARGS", idxStr), "")
		var args []string
		if argsStr != "" {
			args = strings.Fields(argsStr)
		}

		// Find and collect environment variables for this server
		// Only check with SPL_ prefixed variables
		splEnvPrefix := fmt.Sprintf("SPL_MCPS_%s_ENV_", idxStr)
		var envVars []string

		for _, fullEnv := range os.Environ() {
			parts := strings.SplitN(fullEnv, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := parts[0]
			value := parts[1]

			if strings.HasPrefix(key, splEnvPrefix) {
				envKey := strings.TrimPrefix(key, splEnvPrefix)
				envVars = append(envVars, fmt.Sprintf("%s=%s", envKey, value))
			}
		}

		mcpServers[serverID] = types.MCPServerConnection{
			URL:         getEnvString(fmt.Sprintf("SPL_MCPS_%s_URL", idxStr), ""),
			APIKey:      getEnvString(fmt.Sprintf("SPL_MCPS_%s_API_KEY", idxStr), ""),
			Command:     getEnvString(fmt.Sprintf("SPL_MCPS_%s_COMMAND", idxStr), ""),
			Args:        args,
			Environment: envVars,
		}
	}

	// MCP Connector Config - Retry
	cm.mcpConnectorConfig = types.MCPConnectorConfig{
		McpServers: mcpServers,
		RetryConfig: types.RetryConfig{
			MaxRetries:        getEnvInt("SPL_MSPS_RETRY_MAX_RETRIES", 3),
			InitialBackoff:    getEnvFloat("SPL_MSPS_RETRY_INITIAL_BACKOFF", 1.0),
			MaxBackoff:        getEnvFloat("SPL_MSPS_RETRY_MAX_BACKOFF", 30.0),
			BackoffMultiplier: getEnvFloat("SPL_MSPS_RETRY_BACKOFF_MULTIPLIER", 2.0),
		},
	}

	// LLM Service Config
	apiKey := getEnvString("SPL_LLM_API_KEY", "")
	provider := getEnvString("SPL_LLM_PROVIDER", "")
	model := getEnvString("SPL_LLM_MODEL", "")
	promptTemplate := getEnvString("SPL_LLM_PROMPT_TEMPLATE", "")

	// Temperature and max tokens handling
	var temperature float64
	var maxTokens int
	var isTemperatureSet, isMaxTokensSet bool

	// Check if the temperature is explicitly set
	temperatureStr := os.Getenv("SPL_LLM_TEMPERATURE")
	if temperatureStr != "" {
		var err error
		temperature, err = strconv.ParseFloat(temperatureStr, 64)
		if err != nil {
			temperature = 0.7 // Default
		} else {
			isTemperatureSet = true
		}
	} else {
		temperature = 0.7 // Default
	}

	// Check if max tokens is explicitly set
	maxTokensStr := os.Getenv("SPL_LLM_MAX_TOKENS")
	if maxTokensStr != "" {
		var err error
		maxTokens, err = strconv.Atoi(maxTokensStr)
		if err != nil {
			maxTokens = 0 // Default, meaning no limit
		} else {
			isMaxTokensSet = true
		}
	} else {
		maxTokens = 0 // Default, meaning no limit
	}

	// Validate required fields for LLM Service Config
	if apiKey == "" {
		validationErrors = append(validationErrors, "SPL_LLM_API_KEY environment variable is required")
	}
	if provider == "" {
		validationErrors = append(validationErrors, "SPL_LLM_PROVIDER environment variable is required")
	}
	if model == "" {
		validationErrors = append(validationErrors, "SPL_LLM_MODEL environment variable is required")
	}
	if promptTemplate == "" {
		validationErrors = append(validationErrors, "SPL_LLM_PROMPT_TEMPLATE environment variable is required")
	}

	// Validate prompt template
	if promptTemplate != "" {
		err := cm.validatePromptTemplate(promptTemplate, cm.mcpServerConfig.Tool.ArgumentName)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid prompt template: %v", err))
		}
	}

	// Set validated LLM Service Config
	cm.llmServiceConfig = types.LLMConfig{
		Provider:             provider,
		Model:                model,
		APIKey:               apiKey,
		MaxTokens:            maxTokens,
		IsMaxTokensSet:       isMaxTokensSet,
		Temperature:          temperature,
		IsTemperatureSet:     isTemperatureSet,
		SystemPromptTemplate: promptTemplate,
		RetryConfig: types.RetryConfig{
			MaxRetries:        getEnvInt("SPL_LLM_RETRY_MAX_RETRIES", 3),
			InitialBackoff:    getEnvFloat("SPL_LLM_RETRY_INITIAL_BACKOFF", 1.0),
			MaxBackoff:        getEnvFloat("SPL_LLM_RETRY_MAX_BACKOFF", 30.0),
			BackoffMultiplier: getEnvFloat("SPL_LLM_RETRY_BACKOFF_MULTIPLIER", 2.0),
		},
	}

	// Log Config
	cm.logConfig = types.LogConfig{
		Level:  cm.loadLevelFromName(getEnvString("SPL_LOG_LEVEL", "info")),
		Output: cm.loadOutputFromName(getEnvString("SPL_LOG_OUTPUT", "stderr")),
	}

	// Agent Config
	cm.agentConfig = types.AgentConfig{
		Tool:                 cm.mcpServerConfig.Tool,
		Model:                cm.llmServiceConfig.Model,
		SystemPromptTemplate: cm.llmServiceConfig.SystemPromptTemplate,
		MaxTokens:            getEnvInt("SPL_CHAT_MAX_TOKENS", 0),
		CompactionStrategy:   getEnvString("SPL_CHAT_COMPACTION_STRATEGY", "delete-old"),
		MaxLLMIterations:     getEnvInt("SPL_CHAT_MAX_ITERATIONS", 25),
	}

	// If there are validation errors, return them
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// Helper function to find server indices from environment variables
func findServerIndices() []int {
	indices := make(map[int]bool)

	// Look for all environment variables starting with SPL_MCPS_
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) < 1 {
			continue
		}

		key := parts[0]

		// Only check for SPL_ prefixed variables
		if strings.HasPrefix(key, "SPL_MCPS_") {
			// Extract the index from the key (e.g., SPL_MCPS_0_ID -> 0)
			parts = strings.Split(key, "_")
			if len(parts) < 4 {
				continue
			}

			idx, err := strconv.Atoi(parts[2])
			if err != nil {
				continue
			}

			indices[idx] = true
		}
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

func (cm *Manager) GetAgentConfig() types.AgentConfig {
	return cm.agentConfig
}

// GetStringMap returns a string map value from the configuration.
func (cm *Manager) GetStringMap(key string) (map[string]string, bool) {
	// Implementation to be added if needed
	return nil, false
}

// loadOutputFromName loads an io.Writer based on the output name
func (cm *Manager) loadOutputFromName(name string) io.Writer {
	switch strings.ToLower(name) {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		// Treat anything else as a file path and try to open for appending
		file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			cm.logger.Warnf("Failed to open log file %s: %v, defaulting to stderr", name, err)
			return os.Stderr
		}
		cm.logger.Infof("Logging to file: %s", name)
		return file
	}
}

// loadLevelFromName loads a logrus.Level from a string level name
func (cm *Manager) loadLevelFromName(name string) logrus.Level {
	switch strings.ToLower(name) {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		cm.logger.Warnf("Invalid log level specified: %s, defaulting to info", name)
		return logrus.InfoLevel
	}
}
