// Package mcplogger provides a wrapper around logrus that implements MCP logging.
package mcplogger

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// MCPLogger defines the interface for our MCP-aware logger
type MCPLogger interface {
	SetLevel(level logrus.Level)
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) *MCPLogEntry
	WithFields(fields logrus.Fields) *MCPLogEntry
}

// MCPLogrus wraps a logrus logger and adds MCP logging capabilities
type MCPLogrus struct {
	underlying *logrus.Logger
	mcpServer  *server.MCPServer
	minLevel   logrus.Level
}

// MCPLogEntry is a wrapper around logrus.Entry that supports MCP logging
type MCPLogEntry struct {
	underlying *logrus.Entry
	mcpServer  *server.MCPServer
	minLevel   logrus.Level
	fields     logrus.Fields
}

// NewMCPLogger creates a new MCPLogrus instance that wraps the provided logrus logger
func NewMCPLogger(underlying *logrus.Logger, mcpServer *server.MCPServer) *MCPLogrus {
	logger := &MCPLogrus{
		underlying: underlying,
		mcpServer:  mcpServer,
		minLevel:   logrus.InfoLevel, // Default level
	}

	// Register a handler for the logging/setLevel method
	if mcpServer != nil {
		logger.registerLoggingSetLevelHandler()
	}

	return logger
}

