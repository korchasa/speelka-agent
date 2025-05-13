package types

// MetaInfo contains metadata about a direct call execution or agent run.
type MetaInfo struct {
	Tokens           int     `json:"tokens"`
	Cost             float64 `json:"cost"`
	DurationMs       int64   `json:"duration_ms"`
	PromptTokens     int     `json:"prompt_tokens,omitempty"`
	CompletionTokens int     `json:"completion_tokens,omitempty"`
	ReasoningTokens  int     `json:"reasoning_tokens,omitempty"`
}
