package request_secret

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kotakarthik/secure-actions/internal/secrets"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var validIdentifier = regexp.MustCompile(`^[a-z0-9-]+$`)

type Dependencies struct {
	SecretManager secrets.Manager
	// AuditRecorder audit.Recorder
	// Logger        logger.Logger
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

	name := strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))
	if !validIdentifier.MatchString(name) {
		return nil, Response{}, fmt.Errorf("invalid identifier %q: only letters, numbers, spaces, and hyphens are allowed", input.Name)
	}

	message := h.buildElicitMessage(input)

	log.Printf("[request_secret] eliciting value for secret %q", name)

	result, err := req.Session.Elicit(ctx, &mcp.ElicitParams{
		Message: message,
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"value": map[string]any{
					"type":  "string",
					"title": "Secret value",
				},
			},
			"required": []string{"value"},
		},
	})
	if err != nil {
		log.Printf("[request_secret] elicitation error: %v", err)
		return nil, Response{}, fmt.Errorf("elicit: %w", err)
	}

	log.Printf("[request_secret] elicitation action: %q", result.Action)

	if result.Action != "accept" {
		return nil, Response{
			SecretName: name,
			Stored:     false,
			Message:    "Secret entry cancelled by user",
		}, nil
	}

	raw, ok := result.Content["value"]
	if !ok {
		return nil, Response{}, fmt.Errorf("elicitation response missing 'value' field")
	}
	value, ok := raw.(string)
	if !ok {
		return nil, Response{}, fmt.Errorf("elicitation 'value' field is not a string")
	}
	if strings.TrimSpace(value) == "" {
		return nil, Response{}, fmt.Errorf("secret value cannot be empty")
	}

	log.Printf("[request_secret] encrypting and storing secret %q", name)

	// Passphrase is derived from the identifier so it's deterministic per-name.
	// A real implementation would use a proper KMS or user-supplied master key.
	encrypted, err := secrets.Encrypt(value, name)
	if err != nil {
		log.Printf("[request_secret] encrypt error: %v", err)
		return nil, Response{}, fmt.Errorf("encrypt: %w", err)
	}

	if err := h.deps.SecretManager.Set(ctx, name, encrypted); err != nil {
		log.Printf("[request_secret] store error: %v", err)
		return nil, Response{}, fmt.Errorf("store: %w", err)
	}

	log.Printf("[request_secret] secret %q stored successfully", name)

	return nil, Response{
		SecretName: name,
		Stored:     true,
		Message:    fmt.Sprintf("Secret %q stored successfully", name),
	}, nil
}

func (h *Handler) buildElicitMessage(input Request) string {
	var b strings.Builder

	fmt.Fprintf(&b, "**Secret: `%s`**\n\n", input.Name)

	if input.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", input.Description)
	}

	if input.Prompt != "" {
		fmt.Fprintf(&b, "%s\n\n", input.Prompt)
	}

	b.WriteString("Enter the secret value below. The value will be encrypted before storage and will not be visible again.")

	return b.String()
}
