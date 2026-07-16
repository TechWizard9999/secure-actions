package ping

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Request struct{}

type Response struct {
	Message string `json:"message" jsonschema:"The pong response"`
}

type Handler struct{}

func New() *Handler {
	return &Handler{}
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