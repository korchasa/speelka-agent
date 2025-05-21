// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

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
