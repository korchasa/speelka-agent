package utils

import (
	"testing"
	"time"
)

func TestSDump(t *testing.T) {
	type foo struct{ Bar string }
	out := SDump(foo{Bar: "baz"})
	if len(out) == 0 || out[0] != '\n' {
		t.Errorf("SDump should start with newline, got: %q", out)
	}
}

func TestCompactJson(t *testing.T) {
	m := map[string]interface{}{"a": 1}
	out := CompactJson(m)
	if out != `{"a":1}` {
		t.Errorf("unexpected json: %s", out)
	}
	bad := func() {}
	if CompactJson(bad) != "{}" {
		t.Error("CompactJson should return {} on marshal error")
	}
}

func TestJoin(t *testing.T) {
	arr := []string{"a", "b", "c"}
	if Join(arr, "-") != "a-b-c" {
		t.Error("Join failed")
	}
}

func TestSinceMs(t *testing.T) {
	start := time.Now().Add(-time.Millisecond * 10)
	if SinceMs(start) < 0 {
		t.Error("SinceMs should not be negative")
	}
}

func TestNow(t *testing.T) {
	now := Now()
	if time.Since(now) > time.Second {
		t.Error("Now returned too old time")
	}
}
