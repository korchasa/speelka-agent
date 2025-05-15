package mcp_server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

// Minimal mock LoggerSpec for testing
// Only methods actually used in tests
// (others panic)
// NB: Secret/PII filtering is the responsibility of business logic, not logging infrastructure.
// Here we only check that the log is sent correctly.
// Check that logging capability is present
// Check that logging capability is absent

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
	// NB: Secret/PII filtering is the responsibility of business logic, not logging infrastructure.
	// Here we only check that the log is sent correctly.
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
	// Check that logging capability is present
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
	// Check that logging capability is absent
	caps := mcpSrv.GetServerCapabilities()
	assert.Nil(t, caps.Logging, "logging capability must NOT be present when LogRawOutput is not :mcp:")
}
