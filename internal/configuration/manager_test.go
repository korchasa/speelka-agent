package configuration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	log "github.com/sirupsen/logrus"
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

	// Set environment variables for test
	for k, v := range env {
		os.Setenv(k, v)
	}

	// Restore environment after test
	defer func() {
		// Clear all environment variables that were set in the test
		for k := range env {
			os.Unsetenv(k)
		}

		// Restore original environment
		for k, v := range originalEnv {
			if _, present := env[k]; present {
				os.Setenv(k, v)
			}
		}
	}()

	// Run the test
	testFunc()
}

func TestLoadEnvironmentConfiguration(t *testing.T) {
	// Create a logger for testing
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	t.Run("basic configuration", func(t *testing.T) {
		env := map[string]string{
			"AGENT_NAME":                "test-agent",
			"AGENT_VERSION":             "1.0.0",
			"TOOL_NAME":                 "test-tool",
			"TOOL_DESCRIPTION":          "Test tool description",
			"TOOL_ARGUMENT_NAME":        "query",
			"TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"LLM_PROVIDER":              "openai",
			"LLM_MODEL":                 "gpt-4o",
			"LLM_API_KEY":               "test-api-key",
			"LLM_MAX_TOKENS":            "100",
			"LLM_TEMPERATURE":           "0.5",
			"LLM_PROMPT_TEMPLATE":       "Template with {{query}} and {{tools}} placeholders",
			"RUNTIME_LOG_LEVEL":         "info",
			"RUNTIME_LOG_OUTPUT":        "stdout",
			"RUNTIME_STDIO_ENABLED":     "true",
			"RUNTIME_STDIO_BUFFER_SIZE": "8192",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(logger)

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
		})
	})

	t.Run("MCP servers configuration", func(t *testing.T) {
		env := map[string]string{
			// Basic required config
			"AGENT_NAME":                "test-agent",
			"AGENT_VERSION":             "1.0.0",
			"TOOL_NAME":                 "test-tool",
			"TOOL_DESCRIPTION":          "Test tool description",
			"TOOL_ARGUMENT_NAME":        "query",
			"TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"LLM_PROVIDER":              "openai",
			"LLM_MODEL":                 "gpt-4o",
			"LLM_PROMPT_TEMPLATE":       "Template with {{query}} and {{tools}} placeholders",

			// MCP Server configs
			"MCPS_0_ID":      "time-server",
			"MCPS_0_COMMAND": "docker",
			"MCPS_0_ARGS":    "run -i --rm mcp/time",

			"MCPS_1_ID":      "filesystem-server",
			"MCPS_1_COMMAND": "mcp-filesystem-server",
			"MCPS_1_ARGS":    ".",

			// MCPS retry config
			"MSPS_RETRY_MAX_RETRIES":        "3",
			"MSPS_RETRY_INITIAL_BACKOFF":    "1.0",
			"MSPS_RETRY_MAX_BACKOFF":        "30.0",
			"MSPS_RETRY_BACKOFF_MULTIPLIER": "2.0",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(logger)

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

	t.Run("missing required config", func(t *testing.T) {
		env := map[string]string{
			// Missing AGENT_NAME, TOOL_NAME, and other required fields
			"AGENT_VERSION": "1.0.0",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(logger)

			// Load configuration - should fail validation
			err := cm.LoadConfiguration(context.Background())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "AGENT_NAME environment variable is required")
			assert.Contains(t, err.Error(), "TOOL_NAME environment variable is required")
			assert.Contains(t, err.Error(), "TOOL_DESCRIPTION environment variable is required")
			assert.Contains(t, err.Error(), "TOOL_ARGUMENT_NAME environment variable is required")
			assert.Contains(t, err.Error(), "TOOL_ARGUMENT_DESCRIPTION environment variable is required")
		})
	})

	t.Run("invalid prompt template", func(t *testing.T) {
		env := map[string]string{
			// Basic required config
			"AGENT_NAME":                "test-agent",
			"AGENT_VERSION":             "1.0.0",
			"TOOL_NAME":                 "test-tool",
			"TOOL_DESCRIPTION":          "Test tool description",
			"TOOL_ARGUMENT_NAME":        "query",
			"TOOL_ARGUMENT_DESCRIPTION": "Test query description",
			"LLM_PROVIDER":              "openai",
			"LLM_MODEL":                 "gpt-4o",
			// Missing {{tools}} placeholder
			"LLM_PROMPT_TEMPLATE": "Template with only {{query}} placeholder",
		}

		withEnvironment(t, env, func() {
			// Create a configuration manager
			cm := configuration.NewConfigurationManager(logger)

			// Load configuration - should fail validation
			err := cm.LoadConfiguration(context.Background())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "prompt template validation failed")
			assert.Contains(t, err.Error(), "missing required placeholder(s): tools")
		})
	})
}

func TestValidatePromptTemplate(t *testing.T) {
	// Create a logger for testing
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	// Create a configuration manager
	cm := configuration.NewConfigurationManager(logger)

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
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	// Create a configuration manager
	cm := configuration.NewConfigurationManager(logger)

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
