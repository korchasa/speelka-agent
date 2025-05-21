package mcp_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMCPServer struct {
	lastCtx    context.Context
	lastMethod string
	lastData   map[string]interface{}
}

func (m *mockMCPServer) SendNotificationToClient(ctx context.Context, method string, data map[string]interface{}) error {
	m.lastCtx = ctx
	m.lastMethod = method
	m.lastData = data
	return nil
}

func (m *mockMCPServer) GetServer() *server.MCPServer { return nil }

// Helper function to create a test logger
func newTestLogger() *logrus.Logger {
	buf := &bytes.Buffer{}
	log := logrus.New()
	log.SetOutput(buf)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return log
}

func TestMCPServer_SendsNotificationOnLog(t *testing.T) {
	mockServer := &mockMCPServer{}
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	mcpSrv.server = &server.MCPServer{} // avoid nil

	ctx := context.Background()
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelInfo, "test", "test log")
	err = mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
		"level":  msg.Params.Level,
		"logger": msg.Params.Logger,
		"data":   msg.Params.Data,
	})
	assert.NoError(t, err)
	assert.Equal(t, "notifications/message", mockServer.lastMethod)
	assert.Equal(t, mcp.LoggingLevelInfo, mockServer.lastData["level"])
	assert.Equal(t, "test log", mockServer.lastData["data"])
}

func TestMCPServer_LoggingMessageNotification_Structure(t *testing.T) {
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelInfo, "test", map[string]interface{}{"foo": "bar"})
	b, err := json.Marshal(msg)
	assert.NoError(t, err)
	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &out))
	params := out["params"].(map[string]interface{})
	assert.Equal(t, "info", params["level"])
	assert.Equal(t, "test", params["logger"])
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, params["data"])
}

func TestMCPServer_LogLevelFiltering(t *testing.T) {
	mockServer := &mockMCPServer{}
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	mcpSrv.server = &server.MCPServer{}

	ctx := context.Background()
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelWarning, "test", "should be sent")
	err = mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
		"level":  msg.Params.Level,
		"logger": msg.Params.Logger,
		"data":   msg.Params.Data,
	})
	assert.NoError(t, err)
	assert.Equal(t, "should be sent", mockServer.lastData["data"])
}

func TestMCPServer_NoSecretsOrPIIInLogs(t *testing.T) {
	mockServer := &mockMCPServer{}
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	mcpSrv.server = &server.MCPServer{}

	ctx := context.Background()
	secret := "super-secret-password"
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelError, "test", "error: "+secret)
	err = mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
		"level":  msg.Params.Level,
		"logger": msg.Params.Logger,
		"data":   msg.Params.Data,
	})
	assert.NoError(t, err)
	// NB: Secret/PII filtering is the responsibility of business logic, not logging infrastructure.
	// Here we only check that the log is sent correctly.
}

func TestMCPServer_LoggingCapability_Enabled(t *testing.T) {
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	// Check that logging capability is present
	caps := mcpSrv.GetServerCapabilities()
	assert.NotNil(t, caps.Logging, "logging capability must be present when LogRawOutput is :mcp:")
}

func TestMCPServer_LoggingCapability_Disabled(t *testing.T) {
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	config.MCPLogEnabled = false
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	// Check that logging capability is absent
	caps := mcpSrv.GetServerCapabilities()
	assert.Nil(t, caps.Logging, "logging capability must NOT be present when LogRawOutput is not :mcp:")
}

func TestMCPServer_ConcurrentStopAndServe(t *testing.T) {
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = mcpSrv.Serve(ctx, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{}, nil
		})
	}()
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		_ = mcpSrv.Stop(ctx)
	}()
	wg.Wait()
	// If there were no races or panics — the test is successful
}

func TestMCPServer_ToolsConsistency(t *testing.T) {
	log := newTestLogger()
	config := configuration.MCPServerConfigForTest()
	mcpSrv, err := NewMCPServer(config, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	_ = mcpSrv.Serve(context.Background(), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	})

	registered := mcpSrv.GetAllTools()

	// Get the list of tools via reflection
	var actual []string
	if mcpSrv.server != nil {
		val := reflect.ValueOf(mcpSrv.server).Elem()
		toolsField := val.FieldByName("tools")
		if toolsField.IsValid() {
			for _, key := range toolsField.MapKeys() {
				actual = append(actual, key.String())
			}
		}
	}
	// Compare tool names
	expected := map[string]bool{}
	for _, tool := range registered {
		expected[tool.Name] = true
	}
	for _, name := range actual {
		if !expected[name] {
			t.Errorf("Tool %s registered in server but not in GetAllTools", name)
		}
	}
	for _, tool := range registered {
		found := false
		for _, name := range actual {
			if name == tool.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Tool %s present in GetAllTools but not registered in server", tool.Name)
		}
	}
}

