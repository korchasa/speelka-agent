// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import (
	"os"
	"strconv"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/types"
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
	if os.Getenv("SPL_LOG_LEVEL") != "" {
		config.Runtime.Log.RawLevel = os.Getenv("SPL_LOG_LEVEL")
	}
	if os.Getenv("SPL_LOG_OUTPUT") != "" {
		config.Runtime.Log.RawOutput = os.Getenv("SPL_LOG_OUTPUT")
	}
	// Log format (text or json)
	if os.Getenv("SPL_LOG_FORMAT") != "" {
		config.Runtime.Log.RawFormat = os.Getenv("SPL_LOG_FORMAT")
	}

	// Runtime transport configuration
	if os.Getenv("SPL_TRANSPORTS_STDIO_ENABLED") != "" {
		enabled, err := strconv.ParseBool(os.Getenv("SPL_TRANSPORTS_STDIO_ENABLED"))
		if err == nil {
			config.Runtime.Transports.Stdio.Enabled = enabled
		}
	}
	if os.Getenv("SPL_TRANSPORTS_STDIO_BUFFER_SIZE") != "" {
		bufferSize, err := strconv.Atoi(os.Getenv("SPL_TRANSPORTS_STDIO_BUFFER_SIZE"))
		if err == nil {
			config.Runtime.Transports.Stdio.BufferSize = bufferSize
		}
	}
	if os.Getenv("SPL_TRANSPORTS_HTTP_ENABLED") != "" {
		enabled, err := strconv.ParseBool(os.Getenv("SPL_TRANSPORTS_HTTP_ENABLED"))
		if err == nil {
			config.Runtime.Transports.HTTP.Enabled = enabled
		}
	}
	if os.Getenv("SPL_TRANSPORTS_HTTP_HOST") != "" {
		config.Runtime.Transports.HTTP.Host = os.Getenv("SPL_TRANSPORTS_HTTP_HOST")
	}
	if os.Getenv("SPL_TRANSPORTS_HTTP_PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("SPL_TRANSPORTS_HTTP_PORT"))
		if err == nil {
			config.Runtime.Transports.HTTP.Port = port
		}
	}

	// Agent configuration
	if os.Getenv("SPL_AGENT_NAME") != "" {
		config.Agent.Name = os.Getenv("SPL_AGENT_NAME")
	}
	if os.Getenv("SPL_AGENT_VERSION") != "" {
		config.Agent.Version = os.Getenv("SPL_AGENT_VERSION")
	}

	// Tool configuration
	if os.Getenv("SPL_TOOL_NAME") != "" {
		config.Agent.Tool.Name = os.Getenv("SPL_TOOL_NAME")
	}
	if os.Getenv("SPL_TOOL_DESCRIPTION") != "" {
		config.Agent.Tool.Description = os.Getenv("SPL_TOOL_DESCRIPTION")
	}
	if os.Getenv("SPL_TOOL_ARGUMENT_NAME") != "" {
		config.Agent.Tool.ArgumentName = os.Getenv("SPL_TOOL_ARGUMENT_NAME")
	}
	if os.Getenv("SPL_TOOL_ARGUMENT_DESCRIPTION") != "" {
		config.Agent.Tool.ArgumentDescription = os.Getenv("SPL_TOOL_ARGUMENT_DESCRIPTION")
	}

	// LLM configuration
	if os.Getenv("SPL_LLM_PROVIDER") != "" {
		config.Agent.LLM.Provider = os.Getenv("SPL_LLM_PROVIDER")
	}
	if os.Getenv("SPL_LLM_API_KEY") != "" {
		config.Agent.LLM.APIKey = os.Getenv("SPL_LLM_API_KEY")
	}
	if os.Getenv("SPL_LLM_MODEL") != "" {
		config.Agent.LLM.Model = os.Getenv("SPL_LLM_MODEL")
	}
	if os.Getenv("SPL_LLM_PROMPT_TEMPLATE") != "" {
		config.Agent.LLM.PromptTemplate = os.Getenv("SPL_LLM_PROMPT_TEMPLATE")
	}

	// Optional LLM configuration with defaults
	if os.Getenv("SPL_LLM_MAX_TOKENS") != "" {
		maxTokens, err := strconv.Atoi(os.Getenv("SPL_LLM_MAX_TOKENS"))
		if err == nil {
			config.Agent.LLM.MaxTokens = maxTokens
			config.Agent.LLM.IsMaxTokensSet = true
		}
	}

	if os.Getenv("SPL_LLM_TEMPERATURE") != "" {
		temperature, err := strconv.ParseFloat(os.Getenv("SPL_LLM_TEMPERATURE"), 64)
		if err == nil {
			config.Agent.LLM.Temperature = temperature
			config.Agent.LLM.IsTemperatureSet = true
		}
	}

	// LLM Retry configuration
	if os.Getenv("SPL_LLM_RETRY_MAX_RETRIES") != "" {
		maxRetries, err := strconv.Atoi(os.Getenv("SPL_LLM_RETRY_MAX_RETRIES"))
		if err == nil {
			config.Agent.LLM.Retry.MaxRetries = maxRetries
		}
	}

	if os.Getenv("SPL_LLM_RETRY_INITIAL_BACKOFF") != "" {
		initialBackoff, err := strconv.ParseFloat(os.Getenv("SPL_LLM_RETRY_INITIAL_BACKOFF"), 64)
		if err == nil {
			config.Agent.LLM.Retry.InitialBackoff = initialBackoff
		}
	}

	if os.Getenv("SPL_LLM_RETRY_MAX_BACKOFF") != "" {
		maxBackoff, err := strconv.ParseFloat(os.Getenv("SPL_LLM_RETRY_MAX_BACKOFF"), 64)
		if err == nil {
			config.Agent.LLM.Retry.MaxBackoff = maxBackoff
		}
	}

	if os.Getenv("SPL_LLM_RETRY_BACKOFF_MULTIPLIER") != "" {
		backoffMultiplier, err := strconv.ParseFloat(os.Getenv("SPL_LLM_RETRY_BACKOFF_MULTIPLIER"), 64)
		if err == nil {
			config.Agent.LLM.Retry.BackoffMultiplier = backoffMultiplier
		}
	}

	// Chat configuration
	if os.Getenv("SPL_CHAT_MAX_ITERATIONS") != "" {
		maxIterations, err := strconv.Atoi(os.Getenv("SPL_CHAT_MAX_ITERATIONS"))
		if err == nil {
			config.Agent.Chat.MaxLLMIterations = maxIterations
		}
	}

	if os.Getenv("SPL_CHAT_MAX_TOKENS") != "" {
		maxTokens, err := strconv.Atoi(os.Getenv("SPL_CHAT_MAX_TOKENS"))
		if err == nil {
			config.Agent.Chat.MaxTokens = maxTokens
		}
	}

	// Connection retry configuration
	if os.Getenv("SPL_CONNECTIONS_RETRY_MAX_RETRIES") != "" {
		maxRetries, err := strconv.Atoi(os.Getenv("SPL_CONNECTIONS_RETRY_MAX_RETRIES"))
		if err == nil {
			config.Agent.Connections.Retry.MaxRetries = maxRetries
		}
	}

	if os.Getenv("SPL_CONNECTIONS_RETRY_INITIAL_BACKOFF") != "" {
		initialBackoff, err := strconv.ParseFloat(os.Getenv("SPL_CONNECTIONS_RETRY_INITIAL_BACKOFF"), 64)
		if err == nil {
			config.Agent.Connections.Retry.InitialBackoff = initialBackoff
		}
	}

	if os.Getenv("SPL_CONNECTIONS_RETRY_MAX_BACKOFF") != "" {
		maxBackoff, err := strconv.ParseFloat(os.Getenv("SPL_CONNECTIONS_RETRY_MAX_BACKOFF"), 64)
		if err == nil {
			config.Agent.Connections.Retry.MaxBackoff = maxBackoff
		}
	}

	if os.Getenv("SPL_CONNECTIONS_RETRY_BACKOFF_MULTIPLIER") != "" {
		backoffMultiplier, err := strconv.ParseFloat(os.Getenv("SPL_CONNECTIONS_RETRY_BACKOFF_MULTIPLIER"), 64)
		if err == nil {
			config.Agent.Connections.Retry.BackoffMultiplier = backoffMultiplier
		}
	}

	// MCP Servers configuration
	config.Agent.Connections.McpServers = make(map[string]types.MCPServerConnection)
	l.loadMCPServersFromEnv(config)

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
			Command:      command,
			Args:         args,
			Environment:  env,
			IncludeTools: parseEnvList(os.Getenv(prefix + "INCLUDE_TOOLS")),
			ExcludeTools: parseEnvList(os.Getenv(prefix + "EXCLUDE_TOOLS")),
		}
	}
}

// parseEnvList parses a comma- or space-separated environment variable into a string slice.
func parseEnvList(val string) []string {
	if val == "" {
		return nil
	}
	// Support both comma and space as separators
	var out []string
	for _, part := range strings.FieldsFunc(val, func(r rune) bool { return r == ',' || r == ' ' }) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
