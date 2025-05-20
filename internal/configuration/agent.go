// Package configuration defines interfaces for MCP server components.
// Responsibility: Defining interaction contracts between system components
// Features: Contains only interfaces and data structures, without implementation
package configuration

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
