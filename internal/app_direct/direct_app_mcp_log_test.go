package app_direct

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
)

// mockLogger implements LoggerSpec and saves stderr to a buffer
// Used to check MCP log output

type bufStderr struct {
	old *os.File
	r   *os.File
	w   *os.File
	buf bytes.Buffer
}

func (b *bufStderr) start() {
	b.old = os.Stderr
	r, w, _ := os.Pipe()
	b.r = r
	b.w = w
	os.Stderr = w
}

func (b *bufStderr) stop() string {
	b.w.Close()
	os.Stderr = b.old
	ioBuf := make([]byte, 1024)
	n, _ := b.r.Read(ioBuf)
	b.buf.Write(ioBuf[:n])
	return b.buf.String()
}

// mockAgentWithMCPLog emulates a child MCP that writes notifications/message

type mockAgentWithMCPLog struct {
	log types.LoggerSpec
}

func (m *mockAgentWithMCPLog) CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error) {
	m.log.Infof("Child MCP: test log info")
	return "ok", types.MetaInfo{Tokens: 1}, nil
}

func TestDirectApp_ChildMCPLogToStderr(t *testing.T) {
	// Buffer for capturing stderr
	buf := &bufStderr{}
	buf.start()
	defer buf.stop()

	// MCPLogger with mcpLogStub (as in direct-call)
	logger := newTestMCPLogger()
	app := &DirectApp{
		logger: logger,
		agent:  &mockAgentWithMCPLog{log: logger},
	}

	// Call
	_ = app.HandleCall(context.Background(), "test")

	// Check that stderr contains MCP info log
	out := buf.stop()
	if !strings.Contains(out, "[MCP info] Child MCP: test log info") {
		t.Errorf("Expected MCP info log in stderr, but not found. Output: %q", out)
	}
}

// newTestMCPLogger creates MCPLogger with mcpLogStub for testing
func newTestMCPLogger() types.LoggerSpec {
	logger := &testMCPLogger{}
	logger.SetMCPServer(&mcpLogStub{})
	return logger
}

// testMCPLogger implements only Infof for testing
// Passes log to mcpLogStub via SendNotificationToClient

type testMCPLogger struct {
	mcpServer types.MCPServerNotifier
}

func (l *testMCPLogger) SetMCPServer(s types.MCPServerNotifier) { l.mcpServer = s }
func (l *testMCPLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if l.mcpServer != nil {
		_ = l.mcpServer.SendNotificationToClient(context.Background(), "notifications/message", map[string]interface{}{
			"level":   "info",
			"message": msg,
		})
	}
}

// Stubs for the interface

// testLogEntry — empty implementation of types.LogEntrySpec for testing
// All methods are no-op

type testLogEntry struct{}

func (e *testLogEntry) Debug(...interface{})          {}
func (e *testLogEntry) Debugf(string, ...interface{}) {}
func (e *testLogEntry) Info(...interface{})           {}
func (e *testLogEntry) Infof(string, ...interface{})  {}
func (e *testLogEntry) Warn(...interface{})           {}
func (e *testLogEntry) Warnf(string, ...interface{})  {}
func (e *testLogEntry) Error(...interface{})          {}
func (e *testLogEntry) Errorf(string, ...interface{}) {}
func (e *testLogEntry) Fatal(...interface{})          {}
func (e *testLogEntry) Fatalf(string, ...interface{}) {}
