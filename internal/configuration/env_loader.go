// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import (
	"os"
	"strconv"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// EnvLoader implements the LoaderSpec interface for loading configuration from environment variables.
type EnvLoader struct{}

// NewEnvLoader creates a new EnvLoader.
func NewEnvLoader() *EnvLoader {
	return &EnvLoader{}
}

// LoadConfiguration loads configuration from environment variables with SPL_ prefix.
func (l *EnvLoader) LoadConfiguration() (*types.Configuration, error) {
	// Initialize a new configuration
	config := types.NewConfiguration()

	// Parse environment variables
	// Runtime configuration
	config.Runtime.Log.RawLevel = getEnvWithDefault("SPL_LOG_LEVEL", "info")
	config.Runtime.Log.Output = getEnvWithDefault("SPL_LOG_OUTPUT", "stderr")

	// Runtime transport configuration
	if enabled, err := strconv.ParseBool(getEnvWithDefault("SPL_TRANSPORTS_STDIO_ENABLED", "true")); err == nil {
		config.Runtime.Transports.Stdio.Enabled = enabled
	}
	if bufferSize, err := strconv.Atoi(getEnvWithDefault("SPL_TRANSPORTS_STDIO_BUFFER_SIZE", "8192")); err == nil {
		config.Runtime.Transports.Stdio.BufferSize = bufferSize
	}
	if enabled, err := strconv.ParseBool(getEnvWithDefault("SPL_TRANSPORTS_HTTP_ENABLED", "false")); err == nil {
		config.Runtime.Transports.HTTP.Enabled = enabled
	}
	config.Runtime.Transports.HTTP.Host = getEnvWithDefault("SPL_TRANSPORTS_HTTP_HOST", "localhost")
	if port, err := strconv.Atoi(getEnvWithDefault("SPL_TRANSPORTS_HTTP_PORT", "3000")); err == nil {
		config.Runtime.Transports.HTTP.Port = port
	}

	// Agent configuration
	config.Agent.Name = os.Getenv("SPL_AGENT_NAME")
	config.Agent.Version = getEnvWithDefault("SPL_AGENT_VERSION", "1.0.0")

	// Tool configuration
	config.Agent.Tool.Name = os.Getenv("SPL_TOOL_NAME")
	config.Agent.Tool.Description = os.Getenv("SPL_TOOL_DESCRIPTION")
	config.Agent.Tool.ArgumentName = os.Getenv("SPL_TOOL_ARGUMENT_NAME")
	config.Agent.Tool.ArgumentDescription = os.Getenv("SPL_TOOL_ARGUMENT_DESCRIPTION")

	// LLM configuration
	config.Agent.LLM.Provider = os.Getenv("SPL_LLM_PROVIDER")
	config.Agent.LLM.APIKey = os.Getenv("SPL_LLM_API_KEY")
	config.Agent.LLM.Model = os.Getenv("SPL_LLM_MODEL")
	config.Agent.LLM.PromptTemplate = os.Getenv("SPL_LLM_PROMPT_TEMPLATE")

	// Optional LLM configuration with defaults
	if maxTokens, err := strconv.Atoi(getEnvWithDefault("SPL_LLM_MAX_TOKENS", "0")); err == nil {
		config.Agent.LLM.MaxTokens = maxTokens
		config.Agent.LLM.IsMaxTokensSet = true
	}

	if temperature, err := strconv.ParseFloat(getEnvWithDefault("SPL_LLM_TEMPERATURE", "0.7"), 64); err == nil {
		config.Agent.LLM.Temperature = temperature
		config.Agent.LLM.IsTemperatureSet = true
	}

	// LLM Retry configuration
	if maxRetries, err := strconv.Atoi(getEnvWithDefault("SPL_LLM_RETRY_MAX_RETRIES", "3")); err == nil {
		config.Agent.LLM.Retry.MaxRetries = maxRetries
	}

	if initialBackoff, err := strconv.ParseFloat(getEnvWithDefault("SPL_LLM_RETRY_INITIAL_BACKOFF", "1.0"), 64); err == nil {
		config.Agent.LLM.Retry.InitialBackoff = initialBackoff
	}

	if maxBackoff, err := strconv.ParseFloat(getEnvWithDefault("SPL_LLM_RETRY_MAX_BACKOFF", "30.0"), 64); err == nil {
		config.Agent.LLM.Retry.MaxBackoff = maxBackoff
	}

	if backoffMultiplier, err := strconv.ParseFloat(getEnvWithDefault("SPL_LLM_RETRY_BACKOFF_MULTIPLIER", "2.0"), 64); err == nil {
		config.Agent.LLM.Retry.BackoffMultiplier = backoffMultiplier
	}

	// Chat configuration
	if maxIterations, err := strconv.Atoi(getEnvWithDefault("SPL_CHAT_MAX_ITERATIONS", "25")); err == nil {
		config.Agent.Chat.MaxLLMIterations = maxIterations
	}

	if maxTokens, err := strconv.Atoi(getEnvWithDefault("SPL_CHAT_MAX_TOKENS", "0")); err == nil {
		config.Agent.Chat.MaxTokens = maxTokens
	}

	config.Agent.Chat.CompactionStrategy = getEnvWithDefault("SPL_CHAT_COMPACTION_STRATEGY", "delete-old")

	// Connection retry configuration
	if maxRetries, err := strconv.Atoi(getEnvWithDefault("SPL_CONNECTIONS_RETRY_MAX_RETRIES", "3")); err == nil {
		config.Agent.Connections.Retry.MaxRetries = maxRetries
	}

	if initialBackoff, err := strconv.ParseFloat(getEnvWithDefault("SPL_CONNECTIONS_RETRY_INITIAL_BACKOFF", "1.0"), 64); err == nil {
		config.Agent.Connections.Retry.InitialBackoff = initialBackoff
	}

	if maxBackoff, err := strconv.ParseFloat(getEnvWithDefault("SPL_CONNECTIONS_RETRY_MAX_BACKOFF", "30.0"), 64); err == nil {
		config.Agent.Connections.Retry.MaxBackoff = maxBackoff
	}

	if backoffMultiplier, err := strconv.ParseFloat(getEnvWithDefault("SPL_CONNECTIONS_RETRY_BACKOFF_MULTIPLIER", "2.0"), 64); err == nil {
		config.Agent.Connections.Retry.BackoffMultiplier = backoffMultiplier
	}

	// MCP Servers configuration
	config.Agent.Connections.McpServers = make(map[string]types.MCPServerConnection)
	l.loadMCPServersFromEnv(config)

	// If no required values are set, return nil to indicate no config from environment
	if !l.hasRequiredValues(config) {
		return nil, nil
	}

	// Set log level based on the provided string value
	logLevel, err := logrus.ParseLevel(config.Runtime.Log.RawLevel)
	if err == nil {
		config.Runtime.Log.LogLevel = logLevel
	} else {
		config.Runtime.Log.LogLevel = logrus.InfoLevel
	}

	return config, nil
}

