package llm_service

import (
	"context"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/llm_models"
	"github.com/korchasa/speelka-agent-go/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

type mockLLM struct {
	llms.Model
	response *llms.ContentResponse
}

func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.response, nil
}

// NoOpLogEntry implements types.LogEntrySpec as a no-op
var _ types.LogEntrySpec = (*NoOpLogEntry)(nil)

type NoOpLogEntry struct{}

func (NoOpLogEntry) Debug(args ...interface{})                 {}
func (NoOpLogEntry) Debugf(format string, args ...interface{}) {}
func (NoOpLogEntry) Info(args ...interface{})                  {}
func (NoOpLogEntry) Infof(format string, args ...interface{})  {}
func (NoOpLogEntry) Warn(args ...interface{})                  {}
func (NoOpLogEntry) Warnf(format string, args ...interface{})  {}
func (NoOpLogEntry) Error(args ...interface{})                 {}
func (NoOpLogEntry) Errorf(format string, args ...interface{}) {}
func (NoOpLogEntry) Fatal(args ...interface{})                 {}
func (NoOpLogEntry) Fatalf(format string, args ...interface{}) {}

// NoOpLogger implements types.LoggerSpec as a no-op
var _ types.LoggerSpec = (*NoOpLogger)(nil)

type NoOpLogger struct{}

func (NoOpLogger) SetLevel(level logrus.Level)                                {}
func (NoOpLogger) Debug(args ...interface{})                                  {}
func (NoOpLogger) Debugf(format string, args ...interface{})                  {}
func (NoOpLogger) Info(args ...interface{})                                   {}
func (NoOpLogger) Infof(format string, args ...interface{})                   {}
func (NoOpLogger) Warn(args ...interface{})                                   {}
func (NoOpLogger) Warnf(format string, args ...interface{})                   {}
func (NoOpLogger) Error(args ...interface{})                                  {}
func (NoOpLogger) Errorf(format string, args ...interface{})                  {}
func (NoOpLogger) Fatal(args ...interface{})                                  {}
func (NoOpLogger) Fatalf(format string, args ...interface{})                  {}
func (NoOpLogger) WithField(key string, value interface{}) types.LogEntrySpec { return &NoOpLogEntry{} }
func (NoOpLogger) WithFields(fields logrus.Fields) types.LogEntrySpec         { return &NoOpLogEntry{} }
func (NoOpLogger) SetFormatter(formatter logrus.Formatter)                    {}
func (NoOpLogger) HandleMCPSetLevel(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func TestLLMService_SendRequest_ReturnsLLMResponse(t *testing.T) {
	mockResp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "Test response text",
				ToolCalls: []llms.ToolCall{
					{
						ID: "call_1",
						FunctionCall: &llms.FunctionCall{
							Name:      "test_tool",
							Arguments: `{"arg1":"value1"}`,
						},
					},
				},
				FuncCall: &llms.FunctionCall{
					Name:      "test_tool",
					Arguments: `{"arg1":"value1"}`,
				},
				GenerationInfo: map[string]any{
					"CompletionTokens": 10,
					"PromptTokens":     20,
					"ReasoningTokens":  5,
					"TotalTokens":      35,
				},
			},
		},
	}

	svc := &LLMService{
		client:     &mockLLM{response: mockResp},
		logger:     NoOpLogger{},
		config:     types.LLMConfig{}, // Provide default config to avoid nil dereference
		calculator: llm_models.NewCalculator(),
	}

	resp, err := svc.SendRequest(context.Background(), []llms.MessageContent{}, []mcp.Tool{})
	assert.NoError(t, err)
	assert.Equal(t, "Test response text", resp.Text)
	assert.Len(t, resp.Calls, 1)
	assert.Equal(t, "test_tool", resp.Calls[0].ToolName())
	assert.Equal(t, 10, resp.Metadata.Tokens.CompletionTokens)
	assert.Equal(t, 20, resp.Metadata.Tokens.PromptTokens)
	assert.Equal(t, 5, resp.Metadata.Tokens.ReasoningTokens)
	assert.Equal(t, 35, resp.Metadata.Tokens.TotalTokens)
	assert.GreaterOrEqual(t, resp.Metadata.DurationMs, int64(0), "DurationMs should be set and >= 0")
}
