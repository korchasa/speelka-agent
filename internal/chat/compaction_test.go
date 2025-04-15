package chat_test

import (
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestChat_Compaction(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("", log)
	maxTokens := 100 // Use a low value to force compaction

	t.Run("compaction preserves system prompt", func(t *testing.T) {
		ch := chat.NewChat("", "System instruction: You are a helpful AI assistant.", "query", log, calculator, compaction, maxTokens)

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
