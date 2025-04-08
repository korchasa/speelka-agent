// Package types Package contracts defines the interfaces for the MCP server components.
package types

import (
    "context"
    "github.com/mark3labs/mcp-go/mcp"
)

// MCPServerSpec represents the interface for the MCP server.
type MCPServerSpec interface {
    // Start initializes and starts the MCP server.
    // It returns an error if the server fails to start.
    Start(ctx context.Context) error

    // Stop gracefully shuts down the MCP server.
    // It returns an error if the server fails to stop.
    Stop(ctx context.Context) error

    // RegisterTool registers a tool with the MCP server.
    // It returns an error if the tool registration fails.
    RegisterTool(tool mcp.Tool) error

    // GetRegisteredTools returns a list of all registered tools.
    GetRegisteredTools() []mcp.Tool

    // HandleRequest processes a request to a tool.
    // It returns the response from the tool and an error if the request fails.
    HandleRequest(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error)
}

// ParameterSpec represents the specification of a parameter.
type ParameterSpec struct {
    // Type is the data type of the parameter.
    Type string

    // Description is a description of the parameter.
    Description string

    // Required indicates whether the parameter is required.
    Required bool

    // Default is the default value of the parameter if it is not provided.
    Default interface{}

    // Enum is a list of possible values for the parameter.
    Enum []interface{}
}
