package configuration

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/types"
)

// Manager implements the types.ConfigurationManagerSpec interface.
// Responsibility: Managing application configuration by coordinating multiple loaders
type Manager struct {
	logger        types.LoggerSpec
	config        *types.Configuration
	defaultLoader LoaderSpec
	envLoader     LoaderSpec
}

// NewConfigurationManager creates a new instance of ConfigurationManagerSpec.
// Responsibility: Factory method for creating a configuration manager
func NewConfigurationManager(logger types.LoggerSpec) *Manager {
	manager := &Manager{
		logger: logger,
	}
	// Initialize loaders
	manager.defaultLoader = NewDefaultLoader()
	manager.envLoader = NewEnvLoader()

	return manager
}

// LoadConfiguration loads configuration using the configured loaders.
// It first loads default values, then from a configuration file if specified,
// and finally applies environment variables which take precedence.
// Responsibility: Coordinating the loading of configuration from multiple sources
func (cm *Manager) LoadConfiguration(ctx context.Context, configFilePath string) error {
	if err := cm.loadDefaultConfig(); err != nil {
		return err
	}
	if err := cm.loadFileConfig(configFilePath); err != nil {
		return err
	}
	if err := cm.applyEnvConfig(); err != nil {
		return err
	}
	return nil
}

// loadDefaultConfig загружает дефолтную конфигурацию
func (cm *Manager) loadDefaultConfig() error {
	defaultConfig, err := cm.defaultLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load default configuration: %w", err)
	}
	cm.config = defaultConfig
	return nil
}

// loadFileConfig загружает конфиг из файла, если путь задан
func (cm *Manager) loadFileConfig(configFilePath string) error {
	if configFilePath == "" {
		return nil
	}
	var fileLoader LoaderSpec
	if strings.HasSuffix(configFilePath, ".yaml") || strings.HasSuffix(configFilePath, ".yml") {
		fileLoader = NewYAMLLoader(configFilePath)
	} else if strings.HasSuffix(configFilePath, ".json") {
		fileLoader = NewJSONLoader(configFilePath)
	} else {
		return fmt.Errorf("unsupported configuration file format: %s", configFilePath)
	}
	fileConfig, err := fileLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration from file: %w", err)
	}
	if _, err := cm.Apply(cm.config, fileConfig); err != nil {
		return fmt.Errorf("failed to apply file configuration: %w", err)
	}
	return nil
}

// applyEnvConfig применяет переменные окружения
func (cm *Manager) applyEnvConfig() error {
	envConfig, err := cm.envLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}
	if _, err := cm.Apply(cm.config, envConfig); err != nil {
		return fmt.Errorf("failed to apply environment configuration: %w", err)
	}
	return nil
}

// GetConfiguration returns the loaded configuration
func (cm *Manager) GetConfiguration() *types.Configuration {
	return cm.config
}

