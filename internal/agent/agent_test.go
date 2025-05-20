package agent

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	types2 "github.com/korchasa/speelka-agent-go/internal/llm/types"

	"github.com/korchasa/speelka-agent-go/internal/chat"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

// Helper function to create a test logger
func newTestLogger() *logrus.Logger {
	buf := &bytes.Buffer{}
	log := logrus.New()
	log.SetOutput(buf)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return log
}

func TestHandleLLMAnswerToolRequest(t *testing.T) {
	a := NewAgent(configuration.AgentConfig{}, nil, nil, newTestLogger(), nil)
	sess := &chat.Chat{} // Not used in this test
	resp := types2.LLMResponse{}

	t.Run("missing text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{}
		res := a.HandleLLMFinishToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for missing text argument, got success")
		}
	})

	t.Run("nil text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": nil}
		res := a.HandleLLMFinishToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for nil text argument, got success")
		}
	})

	t.Run("non-string text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": 123}
		res := a.HandleLLMFinishToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for non-string text argument, got success")
		}
	})

	t.Run("empty string text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": ""}
		res := a.HandleLLMFinishToolRequest(call, resp, sess)
		if !res.IsError {
			t.Errorf("expected error for empty string text argument, got success")
		}
	})

	t.Run("valid string text argument", func(t *testing.T) {
		call := types.CallToolRequest{}
		call.Params.Arguments = map[string]interface{}{"text": "hello"}
		res := a.HandleLLMFinishToolRequest(call, resp, sess)
		if res.IsError {
			t.Errorf("expected success for valid string, got error")
		}
		if len(res.Content) == 0 {
			t.Errorf("expected content in result")
		}
		found := false
		for _, c := range res.Content {
			if tc, ok := c.(mcp.TextContent); ok && tc.Text == "hello" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected text content 'hello' in result")
		}
	})
}

// --- BEGIN: Unit tests for CallDirect and RunSession ---
type mockToolConnector struct {
	getAllToolsErr error
	tools          []mcp.Tool
	executeToolFn  func(ctx context.Context, call types.CallToolRequest) (*mcp.CallToolResult, error)
}

func (m *mockToolConnector) InitAndConnectToMCPs(ctx context.Context) error { return nil }
func (m *mockToolConnector) ConnectServer(ctx context.Context, serverID string, serverConfig configuration.MCPServerConnection) (client.MCPClient, error) {
	return nil, nil
}
func (m *mockToolConnector) GetAllTools(ctx context.Context) ([]mcp.Tool, error) {
	return m.tools, m.getAllToolsErr
}
func (m *mockToolConnector) ExecuteTool(ctx context.Context, call types.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.executeToolFn != nil {
		return m.executeToolFn(ctx, call)
	}
	return mcp.NewToolResultText("ok"), nil
}
func (m *mockToolConnector) Close() error { return nil }

type mockLLMService struct {
	responses []types2.LLMResponse
	err       error
	callIdx   int
}

func (m *mockLLMService) SendRequest(ctx context.Context, messages []llms.MessageContent, tools []mcp.Tool) (types2.LLMResponse, error) {
	if m.err != nil {
		return types2.LLMResponse{}, m.err
	}
	if m.callIdx < len(m.responses) {
		resp := m.responses[m.callIdx]
		m.callIdx++
		return resp, nil
	}
	return types2.LLMResponse{}, nil
}

func TestAgent_RunSession(t *testing.T) {
	t.Run("error on GetAllTools", func(t *testing.T) {
		chatInstance := chat.NewChat("model", "prompt", "arg", newTestLogger(), nil, 10, 0.0)
		agent := NewAgent(
			configuration.AgentConfig{MaxLLMIterations: 1},
			&mockLLMService{},
			&mockToolConnector{getAllToolsErr: fmt.Errorf("fail")},
			newTestLogger(),
			chatInstance,
		)
		_, _, err := agent.RunSession(context.Background(), "input")
		if err == nil || err.Error() != "fail" && !strings.Contains(err.Error(), "fail") {
			t.Errorf("expected error from GetAllTools, got %v", err)
		}
	})
	t.Run("error on LLMService", func(t *testing.T) {
		chatInstance := chat.NewChat("model", "prompt", "arg", newTestLogger(), nil, 10, 0.0)
		agent := NewAgent(
			configuration.AgentConfig{MaxLLMIterations: 1},
			&mockLLMService{err: fmt.Errorf("llm fail")},
			&mockToolConnector{tools: []mcp.Tool{finishTool}},
			newTestLogger(),
			chatInstance,
		)
		_, _, err := agent.RunSession(context.Background(), "input")
		if err == nil || !strings.Contains(err.Error(), "llm fail") {
			t.Errorf("expected error from LLMService, got %v", err)
		}
	})
	t.Run("exceed max iterations", func(t *testing.T) {
		chatInstance := chat.NewChat("model", "prompt", "arg", newTestLogger(), nil, 10, 0.0)
		llmCall := llms.ToolCall{
			ID:   "call-id-1",
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      "some_tool",
				Arguments: `{"input": "test"}`,
			},
		}
		call, err := types.NewCallToolRequest(llmCall)
		if err != nil {
			t.Fatalf("failed to create CallToolRequest: %v", err)
		}
		agent := NewAgent(
			configuration.AgentConfig{MaxLLMIterations: 1},
			&mockLLMService{responses: []types2.LLMResponse{{Calls: []types.CallToolRequest{call}}}},
			&mockToolConnector{tools: []mcp.Tool{finishTool}},
			newTestLogger(),
			chatInstance,
		)
		_, _, err = agent.RunSession(context.Background(), "input")
		if err == nil || !strings.Contains(err.Error(), "exceeded maximum number of LLM iterations") {
			t.Errorf("expected max iterations error, got %v", err)
		}
	})
	t.Run("success exitTool", func(t *testing.T) {
		chatInstance := chat.NewChat("model", "prompt", "arg", newTestLogger(), nil, 10, 0.0)
		llmCall := llms.ToolCall{
			ID:   "call-id-1",
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      finishTool.Name,
				Arguments: `{"text": "done!"}`,
			},
		}
		call, err := types.NewCallToolRequest(llmCall)
		if err != nil {
			t.Fatalf("failed to create CallToolRequest: %v", err)
		}
		calls := []types.CallToolRequest{call}
		llmResp := types2.LLMResponse{
			Text:  "irrelevant",
			Calls: calls,
			Metadata: types2.LLMResponseMetadata{
				Tokens: types2.LLMResponseTokensMetadata{TotalTokens: 1},
			},
		}
		agent := NewAgent(
			configuration.AgentConfig{MaxLLMIterations: 2},
			&mockLLMService{responses: []types2.LLMResponse{llmResp}},
			&mockToolConnector{tools: []mcp.Tool{finishTool}},
			newTestLogger(),
			chatInstance,
		)
		msg, meta, err := agent.RunSession(context.Background(), "input")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if msg != "done!" {
			t.Errorf("expected 'done!', got %q", msg)
		}
		if meta.Tokens == 0 {
			t.Errorf("expected meta tokens > 0")
		}
	})
}

// --- END: Unit tests for CallDirect and RunSession ---
