// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import "github.com/korchasa/speelka-agent-go/internal/types"

// LoaderSpec defines the interface for configuration loaders
type LoaderSpec interface {
	// LoadConfiguration loads configuration data and returns a Config
	// object. Returns an error if loading fails.
	LoadConfiguration() (*types.Configuration, error)
}
