package agent

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/chat"

	"github.com/korchasa/speelka-agent-go/internal/utils"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

type Agent struct {
	config        types.AgentConfig
	llmService    types.LLMServiceSpec
	toolConnector types.ToolConnectorSpec
	logger        types.LoggerSpec
	chat          *chat.Chat // Injected chat instance
}

var exitTool = mcp.NewTool(
	"answer",
	mcp.WithDescription("Use this tool to answer the user and finish the session. The argument 'text' is the final answer."),
	mcp.WithString(
		"text",
		mcp.Description("The final answer to the user's request"),
		mcp.Required(),
	),
)

// NewAgent creates a new instance of Agent with the given dependencies
func NewAgent(
	config types.AgentConfig,
	llmService types.LLMServiceSpec,
	toolConnector types.ToolConnectorSpec,
	logger types.LoggerSpec,
	chat *chat.Chat,
) *Agent {
	return &Agent{
		config:        config,
		llmService:    llmService,
		toolConnector: toolConnector,
		logger:        logger,
		chat:          chat,
	}
}

// GetAllTools returns all available tools (internal and from MCPs)
func (a *Agent) GetAllTools(ctx context.Context) ([]mcp.Tool, error) {
	mcpTools, err := a.toolConnector.GetAllTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP connector: %w", err)
	}
	mcpTools = append(mcpTools, exitTool)
	var toolNames []string
	for _, t := range mcpTools {
		toolNames = append(toolNames, t.Name)
	}
	a.logger.Infof("Tools for LLM: %v", toolNames)
	return mcpTools, nil
}

// isExitCommand checks if a tool call is for the exit tool
func (a *Agent) isExitCommand(call types.CallToolRequest) bool {
	return call.ToolName() == "answer"
}

// runSession manages the main loop of interaction with LLM and tools, returning the final answer and meta information.
// CallDirect now simply calls runSession and returns the result.
func (a *Agent) runSession(ctx context.Context, input string) (string, types.MetaInfo, error) {
	fmt.Fprintf(os.Stderr, "[runSession] started session\n")
	start := time.Now()
	tools, err := a.GetAllTools(ctx)
	if err != nil {
		return "", types.MetaInfo{}, err
	}
	fmt.Fprintf(os.Stderr, "[runSession] got tools:\n")
	session, err := a.beginSession(input, tools)
	if err != nil {
		return "", types.MetaInfo{}, err
	}
	iteration := 0
	var meta types.MetaInfo
	for iteration < a.config.MaxLLMIterations {
		iteration++
		resp, err := a.llmService.SendRequest(ctx, session.GetLLMMessages(), tools)
		if err != nil {
			return "", types.MetaInfo{}, err
		}
		session.AddAssistantMessage(resp)
		if session.ExceededRequestBudget() {
			info := session.GetInfo()
			return "", types.MetaInfo{
				Tokens:     info.TotalTokens,
				Cost:       info.TotalCost,
				DurationMs: time.Since(start).Milliseconds(),
			}, fmt.Errorf("exceeded request budget: total cost %.4f > budget %.4f", info.TotalCost, info.RequestBudget)
		}
		if len(resp.Calls) == 0 {
			return "", types.MetaInfo{}, fmt.Errorf("LLM returned no tool calls")
		}
		for _, call := range resp.Calls {
			if a.isExitCommand(call) {
				finalMessage, _ := call.Params.Arguments["text"].(string)
				info := session.GetInfo()
				meta = types.MetaInfo{
					Tokens:           info.TotalTokens,
					Cost:             info.TotalCost,
					DurationMs:       time.Since(start).Milliseconds(),
					PromptTokens:     resp.Metadata.Tokens.PromptTokens,
					CompletionTokens: resp.Metadata.Tokens.CompletionTokens,
					ReasoningTokens:  resp.Metadata.Tokens.ReasoningTokens,
				}
				return finalMessage, meta, nil
			}
		}
		a.handleLLMToolCallRequest(ctx, resp, session, iteration)
	}
	info := session.GetInfo()
	return "", types.MetaInfo{
		Tokens:     info.TotalTokens,
		Cost:       info.TotalCost,
		DurationMs: time.Since(start).Milliseconds(),
	}, fmt.Errorf("exceeded maximum number of LLM iterations (%d)", a.config.MaxLLMIterations)
}

// CallDirect now simply calls runSession and returns the result.
func (a *Agent) CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error) {
	return a.runSession(ctx, input)
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
		result, err := a.toolConnector.ExecuteTool(ctx, call)
		if err != nil {
			a.logger.Errorf("failed to execute tool %s: %v", call.ToolName(), err)
			errorResult := mcp.NewToolResultError(fmt.Sprintf("Error: %v", err))
			session.AddToolResult(call, errorResult)
			continue
		}
		session.AddToolResult(call, result)
	}

	a.logger.Infof("Iteration complete: %s", utils.SDump(session.GetInfo()))
}

func (a *Agent) HandleLLMAnswerToolRequest(call types.CallToolRequest, resp types.LLMResponse, session *chat.Chat) *mcp.CallToolResult {
	// Robust nil and type checking for 'text' argument
	argValue, exists := call.Params.Arguments["text"]
	if !exists || argValue == nil {
		fmt.Fprintf(os.Stderr, "[HandleLLMAnswerToolRequest] missing or nil 'text' argument in exit tool call\n")
		return mcp.NewToolResultError("missing or nil 'text' argument in exit tool call")
	}
	finalMessage, ok := argValue.(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "[HandleLLMAnswerToolRequest] invalid 'text' argument type: expected string, got %T\n", argValue)
		return mcp.NewToolResultError(fmt.Sprintf("invalid 'text' argument type: expected string, got %T", argValue))
	}
	if finalMessage == "" {
		fmt.Fprintf(os.Stderr, "[HandleLLMAnswerToolRequest] empty 'text' argument in exit tool call\n")
		return mcp.NewToolResultError("empty 'text' argument in exit tool call")
	}
	a.logger.WithFields(logrus.Fields{
		"request_cost":     resp.Metadata.Cost,
		"request_duration": resp.Metadata.DurationMs,
	}).Infof("<< LLM asked to answer the user with: %s", finalMessage)
	fmt.Fprintf(os.Stderr, "[HandleLLMAnswerToolRequest] Chat ended by LLM with message: %s %s\n", finalMessage, utils.SDump(session.GetInfo()))
	return mcp.NewToolResultText(finalMessage)
}
