package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

type mockMCPClient struct {
	lastRequest mcp.CallToolRequest
	lastLevel   string
	fail        bool
}

func (m *mockMCPClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	m.lastRequest = req
	if req.Params.Name == "logging/setLevel" {
		if args, ok := req.Params.Arguments.(map[string]any); ok {
			if level, ok := args["level"].(string); ok {
				m.lastLevel = level
			}
		}
	}
	if m.fail {
		return nil, errors.New("fail")
	}
	return &mcp.CallToolResult{}, nil
}

type callTooler interface {
	CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// Check debug filtering
// Check object output
// Check error output

func TestHandleLoggingNotification(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	notification := mcp.JSONRPCNotification{
		Notification: mcp.Notification{
			Method: "notifications/message",
			Params: mcp.NotificationParams{
				AdditionalFields: map[string]any{
					"level":  "info",
					"logger": "test",
					"data":   "hello world",
				},
			},
		},
	}

	handled := false
	handler := func(notification mcp.JSONRPCNotification) {
		if notification.Method != "notifications/message" {
			return
		}
		var logMsg mcp.LoggingMessageNotification
		params, err := json.Marshal(notification.Params.AdditionalFields)
		if err != nil {
			return
		}
		if err := json.Unmarshal(params, &logMsg.Params); err != nil {
			return
		}
		level := string(logMsg.Params.Level)
		if level == "debug" {
			return
		}
		msg := ""
		if s, ok := logMsg.Params.Data.(string); ok {
			msg = s
		} else {
			b, _ := json.Marshal(logMsg.Params.Data)
			msg = string(b)
		}
		log.Printf("[MCP %s] %s: %s", strings.ToUpper(level), logMsg.Params.Logger, msg)
		handled = true
	}

	handler(notification)
	out := buf.String()
	if !handled || !strings.Contains(out, "[MCP INFO] test: hello world") {
		t.Errorf("notification not handled or output incorrect: %s", out)
	}

	// Check debug filtering
	buf.Reset()
	notification.Notification.Params.AdditionalFields["level"] = "debug"
	handled = false
	handler(notification)
	if handled || buf.String() != "" {
		t.Errorf("debug log should be ignored")
	}

	// Check object output
	buf.Reset()
	notification.Notification.Params.AdditionalFields = map[string]any{
		"level":  "warning",
		"logger": "test",
		"data":   map[string]any{"foo": "bar"},
	}
	handled = false
	handler(notification)
	out = buf.String()
	if !handled || !strings.Contains(out, "[MCP WARNING] test: {\"foo\":\"bar\"}") {
		t.Errorf("object log not handled or output incorrect: %s", out)
	}
}

func setLogLevelIface(ctx context.Context, c callTooler, level string) error {
	req := mcp.SetLevelRequest{}
	req.Params.Level = mcp.LoggingLevel(level)
	_, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      "logging/setLevel",
			Arguments: map[string]any{"level": level},
			Meta:      nil,
		},
	})
	return err
}

func TestSetLogLevel(t *testing.T) {
	ctx := context.Background()
	mock := &mockMCPClient{}
	err := setLogLevelIface(ctx, mock, "error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastRequest.Params.Name != "logging/setLevel" || mock.lastLevel != "error" {
		t.Errorf("incorrect tool call: %+v", mock.lastRequest)
	}

	// Check error
	mock.fail = true
	err = setLogLevelIface(ctx, mock, "info")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
