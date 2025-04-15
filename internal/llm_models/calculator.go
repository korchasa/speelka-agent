package llm_models

import (
	"fmt"

	"github.com/korchasa/speelka-agent-go/internal/types"
)

// Calculator is a concrete implementation of Calculator.
type Calculator struct {
	catalog LLMModelsCatalog
}

// NewCalculator creates a new Calculator using the provided Catalog.
func NewCalculator() types.CalculatorSpec {
	return &Calculator{catalog: NewDefaultCatalog()}
}

// CalculateCost returns the USD cost for the given model and token usage.
// It uses PromptCostPerM and CompletionCostPerM from the catalog.
// If the model is not found, returns an error.
func (c *Calculator) CalculateCost(modelName string, inputTokens, outputTokens int) (float64, error) {
	model, ok := c.catalog.GetModel(modelName)
	if !ok {
		return 0, fmt.Errorf("model not found: %s", modelName)
	}
	inputCost := float64(inputTokens) * model.PromptCostPerM / 1_000_000
	outputCost := float64(outputTokens) * model.CompletionCostPerM / 1_000_000
	// Future: add cached prompt cost if/when supported by metadata
	return inputCost + outputCost, nil
}

// CalculateLLMResponse returns the number of tokens, USD cost, and approximation flag for the given model and LLM response.
func (c *Calculator) CalculateLLMResponse(modelName string, resp types.LLMResponse) (tokens int, cost float64, isApprox bool, err error) {
	model, ok := c.catalog.GetModel(modelName)
	if !ok {
		return 0, 0, false, fmt.Errorf("model not found: %s", modelName)
	}
	// If exact token info is available, use it
	if resp.Metadata.Tokens.TotalTokens != 0 {
		tokens = resp.Metadata.Tokens.TotalTokens
		inputTokens := resp.Metadata.Tokens.PromptTokens
		outputTokens := resp.Metadata.Tokens.CompletionTokens
		inputCost := float64(inputTokens) * model.PromptCostPerM / 1_000_000
		outputCost := float64(outputTokens) * model.CompletionCostPerM / 1_000_000
		cost = inputCost + outputCost
		isApprox = false
		return
	}
	// Fallback: estimate tokens and cost
	inputTokens := 0
	for _, msg := range resp.RequestMessages {
		inputTokens += len(ExtractTextFromMessageForApprox(msg))
	}
	inputTokens = inputTokens / 4
	outputTokens := len(resp.Text) / 4
	tokens = inputTokens + outputTokens
	inputCost := float64(inputTokens) * model.PromptCostPerM / 1_000_000
	outputCost := float64(outputTokens) * model.CompletionCostPerM / 1_000_000
	cost = inputCost + outputCost
	isApprox = true
	return
}