// LogLevelHandlerFunc is the handler for the logging/setLevel tool
func (l *MCPLogrus) LogLevelHandlerFunc(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
func (l *MCPLogrus) registerLoggingSetLevelHandler() {
	// Create a tool for setting log levels
	setLevelTool := mcp.NewTool("logging/setLevel",
		mcp.WithString("level", mcp.Required(), mcp.Description("Log level to set")),
	)

	// Register the tool with the server
	l.mcpServer.AddTool(setLevelTool, l.LogLevelHandlerFunc)
}

// SetLevel sets the minimum level for both the underlying logger and MCP notifications
func (l *MCPLogrus) SetLevel(level logrus.Level) {
	l.underlying.SetLevel(level)
	l.minLevel = level
}

// Debug logs a message at level Debug
func (l *MCPLogrus) Debug(args ...interface{}) {
	l.underlying.Debug(args...)
	l.sendNotification(logrus.DebugLevel, fmt.Sprint(args...), nil)
}

// Debugf logs a formatted message at level Debug
func (l *MCPLogrus) Debugf(format string, args ...interface{}) {
	l.underlying.Debugf(format, args...)
	l.sendNotification(logrus.DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Info logs a message at level Info
func (l *MCPLogrus) Info(args ...interface{}) {
	l.underlying.Info(args...)
	l.sendNotification(logrus.InfoLevel, fmt.Sprint(args...), nil)
}

// Infof logs a formatted message at level Info
func (l *MCPLogrus) Infof(format string, args ...interface{}) {
	l.underlying.Infof(format, args...)
	l.sendNotification(logrus.InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warn logs a message at level Warn
func (l *MCPLogrus) Warn(args ...interface{}) {
	l.underlying.Warn(args...)
	l.sendNotification(logrus.WarnLevel, fmt.Sprint(args...), nil)
}

// Warnf logs a formatted message at level Warn
func (l *MCPLogrus) Warnf(format string, args ...interface{}) {
	l.underlying.Warnf(format, args...)
	l.sendNotification(logrus.WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Error logs a message at level Error
func (l *MCPLogrus) Error(args ...interface{}) {
	l.underlying.Error(args...)
	l.sendNotification(logrus.ErrorLevel, fmt.Sprint(args...), nil)
}

// Errorf logs a formatted message at level Error
func (l *MCPLogrus) Errorf(format string, args ...interface{}) {
	l.underlying.Errorf(format, args...)
	l.sendNotification(logrus.ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// Fatal logs a message at level Fatal
func (l *MCPLogrus) Fatal(args ...interface{}) {
	l.underlying.Fatal(args...)
	// We don't need to send a notification here because Fatal will exit the program
}

// Fatalf logs a formatted message at level Fatal
func (l *MCPLogrus) Fatalf(format string, args ...interface{}) {
	l.underlying.Fatalf(format, args...)
	// We don't need to send a notification here because Fatalf will exit the program
}

// WithField returns an entry with a single field
func (l *MCPLogrus) WithField(key string, value interface{}) *MCPLogEntry {
	return &MCPLogEntry{
		underlying: l.underlying.WithField(key, value),
		mcpServer:  l.mcpServer,
		minLevel:   l.minLevel,
		fields:     logrus.Fields{key: value},
	}
}

// WithFields returns an entry with multiple fields
func (l *MCPLogrus) WithFields(fields logrus.Fields) *MCPLogEntry {
	return &MCPLogEntry{
		underlying: l.underlying.WithFields(fields),
		mcpServer:  l.mcpServer,
		minLevel:   l.minLevel,
		fields:     fields,
	}
}

// Debug logs a message at level Debug with fields
func (e *MCPLogEntry) Debug(args ...interface{}) {
	e.underlying.Debug(args...)
	e.sendNotification(logrus.DebugLevel, fmt.Sprint(args...))
}

// Debugf logs a formatted message at level Debug with fields
func (e *MCPLogEntry) Debugf(format string, args ...interface{}) {
	e.underlying.Debugf(format, args...)
	e.sendNotification(logrus.DebugLevel, fmt.Sprintf(format, args...))
}

// Info logs a message at level Info with fields
func (e *MCPLogEntry) Info(args ...interface{}) {
	e.underlying.Info(args...)
	e.sendNotification(logrus.InfoLevel, fmt.Sprint(args...))
}

// Infof logs a formatted message at level Info with fields
func (e *MCPLogEntry) Infof(format string, args ...interface{}) {
	e.underlying.Infof(format, args...)
	e.sendNotification(logrus.InfoLevel, fmt.Sprintf(format, args...))
}

// Warn logs a message at level Warn with fields
func (e *MCPLogEntry) Warn(args ...interface{}) {
	e.underlying.Warn(args...)
	e.sendNotification(logrus.WarnLevel, fmt.Sprint(args...))
}

// Warnf logs a formatted message at level Warn with fields
func (e *MCPLogEntry) Warnf(format string, args ...interface{}) {
	e.underlying.Warnf(format, args...)
	e.sendNotification(logrus.WarnLevel, fmt.Sprintf(format, args...))
}

// Error logs a message at level Error with fields
func (e *MCPLogEntry) Error(args ...interface{}) {
	e.underlying.Error(args...)
	e.sendNotification(logrus.ErrorLevel, fmt.Sprint(args...))
}

// Errorf logs a formatted message at level Error with fields
func (e *MCPLogEntry) Errorf(format string, args ...interface{}) {
	e.underlying.Errorf(format, args...)
	e.sendNotification(logrus.ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatal logs a message at level Fatal with fields
func (e *MCPLogEntry) Fatal(args ...interface{}) {
	e.underlying.Fatal(args...)
	// We don't need to send a notification here because Fatal will exit the program
}

// Fatalf logs a formatted message at level Fatal with fields
func (e *MCPLogEntry) Fatalf(format string, args ...interface{}) {
	e.underlying.Fatalf(format, args...)
	// We don't need to send a notification here because Fatalf will exit the program
}

// sendNotification sends a log notification via MCP if the level is at or above the minimum level
func (l *MCPLogrus) sendNotification(level logrus.Level, msg string, fields logrus.Fields) {
	if l.mcpServer == nil || level < l.minLevel {
		return
	}

	mcpLevel := logrusToMCPLevel(level)

	// Prepare notification data
	data := map[string]interface{}{
		"level":   mcpLevel,
		"message": msg,
	}

	if len(fields) > 0 {
		data["data"] = fields
	}

	// Send the notification to all clients
	// We ignore errors because logging should not fail the application
	// The sendNotificationToAllClients method is unexported, so we use SendNotificationToClient instead
	if ctx := context.Background(); ctx != nil {
		_ = l.mcpServer.SendNotificationToClient(ctx, "notifications/message", data)
	}
}

// sendNotification sends a log notification via MCP if the level is at or above the minimum level
func (e *MCPLogEntry) sendNotification(level logrus.Level, msg string) {
	if e.mcpServer == nil || level < e.minLevel {
		return
	}

	mcpLevel := logrusToMCPLevel(level)

	// Prepare notification data
	data := map[string]interface{}{
		"level":   mcpLevel,
		"message": msg,
	}

	if len(e.fields) > 0 {
		data["data"] = e.fields
	}

	// Send the notification to all clients
	// We ignore errors because logging should not fail the application
	// The sendNotificationToAllClients method is unexported, so we use SendNotificationToClient instead
	if ctx := context.Background(); ctx != nil {
		_ = e.mcpServer.SendNotificationToClient(ctx, "notifications/message", data)
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
