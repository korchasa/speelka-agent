package types

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"testing/quick"

	"io"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()
	assert.NotNil(t, config)
}

func TestConfigurationValidate(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		config        *Configuration
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid configuration",
			config: &Configuration{
				Runtime: RuntimeConfig{
					Log: RuntimeLogConfig{
						RawDefaultLevel: "debug",
						RawOutput:       LogOutputStdout,
					},
				},
				Agent: ConfigAgent{
					Name: "TestAgent",
					Tool: AgentToolConfig{
						Name:                "TestTool",
						Description:         "Test tool description",
						ArgumentName:        "query",
						ArgumentDescription: "Query to process",
					},
					LLM: AgentLLMConfig{
						Provider:       "openai",
						Model:          "gpt-4",
						APIKey:         "test-api-key",
						PromptTemplate: "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Missing Agent name",
			config: &Configuration{
				Agent: ConfigAgent{
					Tool: AgentToolConfig{
						Name:                "TestTool",
						Description:         "Test tool description",
						ArgumentName:        "query",
						ArgumentDescription: "Query to process",
					},
					LLM: AgentLLMConfig{
						Provider:       "openai",
						Model:          "gpt-4",
						APIKey:         "test-api-key",
						PromptTemplate: "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}",
					},
				},
			},
			expectError:   true,
			errorContains: "Agent name is required",
		},
		{
			name: "Missing Tool name",
			config: &Configuration{
				Agent: ConfigAgent{
					Name: "TestAgent",
					Tool: AgentToolConfig{
						Description:         "Test tool description",
						ArgumentName:        "query",
						ArgumentDescription: "Query to process",
					},
					LLM: AgentLLMConfig{
						Provider:       "openai",
						Model:          "gpt-4",
						APIKey:         "test-api-key",
						PromptTemplate: "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}",
					},
				},
			},
			expectError:   true,
			errorContains: "Tool name is required",
		},
		{
			name: "Missing LLM API key",
			config: &Configuration{
				Agent: ConfigAgent{
					Name: "TestAgent",
					Tool: AgentToolConfig{
						Name:                "TestTool",
						Description:         "Test tool description",
						ArgumentName:        "query",
						ArgumentDescription: "Query to process",
					},
					LLM: AgentLLMConfig{
						Provider:       "openai",
						Model:          "gpt-4",
						PromptTemplate: "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}",
					},
				},
			},
			expectError:   true,
			errorContains: "LLM API key is required",
		},
		{
			name: "Valid Prompt Template - No tools placeholder",
			config: &Configuration{
				Agent: ConfigAgent{
					Name: "TestAgent",
					Tool: AgentToolConfig{
						Name:                "TestTool",
						Description:         "Test tool description",
						ArgumentName:        "query",
						ArgumentDescription: "Query to process",
					},
					LLM: AgentLLMConfig{
						Provider:       "openai",
						Model:          "gpt-4",
						APIKey:         "test-api-key",
						PromptTemplate: "You are a helpful assistant. User query: {{query}}",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid Prompt Template - Missing query",
			config: &Configuration{
				Agent: ConfigAgent{
					Name: "TestAgent",
					Tool: AgentToolConfig{
						Name:                "TestTool",
						Description:         "Test tool description",
						ArgumentName:        "query",
						ArgumentDescription: "Query to process",
					},
					LLM: AgentLLMConfig{
						Provider:       "openai",
						Model:          "gpt-4",
						APIKey:         "test-api-key",
						PromptTemplate: "You are a helpful assistant. Available tools: {{tools}}",
					},
				},
			},
			expectError:   true,
			errorContains: "template must contain either {{query}} or {{input}} placeholder",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePromptTemplate(t *testing.T) {
	config := NewConfiguration()

	t.Run("valid template with all placeholders", func(t *testing.T) {
		// Test a template with both required placeholders
		template := `This is a template with {{query}} and {{tools}} placeholders`
		err := config.validatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("valid template with input alternative", func(t *testing.T) {
		// Test a template using 'input' instead of query
		template := `This is a template with {{input}} and {{tools}} placeholders`
		err := config.validatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("valid template with only query placeholder", func(t *testing.T) {
		// Test a template with only query placeholder (no tools)
		template := `Template with only {{query}} placeholder`
		err := config.validatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("invalid template missing query placeholder", func(t *testing.T) {
		// Test a template missing the query placeholder
		template := `Template with only {{tools}} placeholder`
		err := config.validatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template must contain either {{query}} or {{input}} placeholder")
	})

	t.Run("invalid empty template", func(t *testing.T) {
		// Test an empty template
		template := ``
		err := config.validatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestExtractPlaceholders(t *testing.T) {
	config := NewConfiguration()

	t.Run("extract multiple placeholders", func(t *testing.T) {
		template := `This is a {{test}} template with {{multiple}} placeholders including {{tools}}`
		placeholders, err := config.extractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"test", "multiple", "tools"}, placeholders)
	})

	t.Run("extract placeholders with whitespace", func(t *testing.T) {
		template := `This has {{ spaced }} placeholders and {{unspaced}} ones`
		placeholders, err := config.extractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"spaced", "unspaced"}, placeholders)
	})

	t.Run("handle no placeholders", func(t *testing.T) {
		template := `This template has no placeholders`
		placeholders, err := config.extractPlaceholders(template)
		assert.NoError(t, err)
		assert.Empty(t, placeholders)
	})

	t.Run("handle empty template", func(t *testing.T) {
		template := ``
		placeholders, err := config.extractPlaceholders(template)
		assert.NoError(t, err)
		assert.Empty(t, placeholders)
	})
}

func TestContains(t *testing.T) {
	t.Run("string present in slice", func(t *testing.T) {
		slice := []string{"a", "b", "c", "test"}
		result := contains(slice, "test")
		assert.True(t, result)
	})

	t.Run("string not present in slice", func(t *testing.T) {
		slice := []string{"a", "b", "c", "test"}
		result := contains(slice, "missing")
		assert.False(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		var slice []string
		result := contains(slice, "test")
		assert.False(t, result)
	})
}

func TestConfiguration_Apply(t *testing.T) {
	// Create a base configuration with some values
	baseConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawDefaultLevel: "info",
				RawOutput:       LogOutputStdout,
				LogLevel:        logrus.InfoLevel,
				Output:          os.Stdout,
			},
			Transports: RuntimeTransportConfig{
				Stdio: RuntimeStdioConfig{
					Enabled:    false,
					BufferSize: 8192,
				},
				HTTP: RuntimeHTTPConfig{
					Enabled: false,
					Host:    "localhost",
					Port:    3000,
				},
			},
		},
		Agent: ConfigAgent{
			Name:    "base-agent",
			Version: "1.0.0",
			Tool: AgentToolConfig{
				Name:                "base-tool",
				Description:         "Base tool description",
				ArgumentName:        "query",
				ArgumentDescription: "Base query description",
			},
			Chat: AgentChatConfig{
				MaxTokens:        2000,
				MaxLLMIterations: 10,
			},
			LLM: AgentLLMConfig{
				Provider:       "openai",
				Model:          "gpt-3.5-turbo",
				APIKey:         "base-api-key",
				MaxTokens:      1000,
				Temperature:    0.7,
				PromptTemplate: "Base template with {{query}} and {{tools}}",
				Retry: LLMRetryConfig{
					MaxRetries:        3,
					InitialBackoff:    1.0,
					MaxBackoff:        30.0,
					BackoffMultiplier: 2.0,
				},
			},
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"server1": {
						URL:    "http://server1.example.com",
						APIKey: "server1-key",
						Args:   []string{"--arg1", "--arg2"},
					},
				},
				Retry: ConnectionRetryConfig{
					MaxRetries:        3,
					InitialBackoff:    1.0,
					MaxBackoff:        30.0,
					BackoffMultiplier: 2.0,
				},
			},
		},
	}

	// Create a new configuration with some values to overlay
	newConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawDefaultLevel: "debug",
				// Output not specified, should keep stdout
			},
			Transports: RuntimeTransportConfig{
				Stdio: RuntimeStdioConfig{
					Enabled:    true,
					BufferSize: 16384,
				},
				HTTP: RuntimeHTTPConfig{
					Enabled: true,
					Port:    8080,
				},
			},
		},
		Agent: ConfigAgent{
			// Name not specified, should keep base-agent
			Version: "1.1.0",
			Tool: AgentToolConfig{
				Name:        "new-tool",
				Description: "New tool description",
				// ArgumentName not specified, should keep query
				// ArgumentDescription not specified, should keep base description
			},
			Chat: AgentChatConfig{
				MaxTokens: 2500,
				// MaxLLMIterations not specified, should keep 10
			},
			LLM: AgentLLMConfig{
				// Provider not specified, should keep openai
				Model:     "gpt-4",
				APIKey:    "new-api-key",
				MaxTokens: 1500,
				// Temperature not specified, should keep 0.7
				PromptTemplate: "New template with {{query}} and {{tools}}",
				IsMaxTokensSet: true,
				Retry: LLMRetryConfig{
					MaxRetries: 5,
					// Other retry settings not specified, should keep defaults
				},
			},
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"server2": {
						URL:    "http://server2.example.com",
						APIKey: "server2-key",
						Args:   []string{"--arg3", "--arg4"},
					},
				},
				Retry: ConnectionRetryConfig{
					InitialBackoff: 2.0,
					// Other retry settings not specified, should keep defaults
				},
			},
		},
	}

	// Apply the new configuration to the base
	result, err := baseConfig.Apply(newConfig)
	assert.NoError(t, err)

	// Debug transport settings
	fmt.Printf("Base Transports before apply:\n")
	fmt.Printf("Stdio.Enabled: %v\n", baseConfig.Runtime.Transports.Stdio.Enabled)
	fmt.Printf("Stdio.BufferSize: %v\n", baseConfig.Runtime.Transports.Stdio.BufferSize)
	fmt.Printf("HTTP.Enabled: %v\n", baseConfig.Runtime.Transports.HTTP.Enabled)
	fmt.Printf("HTTP.Host: %v\n", baseConfig.Runtime.Transports.HTTP.Host)
	fmt.Printf("HTTP.Port: %v\n", baseConfig.Runtime.Transports.HTTP.Port)

	fmt.Printf("\nNew Transports in newConfig:\n")
	fmt.Printf("Stdio.Enabled: %v\n", newConfig.Runtime.Transports.Stdio.Enabled)
	fmt.Printf("Stdio.BufferSize: %v\n", newConfig.Runtime.Transports.Stdio.BufferSize)
	fmt.Printf("HTTP.Enabled: %v\n", newConfig.Runtime.Transports.HTTP.Enabled)
	fmt.Printf("HTTP.Host: %v\n", newConfig.Runtime.Transports.HTTP.Host)
	fmt.Printf("HTTP.Port: %v\n", newConfig.Runtime.Transports.HTTP.Port)

	fmt.Printf("\nBase Transports after apply:\n")
	fmt.Printf("Stdio.Enabled: %v\n", baseConfig.Runtime.Transports.Stdio.Enabled)
	fmt.Printf("Stdio.BufferSize: %v\n", baseConfig.Runtime.Transports.Stdio.BufferSize)
	fmt.Printf("HTTP.Enabled: %v\n", baseConfig.Runtime.Transports.HTTP.Enabled)
	fmt.Printf("HTTP.Host: %v\n", baseConfig.Runtime.Transports.HTTP.Host)
	fmt.Printf("HTTP.Port: %v\n", baseConfig.Runtime.Transports.HTTP.Port)

	// Verify it returns the same instance
	assert.Equal(t, baseConfig, result)

	// Verify that the values were updated correctly
	assert.Equal(t, "debug", baseConfig.Runtime.Log.RawDefaultLevel)

	// Verify runtime transport values
	assert.Equal(t, true, baseConfig.Runtime.Transports.Stdio.Enabled, "Stdio.Enabled should be true")
	assert.Equal(t, 16384, baseConfig.Runtime.Transports.Stdio.BufferSize, "Stdio.BufferSize should be 16384")
	assert.Equal(t, true, baseConfig.Runtime.Transports.HTTP.Enabled, "HTTP.Enabled should be true")
	assert.Equal(t, "localhost", baseConfig.Runtime.Transports.HTTP.Host, "HTTP.Host should be localhost")
	assert.Equal(t, 8080, baseConfig.Runtime.Transports.HTTP.Port, "HTTP.Port should be 8080")

	// Verify agent values
	assert.Equal(t, "base-agent", baseConfig.Agent.Name)
	assert.Equal(t, "1.1.0", baseConfig.Agent.Version)
	assert.Equal(t, "new-tool", baseConfig.Agent.Tool.Name)
	assert.Equal(t, "New tool description", baseConfig.Agent.Tool.Description)
	assert.Equal(t, "query", baseConfig.Agent.Tool.ArgumentName)
	assert.Equal(t, "Base query description", baseConfig.Agent.Tool.ArgumentDescription)

	// Verify LLM values
	assert.Equal(t, "openai", baseConfig.Agent.LLM.Provider)
	assert.Equal(t, "gpt-4", baseConfig.Agent.LLM.Model)
	assert.Equal(t, "new-api-key", baseConfig.Agent.LLM.APIKey)
	assert.Equal(t, 1500, baseConfig.Agent.LLM.MaxTokens)
	assert.Equal(t, 0.7, baseConfig.Agent.LLM.Temperature)
	assert.Equal(t, "New template with {{query}} and {{tools}}", baseConfig.Agent.LLM.PromptTemplate)

	// Verify Chat values
	assert.Equal(t, 2500, baseConfig.Agent.Chat.MaxTokens)
	assert.Equal(t, 10, baseConfig.Agent.Chat.MaxLLMIterations)

	// Verify LLM Retry values
	assert.Equal(t, 5, baseConfig.Agent.LLM.Retry.MaxRetries)
	assert.Equal(t, 1.0, baseConfig.Agent.LLM.Retry.InitialBackoff)
	assert.Equal(t, 30.0, baseConfig.Agent.LLM.Retry.MaxBackoff)
	assert.Equal(t, 2.0, baseConfig.Agent.LLM.Retry.BackoffMultiplier)

	// Verify Connection Retry values
	assert.Equal(t, 3, baseConfig.Agent.Connections.Retry.MaxRetries)
	assert.Equal(t, 2.0, baseConfig.Agent.Connections.Retry.InitialBackoff)
	assert.Equal(t, 30.0, baseConfig.Agent.Connections.Retry.MaxBackoff)
	assert.Equal(t, 2.0, baseConfig.Agent.Connections.Retry.BackoffMultiplier)

	// Verify that the connections were merged, not replaced
	assert.Len(t, baseConfig.Agent.Connections.McpServers, 2)
	assert.Contains(t, baseConfig.Agent.Connections.McpServers, "server1")
	assert.Contains(t, baseConfig.Agent.Connections.McpServers, "server2")
	assert.Equal(t, "http://server1.example.com", baseConfig.Agent.Connections.McpServers["server1"].URL)
	assert.Equal(t, "http://server2.example.com", baseConfig.Agent.Connections.McpServers["server2"].URL)

	// Test with an invalid log level
	invalidLogConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawDefaultLevel: "invalid_level",
				RawOutput:       LogOutputStderr,
			},
		},
	}

	_, err = baseConfig.Apply(invalidLogConfig)
	assert.Error(t, err)
	assert.Equal(t, "invalid_level", baseConfig.Runtime.Log.RawDefaultLevel)

	// Test with file output
	// First create a temp directory for the test
	tempDir, err := os.MkdirTemp("", "config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "test.log")
	fileLogConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawDefaultLevel: "warn",
				RawOutput:       logFilePath,
			},
		},
	}

	_, err = baseConfig.Apply(fileLogConfig)
	assert.NoError(t, err)
	assert.Equal(t, "warn", baseConfig.Runtime.Log.RawDefaultLevel)
	// Writer should be a file now, not stdout or stderr
	assert.NotEqual(t, os.Stdout, baseConfig.Runtime.Log.Output)
	assert.NotEqual(t, os.Stderr, baseConfig.Runtime.Log.Output)

	// Close the file to avoid resource leaks in the test
	if file, ok := baseConfig.Runtime.Log.Output.(*os.File); ok && file != os.Stdout && file != os.Stderr {
		file.Close()
	}

	// Test precedence: config file with empty APIKey, env with non-empty APIKey
	baseConfigWithEmptyAPIKey := &Configuration{
		Agent: ConfigAgent{
			LLM: AgentLLMConfig{
				APIKey: "",
			},
		},
	}
	overlayConfigWithEnvAPIKey := &Configuration{
		Agent: ConfigAgent{
			LLM: AgentLLMConfig{
				APIKey: "env-api-key",
			},
		},
	}
	_, err = baseConfigWithEmptyAPIKey.Apply(overlayConfigWithEnvAPIKey)
	assert.NoError(t, err)
	assert.Equal(t, "env-api-key", baseConfigWithEmptyAPIKey.Agent.LLM.APIKey, "Environment variable should override empty config file value for APIKey")
}

