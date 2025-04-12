// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
)

// AgentSpec represents the interface for the Agent component.
// Responsibility: Defining the contract for the Agent component
// Features: Defines methods for starting, stopping, and handling requests
type AgentSpec interface {
	// HandleRequest processes a request to the Agent.
	// It returns the response from the Agent and an error if the request fails.
	HandleRequest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// RegisterTools registers all tools with the MCP server.
	RegisterTools()
}

// AgentConfig represents the configuration for the Agent.
// Responsibility: Storing all settings needed by the Agent
// Features: Includes tool configuration, LLM configuration, and chat configuration
type AgentConfig struct {
	// Tool configuration
	Tool MCPServerToolConfig

	// LLM configuration
	Model                string
	SystemPromptTemplate string

	// Chat configuration
	MaxTokens          int
	CompactionStrategy string
}
