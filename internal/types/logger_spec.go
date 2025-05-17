// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

const (
	LogOutputStdout = ":stdout:"
	LogOutputStderr = ":stderr:"
	LogOutputMCP    = ":mcp:"
)

// LogConfig represents the configuration for logging.
// Responsibility: Storing logging system settings
// Features: Defines the level, format, and output location for logs
// LogConfig is used only for business logic, not for parsing.
// DefaultLevel is the string value from config ("info", "debug", etc.)
// Output is the output identifier string (":stdout:", ":stderr:", ":mcp:", file path)
// Format is the formatter identifier string ("custom", "json", "text", etc.)
// Level is the computed logrus.Level
// UseMCPLogs indicates whether to use MCP logging
type LogConfig struct {
	// DefaultLevel is the string value from config ("info", "debug", etc.)
	DefaultLevel string
	// Format is the formatter identifier string ("custom", "json", "text", etc.)
	Format string
	// Level is the computed logrus.Level
	Level logrus.Level
	// DisableMCP disables MCP notifications even if server is connected
	DisableMCP bool
}

// LogEntrySpec defines the interface for a log entry with fields.
// Responsibility: Providing a unified interface for log entries
// Features: Supports different log levels and structured logging
type LogEntrySpec interface {
	// Debug logs a message at the debug level.
	Debug(args ...interface{})

	// Debugf logs a formatted message at the debug level.
	Debugf(format string, args ...interface{})

	// Info logs a message at the info level.
	Info(args ...interface{})

	// Infof logs a formatted message at the info level.
	Infof(format string, args ...interface{})

	// Warn logs a message at the warn level.
	Warn(args ...interface{})

	// Warnf logs a formatted message at the warn level.
	Warnf(format string, args ...interface{})

	// Error logs a message at the error level.
	Error(args ...interface{})

	// Errorf logs a formatted message at the error level.
	Errorf(format string, args ...interface{})

	// Fatal logs a message at the fatal level and then exits.
	Fatal(args ...interface{})

	// Fatalf logs a formatted message at the fatal level and then exits.
	Fatalf(format string, args ...interface{})
}

// MCPServerNotifier defines the interface for sending MCP notifications (locally to avoid circular imports)
type MCPServerNotifier interface {
	AddTool(tool *mcp.Tool)
	SendNotificationToClient(ctx context.Context, method string, data map[string]interface{}) error
}

// LoggerSpec defines the interface for our MCP-aware logger
// Responsibility: Providing a unified logging interface
// Features: Supports different log levels, structured logging, and MCP integration
type LoggerSpec interface {
	SetLevel(level logrus.Level)
	SetFormatter(formatter logrus.Formatter)
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
	WithField(key string, value interface{}) LogEntrySpec
	WithFields(fields logrus.Fields) LogEntrySpec
	HandleMCPSetLevel(ctx context.Context, req interface{}) (interface{}, error)
}
