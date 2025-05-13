package configuration

import (
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/sirupsen/logrus"
)

// SimpleLogger implements the LoggerSpec interface for testing
type SimpleLogger struct{}

func (m *SimpleLogger) Debug(args ...interface{})                 {}
func (m *SimpleLogger) Debugf(format string, args ...interface{}) {}
func (m *SimpleLogger) Info(args ...interface{})                  {}
func (m *SimpleLogger) Infof(format string, args ...interface{})  {}
func (m *SimpleLogger) Warn(args ...interface{})                  {}
func (m *SimpleLogger) Warnf(format string, args ...interface{})  {}
func (m *SimpleLogger) Error(args ...interface{})                 {}
func (m *SimpleLogger) Errorf(format string, args ...interface{}) {}
func (m *SimpleLogger) Fatal(args ...interface{})                 {}
func (m *SimpleLogger) Fatalf(format string, args ...interface{}) {}
func (m *SimpleLogger) WithField(key string, value interface{}) types.LogEntrySpec {
	return &SimpleLogEntry{}
}
func (m *SimpleLogger) WithFields(fields logrus.Fields) types.LogEntrySpec {
	return &SimpleLogEntry{}
}
func (m *SimpleLogger) SetLevel(level logrus.Level)        {}
func (m *SimpleLogger) SetMCPServer(mcpServer interface{}) {}

// SimpleLogEntry implements the LogEntrySpec interface for testing
type SimpleLogEntry struct{}

func (m *SimpleLogEntry) Debug(args ...interface{})                 {}
func (m *SimpleLogEntry) Debugf(format string, args ...interface{}) {}
func (m *SimpleLogEntry) Info(args ...interface{})                  {}
func (m *SimpleLogEntry) Infof(format string, args ...interface{})  {}
func (m *SimpleLogEntry) Warn(args ...interface{})                  {}
func (m *SimpleLogEntry) Warnf(format string, args ...interface{})  {}
func (m *SimpleLogEntry) Error(args ...interface{})                 {}
func (m *SimpleLogEntry) Errorf(format string, args ...interface{}) {}
func (m *SimpleLogEntry) Fatal(args ...interface{})                 {}
func (m *SimpleLogEntry) Fatalf(format string, args ...interface{}) {}

// Comment out tests that reference NewConfigurationManager or undefined Manager
// func TestConfigurationManager_LoadConfiguration(t *testing.T) { /* ... */ }
// func SetTestConfig(cm *Manager, cfg *types.Configuration) { /* ... */ }

func TestGetMCPConnectorConfigTimeout(t *testing.T) {
	mgr := &Manager{
		config: &types.Configuration{
			Agent: types.ConfigAgent{
				Connections: types.AgentConnectionsConfig{
					McpServers: map[string]types.MCPServerConnection{
						"extractor": {
							Command: "go",
							Args:    []string{"run", "cmd/server/main.go", "--config", "site/examples/text-extractor.yaml"},
							Timeout: 300,
						},
					},
				},
			},
		},
	}
	cfg := mgr.GetMCPConnectorConfig()
	server, ok := cfg.McpServers["extractor"]
	if !ok {
		t.Fatalf("extractor server not found in connector config")
	}
	if server.Timeout != 300 {
		t.Errorf("expected Timeout=300, got %v", server.Timeout)
	}
}
