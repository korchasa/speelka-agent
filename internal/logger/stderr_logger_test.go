package logger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIOWriterLogger_BasicLevels(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewIOWriterLogger(nil)
	logger.underlying.SetOutput(buf)
	logger.SetLevel(logrus.DebugLevel)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	out := buf.String()
	assert.Contains(t, out, "debug message")
	assert.Contains(t, out, "info message")
	assert.Contains(t, out, "warn message")
	assert.Contains(t, out, "error message")
}

func TestIOWriterLogger_RespectsLevel(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewIOWriterLogger(nil)
	logger.underlying.SetOutput(buf)
	logger.SetLevel(logrus.WarnLevel)

	logger.Info("should not appear")
	logger.Warn("should appear")
	out := buf.String()
	assert.NotContains(t, out, "should not appear")
	assert.Contains(t, out, "should appear")
}

func TestIOWriterLogger_WithFieldAndFields(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewIOWriterLogger(nil)
	logger.underlying.SetOutput(buf)
	logger.SetLevel(logrus.DebugLevel)

	logger.WithField("foo", "bar").Info("with field")
	logger.WithFields(logrus.Fields{"a": 1, "b": 2}).Warn("with fields")
	out := buf.String()
	assert.Contains(t, out, "with field")
	assert.Contains(t, out, "foo=bar")
	assert.Contains(t, out, "with fields")
	assert.Contains(t, out, "a=1")
	assert.Contains(t, out, "b=2")
}

func TestIOWriterLogger_SetMCPServerDoesNothing(t *testing.T) {
	logger := NewIOWriterLogger(nil)
	// Should not panic or affect anything
	logger.SetMCPServer(nil)
}
