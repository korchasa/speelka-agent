// Package llm_service provides functionality for interacting with large language model (LLM) services.
// Responsibility: Interacting with various LLM providers (OpenAI, Anthropic)
// Features: Sends requests, processes responses, formats prompts, supports retry strategy
package llm_service

import (
	"context"
	"fmt"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/error_handling"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

// LLMService implements the contracts.LLMServiceSpec interface
// Responsibility: Providing a unified API for working with different LLM services
// Features: Encapsulates settings and client for a specific LLM provider
type LLMService struct {
	config types.LLMConfig
	client llms.Model
	logger types.LoggerSpec
}

// NewLLMService creates a new instance of LLMService
// Responsibility: Factory method for creating an LLM service
// Features: Returns an uninitialized service that requires Initialize to be called
func NewLLMService(cfg types.LLMConfig, logger types.LoggerSpec) (*LLMService, error) {
	s := &LLMService{
		config: cfg,
		logger: logger,
	}

	if cfg.Provider == "" {
		return nil, error_handling.NewError(
			"provider is required",
			error_handling.ErrorCategoryValidation,
		)
	}
	if cfg.Model == "" {
		return nil, error_handling.NewError(
			"model is required",
			error_handling.ErrorCategoryValidation,
		)
	}
	if cfg.APIKey == "" {
		return nil, error_handling.NewError(
			"API key is required",
			error_handling.ErrorCategoryValidation,
		)
	}

	// Initialize the appropriate client based on the provider
	var err error
	switch cfg.Provider {
	case "openai":
		s.client, err = openai.New(
			openai.WithToken(s.config.APIKey),
			openai.WithModel(s.config.Model),
		)
		if err != nil {
			return nil, error_handling.WrapError(
				err,
				"failed to initialize OpenAI client",
				error_handling.ErrorCategoryInternal,
			)
		}
	case "anthropic":
		s.client, err = anthropic.New(
			anthropic.WithToken(s.config.APIKey),
			anthropic.WithModel(s.config.Model),
		)
		if err != nil {
			return nil, error_handling.WrapError(
				err,
				"failed to initialize Anthropic client",
				error_handling.ErrorCategoryInternal,
			)
		}
	default:
		return nil, error_handling.NewError(
			fmt.Sprintf("unsupported provider: %s", s.config.Provider),
			error_handling.ErrorCategoryValidation,
		)
	}

	return s, nil

}

// SendRequest sends a request to the LLM with the given prompt and tools
// Responsibility: Communication with the LLM API and getting a response
// Features: Uses a retry strategy to handle transient errors
func (s *LLMService) SendRequest(ctx context.Context, messages []llms.MessageContent, tools []mcp.Tool) (string, []types.CallToolRequest, error) {
	if s.client == nil {
		return "", nil, error_handling.NewError(
			"LLM service not initialized",
			error_handling.ErrorCategoryValidation,
		)
	}

	llmTools, err := types.ConvertToolsToLLM(tools)
	if err != nil {
		return "", nil, error_handling.WrapError(
			err,
			"failed to convert tools to LLM tools",
			error_handling.ErrorCategoryInternal,
		)
	}

	// Define a function that performs the request sending
	var response *llms.ContentResponse
	var message string
	var llmsCalls []llms.ToolCall
	sendFn := func() error {
		var err error
		s.logger.Infof("Send request to LLM with %d messages", len(messages))
		s.logger.Debugf("Details: %s", logger.SDump(map[string]any{"messages": messages, "tools": llmTools}))
		// Prepare options for LLM
		options := []llms.CallOption{
			llms.WithTools(llmTools),
			llms.WithToolChoice("required"),
		}

		// Only add temperature if it was explicitly set in the environment
		if s.config.IsTemperatureSet {
			s.logger.Debugf("Using explicitly set temperature: %f", s.config.Temperature)
			options = append(options, llms.WithTemperature(s.config.Temperature))
		}

		// Add max tokens if it was explicitly set and is greater than 0
		if s.config.IsMaxTokensSet && s.config.MaxTokens > 0 {
			s.logger.Debugf("Using explicitly set max tokens: %d", s.config.MaxTokens)
			options = append(options, llms.WithMaxTokens(s.config.MaxTokens))
		}

		response, err = s.client.GenerateContent(ctx, messages, options...)
		if err != nil {
			// Wrap the error to categorize it as transient for retry attempts
			return error_handling.WrapError(
				err,
				"failed to send request to LLM",
				error_handling.ErrorCategoryTransient,
			)
		}
		s.logger.Infof("LLM response received with %d choices", len(response.Choices))
		s.logger.Debugf("Details: %s", logger.SDump(response))
		if len(response.Choices) == 0 {
			return error_handling.NewError(
				"empty response from LLM",
				error_handling.ErrorCategoryUnknown,
			)
		}
		ch := response.Choices[0]
		if ch.FuncCall == nil {
			return error_handling.NewError(
				"no function call in response",
				error_handling.ErrorCategoryUnknown,
			)
		}
		llmsCalls = ch.ToolCalls
		message = ch.Content
		return nil
	}

	// Use retry with exponential backoff for transient errors
	err = error_handling.RetryWithBackoff(ctx, sendFn, error_handling.RetryConfig{
		MaxRetries:        s.config.RetryConfig.MaxRetries,
		InitialBackoff:    time.Duration(s.config.RetryConfig.InitialBackoff * float64(time.Second)),
		BackoffMultiplier: s.config.RetryConfig.BackoffMultiplier,
		MaxBackoff:        time.Duration(s.config.RetryConfig.MaxBackoff * float64(time.Second)),
	})
	if err != nil {
		// Clean confidential information from the error
		sanitizedErr := error_handling.SanitizeError(err)
		return "", nil, sanitizedErr
	}

	calls := make([]types.CallToolRequest, len(llmsCalls))
	for i, call := range llmsCalls {
		calls[i], err = types.NewCallToolRequest(call)
		if err != nil {
			return "", nil, error_handling.WrapError(
				err,
				"failed to create CallToolRequest",
				error_handling.ErrorCategoryInternal,
			)
		}
	}

	return message, calls, nil
}
