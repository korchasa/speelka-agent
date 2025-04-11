package chat

import (
    "fmt"
    "strings"

    "github.com/korchasa/speelka-agent-go/internal/logger"
    "github.com/korchasa/speelka-agent-go/internal/types"
    "github.com/korchasa/speelka-agent-go/internal/utils"
    "github.com/mark3labs/mcp-go/mcp"
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

// Chat manages the conversation history and message formatting
// Responsibility: Maintaining the chat history and formatting messages for LLM
// Features: Stores message history, adds tool calls and responses to it
type Chat struct {
    promptTemplate string
    argName        string
    messages       []llms.MessageContent
    logger         logger.Spec
}

// NewChat creates a new Chat with the given prompt template and argument name
func NewChat(promptTemplate, argName string, logger logger.Spec) *Chat {
    return &Chat{
        promptTemplate: promptTemplate,
        argName:        argName,
        messages:       make([]llms.MessageContent, 0),
        logger:         logger,
    }
}

func (c *Chat) Begin(input string, tools []mcp.Tool) error {
    toolsDescription, err := c.BuildPromptPartForToolsDescription(tools, DefaultToolsDescriptionTemplate)
    if err != nil {
        return fmt.Errorf("failed to build tools description: %v", err)
    }
    prompt := prompts.PromptTemplate{
        Template:       c.promptTemplate,
        InputVariables: []string{c.argName, "tools"},
        TemplateFormat: prompts.TemplateFormatJinja2,
    }

    values := map[string]any{
        c.argName: input,
        "tools":   toolsDescription,
    }

    result, err := prompt.Format(values)
    if err != nil {
        return fmt.Errorf("failed to format prompt: %v", err)
    }

    c.messages = append(c.messages, llms.TextParts(llms.ChatMessageTypeSystem, result))
    return nil
}

func (c *Chat) GetLLMMessages() []llms.MessageContent {
    return c.messages
}

// AddAssistantMessage adds a message from the assistant (LLM) to the chat history.
//
// Parameters:
//   - content: text of the message from the assistant
func (c *Chat) AddAssistantMessage(content string) {
    if content == "" {
        return
    }
    c.messages = append(c.messages, llms.TextParts(llms.ChatMessageTypeAI, content))
}

// AddToolCall adds a tool call to the chat history.
//
// Parameters:
//   - toolCall: information about the tool call
func (c *Chat) AddToolCall(toolCall types.CallToolRequest) {
    c.messages = append(c.messages, llms.MessageContent{
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

    c.messages = append(c.messages, llms.MessageContent{
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
