package types

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Configuration represents the complete configuration structure that matches the example files
type Configuration struct {
	Runtime RuntimeConfig `json:"runtime" yaml:"runtime"`
	Agent   ConfigAgent   `json:"agent" yaml:"agent"`
}

// RuntimeConfig represents the runtime configuration section
type RuntimeConfig struct {
	Log        RuntimeLogConfig       `json:"log" yaml:"log"`
	Transports RuntimeTransportConfig `json:"transports" yaml:"transports"`
}

// RuntimeLogConfig represents the log configuration section in the runtime config
type RuntimeLogConfig struct {
	RawLevel  string `json:"level" yaml:"level"`
	RawOutput string `json:"output" yaml:"output"`

	// Internal fields, not directly from config file
	LogLevel logrus.Level
	Output   io.Writer
}

// RuntimeTransportConfig represents the transport configuration section in the runtime config
type RuntimeTransportConfig struct {
	Stdio RuntimeStdioConfig `json:"stdio" yaml:"stdio"`
	HTTP  RuntimeHTTPConfig  `json:"http" yaml:"http"`
}

// RuntimeStdioConfig represents the configuration for stdio transport in runtime config
type RuntimeStdioConfig struct {
	Enabled    bool `json:"enabled" yaml:"enabled"`
	BufferSize int  `json:"buffer_size" yaml:"buffer_size"`
}

// RuntimeHTTPConfig represents the configuration for HTTP transport in runtime config
type RuntimeHTTPConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Host    string `json:"host" yaml:"host"`
	Port    int    `json:"port" yaml:"port"`
}

// ConfigAgent represents the agent configuration section
type ConfigAgent struct {
	Name        string                 `json:"name" yaml:"name"`
	Version     string                 `json:"version" yaml:"version"`
	Tool        AgentToolConfig        `json:"tool" yaml:"tool"`
	Chat        AgentChatConfig        `json:"chat" yaml:"chat"`
	LLM         AgentLLMConfig         `json:"llm" yaml:"llm"`
	Connections AgentConnectionsConfig `json:"connections" yaml:"connections"`
}

// AgentChatConfig represents the chat configuration section
type AgentChatConfig struct {
	MaxTokens        int     `json:"max_tokens" yaml:"max_tokens"`
	MaxLLMIterations int     `json:"max_llm_iterations" yaml:"max_llm_iterations"`
	RequestBudget    float64 `json:"request_budget" yaml:"request_budget"`
}

// AgentToolConfig represents the tool configuration section
type AgentToolConfig struct {
	Name                string `json:"name" yaml:"name"`
	Description         string `json:"description" yaml:"description"`
	ArgumentName        string `json:"argument_name" yaml:"argument_name"`
	ArgumentDescription string `json:"argument_description" yaml:"argument_description"`
}

// AgentLLMConfig represents the LLM configuration section
type AgentLLMConfig struct {
	Provider       string         `json:"provider" yaml:"provider"`
	Model          string         `json:"model" yaml:"model"`
	APIKey         string         `json:"api_key" yaml:"api_key"`
	MaxTokens      int            `json:"max_tokens" yaml:"max_tokens"`
	Temperature    float64        `json:"temperature" yaml:"temperature"`
	PromptTemplate string         `json:"prompt_template" yaml:"prompt_template"`
	Retry          LLMRetryConfig `json:"retry" yaml:"retry"`

	// Internal flags
	IsMaxTokensSet   bool
	IsTemperatureSet bool
}

// LLMRetryConfig represents the retry configuration for LLM operations
type LLMRetryConfig struct {
	MaxRetries        int     `json:"max_retries" yaml:"max_retries"`
	InitialBackoff    float64 `json:"initial_backoff" yaml:"initial_backoff"`
	MaxBackoff        float64 `json:"max_backoff" yaml:"max_backoff"`
	BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`
}

// AgentConnectionsConfig represents the connections configuration section
type AgentConnectionsConfig struct {
	McpServers map[string]MCPServerConnection `json:"mcpServers" yaml:"mcpServers"`
	Retry      ConnectionRetryConfig          `json:"retry" yaml:"retry"`
}

// ConnectionRetryConfig represents the retry configuration for connections
type ConnectionRetryConfig struct {
	MaxRetries        int     `json:"max_retries" yaml:"max_retries"`
	InitialBackoff    float64 `json:"initial_backoff" yaml:"initial_backoff"`
	MaxBackoff        float64 `json:"max_backoff" yaml:"max_backoff"`
	BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`
}

