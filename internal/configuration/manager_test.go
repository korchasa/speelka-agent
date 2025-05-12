package configuration

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleLogger implements the LoggerSpec interface for testing
type SimpleLogger struct{}

func (m *SimpleLogger) Debug(args ...interface{})                 {}
func (m *SimpleLogger) Debugf(format string, args ...interface{}) {}
func (m *SimpleLogger) Info(args ...interface{})                  {}
func (m *SimpleLogger) Infof(format string, args ...interface{})  {}
func (m *SimpleLogger) Warn(args ...interface{})                  {}
func (m *SimpleLogger) Warnf(format string, args ...interface{})  {}
func (m *SimpleLogger) Error(args ...interface{})                 {}
func (m *SimpleLogger) Errorf(format string, args ...interface{}) {}
func (m *SimpleLogger) Fatal(args ...interface{})                 {}
func (m *SimpleLogger) Fatalf(format string, args ...interface{}) {}
func (m *SimpleLogger) WithField(key string, value interface{}) types.LogEntrySpec {
	return &SimpleLogEntry{}
}
func (m *SimpleLogger) WithFields(fields logrus.Fields) types.LogEntrySpec {
	return &SimpleLogEntry{}
}
func (m *SimpleLogger) SetLevel(level logrus.Level)        {}
func (m *SimpleLogger) SetMCPServer(mcpServer interface{}) {}

// SimpleLogEntry implements the LogEntrySpec interface for testing
type SimpleLogEntry struct{}

func (m *SimpleLogEntry) Debug(args ...interface{})                 {}
func (m *SimpleLogEntry) Debugf(format string, args ...interface{}) {}
func (m *SimpleLogEntry) Info(args ...interface{})                  {}
func (m *SimpleLogEntry) Infof(format string, args ...interface{})  {}
func (m *SimpleLogEntry) Warn(args ...interface{})                  {}
func (m *SimpleLogEntry) Warnf(format string, args ...interface{})  {}
func (m *SimpleLogEntry) Error(args ...interface{})                 {}
func (m *SimpleLogEntry) Errorf(format string, args ...interface{}) {}
func (m *SimpleLogEntry) Fatal(args ...interface{})                 {}
func (m *SimpleLogEntry) Fatalf(format string, args ...interface{}) {}

