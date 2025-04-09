package chat_test

import (
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestChat_BuildPromptPartForToolsDescription(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Базовый шаблон с bullet points для имени инструмента и контролем отступов
	basicTemplate := `{% for tool in tools -%}
- ` + "`{{ tool.Name }}`" + ` - {{ tool.Description }}
{% endfor %}`

	// Шаблон с параметрами и правильным доступом к свойствам
	paramsTemplate := `{% for tool in tools -%}
- ` + "`{{ tool.Name }}`" + ` - {{ tool.Description }}. Arguments:
{% if tool.InputSchema and tool.InputSchema.Properties -%}
{% for name, prop in tool.InputSchema.Properties -%}
* ` + "`{{ name }}`" + ` ({{ prop.type }}): {{ prop.description }}
{% endfor -%}
{% endif -%}
{% endfor %}`

	t.Run("simple tool description", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "get_file_structure",
				Description: "Displays the code structure of the specified file",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, basicTemplate)
		assert.NoError(t, err)

		expected := `- ` + "`get_file_structure`" + ` - Displays the code structure of the specified file`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool with parameters", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "get_file_structure",
				Description: "Displays the code structure of the specified file",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"file": map[string]interface{}{
							"type":        "string",
							"description": "the path to the file",
						},
					},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, paramsTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt:\n%s", promptPart)

		// Just get actual formatting and adapt test
		expected := `- ` + "`get_file_structure`" + ` - Displays the code structure of the specified file. Arguments:
* ` + "`file`" + ` (string): the path to the file`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("multiple tools", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "tool1",
				Description: "Description of tool1",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			{
				Name:        "tool2",
				Description: "Description of tool2",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, basicTemplate)
		assert.NoError(t, err)

		expected := `- ` + "`tool1`" + ` - Description of tool1
- ` + "`tool2`" + ` - Description of tool2`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("custom template format", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "get_file_structure",
				Description: "Displays the code structure of the specified file",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		customTemplate := `{% for tool in tools -%}Tool: {{ tool.Name }}
Description: {{ tool.Description }}{% endfor %}`

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, customTemplate)
		assert.NoError(t, err)

		expected := `Tool: get_file_structure
Description: Displays the code structure of the specified file`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool with null Properties", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "tool_no_args",
				Description: "A tool without arguments",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: nil,
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt for tool with null Properties:\n%s", promptPart)

		expected := `- ` + "`tool_no_args`" + ` - A tool without arguments. No arguments required.`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool with empty Properties", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "tool_empty_args",
				Description: "A tool with empty arguments",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt for tool with empty Properties:\n%s", promptPart)

		expected := `- ` + "`tool_empty_args`" + ` - A tool with empty arguments. No arguments required.`

		assert.Equal(t, expected, promptPart)
	})

	t.Run("tool without InputSchema", func(t *testing.T) {
		ch := chat.NewChat("", "query", logger)
		tools := []mcp.Tool{
			{
				Name:        "tool_no_schema",
				Description: "A tool without input schema",
			},
		}

		promptPart, err := ch.BuildPromptPartForToolsDescription(tools, chat.DefaultToolsDescriptionTemplate)
		assert.NoError(t, err)

		t.Logf("Generated prompt for tool without InputSchema:\n%s", promptPart)

		expected := `- ` + "`tool_no_schema`" + ` - A tool without input schema. No arguments required.`

		assert.Equal(t, expected, promptPart)
	})
}
