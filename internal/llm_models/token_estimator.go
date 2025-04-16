package llm_models

import (
	"encoding/json"

	"github.com/tmc/langchaingo/llms"
)

// TokenEstimator provides a shared utility for counting tokens in LLM messages.
type TokenEstimator struct{}

// CountTokens returns the number of tokens in the given message for the specified model.
func (TokenEstimator) CountTokens(message llms.MessageContent) int {
	text := ExtractTextFromMessageForApprox(message)
	if len(text) == 0 {
		return 0
	}
	est := len(text) / 4
	if est < 1 {
		est = 1
	}
	return est
}

// ExtractTextFromMessageForApprox attempts to extract the main text from llms.MessageContent for fallback estimation.
func ExtractTextFromMessageForApprox(msg llms.MessageContent) string {
	for _, part := range msg.Parts {
		// Try to extract from known struct types (e.g., TextContent)
		if txt, ok := part.(interface{ GetText() string }); ok {
			return txt.GetText()
		}
	}
	// Fallback: try to serialize the message and use its length
	jsonStr, err := json.Marshal(msg)
	if err == nil {
		return string(jsonStr)
	}
	return ""
}