func TestConfigurationManager_LoadConfiguration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config-manager-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid YAML configuration file
	validYAML := `
runtime:
  log:
    level: debug
    output: ./test.log

agent:
  name: test-agent
  tool:
    name: test-tool
    description: A test tool
    argument_name: query
    argument_description: The query to process
  llm:
    provider: openai
    model: gpt-4
    api_key: test-api-key
    prompt_template: "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}"
`

	validYAMLPath := filepath.Join(tempDir, "valid-config.yaml")
	err = os.WriteFile(validYAMLPath, []byte(validYAML), 0644)
	require.NoError(t, err)

	// Set up simple logger
	logger := &SimpleLogger{}

	// Test loading from file
	t.Run("Load from file", func(t *testing.T) {
		cm := NewConfigurationManager(logger)

		// Load from file
		err := cm.LoadConfiguration(context.Background(), validYAMLPath)
		assert.NoError(t, err)

		// Verify configuration is loaded
		assert.Equal(t, "test-agent", cm.config.Agent.Name)
		assert.Equal(t, "test-tool", cm.config.Agent.Tool.Name)
		assert.Equal(t, "A test tool", cm.config.Agent.Tool.Description)
		assert.Equal(t, "query", cm.config.Agent.Tool.ArgumentName)
		assert.Equal(t, "The query to process", cm.config.Agent.Tool.ArgumentDescription)
		assert.Equal(t, "openai", cm.config.Agent.LLM.Provider)
		assert.Equal(t, "gpt-4", cm.config.Agent.LLM.Model)
		assert.Equal(t, "test-api-key", cm.config.Agent.LLM.APIKey)
		assert.Equal(t, "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}", cm.config.Agent.LLM.PromptTemplate)
		assert.Equal(t, "debug", cm.config.Runtime.Log.RawLevel)
		assert.Equal(t, "./test.log", cm.config.Runtime.Log.RawOutput)
		assert.Equal(t, logrus.DebugLevel, cm.config.Runtime.Log.LogLevel)

		// Verify getters return the correct values
		mcpServerConfig := cm.GetMCPServerConfig()
		assert.Equal(t, "test-agent", mcpServerConfig.Name)
		assert.Equal(t, "test-tool", mcpServerConfig.Tool.Name)
		assert.Equal(t, "A test tool", mcpServerConfig.Tool.Description)
		assert.Equal(t, "query", mcpServerConfig.Tool.ArgumentName)
		assert.Equal(t, "The query to process", mcpServerConfig.Tool.ArgumentDescription)

		llmConfig := cm.GetLLMConfig()
		assert.Equal(t, "openai", llmConfig.Provider)
		assert.Equal(t, "gpt-4", llmConfig.Model)
		assert.Equal(t, "test-api-key", llmConfig.APIKey)
		assert.Equal(t, "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}", llmConfig.SystemPromptTemplate)

		logConfig := cm.GetLogConfig()
		assert.Equal(t, logrus.DebugLevel, logConfig.Level)
		assert.Equal(t, "./test.log", logConfig.RawOutput)

		agentConfig := cm.GetAgentConfig()
		assert.Equal(t, "test-tool", agentConfig.Tool.Name)
		assert.Equal(t, "A test tool", agentConfig.Tool.Description)
		assert.Equal(t, "gpt-4", agentConfig.Model)
		assert.Equal(t, "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}", agentConfig.SystemPromptTemplate)
	})

	// Test loading without a file (default values)
	t.Run("Load defaults", func(t *testing.T) {
		// Create a default configuration file with minimal settings
		minimalConfigYAML := `
runtime:
  log:
    level: info
    output: stdout

agent:
  name: default-agent
  tool:
    name: default-tool
    description: Default tool
    argument_name: query
    argument_description: The default query
  llm:
    provider: openai
    model: gpt-3.5-turbo
    api_key: default-api-key
    prompt_template: "Default template with {{query}} and {{tools}}"
`
		minimalConfigPath := filepath.Join(tempDir, "default-config.yaml")
		err = os.WriteFile(minimalConfigPath, []byte(minimalConfigYAML), 0644)
		require.NoError(t, err)

		cm := NewConfigurationManager(logger)

		// Set the default config path and load
		err := cm.LoadConfiguration(context.Background(), minimalConfigPath)
		assert.NoError(t, err)

		// Check that default values were loaded
		assert.Equal(t, "default-agent", cm.config.Agent.Name)
		assert.Equal(t, "default-tool", cm.config.Agent.Tool.Name)
		assert.Equal(t, "Default tool", cm.config.Agent.Tool.Description)
		assert.Equal(t, "info", cm.config.Runtime.Log.RawLevel)
		assert.Equal(t, "stdout", cm.config.Runtime.Log.RawOutput)
		assert.Equal(t, "default-api-key", cm.config.Agent.LLM.APIKey)
	})

	// Test loading from an invalid file
	t.Run("Invalid file", func(t *testing.T) {
		cm := NewConfigurationManager(logger)

		// Load from non-existent file
		err := cm.LoadConfiguration(context.Background(), filepath.Join(tempDir, "non-existent.yaml"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration from file")
	})
}

// SetTestConfig sets the unexported config field in Manager for testing purposes only.
// This should never be used in production code.
func SetTestConfig(cm *Manager, cfg *types.Configuration) {
	v := reflect.ValueOf(cm).Elem()
	field := v.FieldByName("config")
	if !field.IsValid() || !field.CanSet() {
		panic("cannot set config field via reflection")
	}
	field.Set(reflect.ValueOf(cfg))
}
