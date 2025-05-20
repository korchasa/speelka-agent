// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/tmc/langchaingo/llms"
)

// LLMResponse represents the response from the LLMService, including text, tool calls, and token usage.
type LLMResponse struct {
	// RequestMessages stores the original messages array sent to the LLM.
	RequestMessages []llms.MessageContent
	// Text is the main response from the LLM.
	Text string
	// Calls is the list of tool/function calls returned by the LLM.
	Calls []types.CallToolRequest
	// Metadata contains metadata about the LLM response.
	Metadata LLMResponseMetadata
}

type LLMResponseMetadata struct {
	Tokens     LLMResponseTokensMetadata
	Cost       float64
	DurationMs int64 // Duration of the LLM request in milliseconds
}

// LLMResponseTokensMetadata represents metadata about the LLM response.
type LLMResponseTokensMetadata struct {
	// CompletionTokens is the number of tokens in the completion/response.
	CompletionTokens int
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int
	// ReasoningTokens is the number of tokens used for reasoning (if available).
	ReasoningTokens int
	// TotalTokens is the total number of tokens used.
	TotalTokens int
}
