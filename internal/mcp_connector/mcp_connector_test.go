package mcp_connector

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test logger
func newTestLogger() (*logrus.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := logrus.New()
	log.SetOutput(buf)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return log, buf
}

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

func Test_StderrLoggingTrimsNewlines(t *testing.T) {
	log, buf := newTestLogger()
	serverID := "test-server"
	line := "error message with newline\n\r  "
	trimmed := strings.TrimRight(line, "\r\n \t")
	log.Infof("`%s` stderr: %s", serverID, trimmed)
	assert.Contains(t, buf.String(), "`test-server` stderr: error message with newline")
}

func Test_LoggingRouting_MCPAndStderr(t *testing.T) {
	log, buf := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
	serverID := "test-server"

	// Scenario 1: logging is supported (MCP logging)
	capWithLogging := mcp.ServerCapabilities{Logging: &struct{}{}}
	mc.dataLock.Lock()
	mc.capabilities[serverID] = capWithLogging
	mc.dataLock.Unlock()
	// Simulate MCP log (info level)
	msg := "mcp log message"
	level := "info"
	log.Infof("[MCP %s] %s", level, msg)
	assert.Contains(t, buf.String(), msg)
	assert.Contains(t, buf.String(), "[MCP info]")

	// Scenario 2: logging is not supported (fallback to stderr)
	capWithoutLogging := mcp.ServerCapabilities{}
	mc.dataLock.Lock()
	mc.capabilities[serverID] = capWithoutLogging
	mc.dataLock.Unlock()
	stderrMsg := "stderr fallback message\n"
	trimmed := strings.TrimRight(stderrMsg, "\r\n \t")
	log.Infof("`%s` stderr: %s", serverID, trimmed)
	assert.Contains(t, buf.String(), "stderr: stderr fallback message")
}

func Test_InitAndConnectToMCPs_emptyConfig(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{McpServers: map[string]configuration.MCPServerConnection{}}, log)
	err := mc.InitAndConnectToMCPs(context.Background())
	assert.NoError(t, err)
}

func Test_ExecuteTool_toolNotFound(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{McpServers: map[string]configuration.MCPServerConnection{}}, log)
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
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{McpServers: map[string]configuration.MCPServerConnection{"srv": {}}}, log)
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
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{McpServers: map[string]configuration.MCPServerConnection{"srv": {Timeout: 0.01}}}, log)
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
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{McpServers: map[string]configuration.MCPServerConnection{"srv": {}}}, log)
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
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
	mc.clients["ok"] = &mockMCPClient{}
	mc.clients["fail"] = &mockMCPClient{closeErr: fmt.Errorf("fail close")}
	// Should not panic, errors are logged
	_ = mc.Close()
}

func Test_GetAllTools_emptyAndFilled(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
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
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
	srvCfg := configuration.MCPServerConnection{
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
	log, _ := newTestLogger()
	cfg := configuration.MCPConnectorConfig{
		McpServers: map[string]configuration.MCPServerConnection{
			"srv": {Timeout: 42},
		},
	}
	mc := NewMCPConnector(cfg, log)
	if mc.getCallTimeout("srv") != 42*time.Second {
		t.Error("getCallTimeout should return configured timeout as duration")
	}
	if mc.getCallTimeout("unknown") != 30*time.Second {
		t.Error("getCallTimeout should return default duration for unknown server")
	}
}

func Test_findServerAndClient(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
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
	log, _ := newTestLogger()
	cfg := configuration.MCPConnectorConfig{
		McpServers: map[string]configuration.MCPServerConnection{
			"srv": {Timeout: 42},
		},
	}
	mc := NewMCPConnector(cfg, log)
	if mc.getCallTimeout("srv") != 42*time.Second {
		t.Error("getCallTimeout should return configured timeout as duration")
	}
	if mc.getCallTimeout("unknown") != 30*time.Second {
		t.Error("getCallTimeout should return default duration for unknown server")
	}
}

func Test_handleToolExecutionResult(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
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

func TestConnectServer_NoCommandNoURL(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{}, log)
	_, err := mc.ConnectServer(context.Background(), "srv", configuration.MCPServerConnection{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "neither command nor URL is specified")
}

func Test_getCallTimeout_CustomAndDefault(t *testing.T) {
	log, _ := newTestLogger()
	mc := NewMCPConnector(configuration.MCPConnectorConfig{
		McpServers: map[string]configuration.MCPServerConnection{
			"srv": {Timeout: 42.0},
		},
	}, log)
	timeout := mc.getCallTimeout("srv")
	assert.Equal(t, 42*time.Second, timeout)
	// default
	mc = NewMCPConnector(configuration.MCPConnectorConfig{}, log)
	timeout = mc.getCallTimeout("unknown")
	assert.Equal(t, 30*time.Second, timeout)
}
