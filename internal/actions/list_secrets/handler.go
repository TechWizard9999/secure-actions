package list_secrets

import (
	"context"
	"fmt"

	"github.com/kotakarthik/secure-actions/internal/secrets"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Request struct{}

type Response struct {
	Secrets []string `json:"secrets" jsonschema:"List of secret identifiers"`
	Count   int      `json:"count" jsonschema:"Total number of secrets"`
}

type Dependencies struct {
	SecretManager secrets.Manager
}

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

	keys, err := h.deps.SecretManager.Keys(ctx)
	if err != nil {
		return nil, Response{}, fmt.Errorf("list secrets: %w", err)
	}

	if keys == nil {
		keys = []string{}
	}

	return nil, Response{
		Secrets: keys,
		Count:   len(keys),
	}, nil
}
