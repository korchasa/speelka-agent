package log_levels

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

// LogrusToMCPLevel converts a logrus level to an MCP log level
func LogrusToMCPLevel(level logrus.Level) string {
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

// MCPToLogrusLevel converts an MCP log level to a logrus level
func MCPToLogrusLevel(level string) (logrus.Level, error) {
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