func TestMCPServerConnection_ToolFilters(t *testing.T) {
	t.Run("includeTools only", func(t *testing.T) {
		cfg := MCPServerConnection{
			IncludeTools: []string{"foo", "bar"},
		}
		assert.Equal(t, []string{"foo", "bar"}, cfg.IncludeTools)
		assert.Nil(t, cfg.ExcludeTools)
	})
	t.Run("excludeTools only", func(t *testing.T) {
		cfg := MCPServerConnection{
			ExcludeTools: []string{"baz"},
		}
		assert.Nil(t, cfg.IncludeTools)
		assert.Equal(t, []string{"baz"}, cfg.ExcludeTools)
	})
	t.Run("both includeTools and excludeTools", func(t *testing.T) {
		cfg := MCPServerConnection{
			IncludeTools: []string{"foo", "bar"},
			ExcludeTools: []string{"bar"},
		}
		assert.Equal(t, []string{"foo", "bar"}, cfg.IncludeTools)
		assert.Equal(t, []string{"bar"}, cfg.ExcludeTools)
	})
	t.Run("neither includeTools nor excludeTools", func(t *testing.T) {
		cfg := MCPServerConnection{}
		assert.Nil(t, cfg.IncludeTools)
		assert.Nil(t, cfg.ExcludeTools)
	})
}

