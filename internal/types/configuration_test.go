package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()
	assert.NotNil(t, config)
}

func TestConfiguration_Converters(t *testing.T) {
	baseConfig := NewConfiguration()
	baseConfig.Runtime.Log.DefaultLevel = "info"
	baseConfig.Runtime.Transports.Stdio.Enabled = true
	baseConfig.Runtime.Transports.Stdio.BufferSize = 4096
	baseConfig.Runtime.Transports.HTTP.Enabled = true
	baseConfig.Runtime.Transports.HTTP.Host = "127.0.0.1"
	baseConfig.Runtime.Transports.HTTP.Port = 8080

	baseConfig.Agent.Name = "agent-test"
	baseConfig.Agent.Version = "v1.2.3"
	baseConfig.Agent.Tool.Name = "tool1"
	baseConfig.Agent.Tool.Description = "desc1"
	baseConfig.Agent.Tool.ArgumentName = "arg1"
	baseConfig.Agent.Tool.ArgumentDescription = "argdesc1"
	baseConfig.Agent.Chat.MaxTokens = 1234
	baseConfig.Agent.Chat.MaxLLMIterations = 7
	baseConfig.Agent.Chat.RequestBudget = 1.23
	baseConfig.Agent.LLM.Provider = "openai"
	baseConfig.Agent.LLM.Model = "gpt-4"
	baseConfig.Agent.LLM.APIKey = "api-key-123"
	baseConfig.Agent.LLM.MaxTokens = 2048
	baseConfig.Agent.LLM.IsMaxTokensSet = true
	baseConfig.Agent.LLM.Temperature = 0.5
	baseConfig.Agent.LLM.PromptTemplate = "prompt {{arg1}}"
	baseConfig.Agent.LLM.Retry.MaxRetries = 2
	baseConfig.Agent.LLM.Retry.InitialBackoff = 1.5
	baseConfig.Agent.LLM.Retry.MaxBackoff = 10.0
	baseConfig.Agent.LLM.Retry.BackoffMultiplier = 2.5
	baseConfig.Agent.Connections.McpServers = map[string]MCPServerConnection{
		"srv1": {
			URL:    "http://srv1",
			APIKey: "srv1-key",
		},
	}
	baseConfig.Agent.Connections.Retry.MaxRetries = 3
	baseConfig.Agent.Connections.Retry.InitialBackoff = 2.0
	baseConfig.Agent.Connections.Retry.MaxBackoff = 20.0
	baseConfig.Agent.Connections.Retry.BackoffMultiplier = 3.0

	t.Run("GetAgentConfig", func(t *testing.T) {
		agentCfg := baseConfig.GetAgentConfig()
		assert.Equal(t, "tool1", agentCfg.Tool.Name)
		assert.Equal(t, "gpt-4", agentCfg.Model)
		assert.Equal(t, "prompt {{arg1}}", agentCfg.SystemPromptTemplate)
		assert.Equal(t, 1234, agentCfg.MaxTokens)
		assert.Equal(t, 7, agentCfg.MaxLLMIterations)
	})

	t.Run("GetLLMConfig", func(t *testing.T) {
		llmCfg := baseConfig.GetLLMConfig()
		assert.Equal(t, "openai", llmCfg.Provider)
		assert.Equal(t, "gpt-4", llmCfg.Model)
		assert.Equal(t, "api-key-123", llmCfg.APIKey)
		assert.Equal(t, 2048, llmCfg.MaxTokens)
		assert.True(t, llmCfg.IsMaxTokensSet)
		assert.Equal(t, 0.5, llmCfg.Temperature)
		assert.Equal(t, "prompt {{arg1}}", llmCfg.SystemPromptTemplate)
		assert.Equal(t, 2, llmCfg.RetryConfig.MaxRetries)
		assert.Equal(t, 1.5, llmCfg.RetryConfig.InitialBackoff)
		assert.Equal(t, 10.0, llmCfg.RetryConfig.MaxBackoff)
		assert.Equal(t, 2.5, llmCfg.RetryConfig.BackoffMultiplier)
	})

	t.Run("GetMCPServerConfig", func(t *testing.T) {
		mcpSrvCfg := baseConfig.GetMCPServerConfig()
		assert.Equal(t, "agent-test", mcpSrvCfg.Name)
		assert.Equal(t, "v1.2.3", mcpSrvCfg.Version)
		assert.Equal(t, true, mcpSrvCfg.HTTP.Enabled)
		assert.Equal(t, "127.0.0.1", mcpSrvCfg.HTTP.Host)
		assert.Equal(t, 8080, mcpSrvCfg.HTTP.Port)
		assert.Equal(t, true, mcpSrvCfg.Stdio.Enabled)
		assert.Equal(t, 4096, mcpSrvCfg.Stdio.BufferSize)
		assert.Equal(t, "tool1", mcpSrvCfg.Tool.Name)
	})

	t.Run("GetMCPConnectorConfig", func(t *testing.T) {
		mcpConnCfg := baseConfig.GetMCPConnectorConfig()
		assert.Contains(t, mcpConnCfg.McpServers, "srv1")
		assert.Equal(t, "http://srv1", mcpConnCfg.McpServers["srv1"].URL)
		assert.Equal(t, "srv1-key", mcpConnCfg.McpServers["srv1"].APIKey)
		assert.Equal(t, 3, mcpConnCfg.RetryConfig.MaxRetries)
		assert.Equal(t, 2.0, mcpConnCfg.RetryConfig.InitialBackoff)
		assert.Equal(t, 20.0, mcpConnCfg.RetryConfig.MaxBackoff)
		assert.Equal(t, 3.0, mcpConnCfg.RetryConfig.BackoffMultiplier)
	})
}

