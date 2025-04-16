package chat_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestChat_InitializationAndGetInfo(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	maxTokens := 2048
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, maxTokens, 0.0)

	info := ch.GetInfo()
	assert.Equal(t, "gpt-4o", info.ModelName)
	assert.Equal(t, maxTokens, info.MaxTokens)
	assert.Zero(t, info.TotalTokens)
	assert.Zero(t, info.TotalCost)
	assert.Zero(t, info.LLMRequests)
	assert.Zero(t, info.ToolCallCount)
	assert.Zero(t, info.MessageStackLen)
}

func TestChat_Begin_SystemPromptAndToolDescription(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}. Tools: {{tools}}", "query", log, calculator, compaction, 2048, 0.0)

	tools := []mcp.Tool{
		mcp.NewTool("echo", mcp.WithString("msg", mcp.Required(), mcp.Description("Message to echo"))),
	}
	err := ch.Begin("Hello", tools)
	assert.NoError(t, err)

	msgs := ch.GetLLMMessages()
	assert.Len(t, msgs, 1)
	assert.Equal(t, llms.ChatMessageTypeSystem, msgs[0].Role)
	if len(msgs[0].Parts) > 0 {
		if text, ok := msgs[0].Parts[0].(llms.TextContent); ok {
			assert.Contains(t, text.Text, "Hello")
			assert.Contains(t, text.Text, "echo")
		} else {
			t.Errorf("Expected TextContent in system message part")
		}
	} else {
		t.Errorf("No parts in system message")
	}

	info := ch.GetInfo()
	assert.Equal(t, 1, info.MessageStackLen)
	assert.Greater(t, info.TotalTokens, 0)
}

func TestChat_AddAssistantMessage_TokenCostApproximation(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)
	_ = ch.Begin("Hi", nil)

	resp := types.LLMResponse{
		Text: "This is a test response.",
		Metadata: types.LLMResponseMetadata{
			Tokens: types.LLMResponseTokensMetadata{
				TotalTokens:      10,
				PromptTokens:     5,
				CompletionTokens: 5,
			},
			Cost: 0.001,
		},
	}
	ch.AddAssistantMessage(resp)

	info := ch.GetInfo()
	assert.Equal(t, 19, info.TotalTokens) // 9 (system) + 10 (assistant)
	assert.InDelta(t, 0.001, info.TotalCost, 1e-8)
	assert.False(t, info.IsApproximate)
	assert.Equal(t, 2, info.MessageStackLen)
}

func TestChat_AddAssistantMessage_FallbackEstimation(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)
	_ = ch.Begin("Hi", nil)

	resp := types.LLMResponse{
		Text: "Fallback estimation test message.",
		// No token metadata provided
	}
	ch.AddAssistantMessage(resp)

	info := ch.GetInfo()
	assert.Equal(t, 1, info.LLMRequests)
	assert.Greater(t, info.TotalTokens, 0)
	assert.True(t, info.IsApproximate)
	assert.Greater(t, info.TotalCost, 0.0)
	assert.Equal(t, 2, info.MessageStackLen)
}

func TestChat_AddToolCall_And_AddToolResult(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)
	_ = ch.Begin("Hi", nil)

	// Tool call
	toolCall := llms.ToolCall{
		ID:   "tool-1",
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      "echo",
			Arguments: `{"msg":"hello"}`,
		},
	}
	callReq, err := types.NewCallToolRequest(toolCall)
	assert.NoError(t, err)
	ch.AddToolCall(callReq)

	info := ch.GetInfo()
	assert.Equal(t, 1, info.ToolCallCount)
	assert.Equal(t, 2, info.MessageStackLen)
	assert.Greater(t, info.TotalTokens, 0)

	// Tool result
	result := &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent("Echoed: hello")},
		IsError: false,
	}
	ch.AddToolResult(callReq, result)

	info2 := ch.GetInfo()
	assert.Equal(t, 3, info2.MessageStackLen)
	assert.Greater(t, info2.TotalTokens, 0)
}

func TestChat_AddToolResult_ErrorHandling(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)
	_ = ch.Begin("Hi", nil)

	toolCall := llms.ToolCall{
		ID:   "tool-err",
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      "fail",
			Arguments: `{"msg":"fail"}`,
		},
	}
	callReq, err := types.NewCallToolRequest(toolCall)
	assert.NoError(t, err)
	ch.AddToolCall(callReq)

	errorResult := &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent("Something went wrong!")},
		IsError: true,
	}
	ch.AddToolResult(callReq, errorResult)

	info := ch.GetInfo()
	assert.Equal(t, 3, info.MessageStackLen)

	msgs := ch.GetLLMMessages()
	found := false
	for _, msg := range msgs {
		if msg.Role == llms.ChatMessageTypeTool {
			for _, part := range msg.Parts {
				if toolResp, ok := part.(llms.ToolCallResponse); ok {
					if toolResp.ToolCallID == callReq.ID && toolResp.Name == callReq.ToolName() {
						if toolResp.Content != "" &&
							strings.Contains(toolResp.Content, "Error") &&
							strings.Contains(toolResp.Content, "Something went wrong!") {
							found = true
						}
					}
				}
			}
		}
	}
	assert.True(t, found, "Error tool result should be present in the message stack and contain the error message")
}

