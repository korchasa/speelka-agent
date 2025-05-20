package cost

import (
	"github.com/korchasa/speelka-agent-go/internal/llm/types"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestCalculator_CalculateLLMResponse(t *testing.T) {
	calc := NewCalculator()

	t.Run("returns correct tokens and cost for known model with full token metadata", func(t *testing.T) {
		resp := types.LLMResponse{
			Text: "Hello world!",
			Metadata: types.LLMResponseMetadata{
				Tokens: types.LLMResponseTokensMetadata{
					PromptTokens:     1000,
					CompletionTokens: 500,
					TotalTokens:      1500,
				},
			},
		}
		tokens, cost, approx, err := calc.CalculateLLMResponse("gpt-4o", resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens != 1500 {
			t.Errorf("expected 1500 tokens, got %d", tokens)
		}
		// gpt-4o: $0.0025/1K prompt, $0.01/1K completion
		// cost = 1000*0.0025/1000 + 500*0.01/1000 = 0.0025 + 0.005 = 0.0075
		if cost < 0.0074 || cost > 0.0076 {
			t.Errorf("expected cost ~0.0075, got %f", cost)
		}
		if approx {
			t.Error("expected approx=false for full token info")
		}
	})

	t.Run("returns error for unknown model", func(t *testing.T) {
		resp := types.LLMResponse{
			Text: "irrelevant",
			Metadata: types.LLMResponseMetadata{
				Tokens: types.LLMResponseTokensMetadata{
					PromptTokens:     10,
					CompletionTokens: 10,
					TotalTokens:      20,
				},
			},
		}
		_, _, _, err := calc.CalculateLLMResponse("nonexistent-model", resp)
		if err == nil {
			t.Error("expected error for unknown model, got nil")
		}
	})

	t.Run("estimates tokens and cost if token metadata is missing", func(t *testing.T) {
		resp := types.LLMResponse{
			Text: "This is a test message with 40 chars.",
			RequestMessages: []llms.MessageContent{
				llms.TextParts("user", "input message with 36 chars."),
			},
			Metadata: types.LLMResponseMetadata{
				Tokens: types.LLMResponseTokensMetadata{},
			},
		}
		tokens, cost, approx, err := calc.CalculateLLMResponse("gpt-4o", resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens <= 0 {
			t.Errorf("expected positive token count, got %d", tokens)
		}
		if cost <= 0 {
			t.Errorf("expected positive cost, got %f", cost)
		}
		if !approx {
			t.Error("expected approx=true for fallback estimation")
		}
	})

	t.Run("zero tokens returns zero cost and sets approx", func(t *testing.T) {
		resp := types.LLMResponse{
			Text: "",
			Metadata: types.LLMResponseMetadata{
				Tokens: types.LLMResponseTokensMetadata{},
			},
		}
		tokens, cost, approx, err := calc.CalculateLLMResponse("gpt-4o", resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens != 0 {
			t.Errorf("expected 0 tokens, got %d", tokens)
		}
		if cost != 0 {
			t.Errorf("expected 0 cost, got %f", cost)
		}
		if !approx {
			t.Error("expected approx=true for zero tokens")
		}
	})

	t.Run("partial token info: only prompt tokens", func(t *testing.T) {
		resp := types.LLMResponse{
			Text: "",
			Metadata: types.LLMResponseMetadata{
				Tokens: types.LLMResponseTokensMetadata{
					PromptTokens: 1000,
					TotalTokens:  1000,
				},
			},
		}
		tokens, cost, approx, err := calc.CalculateLLMResponse("gpt-4o", resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens != 1000 {
			t.Errorf("expected 1000 tokens, got %d", tokens)
		}
		if cost < 0.0024 || cost > 0.0026 {
			t.Errorf("expected cost ~0.0025, got %f", cost)
		}
		if approx {
			t.Error("expected approx=false for partial token info")
		}
	})
}

func TestCalculator_CalculateCost(t *testing.T) {
	calc := NewCalculator()

	t.Run("known model", func(t *testing.T) {
		cost, err := calc.CalculateCost("gpt-4o", 1000, 500)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cost < 0.0074 || cost > 0.0076 {
			t.Errorf("expected cost ~0.0075, got %f", cost)
		}
	})

	t.Run("unknown model", func(t *testing.T) {
		_, err := calc.CalculateCost("nonexistent-model", 100, 100)
		if err == nil {
			t.Error("expected error for unknown model, got nil")
		}
	})
}
