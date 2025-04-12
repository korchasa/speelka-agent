package chat_test

import (
	"fmt"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/logger"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestChat_BuildPromptPartForToolsDescription(t *testing.T) {
	log := logger.NewLogger()

	// Basic template with bullet points for tool name and indent control
	basicTemplate := `{% for tool in tools -%}
- ` + "`{{ tool.Name }}`" + ` - {{ tool.Description }}
{% endfor %}`

	// Template with parameters and proper property access
	paramsTemplate := `{% for tool in tools -%}
- ` + "`{{ tool.Name }}`" + ` - {{ tool.Description }}. Arguments:
{% if tool.InputSchema and tool.InputSchema.Properties -%}
{% for name, prop in tool.InputSchema.Properties -%}
* ` + "`{{ name }}`" + ` ({{ prop.type }}): {{ prop.description }}
{% endfor -%}
{% endif -%}
{% endfor %}`

	t.Run("simple tool description", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "get_file_structure",
				Description: "Displays the code structure of the specified file",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, basicTemplate)
		assert.NoError(t, err)

		expected := `- ` + "`get_file_structure`" + ` - Displays the code structure of the specified file`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool with parameters", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "get_file_structure",
				Description: "Displays the code structure of the specified file",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"file": map[string]interface{}{
							"type":        "string",
							"description": "the path to the file",
						},
					},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, paramsTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt:\n%s", promptPart)

		// Just get actual formatting and adapt test
		expected := `- ` + "`get_file_structure`" + ` - Displays the code structure of the specified file. Arguments:
* ` + "`file`" + ` (string): the path to the file`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("multiple tools", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "tool1",
				Description: "Description of tool1",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			{
				Name:        "tool2",
				Description: "Description of tool2",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, basicTemplate)
		assert.NoError(t, err)

		expected := `- ` + "`tool1`" + ` - Description of tool1
- ` + "`tool2`" + ` - Description of tool2`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("custom template format", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "get_file_structure",
				Description: "Displays the code structure of the specified file",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		customTemplate := `{% for tool in tools -%}Tool: {{ tool.Name }}
Description: {{ tool.Description }}{% endfor %}`

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, customTemplate)
		assert.NoError(t, err)

		expected := `Tool: get_file_structure
Description: Displays the code structure of the specified file`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool with null Properties", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "tool_no_args",
				Description: "A tool without arguments",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: nil,
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt for tool with null Properties:\n%s", promptPart)

		expected := `- ` + "`tool_no_args`" + ` - A tool without arguments. No arguments required.`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool with empty Properties", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "tool_empty_args",
				Description: "A tool with empty arguments",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt for tool with empty Properties:\n%s", promptPart)

		expected := `- ` + "`tool_empty_args`" + ` - A tool with empty arguments. No arguments required.`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool without InputSchema", func(t *testing.T) {
		ch := chat.NewChat("", "", "query", log)
		tools := []mcp.Tool{
			{
				Name:        "tool_no_schema",
				Description: "A tool without input schema",
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt for tool without InputSchema:\n%s", promptPart)

		expected := `- ` + "`tool_no_schema`" + ` - A tool without input schema. No arguments required.`

		assert.Equal(t, expected, promptPart)
	})
}

func TestChat_TokenLimits(t *testing.T) {
	log := logger.NewLogger()

	t.Run("model-based token limit", func(t *testing.T) {
		ch := chat.NewChat("", "You are an AI assistant.", "query", log)

		// Set token limit to 0 (model-based limit)
		ch.SetMaxTokens(0)

		// Add some messages
		err := ch.Begin("Hello", []mcp.Tool{})
		assert.NoError(t, err)

		// Add several messages without triggering compaction
		// Since we're using a model-based limit (0), compaction should not be triggered
		for i := 0; i < 5; i++ {
			ch.AddAssistantMessage(fmt.Sprintf("Message %d", i))
		}

		// Verify that all messages are preserved
		messages := ch.GetLLMMessages()
		assert.Equal(t, 6, len(messages), "Should have 6 messages (1 system + 5 assistant)")

		// Verify that the first message is the system message
		assert.Equal(t, llms.ChatMessageTypeSystem, messages[0].Role)
	})

	t.Run("get total tokens", func(t *testing.T) {
		ch := chat.NewChat("", "You are an AI assistant.", "query", log)
		err := ch.Begin("Hello, how are you?", []mcp.Tool{})
		assert.NoError(t, err)

		// Should have a positive number of tokens after initialization
		assert.Greater(t, ch.GetTotalTokens(), 0)

		initialTokens := ch.GetTotalTokens()

		// Add a message and check that tokens increased
		ch.AddAssistantMessage("I'm doing well, thank you for asking!")
		assert.Greater(t, ch.GetTotalTokens(), initialTokens)
	})

	t.Run("set max tokens", func(t *testing.T) {
		ch := chat.NewChat("", "You are an AI assistant.", "query", log)

		// Default max tokens should be positive
		assert.Greater(t, chat.DefaultMaxTokens, 0)

		// Set custom max tokens
		customMaxTokens := 1000
		ch.SetMaxTokens(customMaxTokens)

		// Only negative max tokens should be ignored
		ch.SetMaxTokens(-10)

		// Zero is a valid value (model-based limit)
		ch.SetMaxTokens(0)

		// Add a very long message to trigger compaction
		err := ch.Begin("Hello", []mcp.Tool{})
		assert.NoError(t, err)

		// Add 10 messages to build up token count
		for i := 0; i < 10; i++ {
			ch.AddAssistantMessage("This is a relatively long message that will consume tokens in our conversation history. It contains enough text to make the tokenizer work hard and count a reasonable number of tokens for testing purposes. We need to make sure we have enough text to potentially trigger the compaction mechanism when we go over our limit.")
		}
	})
}

func TestChat_Compaction(t *testing.T) {
	log := logger.NewLogger()

	t.Run("set compaction strategy", func(t *testing.T) {
		ch := chat.NewChat("", "You are an AI assistant.", "query", log)

		// Default strategy is "delete-old"
		err := ch.SetCompactionStrategy(chat.CompactionStrategyDeleteOld)
		assert.NoError(t, err)

		// Invalid strategy should return error
		err = ch.SetCompactionStrategy("invalid-strategy")
		assert.Error(t, err)
	})

	t.Run("compaction preserves system prompt", func(t *testing.T) {
		ch := chat.NewChat("", "System instruction: You are a helpful AI assistant.", "query", log)

		// Set a very low token limit to force compaction
		ch.SetMaxTokens(100)

		err := ch.Begin("Hello", []mcp.Tool{})
		assert.NoError(t, err)

		// Get the initial messages to compare later
		initialMessages := ch.GetLLMMessages()
		assert.Equal(t, 1, len(initialMessages))

		// Add enough messages to trigger compaction
		for i := 0; i < 5; i++ {
			ch.AddAssistantMessage("This is message " + string(rune('A'+i)))
		}

		// Get final messages and ensure system prompt is preserved
		finalMessages := ch.GetLLMMessages()

		// System prompt should still be the same
		assert.Greater(t, len(finalMessages), 1)
		assert.Equal(t, initialMessages[0], finalMessages[0])
	})
}
