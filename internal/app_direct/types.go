package app_direct

// Types for direct call (CLI JSON output) mode only.
// Not used in agent core or MCP protocol.

// No import needed for MetaInfo

type MetaInfo struct {
	Tokens           int     `json:"tokens"`
	Cost             float64 `json:"cost"`
	DurationMs       int64   `json:"duration_ms"`
	PromptTokens     int     `json:"prompt_tokens,omitempty"`
	CompletionTokens int     `json:"completion_tokens,omitempty"`
	ReasoningTokens  int     `json:"reasoning_tokens,omitempty"`
}

type DirectCallError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type DirectCallResult struct {
	Success bool            `json:"success"`
	Result  map[string]any  `json:"result"`
	Meta    MetaInfo        `json:"meta"`
	Error   DirectCallError `json:"error"`
}
