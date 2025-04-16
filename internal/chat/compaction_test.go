package chat_test

import (
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestChat_Compaction(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("", log)
	maxTokens := 100 // Use a low value to force compaction

	t.Run("compaction preserves system prompt", func(t *testing.T) {
		ch := chat.NewChat("", "System instruction: You are a helpful AI assistant.", "query", log, calculator, compaction, maxTokens, 0.0)

		err := ch.Begin("Hello", []mcp.Tool{})
		assert.NoError(t, err)

		// Get the initial messages to compare later
		initialMessages := ch.GetLLMMessages()
		assert.Equal(t, 1, len(initialMessages))

		// Add enough messages to trigger compaction
		for i := 0; i < 5; i++ {
			ch.AddAssistantMessage(types.LLMResponse{Text: "This is message " + string(rune('A'+i)), Metadata: types.LLMResponseMetadata{Tokens: types.LLMResponseTokensMetadata{TotalTokens: 42}}})
		}

		// Get final messages and ensure system prompt is preserved
		finalMessages := ch.GetLLMMessages()

		// System prompt should still be the same as the first message
		assert.Greater(t, len(finalMessages), 1)
		assert.Equal(t, initialMessages[0], finalMessages[0], "System prompt should be preserved as the first message after compaction")
	})
}

func TestCompaction_RemovesToolCallAndResult(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("", log)
	maxTokens := 50 // Low value to force compaction

	ch := chat.NewChat("", "System: {{query}}", "query", log, calculator, compaction, maxTokens, 0.0)
	_ = ch.Begin("Test tool call compaction", nil)

	// Properly construct a tool_call using llms.ToolCall and types.NewCallToolRequest
	llmToolCall := llms.ToolCall{
		ID:   "tool-123",
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      "test_tool",
			Arguments: "{}",
		},
	}
	callReq, err := types.NewCallToolRequest(llmToolCall)
	assert.NoError(t, err)
	ch.AddToolCall(callReq)

	// Add a tool_call result
	result := &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent("Result content")},
		IsError: false,
	}
	ch.AddToolResult(callReq, result)

	// Add enough assistant messages to force compaction
	for i := 0; i < 3; i++ {
		ch.AddAssistantMessage(types.LLMResponse{Text: "Msg", Metadata: types.LLMResponseMetadata{Tokens: types.LLMResponseTokensMetadata{TotalTokens: 30}}})
	}

	msgs := ch.GetLLMMessages()
	for _, msg := range msgs {
		for _, part := range msg.Parts {
			if toolCallPart, ok := part.(llms.ToolCall); ok {
				assert.NotEqual(t, "tool-123", toolCallPart.ID, "tool_call should be deleted by compaction")
			}
			if toolResp, ok := part.(llms.ToolCallResponse); ok {
				assert.NotEqual(t, "tool-123", toolResp.ToolCallID, "tool_call result should be deleted by compaction")
			}
		}
	}
}
