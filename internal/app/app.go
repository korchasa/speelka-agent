package app

import (
	"context"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
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
	// Use utility to create agent and dependencies
	agent, mcpServer, err := NewAgentWithServer(a.configManager, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize agent and server: %w", err)
	}
	a.agent = agent
	a.mcpServer = mcpServer
	return nil
}

// Start starts the Agent in daemon or stdio mode
func (a *App) Start(daemonMode bool, ctx context.Context) error {
	if err := a.mcpServer.Serve(ctx, daemonMode, a.agent.HandleRequest); err != nil {
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
	if a.agent == nil {
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: "agent not initialized"},
		}
	}
	da, ok := a.agent.(directAgent)
	if !ok {
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: "agent does not implement directAgent interface"},
		}
	}
	answer, meta, err := da.CallDirect(ctx, input)
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
