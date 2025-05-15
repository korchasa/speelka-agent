package app_direct

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/korchasa/speelka-agent-go/internal/app_mcp"
	"github.com/korchasa/speelka-agent-go/internal/types"
)

// DirectApp handles direct CLI calls, including config loading, agent initialization,
// call execution, JSON output, and proper exit codes.
type DirectApp struct {
	logger        types.LoggerSpec
	agent         directAgent
	configManager types.ConfigurationManagerSpec
}

type directAgent interface {
	CallDirect(ctx context.Context, input string) (string, types.MetaInfo, error)
}

// mcpLogStub реализует types.MCPServerNotifier и выводит MCP-логи в stderr
// Используется только в direct-call режиме для отображения логов пользователю
// Формат: [MCP level] message: ...
type mcpLogStub struct{}

func (m *mcpLogStub) SendNotificationToClient(_ context.Context, method string, data map[string]interface{}) error {
	if method != "notifications/message" {
		return nil
	}
	level, _ := data["level"].(string)
	msg, _ := data["message"].(string)
	fmt.Fprintf(os.Stderr, "[MCP %s] %s\n", level, msg)
	return nil
}

// NewDirectApp creates a new DirectApp with the given logger and configuration manager.
func NewDirectApp(logger types.LoggerSpec, configManager types.ConfigurationManagerSpec) *DirectApp {
	logger.SetMCPServer(&mcpLogStub{})
	return &DirectApp{
		logger:        logger,
		configManager: configManager,
	}
}

// Initialize loads configuration and initializes the Agent application.
func (d *DirectApp) Initialize(ctx context.Context) error {
	d.logger.Debugf("DirectApp.Initialize: start")
	agent, _, err := app_mcp.NewAgentWithServer(d.configManager, d.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}
	da, ok := agent.(directAgent)
	if !ok {
		return fmt.Errorf("agent does not implement directAgent interface")
	}
	d.agent = da
	d.logger.Debugf("DirectApp.Initialize: agent created and assigned")
	return nil
}

// HandleCall executes the direct call on the initialized agent.
func (d *DirectApp) HandleCall(ctx context.Context, input string) types.DirectCallResult {
	if d.agent == nil {
		return types.DirectCallResult{
			Success: false,
			Result:  map[string]any{"answer": ""},
			Meta:    types.MetaInfo{},
			Error:   types.DirectCallError{Type: "internal", Message: "agent not initialized"},
		}
	}
	answer, meta, err := d.agent.CallDirect(ctx, input)
	res := types.DirectCallResult{
		Success: err == nil,
		Result:  map[string]any{"answer": answer},
		Meta:    types.MetaInfo(meta),
		Error:   types.DirectCallError{},
	}
	if err != nil {
		res.Success = false
		res.Result = map[string]any{"answer": ""}
		res.Error = types.DirectCallError{Type: "internal", Message: err.Error()}
	}
	return res
}

// Execute runs the direct call workflow: initialize, call, output JSON, and exit.
func (d *DirectApp) Execute(ctx context.Context, input string) {
	if err := d.Initialize(ctx); err != nil {
		d.outputErrorAndExit("config", err)
	}
	result := d.HandleCall(ctx, input)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		d.outputErrorAndExit("internal", fmt.Errorf("failed to encode result: %w", err))
	}
	if result.Success {
		os.Exit(0)
	}
	switch result.Error.Type {
	case "user", "config":
		os.Exit(1)
	default:
		os.Exit(2)
	}
}

// outputErrorAndExit writes a JSON error to stdout and exits with the appropriate code.
func (d *DirectApp) outputErrorAndExit(errType string, err error) {
	result := types.DirectCallResult{
		Success: false,
		Result:  map[string]any{"answer": ""},
		Meta:    types.MetaInfo{},
		Error:   types.DirectCallError{Type: errType, Message: err.Error()},
	}
	output, _ := json.Marshal(result)
	fmt.Fprintln(os.Stdout, string(output))
	if errType == "user" || errType == "config" {
		os.Exit(1)
	}
	os.Exit(2)
}

// Ensure DirectApp implements the Application interface
var _ app_mcp.Application = (*DirectApp)(nil)

// Start is a no-op for DirectApp (CLI mode)
func (d *DirectApp) Start(daemonMode bool, ctx context.Context) error {
	return nil
}

// Stop is a no-op for DirectApp (CLI mode)
func (d *DirectApp) Stop(shutdownCtx context.Context) error {
	return nil
}
