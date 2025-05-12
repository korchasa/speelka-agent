package app

import (
	"context"
	"fmt"

	agentpkg "github.com/korchasa/speelka-agent-go/internal/agent"
	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/llm_service"
	"github.com/korchasa/speelka-agent-go/internal/mcp_connector"
	"github.com/korchasa/speelka-agent-go/internal/mcp_server"
	"github.com/korchasa/speelka-agent-go/internal/types"
)

// LoadConfiguration loads configuration from file/env.
func LoadConfiguration(ctx context.Context, configPath string, logger types.LoggerSpec) (types.ConfigurationManagerSpec, error) {
	// TODO: implement
	return nil, nil
}

// NewLogger creates a new logger instance.
func NewLogger() types.LoggerSpec {
	// TODO: implement
	return nil
}

// NewAgentWithDeps wires up all agent dependencies and returns an agent instance.
func NewAgentWithDeps(cfg types.AgentConfig, logger types.LoggerSpec, configManager types.ConfigurationManagerSpec) (types.AgentSpec, error) {
	// TODO: implement
	return nil, nil
}

// NewAgentWithServer creates the agent and MCP server, wiring all dependencies (for App).
func NewAgentWithServer(configManager types.ConfigurationManagerSpec, logger types.LoggerSpec) (types.AgentSpec, *mcp_server.MCPServer, error) {
	// Create LLM service
	llmService, err := llm_service.NewLLMService(configManager.GetLLMConfig(), logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create LLM service: %w", err)
	}
	logger.Info("LLM service instance created")

	// Create MCP server
	mcpServer := mcp_server.NewMCPServer(configManager.GetMCPServerConfig(), logger)
	logger.SetMCPServer(mcpServer)
	logger.Info("MCP server instance created")

	// Create MCP connector
	mcpConnector := mcp_connector.NewMCPConnector(configManager.GetMCPConnectorConfig(), logger)
	logger.Info("MCP connector instance created")

	// Initialize and connect to MCPs
	err = mcpConnector.InitAndConnectToMCPs(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init MCP connector: %w", err)
	}
	logger.Info("MCP connector connected successfully")

	// Get Agent configuration
	agentConfig := configManager.GetAgentConfig()

	// Create Calculator
	calculator := llm_models.NewCalculator()

	// Create Chat
	chatInstance := chat.NewChat(
		agentConfig.Model,
		agentConfig.SystemPromptTemplate,
		agentConfig.Tool.ArgumentName,
		logger,
		calculator,
		agentConfig.MaxTokens,
		0.0,
	)

	// Create Agent
	agent := agentpkg.NewAgent(
		agentConfig,
		llmService,
		mcpServer,
		mcpConnector,
		logger,
		chatInstance,
	)
	logger.Info("Agent instance created")

	// Register all tools
	agent.RegisterTools()

	return agent, mcpServer, nil
}