// NewConfiguration creates a new empty Config
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// Validate checks if the configuration is valid
func (c *Configuration) Validate() error {
	var validationErrors []string

	// Validate required fields for Agent
	if c.Agent.Name == "" {
		validationErrors = append(validationErrors, "Agent name is required")
	}

	// Validate Tool configuration
	if c.Agent.Tool.Name == "" {
		validationErrors = append(validationErrors, "Tool name is required")
	}
	if c.Agent.Tool.Description == "" {
		validationErrors = append(validationErrors, "Tool description is required")
	}
	if c.Agent.Tool.ArgumentName == "" {
		validationErrors = append(validationErrors, "Tool argument name is required")
	}
	if c.Agent.Tool.ArgumentDescription == "" {
		validationErrors = append(validationErrors, "Tool argument description is required")
	}

	// Validate required fields for LLM Service Config
	if c.Agent.LLM.APIKey == "" {
		validationErrors = append(validationErrors, "LLM API key is required")
	}
	if c.Agent.LLM.Provider == "" {
		validationErrors = append(validationErrors, "LLM provider is required")
	}
	if c.Agent.LLM.Model == "" {
		validationErrors = append(validationErrors, "LLM model is required")
	}
	if c.Agent.LLM.PromptTemplate == "" {
		validationErrors = append(validationErrors, "LLM prompt template is required")
	}

	// Validate prompt template
	if c.Agent.LLM.PromptTemplate != "" {
		err := c.validatePromptTemplate(c.Agent.LLM.PromptTemplate, c.Agent.Tool.ArgumentName)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid prompt template: %v", err))
		}
	}

	// If there are validation errors, return them
	if len(validationErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// validatePromptTemplate validates that the prompt template contains all required placeholders.
func (c *Configuration) validatePromptTemplate(template string, argumentName string) error {
	// Check if template is empty
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}

	// Extract all placeholders from the template
	placeholders, err := c.extractPlaceholders(template)
	if err != nil {
		return fmt.Errorf("failed to extract placeholders: %w", err)
	}

	// Check that required placeholders are present
	if !contains(placeholders, argumentName) && !contains(placeholders, "input") {
		return fmt.Errorf("template must contain either {{%s}} or {{input}} placeholder", argumentName)
	}

	// Tools should be included to list available tools
	if !contains(placeholders, "tools") {
		return fmt.Errorf("template must contain {{tools}} placeholder")
	}

	return nil
}

// extractPlaceholders extracts all placeholders ({{name}}) from a template string.
func (c *Configuration) extractPlaceholders(template string) ([]string, error) {
	r, err := regexp.Compile(`\{\{([^{}]+)\}\}`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile placeholder regex: %w", err)
	}

	matches := r.FindAllStringSubmatch(template, -1)
	var placeholders []string
	for _, match := range matches {
		if len(match) > 1 {
			placeholders = append(placeholders, strings.TrimSpace(match[1]))
		}
	}

	return placeholders, nil
}

