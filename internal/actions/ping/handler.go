package ping

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Request struct{}

type Response struct {
	Message string `json:"message" jsonschema:"The pong response"`
}

type Dependencies struct{}

type Handler struct {
	deps Dependencies
}

func New(deps Dependencies) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) Execute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input Request,
) (*mcp.CallToolResult, Response, error) {

	return nil, Response{
		Message: "pong",
	}, nil
}