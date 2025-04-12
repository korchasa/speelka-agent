package agent

import (
	"context"
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/llm_service"
	"github.com/korchasa/speelka-agent-go/internal/logger"
	"github.com/korchasa/speelka-agent-go/internal/mcp_connector"
	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
)

// App is responsible for instantiating and managing the Agent and its dependencies
type App struct {
	configManager types.ConfigurationManagerSpec
	agent         *Agent
	logger        logger.Spec
}

// NewApp creates a new instance of App with the given logger
func NewApp(logger logger.Spec) (*App, error) {
	return &App{
		logger: logger,
	}, nil
}

// LoadConfiguration loads the configuration from environment variables
func (a *App) LoadConfiguration(ctx context.Context) error {
	// Create and load configuration manager
	configManager := configuration.NewConfigurationManager(a.logger)
	err := configManager.LoadConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	a.configManager = configManager
	return nil
}

// Initialize creates and initializes all components needed by the Agent
func (a *App) Initialize() error {
	// Create LLM service
	llmService, err := llm_service.NewLLMService(a.configManager.GetLLMConfig(), a.logger)
	if err != nil {
		return fmt.Errorf("failed to create LLM service: %w", err)
	}
	a.logger.Info("LLM service instance created")

	// Create MCP server
	mcpServer := mcp_server.NewMCPServer(a.configManager.GetMCPServerConfig(), a.logger)
	a.logger.Info("MCP server instance created")

	// Create MCP connector
	mcpConnector := mcp_connector.NewMCPConnector(a.configManager.GetMCPConnectorConfig(), a.logger)
	a.logger.Info("MCP connector instance created")

	// Create Agent configuration
	agentConfig := types.AgentConfig{
		Tool:                 a.configManager.GetMCPServerConfig().Tool,
		Model:                a.configManager.GetLLMConfig().Model,
		SystemPromptTemplate: a.configManager.GetLLMConfig().SystemPromptTemplate,
		MaxTokens:            a.configManager.GetChatConfig().MaxTokens,
		CompactionStrategy:   a.configManager.GetChatConfig().CompactionStrategy,
	}

	// Create Agent
	agent := NewAgent(
		agentConfig,
		llmService,
		mcpServer,
		mcpConnector,
		a.logger,
	)
	a.logger.Info("Agent instance created")

	// Register all tools
	agent.RegisterTools()

	a.agent = agent
	return nil
}

// Start starts the Agent in daemon or stdio mode
func (a *App) Start(daemonMode bool, ctx context.Context) error {
	if a.agent == nil {
		return fmt.Errorf("agent not initialized, call Initialize() first")
	}
	return a.agent.Start(daemonMode, ctx)
}

// Stop stops the Agent
func (a *App) Stop(shutdownCtx context.Context) error {
	if a.agent == nil {
		return fmt.Errorf("agent not initialized, nothing to stop")
	}
	return a.agent.Stop(shutdownCtx)
}

// GetAgent returns the Agent instance
func (a *App) GetAgent() *Agent {
	return a.agent
}

// GetMCPServer returns the MCP server instance for external use
func (a *App) GetMCPServer() *mcp_server.MCPServer {
	if a.agent == nil {
		return nil
	}
	return a.agent.GetMCPServer()
}
