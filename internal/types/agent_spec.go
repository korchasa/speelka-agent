// Package types defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package types

import (
	"context"
)

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
	MaxTokens int

	// Agent behavior configuration
	MaxLLMIterations int
}

// AgentSpec represents the interface for the Agent component.
// Responsibility: Defining the contract for the Agent component
// Features: Defines methods for direct call mode
type AgentSpec interface {
	CallDirect(ctx context.Context, input string) (string, MetaInfo, error)
}
