package types

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Configuration represents the complete configuration structure that matches the example files
// Теперь все вложенные структуры inline (анонимные)
type Configuration struct {
	Runtime struct {
		Log struct {
			DefaultLevel string `json:"default_level" yaml:"default_level"`
			Output       string `json:"output" yaml:"output"`
			Format       string `json:"format" yaml:"format"`
		} `json:"log" yaml:"log"`
		Transports struct {
			Stdio struct {
				Enabled    bool `json:"enabled" yaml:"enabled"`
				BufferSize int  `json:"buffer_size" yaml:"buffer_size"`
			} `json:"stdio" yaml:"stdio"`
			HTTP struct {
				Enabled bool   `json:"enabled" yaml:"enabled"`
				Host    string `json:"host" yaml:"host"`
				Port    int    `json:"port" yaml:"port"`
			} `json:"http" yaml:"http"`
		} `json:"transports" yaml:"transports"`
	} `json:"runtime" yaml:"runtime"`
	Agent struct {
		Name    string `json:"name" yaml:"name"`
		Version string `json:"version" yaml:"version"`
		Tool    struct {
			Name                string `json:"name" yaml:"name"`
			Description         string `json:"description" yaml:"description"`
			ArgumentName        string `json:"argument_name" yaml:"argument_name"`
			ArgumentDescription string `json:"argument_description" yaml:"argument_description"`
		} `json:"tool" yaml:"tool"`
		Chat struct {
			MaxTokens        int     `json:"max_tokens" yaml:"max_tokens"`
			MaxLLMIterations int     `json:"max_llm_iterations" yaml:"max_llm_iterations"`
			RequestBudget    float64 `json:"request_budget" yaml:"request_budget"`
		} `json:"chat" yaml:"chat"`
		LLM struct {
			Provider       string  `json:"provider" yaml:"provider"`
			Model          string  `json:"model" yaml:"model"`
			APIKey         string  `json:"api_key" yaml:"api_key"`
			MaxTokens      int     `json:"max_tokens" yaml:"max_tokens"`
			Temperature    float64 `json:"temperature" yaml:"temperature"`
			PromptTemplate string  `json:"prompt_template" yaml:"prompt_template"`
			Retry          struct {
				MaxRetries        int     `json:"max_retries" yaml:"max_retries"`
				InitialBackoff    float64 `json:"initial_backoff" yaml:"initial_backoff"`
				MaxBackoff        float64 `json:"max_backoff" yaml:"max_backoff"`
				BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`
			} `json:"retry" yaml:"retry"`
			// Internal flags
			IsMaxTokensSet   bool
			IsTemperatureSet bool
		} `json:"llm" yaml:"llm"`
		Connections struct {
			McpServers map[string]MCPServerConnection `json:"mcpServers" yaml:"mcpServers"`
			Retry      struct {
				MaxRetries        int     `json:"max_retries" yaml:"max_retries"`
				InitialBackoff    float64 `json:"initial_backoff" yaml:"initial_backoff"`
				MaxBackoff        float64 `json:"max_backoff" yaml:"max_backoff"`
				BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`
			} `json:"retry" yaml:"retry"`
		} `json:"connections" yaml:"connections"`
	} `json:"agent" yaml:"agent"`
}

// NewConfiguration создает новый пустой конфиг
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// GetAgentConfig преобразует *Configuration в AgentConfig
func (c *Configuration) GetAgentConfig() AgentConfig {
	return AgentConfig{
		Tool: MCPServerToolConfig{
			Name:                c.Agent.Tool.Name,
			Description:         c.Agent.Tool.Description,
			ArgumentName:        c.Agent.Tool.ArgumentName,
			ArgumentDescription: c.Agent.Tool.ArgumentDescription,
		},
		Model:                c.Agent.LLM.Model,
		SystemPromptTemplate: c.Agent.LLM.PromptTemplate,
		MaxTokens:            c.Agent.Chat.MaxTokens,
		MaxLLMIterations:     c.Agent.Chat.MaxLLMIterations,
	}
}

