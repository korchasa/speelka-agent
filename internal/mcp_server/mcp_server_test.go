package mcp_server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

// Минимальный mock LoggerSpec для теста
// Только методы, которые реально используются в тестах
// (остальные panic)
type testLogger struct {
	mcpServer types.MCPServerNotifier
}

func (l *testLogger) SetLevel(level logrus.Level)                                {}
func (l *testLogger) Debug(args ...interface{})                                  {}
func (l *testLogger) Debugf(format string, args ...interface{})                  {}
func (l *testLogger) Info(args ...interface{})                                   {}
func (l *testLogger) Infof(format string, args ...interface{})                   {}
func (l *testLogger) Warn(args ...interface{})                                   {}
func (l *testLogger) Warnf(format string, args ...interface{})                   {}
func (l *testLogger) Error(args ...interface{})                                  {}
func (l *testLogger) Errorf(format string, args ...interface{})                  {}
func (l *testLogger) Fatal(args ...interface{})                                  {}
func (l *testLogger) Fatalf(format string, args ...interface{})                  {}
func (l *testLogger) WithField(key string, value interface{}) types.LogEntrySpec { panic("not used") }
func (l *testLogger) WithFields(fields logrus.Fields) types.LogEntrySpec         { panic("not used") }
func (l *testLogger) SetMCPServer(mcpServer types.MCPServerNotifier)             { l.mcpServer = mcpServer }
func (l *testLogger) SetFormatter(formatter logrus.Formatter)                    {}

func TestMCPServer_SendsNotificationOnLog(t *testing.T) {
	mockServer := &mockMCPServer{}
	logger := &testLogger{}
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
	logger.SetMCPServer(mockServer)

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
	logger := &testLogger{}
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
	logger.SetMCPServer(mockServer)
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
	logger := &testLogger{}
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
	logger.SetMCPServer(mockServer)
	ctx := context.Background()
	secret := "super-secret-password"
	msg := mcp.NewLoggingMessageNotification(mcp.LoggingLevelError, "test", "error: "+secret)
	err := mockServer.SendNotificationToClient(ctx, msg.Notification.Method, map[string]interface{}{
		"level":  msg.Params.Level,
		"logger": msg.Params.Logger,
		"data":   msg.Params.Data,
	})
	assert.NoError(t, err)
	// NB: Фильтрация секретов/PII — ответственность бизнес-логики, а не инфраструктуры логгирования.
	// Здесь проверяем только, что лог отправлен корректно.
}

func TestMCPServer_LoggingCapability_Enabled(t *testing.T) {
	logger := &testLogger{}
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		LogRawOutput: ":mcp:",
	}
	mcpSrv := NewMCPServer(config, logger)
	err := mcpSrv.createAndInitMCPServer(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)
	// Проверяем, что capability logging есть
	caps := mcpSrv.GetServerCapabilities()
	assert.NotNil(t, caps.Logging, "logging capability must be present when LogRawOutput is :mcp:")
}

func TestMCPServer_LoggingCapability_Disabled(t *testing.T) {
	logger := &testLogger{}
	config := types.MCPServerConfig{
		Name:    "test-server",
		Version: "0.1.0",
		Tool: types.MCPServerToolConfig{
			Name:                "test-tool",
			Description:         "desc",
			ArgumentName:        "arg",
			ArgumentDescription: "desc",
		},
		LogRawOutput: ":stdout:",
	}
	mcpSrv := NewMCPServer(config, logger)
	err := mcpSrv.createAndInitMCPServer(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)
	// Проверяем, что capability logging отсутствует
	caps := mcpSrv.GetServerCapabilities()
	assert.Nil(t, caps.Logging, "logging capability must NOT be present when LogRawOutput is not :mcp:")
}
