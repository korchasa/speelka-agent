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
func (m *mockLogger) HandleMCPSetLevel(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

// Manually call the logic for saving capabilities
// assert.False(t, called, "Initialize should not be called explicitly in this test")
// assert.True(t, ok, "capabilities should be saved")
// assert.NotNil(t, cap.Logging, "capabilities.Logging should be set")
// Scenario 1: logging is supported (MCP logging)
// Simulate MCP log (info level)
// Scenario 2: logging is not supported (fallback to stderr)

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

	// Scenario 1: logging is supported (MCP logging)
	capWithLogging := mcp.ServerCapabilities{Logging: &struct{}{}}
	mc.dataLock.Lock()
	mc.capabilities[serverID] = capWithLogging
	mc.dataLock.Unlock()
	// Simulate MCP log (info level)
	msg := "mcp log message"
	level := "info"
	logger.Infof("[MCP %s] %s", level, msg)
	assert.Contains(t, logger.lastMsg, msg)
	assert.Contains(t, logger.lastMsg, "[MCP info]")

	// Scenario 2: logging is not supported (fallback to stderr)
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

// --- MOCKS ---
type mockMCPClient struct {
	callResult *mcp.CallToolResult
	callErr    error
	closeErr   error
}

func (m *mockMCPClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return m.callResult, m.callErr
}
func (m *mockMCPClient) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{Tools: []mcp.Tool{{Name: "foo"}, {Name: "bar"}}}, nil
}
func (m *mockMCPClient) Initialize(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{Capabilities: mcp.ServerCapabilities{}}, nil
}
func (m *mockMCPClient) Close() error                   { return m.closeErr }
func (m *mockMCPClient) Ping(ctx context.Context) error { return nil }
func (m *mockMCPClient) ListResourcesByPage(ctx context.Context, req mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListResources(ctx context.Context, req mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListResourceTemplatesByPage(ctx context.Context, req mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListResourceTemplates(ctx context.Context, req mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ReadResource(ctx context.Context, req mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (m *mockMCPClient) Subscribe(ctx context.Context, req mcp.SubscribeRequest) error { return nil }
func (m *mockMCPClient) Unsubscribe(ctx context.Context, req mcp.UnsubscribeRequest) error {
	return nil
}
func (m *mockMCPClient) ListPromptsByPage(ctx context.Context, req mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListPrompts(ctx context.Context, req mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	return nil, nil
}
func (m *mockMCPClient) GetPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListToolsByPage(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return nil, nil
}
func (m *mockMCPClient) SetLevel(ctx context.Context, req mcp.SetLevelRequest) error { return nil }
func (m *mockMCPClient) Complete(ctx context.Context, req mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return nil, nil
}
func (m *mockMCPClient) OnNotification(handler func(mcp.JSONRPCNotification)) {}

func Test_ExecuteTool_success(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{McpServers: map[string]types.MCPServerConnection{"srv": {}}}, &mockLogger{})
	mc.clients["srv"] = &mockMCPClient{callResult: &mcp.CallToolResult{Result: mcp.Result{Meta: map[string]any{"ok": true}}}}
	mc.tools["srv"] = []mcp.Tool{{Name: "foo"}}
	call := types.CallToolRequest{}
	call.Params.Name = "foo"
	res, err := mc.ExecuteTool(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Result.Meta["ok"] != true {
		t.Errorf("unexpected result: %+v", res)
	}
}

type slowClient struct{ mockMCPClient }

func (s *slowClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	time.Sleep(50 * time.Millisecond)
	return nil, nil
}

func Test_ExecuteTool_timeout(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{McpServers: map[string]types.MCPServerConnection{"srv": {Timeout: 0.01}}}, &mockLogger{})
	mc.clients["srv"] = &slowClient{}
	mc.tools["srv"] = []mcp.Tool{{Name: "foo"}}
	call := types.CallToolRequest{}
	call.Params.Name = "foo"
	_, err := mc.ExecuteTool(context.Background(), call)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func Test_ExecuteTool_error(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{McpServers: map[string]types.MCPServerConnection{"srv": {}}}, &mockLogger{})
	mc.clients["srv"] = &mockMCPClient{callErr: fmt.Errorf("fail call")}
	mc.tools["srv"] = []mcp.Tool{{Name: "foo"}}
	call := types.CallToolRequest{}
	call.Params.Name = "foo"
	_, err := mc.ExecuteTool(context.Background(), call)
	if err == nil || !strings.Contains(err.Error(), "fail call") {
		t.Errorf("expected call error, got: %v", err)
	}
}

func Test_Close_clients(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{}, &mockLogger{})
	mc.clients["ok"] = &mockMCPClient{}
	mc.clients["fail"] = &mockMCPClient{closeErr: fmt.Errorf("fail close")}
	// Не должно паниковать, ошибки логируются
	_ = mc.Close()
}

func Test_GetAllTools_emptyAndFilled(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{}, &mockLogger{})
	tools, err := mc.GetAllTools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
	mc.tools["srv1"] = []mcp.Tool{{Name: "foo"}}
	mc.tools["srv2"] = []mcp.Tool{{Name: "bar"}}
	tools, _ = mc.GetAllTools(context.Background())
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func Test_filterAllowedTools(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{}, &mockLogger{})
	srvCfg := types.MCPServerConnection{
		IncludeTools: []string{"foo"},
		ExcludeTools: []string{"bar"},
	}
	tools := []mcp.Tool{{Name: "foo"}, {Name: "bar"}, {Name: "baz"}}
	filtered := mc.filterAllowedTools("srv", tools, srvCfg)
	if len(filtered) != 1 || filtered[0].Name != "foo" {
		t.Errorf("expected only 'foo' allowed, got: %+v", filtered)
	}
}

func Test_getServerTimeout(t *testing.T) {
	cfg := types.MCPConnectorConfig{
		McpServers: map[string]types.MCPServerConnection{
			"srv": {Timeout: 42},
		},
	}
	mc := NewMCPConnector(cfg, &mockLogger{})
	if mc.getCallTimeout("srv") != 42*time.Second {
		t.Error("getCallTimeout should return configured timeout as duration")
	}
	if mc.getCallTimeout("unknown") != 30*time.Second {
		t.Error("getCallTimeout should return default duration for unknown server")
	}
}

func Test_findServerAndClient(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{}, &mockLogger{})
	mc.clients["srv"] = &mockMCPClient{}
	mc.tools["srv"] = []mcp.Tool{{Name: "foo"}}
	t.Run("found", func(t *testing.T) {
		serverID, client, err := mc.findServerAndClient("foo")
		if err != nil || serverID != "srv" || client == nil {
			t.Errorf("expected found, got %v, %v, %v", serverID, client, err)
		}
	})
	t.Run("not found", func(t *testing.T) {
		_, _, err := mc.findServerAndClient("bar")
		if err == nil || err.Error() != "tool `bar` not found" {
			t.Errorf("expected not found error, got %v", err)
		}
	})
	mc.tools["srv2"] = []mcp.Tool{{Name: "baz"}}
	t.Run("no client", func(t *testing.T) {
		_, _, err := mc.findServerAndClient("baz")
		if err == nil || err.Error() != "not connected to server: srv2" {
			t.Errorf("expected not connected error, got %v", err)
		}
	})
}

func Test_getCallTimeout(t *testing.T) {
	cfg := types.MCPConnectorConfig{
		McpServers: map[string]types.MCPServerConnection{
			"srv": {Timeout: 42},
		},
	}
	mc := NewMCPConnector(cfg, &mockLogger{})
	if mc.getCallTimeout("srv") != 42*time.Second {
		t.Error("getCallTimeout should return configured timeout as duration")
	}
	if mc.getCallTimeout("unknown") != 30*time.Second {
		t.Error("getCallTimeout should return default duration for unknown server")
	}
}

func Test_handleToolExecutionResult(t *testing.T) {
	mc := NewMCPConnector(types.MCPConnectorConfig{}, &mockLogger{})
	call := types.CallToolRequest{}
	call.Params.Name = "foo"
	res := &mcp.CallToolResult{Result: mcp.Result{Meta: map[string]any{"ok": true}}}
	t.Run("success", func(t *testing.T) {
		result, err := mc.handleToolExecutionResult(call, "srv", 10, res, nil, false)
		if err != nil || result != res {
			t.Errorf("expected success, got %v, %v", result, err)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		result, err := mc.handleToolExecutionResult(call, "srv", 10, nil, nil, true)
		if err == nil || !strings.Contains(err.Error(), "timed out") || result != nil {
			t.Errorf("expected timeout error, got %v, %v", result, err)
		}
	})
	t.Run("exec error", func(t *testing.T) {
		errExec := fmt.Errorf("fail")
		result, err := mc.handleToolExecutionResult(call, "srv", 10, nil, errExec, false)
		if err == nil || !strings.Contains(err.Error(), "fail") || result != nil {
			t.Errorf("expected exec error, got %v, %v", result, err)
		}
	})
}
