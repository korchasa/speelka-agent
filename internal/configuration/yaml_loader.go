// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/korchasa/speelka-agent-go/internal/types"
)

// YAMLLoader implements the LoaderSpec interface for loading configuration from YAML files.
type YAMLLoader struct {
	filePath string
}

// NewYAMLLoader creates a new YAMLLoader for the specified YAML file.
func NewYAMLLoader(filePath string) *YAMLLoader {
	return &YAMLLoader{
		filePath: filePath,
	}
}

// LoadConfiguration loads configuration from a YAML file.
func (l *YAMLLoader) LoadConfiguration() (*types.Configuration, error) {
	if l.filePath == "" {
		return nil, fmt.Errorf("empty file path provided")
	}

	// Check if file exists
	if _, err := os.Stat(l.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file does not exist: %s", l.filePath)
	}

	// Verify file extension
	ext := strings.ToLower(filepath.Ext(l.filePath))
	if ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("file is not a YAML file: %s", l.filePath)
	}

	// Read file content
	fileContent, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse YAML directly into Configuration struct
	var config types.Configuration
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML configuration: %w", err)
	}

	return &config, nil
}
