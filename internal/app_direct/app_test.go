package app_direct

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

type mockAgent struct {
	callResult string
	callMeta   types.MetaInfo
	callErr    error
}

func (m *mockAgent) CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error) {
	return m.callResult, m.callMeta, m.callErr
}

func TestDirectApp_HandleCall_Success(t *testing.T) {
	app := &DirectApp{agent: &mockAgent{callResult: "ok", callMeta: types.MetaInfo{Tokens: 1}}}
	res := app.HandleCall(context.Background(), "test")
	if !res.Success {
		t.Errorf("expected success, got false")
	}
	if res.Result["answer"] != "ok" {
		t.Errorf("expected answer=ok, got %v", res.Result["answer"])
	}
	if res.Meta.Tokens != 1 {
		t.Errorf("expected tokens=1, got %d", res.Meta.Tokens)
	}
}

func TestDirectApp_HandleCall_Error(t *testing.T) {
	app := &DirectApp{agent: &mockAgent{callErr: errors.New("fail")}}
	res := app.HandleCall(context.Background(), "test")
	if res.Success {
		t.Errorf("expected error, got success")
	}
	if res.Error.Message == "" {
		t.Errorf("expected error message, got empty")
	}
}

func TestDirectApp_HandleCall_NotInitialized(t *testing.T) {
	app := &DirectApp{}
	res := app.HandleCall(context.Background(), "test")
	if res.Success {
		t.Errorf("expected not initialized error")
	}
	if res.Error.Type != "internal" {
		t.Errorf("expected internal error type")
	}
}

func TestDirectApp_JSONOutput(t *testing.T) {
	app := &DirectApp{agent: &mockAgent{callResult: "foo"}}
	res := app.HandleCall(context.Background(), "bar")
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	for _, f := range []string{"success", "result", "meta", "error"} {
		if _, ok := out[f]; !ok {
			t.Errorf("missing field %q in output: %s", f, string(b))
		}
	}
}

// --- MOCKS ---
type mockLogger struct{}

func (m *mockLogger) SetLevel(level logrus.Level)                                {}
func (m *mockLogger) SetFormatter(formatter logrus.Formatter)                    {}
func (m *mockLogger) Debug(args ...interface{})                                  {}
func (m *mockLogger) Debugf(format string, args ...interface{})                  {}
func (m *mockLogger) Info(args ...interface{})                                   {}
func (m *mockLogger) Infof(format string, args ...interface{})                   {}
func (m *mockLogger) Warn(args ...interface{})                                   {}
func (m *mockLogger) Warnf(format string, args ...interface{})                   {}
func (m *mockLogger) Error(args ...interface{})                                  {}
func (m *mockLogger) Errorf(format string, args ...interface{})                  {}
func (m *mockLogger) Fatal(args ...interface{})                                  {}
func (m *mockLogger) Fatalf(format string, args ...interface{})                  {}
func (m *mockLogger) WithField(key string, value interface{}) types.LogEntrySpec { return m }
func (m *mockLogger) WithFields(fields logrus.Fields) types.LogEntrySpec         { return m }
func (m *mockLogger) HandleMCPSetLevel(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

// mockConfigManager реализует types.ConfigurationManagerSpec
// Возвращает пустую конфигурацию для простоты
// Можно расширить для тестов ошибок

type mockConfigManager struct {
	getConfig func() *types.Configuration
	loadErr   error
}

func (m *mockConfigManager) LoadConfiguration(ctx context.Context, configFilePath string) error {
	return m.loadErr
}
func (m *mockConfigManager) GetConfiguration() *types.Configuration {
	if m.getConfig != nil {
		return m.getConfig()
	}
	cfg := &types.Configuration{}
	cfg.Agent.LLM.Provider = "openai"
	cfg.Agent.LLM.Model = "gpt-4"
	cfg.Agent.LLM.PromptTemplate = "test"
	cfg.Agent.Tool.Name = "test"
	cfg.Agent.Tool.ArgumentName = "input"
	cfg.Agent.LLM.APIKey = "dummy-key"
	cfg.Agent.Chat.MaxLLMIterations = 10
	cfg.Agent.Chat.MaxTokens = 1000
	return cfg
}

// --- TESTS FOR Initialize ---
func TestDirectApp_Initialize_Success(t *testing.T) {
	app := NewDirectApp(&mockLogger{}, &mockConfigManager{})
	err := app.Initialize(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.agent == nil {
		t.Fatalf("expected agent to be set")
	}
}

// --- TESTS FOR outputErrorAndExit ---
func TestDirectApp_outputErrorAndExit_Config(t *testing.T) {
	app := &DirectApp{}
	res, code, err := app.outputErrorAndExit("config", errors.New("fail config"))
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if err == nil || err.Error() != "fail config" {
		t.Errorf("expected error message 'fail config', got %v", err)
	}
	if res.Error.Message != "fail config" || res.Error.Type != "config" {
		t.Errorf("unexpected error struct: %+v", res.Error)
	}
}

func TestDirectApp_outputErrorAndExit_Internal(t *testing.T) {
	app := &DirectApp{}
	res, code, err := app.outputErrorAndExit("internal", errors.New("fail internal"))
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if err == nil || err.Error() != "fail internal" {
		t.Errorf("expected error message 'fail internal', got %v", err)
	}
	if res.Error.Message != "fail internal" || res.Error.Type != "internal" {
		t.Errorf("unexpected error struct: %+v", res.Error)
	}
}