func TestConfiguration_Apply_McpServerToolFilters(t *testing.T) {
	base := &Configuration{
		Agent: ConfigAgent{
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv": {
						IncludeTools: []string{"foo", "bar"},
						ExcludeTools: nil,
					},
				},
			},
		},
	}

	// Overlay: only ExcludeTools
	overlay1 := &Configuration{
		Agent: ConfigAgent{
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv": {
						ExcludeTools: []string{"baz"},
					},
				},
			},
		},
	}
	_, err := base.Apply(overlay1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar"}, base.Agent.Connections.McpServers["srv"].IncludeTools)
	assert.Equal(t, []string{"baz"}, base.Agent.Connections.McpServers["srv"].ExcludeTools)

	// Overlay: only IncludeTools (should overwrite)
	overlay2 := &Configuration{
		Agent: ConfigAgent{
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv": {
						IncludeTools: []string{"qux"},
					},
				},
			},
		},
	}
	_, err = base.Apply(overlay2)
	assert.NoError(t, err)
	assert.Equal(t, []string{"qux"}, base.Agent.Connections.McpServers["srv"].IncludeTools)
	assert.Equal(t, []string{"baz"}, base.Agent.Connections.McpServers["srv"].ExcludeTools)

	// Overlay: both IncludeTools and ExcludeTools (should overwrite both)
	overlay3 := &Configuration{
		Agent: ConfigAgent{
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv": {
						IncludeTools: []string{"a", "b"},
						ExcludeTools: []string{"c"},
					},
				},
			},
		},
	}
	_, err = base.Apply(overlay3)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, base.Agent.Connections.McpServers["srv"].IncludeTools)
	assert.Equal(t, []string{"c"}, base.Agent.Connections.McpServers["srv"].ExcludeTools)

	// Overlay: nil IncludeTools/ExcludeTools (should preserve previous values)
	overlay4 := &Configuration{
		Agent: ConfigAgent{
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv": {
						IncludeTools: nil,
						ExcludeTools: nil,
					},
				},
			},
		},
	}
	_, err = base.Apply(overlay4)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, base.Agent.Connections.McpServers["srv"].IncludeTools)
	assert.Equal(t, []string{"c"}, base.Agent.Connections.McpServers["srv"].ExcludeTools)
}

