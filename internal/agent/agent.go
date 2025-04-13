package agent

import (
	"context"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
)

var ExitTool = mcp.NewTool("answer",
	mcp.WithDescription("Send response to the user"),
	mcp.WithString("text",
		mcp.Required(),
		mcp.Description("Text to send to the user"),
	),
)

type Agent struct {
	config       types.AgentConfig
	llmService   types.LLMServiceSpec
	mcpServer    types.MCPServerSpec
	mcpConnector types.MCPConnectorSpec
	logger       types.LoggerSpec
}

// GetMCPServer returns the MCP server instance for external use
func (a *Agent) GetMCPServer() types.MCPServerSpec {
	return a.mcpServer
}

// NewAgent creates a new instance of Agent with the given dependencies
func NewAgent(
	config types.AgentConfig,
	llmService types.LLMServiceSpec,
	mcpServer types.MCPServerSpec,
	mcpConnector types.MCPConnectorSpec,
	logger types.LoggerSpec,
) types.AgentSpec {
	return &Agent{
		config:       config,
		llmService:   llmService,
		mcpServer:    mcpServer,
		mcpConnector: mcpConnector,
		logger:       logger,
	}
}

// RegisterTools registers all tools for the agent
func (a *Agent) RegisterTools() {
	// Register exit tool
	a.mcpServer.AddTool(ExitTool, nil) // No handler needed as we catch exit tool in process

	// Register agent's core tool for handling user queries
	toolConfig := a.config.Tool
	a.mcpServer.AddTool(
		mcp.NewTool(
			toolConfig.Name,
			mcp.WithDescription(toolConfig.Description),
			mcp.WithString(
				toolConfig.ArgumentName,
				mcp.Description(toolConfig.ArgumentDescription),
				mcp.Required(),
			),
		),
		a.HandleRequest,
	)
}

// GetAllTools returns all available tools (internal and from MCPs)
func (a *Agent) GetAllTools(ctx context.Context) ([]mcp.Tool, error) {
	// Get tools from MCP connector
	mcpTools, err := a.mcpConnector.GetAllTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP connector: %w", err)
	}

	// Combine with exit tool
	allTools := append(mcpTools, ExitTool)

	return allTools, nil
}

// isExitCommand checks if a tool call is for the exit tool
func (a *Agent) isExitCommand(call types.CallToolRequest) bool {
	return call.ToolName() == "answer"
}

// HandleRequest processes the incoming MCP request
func (a *Agent) HandleRequest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a.logger.Debugf(">> HandleRequest: %s", types.SDump(map[string]any{"req": req}))

	toolConfig := a.config.Tool
	if req.Params.Name != toolConfig.Name {
		a.logger.Errorf("invalid tool name: %s", req.Params.Name)
		return mcp.NewToolResultError("invalid tool name"), nil
	}

	// Check if the argument exists and is not nil before type assertion
	argValue, exists := req.Params.Arguments[toolConfig.ArgumentName]
	if !exists || argValue == nil {
		a.logger.Errorf("missing or nil input argument: %s", toolConfig.ArgumentName)
		return mcp.NewToolResultError(fmt.Sprintf("missing or nil input argument: %s", toolConfig.ArgumentName)), nil
	}

	// Safely convert to string
	userRequest, ok := argValue.(string)
	if !ok {
		a.logger.Errorf("invalid input argument type: expected string, got %T", argValue)
		return mcp.NewToolResultError(fmt.Sprintf("invalid input argument type: expected string, got %T", argValue)), nil
	}

	if userRequest == "" {
		a.logger.Errorf("empty input variable")
		return mcp.NewToolResultError("empty input variable"), nil
	}

	a.logger.Infof(">> Request from client: %s", userRequest)

	return a.process(ctx, userRequest)
}

// process processes user requests through the LLM and tool execution
func (a *Agent) process(ctx context.Context, userRequest string) (*mcp.CallToolResult, error) {
	tools, err := a.GetAllTools(ctx)
	if err != nil {
		a.logger.Errorf("failed to get tools: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get tools: %s", err)), nil
	}

	history := chat.NewChat(
		a.config.Model,
		a.config.SystemPromptTemplate,
		a.config.Tool.ArgumentName,
		a.logger,
	)

	// Configure chat compaction settings from configuration
	history.SetMaxTokens(a.config.MaxTokens)
	if err := history.SetCompactionStrategy(a.config.CompactionStrategy); err != nil {
		a.logger.Warnf("Error setting chat compaction strategy: %v. Using default.", err)
	}

	a.logger.Infof("Chat configured with max tokens: %d, compaction strategy: %s",
		a.config.MaxTokens, a.config.CompactionStrategy)

	err = history.Begin(userRequest, tools)
	if err != nil {
		a.logger.Errorf("failed to begin chat: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to begin chat: %s", err)), nil
	}

	var finalMessage string
	iteration := 0

	// Main loop for LLM and tool interaction
	for iteration < a.config.MaxLLMIterations {
		iteration++

		a.logger.WithField("iteration", iteration).Infof(">> Send request to LLM")
		message, calls, err := a.llmService.SendRequest(ctx, history.GetLLMMessages(), tools)
		if err != nil {
			a.logger.Errorf("failed to send request to LLM: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to send request to LLM: %s", err)), nil
		}
		a.logger.Infof("<< LLM response received with %d choices", len(calls))
		a.logger.Debugf("<< Details: %s", types.SDump(map[string]any{"message": message, "calls": calls}))

		for _, call := range calls {
			if a.isExitCommand(call) {
				finalMessage = call.Params.Arguments["text"].(string)
				history.AddAssistantMessage(finalMessage)
				a.logger.Infof("<< LLM response received with final message: %s", finalMessage)
				a.logger.Infof("Chat ended with total tokens: %d", history.GetTotalTokens())
				return mcp.NewToolResultText(finalMessage), nil
			}
		}

		// Process tool calls
		history.AddAssistantMessage(message)

		// If there are no tool calls, assume we're done
		if len(calls) == 0 {
			a.logger.Infof("<< LLM response received with no tool calls, assuming final message: %s", message)
			a.logger.Infof("Chat ended with total tokens: %d", history.GetTotalTokens())
			return mcp.NewToolResultText(message), nil
		}

		// Execute tool calls
		for _, call := range calls {
			a.logger.Infof(">> Process tool call: %s", call.ToolName())
			a.logger.Debugf(">> Details: %s", types.SDump(call))

			// Add tool call to history
			history.AddToolCall(call)

			// Execute the tool
			a.logger.Infof(">> Execute tool `%s` with args: %s", call.ToolName(), types.SDump(call.Params.Arguments))
			result, err := a.mcpConnector.ExecuteTool(ctx, call)
			if err != nil {
				a.logger.Errorf("failed to execute tool %s: %v", call.ToolName(), err)
				errorResult := mcp.NewToolResultError(fmt.Sprintf("Error: %v", err))
				history.AddToolResult(call, errorResult)
				continue
			}

			// Add result to history
			history.AddToolResult(call, result)
			a.logger.Infof("<< Tool %s execution complete", call.ToolName())
		}
	}

	// If we reach here, we've exceeded the maximum number of iterations
	errMsg := fmt.Sprintf("exceeded maximum number of LLM iterations (%d)", a.config.MaxLLMIterations)
	a.logger.Errorf(errMsg)
	a.logger.Infof("Chat ended with total tokens: %d", history.GetTotalTokens())
	return mcp.NewToolResultError(errMsg), nil
}
