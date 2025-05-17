package mcp_connector

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// testableTimeoutSelect is a helper for testing manual timeout logic in ExecuteTool.
// It races a result and error channel against a timer, returning which event occurred.
func testableTimeoutSelect(resultCh <-chan *mcp.CallToolResult, errCh <-chan error, timeout time.Duration, cancel context.CancelFunc) (result *mcp.CallToolResult, err error, timedOut bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case result = <-resultCh:
		return result, nil, false
	case err = <-errCh:
		return nil, err, false
	case <-timer.C:
		cancel()
		return nil, nil, true
	}
}
