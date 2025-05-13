package configuration

import (
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvLoader_LoadConfiguration(t *testing.T) {
	// Setup
	t.Cleanup(func() {
		// Clear environment variables
		for _, envVar := range os.Environ() {
			key, _ := splitEnvVar(envVar)
			if strings.HasPrefix(key, "SPL_") {
				os.Unsetenv(key)
			}
		}
	})

	// Test scenario: No environment variables set
	t.Run("NoEnvironmentVariables", func(t *testing.T) {
		// Clear environment variables
		for _, envVar := range os.Environ() {
			key, _ := splitEnvVar(envVar)
			if strings.HasPrefix(key, "SPL_") {
				os.Unsetenv(key)
			}
		}

		loader := NewEnvLoader()
		config, err := loader.LoadConfiguration()

		// Should return nil config and nil error when no required env vars are set
		assert.NoError(t, err, "Error should be nil when no required env vars are set")
		assert.NotNil(t, config, "Config should not be nil when no required env vars are set")
	})

	// Test scenario: Only required environment variables set
	t.Run("OnlyRequiredVariables", func(t *testing.T) {
		// Clear environment variables
		for _, envVar := range os.Environ() {
			key, _ := splitEnvVar(envVar)
			if strings.HasPrefix(key, "SPL_") {
				os.Unsetenv(key)
			}
		}

		// Set required environment variables
		os.Setenv("SPL_AGENT_NAME", "test-agent")
		os.Setenv("SPL_TOOL_NAME", "test-tool")
		os.Setenv("SPL_TOOL_DESCRIPTION", "Tool for testing")
		os.Setenv("SPL_TOOL_ARGUMENT_NAME", "input")
		os.Setenv("SPL_TOOL_ARGUMENT_DESCRIPTION", "Input for the tool")
		os.Setenv("SPL_LLM_PROVIDER", "openai")
		os.Setenv("SPL_LLM_MODEL", "gpt-4")
		os.Setenv("SPL_LLM_API_KEY", "test-api-key")
		os.Setenv("SPL_LLM_PROMPT_TEMPLATE", "Process this {{input}}. Available tools: {{tools}}")

		loader := NewEnvLoader()
		config, err := loader.LoadConfiguration()

		require.NoError(t, err)
		require.NotNil(t, config, "Configuration should not be nil when required env vars are set")

		// Assert that required values were set
		assert.Equal(t, "test-agent", config.Agent.Name)
		assert.Equal(t, "test-tool", config.Agent.Tool.Name)
		assert.Equal(t, "Tool for testing", config.Agent.Tool.Description)
		assert.Equal(t, "input", config.Agent.Tool.ArgumentName)
		assert.Equal(t, "Input for the tool", config.Agent.Tool.ArgumentDescription)
		assert.Equal(t, "openai", config.Agent.LLM.Provider)
		assert.Equal(t, "gpt-4", config.Agent.LLM.Model)
		assert.Equal(t, "test-api-key", config.Agent.LLM.APIKey)
		assert.Equal(t, "Process this {{input}}. Available tools: {{tools}}", config.Agent.LLM.PromptTemplate)

		// Assert that optional values were set to defaults
		assert.Equal(t, "", config.Runtime.Log.RawLevel)
		assert.Equal(t, "", config.Runtime.Log.RawOutput)
	})

	// Test scenario: Override default values
	t.Run("OverrideDefaultValues", func(t *testing.T) {
		// Set required environment variables
		os.Setenv("SPL_AGENT_NAME", "test-agent")
		os.Setenv("SPL_TOOL_NAME", "test-tool")
		os.Setenv("SPL_TOOL_DESCRIPTION", "Tool for testing")
		os.Setenv("SPL_TOOL_ARGUMENT_NAME", "input")
		os.Setenv("SPL_TOOL_ARGUMENT_DESCRIPTION", "Input for the tool")
		os.Setenv("SPL_LLM_PROVIDER", "openai")
		os.Setenv("SPL_LLM_MODEL", "gpt-4")
		os.Setenv("SPL_LLM_API_KEY", "test-api-key")
		os.Setenv("SPL_LLM_PROMPT_TEMPLATE", "Process this {{input}}. Available tools: {{tools}}")

		// Override default values
		os.Setenv("SPL_LOG_LEVEL", "debug")
		os.Setenv("SPL_LOG_OUTPUT", "stdout")
		os.Setenv("SPL_LOG_FORMAT", "json")
		os.Setenv("SPL_LLM_MAX_TOKENS", "1000")
		os.Setenv("SPL_LLM_TEMPERATURE", "0.5")
		os.Setenv("SPL_LLM_RETRY_MAX_RETRIES", "5")
		os.Setenv("SPL_LLM_RETRY_INITIAL_BACKOFF", "2.0")
		os.Setenv("SPL_LLM_RETRY_MAX_BACKOFF", "60.0")
		os.Setenv("SPL_LLM_RETRY_BACKOFF_MULTIPLIER", "3.0")
		os.Setenv("SPL_CHAT_MAX_ITERATIONS", "50")
		os.Setenv("SPL_CHAT_MAX_TOKENS", "2000")

		loader := NewEnvLoader()
		config, err := loader.LoadConfiguration()

		require.NoError(t, err)
		require.NotNil(t, config, "Configuration should not be nil when required env vars are set")

		// Assert overridden values
		assert.Equal(t, "debug", config.Runtime.Log.RawLevel)
		assert.Equal(t, "stdout", config.Runtime.Log.RawOutput)
		assert.Equal(t, "json", config.Runtime.Log.RawFormat)
		// After Apply, check parsed fields
		config.Apply(config)
		assert.Equal(t, os.Stdout, config.Runtime.Log.Output)
		assert.Equal(t, logrus.DebugLevel, config.Runtime.Log.LogLevel)
		assert.Equal(t, 1000, config.Agent.LLM.MaxTokens)
		assert.Equal(t, 0.5, config.Agent.LLM.Temperature)
		assert.Equal(t, 5, config.Agent.LLM.Retry.MaxRetries)
		assert.Equal(t, 2.0, config.Agent.LLM.Retry.InitialBackoff)
		assert.Equal(t, 60.0, config.Agent.LLM.Retry.MaxBackoff)
		assert.Equal(t, 3.0, config.Agent.LLM.Retry.BackoffMultiplier)
		assert.Equal(t, 50, config.Agent.Chat.MaxLLMIterations)
		assert.Equal(t, 2000, config.Agent.Chat.MaxTokens)
	})

	// Test scenario: MCP Server configuration
	t.Run("ConfigureMCPServers", func(t *testing.T) {
		// Set required environment variables
		os.Setenv("SPL_AGENT_NAME", "test-agent")
		os.Setenv("SPL_TOOL_NAME", "test-tool")
		os.Setenv("SPL_TOOL_DESCRIPTION", "Tool for testing")
		os.Setenv("SPL_TOOL_ARGUMENT_NAME", "input")
		os.Setenv("SPL_TOOL_ARGUMENT_DESCRIPTION", "Input for the tool")
		os.Setenv("SPL_LLM_PROVIDER", "openai")
		os.Setenv("SPL_LLM_MODEL", "gpt-4")
		os.Setenv("SPL_LLM_API_KEY", "test-api-key")
		os.Setenv("SPL_LLM_PROMPT_TEMPLATE", "Process this {{input}}. Available tools: {{tools}}")

		// Configure MCP servers
		os.Setenv("SPL_MCPS_0_ID", "time")
		os.Setenv("SPL_MCPS_0_COMMAND", "docker")
		os.Setenv("SPL_MCPS_0_ARGS", "run -i --rm mcp/time")
		os.Setenv("SPL_MCPS_0_ENV_TEST_VAR", "test_value")

		os.Setenv("SPL_MCPS_1_ID", "filesystem")
		os.Setenv("SPL_MCPS_1_COMMAND", "mcp-filesystem-server")
		os.Setenv("SPL_MCPS_1_ARGS", "/path/to/directory")

		// Test: Only IncludeTools (comma-separated)
		os.Setenv("SPL_MCPS_0_INCLUDE_TOOLS", "foo,bar,baz")
		os.Unsetenv("SPL_MCPS_0_EXCLUDE_TOOLS")
		// Test: Only ExcludeTools (space-separated)
		os.Unsetenv("SPL_MCPS_1_INCLUDE_TOOLS")
		os.Setenv("SPL_MCPS_1_EXCLUDE_TOOLS", "qux quux")

		loader := NewEnvLoader()
		config, err := loader.LoadConfiguration()

		require.NoError(t, err)
		require.NotNil(t, config, "Configuration should not be nil when required env vars are set")

		// Assert MCP server configurations
		assert.Len(t, config.Agent.Connections.McpServers, 2)

		// Check first server (IncludeTools only)
		timeServer, exists := config.Agent.Connections.McpServers["time"]
		assert.True(t, exists, "Time server should exist")
		assert.Equal(t, "docker", timeServer.Command)
		assert.Equal(t, []string{"run", "-i", "--rm", "mcp/time"}, timeServer.Args)
		assert.Contains(t, timeServer.Environment, "TEST_VAR=test_value")
		assert.Equal(t, []string{"foo", "bar", "baz"}, timeServer.IncludeTools)
		assert.Nil(t, timeServer.ExcludeTools)

		// Check second server (ExcludeTools only)
		fsServer, exists := config.Agent.Connections.McpServers["filesystem"]
		assert.True(t, exists, "Filesystem server should exist")
		assert.Equal(t, "mcp-filesystem-server", fsServer.Command)
		assert.Equal(t, []string{"/path/to/directory"}, fsServer.Args)
		assert.Nil(t, fsServer.IncludeTools)
		assert.Equal(t, []string{"qux", "quux"}, fsServer.ExcludeTools)

		// Test: Both IncludeTools and ExcludeTools (space and comma)
		os.Setenv("SPL_MCPS_2_ID", "hybrid")
		os.Setenv("SPL_MCPS_2_COMMAND", "hybrid-server")
		os.Setenv("SPL_MCPS_2_INCLUDE_TOOLS", "alpha beta,gamma")
		os.Setenv("SPL_MCPS_2_EXCLUDE_TOOLS", "delta, epsilon zeta")
		loader2 := NewEnvLoader()
		config2, err2 := loader2.LoadConfiguration()
		require.NoError(t, err2)
		hybridServer, exists := config2.Agent.Connections.McpServers["hybrid"]
		assert.True(t, exists, "Hybrid server should exist")
		assert.Equal(t, []string{"alpha", "beta", "gamma"}, hybridServer.IncludeTools)
		assert.Equal(t, []string{"delta", "epsilon", "zeta"}, hybridServer.ExcludeTools)

		// Test: Neither IncludeTools nor ExcludeTools
		os.Setenv("SPL_MCPS_3_ID", "plain")
		os.Setenv("SPL_MCPS_3_COMMAND", "plain-server")
		os.Unsetenv("SPL_MCPS_3_INCLUDE_TOOLS")
		os.Unsetenv("SPL_MCPS_3_EXCLUDE_TOOLS")
		loader3 := NewEnvLoader()
		config3, err3 := loader3.LoadConfiguration()
		require.NoError(t, err3)
		plainServer, exists := config3.Agent.Connections.McpServers["plain"]
		assert.True(t, exists, "Plain server should exist")
		assert.Nil(t, plainServer.IncludeTools)
		assert.Nil(t, plainServer.ExcludeTools)
	})
}

// Helper function to split environment variables
func splitEnvVar(envVar string) (string, string) {
	for i := 0; i < len(envVar); i++ {
		if envVar[i] == '=' {
			return envVar[:i], envVar[i+1:]
		}
	}
	return envVar, ""
}
