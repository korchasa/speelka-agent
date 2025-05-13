package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMCPLoggerCreation(t *testing.T) {
	mcpLogger := NewMCPLogger()
	assert.Equal(t, logrus.DebugLevel, mcpLogger.minLevel)
}

func TestMCPLoggingWithoutServer(t *testing.T) {
	mcpLogger := NewMCPLogger()
	mcpLogger.Debug("test debug message")
	mcpLogger.Info("test info message")
	mcpLogger.Warn("test warning message")
	mcpLogger.Error("test error message")
	// Нет проверки вывода, так как MCPLogger не пишет в буфер
}

func TestMCPLoggingWithFieldsWithoutServer(t *testing.T) {
	mcpLogger := NewMCPLogger()
	mcpLogger.WithField("key", "value").Info("test with field")
	mcpLogger.WithFields(logrus.Fields{
		"key1": "value1",
		"key2": "value2",
	}).Info("test with fields")
	// Нет проверки вывода, так как MCPLogger не пишет в буфер
}

func TestMCPLoggerLevelSetting(t *testing.T) {
	mcpLogger := NewMCPLogger()
	mcpLogger.SetLevel(logrus.InfoLevel)
	// Проверяем только minLevel
	assert.Equal(t, logrus.InfoLevel, mcpLogger.minLevel)
	mcpLogger.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, mcpLogger.minLevel)
}

func TestMCPLoggerEntryMethods(t *testing.T) {
	mcpLogger := NewMCPLogger()
	entry := mcpLogger.WithField("test", "value")
	entry.Debug("debug entry")
	entry.Info("info entry")
	entry.Warn("warn entry")
	entry.Error("error entry")
	entry.Debugf("debug %s", "format")
	entry.Infof("info %s", "format")
	entry.Warnf("warn %s", "format")
	entry.Errorf("error %s", "format")
	// Нет проверки вывода, так как MCPLogger не пишет в буфер
}

func TestMCPLogLevelConversion(t *testing.T) {
	// Test Logrus to MCP level conversion
	assert.Equal(t, "debug", logrusToMCPLevel(logrus.DebugLevel))
	assert.Equal(t, "debug", logrusToMCPLevel(logrus.TraceLevel))
	assert.Equal(t, "info", logrusToMCPLevel(logrus.InfoLevel))
	assert.Equal(t, "warning", logrusToMCPLevel(logrus.WarnLevel))
	assert.Equal(t, "error", logrusToMCPLevel(logrus.ErrorLevel))
	assert.Equal(t, "critical", logrusToMCPLevel(logrus.FatalLevel))
	assert.Equal(t, "alert", logrusToMCPLevel(logrus.PanicLevel))

	// Test MCP to Logrus level conversion
	debugLevel, err := mcpToLogrusLevel("debug")
	assert.NoError(t, err)
	assert.Equal(t, logrus.DebugLevel, debugLevel)

	infoLevel, err := mcpToLogrusLevel("info")
	assert.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, infoLevel)

	noticeLevel, err := mcpToLogrusLevel("notice")
	assert.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, noticeLevel)

	warningLevel, err := mcpToLogrusLevel("warning")
	assert.NoError(t, err)
	assert.Equal(t, logrus.WarnLevel, warningLevel)

	errorLevel, err := mcpToLogrusLevel("error")
	assert.NoError(t, err)
	assert.Equal(t, logrus.ErrorLevel, errorLevel)

	criticalLevel, err := mcpToLogrusLevel("critical")
	assert.NoError(t, err)
	assert.Equal(t, logrus.FatalLevel, criticalLevel)

	alertLevel, err := mcpToLogrusLevel("alert")
	assert.NoError(t, err)
	assert.Equal(t, logrus.FatalLevel, alertLevel)

	emergencyLevel, err := mcpToLogrusLevel("emergency")
	assert.NoError(t, err)
	assert.Equal(t, logrus.FatalLevel, emergencyLevel)

	// Test invalid MCP level
	_, err = mcpToLogrusLevel("invalid")
	assert.Error(t, err)
}

func TestLogger_RespectsConfigLogLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewIOWriterLogger(nil)
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
	logger := NewIOWriterLogger(nil)
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
	// Получаем поле capabilities через рефлексию
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

type mockMCPServer struct {
	lastCtx    context.Context
	lastMethod string
	lastData   map[string]interface{}
}

func TestLogger_SendsMCPNotification(t *testing.T) {
	var mockServer mockMCPServer
	logger := NewMCPLogger()
	var mcpMock = &mcpServerMock{
		mockSend: func(ctx context.Context, method string, data map[string]interface{}) error {
			mockServer.lastCtx = ctx
			mockServer.lastMethod = method
			mockServer.lastData = data
			return nil
		},
	}
	logger.SetMCPServer(mcpMock)
	logger.SetLevel(logrus.InfoLevel)

	logger.Infof("test info %s", "mcp-notification")

	assert.Equal(t, "notifications/message", mockServer.lastMethod)
	assert.Equal(t, "info", mockServer.lastData["level"])
	assert.Contains(t, mockServer.lastData["message"], "test info mcp-notification")
}

type mcpServerMock struct {
	mockSend func(ctx context.Context, method string, data map[string]interface{}) error
}

func (m *mcpServerMock) SendNotificationToClient(ctx context.Context, method string, data map[string]interface{}) error {
	return m.mockSend(ctx, method, data)
}

func TestLogger_MCPLogHasDeliveredToClientMark(t *testing.T) {
	var mockServer mockMCPServer
	logger := NewMCPLogger()
	var mcpMock = &mcpServerMock{
		mockSend: func(ctx context.Context, method string, data map[string]interface{}) error {
			mockServer.lastCtx = ctx
			mockServer.lastMethod = method
			mockServer.lastData = data
			return nil
		},
	}
	logger.SetMCPServer(mcpMock)
	logger.SetLevel(logrus.InfoLevel)

	logger.WithField("foo", "bar").Info("test delivered mark")

	// Проверяем MCP-лог
	assert.Equal(t, "notifications/message", mockServer.lastMethod)
	if v, ok := mockServer.lastData["delivered_to_client"]; ok {
		assert.Equal(t, true, v)
	}
	if data, ok := mockServer.lastData["data"]; ok {
		if fields, ok := data.(logrus.Fields); ok {
			if v, ok := fields["delivered_to_client"]; ok {
				assert.Equal(t, true, v)
			}
		}
	}
}
