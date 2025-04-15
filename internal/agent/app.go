// TODO: move to separate package
package agent

import (
	"context"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/llm_service"
	"github.com/korchasa/speelka-agent-go/internal/mcp_connector"
	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
)

// App is responsible for instantiating and managing the Agent and its dependencies
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
	// Create LLM service
	llmService, err := llm_service.NewLLMService(a.configManager.GetLLMConfig(), a.logger)
	if err != nil {
		return fmt.Errorf("failed to create LLM service: %w", err)
	}
	a.logger.Info("LLM service instance created")

	// Create MCP server
	a.mcpServer = mcp_server.NewMCPServer(a.configManager.GetMCPServerConfig(), a.logger)
	a.logger.SetMCPServer(a.mcpServer)
	a.logger.Info("MCP server instance created")

	// Create MCP connector
	mcpConnector := mcp_connector.NewMCPConnector(a.configManager.GetMCPConnectorConfig(), a.logger)
	a.logger.Info("MCP connector instance created")

	// First, initialize and connect to MCPs
	err = mcpConnector.InitAndConnectToMCPs(ctx)
	if err != nil {
		return fmt.Errorf("failed to init MCP connector: %w", err)
	}
	a.logger.Info("MCP connector connected successfully")

	// Get Agent configuration
	agentConfig := a.configManager.GetAgentConfig()

	// Create Calculator and CompactionStrategy
	calculator := llm_models.NewCalculator()
	compactionStrategy, err := chat.GetCompactionStrategy(agentConfig.CompactionStrategy, agentConfig.Model, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create compaction strategy: %w", err)
	}

	// Create Chat
	chatInstance := chat.NewChat(
		agentConfig.Model,
		agentConfig.SystemPromptTemplate,
		agentConfig.Tool.ArgumentName,
		a.logger,
		calculator,
		compactionStrategy,
		agentConfig.MaxTokens,
	)

	// Create Agent
	agent := NewAgent(
		agentConfig,
		llmService,
		a.mcpServer,
		mcpConnector,
		a.logger,
		chatInstance,
	)
	a.logger.Info("Agent instance created")

	// Register all tools
	agent.RegisterTools()

	a.agent = agent
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
