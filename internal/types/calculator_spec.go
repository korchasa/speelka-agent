package types

// CalculatorSpec computes monetary cost for LLM usage.
type CalculatorSpec interface {
	// CalculateLLMResponse returns the number of tokens, USD cost, and approximation flag for the given model and LLM response.
	CalculateLLMResponse(modelName string, resp LLMResponse) (tokens int, cost float64, isApprox bool, err error)
}
