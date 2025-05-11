package agent

import (
	"context"
)

type directAgent interface {
	CallDirect(ctx context.Context, input string) (string, MetaInfo, error)
}

type DirectApp struct {
	Agent directAgent
}

func (d *DirectApp) HandleCall(ctx context.Context, input string) DirectCallResult {
	answer, meta, err := d.Agent.CallDirect(ctx, input)
	res := DirectCallResult{
		Success: err == nil,
		Result:  map[string]any{"answer": answer},
		Meta:    meta,
		Error:   DirectCallError{},
	}
	if err != nil {
		res.Success = false
		res.Result = map[string]any{"answer": ""}
		res.Error = DirectCallError{Type: "internal", Message: err.Error()}
	}
	return res
}
