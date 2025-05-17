package app_direct

import (
	"context"
	"errors"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/agent"
	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/llm_service"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// DirectApp handles direct CLI calls, including config loading, agent initialization,
// call execution, JSON output, and proper exit codes.
type DirectApp struct {
	logger        types.LoggerSpec
	agent         directAgent
	configManager types.ConfigurationManagerSpec
}

type directAgent interface {
	CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error)
}

// NewDirectApp creates a new DirectApp with the given logger and configuration manager.
func NewDirectApp(logger types.LoggerSpec, configManager types.ConfigurationManagerSpec) *DirectApp {
	return &DirectApp{
		logger:        logger,
		configManager: configManager,
	}
}

// Initialize loads configuration and initializes the Agent application.
func (d *DirectApp) Initialize(ctx context.Context) error {
	d.logger.Debugf("DirectApp.Initialize: start")
	agent, err := NewAgentCLI(d.configManager, d.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}
	da, ok := agent.(directAgent)
	if !ok {
		return fmt.Errorf("agent does not implement directAgent interface")
	}
	d.agent = da
	d.logger.Debugf("DirectApp.Initialize: agent created and assigned")
	return nil
}

// HandleCall executes the direct call on the initialized agent.
func (d *DirectApp) HandleCall(ctx context.Context, input string) types.DirectCallResult {
	if d.agent == nil {
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: "agent not initialized"},
		}
	}
	answer, meta, err := d.agent.CallDirect(ctx, input)
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

// Execute runs the direct call workflow: initialize, call, output JSON, and exit.
func (d *DirectApp) Execute(ctx context.Context, input string) (types.DirectCallResult, int, error) {
	if err := d.Initialize(ctx); err != nil {
		return d.outputErrorAndExit("config", err)
	}
	result := d.HandleCall(ctx, input)
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

// outputErrorAndExit prepares a JSON error result and code, does not exit.
func (d *DirectApp) outputErrorAndExit(errType string, err error) (types.DirectCallResult, int, error) {
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

// Start is a no-op for DirectApp (CLI mode)
func (d *DirectApp) Start(daemonMode bool, ctx context.Context) error {
	return nil
}

// Stop is a no-op for DirectApp (CLI mode)
func (d *DirectApp) Stop(shutdownCtx context.Context) error {
	return nil
}

// NewAgentCLI creates an agent for CLI mode without MCPServer and external MCP connectors.
// Dummy ToolConnector for CLI: provides no external tools, only exitTool
func NewAgentCLI(configManager types.ConfigurationManagerSpec, logger types.LoggerSpec) (types.AgentSpec, error) {
	conf := configManager.GetConfiguration()
	llmService, err := llm_service.NewLLMService(conf.GetLLMConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}
	logger.Info("LLM service instance created (CLI mode)")

	// Dummy ToolConnector для CLI: не предоставляет внешних инструментов, только exitTool
	dummyConnector := &dummyToolConnector{}

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
		dummyConnector,
		logger,
		chatInstance,
	)
	logger.Info("Agent instance created (CLI mode)")
	return agent, nil
}

type dummyToolConnector struct{}

func (d *dummyToolConnector) InitAndConnectToMCPs(ctx context.Context) error { return nil }
func (d *dummyToolConnector) ConnectServer(ctx context.Context, serverID string, serverConfig types.MCPServerConnection) (client.MCPClient, error) {
	return nil, nil
}
func (d *dummyToolConnector) GetAllTools(ctx context.Context) ([]mcp.Tool, error) {
	return []mcp.Tool{}, nil
}
func (d *dummyToolConnector) ExecuteTool(ctx context.Context, call types.CallToolRequest) (*mcp.CallToolResult, error) {
	return nil, fmt.Errorf("no external tools in CLI mode")
}
func (d *dummyToolConnector) Close() error { return nil }
