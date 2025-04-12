// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

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

// LoggerSpec defines the interface for our MCP-aware logger
// Responsibility: Providing a unified logging interface
// Features: Supports different log levels, structured logging, and MCP integration
type LoggerSpec interface {
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
	WithField(key string, value interface{}) LogEntrySpec
	WithFields(fields logrus.Fields) LogEntrySpec
	SetMCPServer(mcpServer interface{})
}

// SDump is a utility function for dumping objects as strings for logging.
// This is defined here to avoid circular dependencies.
func SDump(v interface{}) string {
	return fmt.Sprintf("%+v", v)
}
