// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

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

    // Temperature - controls the randomness of the LLM output.
    Temperature float64

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