// contains checks if a string slice contains a specific string.
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// parseLogLevel converts a string log level to a logrus.Level
func (c *Configuration) parseLogLevel(level string) (logrus.Level, error) {
	switch strings.ToLower(level) {
	case "panic":
		return logrus.PanicLevel, nil
	case "fatal":
		return logrus.FatalLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "warn", "warning":
		return logrus.WarnLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "debug":
		return logrus.DebugLevel, nil
	case "trace":
		return logrus.TraceLevel, nil
	default:
		return logrus.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// Apply the changes from another configuration to this one, without overwriting values not set in the new config
func (c *Configuration) Apply(newConfig *Configuration) *Configuration {
	// Handle nil case
	if newConfig == nil {
		return c
	}

	// Apply Runtime configuration
	if newConfig.Runtime.Log.RawLevel != "" {
		c.Runtime.Log.RawLevel = newConfig.Runtime.Log.RawLevel
	}
	if newConfig.Runtime.Log.RawOutput != "" {
		c.Runtime.Log.RawOutput = newConfig.Runtime.Log.RawOutput
	}

	// Apply transports configuration
	// HTTP settings - directly apply boolean for Enabled
	c.Runtime.Transports.HTTP.Enabled = newConfig.Runtime.Transports.HTTP.Enabled
	if newConfig.Runtime.Transports.HTTP.Host != "" {
		c.Runtime.Transports.HTTP.Host = newConfig.Runtime.Transports.HTTP.Host
	}
	if newConfig.Runtime.Transports.HTTP.Port != 0 {
		c.Runtime.Transports.HTTP.Port = newConfig.Runtime.Transports.HTTP.Port
	}

	// Stdio settings
	if newConfig.Runtime.Transports.Stdio.Enabled != c.Runtime.Transports.Stdio.Enabled {
		c.Runtime.Transports.Stdio.Enabled = newConfig.Runtime.Transports.Stdio.Enabled
	}
	if newConfig.Runtime.Transports.Stdio.BufferSize != 0 {
		c.Runtime.Transports.Stdio.BufferSize = newConfig.Runtime.Transports.Stdio.BufferSize
	}

	// Process the log level
	logLevel, err := c.parseLogLevel(c.Runtime.Log.RawLevel)
	if err == nil {
		c.Runtime.Log.LogLevel = logLevel
	} else {
		// Default to info level if invalid
		c.Runtime.Log.LogLevel = logrus.InfoLevel
	}

	// Handle log file output if RawOutput is a file path
	if c.Runtime.Log.RawOutput != "" {
		if c.Runtime.Log.RawOutput == "stdout" {
			c.Runtime.Log.Output = os.Stdout
		} else if c.Runtime.Log.RawOutput == "stderr" {
			c.Runtime.Log.Output = os.Stderr
		} else {
			// Try to open the log file
			file, err := os.OpenFile(c.Runtime.Log.RawOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// Default to stderr on error
				c.Runtime.Log.Output = os.Stderr
			} else {
				c.Runtime.Log.Output = file
			}
		}
	}

	// Apply Agent configuration
	if newConfig.Agent.Name != "" {
		c.Agent.Name = newConfig.Agent.Name
	}
	if newConfig.Agent.Version != "" {
		c.Agent.Version = newConfig.Agent.Version
	}

	// Apply Tool configuration
	if newConfig.Agent.Tool.Name != "" {
		c.Agent.Tool.Name = newConfig.Agent.Tool.Name
	}
	if newConfig.Agent.Tool.Description != "" {
		c.Agent.Tool.Description = newConfig.Agent.Tool.Description
	}
	if newConfig.Agent.Tool.ArgumentName != "" {
		c.Agent.Tool.ArgumentName = newConfig.Agent.Tool.ArgumentName
	}
	if newConfig.Agent.Tool.ArgumentDescription != "" {
		c.Agent.Tool.ArgumentDescription = newConfig.Agent.Tool.ArgumentDescription
	}

	// Apply LLM configuration
	if newConfig.Agent.LLM.Provider != "" {
		c.Agent.LLM.Provider = newConfig.Agent.LLM.Provider
	}
	if newConfig.Agent.LLM.Model != "" {
		c.Agent.LLM.Model = newConfig.Agent.LLM.Model
	}
	if newConfig.Agent.LLM.APIKey != "" || c.Agent.LLM.APIKey == "" {
		c.Agent.LLM.APIKey = newConfig.Agent.LLM.APIKey
	}
	if newConfig.Agent.LLM.IsMaxTokensSet {
		c.Agent.LLM.MaxTokens = newConfig.Agent.LLM.MaxTokens
		c.Agent.LLM.IsMaxTokensSet = true
	}
	if newConfig.Agent.LLM.IsTemperatureSet {
		c.Agent.LLM.Temperature = newConfig.Agent.LLM.Temperature
		c.Agent.LLM.IsTemperatureSet = true
	}
	if newConfig.Agent.LLM.PromptTemplate != "" {
		c.Agent.LLM.PromptTemplate = newConfig.Agent.LLM.PromptTemplate
	}

	// Apply retry configuration
	if newConfig.Agent.LLM.Retry.MaxRetries != 0 {
		c.Agent.LLM.Retry.MaxRetries = newConfig.Agent.LLM.Retry.MaxRetries
	}
	if newConfig.Agent.LLM.Retry.InitialBackoff != 0 {
		c.Agent.LLM.Retry.InitialBackoff = newConfig.Agent.LLM.Retry.InitialBackoff
	}
	if newConfig.Agent.LLM.Retry.MaxBackoff != 0 {
		c.Agent.LLM.Retry.MaxBackoff = newConfig.Agent.LLM.Retry.MaxBackoff
	}
	if newConfig.Agent.LLM.Retry.BackoffMultiplier != 0 {
		c.Agent.LLM.Retry.BackoffMultiplier = newConfig.Agent.LLM.Retry.BackoffMultiplier
	}

	// Apply chat configuration
	if newConfig.Agent.Chat.MaxTokens != 0 {
		c.Agent.Chat.MaxTokens = newConfig.Agent.Chat.MaxTokens
	}
	if newConfig.Agent.Chat.MaxLLMIterations != 0 {
		c.Agent.Chat.MaxLLMIterations = newConfig.Agent.Chat.MaxLLMIterations
	}
	if newConfig.Agent.Chat.RequestBudget != 0 {
		c.Agent.Chat.RequestBudget = newConfig.Agent.Chat.RequestBudget
	}

	// Apply Connections configuration if any MCP servers are defined
	if len(newConfig.Agent.Connections.McpServers) > 0 {
		// Initialize the map if it doesn't exist
		if c.Agent.Connections.McpServers == nil {
			c.Agent.Connections.McpServers = make(map[string]MCPServerConnection)
		}

		// Merge MCP servers
		for name, newServer := range newConfig.Agent.Connections.McpServers {
			oldServer, exists := c.Agent.Connections.McpServers[name]
			if !exists {
				c.Agent.Connections.McpServers[name] = newServer
				continue
			}

			// Overlay fields for each server
			if newServer.URL != "" {
				oldServer.URL = newServer.URL
			}
			if newServer.APIKey != "" {
				oldServer.APIKey = newServer.APIKey
			}
			if newServer.Command != "" {
				oldServer.Command = newServer.Command
			}
			if len(newServer.Args) > 0 {
				oldServer.Args = newServer.Args
			}
			if len(newServer.Environment) > 0 {
				oldServer.Environment = newServer.Environment
			}
			if newServer.IncludeTools != nil {
				oldServer.IncludeTools = newServer.IncludeTools
			}
			if newServer.ExcludeTools != nil {
				oldServer.ExcludeTools = newServer.ExcludeTools
			}
			c.Agent.Connections.McpServers[name] = oldServer
		}
	}

	// Update Connection retry settings
	if newConfig.Agent.Connections.Retry.MaxRetries != 0 {
		c.Agent.Connections.Retry.MaxRetries = newConfig.Agent.Connections.Retry.MaxRetries
	}
	if newConfig.Agent.Connections.Retry.InitialBackoff != 0 {
		c.Agent.Connections.Retry.InitialBackoff = newConfig.Agent.Connections.Retry.InitialBackoff
	}
	if newConfig.Agent.Connections.Retry.MaxBackoff != 0 {
		c.Agent.Connections.Retry.MaxBackoff = newConfig.Agent.Connections.Retry.MaxBackoff
	}
	if newConfig.Agent.Connections.Retry.BackoffMultiplier != 0 {
		c.Agent.Connections.Retry.BackoffMultiplier = newConfig.Agent.Connections.Retry.BackoffMultiplier
	}

	return c
}

// RedactedCopy returns a copy of the configuration with all sensitive/private data redacted for safe logging.
func (c *Configuration) RedactedCopy() *Configuration {
	copy := *c // shallow copy
	// Redact LLM API key
	copy.Agent.LLM.APIKey = "***REDACTED***"
	// Redact MCP server API keys
	if copy.Agent.Connections.McpServers != nil {
		redactedServers := make(map[string]MCPServerConnection, len(copy.Agent.Connections.McpServers))
		for k, v := range copy.Agent.Connections.McpServers {
			redacted := v
			redacted.APIKey = "***REDACTED***"
			redactedServers[k] = redacted
		}
		copy.Agent.Connections.McpServers = redactedServers
	}
	return &copy
}
