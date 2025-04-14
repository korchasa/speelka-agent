package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLLoader_LoadConfiguration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yaml-loader-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid YAML configuration file
	validYAML := `
runtime:
  log:
    level: debug
    output: ./test.log

agent:
  name: test-agent
  tool:
    name: test-tool
    description: A test tool
    argument_name: query
    argument_description: The query to process
  llm:
    provider: openai
    model: gpt-4
    api_key: test-api-key
    prompt_template: "You are a helpful assistant. User query: {{query}} Available tools: {{tools}}"
`

	validYAMLPath := filepath.Join(tempDir, "valid-config.yaml")
	err = os.WriteFile(validYAMLPath, []byte(validYAML), 0644)
	require.NoError(t, err)

	// Create a valid YAML configuration file with .yml extension
	validYMLPath := filepath.Join(tempDir, "valid-config.yml")
	err = os.WriteFile(validYMLPath, []byte(validYAML), 0644)
	require.NoError(t, err)

	// Create an invalid YAML file
	invalidYAML := `
runtime
  log:
    level: debug
    INVALID: YAML: SYNTAX
`

	invalidYAMLPath := filepath.Join(tempDir, "invalid-config.yaml")
	err = os.WriteFile(invalidYAMLPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	// Create a non-YAML file
	nonYAMLPath := filepath.Join(tempDir, "config.txt")
	err = os.WriteFile(nonYAMLPath, []byte("This is not a YAML file"), 0644)
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
			name:        "Valid YAML file with .yaml extension",
			filePath:    validYAMLPath,
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
				assert.Equal(t, "./test.log", config.Runtime.Log.Output)
			},
		},
		{
			name:        "Valid YAML file with .yml extension",
			filePath:    validYMLPath,
			expectError: false,
			validate: func(t *testing.T, config *types.Configuration) {
				assert.Equal(t, "test-agent", config.Agent.Name)
				assert.Equal(t, "test-tool", config.Agent.Tool.Name)
			},
		},
		{
			name:        "Invalid YAML file",
			filePath:    invalidYAMLPath,
			expectError: true,
			errorMsg:    "failed to parse YAML configuration",
		},
		{
			name:        "Non-YAML file",
			filePath:    nonYAMLPath,
			expectError: true,
			errorMsg:    "file is not a YAML file",
		},
		{
			name:        "Non-existent file",
			filePath:    filepath.Join(tempDir, "non-existent.yaml"),
			expectError: true,
			errorMsg:    "configuration file does not exist",
		},
		{
			name:        "Empty file path",
			filePath:    "",
			expectError: true,
			errorMsg:    "empty file path provided",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewYAMLLoader(tc.filePath)
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
