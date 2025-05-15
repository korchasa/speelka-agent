package app_direct

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// mockLogger реализует LoggerSpec и сохраняет stderr в буфер
// Используется для проверки вывода MCP-логов

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

// mockAgentWithMCPLog эмулирует вызов дочернего MCP, который пишет notifications/message

type mockAgentWithMCPLog struct {
	log types.LoggerSpec
}

func (m *mockAgentWithMCPLog) CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error) {
	m.log.Infof("Дочерний MCP: тестовый лог info")
	return "ok", types.MetaInfo{Tokens: 1}, nil
}

func TestDirectApp_ChildMCPLogToStderr(t *testing.T) {
	// Буфер для перехвата stderr
	buf := &bufStderr{}
	buf.start()
	defer buf.stop()

	// MCPLogger с mcpLogStub (как в direct-call)
	logger := newTestMCPLogger()
	app := &DirectApp{
		logger: logger,
		agent:  &mockAgentWithMCPLog{log: logger},
	}

	// Вызов
	_ = app.HandleCall(context.Background(), "test")

	// Проверяем, что в stderr есть MCP info лог
	out := buf.stop()
	if !strings.Contains(out, "[MCP info] Дочерний MCP: тестовый лог info") {
		t.Errorf("Ожидался MCP info лог в stderr, но его нет. Вывод: %q", out)
	}
}

// newTestMCPLogger создаёт MCPLogger с mcpLogStub для теста
func newTestMCPLogger() types.LoggerSpec {
	logger := &testMCPLogger{}
	logger.SetMCPServer(&mcpLogStub{})
	return logger
}

// testMCPLogger реализует только Infof для теста
// Пробрасывает лог в mcpLogStub через SendNotificationToClient

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

// Заглушки для интерфейса
func (l *testMCPLogger) SetLevel(level logrus.Level)                      {}
func (l *testMCPLogger) Debug(...interface{})                             {}
func (l *testMCPLogger) Debugf(string, ...interface{})                    {}
func (l *testMCPLogger) Info(...interface{})                              {}
func (l *testMCPLogger) Warn(...interface{})                              {}
func (l *testMCPLogger) Warnf(string, ...interface{})                     {}
func (l *testMCPLogger) Error(...interface{})                             {}
func (l *testMCPLogger) Errorf(string, ...interface{})                    {}
func (l *testMCPLogger) Fatal(...interface{})                             {}
func (l *testMCPLogger) Fatalf(string, ...interface{})                    {}
func (l *testMCPLogger) WithField(string, interface{}) types.LogEntrySpec { return &testLogEntry{} }
func (l *testMCPLogger) WithFields(logrus.Fields) types.LogEntrySpec      { return &testLogEntry{} }
func (l *testMCPLogger) SetFormatter(_ logrus.Formatter)                  {}

// testLogEntry — пустая реализация types.LogEntrySpec для теста
// Все методы — no-op

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
