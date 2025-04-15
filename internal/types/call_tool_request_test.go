package types

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestCallToolRequest_String(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		call := CallToolRequest{
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
		assert.Equal(t, "testTool({})", call.String())
	})

	t.Run("one argument", func(t *testing.T) {
		call := CallToolRequest{
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
		assert.Equal(t, "testTool({\"foo\":\"bar\"})", call.String())
	})

	t.Run("multiple arguments", func(t *testing.T) {
		call := CallToolRequest{
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
		ok := out == "testTool({\"foo\":\"bar\",\"num\":42})" || out == "testTool({\"num\":42,\"foo\":\"bar\"})"
		assert.True(t, ok, "got: %s", out)
	})
}
