// Package llm_service provides functionality for interacting with large language model (LLM) services.
// Responsibility: Interacting with various LLM providers (OpenAI, Anthropic)
// Features: Sends requests, processes responses, formats prompts, supports retry strategy
package llm

import (
	"context"
	"fmt"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/llm/cost"
	llmtypes "github.com/korchasa/speelka-agent-go/internal/llm/types"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/korchasa/speelka-agent-go/internal/utils/dump"
	"github.com/korchasa/speelka-agent-go/internal/utils/tools"
	"strings"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/error_handling"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

// LLMService implements the contracts.LLMServiceSpec interface
// Responsibility: Providing a unified API for working with different LLM services
// Features: Encapsulates settings and client for a specific LLM provider
type LLMService struct {
	config     configuration.LLMConfig
	client     llms.Model
	logger     loggerSpec
	calculator calculatorSpec
}

type calculatorSpec interface {
	// CalculateLLMResponse returns the number of tokens, USD cost, and approximation flag for the given model and LLM response.
	CalculateLLMResponse(modelName string, resp llmtypes.LLMResponse) (tokens int, cost float64, isApprox bool, err error)
}

type loggerSpec interface {
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NewLLMService creates a new instance of LLMService
// Responsibility: Factory method for creating an LLM service
// Features: Returns an uninitialized service that requires Initialize to be called
func NewLLMService(cfg configuration.LLMConfig, logger loggerSpec) (*LLMService, error) {
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

	s.calculator = cost.NewCalculator()

	return s, nil

}

// SendRequest sends a request to the LLM with the given prompt and tools
// Responsibility: Communication with the LLM API and getting a response
// Features: Uses a retry strategy to handle transient errors
func (s *LLMService) SendRequest(ctx context.Context, messages []llms.MessageContent, toolsForLLM []mcp.Tool) (llmtypes.LLMResponse, error) {
	if s.client == nil {
		return llmtypes.LLMResponse{}, error_handling.NewError(
			"LLM service not initialized",
			error_handling.ErrorCategoryValidation,
		)
	}

	llmTools, err := tools.ConvertToolsToLLM(toolsForLLM)
	if err != nil {
		return llmtypes.LLMResponse{}, error_handling.WrapError(
			err,
			"failed to convert tools to LLM tools",
			error_handling.ErrorCategoryInternal,
		)
	}

	// Measure duration
	startTime := time.Now()

	// Define a function that performs the request sending
	var response *llms.ContentResponse
	var message string
	var llmsCalls []llms.ToolCall
	sendFn := func() error {
		var err error
		// Prepare options for LLM
		options := []llms.CallOption{
			llms.WithTools(llmTools),
			llms.WithToolChoice("required"),
		}
		// Only add temperature if it was explicitly set in the environment
		if s.config.IsTemperatureSet {
			options = append(options, llms.WithTemperature(s.config.Temperature))
		}
		// Add max tokens if it was explicitly set and is greater than 0
		if s.config.IsMaxTokensSet && s.config.MaxTokens > 0 {
			options = append(options, llms.WithMaxTokens(s.config.MaxTokens))
		}

		// Compose detailed logging of messages
		var msgDetails []string
		for _, m := range messages {
			var partDetails []string
			for _, p := range m.Parts {
				partDetails = append(partDetails, fmt.Sprintf("%T: %v", p, p))
			}
			msgDetails = append(msgDetails, fmt.Sprintf("[%s] %s", m.Role, strings.Join(partDetails, ", ")))
		}
		joinedDetails := ""
		if len(msgDetails) > 0 {
			joinedDetails = " | Messages: " + strings.Join(msgDetails, "; ")
		}
		s.logger.Infof(
			">> [LLM] Calling GenerateContent (model=%s, provider=%s)%s...",
			s.config.Model,
			s.config.Provider,
			joinedDetails,
		)
		startGen := time.Now()
		response, err = s.client.GenerateContent(ctx, messages, options...)
		genDuration := time.Since(startGen)
		if err != nil {
			s.logger.Errorf("<< [LLM] GenerateContent error after %v: %v", genDuration, err)
			// Wrap the error to categorize it as transient for retry attempts
			return error_handling.WrapError(
				err,
				"failed to send request to LLM",
				error_handling.ErrorCategoryTransient,
			)
		}
		s.logger.Infof("<< [LLM] GenerateContent success after %v", genDuration)
		s.logger.Debugf("<< LLM response received with %d choices: %s", len(response.Choices), dump.SDump(response))
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
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		// Clean confidential information from the error
		sanitizedErr := error_handling.SanitizeError(err)
		return llmtypes.LLMResponse{}, sanitizedErr
	}

	calls := make([]types.CallToolRequest, len(llmsCalls))
	for i, call := range llmsCalls {
		calls[i], err = types.NewCallToolRequest(call)
		if err != nil {
			return llmtypes.LLMResponse{}, error_handling.WrapError(
				err,
				"failed to create CallToolRequest",
				error_handling.ErrorCategoryInternal,
			)
		}
	}

	// Extract token usage from GenerationInfo if available
	var completionTokens, promptTokens, reasoningTokens, totalTokens int
	if response != nil && len(response.Choices) > 0 {
		genInfo := response.Choices[0].GenerationInfo
		if genInfo != nil {
			if v, ok := genInfo["CompletionTokens"]; ok {
				if n, ok := v.(int); ok {
					completionTokens = n
				} else if f, ok := v.(float64); ok {
					completionTokens = int(f)
				}
			}
			if v, ok := genInfo["PromptTokens"]; ok {
				if n, ok := v.(int); ok {
					promptTokens = n
				} else if f, ok := v.(float64); ok {
					promptTokens = int(f)
				}
			}
			if v, ok := genInfo["ReasoningTokens"]; ok {
				if n, ok := v.(int); ok {
					reasoningTokens = n
				} else if f, ok := v.(float64); ok {
					reasoningTokens = int(f)
				}
			}
			if v, ok := genInfo["TotalTokens"]; ok {
				if n, ok := v.(int); ok {
					totalTokens = n
				} else if f, ok := v.(float64); ok {
					totalTokens = int(f)
				}
			}
		}
	}

	tokensMetadata := llmtypes.LLMResponseTokensMetadata{
		CompletionTokens: completionTokens,
		PromptTokens:     promptTokens,
		ReasoningTokens:  reasoningTokens,
		TotalTokens:      totalTokens,
	}

	// Compose and return the response
	llmResp := llmtypes.LLMResponse{
		RequestMessages: messages,
		Text:            message,
		Calls:           calls,
		Metadata: llmtypes.LLMResponseMetadata{
			Tokens:     tokensMetadata,
			DurationMs: durationMs,
		},
	}
	if s.calculator != nil {
		_, amount, _, err := s.calculator.CalculateLLMResponse(s.config.Model, llmResp)
		if err != nil {
			s.logger.Warnf("Failed to calculate cost: %v", err)
			amount = 0
		}
		llmResp.Metadata.Cost = amount
	}
	return llmResp, nil
}
