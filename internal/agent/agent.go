package agent

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/korchasa/speelka-agent-go/internal/chat"

    "github.com/korchasa/speelka-agent-go/internal/utils"

    "github.com/korchasa/speelka-agent-go/internal/types"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/sirupsen/logrus"
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
    chat         *chat.Chat // Injected chat instance
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
    chat *chat.Chat,
) types.AgentSpec {
    return &Agent{
        config:       config,
        llmService:   llmService,
        mcpServer:    mcpServer,
        mcpConnector: mcpConnector,
        logger:       logger,
        chat:         chat,
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
    a.logger.Debugf("> HandleRequest: %s", utils.SDump(map[string]any{"req": req}))

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

    a.logger.Infof("> Request from client: %s", userRequest)

    res, err := a.process(ctx, userRequest)
    if err != nil {
        a.logger.Error(err.Error())
        return mcp.NewToolResultError(err.Error()), nil
    }
    return res, nil
}

// process processes user requests through the LLM and tool execution
func (a *Agent) process(ctx context.Context, userRequest string) (*mcp.CallToolResult, error) {
    tools, err := a.GetAllTools(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get tools: %w", err)
    }

    session, err := a.beginSession(userRequest, tools)
    if err != nil {
        return nil, fmt.Errorf("failed to begin session: %w", err)
    }

    iteration := 0
    // Main loop for LLM and tool interaction
    for iteration < a.config.MaxLLMIterations {
        iteration++

        a.logger.WithFields(logrus.Fields{
            "messages": len(session.GetLLMMessages()),
            "tools":    len(tools),
        }).Infof(">> Start iteration %d: sending request to LLM", iteration)

        resp, err := a.llmService.SendRequest(ctx, session.GetLLMMessages(), tools)
        if err != nil {
            return nil, fmt.Errorf("failed to send request to LLM: %w", err)
        }
        session.AddAssistantMessage(resp)

        // Enforce request budget after each LLM response
        if session.ExceededRequestBudget() {
            return a.handleRequestBudgetExceeded(session)
        }

        if len(resp.Calls) == 0 {
            return nil, fmt.Errorf("LLM returned no tool calls")
        }

        for _, call := range resp.Calls {
            if a.isExitCommand(call) {
                return a.handleLLMAnswerToolRequest(call, resp, session), nil
            }
        }
        a.handleLLMToolCallRequest(ctx, resp, session, iteration)
    }

    return a.handleIterationLimit(session)
}

// handleRequestBudgetExceeded returns a tool result error when the request budget is exceeded
func (a *Agent) handleRequestBudgetExceeded(session *chat.Chat) (*mcp.CallToolResult, error) {
    info := session.GetInfo()
    errMsg := fmt.Sprintf("exceeded request budget: total cost %.4f > budget %.4f", info.TotalCost, info.RequestBudget)
    a.logger.Errorf(errMsg)
    a.logger.Infof("Chat ended by exceeding request budget: %s", utils.SDump(info))
    return mcp.NewToolResultError(errMsg), nil
}

func (a *Agent) beginSession(userRequest string, tools []mcp.Tool) (*chat.Chat, error) {
    // Create a new Chat instance for each session, passing request budget
    var calculator types.CalculatorSpec = nil
    if svc, ok := a.llmService.(interface{ GetCalculator() types.CalculatorSpec }); ok {
        calculator = svc.GetCalculator()
    }
    session := chat.NewChat(
        a.config.Model,
        a.config.SystemPromptTemplate,
        a.config.Tool.ArgumentName,
        a.logger,
        calculator,
        a.config.MaxTokens,
        0.0, // No request budget in AgentConfig, use 0.0 (unlimited)
    )
    info := session.GetInfo()
    a.logger.Infof("Chat configured with max tokens: %d, request budget: %.4f", info.MaxTokens, info.RequestBudget)

    err := session.Begin(userRequest, tools)
    if err != nil {
        return nil, fmt.Errorf("failed to begin session: %w", err)
    }
    return session, nil
}

func (a *Agent) handleLLMToolCallRequest(ctx context.Context, resp types.LLMResponse, session *chat.Chat, _ int) {
    var toolCalls []string
    for _, call := range resp.Calls {
        toolCalls = append(toolCalls, call.String())
    }
    a.logger.WithFields(logrus.Fields{
        "request_cost":     resp.Metadata.Cost,
        "request_duration": resp.Metadata.DurationMs,
    }).Infof("<< LLM asked to call tools:\n%s", strings.Join(toolCalls, "\n"))
    for _, call := range resp.Calls {
        session.AddToolCall(call)
		
        a.logger.Infof(">>> Execute tool `%s`", call.String())
        a.logger.Debugf(">>> Details: %s", call.Params.Arguments)
        n := time.Now()
        result, err := a.mcpConnector.ExecuteTool(ctx, call)
        if err != nil {
            a.logger.Errorf("failed to execute tool %s: %v", call.ToolName(), err)
            errorResult := mcp.NewToolResultError(fmt.Sprintf("Error: %v", err))
            session.AddToolResult(call, errorResult)
            continue
        }
        session.AddToolResult(call, result)
        duration := time.Since(n)
        a.logger.Infof("<<< Tool execution complete in %s", duration)
    }

    a.logger.Infof("Iteration complete: %s", utils.SDump(session.GetInfo()))
}

func (a *Agent) handleIterationLimit(session *chat.Chat) (*mcp.CallToolResult, error) {
    // If we reach here, we've exceeded the maximum number of iterations
    errMsg := fmt.Sprintf("exceeded maximum number of LLM iterations (%d)", a.config.MaxLLMIterations)
    a.logger.Errorf(errMsg)
    a.logger.Infof("Chat ended by exceeding max iterations: %s", utils.SDump(session.GetInfo()))
    return mcp.NewToolResultError(errMsg), nil
}

func (a *Agent) handleLLMAnswerToolRequest(call types.CallToolRequest, resp types.LLMResponse, session *chat.Chat) *mcp.CallToolResult {
    finalMessage := call.Params.Arguments["text"].(string)
    a.logger.WithFields(logrus.Fields{
        "request_cost":     resp.Metadata.Cost,
        "request_duration": resp.Metadata.DurationMs,
    }).Infof("<< LLM asked to answer the user with: %s", finalMessage)
    a.logger.Infof("Chat ended by LLM with message: %s %s", finalMessage, utils.SDump(session.GetInfo()))
    return mcp.NewToolResultText(finalMessage)
}
