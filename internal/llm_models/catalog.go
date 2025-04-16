// Package llm_models provides a catalog of LLM models and their pricing.
// Based on project https://github.com/AgentOps-AI/tokencost
package llm_models

import (
	"strings"
)

// ModelInfo holds pricing and token info for an LLM model.
type ModelInfo struct {
	Name                 string   // Canonical model name
	PromptCostPerM       float64  // USD per 1M prompt tokens
	CachedPromptCostPerM float64  // USD per 1M cached prompt tokens
	CompletionCostPerM   float64  // USD per 1M completion tokens
	MaxPromptTokens      int      // Maximum prompt tokens
	MaxCompletionTokens  int      // Maximum completion tokens
	Aliases              []string // Alternative names/aliases
}

// LLMModelsCatalog provides lookup for LLM model pricing and limits.
type LLMModelsCatalog interface {
	// GetModel returns ModelInfo for a given model name (case-insensitive, alias-aware).
	GetModel(name string) (ModelInfo, bool)
	// ListModels returns all known models.
	ListModels() []ModelInfo
}

// Catalog is a concrete implementation of LLMModelsCatalog.
type Catalog struct {
	models map[string]ModelInfo // key: normalized name
	alias  map[string]string    // alias -> canonical name
}

// normalizeName lowercases and trims the model name for consistent lookup.
func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// NewDefaultCatalog returns a catalog populated with all model data and aliases from the pricing table.
func NewDefaultCatalog() LLMModelsCatalog {
	models := map[string]ModelInfo{
		// OpenAI GPT-4 family
		"gpt-4":       {Name: "gpt-4", PromptCostPerM: 30.00, CompletionCostPerM: 60.00, MaxPromptTokens: 8192, MaxCompletionTokens: 4096, Aliases: []string{"gpt-4-0314", "gpt-4-0613"}},
		"gpt-4-32k":   {Name: "gpt-4-32k", PromptCostPerM: 60.00, CompletionCostPerM: 120.00, MaxPromptTokens: 32768, MaxCompletionTokens: 4096, Aliases: []string{"gpt-4-32k-0314", "gpt-4-32k-0613"}},
		"gpt-4o":      {Name: "gpt-4o", PromptCostPerM: 2.5, CompletionCostPerM: 10.0, MaxPromptTokens: 128000, MaxCompletionTokens: 16384, Aliases: []string{"gpt-4o-2024-08-06", "gpt-4o-2024-05-13", "chatgpt-4o-latest", "gpt-4o-audio-preview", "gpt-4o-audio-preview-2024-10-01"}},
		"gpt-4o-mini": {Name: "gpt-4o-mini", PromptCostPerM: 0.15, CompletionCostPerM: 0.6, MaxPromptTokens: 128000, MaxCompletionTokens: 16384, Aliases: []string{"gpt-4o-mini-2024-07-18"}},
		"gpt-4-turbo": {Name: "gpt-4-turbo", PromptCostPerM: 10.0, CompletionCostPerM: 30.0, MaxPromptTokens: 128000, MaxCompletionTokens: 4096, Aliases: []string{"gpt-4-turbo-preview", "gpt-4-turbo-2024-04-09", "gpt-4-1106-preview", "gpt-4-0125-preview", "gpt-4-vision-preview", "gpt-4-1106-vision-preview"}},
		"gpt-4.1": {
			Name:                 "gpt-4.1",
			PromptCostPerM:       2,
			CachedPromptCostPerM: 0.5,
			CompletionCostPerM:   8.0,
			MaxPromptTokens:      1047576,
			MaxCompletionTokens:  32768,
			Aliases:              []string{"gpt-4.1-2025-04-14"},
		},
		"gpt-4.1-mini": {
			Name:                 "gpt-4.1",
			PromptCostPerM:       0.4,
			CachedPromptCostPerM: 0.1,
			CompletionCostPerM:   1.6,
			MaxPromptTokens:      1047576,
			MaxCompletionTokens:  32768,
			Aliases:              []string{},
		},
		"gpt-4.1-nano": {
			Name:                 "gpt-4.1",
			PromptCostPerM:       0.1,
			CachedPromptCostPerM: 0.03,
			CompletionCostPerM:   0.4,
			MaxPromptTokens:      1047576,
			MaxCompletionTokens:  32768,
			Aliases:              []string{},
		},
		// OpenAI GPT-3.5 family
		"gpt-3.5-turbo": {Name: "gpt-3.5-turbo", PromptCostPerM: 1.5, CompletionCostPerM: 2.0, MaxPromptTokens: 16385, MaxCompletionTokens: 4096, Aliases: []string{"gpt-3.5-turbo-0301", "gpt-3.5-turbo-0613", "gpt-3.5-turbo-1106", "gpt-3.5-turbo-0125", "gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613"}},
		// OpenAI fine-tuned
		"ft:gpt-3.5-turbo":          {Name: "ft:gpt-3.5-turbo", PromptCostPerM: 3.0, CompletionCostPerM: 6.0, MaxPromptTokens: 16385, MaxCompletionTokens: 4096, Aliases: []string{"ft:gpt-3.5-turbo-0125", "ft:gpt-3.5-turbo-1106", "ft:gpt-3.5-turbo-0613"}},
		"ft:gpt-4-0613":             {Name: "ft:gpt-4-0613", PromptCostPerM: 30.0, CompletionCostPerM: 60.0, MaxPromptTokens: 8192, MaxCompletionTokens: 4096, Aliases: nil},
		"ft:gpt-4o-2024-08-06":      {Name: "ft:gpt-4o-2024-08-06", PromptCostPerM: 3.75, CompletionCostPerM: 15.0, MaxPromptTokens: 128000, MaxCompletionTokens: 16384, Aliases: nil},
		"ft:gpt-4o-mini-2024-07-18": {Name: "ft:gpt-4o-mini-2024-07-18", PromptCostPerM: 0.3, CompletionCostPerM: 1.2, MaxPromptTokens: 128000, MaxCompletionTokens: 16384, Aliases: nil},
		// OpenAI legacy
		"ft:davinci-002": {Name: "ft:davinci-002", PromptCostPerM: 2.0, CompletionCostPerM: 2.0, MaxPromptTokens: 16384, MaxCompletionTokens: 4096, Aliases: nil},
		"ft:babbage-002": {Name: "ft:babbage-002", PromptCostPerM: 0.4, CompletionCostPerM: 0.4, MaxPromptTokens: 16384, MaxCompletionTokens: 4096, Aliases: nil},
		// O1 models
		"o1-mini":    {Name: "o1-mini", PromptCostPerM: 1.1, CompletionCostPerM: 4.4, MaxPromptTokens: 128000, MaxCompletionTokens: 65536, Aliases: []string{"o1-mini-2024-09-12"}},
		"o1-preview": {Name: "o1-preview", PromptCostPerM: 15.0, CompletionCostPerM: 60.0, MaxPromptTokens: 128000, MaxCompletionTokens: 32768, Aliases: []string{"o1-preview-2024-09-12"}},
		"o1-pro":     {Name: "o1-pro", PromptCostPerM: 150.0, CompletionCostPerM: 600.0, MaxPromptTokens: 200000, MaxCompletionTokens: 100000, Aliases: []string{"o1-pro-2025-03-19"}},
		// Anthropic Claude (examples, not exhaustive)
		"claude-3-opus":   {Name: "claude-3-opus", PromptCostPerM: 15.0, CompletionCostPerM: 75.0, MaxPromptTokens: 200000, MaxCompletionTokens: 4096, Aliases: []string{"claude-3-opus-20240229"}},
		"claude-3-sonnet": {Name: "claude-3-sonnet", PromptCostPerM: 3.0, CompletionCostPerM: 15.0, MaxPromptTokens: 200000, MaxCompletionTokens: 4096, Aliases: []string{"claude-3-sonnet-20240229"}},
		"claude-3-haiku":  {Name: "claude-3-haiku", PromptCostPerM: 0.25, CompletionCostPerM: 1.25, MaxPromptTokens: 200000, MaxCompletionTokens: 4096, Aliases: []string{"claude-3-haiku-20240307"}},
		// Azure OpenAI (examples)
		"azure/gpt-4o-2024-08-06": {Name: "azure/gpt-4o-2024-08-06", PromptCostPerM: 2.75, CompletionCostPerM: 11.0, MaxPromptTokens: 128000, MaxCompletionTokens: 16384, Aliases: []string{"azure/us/gpt-4o-2024-08-06", "azure/eu/gpt-4o-2024-08-06", "azure/global/gpt-4o-2024-08-06"}},
		"azure/gpt-4o-2024-11-20": {Name: "azure/gpt-4o-2024-11-20", PromptCostPerM: 2.75, CompletionCostPerM: 11.0, MaxPromptTokens: 128000, MaxCompletionTokens: 16384, Aliases: []string{"azure/us/gpt-4o-2024-11-20", "azure/eu/gpt-4o-2024-11-20", "azure/global/gpt-4o-2024-11-20"}},
		// Gemini (examples)
		"gemini/gemini-2.0-pro-exp-02-05":            {Name: "gemini/gemini-2.0-pro-exp-02-05", PromptCostPerM: 0.0, CompletionCostPerM: 0.0, MaxPromptTokens: 2097152, MaxCompletionTokens: 8192, Aliases: nil},
		"gemini/gemini-2.0-flash-thinking-exp-01-21": {Name: "gemini/gemini-2.0-flash-thinking-exp-01-21", PromptCostPerM: 0.0, CompletionCostPerM: 0.0, MaxPromptTokens: 1048576, MaxCompletionTokens: 65536, Aliases: nil},
		// DALL-E (examples, not used for LLM cost)
		"256-x-256/dall-e-2": {Name: "256-x-256/dall-e-2", PromptCostPerM: 0.0, CompletionCostPerM: 0.0, MaxPromptTokens: 0, MaxCompletionTokens: 0, Aliases: nil},
		// Add more models as needed from the table...
	}
	// Aliasing for all variants
	alias := map[string]string{}
	for canonical, info := range models {
		alias[normalizeName(canonical)] = canonical
		for _, a := range info.Aliases {
			alias[normalizeName(a)] = canonical
		}
	}
	return &Catalog{models: models, alias: alias}
}

// GetModel returns ModelInfo for a given model name (case-insensitive, alias-aware).
func (c *Catalog) GetModel(nameOrAlias string) (ModelInfo, bool) {
	norm := normalizeName(nameOrAlias)
	if model, ok := c.models[norm]; ok {
		return model, true
	}
	if canonical, ok := c.alias[norm]; ok {
		return c.models[canonical], true
	}
	return ModelInfo{}, false
}

// ListModels returns all known models.
func (c *Catalog) ListModels() []ModelInfo {
	out := make([]ModelInfo, 0, len(c.models))
	for _, m := range c.models {
		out = append(out, m)
	}
	return out
}
