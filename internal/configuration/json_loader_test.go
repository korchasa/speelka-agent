package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONLoader_LoadConfiguration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "json-loader-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid JSON configuration file
	validJSON := `
{
  "runtime": {
    "log": {
      "level": "debug",
      "output": "./test.log"
    }
  },
  "agent": {
    "name": "test-agent",
    "tool": {
      "name": "test-tool",
      "description": "A test tool",
      "argument_name": "query",
      "argument_description": "The query to process"
    },
    "llm": {
      "provider": "openai",
      "model": "gpt-4",
      "api_key": "test-api-key",
      "prompt_template": "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}"
    }
  }
}
`

	validJSONPath := filepath.Join(tempDir, "valid-config.json")
	err = os.WriteFile(validJSONPath, []byte(validJSON), 0644)
	require.NoError(t, err)

	// Create an invalid JSON file
	invalidJSON := `
{
  "runtime": {
    "log": {
      "level": "debug",
      "output": invalid json syntax
    }
  }
}
`

	invalidJSONPath := filepath.Join(tempDir, "invalid-config.json")
	err = os.WriteFile(invalidJSONPath, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	// Create a non-JSON file
	nonJSONPath := filepath.Join(tempDir, "config.txt")
	err = os.WriteFile(nonJSONPath, []byte("This is not a JSON file"), 0644)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name        string
		filePath    string
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, config *types.Configuration)
	}{
		{
			name:        "Valid JSON file",
			filePath:    validJSONPath,
			expectError: false,
			validate: func(t *testing.T, config *types.Configuration) {
				assert.Equal(t, "test-agent", config.Agent.Name)
				assert.Equal(t, "test-tool", config.Agent.Tool.Name)
				assert.Equal(t, "A test tool", config.Agent.Tool.Description)
				assert.Equal(t, "query", config.Agent.Tool.ArgumentName)
				assert.Equal(t, "The query to process", config.Agent.Tool.ArgumentDescription)
				assert.Equal(t, "openai", config.Agent.LLM.Provider)
				assert.Equal(t, "gpt-4", config.Agent.LLM.Model)
				assert.Equal(t, "test-api-key", config.Agent.LLM.APIKey)
				assert.Equal(t, "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}", config.Agent.LLM.PromptTemplate)
				assert.Equal(t, "debug", config.Runtime.Log.RawLevel)
				assert.Equal(t, "./test.log", config.Runtime.Log.RawOutput)
				// After Apply, check parsed fields
				config.Apply(config)
				assert.NotNil(t, config.Runtime.Log.Output)
				assert.Equal(t, logrus.DebugLevel, config.Runtime.Log.LogLevel)
			},
		},
		{
			name:        "Invalid JSON file",
			filePath:    invalidJSONPath,
			expectError: true,
			errorMsg:    "failed to parse JSON configuration",
		},
		{
			name:        "Non-JSON file",
			filePath:    nonJSONPath,
			expectError: true,
			errorMsg:    "file is not a JSON file",
		},
		{
			name:        "Non-existent file",
			filePath:    filepath.Join(tempDir, "non-existent.json"),
			expectError: true,
			errorMsg:    "configuration file does not exist",
		},
		{
			name:        "Empty file path",
			filePath:    "",
			expectError: true,
			errorMsg:    "empty file path provided",
		},
		{
			name: "MCP server timeout parameter",
			filePath: func() string {
				path := filepath.Join(tempDir, "timeout-config.json")
				json := `{
  "runtime": {
    "log": { "level": "info", "output": "stdout" }
  },
  "agent": {
    "name": "timeout-agent",
    "tool": {
      "name": "timeout-tool",
      "description": "Tool with timeout",
      "argument_name": "input",
      "argument_description": "Input"
    },
    "llm": {
      "provider": "openai",
      "model": "gpt-4",
      "api_key": "test-api-key",
      "prompt_template": "Test {{input}}. Tools: {{tools}}"
    },
    "connections": {
      "mcpServers": {
        "slow": {
          "command": "slow-server",
          "timeout": 42
        }
      }
    }
  }
}
`
				_ = os.WriteFile(path, []byte(json), 0644)
				return path
			}(),
			expectError: false,
			validate: func(t *testing.T, config *types.Configuration) {
				config.Apply(config)
				server, ok := config.Agent.Connections.McpServers["slow"]
				assert.True(t, ok)
				assert.Equal(t, 42.0, server.Timeout)
			},
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewJSONLoader(tc.filePath)
			config, err := loader.LoadConfiguration()

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				if tc.validate != nil {
					tc.validate(t, config)
				}
			}
		})
	}
}