func TestConfiguration_RedactedCopy(t *testing.T) {
	orig := &Configuration{
		Agent: ConfigAgent{
			LLM: AgentLLMConfig{
				APIKey: "super-secret-llm-key",
			},
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"server1": {
						APIKey: "server1-key",
						URL:    "http://server1",
					},
					"server2": {
						APIKey: "server2-key",
						URL:    "http://server2",
					},
				},
			},
		},
	}

	redacted := orig.RedactedCopy()
	assert.Equal(t, "***REDACTED***", redacted.Agent.LLM.APIKey)
	for k, v := range redacted.Agent.Connections.McpServers {
		assert.Equal(t, "***REDACTED***", v.APIKey, "APIKey for %s should be redacted", k)
	}
	// Ensure other fields are unchanged
	assert.Equal(t, orig.Agent.LLM.Model, redacted.Agent.LLM.Model)
	assert.Equal(t, orig.Agent.Connections.McpServers["server1"].URL, redacted.Agent.Connections.McpServers["server1"].URL)
}

func TestConfiguration_Converters(t *testing.T) {
	baseConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawDefaultLevel: "info",
				RawOutput:       LogOutputStdout,
			},
			Transports: RuntimeTransportConfig{
				Stdio: RuntimeStdioConfig{
					Enabled:    true,
					BufferSize: 4096,
				},
				HTTP: RuntimeHTTPConfig{
					Enabled: true,
					Host:    "127.0.0.1",
					Port:    8080,
				},
			},
		},
		Agent: ConfigAgent{
			Name:    "agent-test",
			Version: "v1.2.3",
			Tool: AgentToolConfig{
				Name:                "tool1",
				Description:         "desc1",
				ArgumentName:        "arg1",
				ArgumentDescription: "argdesc1",
			},
			Chat: AgentChatConfig{
				MaxTokens:        1234,
				MaxLLMIterations: 7,
			},
			LLM: AgentLLMConfig{
				Provider:         "openai",
				Model:            "gpt-4",
				APIKey:           "api-key-123",
				MaxTokens:        2048,
				IsMaxTokensSet:   true,
				Temperature:      0.5,
				IsTemperatureSet: true,
				PromptTemplate:   "prompt {{arg1}}",
				Retry: LLMRetryConfig{
					MaxRetries:        2,
					InitialBackoff:    1.5,
					MaxBackoff:        10.0,
					BackoffMultiplier: 2.5,
				},
			},
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv1": {
						URL:    "http://srv1",
						APIKey: "srv1-key",
					},
				},
				Retry: ConnectionRetryConfig{
					MaxRetries:        3,
					InitialBackoff:    2.0,
					MaxBackoff:        20.0,
					BackoffMultiplier: 3.0,
				},
			},
		},
	}

	t.Run("ToAgentConfig", func(t *testing.T) {
		agentCfg := baseConfig.ToAgentConfig()
		assert.Equal(t, "tool1", agentCfg.Tool.Name)
		assert.Equal(t, "gpt-4", agentCfg.Model)
		assert.Equal(t, "prompt {{arg1}}", agentCfg.SystemPromptTemplate)
		assert.Equal(t, 1234, agentCfg.MaxTokens)
		assert.Equal(t, 7, agentCfg.MaxLLMIterations)
	})

	t.Run("ToLLMConfig", func(t *testing.T) {
		llmCfg := baseConfig.ToLLMConfig()
		assert.Equal(t, "openai", llmCfg.Provider)
		assert.Equal(t, "gpt-4", llmCfg.Model)
		assert.Equal(t, "api-key-123", llmCfg.APIKey)
		assert.Equal(t, 2048, llmCfg.MaxTokens)
		assert.True(t, llmCfg.IsMaxTokensSet)
		assert.Equal(t, 0.5, llmCfg.Temperature)
		assert.True(t, llmCfg.IsTemperatureSet)
		assert.Equal(t, "prompt {{arg1}}", llmCfg.SystemPromptTemplate)
		assert.Equal(t, 2, llmCfg.RetryConfig.MaxRetries)
		assert.Equal(t, 1.5, llmCfg.RetryConfig.InitialBackoff)
		assert.Equal(t, 10.0, llmCfg.RetryConfig.MaxBackoff)
		assert.Equal(t, 2.5, llmCfg.RetryConfig.BackoffMultiplier)
	})

	t.Run("ToMCPServerConfig", func(t *testing.T) {
		mcpSrvCfg := baseConfig.ToMCPServerConfig()
		assert.Equal(t, "agent-test", mcpSrvCfg.Name)
		assert.Equal(t, "v1.2.3", mcpSrvCfg.Version)
		assert.Equal(t, true, mcpSrvCfg.HTTP.Enabled)
		assert.Equal(t, "127.0.0.1", mcpSrvCfg.HTTP.Host)
		assert.Equal(t, 8080, mcpSrvCfg.HTTP.Port)
		assert.Equal(t, true, mcpSrvCfg.Stdio.Enabled)
		assert.Equal(t, 4096, mcpSrvCfg.Stdio.BufferSize)
		assert.Equal(t, "tool1", mcpSrvCfg.Tool.Name)
		assert.Equal(t, LogOutputStdout, mcpSrvCfg.LogRawOutput)
	})

	t.Run("ToMCPConnectorConfig", func(t *testing.T) {
		mcpConnCfg := baseConfig.ToMCPConnectorConfig()
		assert.Contains(t, mcpConnCfg.McpServers, "srv1")
		assert.Equal(t, "http://srv1", mcpConnCfg.McpServers["srv1"].URL)
		assert.Equal(t, "srv1-key", mcpConnCfg.McpServers["srv1"].APIKey)
		assert.Equal(t, 3, mcpConnCfg.RetryConfig.MaxRetries)
		assert.Equal(t, 2.0, mcpConnCfg.RetryConfig.InitialBackoff)
		assert.Equal(t, 20.0, mcpConnCfg.RetryConfig.MaxBackoff)
		assert.Equal(t, 3.0, mcpConnCfg.RetryConfig.BackoffMultiplier)
	})
}

