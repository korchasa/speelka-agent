package log_formatter

import (
	"bytes"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestCustomLogFormatter_Format_basic(t *testing.T) {
	f := &CustomLogFormatter{}
	entry := &logrus.Entry{
		Message: "hello",
		Level:   logrus.InfoLevel,
		Time:    time.Now(),
		Data:    map[string]interface{}{},
	}
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(out, []byte("hello")) {
		t.Error("message not found in output")
	}
}

func TestCustomLogFormatter_appendKeyValue(t *testing.T) {
	var b bytes.Buffer
	f := &CustomLogFormatter{}
	f.appendKeyValue(&b, "foo", "bar")
	if b.String() != "foo=bar" {
		t.Errorf("unexpected output: %s", b.String())
	}
}

func TestNewCustomLogFormatter(t *testing.T) {
	f := NewCustomLogFormatter()
	if f == nil {
		t.Fatal("NewCustomLogFormatter returned nil")
	}
	if f.TimestampFormat != "15:04:05" {
		t.Errorf("unexpected TimestampFormat: %s", f.TimestampFormat)
	}
	if f.FullTimestamp != false {
		t.Errorf("unexpected FullTimestamp: %v", f.FullTimestamp)
	}
}
