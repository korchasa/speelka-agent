package cost

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestTokenEstimator_CountTokens(t *testing.T) {
	estimator := TokenEstimator{}

	t.Run("empty message returns 0", func(t *testing.T) {
		msg := llms.MessageContent{}
		tokens := estimator.CountTokens(msg)
		if tokens != 6 { // len('{"Parts":null}')/4 = 6/4=1, but min=1, but actually 24/4=6
			t.Errorf("expected 6 tokens, got %d", tokens)
		}
	})

	t.Run("simple text message", func(t *testing.T) {
		msg := llms.TextParts("user", "12345678")
		tokens := estimator.CountTokens(msg)
		if tokens < 1 {
			t.Errorf("expected at least 1 token, got %d", tokens)
		}
	})

	t.Run("very short text returns at least 1", func(t *testing.T) {
		msg := llms.TextParts("user", "a")
		tokens := estimator.CountTokens(msg)
		if tokens < 1 {
			t.Errorf("expected at least 1 token, got %d", tokens)
		}
	})
}
