package app_mcp

import (
	"context"
	"fmt"
	"os"

	"github.com/korchasa/speelka-agent-go/internal/agent"
	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/llm_service"
	"github.com/korchasa/speelka-agent-go/internal/mcp_connector"
	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
)

// App is responsible for instantiating and managing the Agent and its dependencies
// (for server/daemon mode)
type App struct {
	configManager types.ConfigurationManagerSpec
	agent         types.AgentSpec
	mcpServer     *mcp_server.MCPServer
	logger        types.LoggerSpec
}

// NewApp creates a new instance of App with the given logger and configuration manager
func NewApp(logger types.LoggerSpec, configManager types.ConfigurationManagerSpec) (*App, error) {
	return &App{
		logger:        logger,
		configManager: configManager,
	}, nil
}

// Initialize creates and initializes all components needed by the Agent
func (a *App) Initialize(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "[Initialize] initialization started\n")
	// Use utility to create agent and dependencies
	agent, mcpServer, err := NewAgentServerMode(a.configManager, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize agent and server: %w", err)
	}
	a.agent = agent
	a.mcpServer = mcpServer
	a.mcpServer.SetMainToolHandler(a.DispatchMCPCall)
	return nil
}

// Start starts the Agent in daemon or stdio mode
func (a *App) Start(daemonMode bool, ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "[Start] daemonMode: %v\n", daemonMode)
	if err := a.mcpServer.Serve(ctx, daemonMode, a.DispatchMCPCall); err != nil {
		return fmt.Errorf("failed to serve mcp server: %w", err)
	}
	return nil
}

// Stop stops the Agent
func (a *App) Stop(shutdownCtx context.Context) error {
	if err := a.mcpServer.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop HTTP MCP server: %w", err)
	}
	a.logger.Info("Server shutdown complete")
	return nil
}

// DirectAgent returns the agent as a directAgent interface for direct call mode
func (a *App) DirectAgent() directAgent {
	return a.agent.(directAgent)
}

type directAgent interface {
	CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error)
}

// HandleCall executes the direct call on the initialized agent, matching DirectApp's interface.
func (a *App) HandleCall(ctx context.Context, input string) types.DirectCallResult {
	logHandleCallStep("incoming request", input)
	if a.agent == nil {
		return buildDirectCallResult("", types.MetaInfo{}, fmt.Errorf("agent not initialized"))
	}
	logHandleCallStep("agent initialized", "")
	da, ok := a.agent.(directAgent)
	if !ok {
		logHandleCallStep("agent does not implement directAgent interface", "")
		return buildDirectCallResult("", types.MetaInfo{}, fmt.Errorf("agent does not implement directAgent interface"))
	}
	logHandleCallStep("agent implements directAgent interface", "")
	answer, meta, err := da.CallDirect(ctx, input)
	logHandleCallStep("answer", answer)
	logHandleCallStep("meta", fmt.Sprintf("%+v", meta))
	logHandleCallStep("error", fmt.Sprintf("%v", err))
	return buildDirectCallResult(answer, meta, err)
}

// Application is the shared interface for all application types.
type Application interface {
	Initialize(ctx context.Context) error
	Start(daemonMode bool, ctx context.Context) error
	Stop(shutdownCtx context.Context) error
	HandleCall(ctx context.Context, input string) types.DirectCallResult
}

// Обновить сигнатуру DispatchMCPCall
func (a *App) DispatchMCPCall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logHandleCallStep("handler called", "")
	logHandleCallStep("incoming request", fmt.Sprintf("%+v", req))
	toolName := req.Params.Name
	argName := a.configManager.GetConfiguration().GetAgentConfig().Tool.ArgumentName
	if err := validateToolName(toolName, a.configManager); err != nil {
		logHandleCallStep("invalid tool name", toolName)
		return mcp.NewToolResultError(err.Error()), nil
	}
	userInput, err := extractUserInput(req.Params.Arguments, argName)
	if err != nil {
		logHandleCallStep("argument error", err.Error())
		return mcp.NewToolResultError(err.Error()), nil
	}
	logHandleCallStep("calling agent core with input", userInput)
	answer, _, err := a.agent.CallDirect(ctx, userInput)
	if err != nil {
		logHandleCallStep("core error", err.Error())
		return mcp.NewToolResultError(err.Error()), nil
	}
	logHandleCallStep("successful answer", answer)
	return mcp.NewToolResultText(answer), nil
}

// --- Приватные функции ---

func validateToolName(toolName string, config types.ConfigurationManagerSpec) error {
	expected := config.GetConfiguration().GetAgentConfig().Tool.Name
	if toolName != expected {
		return fmt.Errorf("invalid tool name: %s", toolName)
	}
	return nil
}

func extractUserInput(arguments map[string]interface{}, argName string) (string, error) {
	argValue, ok := arguments[argName]
	if !ok || argValue == nil {
		return "", fmt.Errorf("missing or nil input argument: %s", argName)
	}
	userInput, ok := argValue.(string)
	if !ok {
		return "", fmt.Errorf("invalid input argument type: expected string, got %T", argValue)
	}
	if userInput == "" {
		return "", fmt.Errorf("empty input variable")
	}
	return userInput, nil
}

func buildDirectCallResult(answer string, meta types.MetaInfo, err error) types.DirectCallResult {
	if err != nil {
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: err.Error()},
		}
	}
	return types.DirectCallResult{
		Success: true,
		Result:  map[string]any{"answer": answer},
		Meta:    types.MetaInfo(meta),
		Error:   types.DirectCallError{},
	}
}

func logHandleCallStep(step string, value string) {
	fmt.Fprintf(os.Stderr, "[HandleCall] %s: %s\n", step, value)
}

// NewAgentServerMode создает агент и MCPServer для server/daemon режима.
func NewAgentServerMode(configManager types.ConfigurationManagerSpec, logger types.LoggerSpec) (types.AgentSpec, *mcp_server.MCPServer, error) {
	conf := configManager.GetConfiguration()
	llmService, err := llm_service.NewLLMService(conf.GetLLMConfig(), logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create LLM service: %w", err)
	}
	logger.Info("LLM service instance created (server mode)")

	// MCP ToolConnector для server mode
	toolConnector := mcp_connector.NewMCPConnector(conf.GetMCPConnectorConfig(), logger)
	logger.Info("ToolConnector instance created (server mode)")

	agentConfig := conf.GetAgentConfig()
	calculator := llm_models.NewCalculator()
	chatInstance := chat.NewChat(
		agentConfig.Model,
		agentConfig.SystemPromptTemplate,
		agentConfig.Tool.ArgumentName,
		logger,
		calculator,
		agentConfig.MaxTokens,
		0.0,
	)
	agent := agent.NewAgent(
		agentConfig,
		llmService,
		toolConnector,
		logger,
		chatInstance,
	)
	logger.Info("Agent instance created (server mode)")

	mcpServer := mcp_server.NewMCPServer(conf.GetMCPServerConfig(), logger)
	logger.Info("MCPServer instance created (server mode)")

	return agent, mcpServer, nil
}
