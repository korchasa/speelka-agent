package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"io"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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
	baseConfig.Agent.LLM.Provider = "openai"
	baseConfig.Agent.LLM.Model = "gpt-4"
	baseConfig.Agent.LLM.APIKey = "api-key-123"
	baseConfig.Agent.LLM.MaxTokens = 2048
	baseConfig.Agent.LLM.IsMaxTokensSet = true
	baseConfig.Agent.LLM.Temperature = 0.5
	baseConfig.Agent.LLM.IsTemperatureSet = true
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
		assert.True(t, llmCfg.IsTemperatureSet)
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

func TestConfiguration_Serialization_Golden(t *testing.T) {
	cfg := NewConfiguration()
	goldenPath := filepath.Join("testdata", "configuration_golden.json")
	actual, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	goldenFile, err := os.Open(goldenPath)
	if err != nil {
		// If the golden file is not found, create it
		t.Logf("Golden file not found, creating: %s", goldenPath)
		f, err := os.Create(goldenPath)
		if err != nil {
			t.Fatalf("failed to create golden file: %v", err)
		}
		defer f.Close()
		_, err = f.Write(actual)
		if err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}
	defer goldenFile.Close()
	golden, err := io.ReadAll(goldenFile)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}
	if string(actual) != string(golden) {
		t.Errorf("configuration serialization mismatch with golden file\nActual:\n%s\nGolden:\n%s", actual, golden)
	}
}

func TestBuildLogConfig(t *testing.T) {
	t.Run("valid info custom", func(t *testing.T) {
		raw := struct {
			DefaultLevel string `json:"default_level" yaml:"default_level"`
			Format       string `json:"format" yaml:"format"`
			DisableMCP   bool   `json:"disable_mcp" yaml:"disable_mcp"`
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
			DefaultLevel string `json:"default_level" yaml:"default_level"`
			Format       string `json:"format" yaml:"format"`
			DisableMCP   bool   `json:"disable_mcp" yaml:"disable_mcp"`
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
			DefaultLevel string `json:"default_level" yaml:"default_level"`
			Format       string `json:"format" yaml:"format"`
			DisableMCP   bool   `json:"disable_mcp" yaml:"disable_mcp"`
		}{
			DefaultLevel: "badlevel",
			Format:       "text",
			DisableMCP:   false,
		}
		_, err := BuildLogConfig(raw)
		assert.Error(t, err)
	})
}

func TestConfiguration_Unmarshal_Inline(t *testing.T) {
	yamlData := `
runtime:
  log:
    default_level: info
    output: ':mcp:'
    format: text
  transports:
    stdio:
      enabled: true
      buffer_size: 1024
    http:
      enabled: false
      host: localhost
      port: 3000
agent:
  name: "speelka-agent"
  version: "v1.0.0"
  tool:
    name: "process"
    description: "Process tool for user queries"
    argument_name: "input"
    argument_description: "User query"
  chat:
    max_tokens: 0
    max_llm_iterations: 25
    request_budget: 0.0
  llm:
    provider: "openai"
    api_key: "dummy-api-key"
    model: "gpt-4o"
    temperature: 0.7
    prompt_template: "You are a helpful assistant. {{input}}. Available tools: {{tools}}"
    retry:
      max_retries: 3
      initial_backoff: 1.0
      max_backoff: 30.0
      backoff_multiplier: 2.0
  connections:
    mcpServers:
      time:
        command: "docker"
        args: ["run", "-i", "--rm", "mcp/time"]
        timeout: 10
      filesystem:
        command: "mcp-filesystem-server"
        args: ["/path/to/directory"]
    retry:
      max_retries: 2
      initial_backoff: 1.5
      max_backoff: 10.0
      backoff_multiplier: 2.5
`
	var cfg Configuration
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "speelka-agent", cfg.Agent.Name)
	assert.Equal(t, 25, cfg.Agent.Chat.MaxLLMIterations)
	assert.Equal(t, "openai", cfg.Agent.LLM.Provider)
	assert.Equal(t, 1024, cfg.Runtime.Transports.Stdio.BufferSize)
	assert.Equal(t, false, cfg.Runtime.Transports.HTTP.Enabled)
	assert.Equal(t, "docker", cfg.Agent.Connections.McpServers["time"].Command)
}
