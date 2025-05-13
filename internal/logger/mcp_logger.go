package logger

import (
	"context"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// MCPLogger implements LoggerSpec, logs only via MCP notifications (no stderr)
type MCPLogger struct {
	mcpServer types.MCPServerNotifier
	minLevel  logrus.Level
}

func NewMCPLogger() *MCPLogger {
	return &MCPLogger{
		mcpServer: nil,
		minLevel:  logrus.DebugLevel,
	}
}

func (l *MCPLogger) SetLevel(level logrus.Level) { l.minLevel = level }

func (l *MCPLogger) Debug(args ...interface{}) { l.send(logrus.DebugLevel, fmt.Sprint(args...), nil) }
func (l *MCPLogger) Debugf(format string, args ...interface{}) {
	l.send(logrus.DebugLevel, fmt.Sprintf(format, args...), nil)
}
func (l *MCPLogger) Info(args ...interface{}) { l.send(logrus.InfoLevel, fmt.Sprint(args...), nil) }
func (l *MCPLogger) Infof(format string, args ...interface{}) {
	l.send(logrus.InfoLevel, fmt.Sprintf(format, args...), nil)
}
func (l *MCPLogger) Warn(args ...interface{}) { l.send(logrus.WarnLevel, fmt.Sprint(args...), nil) }
func (l *MCPLogger) Warnf(format string, args ...interface{}) {
	l.send(logrus.WarnLevel, fmt.Sprintf(format, args...), nil)
}
func (l *MCPLogger) Error(args ...interface{}) { l.send(logrus.ErrorLevel, fmt.Sprint(args...), nil) }
func (l *MCPLogger) Errorf(format string, args ...interface{}) {
	l.send(logrus.ErrorLevel, fmt.Sprintf(format, args...), nil)
}
func (l *MCPLogger) Fatal(args ...interface{}) {
	l.send(logrus.FatalLevel, fmt.Sprint(args...), nil)
	panic(fmt.Sprint(args...))
}
func (l *MCPLogger) Fatalf(format string, args ...interface{}) {
	l.send(logrus.FatalLevel, fmt.Sprintf(format, args...), nil)
	panic(fmt.Sprintf(format, args...))
}

func (l *MCPLogger) WithField(key string, value interface{}) types.LogEntrySpec {
	return &MCPEntry{logger: l, fields: logrus.Fields{key: value}}
}
func (l *MCPLogger) WithFields(fields logrus.Fields) types.LogEntrySpec {
	return &MCPEntry{logger: l, fields: fields}
}
func (l *MCPLogger) SetMCPServer(mcpServer types.MCPServerNotifier) { l.mcpServer = mcpServer }

// send отправляет лог только если MCPServer установлен и уровень >= minLevel
func (l *MCPLogger) send(level logrus.Level, msg string, fields logrus.Fields) {
	if l.mcpServer == nil {
		return
	}
	if level > l.minLevel {
		return
	}
	data := map[string]interface{}{
		"level":               logrusToMCPLevel(level),
		"message":             msg,
		"delivered_to_client": true,
	}
	if len(fields) > 0 {
		fieldsCopy := make(map[string]interface{}, len(fields))
		for k, v := range fields {
			fieldsCopy[k] = v
		}
		fieldsCopy["delivered_to_client"] = true
		data["data"] = fieldsCopy
	}
	_ = l.mcpServer.SendNotificationToClient(context.Background(), "notifications/message", data)
}

// MCPEntry implements LogEntrySpec for MCPLogger
// Only sends notifications via MCP, no stderr
type MCPEntry struct {
	logger *MCPLogger
	fields logrus.Fields
}

func (e *MCPEntry) Debug(args ...interface{}) {
	e.logger.send(logrus.DebugLevel, fmt.Sprint(args...), e.fields)
}
func (e *MCPEntry) Debugf(format string, args ...interface{}) {
	e.logger.send(logrus.DebugLevel, fmt.Sprintf(format, args...), e.fields)
}
func (e *MCPEntry) Info(args ...interface{}) {
	e.logger.send(logrus.InfoLevel, fmt.Sprint(args...), e.fields)
}
func (e *MCPEntry) Infof(format string, args ...interface{}) {
	e.logger.send(logrus.InfoLevel, fmt.Sprintf(format, args...), e.fields)
}
func (e *MCPEntry) Warn(args ...interface{}) {
	e.logger.send(logrus.WarnLevel, fmt.Sprint(args...), e.fields)
}
func (e *MCPEntry) Warnf(format string, args ...interface{}) {
	e.logger.send(logrus.WarnLevel, fmt.Sprintf(format, args...), e.fields)
}
func (e *MCPEntry) Error(args ...interface{}) {
	e.logger.send(logrus.ErrorLevel, fmt.Sprint(args...), e.fields)
}
func (e *MCPEntry) Errorf(format string, args ...interface{}) {
	e.logger.send(logrus.ErrorLevel, fmt.Sprintf(format, args...), e.fields)
}
func (e *MCPEntry) Fatal(args ...interface{}) {
	e.logger.send(logrus.FatalLevel, fmt.Sprint(args...), e.fields)
	panic(fmt.Sprint(args...))
}
func (e *MCPEntry) Fatalf(format string, args ...interface{}) {
	e.logger.send(logrus.FatalLevel, fmt.Sprintf(format, args...), e.fields)
	panic(fmt.Sprintf(format, args...))
}

func (l *MCPLogger) SetFormatter(formatter logrus.Formatter) {}