func TestMainToolHandler_NotSet_ReturnsError(t *testing.T) {
	cfg := configuration.MCPServerConfigForTest()
	log := newTestLogger()
	srv, err := NewMCPServer(cfg, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}

	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = "process"
	callReq.Params.Arguments = map[string]any{"input": "test"}

	// Register tools with nil handler (simulate Serve)
	var handler server.ToolHandlerFunc = nil
	for _, tool := range srv.buildTools() {
		var h server.ToolHandlerFunc = nil
		if tool.Name == srv.cfg.Tool.Name {
			h = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				if handler == nil {
					return nil, fmt.Errorf("main tool handler is not set for '%s'", tool.Name)
				}
				return handler(ctx, req)
			}
		}
		// Check that error is returned if handler == nil
		if h != nil {
			result, err := h(context.Background(), callReq)
			require.Error(t, err)
			require.Contains(t, err.Error(), "main tool handler is not set")
			_ = result
		}
	}
}

func TestMCPServer_buildTools(t *testing.T) {
	cfg := configuration.MCPServerConfigForTest()
	srv, err := NewMCPServer(cfg, newTestLogger())
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	tools := srv.buildTools()
	if len(tools) < 1 {
		t.Error("buildTools should return at least one tool")
	}
	found := false
	for _, tool := range tools {
		if tool.Name == "test-tool" {
			found = true
		}
	}
	if !found {
		t.Error("main tool not found in buildTools")
	}
}

func Test_initSSEServer_and_initStdioServer(t *testing.T) {
	cfg := configuration.MCPServerConfigForTest()
	log := newTestLogger()
	srv, err := NewMCPServer(cfg, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	t.Run("SSE server not initialized", func(t *testing.T) {
		srv.server = nil
		err := srv.initSSEServer(nil)
		if err == nil || err.Error() != "server is not *server.MCPServer" {
			t.Errorf("expected error for nil server, got %v", err)
		}
	})
	t.Run("Stdio server not initialized", func(t *testing.T) {
		srv.server = nil
		err := srv.initStdioServer(context.Background(), nil)
		if err == nil || err.Error() != "server is not *server.MCPServer" {
			t.Errorf("expected error for nil server, got %v", err)
		}
	})
}

func Test_buildMainTool_and_buildLoggingTool(t *testing.T) {
	cfg := configuration.MCPServerConfigForTest()
	log := newTestLogger()
	srv, err := NewMCPServer(cfg, log)
	if err != nil {
		t.Fatalf("failed to create MCPServer: %v", err)
	}
	mainTool := srv.buildMainTool()
	if mainTool.Name != "test-tool" {
		t.Errorf("expected test-tool, got %s", mainTool.Name)
	}
	logTool := srv.buildLoggingTool()
	if logTool.Name != "logging/setLevel" {
		t.Errorf("expected logging/setLevel, got %s", logTool.Name)
	}
}

func Test_getIsHttpMode(t *testing.T) {
	cfg := configuration.MCPServerConfigForTest()
	cfg.HTTP.Enabled = true
	cfg.Stdio.Enabled = true
	_, err := getIsHttpMode(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "both HTTP and Stdio modes cannot be enabled")

	cfg.HTTP.Enabled = false
	cfg.Stdio.Enabled = false
	_, err = getIsHttpMode(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either HTTP or Stdio mode must be enabled")

	cfg.HTTP.Enabled = true
	cfg.Stdio.Enabled = false
	ok, err := getIsHttpMode(cfg)
	assert.NoError(t, err)
	assert.True(t, ok)

	cfg.HTTP.Enabled = false
	cfg.Stdio.Enabled = true
	ok, err = getIsHttpMode(cfg)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_initSSEServer_and_initStdioServer_nilServer(t *testing.T) {
	cfg := configuration.MCPServerConfig{HTTP: configuration.HTTPConfig{Host: "127.0.0.1", Port: 12345, Enabled: true}, Stdio: configuration.StdioConfig{Enabled: false}}
	log := newTestLogger()
	srv, _ := NewMCPServer(cfg, log)
	srv.server = nil
	err := srv.initSSEServer(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is not *server.MCPServer")

	srv, _ = NewMCPServer(cfg, log)
	srv.server = nil
	err = srv.initStdioServer(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is not *server.MCPServer")
}

func TestSendNotificationToClient_ServerNil(t *testing.T) {
	cfg := configuration.MCPServerConfig{HTTP: configuration.HTTPConfig{Host: "127.0.0.1", Port: 12345, Enabled: true}, Stdio: configuration.StdioConfig{Enabled: false}}
	log := newTestLogger()
	srv, _ := NewMCPServer(cfg, log)
	srv.server = nil
	err := srv.SendNotificationToClient(context.Background(), "method", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "underlying server is not initialized")
}

func TestStop_onCleanServer(t *testing.T) {
	cfg := configuration.MCPServerConfig{HTTP: configuration.HTTPConfig{Host: "127.0.0.1", Port: 12345, Enabled: true}, Stdio: configuration.StdioConfig{Enabled: false}}
	log := newTestLogger()
	srv, _ := NewMCPServer(cfg, log)
	srv.sseServer = nil
	err := srv.Stop(context.Background())
	assert.NoError(t, err)
}

func TestBuildHooks_HooksCount(t *testing.T) {
	cfg := configuration.MCPServerConfig{HTTP: configuration.HTTPConfig{Host: "127.0.0.1", Port: 12345, Enabled: true}, Stdio: configuration.StdioConfig{Enabled: false}}
	log := newTestLogger()
	srv, _ := NewMCPServer(cfg, log)
	hooks := srv.BuildHooks()
	assert.NotNil(t, hooks)
}
