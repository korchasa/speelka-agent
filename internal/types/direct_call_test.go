package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMetaInfo_JSONFields(t *testing.T) {
	meta := MetaInfo{
		Tokens:           42,
		Cost:             0.1234,
		DurationMs:       1000,
		PromptTokens:     10,
		CompletionTokens: 32,
		ReasoningTokens:  0,
	}
	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	str := string(b)
	if !(strings.Contains(str, "tokens") && strings.Contains(str, "cost") && strings.Contains(str, "duration_ms")) {
		t.Errorf("missing required fields in JSON: %s", str)
	}
}

func TestDirectCallError_JSONFields(t *testing.T) {
	err := DirectCallError{Type: "user", Message: "bad input", Details: map[string]any{"foo": 1}}
	b, err2 := json.Marshal(err)
	if err2 != nil {
		t.Fatalf("marshal failed: %v", err2)
	}
	str := string(b)
	if !(strings.Contains(str, "type") && strings.Contains(str, "message")) {
		t.Errorf("missing required fields in JSON: %s", str)
	}
}

func TestDirectCallResult_AlwaysHasFields(t *testing.T) {
	res := DirectCallResult{
		Success: true,
		Result:  map[string]any{"answer": "hi"},
		Meta:    MetaInfo{Tokens: 1, Cost: 0.1, DurationMs: 10},
		Error:   DirectCallError{},
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	str := string(b)
	for _, f := range []string{"success", "result", "meta", "error"} {
		if !strings.Contains(str, f) {
			t.Errorf("missing field %q in JSON: %s", f, str)
		}
	}
}
