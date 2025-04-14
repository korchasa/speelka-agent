package configuration

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLoader_LoadConfiguration(t *testing.T) {
	// Create a new DefaultLoader
	loader := NewDefaultLoader()

	// Load default configuration
	config, err := loader.LoadConfiguration()

	// Assert that there is no error
	assert.NoError(t, err)

	// Assert that the config is not nil
	assert.NotNil(t, config)

	// Verify default runtime values
	assert.Equal(t, "info", config.Runtime.Log.RawLevel)
	assert.Equal(t, "stderr", config.Runtime.Log.RawOutput)

	// Verify default transport values
	assert.Equal(t, true, config.Runtime.Transports.Stdio.Enabled)
	assert.Equal(t, 8192, config.Runtime.Transports.Stdio.BufferSize)
	assert.Equal(t, false, config.Runtime.Transports.HTTP.Enabled)
	assert.Equal(t, "localhost", config.Runtime.Transports.HTTP.Host)
	assert.Equal(t, 3000, config.Runtime.Transports.HTTP.Port)

	// Verify default agent values
	assert.Equal(t, "speelka-agent", config.Agent.Name)
	assert.Equal(t, "1.0.0", config.Agent.Version)
	assert.Equal(t, "process", config.Agent.Tool.Name)
	assert.Equal(t, "Process user queries with LLM", config.Agent.Tool.Description)
	assert.Equal(t, "input", config.Agent.Tool.ArgumentName)
	assert.Equal(t, "The user query to process", config.Agent.Tool.ArgumentDescription)

	// Verify default LLM values
	assert.Equal(t, "openai", config.Agent.LLM.Provider)
	assert.Equal(t, "gpt-4", config.Agent.LLM.Model)
	assert.Equal(t, 0.7, config.Agent.LLM.Temperature)

	// Verify default LLM retry values
	assert.Equal(t, 3, config.Agent.LLM.Retry.MaxRetries)
	assert.Equal(t, 1.0, config.Agent.LLM.Retry.InitialBackoff)
	assert.Equal(t, 30.0, config.Agent.LLM.Retry.MaxBackoff)
	assert.Equal(t, 2.0, config.Agent.LLM.Retry.BackoffMultiplier)

	// Verify default Chat values
	assert.Equal(t, 8192, config.Agent.Chat.MaxTokens)
	assert.Equal(t, "delete-old", config.Agent.Chat.CompactionStrategy)
	assert.Equal(t, 25, config.Agent.Chat.MaxLLMIterations)

	// Verify default Connection retry values
	assert.Equal(t, 3, config.Agent.Connections.Retry.MaxRetries)
	assert.Equal(t, 1.0, config.Agent.Connections.Retry.InitialBackoff)
	assert.Equal(t, 30.0, config.Agent.Connections.Retry.MaxBackoff)
	assert.Equal(t, 2.0, config.Agent.Connections.Retry.BackoffMultiplier)

	// Verify empty server connections map
	assert.NotNil(t, config.Agent.Connections.McpServers)
	assert.Empty(t, config.Agent.Connections.McpServers)

	// After Apply, check parsed fields
	config.Apply(config)
	assert.Equal(t, "info", config.Runtime.Log.RawLevel)
	assert.Equal(t, "stderr", config.Runtime.Log.RawOutput)
	assert.Equal(t, os.Stderr, config.Runtime.Log.Output)
	assert.Equal(t, logrus.InfoLevel, config.Runtime.Log.LogLevel)
}
