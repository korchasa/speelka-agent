package llm

import (
	"bytes"
	"context"
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	"github.com/korchasa/speelka-agent-go/internal/llm/cost"
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

// Вспомогательная функция для создания тестового логгера
func newTestLogger() *logrus.Logger {
	buf := &bytes.Buffer{}
	log := logrus.New()
	log.SetOutput(buf)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return log
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
		logger:     newTestLogger(),
		config:     configuration.LLMConfig{}, // Provide default config to avoid nil dereference
		calculator: cost.NewCalculator(),
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

func TestNewLLMService_Validation(t *testing.T) {
	logger := newTestLogger()
	cfg := configuration.LLMConfig{Provider: "", Model: "gpt", APIKey: "key"}
	_, err := NewLLMService(cfg, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider is required")

	cfg = configuration.LLMConfig{Provider: "openai", Model: "", APIKey: "key"}
	_, err = NewLLMService(cfg, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is required")

	cfg = configuration.LLMConfig{Provider: "openai", Model: "gpt", APIKey: ""}
	_, err = NewLLMService(cfg, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key is required")

	cfg = configuration.LLMConfig{Provider: "gibberish", Model: "gpt", APIKey: "key"}
	_, err = NewLLMService(cfg, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestLLMService_SendRequest_NotInitialized(t *testing.T) {
	svc := &LLMService{client: nil, logger: newTestLogger(), config: configuration.LLMConfig{}}
	resp, err := svc.SendRequest(context.Background(), nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM service not initialized")
	assert.Empty(t, resp.Text)
}

func TestLLMService_SendRequest_EmptyResponse(t *testing.T) {
	mockResp := &llms.ContentResponse{Choices: []*llms.ContentChoice{}}
	svc := &LLMService{
		client:     &mockLLM{response: mockResp},
		logger:     newTestLogger(),
		config:     configuration.LLMConfig{},
		calculator: cost.NewCalculator(),
	}
	resp, err := svc.SendRequest(context.Background(), nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from LLM")
	assert.Empty(t, resp.Text)
}

func TestLLMService_SendRequest_NoFuncCall(t *testing.T) {
	mockResp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{FuncCall: nil}},
	}
	svc := &LLMService{
		client:     &mockLLM{response: mockResp},
		logger:     newTestLogger(),
		config:     configuration.LLMConfig{},
		calculator: cost.NewCalculator(),
	}
	resp, err := svc.SendRequest(context.Background(), nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no function call in response")
	assert.Empty(t, resp.Text)
}

func TestLLMService_SendRequest_ConvertToolsError(t *testing.T) {
	// Pass incorrect RawInputSchema
	badTool := mcp.Tool{Name: "bad", RawInputSchema: []byte("{bad json")}
	svc := &LLMService{
		client:     &mockLLM{response: &llms.ContentResponse{Choices: []*llms.ContentChoice{{FuncCall: &llms.FunctionCall{Name: "bad"}}}}},
		logger:     newTestLogger(),
		config:     configuration.LLMConfig{},
		calculator: cost.NewCalculator(),
	}
	resp, err := svc.SendRequest(context.Background(), nil, []mcp.Tool{badTool})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to convert tools to LLM tools")
	assert.Empty(t, resp.Text)
}
