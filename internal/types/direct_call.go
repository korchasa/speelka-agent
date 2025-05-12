package types

// DirectCallError represents an error in direct call (CLI JSON output) mode.
type DirectCallError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// DirectCallResult represents the result of a direct call (CLI JSON output) mode.
type DirectCallResult struct {
	Success bool            `json:"success"`
	Result  map[string]any  `json:"result"`
	Meta    MetaInfo        `json:"meta"`
	Error   DirectCallError `json:"error"`
}
