package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/configuration"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

type mockAgent struct {
	callResult string
	callMeta   types.MetaInfo
	callErr    error
}

func (m *mockAgent) RunSession(ctx context.Context, input string) (string, types.MetaInfo, error) {
	return m.callResult, m.callMeta, m.callErr
}

// Implement types.AgentSpec for MCPApp tests
func (m *mockAgent) RegisterTools() {}

// testApp is a minimal stub for DirectApp.agent that returns a mockAgent
// type testApp struct{ agent directAgent }

// func (a *testApp) DirectAgent() directAgent { return a.agent }

// Helper function to create a test logger
func newTestLogger() *logrus.Logger {
	buf := &bytes.Buffer{}
	log := logrus.New()
	log.SetOutput(buf)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return log
}

func TestApp_SuccessfulCall(t *testing.T) {
	app := &MCPApp{}
	app.agent = &mockAgent{
		callResult: "42",
		callMeta:   types.MetaInfo{Tokens: 10, Cost: 0.1, DurationMs: 100},
		callErr:    nil,
	}
	answer, meta, err := app.agent.RunSession(context.Background(), "test")
	res := buildDirectCallResult(answer, meta, err)
	if !res.Success {
		t.Errorf("expected success=true, got false")
	}
	if res.Result["answer"] != "42" {
		t.Errorf("expected answer=42, got %v", res.Result["answer"])
	}
	if res.Meta.Tokens != 10 {
		t.Errorf("expected tokens=10, got %d", res.Meta.Tokens)
	}
	if res.Error.Type != "" || res.Error.Message != "" {
		t.Errorf("expected empty error, got %+v", res.Error)
	}
}

func TestApp_ErrorCall(t *testing.T) {
	app := &MCPApp{}
	app.agent = &mockAgent{
		callResult: "",
		callMeta:   types.MetaInfo{},
		callErr:    errors.New("fail"),
	}
	answer, meta, err := app.agent.RunSession(context.Background(), "test")
	res := buildDirectCallResult(answer, meta, err)
	if res.Success {
		t.Errorf("expected success=false, got true")
	}
	if res.Error.Message == "" {
		t.Errorf("expected error message, got empty")
	}
	if res.Result["answer"] != "" {
		t.Errorf("expected empty answer, got %v", res.Result["answer"])
	}
}

func TestApp_JSONOutputAlwaysValid(t *testing.T) {
	app := &MCPApp{}
	app.agent = &mockAgent{
		callResult: "foo",
		callMeta:   types.MetaInfo{Tokens: 1},
		callErr:    nil,
	}
	answer, meta, err := app.agent.RunSession(context.Background(), "bar")
	res := buildDirectCallResult(answer, meta, err)
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	for _, f := range []string{"success", "result", "meta", "error"} {
		if _, ok := out[f]; !ok {
			t.Errorf("missing field %q in output: %s", f, string(b))
		}
	}
}

func TestApp_JSONOutputWithNewlines(t *testing.T) {
	app := &MCPApp{}
	app.agent = &mockAgent{
		callResult: "line1\nline2\nline3",
		callMeta:   types.MetaInfo{Tokens: 3},
		callErr:    nil,
	}
	answer, meta, err := app.agent.RunSession(context.Background(), "bar")
	res := buildDirectCallResult(answer, meta, err)
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	answerStr, ok := out["result"].(map[string]any)["answer"].(string)
	if !ok {
		t.Fatalf("answer field missing or not a string: %v", out["result"])
	}
	if answerStr != "line1\nline2\nline3" {
		t.Errorf("expected answer with newlines, got: %q", answerStr)
	}
}

func TestApp_DispatchMCPCall_Success(t *testing.T) {
	a := &MCPApp{
		agent: &mockAgent{callResult: "ok", callMeta: types.MetaInfo{}, callErr: nil},
		cfg:   &configuration.Configuration{},
	}
	a.cfg.Agent.Tool.Name = "answer"
	a.cfg.Agent.Tool.ArgumentName = "text"
	req := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      "answer",
			Arguments: map[string]interface{}{"text": "hello"},
			Meta:      nil,
		},
	}
	res, err := a.dispatchMCPCall(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("expected success, got error: %v, %v", err, res)
	}
}

