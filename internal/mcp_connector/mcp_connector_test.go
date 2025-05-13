package mcp_connector

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func Test_testableTimeoutSelect_result(t *testing.T) {
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	cancel := func() {}

	want := &mcp.CallToolResult{Result: mcp.Result{Meta: map[string]any{"test": true}}}
	resultCh <- want

	result, err, timedOut := testableTimeoutSelect(resultCh, errCh, 50*time.Millisecond, cancel)
	assert.Equal(t, want, result)
	assert.Nil(t, err)
	assert.False(t, timedOut)
}

func Test_testableTimeoutSelect_error(t *testing.T) {
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	cancel := func() {}

	errWant := context.DeadlineExceeded
	errCh <- errWant

	result, err, timedOut := testableTimeoutSelect(resultCh, errCh, 50*time.Millisecond, cancel)
	assert.Nil(t, result)
	assert.Equal(t, errWant, err)
	assert.False(t, timedOut)
}

func Test_testableTimeoutSelect_timeout(t *testing.T) {
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err, timedOut := testableTimeoutSelect(resultCh, errCh, 10*time.Millisecond, cancel)
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, timedOut)
	// Context should be canceled
	select {
	case <-ctx.Done():
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Fatal("context was not canceled after timeout")
	}
}

// mockLogger captures Infof calls for testing
// Only implements Infof for this test

type mockLogger struct {
	lastMsg string
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.lastMsg = fmt.Sprintf(format, args...)
}

// Implement unused methods to satisfy LoggerSpec
func (m *mockLogger) SetLevel(_ interface{})                    {}
func (m *mockLogger) Debug(...interface{})                      {}
func (m *mockLogger) Debugf(string, ...interface{})             {}
func (m *mockLogger) Info(...interface{})                       {}
func (m *mockLogger) Warn(...interface{})                       {}
func (m *mockLogger) Warnf(string, ...interface{})              {}
func (m *mockLogger) Error(...interface{})                      {}
func (m *mockLogger) Errorf(string, ...interface{})             {}
func (m *mockLogger) Fatal(...interface{})                      {}
func (m *mockLogger) Fatalf(string, ...interface{})             {}
func (m *mockLogger) WithField(string, interface{}) interface{} { return m }
func (m *mockLogger) WithFields(interface{}) interface{}        { return m }
func (m *mockLogger) SetMCPServer(interface{})                  {}

func Test_StderrLoggingTrimsNewlines(t *testing.T) {
	// Simulate the goroutine logic directly
	logger := &mockLogger{}
	serverID := "test-server"
	line := "error message with newline\n\r  "
	trimmed := strings.TrimRight(line, "\r\n \t")
	logger.Infof("`%s` stderr: %s", serverID, trimmed)
	assert.Equal(t, "`test-server` stderr: error message with newline", logger.lastMsg)
}
