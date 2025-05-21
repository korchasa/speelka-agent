package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/agent"
	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/llm"
	"github.com/korchasa/speelka-agent-go/internal/llm/cost"
	"github.com/korchasa/speelka-agent-go/internal/mcp_connector"
	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

// MCPApp is responsible for instantiating and managing the Agent and its dependencies
// (for server/daemon mode)
type MCPApp struct {
	cfg       *configuration.Configuration
	agent     agentSpec
	mcpServer *mcp_server.MCPServer
	logger    *logrus.Logger
}

// NewMCPApp creates a new instance of MCPApp with the given logger and configuration
func NewMCPApp(logger *logrus.Logger, cfg *configuration.Configuration) (*MCPApp, error) {
	logger.Infof("MCPApp: creating new instance with config: %+v", cfg)
	return &MCPApp{
		logger: logger,
		cfg:    cfg,
	}, nil
}

type agentSpec interface {
	RunSession(ctx context.Context, input string) (string, types.MetaInfo, error)
}

// Initialize creates and initializes all components needed by the Agent
func (a *MCPApp) Initialize(ctx context.Context) error {
	ag, err := buildAgent(ctx, a.cfg, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize agent and server: %w", err)
	}
	a.agent = ag
	return nil
}

// Start starts the Agent in daemon or stdio mode
func (a *MCPApp) Start(ctx context.Context) (err error) {
	a.mcpServer, err = mcp_server.NewMCPServer(a.cfg.GetMCPServerConfig(), a.logger)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}
	a.logger.Info("MCPServer instance created (server mode)")
	if err = a.mcpServer.Serve(ctx, a.dispatchMCPCall); err != nil {
		return fmt.Errorf("failed to serve mcp server: %w", err)
	}
	return nil
}

// Stop stops the Agent
func (a *MCPApp) Stop(shutdownCtx context.Context) error {
	if err := a.mcpServer.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop HTTP MCP server: %w", err)
	}
	a.logger.Info("Server shutdown complete")
	return nil
}

// ExecuteDirectCall runs the direct call workflow: initialize, call, output JSON, and exit.
func (a *MCPApp) ExecuteDirectCall(ctx context.Context, input string) (types.DirectCallResult, int, error) {
	if a.agent == nil {
		return a.outputErrorAndExit("config", fmt.Errorf("agent not initialized"))
	}
	result := a.handleDirectCall(ctx, input)
	if result.Success {
		return result, 0, nil
	}
	switch result.Error.Type {
	case "user", "config":
		return result, 1, errors.New(result.Error.Message)
	default:
		return result, 2, errors.New(result.Error.Message)
	}
}

// handleDirectCall executes the direct call on the initialized agent.
func (a *MCPApp) handleDirectCall(ctx context.Context, input string) types.DirectCallResult {
	answer, meta, err := a.agent.RunSession(ctx, input)
	res := types.DirectCallResult{
		Success: err == nil,
		Result:  map[string]any{"answer": answer},
		Meta:    meta,
		Error:   types.DirectCallError{},
	}
	if err != nil {
		res.Success = false
		res.Result = map[string]any{"answer": ""}
		res.Error = types.DirectCallError{Type: "internal", Message: err.Error()}
	}
	return res
}

func (a *MCPApp) dispatchMCPCall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolName := req.Params.Name
	argName := a.cfg.GetAgentConfig().Tool.ArgumentName
	if err := validateToolName(toolName, a.cfg); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments is not a map"), nil
	}
	userInput, err := extractUserInput(args, argName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	answer, _, err := a.agent.RunSession(ctx, userInput)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(answer), nil
}

// outputErrorAndExit prepares a JSON error result and code, does not exit.
func (a *MCPApp) outputErrorAndExit(errType string, err error) (types.DirectCallResult, int, error) {
	result := types.DirectCallResult{
		Success: false,
		Result:  map[string]any{"answer": ""},
		Meta:    types.MetaInfo{},
		Error:   types.DirectCallError{Type: errType, Message: err.Error()},
	}
	code := 2
	if errType == "user" || errType == "config" {
		code = 1
	}
	return result, code, err
}

func validateToolName(toolName string, cfg *configuration.Configuration) error {
	expected := cfg.GetAgentConfig().Tool.Name
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
		Meta:    meta,
		Error:   types.DirectCallError{},
	}
}

// buildAgent creates an agent and MCPServer for server/daemon mode.
func buildAgent(ctx context.Context, cfg *configuration.Configuration, log *logrus.Logger) (agentSpec, error) {
	llmService, err := llm.NewLLMService(cfg.GetLLMConfig(), log)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}
	log.Info("LLM service instance created (server mode)")

	// MCP ToolConnector for server mode
	toolConnector := mcp_connector.NewMCPConnector(cfg.GetMCPConnectorConfig(), log)
	log.Info("ToolConnector instance created (server mode)")

	// initialization of MCP connections and loading tools
	if err := toolConnector.InitAndConnectToMCPs(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP connections: %w", err)
	}

	agentConfig := cfg.GetAgentConfig()
	calculator := cost.NewCalculator()
	chatInstance := chat.NewChat(
		agentConfig.Model,
		agentConfig.SystemPromptTemplate,
		agentConfig.Tool.ArgumentName,
		log,
		calculator,
		agentConfig.MaxTokens,
		0.0,
	)
	ag := agent.NewAgent(
		agentConfig,
		llmService,
		toolConnector,
		log,
		chatInstance,
	)
	log.Info("Agent instance created (server mode)")

	return ag, nil
}
