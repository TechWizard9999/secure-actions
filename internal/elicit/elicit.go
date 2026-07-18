package elicit

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Elicitor interface {
	Elicit(ctx context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error)
}

type SessionElicitor struct {
	Session *mcp.ServerSession
}

func (s *SessionElicitor) Elicit(ctx context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	return s.Session.Elicit(ctx, params)
}
