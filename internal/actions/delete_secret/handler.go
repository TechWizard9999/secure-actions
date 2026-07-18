package delete_secret

import (
	"context"
	"fmt"

	"github.com/kotakarthik/secure-actions/internal/elicit"
	"github.com/kotakarthik/secure-actions/internal/secrets"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Request struct {
	Name string `json:"name" jsonschema:"Identifier of the secret to delete"`
}

type Response struct {
	SecretName string `json:"secretName" jsonschema:"Identifier of the deleted secret"`
	Deleted    bool   `json:"deleted" jsonschema:"Whether the secret was deleted"`
	Message    string `json:"message" jsonschema:"Status message"`
}

type Dependencies struct {
	SecretManager secrets.Manager
	Elicitor      elicit.Elicitor
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

	_, found, err := h.deps.SecretManager.Get(ctx, input.Name)
	if err != nil {
		return nil, Response{}, fmt.Errorf("lookup secret: %w", err)
	}
	if !found {
		return nil, Response{
			SecretName: input.Name,
			Deleted:    false,
			Message:    fmt.Sprintf("Secret %q not found", input.Name),
		}, nil
	}

	e := h.elicitor(req)

	result, err := e.Elicit(ctx, &mcp.ElicitParams{
		Message: fmt.Sprintf("Are you sure you want to delete secret %q? This action cannot be undone.", input.Name),
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"confirm": map[string]any{
					"type":  "string",
					"title": "Confirm deletion",
					"enum":  []string{"yes", "no"},
				},
			},
			"required": []string{"confirm"},
		},
	})
	if err != nil {
		return nil, Response{}, fmt.Errorf("elicit confirmation: %w", err)
	}

	if result.Action != "accept" || result.Content["confirm"] != "yes" {
		return nil, Response{
			SecretName: input.Name,
			Deleted:    false,
			Message:    "Deletion cancelled by user",
		}, nil
	}

	if err := h.deps.SecretManager.Delete(ctx, input.Name); err != nil {
		return nil, Response{}, fmt.Errorf("delete secret: %w", err)
	}

	return nil, Response{
		SecretName: input.Name,
		Deleted:    true,
		Message:    fmt.Sprintf("Secret %q deleted successfully", input.Name),
	}, nil
}

func (h *Handler) elicitor(req *mcp.CallToolRequest) elicit.Elicitor {
	if h.deps.Elicitor != nil {
		return h.deps.Elicitor
	}
	return &elicit.SessionElicitor{Session: req.Session}
}
