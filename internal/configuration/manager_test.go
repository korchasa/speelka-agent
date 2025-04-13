package configuration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/logger"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/stretchr/testify/assert"
)

// Helper function to set environment variables for tests and clean them up after
func withEnvironment(t *testing.T, env map[string]string, testFunc func()) {
	// Save current environment to restore later
	originalEnv := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Clear any existing SPL_ environment variables to prevent interference between tests
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 && strings.HasPrefix(pair[0], "SPL_") {
			_ = os.Unsetenv(pair[0])
		}
	}

	// Set environment variables for test
	for k, v := range env {
		_ = os.Setenv(k, v)
	}

	// Restore environment after test
	defer func() {
		// Clear all environment variables that were set in the test
		for k := range env {
			_ = os.Unsetenv(k)
		}

		// Clear any remaining SPL_ environment variables
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			if len(pair) == 2 && strings.HasPrefix(pair[0], "SPL_") {
				_ = os.Unsetenv(pair[0])
			}
		}

		// Restore original environment
		for k, v := range originalEnv {
			if _, present := env[k]; present {
				_ = os.Setenv(k, v)
			}
		}
	}()

	// Run the test
	testFunc()
}

func TestLoadEnvironmentConfiguration(t *testing.T) {
	// Create a logger for testing
	log := logger.NewLogger()

	t.Run("basic configuration", func(t *testing.T) {
		env := map[string]string{
			"SPL_AGENT_NAME":                "test-agent",
			"SPL_AGENT_VERSION":             "1.0.0",
			"SPL_TOOL_NAME":                 "test-tool",
			"SPL_TOOL_DESCRIPTION":          "Test tool description",
			"SPL_TOOL_ARGUMENT_NAME":        "query",
			"SPL_TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"SPL_LLM_PROVIDER":              "openai",
			"SPL_LLM_MODEL":                 "gpt-4o",
			"SPL_LLM_API_KEY":               "test-api-key",
			"SPL_LLM_MAX_TOKENS":            "100",
			"SPL_LLM_TEMPERATURE":           "0.5",
			"SPL_LLM_PROMPT_TEMPLATE":       "Template with {{query}} and {{tools}} placeholders",
			"SPL_LOG_LEVEL":                 "info",
			"SPL_LOG_OUTPUT":                "stdout",
			"SPL_RUNTIME_STDIO_ENABLED":     "true",
			"SPL_RUNTIME_STDIO_BUFFER_SIZE": "8192",
			"SPL_CHAT_MAX_TOKENS":           "1000",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(log)

			// Load configuration
			err := cm.LoadConfiguration(context.Background())
			assert.NoError(t, err)

			// Validate configuration
			mcpConfig := cm.GetMCPServerConfig()
			assert.Equal(t, "test-agent", mcpConfig.Name)
			assert.Equal(t, "1.0.0", mcpConfig.Version)
			assert.Equal(t, "test-tool", mcpConfig.Tool.Name)
			assert.Equal(t, "Test tool description", mcpConfig.Tool.Description)
			assert.Equal(t, "query", mcpConfig.Tool.ArgumentName)
			assert.Equal(t, "Test query description", mcpConfig.Tool.ArgumentDescription)
			assert.True(t, mcpConfig.Stdio.Enabled)
			assert.Equal(t, 8192, mcpConfig.Stdio.BufferSize)

			llmConfig := cm.GetLLMConfig()
			assert.Equal(t, "openai", llmConfig.Provider)
			assert.Equal(t, "gpt-4o", llmConfig.Model)
			assert.Equal(t, "test-api-key", llmConfig.APIKey)
			assert.Equal(t, 100, llmConfig.MaxTokens)
			assert.Equal(t, 0.5, llmConfig.Temperature)
			assert.Equal(t, "Template with {{query}} and {{tools}} placeholders", llmConfig.SystemPromptTemplate)

			// Verify agent configuration
			agentConfig := cm.GetAgentConfig()
			assert.Equal(t, 1000, agentConfig.MaxTokens)
			assert.Equal(t, "delete-old", agentConfig.CompactionStrategy)
			assert.Equal(t, "test-tool", agentConfig.Tool.Name)
			assert.Equal(t, "Test tool description", agentConfig.Tool.Description)
			assert.Equal(t, "query", agentConfig.Tool.ArgumentName)
			assert.Equal(t, "Test query description", agentConfig.Tool.ArgumentDescription)
			assert.Equal(t, "gpt-4o", agentConfig.Model)
			assert.Equal(t, "Template with {{query}} and {{tools}} placeholders", agentConfig.SystemPromptTemplate)
			assert.Equal(t, 1000, agentConfig.MaxTokens)
			assert.Equal(t, "delete-old", agentConfig.CompactionStrategy)
		})
	})

	t.Run("default chat configuration", func(t *testing.T) {
		env := map[string]string{
			"SPL_AGENT_NAME":                "test-agent",
			"SPL_AGENT_VERSION":             "1.0.0",
			"SPL_TOOL_NAME":                 "test-tool",
			"SPL_TOOL_DESCRIPTION":          "Test tool description",
			"SPL_TOOL_ARGUMENT_NAME":        "query",
			"SPL_TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"SPL_LLM_PROVIDER":              "openai",
			"SPL_LLM_MODEL":                 "gpt-4o",
			"SPL_LLM_API_KEY":               "test-api-key",
			"SPL_LLM_PROMPT_TEMPLATE":       "Template with {{query}} and {{tools}} placeholders",
			// Deliberately not setting SPL_CHAT_MAX_TOKENS to test default value
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(log)

			// Load configuration
			err := cm.LoadConfiguration(context.Background())
			assert.NoError(t, err)

			// Verify chat configuration default values through agent config
			agentConfig := cm.GetAgentConfig()
			assert.Equal(t, 0, agentConfig.MaxTokens, "Default value for SPL_CHAT_MAX_TOKENS should be 0")
			assert.Equal(t, "delete-old", agentConfig.CompactionStrategy)
		})
	})

	t.Run("MCP servers configuration", func(t *testing.T) {
		env := map[string]string{
			// Basic required config
			"SPL_AGENT_NAME":                "test-agent",
			"SPL_AGENT_VERSION":             "1.0.0",
			"SPL_TOOL_NAME":                 "test-tool",
			"SPL_TOOL_DESCRIPTION":          "Test tool description",
			"SPL_TOOL_ARGUMENT_NAME":        "query",
			"SPL_TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"SPL_LLM_PROVIDER":              "openai",
			"SPL_LLM_MODEL":                 "gpt-4o",
			"SPL_LLM_API_KEY":               "test-api-key",
			"SPL_LLM_PROMPT_TEMPLATE":       "Template with {{query}} and {{tools}} placeholders",

			// MCP Server configs
			"SPL_MCPS_0_ID":      "time-server",
			"SPL_MCPS_0_COMMAND": "docker",
			"SPL_MCPS_0_ARGS":    "run -i --rm mcp/time",

			"SPL_MCPS_1_ID":      "filesystem-server",
			"SPL_MCPS_1_COMMAND": "mcp-filesystem-server",
			"SPL_MCPS_1_ARGS":    ".",

			// MCPS retry config
			"SPL_MSPS_RETRY_MAX_RETRIES":        "3",
			"SPL_MSPS_RETRY_INITIAL_BACKOFF":    "1.0",
			"SPL_MSPS_RETRY_MAX_BACKOFF":        "30.0",
			"SPL_MSPS_RETRY_BACKOFF_MULTIPLIER": "2.0",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(log)

			// Load configuration
			err := cm.LoadConfiguration(context.Background())
			assert.NoError(t, err)

			// Validate MCP servers configuration
			mcpConnConfig := cm.GetMCPConnectorConfig()

			// Check servers
			assert.Len(t, mcpConnConfig.McpServers, 2)

			// Check first server
			timeServer, exists := mcpConnConfig.McpServers["time-server"]
			assert.True(t, exists)
			assert.Equal(t, "docker", timeServer.Command)
			assert.Equal(t, []string{"run", "-i", "--rm", "mcp/time"}, timeServer.Args)

			// Check second server
			fsServer, exists := mcpConnConfig.McpServers["filesystem-server"]
			assert.True(t, exists)
			assert.Equal(t, "mcp-filesystem-server", fsServer.Command)
			assert.Equal(t, []string{"."}, fsServer.Args)

			// Check retry config
			assert.Equal(t, 3, mcpConnConfig.RetryConfig.MaxRetries)
			assert.Equal(t, 1.0, mcpConnConfig.RetryConfig.InitialBackoff)
			assert.Equal(t, 30.0, mcpConnConfig.RetryConfig.MaxBackoff)
			assert.Equal(t, 2.0, mcpConnConfig.RetryConfig.BackoffMultiplier)
		})
	})

	t.Run("MCP server environment variables", func(t *testing.T) {
		env := map[string]string{
			// Basic required config
			"SPL_AGENT_NAME":                "test-agent",
			"SPL_AGENT_VERSION":             "1.0.0",
			"SPL_TOOL_NAME":                 "test-tool",
			"SPL_TOOL_DESCRIPTION":          "Test tool description",
			"SPL_TOOL_ARGUMENT_NAME":        "query",
			"SPL_TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"SPL_LLM_PROVIDER":              "openai",
			"SPL_LLM_MODEL":                 "gpt-4o",
			"SPL_LLM_API_KEY":               "test-api-key",
			"SPL_LLM_PROMPT_TEMPLATE":       "Template with {{query}} and {{tools}} placeholders",

			// MCP Server config with environment variables
			"SPL_MCPS_0_ID":           "fetcher",
			"SPL_MCPS_0_COMMAND":      "npx",
			"SPL_MCPS_0_ARGS":         "-y fetcher-mcp",
			"SPL_MCPS_0_ENV_NODE_ENV": "production",
			"SPL_MCPS_0_ENV_DEBUG":    "true",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(log)

			// Load configuration
			err := cm.LoadConfiguration(context.Background())
			assert.NoError(t, err)

			// Validate MCP servers configuration
			mcpConnConfig := cm.GetMCPConnectorConfig()

			// Check servers
			assert.Len(t, mcpConnConfig.McpServers, 1)

			// Check environment variables
			server, exists := mcpConnConfig.McpServers["fetcher"]
			assert.True(t, exists)
			assert.Equal(t, "npx", server.Command)
			assert.Equal(t, []string{"-y", "fetcher-mcp"}, server.Args)

			// The test should fail because environment variables aren't being loaded
			assert.Contains(t, server.Environment, "NODE_ENV=production")
			assert.Contains(t, server.Environment, "DEBUG=true")
		})
	})

	t.Run("missing required config", func(t *testing.T) {
		env := map[string]string{
			// Missing SPL_AGENT_NAME, SPL_TOOL_NAME, and other required fields
			"SPL_AGENT_VERSION": "1.0.0",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(log)

			// Load configuration - should fail validation
			err := cm.LoadConfiguration(context.Background())
			assert.Error(t, err)

			// Check for MCP Server Config validation errors
			assert.Contains(t, err.Error(), "SPL_AGENT_NAME environment variable is required")
			assert.Contains(t, err.Error(), "SPL_TOOL_NAME environment variable is required")
			assert.Contains(t, err.Error(), "SPL_TOOL_DESCRIPTION environment variable is required")
			assert.Contains(t, err.Error(), "SPL_TOOL_ARGUMENT_NAME environment variable is required")
			assert.Contains(t, err.Error(), "SPL_TOOL_ARGUMENT_DESCRIPTION environment variable is required")

			// Check for LLM Service Config validation errors
			assert.Contains(t, err.Error(), "SPL_LLM_API_KEY environment variable is required")
			assert.Contains(t, err.Error(), "SPL_LLM_PROVIDER environment variable is required")
			assert.Contains(t, err.Error(), "SPL_LLM_MODEL environment variable is required")
			assert.Contains(t, err.Error(), "SPL_LLM_PROMPT_TEMPLATE environment variable is required")
		})
	})

	t.Run("invalid prompt template", func(t *testing.T) {
		env := map[string]string{
			// Basic required config
			"SPL_AGENT_NAME":                "test-agent",
			"SPL_AGENT_VERSION":             "1.0.0",
			"SPL_TOOL_NAME":                 "test-tool",
			"SPL_TOOL_DESCRIPTION":          "Test tool description",
			"SPL_TOOL_ARGUMENT_NAME":        "query",
			"SPL_TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"SPL_LLM_PROVIDER":              "openai",
			"SPL_LLM_MODEL":                 "gpt-4o",
			"SPL_LLM_API_KEY":               "test-api-key",
			// Missing {{tools}} placeholder
			"SPL_LLM_PROMPT_TEMPLATE": "Template with only {{query}} placeholder",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(log)

			// Load configuration - should fail prompt template validation
			err := cm.LoadConfiguration(context.Background())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid prompt template")
			assert.Contains(t, err.Error(), "missing required placeholder")
		})
	})
}

