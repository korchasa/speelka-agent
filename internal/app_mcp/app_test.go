package app_mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
)

type mockAgent struct {
	callResult string
	callMeta   types.MetaInfo
	callErr    error
}

func (m *mockAgent) CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error) {
	return m.callResult, m.callMeta, m.callErr
}

// Implement types.AgentSpec for App tests
func (m *mockAgent) RegisterTools() {}

// testApp is a minimal stub for DirectApp.agent that returns a mockAgent
// type testApp struct{ agent directAgent }

// func (a *testApp) DirectAgent() directAgent { return a.agent }

func TestApp_SuccessfulCall(t *testing.T) {
	app := &App{}
	app.agent = &mockAgent{
		callResult: "42",
		callMeta:   types.MetaInfo{Tokens: 10, Cost: 0.1, DurationMs: 100},
		callErr:    nil,
	}
	res := app.HandleCall(context.Background(), "test")
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
	app := &App{}
	app.agent = &mockAgent{
		callResult: "",
		callMeta:   types.MetaInfo{},
		callErr:    errors.New("fail"),
	}
	res := app.HandleCall(context.Background(), "test")
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
	app := &App{}
	app.agent = &mockAgent{
		callResult: "foo",
		callMeta:   types.MetaInfo{Tokens: 1},
		callErr:    nil,
	}
	res := app.HandleCall(context.Background(), "bar")
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
	app := &App{}
	app.agent = &mockAgent{
		callResult: "line1\nline2\nline3",
		callMeta:   types.MetaInfo{Tokens: 3},
		callErr:    nil,
	}
	res := app.HandleCall(context.Background(), "bar")
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	answer, ok := out["result"].(map[string]any)["answer"].(string)
	if !ok {
		t.Fatalf("answer field missing or not a string: %v", out["result"])
	}
	if answer != "line1\nline2\nline3" {
		t.Errorf("expected answer with newlines, got: %q", answer)
	}
}

func TestApp_DispatchMCPCall_Success(t *testing.T) {
	a := &App{
		agent:         &mockAgent{callResult: "ok", callMeta: types.MetaInfo{}, callErr: nil},
		configManager: &mockConfigManager{toolName: "answer", argName: "text"},
	}
	req := mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      "answer",
			Arguments: map[string]interface{}{"text": "hello"},
		},
	}
	res, err := a.DispatchMCPCall(context.Background(), req)
	if err != nil || res.IsError {
		t.Errorf("expected success, got error: %v, %v", err, res)
	}
}

func TestApp_DispatchMCPCall_InvalidTool(t *testing.T) {
	a := &App{agent: &mockAgent{}, configManager: &mockConfigManager{toolName: "answer", argName: "text"}}
	req := mcp.CallToolRequest{Params: struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
		Meta      *struct {
			ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
		} `json:"_meta,omitempty"`
	}{Name: "notanswer", Arguments: map[string]interface{}{"text": "hi"}}}
	res, _ := a.DispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for invalid tool name")
	}
}

func TestApp_DispatchMCPCall_MissingArgument(t *testing.T) {
	a := &App{agent: &mockAgent{}, configManager: &mockConfigManager{toolName: "answer", argName: "text"}}
	req := mcp.CallToolRequest{Params: struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
		Meta      *struct {
			ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
		} `json:"_meta,omitempty"`
	}{Name: "answer", Arguments: map[string]interface{}{}}}
	res, _ := a.DispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for missing argument")
	}
}

func TestApp_DispatchMCPCall_EmptyInput(t *testing.T) {
	a := &App{agent: &mockAgent{}, configManager: &mockConfigManager{toolName: "answer", argName: "text"}}
	req := mcp.CallToolRequest{Params: struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
		Meta      *struct {
			ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
		} `json:"_meta,omitempty"`
	}{Name: "answer", Arguments: map[string]interface{}{"text": ""}}}
	res, _ := a.DispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for empty input")
	}
}

