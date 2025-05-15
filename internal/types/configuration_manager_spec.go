// Package types defines the interfaces for the MCP server components.
// Responsibility: Defining types and interfaces for all system components
// Features: Serves as the foundation for component implementation and ensures loose coupling between them
package types

import (
	"context"
)

// ConfigurationManagerSpec represents the interface for managing configuration.
// Responsibility: Providing unified access to system configuration
// Features: Supports configuration loading from files and strings,
// provides access to various types of configuration parameters
// Теперь: только загрузка и возврат итоговой структуры конфигурации
// Вся бизнес-логика и валидация — в types.Configuration
type ConfigurationManagerSpec interface {
	// LoadConfiguration loads configuration from various sources based on context.
	// It first tries to load from a configuration file if specified,
	// then applies environment variables (which take precedence).
	// Returns an error if the loading fails.
	LoadConfiguration(ctx context.Context, configFilePath string) error

	// GetConfiguration returns the final loaded configuration.
	GetConfiguration() *Configuration
}