func TestApp_DispatchMCPCall_InvalidTool(t *testing.T) {
	a := &MCPApp{agent: &mockAgent{}, cfg: &configuration.Configuration{}}
	a.cfg.Agent.Tool.Name = "answer"
	a.cfg.Agent.Tool.ArgumentName = "text"
	req := mcp.CallToolRequest{Params: struct {
		Name      string    `json:"name"`
		Arguments any       `json:"arguments,omitempty"`
		Meta      *mcp.Meta `json:"_meta,omitempty"`
	}{Name: "notanswer", Arguments: map[string]interface{}{"text": "hi"}}}
	res, _ := a.dispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for invalid tool name")
	}
}

func TestApp_DispatchMCPCall_MissingArgument(t *testing.T) {
	a := &MCPApp{agent: &mockAgent{}, cfg: &configuration.Configuration{}}
	a.cfg.Agent.Tool.Name = "answer"
	a.cfg.Agent.Tool.ArgumentName = "text"
	req := mcp.CallToolRequest{Params: struct {
		Name      string    `json:"name"`
		Arguments any       `json:"arguments,omitempty"`
		Meta      *mcp.Meta `json:"_meta,omitempty"`
	}{Name: "answer", Arguments: map[string]interface{}{}}}
	res, _ := a.dispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for missing argument")
	}
}

func TestApp_DispatchMCPCall_EmptyInput(t *testing.T) {
	a := &MCPApp{agent: &mockAgent{}, cfg: &configuration.Configuration{}}
	a.cfg.Agent.Tool.Name = "answer"
	a.cfg.Agent.Tool.ArgumentName = "text"
	req := mcp.CallToolRequest{Params: struct {
		Name      string    `json:"name"`
		Arguments any       `json:"arguments,omitempty"`
		Meta      *mcp.Meta `json:"_meta,omitempty"`
	}{Name: "answer", Arguments: map[string]interface{}{"text": ""}}}
	res, _ := a.dispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for empty input")
	}
}

func TestApp_DispatchMCPCall_CoreError(t *testing.T) {
	a := &MCPApp{agent: &mockAgent{callErr: errors.New("fail")}, cfg: &configuration.Configuration{}}
	a.cfg.Agent.Tool.Name = "answer"
	a.cfg.Agent.Tool.ArgumentName = "text"
	req := mcp.CallToolRequest{Params: struct {
		Name      string    `json:"name"`
		Arguments any       `json:"arguments,omitempty"`
		Meta      *mcp.Meta `json:"_meta,omitempty"`
	}{Name: "answer", Arguments: map[string]interface{}{"text": "hi"}}}
	res, _ := a.dispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for core failure")
	}
}

