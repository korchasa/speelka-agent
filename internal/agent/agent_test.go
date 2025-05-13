package agent

import (
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

type dummyLogger struct{}

func (d *dummyLogger) SetLevel(level logrus.Level)               {}
func (d *dummyLogger) Debug(args ...interface{})                 {}
func (d *dummyLogger) Debugf(format string, args ...interface{}) {}
func (d *dummyLogger) Info(args ...interface{})                  {}
func (d *dummyLogger) Infof(format string, args ...interface{})  {}
func (d *dummyLogger) Warn(args ...interface{})                  {}
func (d *dummyLogger) Warnf(format string, args ...interface{})  {}
func (d *dummyLogger) Error(args ...interface{})                 {}
func (d *dummyLogger) Errorf(format string, args ...interface{}) {}
func (d *dummyLogger) Fatal(args ...interface{})                 {}
func (d *dummyLogger) Fatalf(format string, args ...interface{}) {}
func (d *dummyLogger) WithField(key string, value interface{}) types.LogEntrySpec {
	return &dummyLogEntry{}
}
func (d *dummyLogger) WithFields(fields logrus.Fields) types.LogEntrySpec { return &dummyLogEntry{} }
func (d *dummyLogger) SetMCPServer(mcpServer interface{})                 {}

type dummyLogEntry struct{}

func (d *dummyLogEntry) Debug(args ...interface{})                 {}
func (d *dummyLogEntry) Debugf(format string, args ...interface{}) {}
func (d *dummyLogEntry) Info(args ...interface{})                  {}
func (d *dummyLogEntry) Infof(format string, args ...interface{})  {}
func (d *dummyLogEntry) Warn(args ...interface{})                  {}
func (d *dummyLogEntry) Warnf(format string, args ...interface{})  {}
func (d *dummyLogEntry) Error(args ...interface{})                 {}
func (d *dummyLogEntry) Errorf(format string, args ...interface{}) {}
func (d *dummyLogEntry) Fatal(args ...interface{})                 {}
func (d *dummyLogEntry) Fatalf(format string, args ...interface{}) {}

func TestHandleLLMAnswerToolRequest(t *testing.T) {
	a := &Agent{logger: &dummyLogger{}}
	sess := &chat.Chat{} // Not used in this test
	resp := types.LLMResponse{}

	t.Run("missing text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{}
		res := a.handleLLMAnswerToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for missing text argument, got success")
		}
	})

	t.Run("nil text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": nil}
		res := a.handleLLMAnswerToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for nil text argument, got success")
		}
	})

	t.Run("non-string text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": 123}
		res := a.handleLLMAnswerToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for non-string text argument, got success")
		}
	})

	t.Run("empty string text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": ""}
		res := a.handleLLMAnswerToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for empty string text argument, got success")
		}
	})

	t.Run("valid string text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": "hello"}
		res := a.handleLLMAnswerToolRequest(call, resp, sess)
		if res.IsError {
			t.Errorf("expected success for valid string, got error")
		}
		if len(res.Content) == 0 {
			t.Errorf("expected content in result")
		}
		found := false
		for _, c := range res.Content {
			if tc, ok := c.(mcp.TextContent); ok && tc.Text == "hello" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected text content 'hello' in result")
		}
	})
}
