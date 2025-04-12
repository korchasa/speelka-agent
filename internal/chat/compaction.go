package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// DefaultTokenCountModel model to use for token counting if none is specified
const DefaultTokenCountModel = "gpt-3.5-turbo"

// estimateTokenCount estimates the number of tokens in a message using langchaingo's CountTokens
func estimateTokenCount(message llms.MessageContent, model string) int {
	// Serialize the message to JSON for complete token counting
	jsonStr, err := json.Marshal(message)
	if err != nil {
		return 0
	}

	// Use the specified model or default if none provided
	modelToUse := model
	if modelToUse == "" {
		modelToUse = DefaultTokenCountModel
	}

	// Use langchaingo's CountTokens for accurate counting
	return llms.CountTokens(modelToUse, string(jsonStr))
}

// Logger defines the interface for logging
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// CompactionStrategyDeleteOld Compaction strategy constants
const (
	CompactionStrategyDeleteOld = "delete-old"
)

// CompactionStrategy defines the interface for message history compaction strategies
type CompactionStrategy interface {
	// Compact reduces the token count by removing or summarizing messages
	// Returns the compacted messages and the new token count
	Compact(messages []llms.MessageContent, currentTokens, maxTokens int) ([]llms.MessageContent, int)
	// Name returns the name of the strategy
	Name() string
}

// GetCompactionStrategy returns a compaction strategy based on the provided name
func GetCompactionStrategy(name string, model string, logger Logger) (CompactionStrategy, error) {
	switch strings.ToLower(name) {
	case CompactionStrategyDeleteOld:
		return NewDeleteOldStrategy(model, logger), nil
	default:
		return nil, fmt.Errorf("unknown compaction strategy: %s", name)
	}
}

// DeleteOldStrategy implements the strategy to delete oldest messages first
type DeleteOldStrategy struct {
	model  string
	logger Logger
}

// NewDeleteOldStrategy creates a new instance of DeleteOldStrategy
func NewDeleteOldStrategy(model string, logger Logger) *DeleteOldStrategy {
	return &DeleteOldStrategy{
		model:  model,
		logger: logger,
	}
}

func (s *DeleteOldStrategy) Name() string {
	return CompactionStrategyDeleteOld
}

// Compact implements the CompactionStrategy interface by removing the oldest messages.
// It always preserves the system prompt (first message)
func (s *DeleteOldStrategy) Compact(messages []llms.MessageContent, currentTokens, maxTokens int) ([]llms.MessageContent, int) {
	if len(messages) <= 1 || currentTokens <= maxTokens {
		return messages, currentTokens
	}

	// Always preserve the system prompt (first message)
	systemPrompt := messages[0]
	systemPromptTokens := estimateTokenCount(systemPrompt, s.model)

	// Start with just the system prompt
	result := []llms.MessageContent{systemPrompt}
	resultTokens := systemPromptTokens

	// Add as many recent messages as possible without exceeding the token limit
	for i := len(messages) - 1; i > 0; i-- {
		msgTokens := estimateTokenCount(messages[i], s.model)

		if resultTokens+msgTokens <= maxTokens {
			// Add this message at the beginning (after system prompt)
			// This maintains the correct chronological order
			result = append([]llms.MessageContent{messages[i]}, result...)
			resultTokens += msgTokens
		} else {
			// Skip this message as it would exceed the token limit
			s.logger.Debugf("Skipping message index %d during compaction to stay within token limit", i)
		}
	}

	s.logger.Infof("Compacted chat history from %d to %d messages, tokens reduced from %d to %d",
		len(messages), len(result), currentTokens, resultTokens)

	return result, resultTokens
}