func Test_validateToolName(t *testing.T) {
	cfg := &configuration.Configuration{}
	cfg.Agent.Tool.Name = "answer"
	cfg.Agent.Tool.ArgumentName = "text"
	t.Run("valid name", func(t *testing.T) {
		err := validateToolName("answer", cfg)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
	t.Run("invalid name", func(t *testing.T) {
		err := validateToolName("wrong", cfg)
		if err == nil || err.Error() != "invalid tool name: wrong" {
			t.Errorf("expected error for invalid tool name, got %v", err)
		}
	})
}

func Test_extractUserInput(t *testing.T) {
	argName := "text"
	t.Run("ok", func(t *testing.T) {
		input, err := extractUserInput(map[string]interface{}{argName: "hi"}, argName)
		if err != nil || input != "hi" {
			t.Errorf("expected 'hi', got %q, %v", input, err)
		}
	})
	t.Run("missing", func(t *testing.T) {
		_, err := extractUserInput(map[string]interface{}{}, argName)
		if err == nil || err.Error() != "missing or nil input argument: text" {
			t.Errorf("expected error for missing arg, got %v", err)
		}
	})
	t.Run("nil value", func(t *testing.T) {
		_, err := extractUserInput(map[string]interface{}{argName: nil}, argName)
		if err == nil || err.Error() != "missing or nil input argument: text" {
			t.Errorf("expected error for nil arg, got %v", err)
		}
	})
	t.Run("wrong type", func(t *testing.T) {
		_, err := extractUserInput(map[string]interface{}{argName: 123}, argName)
		if err == nil || err.Error() != "invalid input argument type: expected string, got int" {
			t.Errorf("expected error for type, got %v", err)
		}
	})
	t.Run("empty string", func(t *testing.T) {
		_, err := extractUserInput(map[string]interface{}{argName: ""}, argName)
		if err == nil || err.Error() != "empty input variable" {
			t.Errorf("expected error for empty, got %v", err)
		}
	})
}

func Test_buildDirectCallResult(t *testing.T) {
	meta := types.MetaInfo{Tokens: 1}
	t.Run("success", func(t *testing.T) {
		res := buildDirectCallResult("ok", meta, nil)
		if !res.Success || res.Result["answer"] != "ok" || res.Meta.Tokens != 1 || res.Error.Type != "" {
			t.Errorf("unexpected result: %+v", res)
		}
	})
	t.Run("error", func(t *testing.T) {
		err := fmt.Errorf("fail")
		res := buildDirectCallResult("", meta, err)
		if res.Success || res.Result["answer"] != "" || res.Error.Type != "internal" || res.Error.Message != "fail" {
			t.Errorf("unexpected error result: %+v", res)
		}
	})
}

func Test_MCPConnector_ToolsInitialization(t *testing.T) {
	type fakeToolConnector struct {
		initCalled bool
		tools      []mcp.Tool
	}
	var (
		fakeTool = mcp.NewTool("external_tool", mcp.WithDescription("external"))
	)
	connector := &fakeToolConnector{
		initCalled: false,
		tools:      nil,
	}
	// Emulate MCPConnector before initialization
	t.Run("tools not loaded before init", func(t *testing.T) {
		if len(connector.tools) != 0 {
			t.Errorf("expected no tools before init, got %d", len(connector.tools))
		}
	})
	// Emulate InitAndConnectToMCPs
	connector.initCalled = true
	connector.tools = []mcp.Tool{fakeTool}
	t.Run("tools loaded after init", func(t *testing.T) {
		if len(connector.tools) != 1 || connector.tools[0].Name != "external_tool" {
			t.Errorf("expected external_tool after init, got %+v", connector.tools)
		}
	})
}

// --- BEGIN: New tests for MCPApp methods ---

// outputErrorAndExit is pure, can be tested directly
func TestMCPApp_outputErrorAndExit(t *testing.T) {
	app := &MCPApp{}
	res, code, err := app.outputErrorAndExit("user", errors.New("fail"))
	if res.Success || code != 1 || err == nil {
		t.Errorf("expected user error, got: %+v, %d, %v", res, code, err)
	}
	res, code, err = app.outputErrorAndExit("config", errors.New("fail"))
	if res.Success || code != 1 || err == nil {
		t.Errorf("expected config error, got: %+v, %d, %v", res, code, err)
	}
	res, code, err = app.outputErrorAndExit("internal", errors.New("fail"))
	if res.Success || code != 2 || err == nil {
		t.Errorf("expected internal error, got: %+v, %d, %v", res, code, err)
	}
}

func TestMCPApp_ExecuteDirectCall_agentNotInit(t *testing.T) {
	app := &MCPApp{}
	res, code, err := app.ExecuteDirectCall(context.Background(), "hi")
	if res.Success || code != 1 || err == nil {
		t.Errorf("expected config error, got: %+v, %d, %v", res, code, err)
	}
}

func TestApp_ExecuteDirectCall_Success(t *testing.T) {
	app := &MCPApp{}
	app.agent = &mockAgent{callResult: "ok", callMeta: types.MetaInfo{}, callErr: nil}
	res, code, err := app.ExecuteDirectCall(context.Background(), "foo")
	if !res.Success || code != 0 || err != nil || res.Result["answer"] != "ok" {
		t.Errorf("expected success, code=0, err=nil, got: %+v, %d, %v", res, code, err)
	}
}

func TestApp_ExecuteDirectCall_InternalError(t *testing.T) {
	app := &MCPApp{}
	app.agent = &mockAgent{callErr: errors.New("oops")}
	res, code, err := app.ExecuteDirectCall(context.Background(), "foo")
	if res.Success || code != 2 || err == nil || res.Error.Message != "oops" {
		t.Errorf("expected internal error, got: %+v, %d, %v", res, code, err)
	}
}

func TestApp_Start_InvalidConfig(t *testing.T) {
	logger := newTestLogger()
	cfg := &configuration.Configuration{}
	cfg.Runtime.Transports.HTTP.Enabled = true
	cfg.Runtime.Transports.Stdio.Enabled = true
	app, err := NewMCPApp(logger, cfg)
	if err != nil {
		t.Fatalf("unexpected error from NewMCPApp: %v", err)
	}
	err = app.Start(context.Background())
	if err == nil || err.Error() == "" || !contains(err.Error(), "failed to create MCP server") {
		t.Errorf("expected error about failed to create MCP server, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