// loadMCPServersFromEnv loads MCP server configurations from environment variables
func (l *EnvLoader) loadMCPServersFromEnv(config *types.Configuration) {
	// Check for MCP servers defined with indexed environment variables
	for i := 0; ; i++ {
		prefix := "SPL_MCPS_" + strconv.Itoa(i) + "_"

		id := os.Getenv(prefix + "ID")
		if id == "" {
			break // No more servers defined
		}

		command := os.Getenv(prefix + "COMMAND")
		argsStr := os.Getenv(prefix + "ARGS")

		var args []string
		if argsStr != "" {
			args = strings.Fields(argsStr)
		}

		// Create environment variable list for this server
		var env []string
		envPrefix := prefix + "ENV_"
		for _, pair := range os.Environ() {
			if strings.HasPrefix(pair, envPrefix) {
				env = append(env, strings.TrimPrefix(pair, envPrefix))
			}
		}

		// Add the server to the config
		config.Agent.Connections.McpServers[id] = types.MCPServerConnection{
			Command:     command,
			Args:        args,
			Environment: env,
		}
	}
}

// hasRequiredValues checks if the required configuration values are set
func (l *EnvLoader) hasRequiredValues(config *types.Configuration) bool {
	// Check if essential fields are provided
	return config.Agent.Name != "" &&
		config.Agent.Tool.Name != "" &&
		config.Agent.Tool.Description != "" &&
		config.Agent.Tool.ArgumentName != "" &&
		config.Agent.Tool.ArgumentDescription != "" &&
		config.Agent.LLM.Provider != "" &&
		config.Agent.LLM.Model != "" &&
		config.Agent.LLM.APIKey != "" &&
		config.Agent.LLM.PromptTemplate != ""
}

// getEnvWithDefault returns the value of the environment variable or the default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