// Validate checks if the configuration is valid
func (m *Manager) Validate(config *types.Configuration) error {
	var validationErrors []string

	if err := m.validateAgent(config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if err := m.validateTool(config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if err := m.validateLLM(config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if err := m.validatePrompt(config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(validationErrors, "; "))
	}
	return nil
}

func (m *Manager) validateAgent(config *types.Configuration) error {
	if config.Agent.Name == "" {
		return fmt.Errorf("Agent name is required")
	}
	return nil
}

func (m *Manager) validateTool(config *types.Configuration) error {
	var errs []string
	if config.Agent.Tool.Name == "" {
		errs = append(errs, "Tool name is required")
	}
	if config.Agent.Tool.Description == "" {
		errs = append(errs, "Tool description is required")
	}
	if config.Agent.Tool.ArgumentName == "" {
		errs = append(errs, "Tool argument name is required")
	}
	if config.Agent.Tool.ArgumentDescription == "" {
		errs = append(errs, "Tool argument description is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (m *Manager) validateLLM(config *types.Configuration) error {
	var errs []string
	if config.Agent.LLM.APIKey == "" {
		errs = append(errs, "LLM API key is required")
	}
	if config.Agent.LLM.Provider == "" {
		errs = append(errs, "LLM provider is required")
	}
	if config.Agent.LLM.Model == "" {
		errs = append(errs, "LLM model is required")
	}
	if config.Agent.LLM.PromptTemplate == "" {
		errs = append(errs, "LLM prompt template is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (m *Manager) validatePrompt(config *types.Configuration) error {
	if config.Agent.LLM.PromptTemplate != "" {
		err := m.validatePromptTemplate(config.Agent.LLM.PromptTemplate, config.Agent.Tool.ArgumentName)
		if err != nil {
			return fmt.Errorf("Invalid prompt template: %v", err)
		}
	}
	return nil
}

func (m *Manager) validatePromptTemplate(template string, argumentName string) error {
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}
	placeholders, err := m.extractPlaceholders(template)
	if err != nil {
		return fmt.Errorf("failed to extract placeholders: %w", err)
	}
	if !contains(placeholders, argumentName) && !contains(placeholders, "input") {
		return fmt.Errorf("template must contain either {{%s}} or {{input}} placeholder", argumentName)
	}
	return nil
}

func (m *Manager) extractPlaceholders(template string) ([]string, error) {
	r, err := regexp.Compile(`\{\{([^{}]+)\}\}`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile placeholder regex: %w", err)
	}
	matches := r.FindAllStringSubmatch(template, -1)
	var placeholders []string
	for _, match := range matches {
		if len(match) > 1 {
			placeholders = append(placeholders, strings.TrimSpace(match[1]))
		}
	}
	return placeholders, nil
}

func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// Apply the changes from another configuration to this one, without overwriting values not set in the new config
func (m *Manager) Apply(base, newConfig *types.Configuration) (*types.Configuration, error) {
	if newConfig == nil {
		return base, nil
	}
	// Log
	if newConfig.Runtime.Log.DefaultLevel != "" {
		base.Runtime.Log.DefaultLevel = newConfig.Runtime.Log.DefaultLevel
	}
	if newConfig.Runtime.Log.Output != "" {
		base.Runtime.Log.Output = newConfig.Runtime.Log.Output
	}
	if newConfig.Runtime.Log.Format != "" {
		base.Runtime.Log.Format = newConfig.Runtime.Log.Format
	}
	// Transports
	if newConfig.Runtime.Transports.Stdio.Enabled != base.Runtime.Transports.Stdio.Enabled {
		base.Runtime.Transports.Stdio.Enabled = newConfig.Runtime.Transports.Stdio.Enabled
	}
	if newConfig.Runtime.Transports.Stdio.BufferSize != 0 {
		base.Runtime.Transports.Stdio.BufferSize = newConfig.Runtime.Transports.Stdio.BufferSize
	}
	if newConfig.Runtime.Transports.HTTP.Enabled != base.Runtime.Transports.HTTP.Enabled {
		base.Runtime.Transports.HTTP.Enabled = newConfig.Runtime.Transports.HTTP.Enabled
	}
	if newConfig.Runtime.Transports.HTTP.Host != "" {
		base.Runtime.Transports.HTTP.Host = newConfig.Runtime.Transports.HTTP.Host
	}
	if newConfig.Runtime.Transports.HTTP.Port != 0 {
		base.Runtime.Transports.HTTP.Port = newConfig.Runtime.Transports.HTTP.Port
	}
	// Agent
	if newConfig.Agent.Name != "" {
		base.Agent.Name = newConfig.Agent.Name
	}
	if newConfig.Agent.Version != "" {
		base.Agent.Version = newConfig.Agent.Version
	}
	// Tool
	if newConfig.Agent.Tool.Name != "" {
		base.Agent.Tool.Name = newConfig.Agent.Tool.Name
	}
	if newConfig.Agent.Tool.Description != "" {
		base.Agent.Tool.Description = newConfig.Agent.Tool.Description
	}
	if newConfig.Agent.Tool.ArgumentName != "" {
		base.Agent.Tool.ArgumentName = newConfig.Agent.Tool.ArgumentName
	}
	if newConfig.Agent.Tool.ArgumentDescription != "" {
		base.Agent.Tool.ArgumentDescription = newConfig.Agent.Tool.ArgumentDescription
	}
	// Chat
	if newConfig.Agent.Chat.MaxTokens != 0 {
		base.Agent.Chat.MaxTokens = newConfig.Agent.Chat.MaxTokens
	}
	if newConfig.Agent.Chat.MaxLLMIterations != 0 {
		base.Agent.Chat.MaxLLMIterations = newConfig.Agent.Chat.MaxLLMIterations
	}
	if newConfig.Agent.Chat.RequestBudget != 0 {
		base.Agent.Chat.RequestBudget = newConfig.Agent.Chat.RequestBudget
	}
	// LLM
	if newConfig.Agent.LLM.Provider != "" {
		base.Agent.LLM.Provider = newConfig.Agent.LLM.Provider
	}
	if newConfig.Agent.LLM.Model != "" {
		base.Agent.LLM.Model = newConfig.Agent.LLM.Model
	}
	if newConfig.Agent.LLM.APIKey != "" {
		base.Agent.LLM.APIKey = newConfig.Agent.LLM.APIKey
	}
	if newConfig.Agent.LLM.MaxTokens != 0 {
		base.Agent.LLM.MaxTokens = newConfig.Agent.LLM.MaxTokens
		base.Agent.LLM.IsMaxTokensSet = true
	}
	if newConfig.Agent.LLM.Temperature != 0 {
		base.Agent.LLM.Temperature = newConfig.Agent.LLM.Temperature
		base.Agent.LLM.IsTemperatureSet = true
	}
	if newConfig.Agent.LLM.PromptTemplate != "" {
		base.Agent.LLM.PromptTemplate = newConfig.Agent.LLM.PromptTemplate
	}
	// LLM Retry
	if newConfig.Agent.LLM.Retry.MaxRetries != 0 {
		base.Agent.LLM.Retry.MaxRetries = newConfig.Agent.LLM.Retry.MaxRetries
	}
	if newConfig.Agent.LLM.Retry.InitialBackoff != 0 {
		base.Agent.LLM.Retry.InitialBackoff = newConfig.Agent.LLM.Retry.InitialBackoff
	}
	if newConfig.Agent.LLM.Retry.MaxBackoff != 0 {
		base.Agent.LLM.Retry.MaxBackoff = newConfig.Agent.LLM.Retry.MaxBackoff
	}
	if newConfig.Agent.LLM.Retry.BackoffMultiplier != 0 {
		base.Agent.LLM.Retry.BackoffMultiplier = newConfig.Agent.LLM.Retry.BackoffMultiplier
	}
	// Connections
	if len(newConfig.Agent.Connections.McpServers) > 0 {
		if base.Agent.Connections.McpServers == nil {
			base.Agent.Connections.McpServers = make(map[string]types.MCPServerConnection)
		}
		for name, newServer := range newConfig.Agent.Connections.McpServers {
			oldServer, exists := base.Agent.Connections.McpServers[name]
			if !exists {
				base.Agent.Connections.McpServers[name] = newServer
				continue
			}
			if newServer.URL != "" {
				oldServer.URL = newServer.URL
			}
			if newServer.APIKey != "" {
				oldServer.APIKey = newServer.APIKey
			}
			if newServer.Command != "" {
				oldServer.Command = newServer.Command
			}
			if len(newServer.Args) > 0 {
				oldServer.Args = newServer.Args
			}
			if newServer.Timeout != 0 {
				oldServer.Timeout = newServer.Timeout
			}
			base.Agent.Connections.McpServers[name] = oldServer
		}
	}
	if newConfig.Agent.Connections.Retry.MaxRetries != 0 {
		base.Agent.Connections.Retry.MaxRetries = newConfig.Agent.Connections.Retry.MaxRetries
	}
	if newConfig.Agent.Connections.Retry.InitialBackoff != 0 {
		base.Agent.Connections.Retry.InitialBackoff = newConfig.Agent.Connections.Retry.InitialBackoff
	}
	if newConfig.Agent.Connections.Retry.MaxBackoff != 0 {
		base.Agent.Connections.Retry.MaxBackoff = newConfig.Agent.Connections.Retry.MaxBackoff
	}
	if newConfig.Agent.Connections.Retry.BackoffMultiplier != 0 {
		base.Agent.Connections.Retry.BackoffMultiplier = newConfig.Agent.Connections.Retry.BackoffMultiplier
	}
	return base, nil
}

// RedactedCopy возвращает копию конфигурации с замаскированными приватными данными для безопасного логирования.
func RedactedCopy(config *types.Configuration) *types.Configuration {
	copy := *config // shallow copy
	copy.Agent.LLM.APIKey = "***REDACTED***"
	if copy.Agent.Connections.McpServers != nil {
		redactedServers := make(map[string]types.MCPServerConnection, len(copy.Agent.Connections.McpServers))
		for k, v := range copy.Agent.Connections.McpServers {
			redacted := v
			redacted.APIKey = "***REDACTED***"
			redactedServers[k] = redacted
		}
		copy.Agent.Connections.McpServers = redactedServers
	}
	return &copy
}

// GetAgentConfig возвращает бизнес-структуру AgentConfig на основе rawConfig
func (cm *Manager) GetAgentConfig() types.AgentConfig {
	// Пока rawConfig не заполняется загрузчиками, используем cm.config для обратной совместимости
	if cm.config == nil {
		return types.AgentConfig{}
	}
	return types.AgentConfig{
		Tool: types.MCPServerToolConfig{
			Name:                cm.config.Agent.Tool.Name,
			Description:         cm.config.Agent.Tool.Description,
			ArgumentName:        cm.config.Agent.Tool.ArgumentName,
			ArgumentDescription: cm.config.Agent.Tool.ArgumentDescription,
		},
		Model:                cm.config.Agent.LLM.Model,
		SystemPromptTemplate: cm.config.Agent.LLM.PromptTemplate,
		MaxTokens:            cm.config.Agent.Chat.MaxTokens,
		MaxLLMIterations:     cm.config.Agent.Chat.MaxLLMIterations,
	}
}
