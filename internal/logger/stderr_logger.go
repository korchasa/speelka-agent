package logger

import (
	"io"
	"os"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// IOWriterLogger implements LoggerSpec, logs only to the provided io.Writer (no MCP)
type IOWriterLogger struct {
	underlying *logrus.Logger
	minLevel   logrus.Level
	writer     io.Writer
}

// NewIOWriterLogger creates a logger that writes only to the given io.Writer (default: os.Stderr)
func NewIOWriterLogger(writer io.Writer) *IOWriterLogger {
	if writer == nil {
		writer = os.Stderr
	}
	underlying := logrus.New()
	underlying.SetLevel(logrus.DebugLevel)
	underlying.SetOutput(writer)
	underlying.SetReportCaller(false)
	return &IOWriterLogger{
		underlying: underlying,
		minLevel:   logrus.DebugLevel,
		writer:     writer,
	}
}

func (l *IOWriterLogger) SetLevel(level logrus.Level) {
	l.underlying.SetLevel(level)
	l.minLevel = level
}

func (l *IOWriterLogger) Debug(args ...interface{}) { l.underlying.Debug(args...) }
func (l *IOWriterLogger) Debugf(format string, args ...interface{}) {
	l.underlying.Debugf(format, args...)
}
func (l *IOWriterLogger) Info(args ...interface{}) { l.underlying.Info(args...) }
func (l *IOWriterLogger) Infof(format string, args ...interface{}) {
	l.underlying.Infof(format, args...)
}
func (l *IOWriterLogger) Warn(args ...interface{}) { l.underlying.Warn(args...) }
func (l *IOWriterLogger) Warnf(format string, args ...interface{}) {
	l.underlying.Warnf(format, args...)
}
func (l *IOWriterLogger) Error(args ...interface{}) { l.underlying.Error(args...) }
func (l *IOWriterLogger) Errorf(format string, args ...interface{}) {
	l.underlying.Errorf(format, args...)
}
func (l *IOWriterLogger) Fatal(args ...interface{}) { l.underlying.Fatal(args...) }
func (l *IOWriterLogger) Fatalf(format string, args ...interface{}) {
	l.underlying.Fatalf(format, args...)
}

func (l *IOWriterLogger) WithField(key string, value interface{}) types.LogEntrySpec {
	return &IOWriterEntry{l.underlying.WithField(key, value)}
}
func (l *IOWriterLogger) WithFields(fields logrus.Fields) types.LogEntrySpec {
	return &IOWriterEntry{l.underlying.WithFields(fields)}
}
func (l *IOWriterLogger) SetMCPServer(_ types.MCPServerNotifier) {}

// IOWriterEntry implements LogEntrySpec for IOWriterLogger
// Only logs to the provided io.Writer, no MCP
type IOWriterEntry struct {
	underlying *logrus.Entry
}

func (e *IOWriterEntry) Debug(args ...interface{}) { e.underlying.Debug(args...) }
func (e *IOWriterEntry) Debugf(format string, args ...interface{}) {
	e.underlying.Debugf(format, args...)
}
func (e *IOWriterEntry) Info(args ...interface{}) { e.underlying.Info(args...) }
func (e *IOWriterEntry) Infof(format string, args ...interface{}) {
	e.underlying.Infof(format, args...)
}
func (e *IOWriterEntry) Warn(args ...interface{}) { e.underlying.Warn(args...) }
func (e *IOWriterEntry) Warnf(format string, args ...interface{}) {
	e.underlying.Warnf(format, args...)
}
func (e *IOWriterEntry) Error(args ...interface{}) { e.underlying.Error(args...) }
func (e *IOWriterEntry) Errorf(format string, args ...interface{}) {
	e.underlying.Errorf(format, args...)
}
func (e *IOWriterEntry) Fatal(args ...interface{}) { e.underlying.Fatal(args...) }
func (e *IOWriterEntry) Fatalf(format string, args ...interface{}) {
	e.underlying.Fatalf(format, args...)
}

func (l *IOWriterLogger) SetFormatter(formatter logrus.Formatter) {
	l.underlying.SetFormatter(formatter)
}
