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
	m.applyLogConfig(&base.Runtime.Log, &newConfig.Runtime.Log)
	m.applyTransportsConfig(&base.Runtime.Transports, &newConfig.Runtime.Transports)
	m.applyAgentConfig(&base.Agent, &newConfig.Agent)
	m.applyToolConfig(&base.Agent.Tool, &newConfig.Agent.Tool)
	m.applyLLMConfig(&base.Agent.LLM, &newConfig.Agent.LLM)
	m.applyLLMRetryConfig(&base.Agent.LLM.Retry, &newConfig.Agent.LLM.Retry)
	m.applyChatConfig(&base.Agent.Chat, &newConfig.Agent.Chat)
	m.applyConnectionsConfig(&base.Agent.Connections, &newConfig.Agent.Connections)
	return base, nil
}

// applyLogConfig копирует все поля лог-конфига из overlay в base, если они заданы
func (m *Manager) applyLogConfig(base, overlay *types.RuntimeLogConfig) {
	if overlay.DefaultLevel != "" {
		base.DefaultLevel = overlay.DefaultLevel
	}
	if overlay.Output != "" {
		base.Output = overlay.Output
	}
	if overlay.Format != "" {
		base.Format = overlay.Format
	}
}

func (m *Manager) applyTransportsConfig(base, overlay *types.RuntimeTransportConfig) {
	base.HTTP.Enabled = overlay.HTTP.Enabled
	if overlay.HTTP.Host != "" {
		base.HTTP.Host = overlay.HTTP.Host
	}
	if overlay.HTTP.Port != 0 {
		base.HTTP.Port = overlay.HTTP.Port
	}
	if overlay.Stdio.Enabled != base.Stdio.Enabled {
		base.Stdio.Enabled = overlay.Stdio.Enabled
	}
	if overlay.Stdio.BufferSize != 0 {
		base.Stdio.BufferSize = overlay.Stdio.BufferSize
	}
}

func (m *Manager) applyAgentConfig(base, overlay *types.ConfigAgent) {
	if overlay.Name != "" {
		base.Name = overlay.Name
	}
	if overlay.Version != "" {
		base.Version = overlay.Version
	}
}

func (m *Manager) applyToolConfig(base, overlay *types.AgentToolConfig) {
	if overlay.Name != "" {
		base.Name = overlay.Name
	}
	if overlay.Description != "" {
		base.Description = overlay.Description
	}
	if overlay.ArgumentName != "" {
		base.ArgumentName = overlay.ArgumentName
	}
	if overlay.ArgumentDescription != "" {
		base.ArgumentDescription = overlay.ArgumentDescription
	}
}

func (m *Manager) applyLLMConfig(base, overlay *types.AgentLLMConfig) {
	if overlay.Provider != "" {
		base.Provider = overlay.Provider
	}
	if overlay.Model != "" {
		base.Model = overlay.Model
	}
	if overlay.APIKey != "" {
		base.APIKey = overlay.APIKey
	}
	if overlay.IsMaxTokensSet {
		base.MaxTokens = overlay.MaxTokens
		base.IsMaxTokensSet = true
	}
	if overlay.IsTemperatureSet {
		base.Temperature = overlay.Temperature
		base.IsTemperatureSet = true
	}
	if overlay.PromptTemplate != "" {
		base.PromptTemplate = overlay.PromptTemplate
	}
}

func (m *Manager) applyLLMRetryConfig(base, overlay *types.LLMRetryConfig) {
	if overlay.MaxRetries != 0 {
		base.MaxRetries = overlay.MaxRetries
	}
	if overlay.InitialBackoff != 0 {
		base.InitialBackoff = overlay.InitialBackoff
	}
	if overlay.MaxBackoff != 0 {
		base.MaxBackoff = overlay.MaxBackoff
	}
	if overlay.BackoffMultiplier != 0 {
		base.BackoffMultiplier = overlay.BackoffMultiplier
	}
}

func (m *Manager) applyChatConfig(base, overlay *types.AgentChatConfig) {
	if overlay.MaxTokens != 0 {
		base.MaxTokens = overlay.MaxTokens
	}
	if overlay.MaxLLMIterations != 0 {
		base.MaxLLMIterations = overlay.MaxLLMIterations
	}
	if overlay.RequestBudget != 0 {
		base.RequestBudget = overlay.RequestBudget
	}
}

func (m *Manager) applyConnectionsConfig(base, overlay *types.AgentConnectionsConfig) {
	if len(overlay.McpServers) > 0 {
		if base.McpServers == nil {
			base.McpServers = make(map[string]types.MCPServerConnection)
		}
		for name, newServer := range overlay.McpServers {
			oldServer, exists := base.McpServers[name]
			if !exists {
				base.McpServers[name] = newServer
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
			if len(newServer.Environment) > 0 {
				oldServer.Environment = newServer.Environment
			}
			if newServer.IncludeTools != nil {
				oldServer.IncludeTools = newServer.IncludeTools
			}
			if newServer.ExcludeTools != nil {
				oldServer.ExcludeTools = newServer.ExcludeTools
			}
			if newServer.Timeout != 0 {
				oldServer.Timeout = newServer.Timeout
			}
			base.McpServers[name] = oldServer
		}
	}
	if overlay.Retry.MaxRetries != 0 {
		base.Retry.MaxRetries = overlay.Retry.MaxRetries
	}
	if overlay.Retry.InitialBackoff != 0 {
		base.Retry.InitialBackoff = overlay.Retry.InitialBackoff
	}
	if overlay.Retry.MaxBackoff != 0 {
		base.Retry.MaxBackoff = overlay.Retry.MaxBackoff
	}
	if overlay.Retry.BackoffMultiplier != 0 {
		base.Retry.BackoffMultiplier = overlay.Retry.BackoffMultiplier
	}
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
