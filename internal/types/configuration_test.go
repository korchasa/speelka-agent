package types

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
						RawLevel: "debug",
						Output:   "stdout",
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
			name: "Invalid Prompt Template - Missing tools",
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
			expectError:   true,
			errorContains: "template must contain {{tools}} placeholder",
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

	t.Run("invalid template missing query placeholder", func(t *testing.T) {
		// Test a template missing the query placeholder
		template := `Template with only {{tools}} placeholder`
		err := config.validatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template must contain either {{query}} or {{input}} placeholder")
	})

	t.Run("invalid template missing tools placeholder", func(t *testing.T) {
		// Test a template missing the tools placeholder
		template := `Template with only {{query}} placeholder`
		err := config.validatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template must contain {{tools}} placeholder")
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
				RawLevel: "info",
				Output:   "stdout",
				LogLevel: logrus.InfoLevel,
				Writer:   os.Stdout,
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
				MaxTokens:          2000,
				CompactionStrategy: "delete-old",
				MaxLLMIterations:   10,
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
				RawLevel: "debug",
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
				// CompactionStrategy not specified, should keep delete-old
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
	result := baseConfig.Apply(newConfig)

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
	assert.Equal(t, "debug", baseConfig.Runtime.Log.RawLevel)
	assert.Equal(t, "stdout", baseConfig.Runtime.Log.Output)
	assert.Equal(t, logrus.DebugLevel, baseConfig.Runtime.Log.LogLevel)
	assert.Equal(t, os.Stdout, baseConfig.Runtime.Log.Writer)

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
	assert.Equal(t, "delete-old", baseConfig.Agent.Chat.CompactionStrategy)
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
				RawLevel: "invalid_level",
				Output:   "stderr",
			},
		},
	}

	baseConfig.Apply(invalidLogConfig)
	assert.Equal(t, "invalid_level", baseConfig.Runtime.Log.RawLevel)
	assert.Equal(t, "stderr", baseConfig.Runtime.Log.Output)
	// Should default to info level for invalid log level
	assert.Equal(t, logrus.InfoLevel, baseConfig.Runtime.Log.LogLevel)
	assert.Equal(t, os.Stderr, baseConfig.Runtime.Log.Writer)

	// Test with file output
	// First create a temp directory for the test
	tempDir, err := os.MkdirTemp("", "config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "test.log")
	fileLogConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				RawLevel: "warn",
				Output:   logFilePath,
			},
		},
	}

	baseConfig.Apply(fileLogConfig)
	assert.Equal(t, "warn", baseConfig.Runtime.Log.RawLevel)
	assert.Equal(t, logFilePath, baseConfig.Runtime.Log.Output)
	assert.Equal(t, logrus.WarnLevel, baseConfig.Runtime.Log.LogLevel)
	// Writer should be a file now, not stdout or stderr
	assert.NotEqual(t, os.Stdout, baseConfig.Runtime.Log.Writer)
	assert.NotEqual(t, os.Stderr, baseConfig.Runtime.Log.Writer)

	// Close the file to avoid resource leaks in the test
	if file, ok := baseConfig.Runtime.Log.Writer.(*os.File); ok && file != os.Stdout && file != os.Stderr {
		file.Close()
	}
}

/* These functions might be defined elsewhere or have been removed
func TestLoadLevelFromName(t *testing.T) {
	tests := []struct {
		levelName      string
		expectedLevel  logrus.Level
		expectedOutput string
		expectError    bool
	}{
		{"debug", logrus.DebugLevel, "debug", false},
		{"Debug", logrus.DebugLevel, "debug", false},
		{"DEBUG", logrus.DebugLevel, "debug", false},
		{"info", logrus.InfoLevel, "info", false},
		{"Info", logrus.InfoLevel, "info", false},
		{"INFO", logrus.InfoLevel, "info", false},
		{"warn", logrus.WarnLevel, "warn", false},
		{"Warn", logrus.WarnLevel, "warn", false},
		{"WARN", logrus.WarnLevel, "warn", false},
		{"warning", logrus.WarnLevel, "warn", false},
		{"Warning", logrus.WarnLevel, "warn", false},
		{"WARNING", logrus.WarnLevel, "warn", false},
		{"error", logrus.ErrorLevel, "error", false},
		{"Error", logrus.ErrorLevel, "error", false},
		{"ERROR", logrus.ErrorLevel, "error", false},
		{"fatal", logrus.FatalLevel, "fatal", false},
		{"Fatal", logrus.FatalLevel, "fatal", false},
		{"FATAL", logrus.FatalLevel, "fatal", false},
		{"panic", logrus.PanicLevel, "panic", false},
		{"Panic", logrus.PanicLevel, "panic", false},
		{"PANIC", logrus.PanicLevel, "panic", false},
		{"unknown", logrus.InfoLevel, "info", true},
		{"", logrus.InfoLevel, "info", true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Level %s", test.levelName), func(t *testing.T) {
			level, output, err := LoadLevelFromName(test.levelName)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expectedLevel, level)
			assert.Equal(t, test.expectedOutput, output)
		})
	}
}

func TestLoadOutputFromName(t *testing.T) {
	tests := []struct {
		outputName   string
		expectStdout bool
		expectStderr bool
		expectFile   bool
		expectError  bool
	}{
		{"stdout", true, false, false, false},
		{"STDOUT", true, false, false, false},
		{"Stdout", true, false, false, false},
		{"stderr", false, true, false, false},
		{"STDERR", false, true, false, false},
		{"Stderr", false, true, false, false},
		{"test.log", false, false, true, false},
		{"/var/log/app.log", false, false, true, false},
		{"./log/app.log", false, false, true, false},
		{"", false, true, false, false}, // Default to stderr
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Output %s", test.outputName), func(t *testing.T) {
			writer, err := LoadOutputFromName(test.outputName)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if test.expectStdout {
				assert.Equal(t, os.Stdout, writer)
			}

			if test.expectStderr {
				assert.Equal(t, os.Stderr, writer)
			}

			// Can't easily check if a specific file was opened,
			// so just check it's not stdout/stderr
			if test.expectFile {
				assert.NotEqual(t, os.Stdout, writer)
				assert.NotEqual(t, os.Stderr, writer)
			}
		})
	}
}
*/
