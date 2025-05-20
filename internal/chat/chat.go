package chat

import (
	"fmt"
	"github.com/korchasa/speelka-agent-go/internal/llm/cost"
	types2 "github.com/korchasa/speelka-agent-go/internal/llm/types"
	"github.com/korchasa/speelka-agent-go/internal/utils/dump"
	"strings"

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
	argumentName   string
	messagesStack  []llms.MessageContent
	logger         loggerSpec

	// Unified chat info struct
	info types.ChatInfo

	// Store LLMResponse objects for assistant messages
	llmMessagesHistory []types2.LLMResponse

	// Cost calculator (should be set from llm_models)
	calculator calculatorSpec

	// Request budget (USD or token-equivalent)
	requestBudget float64
}

type calculatorSpec interface {
	// CalculateLLMResponse returns the number of tokens, USD cost, and approximation flag for the given model and LLM response.
	CalculateLLMResponse(modelName string, resp types2.LLMResponse) (tokens int, cost float64, isApprox bool, err error)
}

type loggerSpec interface {
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

// NewChat creates a new Chat with the given prompt template, calculator, max tokens, and request budget
func NewChat(model string, promptTemplate string, argumentName string, logger loggerSpec, calculator calculatorSpec, maxTokens int, requestBudget float64) *Chat {
	if calculator == nil {
		calculator = cost.NewCalculator()
	}
	if maxTokens < 0 {
		logger.Warnf("Invalid max tokens value %d, using default %d", maxTokens, DefaultMaxTokens)
		maxTokens = DefaultMaxTokens
	}
	return &Chat{
		promptTemplate: promptTemplate,
		argumentName:   argumentName,
		messagesStack:  make([]llms.MessageContent, 0),
		logger:         logger,
		info: types.ChatInfo{
			ModelName:     model,
			MaxTokens:     maxTokens,
			RequestBudget: requestBudget,
		},
		calculator:    calculator,
		requestBudget: requestBudget,
	}
}

// GetInfo returns a summary of the chat state (tokens, cost, etc.).
func (c *Chat) GetInfo() types.ChatInfo {
	return c.info
}

func (c *Chat) Begin(input string, tools []mcp.Tool) error {
	toolsDescription, err := c.BuildPromptPartForToolsDescription(tools, DefaultToolsDescriptionTemplate)
	if err != nil {
		return fmt.Errorf("failed to build tools description: %v", err)
	}
	prompt := prompts.PromptTemplate{
		Template:       c.promptTemplate,
		InputVariables: []string{c.argumentName, "input", "tools"},
		TemplateFormat: prompts.TemplateFormatJinja2,
	}
	values := map[string]any{
		c.argumentName: input,
		"input":        input,
		"tools":        toolsDescription,
	}
	result, err := prompt.Format(values)
	if err != nil {
		return fmt.Errorf("failed to format prompt: %v", err)
	}
	systemMessage := llms.TextParts(llms.ChatMessageTypeSystem, result)
	c.messagesStack = append(c.messagesStack, systemMessage)
	tokenEstimator := cost.TokenEstimator{}
	// Count tokens for the system message
	messageTokens := tokenEstimator.CountTokens(systemMessage)
	c.info.TotalTokens += messageTokens
	c.info.MessageStackLen = len(c.messagesStack)
	c.logger.Debugf("Added system message with %d tokens, total now %d", messageTokens, c.info.TotalTokens)
	return nil
}

func (c *Chat) GetLLMMessages() []llms.MessageContent {
	return c.messagesStack
}

// AddAssistantMessage adds a message from the assistant (LLM) to the chat history.
func (c *Chat) AddAssistantMessage(response types2.LLMResponse) {
	message := llms.TextParts(llms.ChatMessageTypeAI, response.Text)

	tokens := response.Metadata.Tokens.TotalTokens
	cost := response.Metadata.Cost
	isApprox := false
	if tokens == 0 {
		// Fallback to calculator if no token info
		tokens, cost, isApprox, _ = c.calculator.CalculateLLMResponse(c.info.ModelName, response)
	}

	c.messagesStack = append(c.messagesStack, message)
	c.llmMessagesHistory = append(c.llmMessagesHistory, response)

	// Only increment, never decrease
	c.info.TotalTokens += tokens
	c.info.TotalCost += cost
	if isApprox {
		c.info.IsApproximate = true
	}
	c.info.LLMRequests = len(c.llmMessagesHistory)
	c.info.MessageStackLen = len(c.messagesStack)

	c.logger.Debugf("Added assistant message, total tokens: %d, cost: %f, approx: %v", c.info.TotalTokens, c.info.TotalCost, c.info.IsApproximate)
}

// AddToolCall adds a tool call to the chat history.
func (c *Chat) AddToolCall(toolCall types.CallToolRequest) {
	llmCall := toolCall.ToLLM()
	if llmCall.FunctionCall == nil || llmCall.ID == "" {
		c.logger.Warnf("AddToolCall: toolCall has nil FunctionCall or empty ID: %+v", toolCall)
		return
	}
	message := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   llmCall.ID,
				Type: llmCall.Type,
				FunctionCall: &llms.FunctionCall{
					Name:      llmCall.FunctionCall.Name,
					Arguments: llmCall.FunctionCall.Arguments,
				},
			},
		},
	}

	tokenEstimator := cost.TokenEstimator{}
	messageTokens := tokenEstimator.CountTokens(message)

	c.messagesStack = append(c.messagesStack, message)
	c.info.TotalTokens += messageTokens
	c.info.MessageStackLen = len(c.messagesStack)
	c.info.ToolCallCount++

	c.logger.Debugf("Added tool call with %d tokens, total now %d", messageTokens, c.info.TotalTokens)
}

// AddToolResult adds the result of a tool execution to the chat history.
func (c *Chat) AddToolResult(toolCall types.CallToolRequest, result *mcp.CallToolResult) {
	resultStr := "Result: "
	if result.IsError {
		resultStr += fmt.Sprintf("Error: %s", dump.SDump(map[string]any{"error": result.Content}))
	} else {
		resultStr += fmt.Sprintf("%v", result.Content)
	}
	c.info.ToolCallCount++

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

	tokenEstimator := cost.TokenEstimator{}
	messageTokens := tokenEstimator.CountTokens(message)

	c.messagesStack = append(c.messagesStack, message)
	c.info.TotalTokens += messageTokens
	c.info.MessageStackLen = len(c.messagesStack)

	c.logger.Debugf("Added tool result with %d tokens, total now %d", messageTokens, c.info.TotalTokens)
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

// ExceededRequestBudget returns true if the total cost exceeds the configured request budget (if > 0)
func (c *Chat) ExceededRequestBudget() bool {
	if c.requestBudget > 0 && c.info.TotalCost > c.requestBudget {
		c.logger.Warnf("Request budget exceeded: total cost %.4f > budget %.4f", c.info.TotalCost, c.requestBudget)
		return true
	}
	return false
}
