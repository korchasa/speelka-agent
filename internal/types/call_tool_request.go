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
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      call.FunctionCall.Name,
			Arguments: args,
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
	if len(args) == 0 {
		argsStr = "{}"
	} else {
		b, err := json.Marshal(args)
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
