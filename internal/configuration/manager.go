package configuration

import (
	"context"
	"fmt"
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
	defaultConfig, err := cm.defaultLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load default configuration: %w", err)
	}
	cm.config = defaultConfig

	// Load from file if specified
	if configFilePath != "" {
		var fileLoader LoaderSpec

		// Choose loader based on file extension
		if strings.HasSuffix(configFilePath, ".yaml") || strings.HasSuffix(configFilePath, ".yml") {
			fileLoader = NewYAMLLoader(configFilePath)
		} else if strings.HasSuffix(configFilePath, ".json") {
			fileLoader = NewJSONLoader(configFilePath)
		} else {
			return fmt.Errorf("unsupported configuration file format: %s", configFilePath)
		}

		// Load configuration from file
		fileConfig, err := fileLoader.LoadConfiguration()
		if err != nil {
			return fmt.Errorf("failed to load configuration from file: %w", err)
		}

		// Apply file configuration to default configuration instead of replacing it
		if _, err := cm.config.Apply(fileConfig); err != nil {
			return fmt.Errorf("failed to apply file configuration: %w", err)
		}
	}

	// Load and apply environment variables (highest precedence)
	envConfig, err := cm.envLoader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Apply environment variables if they exist
	if _, err := cm.config.Apply(envConfig); err != nil {
		return fmt.Errorf("failed to apply environment configuration: %w", err)
	}
	return nil
}

// GetConfiguration returns the loaded configuration
func (cm *Manager) GetConfiguration() *types.Configuration {
	return cm.config
}
