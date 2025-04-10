package types

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToolToLLM(t *testing.T) {
	t.Run("with InputSchema", func(t *testing.T) {
		// Setup test tool with InputSchema
		inputTool := mcp.Tool{
			Name:        "test_tool",
			Description: "A test tool",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"param1": map[string]interface{}{
						"type":        "string",
						"description": "Parameter 1",
					},
					"param2": map[string]interface{}{
						"type":        "integer",
						"description": "Parameter 2",
					},
				},
				Required: []string{"param1"},
			},
		}

		// Convert to LLM tool
		result, err := ConvertToolToLLM(inputTool)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "function", result.Type)
		assert.Equal(t, "test_tool", result.Function.Name)
		assert.Equal(t, "A test tool", result.Function.Description)

		// Verify parameters converted correctly
		params, ok := result.Function.Parameters.(map[string]interface{})
		require.True(t, ok, "Parameters should be a map[string]interface{}")

		assert.Equal(t, "object", params["type"])

		props, ok := params["properties"].(map[string]interface{})
		require.True(t, ok, "Properties should be a map[string]interface{}")

		param1, ok := props["param1"].(map[string]interface{})
		require.True(t, ok, "param1 should be a map[string]interface{}")
		assert.Equal(t, "string", param1["type"])
		assert.Equal(t, "Parameter 1", param1["description"])

		param2, ok := props["param2"].(map[string]interface{})
		require.True(t, ok, "param2 should be a map[string]interface{}")
		assert.Equal(t, "integer", param2["type"])
		assert.Equal(t, "Parameter 2", param2["description"])

		required, ok := params["required"].([]interface{})
		require.True(t, ok, "Required should be a []interface{}")
		assert.Contains(t, required, "param1")
	})

	t.Run("with RawInputSchema", func(t *testing.T) {
		// Setup test tool with RawInputSchema
		rawSchema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"rawParam1": {
					"type": "string",
					"description": "Raw Parameter 1"
				},
				"rawParam2": {
					"type": "boolean",
					"description": "Raw Parameter 2"
				}
			},
			"required": ["rawParam2"]
		}`)

		inputTool := mcp.Tool{
			Name:           "raw_tool",
			Description:    "A tool with raw schema",
			RawInputSchema: rawSchema,
		}

		// Convert to LLM tool
		result, err := ConvertToolToLLM(inputTool)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "function", result.Type)
		assert.Equal(t, "raw_tool", result.Function.Name)
		assert.Equal(t, "A tool with raw schema", result.Function.Description)

		// Verify parameters converted correctly
		params, ok := result.Function.Parameters.(map[string]interface{})
		require.True(t, ok, "Parameters should be a map[string]interface{}")

		assert.Equal(t, "object", params["type"])

		props, ok := params["properties"].(map[string]interface{})
		require.True(t, ok, "Properties should be a map[string]interface{}")

		rawParam1, ok := props["rawParam1"].(map[string]interface{})
		require.True(t, ok, "rawParam1 should be a map[string]interface{}")
		assert.Equal(t, "string", rawParam1["type"])
		assert.Equal(t, "Raw Parameter 1", rawParam1["description"])

		rawParam2, ok := props["rawParam2"].(map[string]interface{})
		require.True(t, ok, "rawParam2 should be a map[string]interface{}")
		assert.Equal(t, "boolean", rawParam2["type"])
		assert.Equal(t, "Raw Parameter 2", rawParam2["description"])

		required, ok := params["required"].([]interface{})
		require.True(t, ok, "Required should be a []interface{}")
		assert.Contains(t, required, "rawParam2")
	})

	t.Run("with invalid RawInputSchema", func(t *testing.T) {
		// Setup test tool with invalid RawInputSchema
		invalidRawSchema := json.RawMessage(`{invalid json`)

		inputTool := mcp.Tool{
			Name:           "invalid_tool",
			Description:    "A tool with invalid raw schema",
			RawInputSchema: invalidRawSchema,
		}

		// Convert to LLM tool
		_, err := ConvertToolToLLM(inputTool)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("with empty tool", func(t *testing.T) {
		// Setup empty tool
		emptyTool := mcp.Tool{
			Name:        "empty_tool",
			Description: "An empty tool",
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		}

		// Convert to LLM tool
		result, err := ConvertToolToLLM(emptyTool)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "function", result.Type)
		assert.Equal(t, "empty_tool", result.Function.Name)
		assert.Equal(t, "An empty tool", result.Function.Description)

		// Verify parameters converted correctly
		params, ok := result.Function.Parameters.(map[string]interface{})
		require.True(t, ok, "Parameters should be a map[string]interface{}")

		assert.Equal(t, "object", params["type"])

		props, ok := params["properties"].(map[string]interface{})
		require.True(t, ok, "Properties should be a map[string]interface{}")
		assert.Empty(t, props)
	})

	t.Run("with tool without properties", func(t *testing.T) {
		// Tool with Properties = nil
		toolWithoutProps := mcp.Tool{
			Name:        "no_props_tool",
			Description: "A tool without properties",
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: nil,
			},
		}

		// Convert to LLM tool
		result, err := ConvertToolToLLM(toolWithoutProps)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "function", result.Type)
		assert.Equal(t, "no_props_tool", result.Function.Name)
		assert.Equal(t, "A tool without properties", result.Function.Description)

		// Verify parameters converted correctly
		params, ok := result.Function.Parameters.(map[string]interface{})
		require.True(t, ok, "Parameters should be a map[string]interface{}")

		assert.Equal(t, "object", params["type"])

		props, ok := params["properties"].(map[string]interface{})
		require.True(t, ok, "Properties should be a map[string]interface{}")
		assert.Empty(t, props)
	})

	t.Run("with tool without InputSchema", func(t *testing.T) {
		// Tool completely without InputSchema
		toolWithoutSchema := mcp.Tool{
			Name:        "no_schema_tool",
			Description: "A tool without input schema",
		}

		// Convert to LLM tool
		result, err := ConvertToolToLLM(toolWithoutSchema)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "function", result.Type)
		assert.Equal(t, "no_schema_tool", result.Function.Name)
		assert.Equal(t, "A tool without input schema", result.Function.Description)

		// Verify parameters converted correctly - should be an empty object
		params, ok := result.Function.Parameters.(map[string]interface{})
		require.True(t, ok, "Parameters should be a map[string]interface{}")

		// In this case we should get either an empty InputSchema or a schema with empty Properties
		assert.Contains(t, []string{"", "object"}, params["type"])
		if params["type"] == "object" {
			props, propsOk := params["properties"].(map[string]interface{})
			if propsOk {
				assert.Empty(t, props)
			}
		}
	})
}

func TestConvertToolsToLLM(t *testing.T) {
	t.Run("convert multiple tools", func(t *testing.T) {
		// Setup multiple tools
		tools := []mcp.Tool{
			{
				Name:        "tool1",
				Description: "Tool 1",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"param": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			{
				Name:        "tool2",
				Description: "Tool 2",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"param": map[string]interface{}{
							"type": "integer",
						},
					},
				},
			},
		}

		// Convert to LLM tools
		results, err := ConvertToolsToLLM(tools)

		// Assert
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// Check first tool
		assert.Equal(t, "tool1", results[0].Function.Name)
		assert.Equal(t, "Tool 1", results[0].Function.Description)

		// Check second tool
		assert.Equal(t, "tool2", results[1].Function.Name)
		assert.Equal(t, "Tool 2", results[1].Function.Description)
	})

	t.Run("error in one tool", func(t *testing.T) {
		// Setup tools with one invalid
		tools := []mcp.Tool{
			{
				Name:        "valid_tool",
				Description: "Valid Tool",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			{
				Name:           "invalid_tool",
				Description:    "Invalid Tool",
				RawInputSchema: json.RawMessage(`{invalid`),
			},
		}

		// Convert to LLM tools
		results, err := ConvertToolsToLLM(tools)

		// Assert
		require.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to convert tool to LLM tool")
	})

	t.Run("empty tools list", func(t *testing.T) {
		// Setup empty tools list
		var tools []mcp.Tool

		// Convert to LLM tools
		results, err := ConvertToolsToLLM(tools)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