func TestConfiguration_Overlay_PropertyBased(t *testing.T) {
	f := func(base, overlay Configuration) bool {
		// Копируем base для overlay
		baseCopy := base
		result, err := baseCopy.Apply(&overlay)
		if err != nil {
			t.Logf("Apply error: %v", err)
			return false
		}
		// Проверяем, что overlay не затирает дефолтные значения zero-value полями
		if overlay.Agent.LLM.APIKey == "" && result.Agent.LLM.APIKey != base.Agent.LLM.APIKey {
			t.Logf("APIKey zero-value overlay: want %q, got %q", base.Agent.LLM.APIKey, result.Agent.LLM.APIKey)
			return false
		}
		// Проверяем, что overlay мержит map, а не заменяет
		if len(base.Agent.Connections.McpServers) > 0 && len(overlay.Agent.Connections.McpServers) > 0 {
			for k, v := range base.Agent.Connections.McpServers {
				if _, ok := result.Agent.Connections.McpServers[k]; !ok {
					t.Logf("McpServers map merge lost key: %q", k)
					return false
				}
				// Проверяем, что overlay не затирает поля, если они zero-value
				if overlay.Agent.Connections.McpServers[k].APIKey == "" && result.Agent.Connections.McpServers[k].APIKey != v.APIKey {
					t.Logf("McpServers[%q].APIKey zero-value overlay: want %q, got %q", k, v.APIKey, result.Agent.Connections.McpServers[k].APIKey)
					return false
				}
			}
		}
		return true
	}
	cfg := &quick.Config{
		MaxCount: 50,
		Values: func(args []reflect.Value, r *rand.Rand) {
			args[0] = reflect.ValueOf(randomConfig(r))
			args[1] = reflect.ValueOf(randomConfig(r))
		},
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// randomConfig генерирует случайную конфигурацию для property-based тестов
func randomConfig(r *rand.Rand) Configuration {
	cfg := NewConfiguration()
	cfg.Agent.Name = randomString(r)
	cfg.Agent.LLM.APIKey = randomString(r)
	cfg.Agent.LLM.Model = randomString(r)
	cfg.Agent.LLM.Provider = randomString(r)
	cfg.Agent.LLM.PromptTemplate = randomString(r)
	cfg.Agent.LLM.MaxTokens = r.Intn(10000)
	cfg.Agent.LLM.Temperature = r.Float64()
	cfg.Agent.Connections.McpServers = make(map[string]MCPServerConnection)
	if r.Intn(2) == 1 {
		key := randomString(r)
		cfg.Agent.Connections.McpServers[key] = MCPServerConnection{
			URL:    randomString(r),
			APIKey: randomString(r),
		}
	}
	return *cfg
}

func randomString(r *rand.Rand) string {
	length := r.Intn(5)
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(r.Intn(26) + 97)
	}
	return string(b)
}

func TestConfiguration_Serialization_Golden(t *testing.T) {
	cfg := NewConfiguration()
	goldenPath := filepath.Join("testdata", "configuration_golden.json")
	actual, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	goldenFile, err := os.Open(goldenPath)
	if err != nil {
		// Если golden-файл не найден — создаём его
		t.Logf("Golden file not found, creating: %s", goldenPath)
		f, err := os.Create(goldenPath)
		if err != nil {
			t.Fatalf("failed to create golden file: %v", err)
		}
		defer f.Close()
		_, err = f.Write(actual)
		if err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}
	defer goldenFile.Close()
	golden, err := io.ReadAll(goldenFile)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}
	if string(actual) != string(golden) {
		t.Errorf("configuration serialization mismatch with golden file\nActual:\n%s\nGolden:\n%s", actual, golden)
	}
}

