package chat

import (
	"fmt"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/tmc/langchaingo/llms"
)

// DefaultTokenCountModel model to use for token counting if none is specified
const DefaultTokenCountModel = "gpt-3.5-turbo"

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
	model        string
	logger       Logger
	tokenCounter llm_models.TokenEstimator
}

// NewDeleteOldStrategy creates a new instance of DeleteOldStrategy
func NewDeleteOldStrategy(model string, logger Logger) *DeleteOldStrategy {
	return &DeleteOldStrategy{
		model:        model,
		logger:       logger,
		tokenCounter: llm_models.TokenEstimator{},
	}
}

func (s *DeleteOldStrategy) Name() string {
	return CompactionStrategyDeleteOld
}

// Compact implements the CompactionStrategy interface by removing the oldest messages.
// It always preserves the system prompt (first message) if present.
func (s *DeleteOldStrategy) Compact(messages []llms.MessageContent, currentTokens, maxTokens int) ([]llms.MessageContent, int) {
	if len(messages) == 0 || currentTokens <= maxTokens {
		return messages, currentTokens
	}

	var systemPrompt *llms.MessageContent
	startIdx := 0
	if len(messages) > 0 && messages[0].Role == llms.ChatMessageTypeSystem {
		systemPrompt = &messages[0]
		startIdx = 1
	}

	resultTokens := 0
	if systemPrompt != nil {
		resultTokens = s.tokenCounter.CountTokens(*systemPrompt)
	}

	// Collect as many recent messages as possible (excluding system prompt)
	collected := make([]llms.MessageContent, 0, len(messages)-startIdx)
	for i := len(messages) - 1; i >= startIdx; i-- {
		msgTokens := s.tokenCounter.CountTokens(messages[i])
		if resultTokens+msgTokens <= maxTokens {
			collected = append(collected, messages[i])
			resultTokens += msgTokens
		} else {
			s.logger.Debugf("Skipping message index %d during compaction to stay within token limit", i)
		}
	}

	// Reverse collected messages to restore chronological order
	for i, j := 0, len(collected)-1; i < j; i, j = i+1, j-1 {
		collected[i], collected[j] = collected[j], collected[i]
	}

	// Build final result
	result := make([]llms.MessageContent, 0, 1+len(collected))
	if systemPrompt != nil {
		result = append(result, *systemPrompt)
	}
	result = append(result, collected...)

	s.logger.Infof("Compacted chat history from %d to %d messages, tokens reduced from %d to %d",
		len(messages), len(result), currentTokens, resultTokens)

	return result, resultTokens
}
