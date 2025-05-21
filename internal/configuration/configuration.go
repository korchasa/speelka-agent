package configuration

import (
	"fmt"
	"github.com/korchasa/speelka-agent-go/internal/utils/log_formatter"

	"github.com/sirupsen/logrus"
)

// Configuration represents the complete configuration structure that matches the example files
// All nested structures are now inline (anonymous)
type Configuration struct {
	Runtime struct {
		Log struct {
			DefaultLevel string `koanf:"defaultlevel" json:"defaultLevel" yaml:"defaultLevel"`
			Format       string `koanf:"format"`
			DisableMCP   bool   `koanf:"disablemcp" json:"disableMcp" yaml:"disableMcp"`
		} `koanf:"log"`
		Transports struct {
			Stdio struct {
				Enabled    bool `koanf:"enabled"`
				BufferSize int  `koanf:"buffersize" json:"bufferSize" yaml:"bufferSize"`
			} `koanf:"stdio"`
			HTTP struct {
				Enabled bool   `koanf:"enabled"`
				Host    string `koanf:"host"`
				Port    int    `koanf:"port"`
			} `koanf:"http"`
		} `koanf:"transports"`
	} `koanf:"runtime"`
	Agent struct {
		Name    string `koanf:"name"`
		Version string `koanf:"version"`
		Tool    struct {
			Name                string `koanf:"name"`
			Description         string `koanf:"description"`
			ArgumentName        string `koanf:"argumentname" json:"argumentName" yaml:"argumentName"`
			ArgumentDescription string `koanf:"argumentdescription" json:"argumentDescription" yaml:"argumentDescription"`
		} `koanf:"tool"`
		Chat struct {
			MaxTokens        int     `koanf:"maxtokens" json:"maxTokens" yaml:"maxTokens"`
			MaxLLMIterations int     `koanf:"maxllmiterations" json:"maxLLMIterations" yaml:"maxLLMIterations"`
			RequestBudget    float64 `koanf:"requestbudget" json:"requestBudget" yaml:"requestBudget"`
		} `koanf:"chat"`
		LLM struct {
			Provider       string  `koanf:"provider"`
			Model          string  `koanf:"model"`
			APIKey         string  `koanf:"apikey" json:"apiKey" yaml:"apiKey"`
			MaxTokens      int     `koanf:"maxtokens" json:"maxTokens" yaml:"maxTokens"`
			Temperature    float64 `koanf:"temperature"`
			PromptTemplate string  `koanf:"prompttemplate" json:"promptTemplate" yaml:"promptTemplate"`
			Retry          struct {
				MaxRetries        int     `koanf:"maxretries" json:"maxRetries" yaml:"maxRetries"`
				InitialBackoff    float64 `koanf:"initialbackoff" json:"initialBackoff" yaml:"initialBackoff"`
				MaxBackoff        float64 `koanf:"maxbackoff" json:"maxBackoff" yaml:"maxBackoff"`
				BackoffMultiplier float64 `koanf:"backoffmultiplier" json:"backoffMultiplier" yaml:"backoffMultiplier"`
			} `koanf:"retry"`
			IsMaxTokensSet bool `koanf:"ismaxtokensset" json:"isMaxTokensSet" yaml:"isMaxTokensSet"`
		} `koanf:"llm"`
		Connections struct {
			McpServers map[string]MCPServerConnection `koanf:"mcpservers" json:"mcpServers" yaml:"mcpServers"`
			Retry      struct {
				MaxRetries        int     `koanf:"maxretries" json:"maxRetries" yaml:"maxRetries"`
				InitialBackoff    float64 `koanf:"initialbackoff" json:"initialBackoff" yaml:"initialBackoff"`
				MaxBackoff        float64 `koanf:"maxbackoff" json:"maxBackoff" yaml:"maxBackoff"`
				BackoffMultiplier float64 `koanf:"backoffmultiplier" json:"backoffMultiplier" yaml:"backoffMultiplier"`
			} `koanf:"retry"`
		} `koanf:"connections"`
	} `koanf:"agent"`
}

// NewConfiguration creates a new empty config
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// GetAgentConfig converts *Configuration to AgentConfig
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

// GetLLMConfig converts *Configuration to LLMConfig
func (c *Configuration) GetLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:             c.Agent.LLM.Provider,
		Model:                c.Agent.LLM.Model,
		APIKey:               c.Agent.LLM.APIKey,
		MaxTokens:            c.Agent.LLM.MaxTokens,
		IsMaxTokensSet:       c.Agent.LLM.IsMaxTokensSet,
		Temperature:          c.Agent.LLM.Temperature,
		SystemPromptTemplate: c.Agent.LLM.PromptTemplate,
		RetryConfig: RetryConfig{
			MaxRetries:        c.Agent.LLM.Retry.MaxRetries,
			InitialBackoff:    c.Agent.LLM.Retry.InitialBackoff,
			MaxBackoff:        c.Agent.LLM.Retry.MaxBackoff,
			BackoffMultiplier: c.Agent.LLM.Retry.BackoffMultiplier,
		},
	}
}

// GetMCPServerConfig converts *Configuration to MCPServerConfig
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
		MCPLogEnabled: !c.Runtime.Log.DisableMCP,
	}
}

// GetMCPConnectorConfig converts *Configuration to MCPConnectorConfig
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

// BuildLogConfig creates a business LogConfig structure from the raw Log config
func (c *Configuration) BuildLogConfig() (LogConfig, error) {
	// Parse log level
	level, err := parseLogLevel(c.Runtime.Log.DefaultLevel)
	if err != nil {
		return LogConfig{}, fmt.Errorf("failed to parse log level `%s`: %w", c.Runtime.Log.DefaultLevel, err)
	}
	formatter, err := parseFormatter(c.Runtime.Log.Format)
	if err != nil {
		return LogConfig{}, fmt.Errorf("failed to parse log format `%s`: %w", c.Runtime.Log.Format, err)
	}
	return LogConfig{
		Level:      level,
		Formatter:  formatter,
		DisableMCP: c.Runtime.Log.DisableMCP,
	}, nil
}

// parseLogLevel converts log level string to logrus.Level
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

func parseFormatter(format string) (logrus.Formatter, error) {
	switch format {
	case "json":
		return &logrus.JSONFormatter{}, nil
	case "text":
		return &logrus.TextFormatter{}, nil
	case "auto":
		return &log_formatter.CustomLogFormatter{}, nil
	default:
		return nil, fmt.Errorf("unexpected log format: %s", format)
	}
}
