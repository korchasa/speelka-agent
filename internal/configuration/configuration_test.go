package configuration

import (
	"testing"

	"github.com/sirupsen/logrus"
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

func TestParseLogLevel(t *testing.T) {
	tbl := []struct {
		in   string
		out  logrus.Level
		fail bool
	}{
		{"panic", logrus.PanicLevel, false},
		{"fatal", logrus.FatalLevel, false},
		{"error", logrus.ErrorLevel, false},
		{"warn", logrus.WarnLevel, false},
		{"warning", logrus.WarnLevel, false},
		{"info", logrus.InfoLevel, false},
		{"debug", logrus.DebugLevel, false},
		{"trace", logrus.TraceLevel, false},
		{"", logrus.InfoLevel, false},
		{"bad", logrus.InfoLevel, true},
	}
	for _, tc := range tbl {
		lvl, err := parseLogLevel(tc.in)
		if tc.fail {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.out, lvl)
		}
	}
}

func TestBuildLogConfig(t *testing.T) {
	c := &Configuration{}
	c.Runtime.Log.DefaultLevel = "info"
	c.Runtime.Log.Format = "text"
	c.Runtime.Log.DisableMCP = false
	cfg, err := c.BuildLogConfig()
	assert.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, cfg.Level)
	_, isText := cfg.Formatter.(*logrus.TextFormatter)
	assert.True(t, isText)
	assert.False(t, cfg.DisableMCP)

	c.Runtime.Log.DefaultLevel = "bad"
	_, err = c.BuildLogConfig()
	assert.Error(t, err)
}

func TestGetAgentConfig(t *testing.T) {
	c := &Configuration{}
	c.Agent.Tool.Name = "tool"
	c.Agent.Tool.Description = "desc"
	c.Agent.Tool.ArgumentName = "arg"
	c.Agent.Tool.ArgumentDescription = "argdesc"
	c.Agent.LLM.Model = "model"
	c.Agent.LLM.PromptTemplate = "tmpl"
	c.Agent.Chat.MaxTokens = 42
	c.Agent.Chat.MaxLLMIterations = 3
	ac := c.GetAgentConfig()
	assert.Equal(t, "tool", ac.Tool.Name)
	assert.Equal(t, "desc", ac.Tool.Description)
	assert.Equal(t, "arg", ac.Tool.ArgumentName)
	assert.Equal(t, "argdesc", ac.Tool.ArgumentDescription)
	assert.Equal(t, "model", ac.Model)
	assert.Equal(t, "tmpl", ac.SystemPromptTemplate)
	assert.Equal(t, 42, ac.MaxTokens)
	assert.Equal(t, 3, ac.MaxLLMIterations)
}

func TestGetLLMConfig(t *testing.T) {
	c := &Configuration{}
	c.Agent.LLM.Provider = "prov"
	c.Agent.LLM.Model = "model"
	c.Agent.LLM.APIKey = "key"
	c.Agent.LLM.MaxTokens = 42
	c.Agent.LLM.IsMaxTokensSet = true
	c.Agent.LLM.Temperature = 0.5
	c.Agent.LLM.PromptTemplate = "tmpl"
	c.Agent.LLM.Retry.MaxRetries = 1
	c.Agent.LLM.Retry.InitialBackoff = 2
	c.Agent.LLM.Retry.MaxBackoff = 3
	c.Agent.LLM.Retry.BackoffMultiplier = 4
	llm := c.GetLLMConfig()
	assert.Equal(t, "prov", llm.Provider)
	assert.Equal(t, "model", llm.Model)
	assert.Equal(t, "key", llm.APIKey)
	assert.Equal(t, 42, llm.MaxTokens)
	assert.True(t, llm.IsMaxTokensSet)
	assert.Equal(t, 0.5, llm.Temperature)
	assert.Equal(t, "tmpl", llm.SystemPromptTemplate)
	assert.Equal(t, 1, llm.RetryConfig.MaxRetries)
	assert.Equal(t, 2.0, llm.RetryConfig.InitialBackoff)
	assert.Equal(t, 3.0, llm.RetryConfig.MaxBackoff)
	assert.Equal(t, 4.0, llm.RetryConfig.BackoffMultiplier)
}

func TestGetMCPServerConfig(t *testing.T) {
	c := &Configuration{}
	c.Agent.Name = "agent"
	c.Agent.Version = "v1"
	c.Runtime.Transports.HTTP.Enabled = true
	c.Runtime.Transports.HTTP.Host = "host"
	c.Runtime.Transports.HTTP.Port = 123
	c.Runtime.Transports.Stdio.Enabled = false
	c.Runtime.Transports.Stdio.BufferSize = 10
	c.Agent.Tool.Name = "tool"
	c.Agent.Tool.Description = "desc"
	c.Agent.Tool.ArgumentName = "arg"
	c.Agent.Tool.ArgumentDescription = "argdesc"
	c.Runtime.Log.DisableMCP = false
	cfg := c.GetMCPServerConfig()
	assert.Equal(t, "agent", cfg.Name)
	assert.Equal(t, "v1", cfg.Version)
	assert.True(t, cfg.HTTP.Enabled)
	assert.Equal(t, "host", cfg.HTTP.Host)
	assert.Equal(t, 123, cfg.HTTP.Port)
	assert.False(t, cfg.Stdio.Enabled)
	assert.Equal(t, 10, cfg.Stdio.BufferSize)
	assert.Equal(t, "tool", cfg.Tool.Name)
	assert.Equal(t, "desc", cfg.Tool.Description)
	assert.Equal(t, "arg", cfg.Tool.ArgumentName)
	assert.Equal(t, "argdesc", cfg.Tool.ArgumentDescription)
	assert.True(t, cfg.MCPLogEnabled)
}

func TestGetMCPConnectorConfig(t *testing.T) {
	c := &Configuration{}
	c.Agent.Connections.McpServers = map[string]MCPServerConnection{"srv": {URL: "url"}}
	c.Agent.Connections.Retry.MaxRetries = 1
	c.Agent.Connections.Retry.InitialBackoff = 2
	c.Agent.Connections.Retry.MaxBackoff = 3
	c.Agent.Connections.Retry.BackoffMultiplier = 4
	cfg := c.GetMCPConnectorConfig()
	assert.Contains(t, cfg.McpServers, "srv")
	assert.Equal(t, "url", cfg.McpServers["srv"].URL)
	assert.Equal(t, 1, cfg.RetryConfig.MaxRetries)
	assert.Equal(t, 2.0, cfg.RetryConfig.InitialBackoff)
	assert.Equal(t, 3.0, cfg.RetryConfig.MaxBackoff)
	assert.Equal(t, 4.0, cfg.RetryConfig.BackoffMultiplier)
}
