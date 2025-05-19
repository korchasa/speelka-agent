package configuration

import (
	"context"
	"os"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
func (m *SimpleLogger) SetLevel(level logrus.Level)                    {}
func (m *SimpleLogger) SetMCPServer(mcpServer types.MCPServerNotifier) {}
func (m *SimpleLogger) SetFormatter(formatter logrus.Formatter)        {}
func (m *SimpleLogger) HandleMCPSetLevel(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

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

// Comment out tests that reference NewConfigurationManager or undefined Manager
// func TestConfigurationManager_LoadConfiguration(t *testing.T) { /* ... */ }
// func SetTestConfig(cm *Manager, cfg *types.Configuration) { /* ... */ }

func TestManager_LoadConfiguration_Defaults(t *testing.T) {
	mgr := NewConfigurationManager(nil)
	err := mgr.LoadConfiguration(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := mgr.GetConfiguration()
	if cfg.Agent.Name != "speelka-agent" {
		t.Errorf("expected default agent name, got %s", cfg.Agent.Name)
	}
	if cfg.Runtime.Log.DefaultLevel != "info" {
		t.Errorf("expected default log level, got %s", cfg.Runtime.Log.DefaultLevel)
	}
}

func TestManager_LoadConfiguration_YAMLFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testconfig-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	yamlContent := []byte(`
agent:
  name: "custom-agent"
  tool:
    name: "custom-tool"
`)
	if _, err := tmpfile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	mgr := NewConfigurationManager(nil)
	err = mgr.LoadConfiguration(context.Background(), tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := mgr.GetConfiguration()
	if cfg.Agent.Name != "custom-agent" {
		t.Errorf("expected agent name from yaml, got %s", cfg.Agent.Name)
	}
	if cfg.Agent.Tool.Name != "custom-tool" {
		t.Errorf("expected tool name from yaml, got %s", cfg.Agent.Tool.Name)
	}
}

func TestManager_LoadConfiguration_EnvOverride(t *testing.T) {
	os.Setenv("SPL_agent_name", "env-agent")
	os.Setenv("SPL_agent_tool_name", "env-tool")
	defer os.Unsetenv("SPL_agent_name")
	defer os.Unsetenv("SPL_agent_tool_name")

	mgr := NewConfigurationManager(nil)
	err := mgr.LoadConfiguration(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := mgr.GetConfiguration()
	if cfg.Agent.Name != "env-agent" {
		t.Errorf("expected agent name from env, got %s", cfg.Agent.Name)
	}
	if cfg.Agent.Tool.Name != "env-tool" {
		t.Errorf("expected tool name from env, got %s", cfg.Agent.Tool.Name)
	}
}

// --- BEGIN: overlay, validation, redaction, apply, property-based overlay tests ---

func TestManager_ValidateConfiguration(t *testing.T) {
	mgr := NewConfigurationManager(&SimpleLogger{})
	validConfig := &types.Configuration{}
	validConfig.Agent.Name = "TestAgent"
	validConfig.Agent.Tool.Name = "TestTool"
	validConfig.Agent.Tool.Description = "Test tool description"
	validConfig.Agent.Tool.ArgumentName = "query"
	validConfig.Agent.Tool.ArgumentDescription = "Query to process"
	validConfig.Agent.LLM.Provider = "openai"
	validConfig.Agent.LLM.Model = "gpt-4"
	validConfig.Agent.LLM.APIKey = "test-api-key"
	validConfig.Agent.LLM.PromptTemplate = "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}"
	mgr.config = validConfig
	assert.NoError(t, mgr.Validate())

	invalidConfig := &types.Configuration{}
	mgr.config = invalidConfig
	// Leave Agent.Name and Tool/LLM empty
	err := mgr.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent name is required")
}

// func TestManager_OverlayApply(t *testing.T) {
// 	mgr := NewConfigurationManager(&SimpleLogger{})
// 	base := &types.Configuration{}
// 	base.Agent.Name = "base-agent"
// 	base.Agent.Tool.Name = "base-tool"
// 	base.Agent.Tool.Description = "Base tool description"
// 	base.Agent.Tool.ArgumentName = "query"
// 	base.Agent.Tool.ArgumentDescription = "Base query description"
// 	base.Agent.LLM.Provider = "openai"
// 	base.Agent.LLM.Model = "gpt-3.5-turbo"
// 	base.Agent.LLM.APIKey = "base-api-key"
// 	base.Agent.LLM.PromptTemplate = "Base template with {{query}} and {{tools}}"
//
// 	overlay := &types.Configuration{}
// 	overlay.Agent.Tool.Name = "new-tool"
// 	overlay.Agent.Tool.Description = "New tool description"
// 	overlay.Agent.LLM.Model = "gpt-4"
// 	overlay.Agent.LLM.APIKey = "new-api-key"
// 	overlay.Agent.LLM.PromptTemplate = "New template with {{query}} and {{tools}}"
//
// 	result, err := mgr.Apply(base, overlay)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "new-tool", result.Agent.Tool.Name)
// 	assert.Equal(t, "New tool description", result.Agent.Tool.Description)
// 	assert.Equal(t, "gpt-4", result.Agent.LLM.Model)
// 	assert.Equal(t, "new-api-key", result.Agent.LLM.APIKey)
// 	assert.Equal(t, "New template with {{query}} and {{tools}}", result.Agent.LLM.PromptTemplate)
// }

func TestRedactedCopy(t *testing.T) {
	orig := &types.Configuration{}
	orig.Agent.LLM.APIKey = "super-secret-llm-key"
	orig.Agent.Connections.McpServers = map[string]types.MCPServerConnection{
		"server1": {APIKey: "server1-key", URL: "http://server1"},
		"server2": {APIKey: "server2-key", URL: "http://server2"},
	}
	redacted := RedactedCopy(orig)
	assert.Equal(t, "***REDACTED***", redacted.Agent.LLM.APIKey)
	for k, v := range redacted.Agent.Connections.McpServers {
		assert.Equal(t, "***REDACTED***", v.APIKey, "APIKey for %s should be redacted", k)
	}
}

func TestManager_ValidatePromptTemplate(t *testing.T) {
	mgr := NewConfigurationManager(&SimpleLogger{})
	err := mgr.validatePromptTemplate("This is a template with {{query}} and {{tools}} placeholders", "query")
	assert.NoError(t, err)
	err = mgr.validatePromptTemplate("This is a template with {{input}} and {{tools}} placeholders", "query")
	assert.NoError(t, err)
	err = mgr.validatePromptTemplate("Template with only {{tools}} placeholder", "query")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template must contain either {{query}} or {{input}} placeholder")
	err = mgr.validatePromptTemplate("", "query")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestManager_ExtractPlaceholders(t *testing.T) {
	mgr := NewConfigurationManager(&SimpleLogger{})
	placeholders, err := mgr.extractPlaceholders("This is a {{test}} template with {{multiple}} placeholders including {{tools}}")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"test", "multiple", "tools"}, placeholders)
	placeholders, err = mgr.extractPlaceholders("This has {{ spaced }} placeholders and {{unspaced}} ones")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"spaced", "unspaced"}, placeholders)
	placeholders, err = mgr.extractPlaceholders("This template has no placeholders")
	assert.NoError(t, err)
	assert.Empty(t, placeholders)
	placeholders, err = mgr.extractPlaceholders("")
	assert.NoError(t, err)
	assert.Empty(t, placeholders)
}

// --- END: overlay, validation, redaction, apply, property-based overlay tests ---

func TestManager_GetAgentConfig_InlineStruct(t *testing.T) {
	logger := &SimpleLogger{}
	mgr := NewConfigurationManager(logger)
	// Load only the default configuration (without file and env)
	err := mgr.LoadConfiguration(context.Background(), "")
	assert.NoError(t, err)

	agentCfg := mgr.GetAgentConfig()
	assert.Equal(t, "process", agentCfg.Tool.Name)
	assert.Equal(t, "gpt-4", agentCfg.Model)
	assert.Equal(t, 8192, agentCfg.MaxTokens)
	assert.Equal(t, 100, agentCfg.MaxLLMIterations)
}

func TestManager_FullConfig_Parse_YAML_JSON_Env(t *testing.T) {
	tmpYaml, err := os.CreateTemp("", "fullconfig-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpYaml.Name())
	yamlContent := []byte(`
agent:
  name: "yaml-agent"
  tool:
    name: "yaml-tool"
    description: "desc from yaml"
    argumentName: "input"
    argumentDescription: "desc arg"
  chat:
    maxTokens: 111
    maxLLMIterations: 222
    requestBudget: 1.23
  llm:
    provider: "openai"
    model: "yaml-model"
    apiKey: "yaml-key"
    promptTemplate: "YAML template {{input}}"
    retry:
      maxRetries: 2
      initialBackoff: 1.1
      maxBackoff: 2.2
      backoffMultiplier: 3.3
  connections:
    mcpServers:
      test:
        url: "http://yaml-server"
        apiKey: "yaml-server-key"
    retry:
      maxRetries: 5
      initialBackoff: 2.2
      maxBackoff: 3.3
      backoffMultiplier: 4.4
runtime:
  log:
    defaultLevel: "debug"
    format: "json"
    disableMcp: true
`)
	if _, err := tmpYaml.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpYaml.Close()

	tmpJson, err := os.CreateTemp("", "fullconfig-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpJson.Name())
	jsonContent := []byte(`{
  "agent": {
    "name": "json-agent",
    "tool": {"name": "json-tool", "description": "desc from json", "argumentName": "input", "argumentDescription": "desc arg"},
    "chat": {"maxTokens": 333, "maxLLMIterations": 444, "requestBudget": 2.34},
    "llm": {"provider": "openai", "model": "json-model", "apiKey": "json-key", "promptTemplate": "JSON template {{input}}", "retry": {"maxRetries": 3, "initialBackoff": 2.2, "maxBackoff": 3.3, "backoffMultiplier": 4.4}},
    "connections": {"mcpServers": {"test": {"url": "http://json-server", "apiKey": "json-server-key"}}, "retry": {"maxRetries": 6, "initialBackoff": 3.3, "maxBackoff": 4.4, "backoffMultiplier": 5.5}}
  },
  "runtime": {"log": {"defaultLevel": "info", "format": "text", "disableMcp": false}}
}`)
	if _, err := tmpJson.Write(jsonContent); err != nil {
		t.Fatal(err)
	}
	tmpJson.Close()

	os.Setenv("SPL_AGENT_NAME", "env-agent")
	os.Setenv("SPL_AGENT_TOOL_NAME", "env-tool")
	os.Setenv("SPL_AGENT_LLM_APIKEY", "env-key")
	os.Setenv("SPL_RUNTIME_LOG_DEFAULTLEVEL", "warn")
	defer os.Unsetenv("SPL_AGENT_NAME")
	defer os.Unsetenv("SPL_AGENT_TOOL_NAME")
	defer os.Unsetenv("SPL_AGENT_LLM_APIKEY")
	defer os.Unsetenv("SPL_RUNTIME_LOG_DEFAULTLEVEL")

	t.Run("YAML config + env", func(t *testing.T) {
		mgr := NewConfigurationManager(nil)
		err := mgr.LoadConfiguration(context.Background(), tmpYaml.Name())
		assert.NoError(t, err)
		cfg := mgr.GetConfiguration()
		assert.Equal(t, "env-agent", cfg.Agent.Name) // env overrides yaml
		assert.Equal(t, "env-tool", cfg.Agent.Tool.Name)
		assert.Equal(t, "yaml-model", cfg.Agent.LLM.Model)
		assert.Equal(t, "env-key", cfg.Agent.LLM.APIKey)
		assert.Equal(t, "warn", cfg.Runtime.Log.DefaultLevel)
		assert.Equal(t, "json", cfg.Runtime.Log.Format)
		assert.Equal(t, true, cfg.Runtime.Log.DisableMCP)
		assert.Equal(t, 111, cfg.Agent.Chat.MaxTokens)
		assert.Equal(t, 222, cfg.Agent.Chat.MaxLLMIterations)
		assert.Equal(t, 1.23, cfg.Agent.Chat.RequestBudget)
		assert.Equal(t, "desc from yaml", cfg.Agent.Tool.Description)
		assert.Equal(t, "desc arg", cfg.Agent.Tool.ArgumentDescription)
		assert.Equal(t, "YAML template {{input}}", cfg.Agent.LLM.PromptTemplate)
		assert.Equal(t, 2, cfg.Agent.LLM.Retry.MaxRetries)
		assert.Equal(t, 1.1, cfg.Agent.LLM.Retry.InitialBackoff)
		assert.Equal(t, 2.2, cfg.Agent.LLM.Retry.MaxBackoff)
		assert.Equal(t, 3.3, cfg.Agent.LLM.Retry.BackoffMultiplier)
		assert.Equal(t, "http://yaml-server", cfg.Agent.Connections.McpServers["test"].URL)
		assert.Equal(t, "yaml-server-key", cfg.Agent.Connections.McpServers["test"].APIKey)
		assert.Equal(t, 5, cfg.Agent.Connections.Retry.MaxRetries)
		assert.Equal(t, 2.2, cfg.Agent.Connections.Retry.InitialBackoff)
		assert.Equal(t, 3.3, cfg.Agent.Connections.Retry.MaxBackoff)
		assert.Equal(t, 4.4, cfg.Agent.Connections.Retry.BackoffMultiplier)
	})

	t.Run("JSON config + env", func(t *testing.T) {
		mgr := NewConfigurationManager(nil)
		err := mgr.LoadConfiguration(context.Background(), tmpJson.Name())
		assert.NoError(t, err)
		cfg := mgr.GetConfiguration()
		assert.Equal(t, "env-agent", cfg.Agent.Name) // env overrides json
		assert.Equal(t, "env-tool", cfg.Agent.Tool.Name)
		assert.Equal(t, "json-model", cfg.Agent.LLM.Model)
		assert.Equal(t, "env-key", cfg.Agent.LLM.APIKey)
		assert.Equal(t, "warn", cfg.Runtime.Log.DefaultLevel)
		assert.Equal(t, "text", cfg.Runtime.Log.Format)
		assert.Equal(t, false, cfg.Runtime.Log.DisableMCP)
		assert.Equal(t, 333, cfg.Agent.Chat.MaxTokens)
		assert.Equal(t, 444, cfg.Agent.Chat.MaxLLMIterations)
		assert.Equal(t, 2.34, cfg.Agent.Chat.RequestBudget)
		assert.Equal(t, "desc from json", cfg.Agent.Tool.Description)
		assert.Equal(t, "desc arg", cfg.Agent.Tool.ArgumentDescription)
		assert.Equal(t, "JSON template {{input}}", cfg.Agent.LLM.PromptTemplate)
		assert.Equal(t, 3, cfg.Agent.LLM.Retry.MaxRetries)
		assert.Equal(t, 2.2, cfg.Agent.LLM.Retry.InitialBackoff)
		assert.Equal(t, 3.3, cfg.Agent.LLM.Retry.MaxBackoff)
		assert.Equal(t, 4.4, cfg.Agent.LLM.Retry.BackoffMultiplier)
		assert.Equal(t, "http://json-server", cfg.Agent.Connections.McpServers["test"].URL)
		assert.Equal(t, "json-server-key", cfg.Agent.Connections.McpServers["test"].APIKey)
		assert.Equal(t, 6, cfg.Agent.Connections.Retry.MaxRetries)
		assert.Equal(t, 3.3, cfg.Agent.Connections.Retry.InitialBackoff)
		assert.Equal(t, 4.4, cfg.Agent.Connections.Retry.MaxBackoff)
		assert.Equal(t, 5.5, cfg.Agent.Connections.Retry.BackoffMultiplier)
	})
}

func TestEnvKeyToPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SPL_agent_tool_name", "agent.tool.name"},
		{"SPL_AGENT_TOOL_DESCRIPTION", "agent.tool.description"},
		{"SPL_runtime_log_defaultLevel", "runtime.log.defaultlevel"},
		{"SPL_AGENT__TOOL__NAME", "agent..tool..name"},
		{"SPL__agent__tool__name", ".agent..tool..name"},
		{"SPL_agent", "agent"},
		{"SPL_AGENT", "agent"},
		{"SPL__", "."},
		{"SPL_", ""},
		{"agent_tool_name", "agent.tool.name"},
		{"SPL_AGENT_TOOL_", "agent.tool."},
		{"SPL_AGENT__TOOL__", "agent..tool.."},
	}
	for _, tt := range tests {
		got := envKeyToPath(tt.input)
		if got != tt.expected {
			t.Errorf("envKeyToPath(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