func TestApp_DispatchMCPCall_CoreError(t *testing.T) {
	a := &App{agent: &mockAgent{callErr: errors.New("fail")}, configManager: &mockConfigManager{toolName: "answer", argName: "text"}}
	req := mcp.CallToolRequest{Params: struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
		Meta      *struct {
			ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
		} `json:"_meta,omitempty"`
	}{Name: "answer", Arguments: map[string]interface{}{"text": "hi"}}}
	res, _ := a.DispatchMCPCall(context.Background(), req)
	if !res.IsError {
		t.Errorf("expected error for core failure")
	}
}

type mockConfigManager struct{ toolName, argName string }

func (m *mockConfigManager) GetConfiguration() *types.Configuration {
	return &types.Configuration{
		Agent: struct {
			Name    string "json:\"name\" yaml:\"name\""
			Version string "json:\"version\" yaml:\"version\""
			Tool    struct {
				Name                string "json:\"name\" yaml:\"name\""
				Description         string "json:\"description\" yaml:\"description\""
				ArgumentName        string "json:\"argument_name\" yaml:\"argument_name\""
				ArgumentDescription string "json:\"argument_description\" yaml:\"argument_description\""
			} "json:\"tool\" yaml:\"tool\""
			Chat struct {
				MaxTokens        int     "json:\"max_tokens\" yaml:\"max_tokens\""
				MaxLLMIterations int     "json:\"max_llm_iterations\" yaml:\"max_llm_iterations\""
				RequestBudget    float64 "json:\"request_budget\" yaml:\"request_budget\""
			} "json:\"chat\" yaml:\"chat\""
			LLM struct {
				Provider       string  "json:\"provider\" yaml:\"provider\""
				Model          string  "json:\"model\" yaml:\"model\""
				APIKey         string  "json:\"api_key\" yaml:\"api_key\""
				MaxTokens      int     "json:\"max_tokens\" yaml:\"max_tokens\""
				Temperature    float64 "json:\"temperature\" yaml:\"temperature\""
				PromptTemplate string  "json:\"prompt_template\" yaml:\"prompt_template\""
				Retry          struct {
					MaxRetries        int     "json:\"max_retries\" yaml:\"max_retries\""
					InitialBackoff    float64 "json:\"initial_backoff\" yaml:\"initial_backoff\""
					MaxBackoff        float64 "json:\"max_backoff\" yaml:\"max_backoff\""
					BackoffMultiplier float64 "json:\"backoff_multiplier\" yaml:\"backoff_multiplier\""
				} "json:\"retry\" yaml:\"retry\""
				IsMaxTokensSet   bool
				IsTemperatureSet bool
			} "json:\"llm\" yaml:\"llm\""
			Connections struct {
				McpServers map[string]types.MCPServerConnection "json:\"mcpServers\" yaml:\"mcpServers\""
				Retry      struct {
					MaxRetries        int     "json:\"max_retries\" yaml:\"max_retries\""
					InitialBackoff    float64 "json:\"initial_backoff\" yaml:\"initial_backoff\""
					MaxBackoff        float64 "json:\"max_backoff\" yaml:\"max_backoff\""
					BackoffMultiplier float64 "json:\"backoff_multiplier\" yaml:\"backoff_multiplier\""
				} "json:\"retry\" yaml:\"retry\""
			} "json:\"connections\" yaml:\"connections\""
		}{
			Tool: struct {
				Name                string "json:\"name\" yaml:\"name\""
				Description         string "json:\"description\" yaml:\"description\""
				ArgumentName        string "json:\"argument_name\" yaml:\"argument_name\""
				ArgumentDescription string "json:\"argument_description\" yaml:\"argument_description\""
			}{
				Name:         m.toolName,
				ArgumentName: m.argName,
			},
		},
	}
}

func (m *mockConfigManager) LoadConfiguration(ctx context.Context, configFilePath string) error {
	return nil
}

func Test_validateToolName(t *testing.T) {
	cfg := &mockConfigManager{toolName: "answer", argName: "text"}
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
