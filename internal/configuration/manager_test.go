package configuration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// mockLogger implements the LoggerSpec interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{})                 {}
func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Info(args ...interface{})                  {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Warn(args ...interface{})                  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{})                 {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}
func (m *mockLogger) Fatal(args ...interface{})                 {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) {}
func (m *mockLogger) Panic(args ...interface{})                 {}
func (m *mockLogger) Panicf(format string, args ...interface{}) {}
func (m *mockLogger) SetMCPServer(mcpServer interface{})        {}
func (m *mockLogger) WithField(key string, value interface{}) types.LogEntrySpec {
	return m
}
func (m *mockLogger) WithFields(fields logrus.Fields) types.LogEntrySpec {
	return m
}

// SetLevel method with the correct signature
func (m *mockLogger) SetLevel(level logrus.Level) {}

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

func TestLoadConfigurationFromYAMLFile(t *testing.T) {
	// Create a temporary YAML file
	content := `
agent:
  name: "test-agent"
  version: "1.0.0"
  tool:
    name: "test-tool"
    description: "Test tool description"
    argument_name: "input"
    argument_description: "Test argument description"
  llm:
    provider: "openai"
    api_key: "test-api-key"
    model: "gpt-4"
    max_tokens: 100
    temperature: 0.5
    prompt_template: "You are a helpful assistant. User query: {{input}} Available tools: {{tools}}"
    retry:
      max_retries: 3
      initial_backoff: 1.0
      max_backoff: 30.0
      backoff_multiplier: 2.0
  connections:
    mcpServers:
      test-server:
        command: "test-command"
        args: ["arg1", "arg2"]
        environment:
          ENV_VAR1: "value1"
          ENV_VAR2: "value2"
    retry:
      max_retries: 2
      initial_backoff: 1.5
      max_backoff: 20.0
      backoff_multiplier: 1.5
  chat:
    max_tokens: 200
    compaction_strategy: "delete-old"
runtime:
  log:
    level: "debug"
    output: "stdout"
  transports:
    stdio:
      enabled: true
      buffer_size: 4096
    http:
      enabled: false
      host: "localhost"
      port: 8080
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create configuration manager with mock logger
	mockLogger := &mockLogger{}
	cm := configuration.NewConfigurationManager(mockLogger)

	// Load configuration from YAML file
	err = cm.LoadConfigurationFromFile(tmpfile.Name())
	assert.NoError(t, err, "Loading configuration from YAML file should not return an error")

	// Verify loaded configuration
	mcpServerConfig := cm.GetMCPServerConfig()
	assert.Equal(t, "test-agent", mcpServerConfig.Name, "Agent name should be loaded correctly")
	assert.Equal(t, "1.0.0", mcpServerConfig.Version, "Agent version should be loaded correctly")
	assert.Equal(t, "test-tool", mcpServerConfig.Tool.Name, "Tool name should be loaded correctly")
	assert.Equal(t, "Test tool description", mcpServerConfig.Tool.Description, "Tool description should be loaded correctly")
	assert.Equal(t, "input", mcpServerConfig.Tool.ArgumentName, "Tool argument name should be loaded correctly")
	assert.Equal(t, "Test argument description", mcpServerConfig.Tool.ArgumentDescription, "Tool argument description should be loaded correctly")
	assert.Equal(t, 8080, mcpServerConfig.HTTP.Port, "HTTP port should be loaded correctly")

	llmConfig := cm.GetLLMConfig()
	assert.Equal(t, "openai", llmConfig.Provider, "LLM provider should be loaded correctly")
	assert.Equal(t, "test-api-key", llmConfig.APIKey, "LLM API key should be loaded correctly")
	assert.Equal(t, "gpt-4", llmConfig.Model, "LLM model should be loaded correctly")
	assert.Equal(t, 100, llmConfig.MaxTokens, "LLM max tokens should be loaded correctly")
	assert.Equal(t, 0.5, llmConfig.Temperature, "LLM temperature should be loaded correctly")
	assert.Equal(t, "You are a helpful assistant. User query: {{input}} Available tools: {{tools}}", llmConfig.SystemPromptTemplate, "LLM prompt template should be loaded correctly")
	assert.Equal(t, 3, llmConfig.RetryConfig.MaxRetries, "LLM retry max retries should be loaded correctly")

	agentConfig := cm.GetAgentConfig()
	assert.Equal(t, 200, agentConfig.MaxTokens, "Agent max tokens should be loaded correctly")
	assert.Equal(t, "delete-old", agentConfig.CompactionStrategy, "Agent compaction strategy should be loaded correctly")

	mcpConnectorConfig := cm.GetMCPConnectorConfig()
	assert.Equal(t, 2, mcpConnectorConfig.RetryConfig.MaxRetries, "MCP connector retry max retries should be loaded correctly")
	assert.Equal(t, 1.5, mcpConnectorConfig.RetryConfig.InitialBackoff, "MCP connector retry initial backoff should be loaded correctly")
	assert.Contains(t, mcpConnectorConfig.McpServers, "test-server", "MCP connector servers should contain test-server")
	assert.Equal(t, "test-command", mcpConnectorConfig.McpServers["test-server"].Command, "MCP connector server command should be loaded correctly")
	assert.Len(t, mcpConnectorConfig.McpServers["test-server"].Args, 2, "MCP connector server args should have 2 items")
	assert.Contains(t, mcpConnectorConfig.McpServers["test-server"].Environment, "ENV_VAR1=value1", "MCP connector server environment should contain ENV_VAR1=value1")
}

func TestLoadConfigurationFromJSONFile(t *testing.T) {
	// Create a temporary JSON file
	content := `{
  "agent": {
    "name": "test-agent-json",
    "version": "1.0.0",
    "tool": {
      "name": "test-tool-json",
      "description": "Test tool description JSON",
      "argument_name": "input",
      "argument_description": "Test argument description JSON"
    },
    "llm": {
      "provider": "anthropic",
      "api_key": "test-api-key-json",
      "model": "claude-3-opus",
      "max_tokens": 150,
      "temperature": 0.7,
      "prompt_template": "You are a helpful JSON assistant. User query: {{input}} Available tools: {{tools}}",
      "retry": {
        "max_retries": 4,
        "initial_backoff": 2.0,
        "max_backoff": 40.0,
        "backoff_multiplier": 2.5
      }
    },
    "connections": {
      "mcpServers": {
        "test-server-json": {
          "command": "test-command-json",
          "args": ["arg1-json", "arg2-json"],
          "environment": {
            "ENV_VAR1_JSON": "value1-json",
            "ENV_VAR2_JSON": "value2-json"
          }
        }
      },
      "retry": {
        "max_retries": 3,
        "initial_backoff": 2.5,
        "max_backoff": 25.0,
        "backoff_multiplier": 2.0
      }
    },
    "chat": {
      "max_tokens": 300,
      "compaction_strategy": "delete-middle"
    }
  },
  "runtime": {
    "log": {
      "level": "info",
      "output": "stderr"
    },
    "transports": {
      "stdio": {
        "enabled": false,
        "buffer_size": 5000
      },
      "http": {
        "enabled": true,
        "host": "0.0.0.0",
        "port": 9090
      }
    }
  }
}`
	tmpfile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create configuration manager with mock logger
	mockLogger := &mockLogger{}
	cm := configuration.NewConfigurationManager(mockLogger)

	// Load configuration from JSON file
	err = cm.LoadConfigurationFromFile(tmpfile.Name())
	assert.NoError(t, err, "Loading configuration from JSON file should not return an error")

	// Verify loaded configuration
	mcpServerConfig := cm.GetMCPServerConfig()
	assert.Equal(t, "test-agent-json", mcpServerConfig.Name, "Agent name should be loaded correctly")
	assert.Equal(t, "test-tool-json", mcpServerConfig.Tool.Name, "Tool name should be loaded correctly")
	assert.Equal(t, "Test tool description JSON", mcpServerConfig.Tool.Description, "Tool description should be loaded correctly")
	assert.Equal(t, 9090, mcpServerConfig.HTTP.Port, "HTTP port should be loaded correctly")
	assert.True(t, mcpServerConfig.HTTP.Enabled, "HTTP should be enabled")
	assert.False(t, mcpServerConfig.Stdio.Enabled, "Stdio should be disabled")

	llmConfig := cm.GetLLMConfig()
	assert.Equal(t, "anthropic", llmConfig.Provider, "LLM provider should be loaded correctly")
	assert.Equal(t, "claude-3-opus", llmConfig.Model, "LLM model should be loaded correctly")
	assert.Equal(t, 150, llmConfig.MaxTokens, "LLM max tokens should be loaded correctly")
	assert.Equal(t, 0.7, llmConfig.Temperature, "LLM temperature should be loaded correctly")
	assert.Equal(t, 4, llmConfig.RetryConfig.MaxRetries, "LLM retry max retries should be loaded correctly")

	agentConfig := cm.GetAgentConfig()
	assert.Equal(t, 300, agentConfig.MaxTokens, "Agent max tokens should be loaded correctly")
	assert.Equal(t, "delete-middle", agentConfig.CompactionStrategy, "Agent compaction strategy should be loaded correctly")

	mcpConnectorConfig := cm.GetMCPConnectorConfig()
	assert.Equal(t, 3, mcpConnectorConfig.RetryConfig.MaxRetries, "MCP connector retry max retries should be loaded correctly")
	assert.Contains(t, mcpConnectorConfig.McpServers, "test-server-json", "MCP connector servers should contain test-server-json")
	assert.Equal(t, "test-command-json", mcpConnectorConfig.McpServers["test-server-json"].Command, "MCP connector server command should be loaded correctly")
}

func TestLoadConfigurationFromInvalidFile(t *testing.T) {
	// Test with non-existent file
	mockLogger := &mockLogger{}
	cm := configuration.NewConfigurationManager(mockLogger)
	err := cm.LoadConfigurationFromFile("non-existent-file.yaml")
	assert.Error(t, err, "Loading from non-existent file should return an error")
	assert.Contains(t, err.Error(), "failed to read configuration file", "Error message should mention file reading failure")

	// Test with invalid YAML file
	tmpfileYAML, err := os.CreateTemp("", "config-invalid-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfileYAML.Name())

	if _, err := tmpfileYAML.Write([]byte("invalid: yaml: content: ]")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfileYAML.Close(); err != nil {
		t.Fatal(err)
	}

	err = cm.LoadConfigurationFromFile(tmpfileYAML.Name())
	assert.Error(t, err, "Loading invalid YAML should return an error")
	assert.Contains(t, err.Error(), "failed to parse YAML configuration", "Error message should mention YAML parsing failure")

	// Test with invalid JSON file
	tmpfileJSON, err := os.CreateTemp("", "config-invalid-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfileJSON.Name())

	if _, err := tmpfileJSON.Write([]byte("{ invalid json }")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfileJSON.Close(); err != nil {
		t.Fatal(err)
	}

	err = cm.LoadConfigurationFromFile(tmpfileJSON.Name())
	assert.Error(t, err, "Loading invalid JSON should return an error")
	assert.Contains(t, err.Error(), "failed to parse JSON configuration", "Error message should mention JSON parsing failure")

	// Test with unsupported file extension
	tmpfileUnsupported, err := os.CreateTemp("", "config-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfileUnsupported.Name())

	if _, err := tmpfileUnsupported.Write([]byte("some text")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfileUnsupported.Close(); err != nil {
		t.Fatal(err)
	}

	err = cm.LoadConfigurationFromFile(tmpfileUnsupported.Name())
	assert.Error(t, err, "Loading file with unsupported extension should return an error")
	assert.Contains(t, err.Error(), "unsupported file format", "Error message should mention unsupported format")
}

func TestConfigurationHierarchy(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
agent:
  name: "yaml-agent"
  version: "1.0.0"
  tool:
    name: "yaml-tool"
    description: "YAML tool description"
    argument_name: "yaml_input"
    argument_description: "YAML argument description"
  llm:
    provider: "openai"
    api_key: "yaml-api-key"
    model: "gpt-4"
    max_tokens: 100
    temperature: 0.5
    prompt_template: "YAML template: {{yaml_input}} {{tools}}"
`
	tmpfileYAML, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfileYAML.Name())

	if _, err := tmpfileYAML.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfileYAML.Close(); err != nil {
		t.Fatal(err)
	}

	// Set all required environment variables for testing
	envVars := map[string]string{
		"SPL_AGENT_NAME":                "env-agent",
		"SPL_TOOL_NAME":                 "env-tool",
		"SPL_TOOL_DESCRIPTION":          "env-tool-description",
		"SPL_TOOL_ARGUMENT_NAME":        "env_input",
		"SPL_TOOL_ARGUMENT_DESCRIPTION": "env-argument-description",
		"SPL_LLM_PROVIDER":              "env-provider",
		"SPL_LLM_API_KEY":               "env-api-key",
		"SPL_LLM_MODEL":                 "env-model",
		"SPL_LLM_PROMPT_TEMPLATE":       "ENV template: {{env_input}} {{tools}}",
	}

	// Save current environment to restore later
	oldEnv := map[string]string{}
	for k, v := range envVars {
		if oldVal, exists := os.LookupEnv(k); exists {
			oldEnv[k] = oldVal
		}
		os.Setenv(k, v)
	}

	// Restore environment after test
	defer func() {
		for k := range envVars {
			if oldVal, exists := oldEnv[k]; exists {
				os.Setenv(k, oldVal)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	// Create configuration manager with mock logger
	mockLogger := &mockLogger{}
	cm := configuration.NewConfigurationManager(mockLogger)

	// First load configuration from YAML file
	err = cm.LoadConfigurationFromFile(tmpfileYAML.Name())
	assert.NoError(t, err, "Loading configuration from YAML file should not return an error")

	// Create first set of assertions to verify file loading
	mcpServerConfigFile := cm.GetMCPServerConfig()
	assert.Equal(t, "yaml-agent", mcpServerConfigFile.Name, "Agent name should be loaded from file")
	assert.Equal(t, "yaml-tool", mcpServerConfigFile.Tool.Name, "Tool name should be loaded from file")
	assert.Equal(t, "YAML tool description", mcpServerConfigFile.Tool.Description, "Tool description should be loaded from file")

	llmConfigFile := cm.GetLLMConfig()
	assert.Equal(t, "openai", llmConfigFile.Provider, "LLM provider should be loaded from file")
	assert.Equal(t, "yaml-api-key", llmConfigFile.APIKey, "LLM API key should be loaded from file")

	// Then load configuration from environment variables
	ctx := context.Background()
	err = cm.LoadConfiguration(ctx)
	assert.NoError(t, err, "Loading configuration from environment variables should not return an error")

	// Verify hierarchy: environment variables should override file values
	mcpServerConfig := cm.GetMCPServerConfig()
	assert.Equal(t, "env-agent", mcpServerConfig.Name, "Agent name should be taken from environment variable")
	assert.Equal(t, "env-tool", mcpServerConfig.Tool.Name, "Tool name should be taken from environment variable")
	assert.Equal(t, "env-tool-description", mcpServerConfig.Tool.Description, "Tool description should be taken from environment variable")
	assert.Equal(t, "env_input", mcpServerConfig.Tool.ArgumentName, "Tool argument name should be taken from environment variable")

	llmConfig := cm.GetLLMConfig()
	assert.Equal(t, "env-provider", llmConfig.Provider, "LLM provider should be taken from environment variable")
	assert.Equal(t, "env-api-key", llmConfig.APIKey, "LLM API key should be taken from environment variable")
	assert.Equal(t, "env-model", llmConfig.Model, "LLM model should be taken from environment variable")
	assert.Equal(t, "ENV template: {{env_input}} {{tools}}", llmConfig.SystemPromptTemplate, "LLM prompt template should be taken from environment variable")
}