func TestConfiguration_Apply_LogRawFields(t *testing.T) {
	baseConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawDefaultLevel: "info",
				RawOutput:       LogOutputStdout,
				RawFormat:       "text",
				LogLevel:        logrus.InfoLevel,
				Output:          os.Stdout,
			},
		},
	}

	resetBase := func() {
		baseConfig.Runtime.Log.RawDefaultLevel = "info"
		baseConfig.Runtime.Log.RawOutput = LogOutputStdout
		baseConfig.Runtime.Log.RawFormat = "text"
		baseConfig.Runtime.Log.LogLevel = logrus.InfoLevel
		baseConfig.Runtime.Log.Output = os.Stdout
	}

	t.Run("overlay only RawFormat", func(t *testing.T) {
		resetBase()
		overlay := &Configuration{
			Runtime: RuntimeConfig{
				Log: RuntimeLogConfig{
					RawFormat: "json",
				},
			},
		}
		_, err := baseConfig.Apply(overlay)
		assert.NoError(t, err)
		assert.Equal(t, "json", baseConfig.Runtime.Log.RawFormat)
	})

	t.Run("overlay all Raw fields", func(t *testing.T) {
		resetBase()
		overlay := &Configuration{
			Runtime: RuntimeConfig{
				Log: RuntimeLogConfig{
					RawDefaultLevel: "debug",
					RawOutput:       LogOutputStderr,
					RawFormat:       "json",
				},
			},
		}
		_, err := baseConfig.Apply(overlay)
		assert.NoError(t, err)
		assert.Equal(t, "debug", baseConfig.Runtime.Log.RawDefaultLevel)
		assert.Equal(t, LogOutputStderr, baseConfig.Runtime.Log.RawOutput)
		assert.Equal(t, "json", baseConfig.Runtime.Log.RawFormat)
	})

	t.Run("overlay only RawOutput (file, error)", func(t *testing.T) {
		resetBase()
		overlay := &Configuration{
			Runtime: RuntimeConfig{
				Log: RuntimeLogConfig{
					RawOutput: "/nonexistent/forbidden.log",
				},
			},
		}
		_, err := baseConfig.Apply(overlay)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open log file")
	})

	t.Run("overlay only RawDefaultLevel", func(t *testing.T) {
		resetBase()
		overlay := &Configuration{
			Runtime: RuntimeConfig{
				Log: RuntimeLogConfig{
					RawDefaultLevel: "warn",
				},
			},
		}
		_, err := baseConfig.Apply(overlay)
		assert.NoError(t, err)
		assert.Equal(t, "warn", baseConfig.Runtime.Log.RawDefaultLevel)
	})

	// Проверка, что если RawFormat не задан, старое значение сохраняется
	t.Run("overlay without RawFormat keeps previous", func(t *testing.T) {
		resetBase()
		baseConfig.Runtime.Log.RawFormat = "json"
		overlay := &Configuration{
			Runtime: RuntimeConfig{
				Log: RuntimeLogConfig{
					RawDefaultLevel: "error",
				},
			},
		}
		_, err := baseConfig.Apply(overlay)
		assert.NoError(t, err)
		assert.Equal(t, "json", baseConfig.Runtime.Log.RawFormat)
	})
}
