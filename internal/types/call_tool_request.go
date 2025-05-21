//go:build !exclude_calltoolrequest_string_test
// +build !exclude_calltoolrequest_string_test

package types

import (
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"

	"github.com/mark3labs/mcp-go/mcp"
)

type CallToolRequest struct {
	mcp.CallToolRequest
	llms llms.ToolCall
	ID   string
}

func NewCallToolRequest(call llms.ToolCall) (CallToolRequest, error) {
	// Parse arguments string into map
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(call.FunctionCall.Arguments), &args); err != nil {
		args = make(map[string]interface{})
	}

	req := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      call.FunctionCall.Name,
			Arguments: args,
			Meta:      nil,
		},
	}

	return CallToolRequest{
		llms:            call,
		CallToolRequest: req,
		ID:              call.ID,
	}, nil
}

func (c *CallToolRequest) String() string {
	args := c.Params.Arguments
	var argsStr string
	switch v := args.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			argsStr = "{}"
		} else {
			b, err := json.Marshal(v)
			if err != nil {
				argsStr = "{error}"
			} else {
				argsStr = string(b)
			}
		}
	case nil:
		argsStr = "{}"
	default:
		b, err := json.Marshal(v)
		if err != nil {
			argsStr = "{error}"
		} else {
			argsStr = string(b)
		}
	}
	return fmt.Sprintf("%s(%s)#%s", c.Params.Name, argsStr, c.ID)
}

func (c *CallToolRequest) ToolName() string {
	return c.Params.Name
}

func (c *CallToolRequest) ToLLM() llms.ToolCall {
	return c.llms
}
