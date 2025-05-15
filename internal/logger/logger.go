// Package logger provides a wrapper around logrus that implements MCP logging.
package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/korchasa/speelka-agent-go/internal/types"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// MCPServerNotifier defines the interface for sending MCP notifications
// Used for injecting mock in tests and real MCPServer in production
// MCPServerNotifier abstracts notification sending for logger
// Only SendNotificationToClient is required for logger
//
//go:generate mockgen -destination=mock_mcpserver_notifier.go -package=logger . MCPServerNotifier
type MCPServerNotifier interface {
	SendNotificationToClient(ctx context.Context, method string, data map[string]interface{}) error
}

// Logger wraps a logrus logger and adds MCP logging capabilities
type Logger struct {
	underlying *logrus.Logger
	mcpServer  types.MCPServerNotifier
	minLevel   logrus.Level
}

// NewLogger creates a new Logger instance with an internal logrus logger
func NewLogger() *Logger {
	// Create and configure the underlying logger
	underlying := logrus.New()
	underlying.SetReportCaller(false)

	// Create the Logger instance
	logger := &Logger{
		underlying: underlying,
		mcpServer:  nil,
	}
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(os.Stderr)

	return logger
}

// SetFormatter sets the formatter for the underlying logger
func (l *Logger) SetFormatter(formatter logrus.Formatter) {
	l.underlying.SetFormatter(formatter)
}

// SetMCPServer sets the MCP server instance for this logger and registers the log level handler
func (l *Logger) SetMCPServer(mcpServer types.MCPServerNotifier) {
	l.mcpServer = mcpServer
	if srv, ok := mcpServer.(*server.MCPServer); ok {
		l.registerLoggingSetLevelHandler(srv)
	}
}

// LogLevelHandlerFunc is the handler for the logging/setLevel tool
func (l *Logger) LogLevelHandlerFunc(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters from the request
	levelStr, ok := req.Params.Arguments["level"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid level parameter")
	}

	// Convert MCP log level to logrus level
	logrusLevel, err := mcpToLogrusLevel(levelStr)
	if err != nil {
		return nil, err
	}

	// Set the level
	l.minLevel = logrusLevel
	l.underlying.SetLevel(logrusLevel)

	return &mcp.CallToolResult{}, nil
}

// registerLoggingSetLevelHandler registers a handler for the logging/setLevel method
func (l *Logger) registerLoggingSetLevelHandler(srv *server.MCPServer) {
	setLevelTool := mcp.NewTool("logging/setLevel",
		mcp.WithString("level", mcp.Required(), mcp.Description("Log level to set")),
	)
	srv.AddTool(setLevelTool, l.LogLevelHandlerFunc)
}

// SetLevel sets the minimum level for both the underlying logger and MCP notifications
func (l *Logger) SetLevel(level logrus.Level) {
	l.underlying.SetLevel(level)
	l.minLevel = level
}

func (l *Logger) SetOutput(output io.Writer) {
	l.underlying.SetOutput(output)
}

// Debug logs a message at level Debug
func (l *Logger) Debug(args ...interface{}) {
	l.underlying.Debug(args...)
	l.sendNotification(logrus.DebugLevel, fmt.Sprint(args...), nil)
}

// Debugf logs a formatted message at level Debug
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.underlying.Debugf(format, args...)
	l.sendNotification(logrus.DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Info logs a message at level Info
func (l *Logger) Info(args ...interface{}) {
	l.underlying.Info(args...)
	l.sendNotification(logrus.InfoLevel, fmt.Sprint(args...), nil)
}

// Infof logs a formatted message at level Info
func (l *Logger) Infof(format string, args ...interface{}) {
	l.underlying.Infof(format, args...)
	l.sendNotification(logrus.InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warn logs a message at level Warn
func (l *Logger) Warn(args ...interface{}) {
	l.underlying.Warn(args...)
	l.sendNotification(logrus.WarnLevel, fmt.Sprint(args...), nil)
}

// Warnf logs a formatted message at level Warn
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.underlying.Warnf(format, args...)
	l.sendNotification(logrus.WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Error logs a message at level Error
func (l *Logger) Error(args ...interface{}) {
	l.underlying.Error(args...)
	l.sendNotification(logrus.ErrorLevel, fmt.Sprint(args...), nil)
}

// Errorf logs a formatted message at level Error
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.underlying.Errorf(format, args...)
	l.sendNotification(logrus.ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// Fatal logs a message at level Fatal
func (l *Logger) Fatal(args ...interface{}) {
	l.underlying.Fatal(args...)
	// We don't need to send a notification here because Fatal will exit the program
}

// Fatalf logs a formatted message at level Fatal
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.underlying.Fatalf(format, args...)
	// We don't need to send a notification here because Fatalf will exit the program
}

// WithField returns an entry with a single field
func (l *Logger) WithField(key string, value interface{}) types.LogEntrySpec {
	return &Entry{
		underlying: l.underlying.WithField(key, value),
		mcpServer:  l.mcpServer,
		minLevel:   l.minLevel,
		fields:     logrus.Fields{key: value},
	}
}

// WithFields returns an entry with multiple fields
func (l *Logger) WithFields(fields logrus.Fields) types.LogEntrySpec {
	return &Entry{
		underlying: l.underlying.WithFields(fields),
		mcpServer:  l.mcpServer,
		minLevel:   l.minLevel,
		fields:     fields,
	}
}

// sendNotification sends a log notification via MCP if the level is at or above the minimum level
func (l *Logger) sendNotification(level logrus.Level, msg string, fields logrus.Fields) {
	if l.mcpServer == nil || level < l.minLevel {
		return
	}

	mcpLevel := logrusToMCPLevel(level)

	// Prepare notification data
	data := map[string]interface{}{
		"level":               mcpLevel,
		"message":             msg,
		"delivered_to_client": true,
	}

	if len(fields) > 0 {
		fields["delivered_to_client"] = true // logrus log
		data["data"] = fields
	}

	// Send the notification to all clients
	// We ignore errors because logging should not fail the application
	// The sendNotificationToAllClients method is unexported, so we use SendNotificationToClient instead
	if ctx := context.Background(); ctx != nil {
		_ = l.mcpServer.SendNotificationToClient(ctx, "notifications/message", data)
	}
}

// logrusToMCPLevel converts a logrus level to an MCP log level
func logrusToMCPLevel(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel, logrus.DebugLevel:
		return "debug"
	case logrus.InfoLevel:
		return "info"
	case logrus.WarnLevel:
		return "warning"
	case logrus.ErrorLevel:
		return "error"
	case logrus.FatalLevel:
		return "critical"
	case logrus.PanicLevel:
		return "alert"
	default:
		return "info"
	}
}

// mcpToLogrusLevel converts an MCP log level to a logrus level
func mcpToLogrusLevel(level string) (logrus.Level, error) {
	switch level {
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "notice":
		return logrus.InfoLevel, nil
	case "warning":
		return logrus.WarnLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "critical", "alert", "emergency":
		return logrus.FatalLevel, nil
	default:
		return logrus.InfoLevel, fmt.Errorf("unknown MCP log level: %s", level)
	}
}
