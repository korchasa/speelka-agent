package logger

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// Entry is a wrapper around logrus.Entry that supports MCP logging
type Entry struct {
	underlying *logrus.Entry
	mcpServer  *server.MCPServer
	minLevel   logrus.Level
	fields     logrus.Fields
}

// Debug logs a message at level Debug with fields
func (e *Entry) Debug(args ...interface{}) {
	e.underlying.Debug(args...)
	e.sendNotification(logrus.DebugLevel, fmt.Sprint(args...))
}

// Debugf logs a formatted message at level Debug with fields
func (e *Entry) Debugf(format string, args ...interface{}) {
	e.underlying.Debugf(format, args...)
	e.sendNotification(logrus.DebugLevel, fmt.Sprintf(format, args...))
}

// Info logs a message at level Info with fields
func (e *Entry) Info(args ...interface{}) {
	e.underlying.Info(args...)
	e.sendNotification(logrus.InfoLevel, fmt.Sprint(args...))
}

// Infof logs a formatted message at level Info with fields
func (e *Entry) Infof(format string, args ...interface{}) {
	e.underlying.Infof(format, args...)
	e.sendNotification(logrus.InfoLevel, fmt.Sprintf(format, args...))
}

// Warn logs a message at level Warn with fields
func (e *Entry) Warn(args ...interface{}) {
	e.underlying.Warn(args...)
	e.sendNotification(logrus.WarnLevel, fmt.Sprint(args...))
}

// Warnf logs a formatted message at level Warn with fields
func (e *Entry) Warnf(format string, args ...interface{}) {
	e.underlying.Warnf(format, args...)
	e.sendNotification(logrus.WarnLevel, fmt.Sprintf(format, args...))
}

// Error logs a message at level Error with fields
func (e *Entry) Error(args ...interface{}) {
	e.underlying.Error(args...)
	e.sendNotification(logrus.ErrorLevel, fmt.Sprint(args...))
}

// Errorf logs a formatted message at level Error with fields
func (e *Entry) Errorf(format string, args ...interface{}) {
	e.underlying.Errorf(format, args...)
	e.sendNotification(logrus.ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatal logs a message at level Fatal with fields
func (e *Entry) Fatal(args ...interface{}) {
	e.underlying.Fatal(args...)
	// We don't need to send a notification here because Fatal will exit the program
}

// Fatalf logs a formatted message at level Fatal with fields
func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.underlying.Fatalf(format, args...)
	// We don't need to send a notification here because Fatalf will exit the program
}

// sendNotification sends a log notification via MCP if the level is at or above the minimum level
func (e *Entry) sendNotification(level logrus.Level, msg string) {
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