func TestValidatePromptTemplate(t *testing.T) {
	// Create a logger for testing
	log := logger.NewLogger()

	// Create a configuration manager
	cm := configuration.NewConfigurationManager(log)

	t.Run("valid template with all placeholders", func(t *testing.T) {
		// Test a template with both required placeholders
		template := `This is a template with {{query}} and {{tools}} placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("valid template with additional placeholders", func(t *testing.T) {
		// Test a template with required and additional placeholders
		template := `Template with {{query}}, {{tools}}, and {{extra}} placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("valid template with whitespace in placeholders", func(t *testing.T) {
		// Test a template with whitespace in the placeholder syntax
		template := `Template with {{ query }} and {{ tools }} placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("invalid template missing query placeholder", func(t *testing.T) {
		// Test a template missing the query placeholder
		template := `Template with only {{tools}} placeholder`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required placeholder(s): query")
	})

	t.Run("invalid template missing tools placeholder", func(t *testing.T) {
		// Test a template missing the tools placeholder
		template := `Template with only {{query}} placeholder`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required placeholder(s): tools")
	})

	t.Run("invalid empty template", func(t *testing.T) {
		// Test an empty template
		template := ``
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("invalid template with no placeholders", func(t *testing.T) {
		// Test a template with no placeholders
		template := `This is a template without any placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
	})

	t.Run("different argument name", func(t *testing.T) {
		// Test with a different argument name than "query"
		template := `Template with {{input}} and {{tools}} placeholders`
		err := cm.TestValidatePromptTemplate(template, "input")
		assert.NoError(t, err)

		// Should fail when looking for a different argument name
		err = cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required placeholder(s): query")
	})
}

func TestExtractPlaceholders(t *testing.T) {
	// Create a logger for testing
	log := logger.NewLogger()

	// Create a configuration manager
	cm := configuration.NewConfigurationManager(log)

	t.Run("extract multiple placeholders", func(t *testing.T) {
		template := `This is a {{test}} template with {{multiple}} placeholders including {{tools}}`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"test", "multiple", "tools"}, placeholders)
	})

	t.Run("extract placeholders with whitespace", func(t *testing.T) {
		template := `This has {{ spaced }} placeholders and {{unspaced}} ones`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"spaced", "unspaced"}, placeholders)
	})

	t.Run("handle no placeholders", func(t *testing.T) {
		template := `This template has no placeholders`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.Empty(t, placeholders)
	})

	t.Run("handle empty template", func(t *testing.T) {
		template := ``
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.Empty(t, placeholders)
	})

	t.Run("handle complex nested content", func(t *testing.T) {
		template := `Complex template with {{placeholder}} and code snippets like if (x == y) { return true; }`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"placeholder"}, placeholders)
	})

	t.Run("extract placeholder with numbers and underscores", func(t *testing.T) {
		template := `Template with {{place_holder_123}} containing numbers and underscores`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"place_holder_123"}, placeholders)
	})
}
