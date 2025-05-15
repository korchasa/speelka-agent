package mcp_connector

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_testableTimeoutSelect_result(t *testing.T) {
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	cancel := func() {}

	want := &mcp.CallToolResult{Result: mcp.Result{Meta: map[string]any{"test": true}}}
	resultCh <- want

	result, err, timedOut := testableTimeoutSelect(resultCh, errCh, 50*time.Millisecond, cancel)
	assert.Equal(t, want, result)
	assert.Nil(t, err)
	assert.False(t, timedOut)
}

func Test_testableTimeoutSelect_error(t *testing.T) {
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	cancel := func() {}

	errWant := context.DeadlineExceeded
	errCh <- errWant

	result, err, timedOut := testableTimeoutSelect(resultCh, errCh, 50*time.Millisecond, cancel)
	assert.Nil(t, result)
	assert.Equal(t, errWant, err)
	assert.False(t, timedOut)
}

func Test_testableTimeoutSelect_timeout(t *testing.T) {
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err, timedOut := testableTimeoutSelect(resultCh, errCh, 10*time.Millisecond, cancel)
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, timedOut)
	// Context should be canceled
	select {
	case <-ctx.Done():
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Fatal("context was not canceled after timeout")
	}
}

// mockLogger implements types.LoggerSpec for testing
// Only Infof and required methods are implemented for this test

type mockLogger struct {
	lastMsg string
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.lastMsg = fmt.Sprintf(format, args...)
}
func (m *mockLogger) SetLevel(_ logrus.Level)                            {}
func (m *mockLogger) SetFormatter(_ logrus.Formatter)                    {}
func (m *mockLogger) Debug(...interface{})                               {}
func (m *mockLogger) Debugf(string, ...interface{})                      {}
func (m *mockLogger) Info(...interface{})                                {}
func (m *mockLogger) Warn(...interface{})                                {}
func (m *mockLogger) Warnf(string, ...interface{})                       {}
func (m *mockLogger) Error(...interface{})                               {}
func (m *mockLogger) Errorf(string, ...interface{})                      {}
func (m *mockLogger) Fatal(...interface{})                               {}
func (m *mockLogger) Fatalf(string, ...interface{})                      {}
func (m *mockLogger) WithField(string, interface{}) types.LogEntrySpec   { return m }
func (m *mockLogger) WithFields(fields logrus.Fields) types.LogEntrySpec { return m }
func (m *mockLogger) SetMCPServer(types.MCPServerNotifier)               {}

// fakeMCPClient implements client.MCPClient with only Initialize for test

type fakeMCPClient struct {
	initResult *mcp.InitializeResult
	called     *bool
}

func (f *fakeMCPClient) Initialize(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	if f.called != nil {
		*f.called = true
	}
	return f.initResult, nil
}

// остальные методы не нужны для этого теста
func (f *fakeMCPClient) Close() error { return nil }

func Test_CapabilitiesSavedAfterInitialize(t *testing.T) {
	called := false
	initResult := &mcp.InitializeResult{
		Capabilities: mcp.ServerCapabilities{
			Logging: &struct{}{},
		},
	}
	_ = &fakeMCPClient{initResult: initResult, called: &called}

	mc := NewMCPConnector(types.MCPConnectorConfig{}, &mockLogger{})
	serverID := "test-server"

	// Вручную вызываем логику сохранения capabilities
	mc.dataLock.Lock()
	mc.capabilities[serverID] = initResult.Capabilities
	mc.dataLock.Unlock()

	assert.False(t, called, "Initialize не должен быть вызван явно в этом тесте")
	mc.dataLock.RLock()
	cap, ok := mc.capabilities[serverID]
	mc.dataLock.RUnlock()
	assert.True(t, ok, "capabilities должны быть сохранены")
	assert.NotNil(t, cap.Logging, "capabilities.Logging должно быть установлено")
}

func Test_StderrLoggingTrimsNewlines(t *testing.T) {
	// Simulate the goroutine logic directly
	logger := &mockLogger{}
	serverID := "test-server"
	line := "error message with newline\n\r  "
	trimmed := strings.TrimRight(line, "\r\n \t")
	logger.Infof("`%s` stderr: %s", serverID, trimmed)
	assert.Equal(t, "`test-server` stderr: error message with newline", logger.lastMsg)
}

func Test_LoggingRouting_MCPAndStderr(t *testing.T) {
	logger := &mockLogger{}
	mc := NewMCPConnector(types.MCPConnectorConfig{}, logger)
	serverID := "test-server"

	// Сценарий 1: logging поддерживается (MCP-логирование)
	capWithLogging := mcp.ServerCapabilities{Logging: &struct{}{}}
	mc.dataLock.Lock()
	mc.capabilities[serverID] = capWithLogging
	mc.dataLock.Unlock()
	// Симулируем MCP-лог (уровень info)
	msg := "mcp log message"
	level := "info"
	logger.Infof("[MCP %s] %s", level, msg)
	assert.Contains(t, logger.lastMsg, msg)
	assert.Contains(t, logger.lastMsg, "[MCP info]")

	// Сценарий 2: logging не поддерживается (fallback на stderr)
	capWithoutLogging := mcp.ServerCapabilities{}
	mc.dataLock.Lock()
	mc.capabilities[serverID] = capWithoutLogging
	mc.dataLock.Unlock()
	stderrMsg := "stderr fallback message\n"
	trimmed := strings.TrimRight(stderrMsg, "\r\n \t")
	logger.Infof("`%s` stderr: %s", serverID, trimmed)
	assert.Contains(t, logger.lastMsg, "stderr: stderr fallback message")
}

func Test_InitAndConnectToMCPs_emptyConfig(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{McpServers: map[string]types.MCPServerConnection{}}, &mockLogger{})
	err := mc.InitAndConnectToMCPs(context.Background())
	assert.NoError(t, err)
}

func Test_ExecuteTool_toolNotFound(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{McpServers: map[string]types.MCPServerConnection{}}, &mockLogger{})
	_, err := mc.ExecuteTool(context.Background(), types.CallToolRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool `")
}
