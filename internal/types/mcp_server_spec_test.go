package types

import (
	"testing"
)

func TestMCPServerConnection_IsToolAllowed(t *testing.T) {
	tests := []struct {
		name        string
		conn        MCPServerConnection
		tool        string
		wantAllowed bool
	}{
		{
			name:        "No filters, tool allowed",
			conn:        MCPServerConnection{},
			tool:        "foo",
			wantAllowed: true,
		},
		{
			name:        "Tool in ExcludeTools, disallowed",
			conn:        MCPServerConnection{ExcludeTools: []string{"foo", "bar"}},
			tool:        "foo",
			wantAllowed: false,
		},
		{
			name:        "Tool not in ExcludeTools, allowed",
			conn:        MCPServerConnection{ExcludeTools: []string{"bar"}},
			tool:        "foo",
			wantAllowed: true,
		},
		{
			name:        "Tool in IncludeTools, allowed",
			conn:        MCPServerConnection{IncludeTools: []string{"foo", "baz"}},
			tool:        "foo",
			wantAllowed: true,
		},
		{
			name:        "Tool not in IncludeTools, disallowed",
			conn:        MCPServerConnection{IncludeTools: []string{"bar", "baz"}},
			tool:        "foo",
			wantAllowed: false,
		},
		{
			name:        "Tool in both IncludeTools and ExcludeTools, disallowed",
			conn:        MCPServerConnection{IncludeTools: []string{"foo", "bar"}, ExcludeTools: []string{"foo"}},
			tool:        "foo",
			wantAllowed: false,
		},
		{
			name:        "Tool not in either, allowed (no filters)",
			conn:        MCPServerConnection{},
			tool:        "baz",
			wantAllowed: true,
		},
		{
			name:        "Empty tool name, no filters, allowed",
			conn:        MCPServerConnection{},
			tool:        "",
			wantAllowed: true,
		},
		{
			name:        "Empty tool name, in ExcludeTools, disallowed",
			conn:        MCPServerConnection{ExcludeTools: []string{""}},
			tool:        "",
			wantAllowed: false,
		},
		{
			name:        "Empty tool name, in IncludeTools, allowed",
			conn:        MCPServerConnection{IncludeTools: []string{""}},
			tool:        "",
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.conn.IsToolAllowed(tt.tool)
			if allowed != tt.wantAllowed {
				t.Errorf("IsToolAllowed(%q) = %v, want %v", tt.tool, allowed, tt.wantAllowed)
			}
		})
	}
}
