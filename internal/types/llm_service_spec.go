// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
)

// LLMResponse represents the response from the LLMService, including text, tool calls, and token usage.
type LLMResponse struct {
	// RequestMessages stores the original messages array sent to the LLM.
	RequestMessages []llms.MessageContent
	// Text is the main response from the LLM.
	Text string
	// Calls is the list of tool/function calls returned by the LLM.
	Calls []CallToolRequest
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

// LLMServiceSpec represents the interface for the LLM service.
// Responsibility: Defining the contract for the LLM service
// Features: Defines methods for sending requests to the LLM
type LLMServiceSpec interface {
	// SendRequest sends a request to the LLM with the given prompt and tools.
	// It returns the response struct and an error if the request fails.
	SendRequest(ctx context.Context, messages []llms.MessageContent, tools []mcp.Tool) (LLMResponse, error)
}

// LLMConfig represents the configuration for the LLM service.
// Responsibility: Storing all settings for working with the language model
// Features: Includes parameters for connecting to the provider, model settings, and prompt templates
type LLMConfig struct {
	// Provider - name of the LLM provider (e.g., "openai", "anthropic").
	Provider string

	// Model - name of the LLM model to use.
	Model string

	// APIKey - API key for the LLM provider.
	APIKey string

	// MaxTokens - maximum number of tokens for generation.
	MaxTokens int

	// IsMaxTokensSet - flag indicating if MaxTokens was explicitly set by the user
	IsMaxTokensSet bool

	// Temperature - controls the randomness of the LLM output.
	Temperature float64

	// IsTemperatureSet - flag indicating if Temperature was explicitly set by the user
	IsTemperatureSet bool

	// SystemPromptTemplate - system prompt template.
	SystemPromptTemplate string

	// RetryConfig - configuration for retry attempts on failed requests.
	RetryConfig RetryConfig
}

// RetryConfig represents the configuration for retry attempts on failed requests.
// Responsibility: Configuring the retry strategy
// Features: Defines the number of attempts and wait time between them
type RetryConfig struct {
	// MaxRetries - maximum number of retry attempts.
	MaxRetries int

	// InitialBackoff - initial delay in seconds.
	InitialBackoff float64

	// MaxBackoff - maximum delay in seconds.
	MaxBackoff float64

	// BackoffMultiplier - multiplier for the delay.
	BackoffMultiplier float64
}

// TokenUsage represents information about token usage.
// Responsibility: Tracking the number of tokens used in LLM requests
// Features: Tracks the number of tokens in the request, response, and their sum
type TokenUsage struct {
	// PromptTokens - number of tokens in the prompt.
	PromptTokens int

	// CompletionTokens - number of tokens in the response.
	CompletionTokens int

	// TotalTokens - total number of tokens used.
	TotalTokens int
}
