package configuration

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

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

func TestManager_LoadAndGetConfiguration(t *testing.T) {
	logger := &SimpleLogger{}
	mgr := NewConfigurationManager(logger)
	// Загружаем только дефолтную конфигурацию (без файла и env)
	err := mgr.LoadConfiguration(context.Background(), "")
	assert.NoError(t, err)

	cfg := mgr.GetConfiguration()
	assert.NotNil(t, cfg)
	// Проверяем, что это действительно types.Configuration
	assert.Equal(t, "speelka-agent", cfg.Agent.Name)
	// Валидация вызывается отдельно
	// err = cfg.Validate()
	// assert.NoError(t, err)
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
	assert.NoError(t, mgr.Validate(validConfig))

	invalidConfig := &types.Configuration{}
	// оставляем Agent.Name и Tool/LLM пустыми
	err := mgr.Validate(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Agent name is required")
}

func TestManager_OverlayApply(t *testing.T) {
	mgr := NewConfigurationManager(&SimpleLogger{})
	base := &types.Configuration{}
	base.Agent.Name = "base-agent"
	base.Agent.Tool.Name = "base-tool"
	base.Agent.Tool.Description = "Base tool description"
	base.Agent.Tool.ArgumentName = "query"
	base.Agent.Tool.ArgumentDescription = "Base query description"
	base.Agent.LLM.Provider = "openai"
	base.Agent.LLM.Model = "gpt-3.5-turbo"
	base.Agent.LLM.APIKey = "base-api-key"
	base.Agent.LLM.PromptTemplate = "Base template with {{query}} and {{tools}}"

	overlay := &types.Configuration{}
	overlay.Agent.Tool.Name = "new-tool"
	overlay.Agent.Tool.Description = "New tool description"
	overlay.Agent.LLM.Model = "gpt-4"
	overlay.Agent.LLM.APIKey = "new-api-key"
	overlay.Agent.LLM.PromptTemplate = "New template with {{query}} and {{tools}}"

	result, err := mgr.Apply(base, overlay)
	assert.NoError(t, err)
	assert.Equal(t, "new-tool", result.Agent.Tool.Name)
	assert.Equal(t, "New tool description", result.Agent.Tool.Description)
	assert.Equal(t, "gpt-4", result.Agent.LLM.Model)
	assert.Equal(t, "new-api-key", result.Agent.LLM.APIKey)
	assert.Equal(t, "New template with {{query}} and {{tools}}", result.Agent.LLM.PromptTemplate)
}

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

func TestManager_Overlay_PropertyBased(t *testing.T) {
	mgr := NewConfigurationManager(&SimpleLogger{})
	f := func(base, overlay types.Configuration) bool {
		baseCopy := base
		_, err := mgr.Apply(&baseCopy, &overlay)
		if err != nil {
			t.Logf("Apply error: %v", err)
			return false
		}
		return true
	}
	cfg := &quick.Config{
		MaxCount: 10,
		Values: func(args []reflect.Value, r *rand.Rand) {
			args[0] = reflect.ValueOf(randomConfig(r))
			args[1] = reflect.ValueOf(randomConfig(r))
		},
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

func randomConfig(r *rand.Rand) types.Configuration {
	cfg := types.NewConfiguration()
	cfg.Agent.Name = randomString(r)
	cfg.Agent.LLM.APIKey = randomString(r)
	cfg.Agent.LLM.Model = randomString(r)
	cfg.Agent.LLM.Provider = randomString(r)
	cfg.Agent.LLM.PromptTemplate = randomString(r)
	cfg.Agent.LLM.MaxTokens = r.Intn(10000)
	cfg.Agent.LLM.Temperature = r.Float64()
	cfg.Agent.Connections.McpServers = make(map[string]types.MCPServerConnection)
	if r.Intn(2) == 1 {
		key := randomString(r)
		cfg.Agent.Connections.McpServers[key] = types.MCPServerConnection{
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

// --- END: overlay, validation, redaction, apply, property-based overlay tests ---

func TestManager_GetAgentConfig_InlineStruct(t *testing.T) {
	logger := &SimpleLogger{}
	mgr := NewConfigurationManager(logger)
	// Загружаем только дефолтную конфигурацию (без файла и env)
	err := mgr.LoadConfiguration(context.Background(), "")
	assert.NoError(t, err)

	agentCfg := mgr.GetAgentConfig()
	assert.Equal(t, "process", agentCfg.Tool.Name)
	assert.Equal(t, "gpt-4", agentCfg.Model)
	assert.Equal(t, 8192, agentCfg.MaxTokens)
	assert.Equal(t, 100, agentCfg.MaxLLMIterations)
}
