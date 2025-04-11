package mcplogger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMCPLoggerCreation(t *testing.T) {
	// Create a new logrus logger
	logger := logrus.New()

	// Create a new MCP logger with nil MCPServer (no MCP functionality)
	mcpLogger := NewMCPLogger(logger, nil)

	// Assert that the logger was created with the correct properties
	assert.Equal(t, logger, mcpLogger.underlying)
	assert.Equal(t, logrus.InfoLevel, mcpLogger.minLevel)
}

func TestMCPLoggingWithoutServer(t *testing.T) {
	// Create a new logrus logger with a buffer
	logger := logrus.New()
	buf := new(bytes.Buffer)
	logger.Out = buf
	logger.Level = logrus.DebugLevel

	// Create a new MCP logger with nil MCPServer (no MCP functionality)
	mcpLogger := NewMCPLogger(logger, nil)

	// Test logging at different levels
	mcpLogger.Debug("test debug message")
	mcpLogger.Info("test info message")
	mcpLogger.Warn("test warning message")
	mcpLogger.Error("test error message")

	// Assert that the logs went to the underlying logger
	logStr := buf.String()
	assert.Contains(t, logStr, "test debug message")
	assert.Contains(t, logStr, "test info message")
	assert.Contains(t, logStr, "test warning message")
	assert.Contains(t, logStr, "test error message")
}

func TestMCPLoggingWithFieldsWithoutServer(t *testing.T) {
	// Create a new logrus logger with a buffer
	logger := logrus.New()
	buf := new(bytes.Buffer)
	logger.Out = buf
	logger.Level = logrus.DebugLevel

	// Create a new MCP logger with nil MCPServer (no MCP functionality)
	mcpLogger := NewMCPLogger(logger, nil)

	// Test logging with fields
	mcpLogger.WithField("key", "value").Info("test with field")
	mcpLogger.WithFields(logrus.Fields{
		"key1": "value1",
		"key2": "value2",
	}).Info("test with fields")

	// Assert that the logs went to the underlying logger
	logStr := buf.String()
	assert.Contains(t, logStr, "test with field")
	assert.Contains(t, logStr, "key=value")
	assert.Contains(t, logStr, "test with fields")
	assert.Contains(t, logStr, "key1=value1")
	assert.Contains(t, logStr, "key2=value2")
}

func TestMCPLoggerLevelSetting(t *testing.T) {
	// Create a new logrus logger with a buffer
	logger := logrus.New()
	buf := new(bytes.Buffer)
	logger.Out = buf
	logger.Level = logrus.InfoLevel

	// Create a new MCP logger with nil MCPServer (no MCP functionality)
	mcpLogger := NewMCPLogger(logger, nil)

	// Test that debug messages are not logged initially
	mcpLogger.Debug("debug message that should not appear")
	assert.Empty(t, buf.String())

	// Change the log level
	mcpLogger.SetLevel(logrus.DebugLevel)

	// Test that debug messages are now logged
	mcpLogger.Debug("debug message that should appear")
	assert.Contains(t, buf.String(), "debug message that should appear")
	assert.NotContains(t, buf.String(), "debug message that should not appear")
}

func TestMCPLoggerEntryMethods(t *testing.T) {
	// Create a new logrus logger with a buffer
	logger := logrus.New()
	buf := new(bytes.Buffer)
	logger.Out = buf
	logger.Level = logrus.DebugLevel

	// Create a new MCP logger with nil MCPServer (no MCP functionality)
	mcpLogger := NewMCPLogger(logger, nil)

	// Test log entry methods
	entry := mcpLogger.WithField("test", "value")

	entry.Debug("debug entry")
	entry.Info("info entry")
	entry.Warn("warn entry")
	entry.Error("error entry")

	// Test formatted log entry methods
	entry.Debugf("debug %s", "format")
	entry.Infof("info %s", "format")
	entry.Warnf("warn %s", "format")
	entry.Errorf("error %s", "format")

	// Check all messages were logged
	logStr := buf.String()
	assert.Contains(t, logStr, "debug entry")
	assert.Contains(t, logStr, "info entry")
	assert.Contains(t, logStr, "warn entry")
	assert.Contains(t, logStr, "error entry")
	assert.Contains(t, logStr, "debug format")
	assert.Contains(t, logStr, "info format")
	assert.Contains(t, logStr, "warn format")
	assert.Contains(t, logStr, "error format")
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
