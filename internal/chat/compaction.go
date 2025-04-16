package chat

import (
	"fmt"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/utils"
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
	// Track indices of messages to delete (tool_call and their results)
	toDelete := make(map[int]struct{})
	// First pass: mark tool_call deletions and their results
	for i := len(messages) - 1; i >= startIdx; i-- {
		msg := messages[i]
		msgTokens := s.tokenCounter.CountTokens(msg)
		if resultTokens+msgTokens <= maxTokens {
			collected = append(collected, msg)
			resultTokens += msgTokens
		} else {
			// Check if this message is a tool_call
			for _, part := range msg.Parts {
				if toolCall, ok := part.(llms.ToolCall); ok {
					// Mark this tool_call for deletion
					toDelete[i] = struct{}{}
					// Search for the corresponding tool_call result (ToolCallResponse)
					for j := startIdx; j < len(messages); j++ {
						if j == i {
							continue
						}
						for _, part2 := range messages[j].Parts {
							if toolResp, ok := part2.(llms.ToolCallResponse); ok && toolResp.ToolCallID == toolCall.ID {
								toDelete[j] = struct{}{}
								s.logger.Infof("Deleted tool_call result at index %d for tool_call ID %s: %s", j, toolCall.ID, utils.SDump(messages[j]))
							}
						}
					}
					s.logger.Infof("Deleted tool_call at index %d: %s", i, utils.SDump(msg))
				}
			}
			// If not a tool_call, just mark for deletion
			if _, found := toDelete[i]; !found {
				s.logger.Infof("Deleted message at index %d: %s", i, utils.SDump(msg))
				toDelete[i] = struct{}{}
			}
		}
	}
	// Build the result, skipping deleted indices
	for idx := startIdx; idx < len(messages); idx++ {
		if _, skip := toDelete[idx]; !skip {
			collected = append(collected, messages[idx])
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

	// Always clean up orphaned tool_calls before compaction
	cleaned, _ := ValidateAndCleanupOrphanedToolCalls(result, s.logger)

	return cleaned, resultTokens
}

// ValidateAndCleanupOrphanedToolCalls checks that every tool_call has a matching tool_response.
// If an orphaned tool_call is found, it is removed from the message stack and a warning is logged.
// Returns the cleaned message stack and a boolean indicating if any orphans were found.
func ValidateAndCleanupOrphanedToolCalls(messages []llms.MessageContent, logger Logger) ([]llms.MessageContent, bool) {
	toolCallIDs := map[string]struct{}{}
	toolResponseIDs := map[string]struct{}{}
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if tc, ok := part.(llms.ToolCall); ok {
				toolCallIDs[tc.ID] = struct{}{}
			}
			if tr, ok := part.(llms.ToolCallResponse); ok {
				toolResponseIDs[tr.ToolCallID] = struct{}{}
			}
		}
	}
	orphaned := make(map[string]struct{})
	for id := range toolCallIDs {
		if _, ok := toolResponseIDs[id]; !ok {
			logger.Warnf("Orphaned tool_call: %s. Auto-removing from message stack.", id)
			orphaned[id] = struct{}{}
		}
	}
	if len(orphaned) > 0 {
		var cleaned []llms.MessageContent
		for _, msg := range messages {
			keep := true
			for _, part := range msg.Parts {
				if tc, ok := part.(llms.ToolCall); ok {
					if _, isOrphan := orphaned[tc.ID]; isOrphan {
						keep = false
						break
					}
				}
			}
			if keep {
				cleaned = append(cleaned, msg)
			}
		}
		logger.Warnf("Removed %d orphaned tool_call(s) from message stack.", len(orphaned))
		return cleaned, true
	}
	return messages, false
}
