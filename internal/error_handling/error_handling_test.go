package error_handling

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"
)

func TestNewErrorAndWrapError(t *testing.T) {
	err := NewError("fail", ErrorCategoryValidation)
	if err == nil {
		t.Fatal("NewError returned nil")
	}
	if err.Error() != "fail" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
	if err.Category() != ErrorCategoryValidation {
		t.Errorf("unexpected category: %v", err.Category())
	}
	if err.Cause() != nil {
		t.Errorf("expected nil cause, got: %v", err.Cause())
	}

	base := errors.New("root")
	wrap := WrapError(base, "context", ErrorCategoryInternal)
	if wrap == nil {
		t.Fatal("WrapError returned nil")
	}
	if wrap.Error() != "context: root" {
		t.Errorf("unexpected wrapped error message: %s", wrap.Error())
	}
	if wrap.Cause() != base {
		t.Errorf("expected cause to be base error")
	}
	if wrap.Category() != ErrorCategoryInternal {
		t.Errorf("unexpected category: %v", wrap.Category())
	}
}

func TestUnwrap(t *testing.T) {
	base := errors.New("root")
	wrap := WrapError(base, "context", ErrorCategoryInternal)
	if !errors.Is(wrap, base) {
		t.Error("errors.Is failed to unwrap base error")
	}
}

func TestIsTransient(t *testing.T) {
	transient := NewError("tmp", ErrorCategoryTransient)
	if !IsTransient(transient) {
		t.Error("IsTransient should return true for transient error")
	}
	nonTransient := NewError("fail", ErrorCategoryValidation)
	if IsTransient(nonTransient) {
		t.Error("IsTransient should return false for non-transient error")
	}
	if IsTransient(errors.New("plain error")) {
		t.Error("IsTransient should return false for plain error")
	}
}

func TestRetryWithBackoff_SuccessFirstTry(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), func() error {
		calls++
		return nil
	}, RetryConfig{MaxRetries: 3, InitialBackoff: 1 * time.Millisecond, BackoffMultiplier: 2, MaxBackoff: 10 * time.Millisecond})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryWithBackoff_TransientThenSuccess(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), func() error {
		calls++
		if calls < 2 {
			return NewError("tmp", ErrorCategoryTransient)
		}
		return nil
	}, RetryConfig{MaxRetries: 3, InitialBackoff: 1 * time.Millisecond, BackoffMultiplier: 2, MaxBackoff: 10 * time.Millisecond})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryWithBackoff_Exhausted(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), func() error {
		calls++
		return NewError("tmp", ErrorCategoryTransient)
	}, RetryConfig{MaxRetries: 2, InitialBackoff: 1 * time.Millisecond, BackoffMultiplier: 2, MaxBackoff: 10 * time.Millisecond})
	if err == nil {
		t.Error("expected error after retries exhausted")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls (1+2 retries), got %d", calls)
	}
	if !IsTransient(err) {
		t.Error("final error should be transient")
	}
	if !regexp.MustCompile(`failed after 2 retries`).MatchString(err.Error()) {
		t.Errorf("expected error message to mention retries, got: %s", err.Error())
	}
}

func TestRetryWithBackoff_NonTransient(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), func() error {
		calls++
		return NewError("fail", ErrorCategoryValidation)
	}, RetryConfig{MaxRetries: 2, InitialBackoff: 1 * time.Millisecond, BackoffMultiplier: 2, MaxBackoff: 10 * time.Millisecond})
	if err == nil {
		t.Error("expected error for non-transient")
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestSanitizeError(t *testing.T) {
	cases := []struct {
		in   error
		want string
	}{
		{errors.New("sk-1234567890abcdef"), "[REDACTED]"},
		{errors.New("Bearer abcdefg12345"), "[REDACTED]"},
		{errors.New("Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="), "[REDACTED]=="},
		{errors.New("password: secret123"), "[REDACTED]"},
		{errors.New("4111 1111 1111 1111"), "[REDACTED]"},
		{errors.New("no sensitive info here"), "no sensitive info here"},
	}
	for i, c := range cases {
		got := SanitizeError(c.in)
		if got.Error() != c.want {
			t.Errorf("case %d: got %q, want %q", i, got.Error(), c.want)
		}
	}
}
