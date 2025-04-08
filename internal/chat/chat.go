package chat

import (
	"fmt"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/korchasa/speelka-agent-go/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	log "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

const (
	DefaultToolsDescriptionTemplate = `{% for tool in tools %}
- ` + "`{{ tool.Name }}`" + ` - {{ tool.Description }}
{%- if tool.InputSchema and tool.InputSchema.Properties and tool.InputSchema.Properties|length > 0 -%}
. Arguments:
{% for name, prop in tool.InputSchema.Properties %}
  * ` + "`{{ name }}`" + ` ({{ prop.type }}): {{ prop.description }}
{%- endfor %}
{%- else -%}
. No arguments required.
{%- endif -%}
{% endfor %}`
)

type Chat struct {
	systemPromptTemplate string
	history              []llms.MessageContent
	logger               *log.Logger
}

func NewChat(systemPromptTemplate string, logger *log.Logger) *Chat {
	return &Chat{
		systemPromptTemplate: systemPromptTemplate,
		history:              make([]llms.MessageContent, 0),
		logger:               logger,
	}
}

func (c *Chat) Begin(input string, tools []mcp.Tool) error {
	toolsDescription, err := c.BuildPromptPartForToolsDescription(tools, DefaultToolsDescriptionTemplate)
	if err != nil {
		return fmt.Errorf("failed to build tools description: %v", err)
	}
	prompt := prompts.PromptTemplate{
		Template:       c.systemPromptTemplate,
		InputVariables: []string{"input", "tools"},
		TemplateFormat: prompts.TemplateFormatJinja2,
	}
	result, err := prompt.Format(map[string]any{
		"input": input,
		"tools": toolsDescription,
	})
	if err != nil {
		return fmt.Errorf("failed to format prompt: %v", err)
	}

	c.history = append(c.history, llms.TextParts(llms.ChatMessageTypeSystem, result))
	return nil
}

func (c *Chat) GetLLMMessages() []llms.MessageContent {
	return c.history
}

// AddAssistantMessage adds a message from the assistant (LLM) to the chat history.
//
// Parameters:
//   - content: text of the message from the assistant
func (c *Chat) AddAssistantMessage(content string) {
	if content == "" {
		return
	}
	c.history = append(c.history, llms.TextParts(llms.ChatMessageTypeAI, content))
}

// AddToolCall adds a tool call to the chat history.
//
// Parameters:
//   - toolCall: information about the tool call
func (c *Chat) AddToolCall(toolCall types.CallToolRequest) {
	c.history = append(c.history, llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   toolCall.ID,
				Type: toolCall.ToLLM().Type,
				FunctionCall: &llms.FunctionCall{
					Name:      toolCall.ToLLM().FunctionCall.Name,
					Arguments: toolCall.ToLLM().FunctionCall.Arguments,
				},
			},
		},
	})
}

// AddToolResult adds the result of a tool execution to the chat history.
//
// Parameters:
//   - toolCall: information about the tool call
//   - result: result of the tool execution
func (c *Chat) AddToolResult(toolCall types.CallToolRequest, result *mcp.CallToolResult) {
	resultStr := "Result: "
	if result.IsError {
		resultStr += fmt.Sprintf("Error: %s", utils.SDump(map[string]any{"error": result.Content}))
	} else {
		resultStr += fmt.Sprintf("%v", result.Content)
	}

	c.history = append(c.history, llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Name:       toolCall.ToolName(),
				Content:    resultStr,
			},
		},
	})
}

// BuildPromptPartForToolsDescription generates a formatted description of available tools
// for inclusion in the system prompt.
//
// Parameters:
//   - tools: slice of mcp.Tool to describe
//   - template: template string for formatting tools
//
// Returns:
//   - formatted string containing descriptions of all tools
//   - error if formatting fails
func (c *Chat) BuildPromptPartForToolsDescription(tools []mcp.Tool, template string) (string, error) {
	prompt := prompts.PromptTemplate{
		Template:       template,
		InputVariables: []string{"tools"},
		TemplateFormat: prompts.TemplateFormatJinja2,
	}
	result, err := prompt.Format(map[string]any{
		"tools": tools,
	})
	if err != nil {
		return "", fmt.Errorf("failed to format tools description: %v", err)
	}
	return strings.Trim(result, " \n"), nil
}
