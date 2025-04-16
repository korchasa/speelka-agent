package types

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestCallToolRequest_String(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		call := CallToolRequest{
			ID: "123",
			CallToolRequest: mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name:      "testTool",
					Arguments: nil,
				},
			},
		}
		assert.Equal(t, "testTool({})#123", call.String())
	})

	t.Run("one argument", func(t *testing.T) {
		call := CallToolRequest{
			ID: "456",
			CallToolRequest: mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name:      "testTool",
					Arguments: map[string]interface{}{"foo": "bar"},
				},
			},
		}
		assert.Equal(t, "testTool({\"foo\":\"bar\"})#456", call.String())
	})

	t.Run("multiple arguments", func(t *testing.T) {
		call := CallToolRequest{
			ID: "789",
			CallToolRequest: mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Name:      "testTool",
					Arguments: map[string]interface{}{"foo": "bar", "num": 42},
				},
			},
		}
		// JSON object order is not guaranteed, so check both possibilities
		out := call.String()
		ok := out == "testTool({\"foo\":\"bar\",\"num\":42})#789" || out == "testTool({\"num\":42,\"foo\":\"bar\"})#789"
		assert.True(t, ok, "got: %s", out)
	})
}
