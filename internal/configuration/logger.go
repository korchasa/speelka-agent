package configuration

import "github.com/sirupsen/logrus"

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
	Level logrus.Level
	// Format is the formatter identifier string ("custom", "json", "text", etc.)
	Formatter logrus.Formatter
	// DisableMCP disables MCP notifications even if server is connected
	DisableMCP bool
}