func TestChat_BuildPromptPartForToolsDescription(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)

	tools := []mcp.Tool{
		mcp.NewTool("echo", mcp.WithString("msg", mcp.Required(), mcp.Description("Message to echo"))),
	}
	desc, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
	assert.NoError(t, err)
	assert.Contains(t, desc, "echo")
	assert.Contains(t, desc, "msg")
}

func TestChat_GetLLMMessages_StackCorrectness(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)
	_ = ch.Begin("Hi", nil)

	resp := types.LLMResponse{
		Text: "Test response.",
		Metadata: types.LLMResponseMetadata{
			Tokens: types.LLMResponseTokensMetadata{
				TotalTokens: 7,
			},
		},
	}
	ch.AddAssistantMessage(resp)

	msgs := ch.GetLLMMessages()
	assert.Len(t, msgs, 2)
	assert.Equal(t, llms.ChatMessageTypeSystem, msgs[0].Role)
	assert.Equal(t, llms.ChatMessageTypeAI, msgs[1].Role)
}

func TestChat_RequestBudgetEnforcement(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4", log)
	budget := 0.0015 // Budget for two messages (each 0.001)
	ch := chat.NewChat("gpt-4", "System: {{query}}", "query", log, calculator, compaction, 2048, budget)
	_ = ch.Begin("Hi", nil)

	resp := types.LLMResponse{
		Text: "This is a test response.",
		Metadata: types.LLMResponseMetadata{
			Tokens: types.LLMResponseTokensMetadata{
				PromptTokens:     10,
				CompletionTokens: 10,
				TotalTokens:      20,
			},
			Cost: 0.001,
		},
	}
	ch.AddAssistantMessage(resp)
	assert.False(t, ch.ExceededRequestBudget(), "Should not exceed budget after first message")

	// Add another message to exceed the budget
	ch.AddAssistantMessage(resp)
	assert.True(t, ch.ExceededRequestBudget(), "Should exceed budget after second message")
}

func TestChat_OrphanedToolCallDetection(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, 2048, 0.0)
	_ = ch.Begin("Hi", nil)

	// Add a tool call but do NOT add a tool result (simulate error between call and result)
	toolCall := llms.ToolCall{
		ID:   "orphan-tool-1",
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      "echo",
			Arguments: `{"msg":"hello"}`,
		},
	}
	callReq, err := types.NewCallToolRequest(toolCall)
	assert.NoError(t, err)
	ch.AddToolCall(callReq)

	msgs := ch.GetLLMMessages()

	// Use the same validation as the agent
	var validateToolCallResponses = func(messages []llms.MessageContent, logger types.LoggerSpec) error {
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
		for id := range toolCallIDs {
			if _, ok := toolResponseIDs[id]; !ok {
				logger.Warnf("Orphaned tool_call: %s", id)
				return fmt.Errorf("Orphaned tool_call: %s", id)
			}
		}
		return nil
	}

	err = validateToolCallResponses(msgs, log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Orphaned tool_call: orphan-tool-1")
}

func TestChat_TotalTokensAndCost_NeverDecreaseOnCompaction(t *testing.T) {
	log := logger.NewLogger()
	calculator := llm_models.NewCalculator()
	maxTokens := 20 // Low limit to force compaction quickly
	compaction := chat.NewDeleteOldStrategy("gpt-4o", log)
	ch := chat.NewChat("gpt-4o", "System: {{query}}", "query", log, calculator, compaction, maxTokens, 0.0)
	_ = ch.Begin("Hi", nil)

	resp := types.LLMResponse{
		Text: "0123456789", // 10 chars, ~2 tokens
		Metadata: types.LLMResponseMetadata{
			Tokens: types.LLMResponseTokensMetadata{
				PromptTokens:     2,
				CompletionTokens: 2,
				TotalTokens:      4,
			},
			Cost: 0.001,
		},
	}

	totalTokensPrev := ch.GetInfo().TotalTokens
	totalCostPrev := ch.GetInfo().TotalCost

	// Add enough messages to exceed maxTokens and trigger compaction
	for i := 0; i < 10; i++ {
		ch.AddAssistantMessage(resp)
		info := ch.GetInfo()
		// Save previous values for next iteration
		assert.GreaterOrEqual(t, info.TotalTokens, totalTokensPrev, "TotalTokens decreased after compaction!")
		assert.GreaterOrEqual(t, info.TotalCost, totalCostPrev, "TotalCost decreased after compaction!")
		totalTokensPrev = info.TotalTokens
		totalCostPrev = info.TotalCost
	}
}
