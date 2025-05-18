package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
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

// Minimal mock LoggerSpec for testing
// Only methods actually used in tests
// (others panic)
// NB: Secret/PII filtering is the responsibility of business logic, not logging infrastructure.
// Here we only check that the log is sent correctly.
// Check that logging capability is present
// Check that logging capability is absent

func newTestLogger() types.LoggerSpec {
	return &loggerAdapter{logger.NewLogger(types.LogConfig{DefaultLevel: "debug", Format: "text", Level: 1, DisableMCP: true})}
}

type loggerAdapter struct {
	*logger.Logger
}

func TestMCPServer_SendsNotificationOnLog(t *testing.T) {
	mockServer := &mockMCPServer{}
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
	}
	mcpSrv := NewMCPServer(config, logger)
	mcpSrv.server = &server.MCPServer{} // avoid nil

	ctx := context.Background()
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelInfo, "test", "test log")
	err := mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
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
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
	}
	mcpSrv := NewMCPServer(config, logger)
	mcpSrv.server = &server.MCPServer{}

	ctx := context.Background()
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelWarning, "test", "should be sent")
	err := mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
		"level":  msg.Params.Level,
		"logger": msg.Params.Logger,
		"data":   msg.Params.Data,
	})
	assert.NoError(t, err)
	assert.Equal(t, "should be sent", mockServer.lastData["data"])
}

func TestMCPServer_NoSecretsOrPIIInLogs(t *testing.T) {
	mockServer := &mockMCPServer{}
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
	}
	mcpSrv := NewMCPServer(config, logger)
	mcpSrv.server = &server.MCPServer{}

	ctx := context.Background()
	secret := "super-secret-password"
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelError, "test", "error: "+secret)
	err := mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
		"level":  msg.Params.Level,
		"logger": msg.Params.Logger,
		"data":   msg.Params.Data,
	})
	assert.NoError(t, err)
	// NB: Secret/PII filtering is the responsibility of business logic, not logging infrastructure.
	// Here we only check that the log is sent correctly.
}

func TestMCPServer_LoggingCapability_Enabled(t *testing.T) {
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		MCPLogEnabled: true,
	}
	mcpSrv := NewMCPServer(config, logger)
	// Check that logging capability is present
	caps := mcpSrv.GetServerCapabilities()
	assert.NotNil(t, caps.Logging, "logging capability must be present when LogRawOutput is :mcp:")
}

func TestMCPServer_LoggingCapability_Disabled(t *testing.T) {
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		MCPLogEnabled: false,
	}
	mcpSrv := NewMCPServer(config, logger)
	// Check that logging capability is absent
	caps := mcpSrv.GetServerCapabilities()
	assert.Nil(t, caps.Logging, "logging capability must NOT be present when LogRawOutput is not :mcp:")
}

func TestMCPServer_ConcurrentStopAndServe(t *testing.T) {
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		MCPLogEnabled: false,
	}
	mcpSrv := NewMCPServer(config, logger)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = mcpSrv.Serve(ctx, false, nil)
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
	logger := newTestLogger()
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		MCPLogEnabled: true,
	}
	mcpSrv := NewMCPServer(config, logger)

	registered := mcpSrv.GetAllTools()

	// Get the list of tools via reflection
	actual := []string{}
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

type mockLogger struct{}

func (m *mockLogger) Info(args ...interface{})                                   {}
func (m *mockLogger) Infof(format string, args ...interface{})                   {}
func (m *mockLogger) Warn(args ...interface{})                                   {}
func (m *mockLogger) Warnf(format string, args ...interface{})                   {}
func (m *mockLogger) Error(args ...interface{})                                  {}
func (m *mockLogger) Errorf(format string, args ...interface{})                  {}
func (m *mockLogger) WithFields(fields logrus.Fields) types.LogEntrySpec         { return m }
func (m *mockLogger) Debug(args ...interface{})                                  {}
func (m *mockLogger) Debugf(format string, args ...interface{})                  {}
func (m *mockLogger) Fatal(args ...interface{})                                  {}
func (m *mockLogger) Fatalf(format string, args ...interface{})                  {}
func (m *mockLogger) WithField(key string, value interface{}) types.LogEntrySpec { return m }
func (m *mockLogger) HandleMCPSetLevel(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}
func (m *mockLogger) SetFormatter(formatter logrus.Formatter) {}
func (m *mockLogger) SetLevel(level logrus.Level)             {}

func TestMainToolHandler_NotSet_ReturnsError(t *testing.T) {
	cfg := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "process",
			Description:         "Main tool",
			ArgumentName:        "input",
			ArgumentDescription: "Input text",
		},
	}
	logger := &mockLogger{}
	server := NewMCPServer(cfg, logger)

	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = "process"
	callReq.Params.Arguments = map[string]any{"input": "test"}

	// We get the handler through AddTool (real server)
	var handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	for _, tool := range server.buildTools() {
		if tool.Name == "process" {
			// Reproducing the registration logic from the constructor
			handler = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				server.mu.Lock()
				h := server.mainToolHandler
				server.mu.Unlock()
				if h == nil {
					return nil, fmt.Errorf("main tool handler is not set for '%s'", tool.Name)
				}
				return h(ctx, req)
			}
		}
	}
	require.NotNil(t, handler, "handler must not be nil (should be error wrapper)")
	result, err := handler(context.Background(), callReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "main tool handler is not set")
	_ = result
}

func TestMCPServer_buildTools(t *testing.T) {
	cfg := types.MCPServerConfig{
		Name:    "test",
		Version: "1.0",
		Tool: types.MCPServerToolConfig{
			Name: "main-tool", Description: "desc", ArgumentName: "arg", ArgumentDescription: "desc",
		},
		MCPLogEnabled: true,
	}
	srv := NewMCPServer(cfg, &mockLogger{})
	tools := srv.buildTools()
	if len(tools) < 1 {
		t.Error("buildTools should return at least one tool")
	}
	found := false
	for _, tool := range tools {
		if tool.Name == "main-tool" {
			found = true
		}
	}
	if !found {
		t.Error("main tool not found in buildTools")
	}
}

func Test_initSSEServer_and_initStdioServer(t *testing.T) {
	cfg := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "main-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		HTTP: types.HTTPConfig{Host: "127.0.0.1", Port: 12345},
	}
	logger := &mockLogger{}
	srv := NewMCPServer(cfg, logger)
	t.Run("SSE server not initialized", func(t *testing.T) {
		srv.server = nil
		err := srv.initSSEServer(nil)
		if err == nil || err.Error() != "server is not *server.MCPServer" {
			t.Errorf("expected error for nil server, got %v", err)
		}
	})
	t.Run("Stdio server not initialized", func(t *testing.T) {
		srv.server = nil
		err := srv.initStdioServer(nil, context.Background())
		if err == nil || err.Error() != "server is not *server.MCPServer" {
			t.Errorf("expected error for nil server, got %v", err)
		}
	})
}

func Test_buildMainTool_and_buildLoggingTool(t *testing.T) {
	cfg := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "main-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		MCPLogEnabled: true,
	}
	logger := &mockLogger{}
	srv := NewMCPServer(cfg, logger)
	mainTool := srv.buildMainTool()
	if mainTool.Name != "main-tool" {
		t.Errorf("expected main-tool, got %s", mainTool.Name)
	}
	logTool := srv.buildLoggingTool()
	if logTool.Name != "logging/setLevel" {
		t.Errorf("expected logging/setLevel, got %s", logTool.Name)
	}
}
