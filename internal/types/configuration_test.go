package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"io"

	"github.com/stretchr/testify/assert"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()
	assert.NotNil(t, config)
}

func TestConfiguration_Converters(t *testing.T) {
	baseConfig := &Configuration{
		Runtime: RuntimeConfig{
			Log: RuntimeLogConfig{
				DefaultLevel: "info",
				Output:       LogOutputStdout,
			},
			Transports: RuntimeTransportConfig{
				Stdio: RuntimeStdioConfig{
					Enabled:    true,
					BufferSize: 4096,
				},
				HTTP: RuntimeHTTPConfig{
					Enabled: true,
					Host:    "127.0.0.1",
					Port:    8080,
				},
			},
		},
		Agent: ConfigAgent{
			Name:    "agent-test",
			Version: "v1.2.3",
			Tool: AgentToolConfig{
				Name:                "tool1",
				Description:         "desc1",
				ArgumentName:        "arg1",
				ArgumentDescription: "argdesc1",
			},
			Chat: AgentChatConfig{
				MaxTokens:        1234,
				MaxLLMIterations: 7,
			},
			LLM: AgentLLMConfig{
				Provider:         "openai",
				Model:            "gpt-4",
				APIKey:           "api-key-123",
				MaxTokens:        2048,
				IsMaxTokensSet:   true,
				Temperature:      0.5,
				IsTemperatureSet: true,
				PromptTemplate:   "prompt {{arg1}}",
				Retry: LLMRetryConfig{
					MaxRetries:        2,
					InitialBackoff:    1.5,
					MaxBackoff:        10.0,
					BackoffMultiplier: 2.5,
				},
			},
			Connections: AgentConnectionsConfig{
				McpServers: map[string]MCPServerConnection{
					"srv1": {
						URL:    "http://srv1",
						APIKey: "srv1-key",
					},
				},
				Retry: ConnectionRetryConfig{
					MaxRetries:        3,
					InitialBackoff:    2.0,
					MaxBackoff:        20.0,
					BackoffMultiplier: 3.0,
				},
			},
		},
	}

	t.Run("ToAgentConfig", func(t *testing.T) {
		agentCfg := baseConfig.ToAgentConfig()
		assert.Equal(t, "tool1", agentCfg.Tool.Name)
		assert.Equal(t, "gpt-4", agentCfg.Model)
		assert.Equal(t, "prompt {{arg1}}", agentCfg.SystemPromptTemplate)
		assert.Equal(t, 1234, agentCfg.MaxTokens)
		assert.Equal(t, 7, agentCfg.MaxLLMIterations)
	})

	t.Run("ToLLMConfig", func(t *testing.T) {
		llmCfg := baseConfig.ToLLMConfig()
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

	t.Run("ToMCPServerConfig", func(t *testing.T) {
		mcpSrvCfg := baseConfig.ToMCPServerConfig()
		assert.Equal(t, "agent-test", mcpSrvCfg.Name)
		assert.Equal(t, "v1.2.3", mcpSrvCfg.Version)
		assert.Equal(t, true, mcpSrvCfg.HTTP.Enabled)
		assert.Equal(t, "127.0.0.1", mcpSrvCfg.HTTP.Host)
		assert.Equal(t, 8080, mcpSrvCfg.HTTP.Port)
		assert.Equal(t, true, mcpSrvCfg.Stdio.Enabled)
		assert.Equal(t, 4096, mcpSrvCfg.Stdio.BufferSize)
		assert.Equal(t, "tool1", mcpSrvCfg.Tool.Name)
		assert.Equal(t, LogOutputStdout, mcpSrvCfg.LogRawOutput)
	})

	t.Run("ToMCPConnectorConfig", func(t *testing.T) {
		mcpConnCfg := baseConfig.ToMCPConnectorConfig()
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
		// Если golden-файл не найден — создаём его
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
	t.Run("valid stdout info custom", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "info",
			Output:       LogOutputStdout,
			Format:       "custom",
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, "info", cfg.DefaultLevel)
		assert.Equal(t, LogOutputStdout, cfg.Output)
		assert.Equal(t, "custom", cfg.Format)
		assert.Equal(t, "info", cfg.DefaultLevel)
		assert.Equal(t, false, cfg.UseMCPLogs)
		assert.Equal(t, "info", cfg.Level.String())
	})

	t.Run("valid stderr debug json", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "debug",
			Output:       LogOutputStderr,
			Format:       "json",
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, LogOutputStderr, cfg.Output)
		assert.Equal(t, "json", cfg.Format)
		assert.Equal(t, "debug", cfg.Level.String())
	})

	t.Run("valid mcp warn text", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "warn",
			Output:       LogOutputMCP,
			Format:       "text",
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, LogOutputMCP, cfg.Output)
		assert.Equal(t, true, cfg.UseMCPLogs)
		assert.Equal(t, "warning", cfg.Level.String())
	})

	t.Run("file output", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "error",
			Output:       "/tmp/test.log",
			Format:       "custom",
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, "/tmp/test.log", cfg.Output)
		assert.Equal(t, "error", cfg.Level.String())
	})

	t.Run("invalid level", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "notalevel",
			Output:       LogOutputStdout,
			Format:       "custom",
		}
		_, err := BuildLogConfig(raw)
		assert.Error(t, err)
	})

	t.Run("invalid output", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "info",
			Output:       "",
			Format:       "custom",
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, "", cfg.Output)
	})

	t.Run("invalid format", func(t *testing.T) {
		raw := RuntimeLogConfig{
			DefaultLevel: "info",
			Output:       LogOutputStdout,
			Format:       "unknown",
		}
		cfg, err := BuildLogConfig(raw)
		assert.NoError(t, err)
		assert.Equal(t, "unknown", cfg.Format)
	})
}