// GetLLMConfig преобразует *Configuration в LLMConfig
func (c *Configuration) GetLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:             c.Agent.LLM.Provider,
		Model:                c.Agent.LLM.Model,
		APIKey:               c.Agent.LLM.APIKey,
		MaxTokens:            c.Agent.LLM.MaxTokens,
		IsMaxTokensSet:       c.Agent.LLM.IsMaxTokensSet,
		Temperature:          c.Agent.LLM.Temperature,
		IsTemperatureSet:     c.Agent.LLM.IsTemperatureSet,
		SystemPromptTemplate: c.Agent.LLM.PromptTemplate,
		RetryConfig: RetryConfig{
			MaxRetries:        c.Agent.LLM.Retry.MaxRetries,
			InitialBackoff:    c.Agent.LLM.Retry.InitialBackoff,
			MaxBackoff:        c.Agent.LLM.Retry.MaxBackoff,
			BackoffMultiplier: c.Agent.LLM.Retry.BackoffMultiplier,
		},
	}
}

// GetMCPServerConfig преобразует *Configuration в MCPServerConfig
func (c *Configuration) GetMCPServerConfig() MCPServerConfig {
	return MCPServerConfig{
		Name:    c.Agent.Name,
		Version: c.Agent.Version,
		HTTP: HTTPConfig{
			Enabled: c.Runtime.Transports.HTTP.Enabled,
			Host:    c.Runtime.Transports.HTTP.Host,
			Port:    c.Runtime.Transports.HTTP.Port,
		},
		Stdio: StdioConfig{
			Enabled:    c.Runtime.Transports.Stdio.Enabled,
			BufferSize: c.Runtime.Transports.Stdio.BufferSize,
		},
		Tool: MCPServerToolConfig{
			Name:                c.Agent.Tool.Name,
			Description:         c.Agent.Tool.Description,
			ArgumentName:        c.Agent.Tool.ArgumentName,
			ArgumentDescription: c.Agent.Tool.ArgumentDescription,
		},
		Debug:        false, // Можно добавить отдельное поле в конфиг при необходимости
		LogRawOutput: c.Runtime.Log.Output,
	}
}

// GetMCPConnectorConfig преобразует *Configuration в MCPConnectorConfig
func (c *Configuration) GetMCPConnectorConfig() MCPConnectorConfig {
	return MCPConnectorConfig{
		McpServers: c.Agent.Connections.McpServers,
		RetryConfig: RetryConfig{
			MaxRetries:        c.Agent.Connections.Retry.MaxRetries,
			InitialBackoff:    c.Agent.Connections.Retry.InitialBackoff,
			MaxBackoff:        c.Agent.Connections.Retry.MaxBackoff,
			BackoffMultiplier: c.Agent.Connections.Retry.BackoffMultiplier,
		},
	}
}

// BuildLogConfig создает бизнес-структуру LogConfig из сырого Log-конфига
func BuildLogConfig(raw struct {
	DefaultLevel string `json:"default_level" yaml:"default_level"`
	Output       string `json:"output" yaml:"output"`
	Format       string `json:"format" yaml:"format"`
}) (LogConfig, error) {
	cfg := LogConfig{
		DefaultLevel: raw.DefaultLevel,
		Output:       raw.Output,
		Format:       raw.Format,
	}

	// Парсинг уровня логирования
	level, err := parseLogLevel(raw.DefaultLevel)
	if err != nil {
		return LogConfig{}, err
	}
	cfg.Level = level

	// MCP-вывод
	cfg.UseMCPLogs = (raw.Output == LogOutputMCP)

	return cfg, nil
}

// parseLogLevel преобразует строку уровня логирования в logrus.Level
func parseLogLevel(level string) (logrus.Level, error) {
	switch level {
	case "panic":
		return logrus.PanicLevel, nil
	case "fatal":
		return logrus.FatalLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "warn", "warning":
		return logrus.WarnLevel, nil
	case "info", "":
		return logrus.InfoLevel, nil
	case "debug":
		return logrus.DebugLevel, nil
	case "trace":
		return logrus.TraceLevel, nil
	default:
		return logrus.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}
