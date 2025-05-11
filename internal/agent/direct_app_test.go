package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type mockAgent struct {
	callResult string
	callMeta   MetaInfo
	callErr    error
}

func (m *mockAgent) CallDirect(ctx context.Context, input string) (string, MetaInfo, error) {
	return m.callResult, m.callMeta, m.callErr
}

func TestDirectApp_SuccessfulCall(t *testing.T) {
	dapp := &DirectApp{
		Agent: &mockAgent{
			callResult: "42",
			callMeta:   MetaInfo{Tokens: 10, Cost: 0.1, DurationMs: 100},
			callErr:    nil,
		},
	}
	res := dapp.HandleCall(context.Background(), "test")
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

func TestDirectApp_ErrorCall(t *testing.T) {
	dapp := &DirectApp{
		Agent: &mockAgent{
			callResult: "",
			callMeta:   MetaInfo{},
			callErr:    errors.New("fail"),
		},
	}
	res := dapp.HandleCall(context.Background(), "test")
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

func TestDirectApp_JSONOutputAlwaysValid(t *testing.T) {
	dapp := &DirectApp{
		Agent: &mockAgent{
			callResult: "foo",
			callMeta:   MetaInfo{Tokens: 1},
			callErr:    nil,
		},
	}
	res := dapp.HandleCall(context.Background(), "bar")
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

func TestDirectApp_JSONOutputWithNewlines(t *testing.T) {
	dapp := &DirectApp{
		Agent: &mockAgent{
			callResult: "line1\nline2\nline3",
			callMeta:   MetaInfo{Tokens: 3},
			callErr:    nil,
		},
	}
	res := dapp.HandleCall(context.Background(), "bar")
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
