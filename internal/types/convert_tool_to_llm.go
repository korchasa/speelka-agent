package types

import (
	"encoding/json"

	"github.com/korchasa/speelka-agent-go/internal/error_handling"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
)

// ConvertToolToLLM converts a mcp.Tool to a llms.Tool
// It handles both regular tools with InputSchema and tools with RawInputSchema
// and properly converts property definitions between the two formats.
func ConvertToolToLLM(tool mcp.Tool) (llms.Tool, error) {
	// Create base function definition
	function := &llms.FunctionDefinition{
		Name:        tool.Name,
		Description: tool.Description,
	}

	// Handle schema conversion - need to convert either InputSchema or RawInputSchema to Parameters
	var parametersMap map[string]interface{}

	// Check if we have a RawInputSchema
	if len(tool.RawInputSchema) > 0 {
		// Parse the raw schema directly
		if err := json.Unmarshal(tool.RawInputSchema, &parametersMap); err != nil {
			return llms.Tool{}, err
		}
	} else if tool.InputSchema.Type == "" {
		// Case when InputSchema is completely empty or not initialized
		// Create an empty object as parameters
		parametersMap = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	} else {
		// Check if there are properties in the input schema
		if tool.InputSchema.Properties == nil {
			// If there are no properties, this is a tool without arguments
			// Create an empty object as parameters
			parametersMap = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		} else {
			// Convert InputSchema to a map
			inputSchemaBytes, err := json.Marshal(tool.InputSchema)
			if err != nil {
				return llms.Tool{}, err
			}

			if err := json.Unmarshal(inputSchemaBytes, &parametersMap); err != nil {
				return llms.Tool{}, err
			}
		}
	}

	// Set the parameters
	function.Parameters = parametersMap

	// Create and return the llms.Tool
	toolLLM := llms.Tool{
		Type:     "function",
		Function: function,
	}

	return toolLLM, nil
}

func ConvertToolsToLLM(tools []mcp.Tool) ([]llms.Tool, error) {
	llmTools := make([]llms.Tool, 0)
	for _, tool := range tools {
		t, err := ConvertToolToLLM(tool)
		if err != nil {
			return nil, error_handling.WrapError(
				err,
				"failed to convert tool to LLM tool",
				error_handling.ErrorCategoryInternal,
			)
		}
		llmTools = append(llmTools, t)
	}
	return llmTools, nil
}
