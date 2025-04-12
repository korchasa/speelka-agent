package chat

import (
	"fmt"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/types"
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

	// DefaultMaxTokens Default max tokens if not specified
	DefaultMaxTokens = 8192
)

// Chat manages the conversation history and message formatting
// Responsibility: Maintaining the chat history and formatting messages for LLM
// Features: Stores message history, adds tool calls and responses to it
type Chat struct {
	promptTemplate string
	argName        string
	messages       []llms.MessageContent
	logger         logger.Spec

	// Token tracking
	modelName   string
	totalTokens int
	maxTokens   int

	// Compaction strategy
	compactionStrategy CompactionStrategy
}

// NewChat creates a new Chat with the given prompt template and argument name
func NewChat(model string, promptTemplate, argName string, logger logger.Spec) *Chat {
	return &Chat{
		promptTemplate:     promptTemplate,
		argName:            argName,
		messages:           make([]llms.MessageContent, 0),
		logger:             logger,
		modelName:          model,
		totalTokens:        0,
		maxTokens:          DefaultMaxTokens,
		compactionStrategy: NewDeleteOldStrategy(model, logger),
	}
}

// SetMaxTokens sets the maximum number of tokens allowed in the chat history
// A value of 0 means to use a model-based limit (effectively no limit in the chat component)
func (c *Chat) SetMaxTokens(maxTokens int) {
	if maxTokens < 0 {
		c.logger.Warnf("Invalid max tokens value %d, using default %d", maxTokens, DefaultMaxTokens)
		maxTokens = DefaultMaxTokens
	}
	c.maxTokens = maxTokens
}

// SetCompactionStrategy sets the strategy used for compacting chat history
func (c *Chat) SetCompactionStrategy(strategy string) error {
	switch strategy {
	case CompactionStrategyDeleteOld:
		c.compactionStrategy = NewDeleteOldStrategy(c.modelName, c.logger)
	default:
		return fmt.Errorf("unsupported compaction strategy: %s", strategy)
	}
	return nil
}

// GetTotalTokens returns the current total token count of the chat history
func (c *Chat) GetTotalTokens() int {
	return c.totalTokens
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

	systemMessage := llms.TextParts(llms.ChatMessageTypeSystem, result)
	c.messages = append(c.messages, systemMessage)

	// Count tokens for the system message
	messageTokens := estimateTokenCount(systemMessage, c.modelName)
	c.totalTokens += messageTokens
	c.logger.Debugf("Added system message with %d tokens, total now %d", messageTokens, c.totalTokens)

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

	message := llms.TextParts(llms.ChatMessageTypeAI, content)
	messageTokens := estimateTokenCount(message, c.modelName)

	// Check if adding this message would exceed the token limit
	// maxTokens of 0 means no limit (model-based limit)
	if c.maxTokens > 0 && c.totalTokens+messageTokens > c.maxTokens {
		c.logger.Infof("Token limit of %d would be exceeded by adding message with %d tokens. Compacting history...",
			c.maxTokens, messageTokens)
		c.messages, c.totalTokens = c.compactionStrategy.Compact(c.messages, c.totalTokens, c.maxTokens-messageTokens)
	}

	c.messages = append(c.messages, message)
	c.totalTokens += messageTokens
	c.logger.Debugf("Added assistant message with %d tokens, total now %d", messageTokens, c.totalTokens)
}

// AddToolCall adds a tool call to the chat history.
//
// Parameters:
//   - toolCall: information about the tool call
func (c *Chat) AddToolCall(toolCall types.CallToolRequest) {
	message := llms.MessageContent{
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
	}

	messageTokens := estimateTokenCount(message, c.modelName)

	// Check if adding this message would exceed the token limit
	// maxTokens of 0 means no limit (model-based limit)
	if c.maxTokens > 0 && c.totalTokens+messageTokens > c.maxTokens {
		c.logger.Infof("Token limit of %d would be exceeded by adding tool call with %d tokens. Compacting history...",
			c.maxTokens, messageTokens)
		c.messages, c.totalTokens = c.compactionStrategy.Compact(c.messages, c.totalTokens, c.maxTokens-messageTokens)
	}

	c.messages = append(c.messages, message)
	c.totalTokens += messageTokens
	c.logger.Debugf("Added tool call with %d tokens, total now %d", messageTokens, c.totalTokens)
}

// AddToolResult adds the result of a tool execution to the chat history.
//
// Parameters:
//   - toolCall: information about the tool call
//   - result: result of the tool execution
func (c *Chat) AddToolResult(toolCall types.CallToolRequest, result *mcp.CallToolResult) {
	resultStr := "Result: "
	if result.IsError {
		resultStr += fmt.Sprintf("Error: %s", logger.SDump(map[string]any{"error": result.Content}))
	} else {
		resultStr += fmt.Sprintf("%v", result.Content)
	}

	message := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Name:       toolCall.ToolName(),
				Content:    resultStr,
			},
		},
	}

	messageTokens := estimateTokenCount(message, c.modelName)

	// Check if adding this message would exceed the token limit
	// maxTokens of 0 means no limit (model-based limit)
	if c.maxTokens > 0 && c.totalTokens+messageTokens > c.maxTokens {
		c.logger.Infof("Token limit of %d would be exceeded by adding tool result with %d tokens. Compacting history...",
			c.maxTokens, messageTokens)
		c.messages, c.totalTokens = c.compactionStrategy.Compact(c.messages, c.totalTokens, c.maxTokens-messageTokens)
	}

	c.messages = append(c.messages, message)
	c.totalTokens += messageTokens
	c.logger.Debugf("Added tool result with %d tokens, total now %d", messageTokens, c.totalTokens)
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
