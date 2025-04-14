// Package configuration provides functionality for managing application configuration.
// Responsibility: Loading and providing access to application settings
package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/korchasa/speelka-agent-go/internal/types"
)

// JSONLoader implements the LoaderSpec interface for loading configuration from JSON files.
type JSONLoader struct {
	filePath string
}

// NewJSONLoader creates a new JSONLoader for the specified JSON file.
func NewJSONLoader(filePath string) *JSONLoader {
	return &JSONLoader{
		filePath: filePath,
	}
}

// LoadConfiguration loads configuration from a JSON file.
func (l *JSONLoader) LoadConfiguration() (*types.Configuration, error) {
	if l.filePath == "" {
		return nil, fmt.Errorf("empty file path provided")
	}

	// Check if file exists
	if _, err := os.Stat(l.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file does not exist: %s", l.filePath)
	}

	// Verify file extension
	ext := strings.ToLower(filepath.Ext(l.filePath))
	if ext != ".json" {
		return nil, fmt.Errorf("file is not a JSON file: %s", l.filePath)
	}

	// Read file content
	fileContent, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse JSON directly into Configuration struct
	var config types.Configuration
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}

	return &config, nil
}
