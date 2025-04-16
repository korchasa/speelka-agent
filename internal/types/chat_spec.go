// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
)

// ChatInfo provides a summary of the chat state for reporting and analytics.
type ChatInfo struct {
	TotalTokens     int
	TotalCost       float64
	IsApproximate   bool
	MaxTokens       int
	MessageStackLen int
	LLMRequests     int     // Number of LLM responses (assistant messages)
	ModelName       string  // Name of the LLM model used for this chat
	ToolCallCount   int     // Number of tool calls in the session
	RequestBudget   float64 // Configured cost budget for this chat (USD or token-equivalent)
}

// ChatSpec represents the interface for the Chat component.
// Responsibility: Defining the contract for the Chat component
// Features: Defines methods for managing chat history and interactions
type ChatSpec interface {
	// GetInfo returns a summary of the chat state (tokens, cost, etc.).
	GetInfo() ChatInfo

	// Begin starts a new chat with the given input and available tools.
	// It returns an error if the chat initialization fails.
	Begin(input string, tools []mcp.Tool) error

	// GetLLMMessages returns the messages in the chat history in a format suitable for LLM requests.
	GetLLMMessages() []llms.MessageContent

	// AddAssistantMessage adds a message from the assistant to the chat history.
	AddAssistantMessage(response LLMResponse)

	// AddToolCall adds a tool call to the chat history.
	AddToolCall(toolCall CallToolRequest)

	// AddToolResult adds a tool result to the chat history.
	AddToolResult(toolCall CallToolRequest, result *mcp.CallToolResult)

	// BuildPromptPartForToolsDescription builds a prompt part for describing available tools.
	// It returns the prompt part and an error if the building fails.
	BuildPromptPartForToolsDescription(tools []mcp.Tool, template string) (string, error)
}
