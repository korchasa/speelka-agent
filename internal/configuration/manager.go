package configuration

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	goyaml "gopkg.in/yaml.v3"
)

// Manager implements the types.ConfigurationManagerSpec interface.
// Responsibility: Managing application configuration by coordinating multiple loaders
type Manager struct {
	config *Configuration
	k      *koanf.Koanf
}

// NewConfigurationManager creates a new instance of ConfigurationManagerSpec.
// Responsibility: Factory method for creating a configuration manager
func NewConfigurationManager() *Manager {
	return &Manager{
		k: koanf.New("."),
	}
}

// LoadConfiguration loads configuration using KoanfWrapper.
// Loads default values, then from a configuration file if specified,
// and finally applies environment variables which take precedence.
func (cm *Manager) LoadConfiguration(ctx context.Context, configFilePath string) error {
	cm.k = koanf.New(".")
	if err := cm.k.Load(confmap.Provider(getDefaultConfigMap(), "."), nil); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}
	if configFilePath != "" {
		var parser koanf.Parser
		if strings.HasSuffix(configFilePath, ".yaml") || strings.HasSuffix(configFilePath, ".yml") {
			parser = yaml.Parser()
		} else if strings.HasSuffix(configFilePath, ".json") {
			parser = json.Parser()
		} else {
			return fmt.Errorf("unsupported config file format: %s", configFilePath)
		}
		if err := cm.k.Load(file.Provider(configFilePath), parser); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}
	if err := cm.k.Load(env.Provider("SPL_", ".", envKeyToPath), nil); err != nil {
		return fmt.Errorf("failed to load env: %w", err)
	}
	cfg := &Configuration{}
	unmarshalConf := koanf.UnmarshalConf{Tag: "koanf", FlatPaths: false}
	if err := cm.k.UnmarshalWithConf("", cfg, unmarshalConf); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	cm.config = cfg
	return nil
}

// envKeyToPath converts SPL_* variables to a path for koanf
// Now splits by a single underscore, does not change case.
func envKeyToPath(s string) string {
	s = strings.TrimPrefix(s, "SPL_")
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToLower(p)
	}
	return strings.Join(parts, ".")
}

// GetConfiguration returns the loaded configuration
func (cm *Manager) GetConfiguration() *Configuration {
	return cm.config
}

// Validate checks if the configuration is valid
func (cm *Manager) Validate() error {
	var validationErrors []string

	if err := cm.validateAgent(cm.config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if err := cm.validateTool(cm.config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if err := cm.validateLLM(cm.config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if err := cm.validatePrompt(cm.config); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(validationErrors, "; "))
	}
	return nil
}

func (cm *Manager) validateAgent(config *Configuration) error {
	if config.Agent.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	return nil
}

func (cm *Manager) validateTool(config *Configuration) error {
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

func (cm *Manager) validateLLM(config *Configuration) error {
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

func (cm *Manager) validatePrompt(config *Configuration) error {
	if config.Agent.LLM.PromptTemplate != "" {
		err := cm.validatePromptTemplate(config.Agent.LLM.PromptTemplate, config.Agent.Tool.ArgumentName)
		if err != nil {
			return fmt.Errorf("invalid prompt template: %v", err)
		}
	}
	return nil
}

func (cm *Manager) validatePromptTemplate(template string, argumentName string) error {
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}
	placeholders, err := cm.extractPlaceholders(template)
	if err != nil {
		return fmt.Errorf("failed to extract placeholders: %w", err)
	}
	if !contains(placeholders, argumentName) && !contains(placeholders, "input") {
		return fmt.Errorf("template must contain either {{%s}} or {{input}} placeholder", argumentName)
	}
	return nil
}

func (cm *Manager) extractPlaceholders(template string) ([]string, error) {
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

// RedactedCopy returns a copy of the configuration with private data masked for safe logging.
func RedactedCopy(config *Configuration) *Configuration {
	cpy := *config // shallow copy
	cpy.Agent.LLM.APIKey = "***REDACTED***"
	if cpy.Agent.Connections.McpServers != nil {
		redactedServers := make(map[string]MCPServerConnection, len(cpy.Agent.Connections.McpServers))
		for k, v := range cpy.Agent.Connections.McpServers {
			redacted := v
			redacted.APIKey = "***REDACTED***"
			redactedServers[k] = redacted
		}
		cpy.Agent.Connections.McpServers = redactedServers
	}
	return &cpy
}

// GetAgentConfig returns the business AgentConfig structure based on rawConfig
// While rawConfig is not filled by loaders, use cm.config for backward compatibility
func (cm *Manager) GetAgentConfig() AgentConfig {
	// While rawConfig is not filled by loaders, use cm.config for backward compatibility
	if cm.config == nil {
		return AgentConfig{}
	}
	return AgentConfig{
		Tool: MCPServerToolConfig{
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

// getDefaultConfigMap returns default values for configuration as map[string]interface{}
func getDefaultConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"runtime": map[string]interface{}{
			"log": map[string]interface{}{
				"format":       "text",
				"defaultLevel": "info",
				"disableMcp":   false,
			},
			"transports": map[string]interface{}{
				"stdio": map[string]interface{}{
					"enabled":    true,
					"bufferSize": 8192,
				},
				"http": map[string]interface{}{
					"enabled": false,
					"host":    "localhost",
					"port":    3000,
				},
			},
		},
		"agent": map[string]interface{}{
			"name":    "speelka-agent",
			"version": "1.0.0",
			"tool": map[string]interface{}{
				"name":                "process",
				"description":         "Process user queries with LLM",
				"argumentName":        "input",
				"argumentDescription": "The user query to process",
			},
			"chat": map[string]interface{}{
				"maxTokens":        8192,
				"maxLLMIterations": 100,
				"requestBudget":    1.0,
			},
			"llm": map[string]interface{}{
				"provider":       "openai",
				"model":          "gpt-4",
				"promptTemplate": "You are a helpful assistant. Respond to the following request: {{input}}. Available tools: {{tools}}",
				"temperature":    0.7,
				"apiKey":         "",
				"retry": map[string]interface{}{
					"maxRetries":        3,
					"initialBackoff":    1.0,
					"maxBackoff":        30.0,
					"backoffMultiplier": 2.0,
				},
			},
			"connections": map[string]interface{}{
				"retry": map[string]interface{}{
					"maxRetries":        3,
					"initialBackoff":    1.0,
					"maxBackoff":        30.0,
					"backoffMultiplier": 2.0,
				},
				"mcpServers": map[string]interface{}{},
			},
		},
	}
}

// MarshalConfiguration serializes the current configuration to map[string]interface{} using yaml.Marshal/yaml.Unmarshal
func (cm *Manager) MarshalConfiguration() (map[string]interface{}, error) {
	if cm.config == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}
	b, err := goyaml.Marshal(cm.config)
	if err != nil {
		return nil, fmt.Errorf("marshal to yaml: %w", err)
	}
	var out map[string]interface{}
	if err := goyaml.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}
	return out, nil
}

// UnmarshalConfiguration deserializes map[string]interface{} into Configuration struct using koanf.Unmarshal
func (cm *Manager) UnmarshalConfiguration(data map[string]interface{}) error {
	if cm.k == nil {
		cm.k = koanf.New(".")
	}
	if err := cm.k.Load(confmap.Provider(data, "."), nil); err != nil {
		return fmt.Errorf("failed to load confmap: %w", err)
	}
	cfg := &Configuration{}
	if err := cm.k.Unmarshal("", cfg); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}
	cm.config = cfg
	return nil
}
