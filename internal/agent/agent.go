package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	types2 "github.com/korchasa/speelka-agent-go/internal/llm/types"
	"github.com/korchasa/speelka-agent-go/internal/utils/dump"
	"github.com/mark3labs/mcp-go/client"
	"github.com/tmc/langchaingo/llms"

	"github.com/korchasa/speelka-agent-go/internal/chat"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

type Agent struct {
	config        configuration.AgentConfig
	llmService    llmServiceSpec
	toolConnector toolConnectorSpec
	log           *logrus.Logger
	chat          *chat.Chat // Injected chat instance
}

var finishTool = mcp.NewTool(
	"finish",
	mcp.WithDescription("Use this tool to answer the user and finish the session. The argument 'text' is the final answer."),
	mcp.WithString(
		"text",
		mcp.Description("The final answer to the user's request"),
		mcp.Required(),
	),
)

// calculatorSpec computes monetary cost for LLM usage.
type calculatorSpec interface {
	// CalculateLLMResponse returns the number of tokens, USD cost, and approximation flag for the given model and LLM response.
	CalculateLLMResponse(modelName string, resp types2.LLMResponse) (tokens int, cost float64, isApprox bool, err error)
}

// ToolConnectorSpec represents the interface for the tool connector component.
// Responsibility: Defining the contract for the tool connector
// Features: Defines methods for connecting to tool servers and executing tools
type toolConnectorSpec interface {
	// InitAndConnectToMCPs initializes connections to all configured tool servers.
	// It returns an error if any connection fails.
	InitAndConnectToMCPs(ctx context.Context) error
	// ConnectServer connects to a specific tool server.
	// It returns the client for the server and an error if the connection fails.
	ConnectServer(ctx context.Context, serverID string, serverConfig configuration.MCPServerConnection) (client.MCPClient, error)
	// GetAllTools returns a list of all tools available on all connected tool servers.
	// It returns an error if the tool discovery fails.
	GetAllTools(ctx context.Context) ([]mcp.Tool, error)
	// ExecuteTool executes a tool on the appropriate tool server.
	// It returns the result of the tool execution and an error if the execution fails.
	ExecuteTool(ctx context.Context, call types.CallToolRequest) (*mcp.CallToolResult, error)
	// Close closes all connections to tool servers.
	// It returns an error if any connection fails to close.
	Close() error
}

// llmServiceSpec represents the interface for the LLM service.
// Responsibility: Defining the contract for the LLM service
// Features: Defines methods for sending requests to the LLM
type llmServiceSpec interface {
	// SendRequest sends a request to the LLM with the given prompt and tools.
	// It returns the response struct and an error if the request fails.
	SendRequest(ctx context.Context, messages []llms.MessageContent, tools []mcp.Tool) (types2.LLMResponse, error)
}

// NewAgent creates a new instance of Agent with the given dependencies
func NewAgent(
	config configuration.AgentConfig,
	llmService llmServiceSpec,
	toolConnector toolConnectorSpec,
	log *logrus.Logger,
	chat *chat.Chat,
) *Agent {
	return &Agent{
		config:        config,
		llmService:    llmService,
		toolConnector: toolConnector,
		log:           log,
		chat:          chat,
	}
}

// GetAllTools returns all available tools (internal and from MCPs)
func (a *Agent) GetAllTools(ctx context.Context) ([]mcp.Tool, error) {
	mcpTools, err := a.toolConnector.GetAllTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP connector: %w", err)
	}
	mcpTools = append(mcpTools, finishTool)
	var toolNames []string
	for _, t := range mcpTools {
		toolNames = append(toolNames, t.Name)
	}
	a.log.Infof("Tools for LLM: %v", toolNames)
	return mcpTools, nil
}

// isFinishCommand checks if a tool call is for the finish tool
func (a *Agent) isFinishCommand(call types.CallToolRequest) bool {
	return call.ToolName() == finishTool.Name
}

// RunSession manages the main loop of interaction with LLM and tools, returning the final answer and meta information.
// CallDirect now simply calls RunSession and returns the result.
func (a *Agent) RunSession(ctx context.Context, input string) (string, types.MetaInfo, error) {
	start := time.Now()
	tools, err := a.GetAllTools(ctx)
	if err != nil {
		return "", types.MetaInfo{}, err
	}
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
			if a.isFinishCommand(call) {
				var finalMessage string
				if args, ok := call.Params.Arguments.(map[string]interface{}); ok {
					finalMessage, _ = args["text"].(string)
				}
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

func (a *Agent) beginSession(userRequest string, tools []mcp.Tool) (*chat.Chat, error) {
	// Create a new Chat instance for each session, passing request budget
	var calculator calculatorSpec = nil
	if svc, ok := a.llmService.(interface{ GetCalculator() calculatorSpec }); ok {
		calculator = svc.GetCalculator()
	}
	session := chat.NewChat(
		a.config.Model,
		a.config.SystemPromptTemplate,
		a.config.Tool.ArgumentName,
		a.log,
		calculator,
		a.config.MaxTokens,
		0.0, // No request budget in AgentConfig, use 0.0 (unlimited)
	)
	info := session.GetInfo()
	a.log.Infof("Chat configured with max tokens: %d, request budget: %.4f", info.MaxTokens, info.RequestBudget)

	err := session.Begin(userRequest, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to begin session: %w", err)
	}
	return session, nil
}

func (a *Agent) handleLLMToolCallRequest(ctx context.Context, resp types2.LLMResponse, session *chat.Chat, _ int) {
	var toolCalls []string
	for _, call := range resp.Calls {
		toolCalls = append(toolCalls, call.String())
	}
	a.log.WithFields(logrus.Fields{
		"request_cost":     resp.Metadata.Cost,
		"request_duration": resp.Metadata.DurationMs,
	}).Infof("<< LLM asked to call tools:\n%s", strings.Join(toolCalls, "\n"))
	for _, call := range resp.Calls {
		session.AddToolCall(call)
		result, err := a.toolConnector.ExecuteTool(ctx, call)
		if err != nil {
			a.log.Errorf("failed to execute tool %s: %v", call.ToolName(), err)
			errorResult := mcp.NewToolResultError(fmt.Sprintf("Error: %v", err))
			session.AddToolResult(call, errorResult)
			continue
		}
		session.AddToolResult(call, result)
	}

	a.log.Infof("Iteration complete: %s", dump.SDump(session.GetInfo()))
}

func (a *Agent) HandleLLMFinishToolRequest(call types.CallToolRequest, resp types2.LLMResponse, session *chat.Chat) *mcp.CallToolResult {
	// Robust nil and type checking for 'text' argument
	args, ok := call.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments is not a map in finish tool call")
	}
	argValue, exists := args["text"]
	if !exists || argValue == nil {
		return mcp.NewToolResultError("missing or nil 'text' argument in finish tool call")
	}
	finalMessage, ok := argValue.(string)
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("invalid 'text' argument type: expected string, got %T", argValue))
	}
	if finalMessage == "" {
		return mcp.NewToolResultError("empty 'text' argument in finish tool call")
	}
	a.log.WithFields(logrus.Fields{
		"request_cost":     resp.Metadata.Cost,
		"request_duration": resp.Metadata.DurationMs,
	}).Infof("<< LLM asked to answer the user with: %s", finalMessage)
	return mcp.NewToolResultText(finalMessage)
}
