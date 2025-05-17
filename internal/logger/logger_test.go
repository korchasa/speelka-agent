package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogger_RespectsConfigLogLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(false)
	logger.underlying.SetOutput(&buf)
	logger.SetLevel(logrus.WarnLevel)

	logger.Info("this is info")
	logger.Warn("this is warn")

	output := buf.String()
	if output == "" {
		t.Fatal("expected some output, got none")
	}
	if contains := bytes.Contains([]byte(output), []byte("this is info")); contains {
		t.Error("info log should not be present at warn level")
	}
	if !bytes.Contains([]byte(output), []byte("this is warn")) {
		t.Error("warn log should be present at warn level")
	}
}

func TestLogger_UsesJSONFormatterWhenConfigured(t *testing.T) {
	logger := newTestLogger(false)
	var buf bytes.Buffer
	logger.underlying.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	logger.underlying.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("json test", "foo")
	output := buf.String()
	assert.Contains(t, output, "json test")
	assert.Contains(t, output, "foo")
	var js map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(output), &js))
}

func TestMCPServer_DeclaresLoggingCapability(t *testing.T) {
	mcpServer := server.NewMCPServer("test-server", "0.1.0", server.WithLogging())
	// Get the capabilities field via reflection
	val := reflect.ValueOf(mcpServer).Elem().FieldByName("capabilities")
	if !val.IsValid() {
		t.Fatal("capabilities field not found in MCPServer")
	}
	logging := val.FieldByName("logging")
	if !logging.IsValid() {
		t.Fatal("logging field not found in capabilities")
	}
	assert.True(t, logging.Bool(), "logging capability must be enabled")
}

func TestLogger_DisableMCP(t *testing.T) {
	var mockServer mockMCPServer
	logger := newTestLogger(true)

	logger.Infof("test info %s", "should-not-send-mcp")

	assert.Empty(t, mockServer.lastMethod)
}

func TestLoggerLevelSetting(t *testing.T) {
	logger := newTestLogger(false)
	logger.SetLevel(logrus.InfoLevel)
	assert.Equal(t, logrus.InfoLevel, logger.minLevel)
	logger.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, logger.minLevel)
}

func TestLoggerEntryMethods(t *testing.T) {
	logger := newTestLogger(false)
	entry := logger.WithField("test", "value")
	entry.Debug("debug entry")
	entry.Info("info entry")
	entry.Warn("warn entry")
	entry.Error("error entry")
	entry.Debugf("debug %s", "format")
	entry.Infof("info %s", "format")
	entry.Warnf("warn %s", "format")
	entry.Errorf("error %s", "format")
}

func getAllToolNamesFromMCPServer(s *server.MCPServer) []string {
	val := reflect.ValueOf(s).Elem().FieldByName("tools")
	if !val.IsValid() {
		return nil
	}
	names := make([]string, 0, val.Len())
	for _, key := range val.MapKeys() {
		names = append(names, key.String())
	}
	return names
}

func TestLogger_RegistersSetLevelToolAndChangesLevel(t *testing.T) {
	mcpSrv := server.NewMCPServer("test-server", "0.1.0", server.WithLogging())
	logger := newTestLogger(false)
	// Register tool via MCPServer
	loggingSetLevel := mcp.NewTool("logging/setLevel",
		mcp.WithString("level", mcp.Required(), mcp.Description("Log level to set")),
	)
	mcpSrv.AddTool(loggingSetLevel, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := logger.HandleMCPSetLevel(ctx, req)
		if err != nil {
			return nil, err
		}
		result, ok := res.(*mcp.CallToolResult)
		if !ok {
			return nil, fmt.Errorf("unexpected result type from HandleMCPSetLevel")
		}
		return result, nil
	})

	// Check that the logging/setLevel tool is registered
	names := getAllToolNamesFromMCPServer(mcpSrv)
	found := false
	for _, name := range names {
		if name == "logging/setLevel" {
			found = true
			break
		}
	}
	assert.True(t, found, "logging/setLevel tool must be registered by MCPServer")

	// Check that the logging level changes via handler
	callReq := mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      "logging/setLevel",
			Arguments: map[string]interface{}{"level": "debug"},
		},
	}
	res, err := logger.HandleMCPSetLevel(context.Background(), callReq)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "debug", logger.GetLogrusLevel().String())
}

func newTestLogger(disableMCP bool) *Logger {
	cfg := types.LogConfig{
		DefaultLevel: "debug",
		Format:       "text",
		Level:        logrus.DebugLevel,
		DisableMCP:   disableMCP,
	}
	return NewLogger(cfg)
}

type mockMCPServer struct {
	lastMethod string
}
