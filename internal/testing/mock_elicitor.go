package testing

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MockElicitor struct {
	Calls   []*mcp.ElicitParams
	Results []*mcp.ElicitResult
	Errors  []error
	callIdx int
}

func (m *MockElicitor) Elicit(_ context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	m.Calls = append(m.Calls, params)
	idx := m.callIdx
	m.callIdx++

	if idx < len(m.Errors) && m.Errors[idx] != nil {
		return nil, m.Errors[idx]
	}
	if idx < len(m.Results) {
		return m.Results[idx], nil
	}
	return &mcp.ElicitResult{Action: "cancel"}, nil
}
