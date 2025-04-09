package agent

import (
	"context"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_service"
	"github.com/korchasa/speelka-agent-go/internal/mcp_connector"
	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/korchasa/speelka-agent-go/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

// MaxLLMIterations Maximum number of LLM interaction iterations
const MaxLLMIterations = 25

var ExitTool = mcp.NewTool("answer",
	mcp.WithDescription("Send response to the user"),
	mcp.WithString("text",
		mcp.Required(),
		mcp.Description("Text to send to the user"),
	),
)

type Agent struct {
	configManager types.ConfigurationManagerSpec
	llmService    *llm_service.LLMService
	mcpServer     *mcp_server.MCPServer
	mcpConnector  *mcp_connector.MCPConnector
	logger        *logrus.Logger
}

func NewAgent(configManager types.ConfigurationManagerSpec, logger *logrus.Logger) (*Agent, error) {
	llmService, err := llm_service.NewLLMService(configManager.GetLLMConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}
	logger.Info("LLM service instance created")

	mcpServer := mcp_server.NewMCPServer(configManager.GetMCPServerConfig(), logger)
	logger.Info("MCP server instance created")

	mcpConnector := mcp_connector.NewMCPConnector(configManager.GetMCPConnectorConfig(), logger)
	logger.Info("MCP connector instance created")

	return &Agent{
		configManager: configManager,
		llmService:    llmService,
		mcpServer:     mcpServer,
		mcpConnector:  mcpConnector,
		logger:        logger,
	}, nil
}

func (a *Agent) Start(daemonMode bool, ctx context.Context) error {
	err := a.mcpConnector.InitAndConnectToMCPs(ctx)
	if err != nil {
		return fmt.Errorf("failed to init MCP connector: %w", err)
	}
	a.logger.Info("MCP connector connected successfully")

	if daemonMode {
		a.logger.Info("Running in daemon mode with HTTP SSE MCP server")
		if err := a.mcpServer.ServeDaemon(a.HandleRequest); err != nil {
			return fmt.Errorf("failed to start HTTP MCP server: %w", err)
		}
	} else {
		a.logger.Info("Running in script mode with stdio MCP server")
		if err := a.mcpServer.ServeStdio(a.HandleRequest); err != nil {
			return fmt.Errorf("failed to start Stdio MCP Server: %w", err)
		}
	}
	return nil
}

func (a *Agent) Stop(shutdownCtx context.Context) error {
	if err := a.mcpServer.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop HTTP MCP server: %w", err)
	} else {
		a.logger.Info("HTTP MCP server stopped successfully")
	}
	a.logger.Info("Server shutdown complete")
	return nil
}

// isExitCommand checks if the tool call is a special exit command
// Responsibility: Determining when to end the LLM interaction cycle
// Features: Compares the tool name with the configured exit tool name
func (a *Agent) isExitCommand(call types.CallToolRequest) bool {
	return call.ToolName() == ExitTool.Name
}

// HandleRequest processes the request to the main tool
// Responsibility: Validating and executing tool requests
// Features: Checks for tool availability and parameter correctness before execution
func (a *Agent) HandleRequest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a.logger.Debugf(">> HandleRequest: %s", utils.SDump(map[string]any{"req": req}))

	toolConfig := a.configManager.GetMCPServerConfig().Tool
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

	tools, err := a.mcpConnector.GetAllTools(ctx)
	if err != nil {
		a.logger.Errorf("failed to get tools: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get tools: %s", err)), nil
	}
	tools = append(tools, ExitTool)

	history := chat.NewChat(
		a.configManager.GetLLMConfig().SystemPromptTemplate,
		toolConfig.ArgumentName,
		a.logger,
	)
	err = history.Begin(userRequest, tools)
	if err != nil {
		a.logger.Errorf("failed to begin chat: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to begin chat: %s", err)), nil
	}

	var finalMessage string
	iteration := 0

	// Main loop for LLM and tool interaction
	for iteration < MaxLLMIterations {
		iteration++

		a.logger.WithField("iteration", iteration).Infof(">> Send request to LLM")
		message, calls, err := a.llmService.SendRequest(ctx, history.GetLLMMessages(), tools)
		if err != nil {
			a.logger.Errorf("failed to send request to LLM: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to send request to LLM: %s", err)), nil
		}
		a.logger.Infof("<< LLM response received with %d choices", len(calls))
		a.logger.Debugf("<< Details: %s", utils.SDump(map[string]any{"message": message, "calls": calls}))

		for _, call := range calls {
			if a.isExitCommand(call) {
				finalMessage = call.Params.Arguments["text"].(string)
				history.AddAssistantMessage(finalMessage)
				a.logger.Infof("<< LLM response received with final message: %s", finalMessage)
				return mcp.NewToolResultText(finalMessage), nil
			}
		}

		// Process tool calls
		for _, call := range calls {
			a.logger.Infof(">> Process tool call: %s", call.ToolName())
			a.logger.Debugf(">> Details: %s", utils.SDump(call))

			// If the tool schema has no properties but has arguments,
			// we need to check if this is a valid call
			for _, tool := range tools {
				if tool.Name == call.ToolName() {
					// If the tool doesn't require arguments (inputSchema == null or properties == null),
					// but arguments are provided, log a warning but continue execution
					if (tool.InputSchema.Type == "" || tool.InputSchema.Properties == nil || len(tool.InputSchema.Properties) == 0) &&
						call.Params.Arguments != nil && len(call.Params.Arguments) > 0 {
						a.logger.Warnf("Tool %s called with arguments but doesn't require any: %v",
							tool.Name, call.Params.Arguments)
					}
					break
				}
			}

			// Add tool call to history
			history.AddToolCall(call)

			// Execute the tool
			a.logger.Infof(">> Execute tool `%s` in connector with args: %s", call.ToolName(), utils.SDump(call.Params.Arguments))
			result, err := a.mcpConnector.ExecuteTool(ctx, call)
			if err != nil {
				a.logger.Warnf("can't make a call in connector: %v", err)
				errStr := err.Error()
				result = mcp.NewToolResultError(errStr)
			}
			a.logger.Infof("<< Tool call `%s` success", call.ToolName())
			a.logger.Debugf("<< Details: %s", utils.SDump(map[string]any{"result": result}))

			// Add result to history
			history.AddToolResult(call, result)
		}
	}

	a.logger.Warnf("<< Reached maximum number of LLM iterations (%d)", MaxLLMIterations)
	return mcp.NewToolResultError(fmt.Sprintf("<< Reached maximum number of iterations (%d). Last response: %s", MaxLLMIterations, finalMessage)), nil
}