func TestBuildLogConfig(t *testing.T) {
	t.Run("valid info custom", func(t *testing.T) {
		raw := struct {
			DefaultLevel string `koanf:"defaultlevel" json:"defaultLevel" yaml:"defaultLevel"`
			Format       string `koanf:"format"`
			DisableMCP   bool   `koanf:"disablemcp" json:"disableMcp" yaml:"disableMcp"`
		}{
			DefaultLevel: "info",
			Format:       "custom",
			DisableMCP:   false,
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, "info", cfg.DefaultLevel)
		assert.Equal(t, "custom", cfg.Format)
		assert.Equal(t, false, cfg.DisableMCP)
		assert.Equal(t, "info", cfg.Level.String())
	})

	t.Run("valid debug json disable_mcp", func(t *testing.T) {
		raw := struct {
			DefaultLevel string `koanf:"defaultlevel" json:"defaultLevel" yaml:"defaultLevel"`
			Format       string `koanf:"format"`
			DisableMCP   bool   `koanf:"disablemcp" json:"disableMcp" yaml:"disableMcp"`
		}{
			DefaultLevel: "debug",
			Format:       "json",
			DisableMCP:   true,
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, "debug", cfg.DefaultLevel)
		assert.Equal(t, "json", cfg.Format)
		assert.Equal(t, true, cfg.DisableMCP)
		assert.Equal(t, "debug", cfg.Level.String())
	})

	t.Run("invalid level", func(t *testing.T) {
		raw := struct {
			DefaultLevel string `koanf:"defaultlevel" json:"defaultLevel" yaml:"defaultLevel"`
			Format       string `koanf:"format"`
			DisableMCP   bool   `koanf:"disablemcp" json:"disableMcp" yaml:"disableMcp"`
		}{
			DefaultLevel: "badlevel",
			Format:       "text",
			DisableMCP:   false,
		}
		_, err := BuildLogConfig(raw)
		assert.Error(t, err)
	})
}

func Test_parseLogLevel(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		expect  string
		isError bool
	}{
		{"empty string", "", "info", false},
		{"info", "info", "info", false},
		{"debug", "debug", "debug", false},
		{"warn", "warn", "warning", false},
		{"warning", "warning", "warning", false},
		{"error", "error", "error", false},
		{"fatal", "fatal", "fatal", false},
		{"panic", "panic", "panic", false},
		{"trace", "trace", "trace", false},
		{"invalid", "notalevel", "info", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lvl, err := parseLogLevel(c.input)
			if c.isError {
				if err == nil {
					t.Errorf("expected error for %q, got nil", c.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %q: %v", c.input, err)
				}
				if lvl.String() != c.expect {
					t.Errorf("expected %q, got %q", c.expect, lvl.String())
				}
			}
		})
	}
}
