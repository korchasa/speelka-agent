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
	fmt.Fprintf(os.Stderr, "[HandleCall] incoming request: %q\n", input)
	if a.agent == nil {
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: "agent not initialized"},
		}
	}
	fmt.Fprintf(os.Stderr, "[HandleCall] agent initialized\n")
	da, ok := a.agent.(directAgent)
	if !ok {
		fmt.Fprintf(os.Stderr, "[HandleCall] agent does not implement directAgent interface\n")
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: "agent does not implement directAgent interface"},
		}
	}
	fmt.Fprintf(os.Stderr, "[HandleCall] agent implements directAgent interface\n")
	answer, meta, err := da.CallDirect(ctx, input)
	fmt.Fprintf(os.Stderr, "[HandleCall] answer: %s\n", answer)
	fmt.Fprintf(os.Stderr, "[HandleCall] meta: %+v\n", meta)
	fmt.Fprintf(os.Stderr, "[HandleCall] error: %v\n", err)
	res := types.DirectCallResult{
		Success: err == nil,
		Result:  map[string]any{"answer": answer},
		Meta:    types.MetaInfo(meta),
		Error:   types.DirectCallError{},
	}
	if err != nil {
		res.Success = false
		res.Result = map[string]any{"answer": ""}
		res.Error = types.DirectCallError{Type: "internal", Message: err.Error()}
	}
	return res
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
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] handler called\n")
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] incoming request: %+v\n", req)
	toolName := req.Params.Name
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] toolName: %s\n", toolName)
	argName := a.configManager.GetConfiguration().GetAgentConfig().Tool.ArgumentName
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] argName: %s\n", argName)
	if toolName != a.configManager.GetConfiguration().GetAgentConfig().Tool.Name {
		errMsg := "invalid tool name: " + toolName
		fmt.Fprintf(os.Stderr, "[DispatchMCPCall] %s\n", errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}
	argValue, ok := req.Params.Arguments[argName]
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] argValue: %#v, ok: %v\n", argValue, ok)
	if !ok || argValue == nil {
		errMsg := "missing or nil input argument: " + argName
		fmt.Fprintf(os.Stderr, "[DispatchMCPCall] %s\n", errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}
	userInput, ok := argValue.(string)
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] userInput: %#v, ok: %v\n", userInput, ok)
	if !ok {
		errMsg := "invalid input argument type: expected string, got " + fmt.Sprintf("%T", argValue)
		fmt.Fprintf(os.Stderr, "[DispatchMCPCall] %s\n", errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}
	if userInput == "" {
		errMsg := "empty input variable"
		fmt.Fprintf(os.Stderr, "[DispatchMCPCall] %s\n", errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] calling agent core with input: %q\n", userInput)
	answer, _, err := a.agent.CallDirect(ctx, userInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DispatchMCPCall] core error: %v\n", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	fmt.Fprintf(os.Stderr, "[DispatchMCPCall] successful answer: %s\n", answer)
	return mcp.NewToolResultText(answer), nil
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
