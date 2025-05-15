package logger

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mcpLoggerTestServer struct {
	calls []map[string]interface{}
}

func (m *mcpLoggerTestServer) SendNotificationToClient(_ context.Context, _ string, data map[string]interface{}) error {
	m.calls = append(m.calls, data)
	return nil
}

func TestMCPLogger_BasicLevels(t *testing.T) {
	mock := &mcpLoggerTestServer{}
	logger := NewMCPLogger()
	logger.SetMCPServer(mock)
	logger.SetLevel(logrus.DebugLevel)
	fmt.Printf("minLevel after SetLevel: %v\n", logger.minLevel)

	logger.Debug("debug message")
	fmt.Printf("minLevel before Info: %v\n", logger.minLevel)
	logger.Info("info message")
	fmt.Printf("minLevel before Warn: %v\n", logger.minLevel)
	logger.Warn("warn message")
	fmt.Printf("minLevel before Error: %v\n", logger.minLevel)
	logger.Error("error message")

	fmt.Printf("mock.calls: %+v\n", mock.calls)
	if len(mock.calls) != 4 {
		t.Fatalf("expected 4 calls, got %d: %+v", len(mock.calls), mock.calls)
	}
	levels := []string{"debug", "info", "warning", "error"}
	for i, call := range mock.calls {
		assert.Equal(t, levels[i], call["level"])
	}
}

func TestMCPLogger_RespectsLevel(t *testing.T) {
	mock := &mcpLoggerTestServer{}
	logger := NewMCPLogger()
	logger.SetMCPServer(mock)
	logger.SetLevel(logrus.WarnLevel)

	logger.Info("should not appear")
	logger.Warn("should appear")
	fmt.Printf("mock.calls: %+v\n", mock.calls)
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d: %+v", len(mock.calls), mock.calls)
	}
	assert.Equal(t, "warning", mock.calls[0]["level"])
	assert.Equal(t, "should appear", mock.calls[0]["message"])
}

func TestMCPLogger_WithFieldAndFields(t *testing.T) {
	mock := &mcpLoggerTestServer{}
	logger := NewMCPLogger()
	logger.SetMCPServer(mock)
	logger.SetLevel(logrus.DebugLevel)

	logger.WithField("foo", "bar").Info("with field")
	logger.WithFields(logrus.Fields{"a": 1, "b": 2}).Warn("with fields")
	fmt.Printf("mock.calls: %+v\n", mock.calls)
	if len(mock.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d: %+v", len(mock.calls), mock.calls)
	}
	assert.Contains(t, mock.calls[0]["data"], "foo")
	assert.Equal(t, "bar", mock.calls[0]["data"].(map[string]interface{})["foo"])
	assert.Contains(t, mock.calls[1]["data"], "a")
	assert.Equal(t, 1, mock.calls[1]["data"].(map[string]interface{})["a"])
	assert.Contains(t, mock.calls[1]["data"], "b")
	assert.Equal(t, 2, mock.calls[1]["data"].(map[string]interface{})["b"])
}

func TestMCPLogger_NoMCPServer_NoPanic(t *testing.T) {
	logger := NewMCPLogger()
	logger.SetLevel(logrus.DebugLevel)
	logger.Info("should not panic")
	// No panic, no call
}

func TestMCPLogger_InfoOnly(t *testing.T) {
	mock := &mcpLoggerTestServer{}
	logger := NewMCPLogger()
	logger.SetMCPServer(mock)
	logger.SetLevel(logrus.DebugLevel)

	logger.Info("info message")

	fmt.Printf("mock.calls: %+v\n", mock.calls)
	assert.Len(t, mock.calls, 1)
	assert.Equal(t, "info", mock.calls[0]["level"])
	assert.Equal(t, "info message", mock.calls[0]["message"])
}

func TestMCPLogger_DoesNotCreateFile(t *testing.T) {
	filename := "mcp"
	// Удаляем файл, если он есть
	_ = os.Remove(filename)

	// Симулируем установку output = "mcp" через Apply
	cfg := &types.Configuration{}
	cfg.Runtime.Log.Output = types.LogOutputMCP
	mgr := configuration.NewConfigurationManager(nil)
	_, err := mgr.Apply(cfg, cfg)
	assert.NoError(t, err)

	// Проверяем, что файл не появился
	if _, err := os.Stat(filename); err == nil {
		os.Remove(filename)
		t.Fatalf("file %s should not be created when output=mcp", filename)
	}
}
