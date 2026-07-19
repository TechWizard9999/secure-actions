package elicit

import (
	"context"
	"testing"

	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSessionElicitor_Elicit(t *testing.T) {
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"value": "test-secret"}},
		},
	}

	params := &mcp.ElicitParams{
		Message: "Test message",
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"value": map[string]any{"type": "string"},
			},
			"required": []string{"value"},
		},
	}

	result, err := elicitor.Elicit(context.Background(), params)
	if err != nil {
		t.Fatalf("Elicit: %v", err)
	}

	if result.Action != "accept" {
		t.Fatalf("Action = %q, want accept", result.Action)
	}
	if result.Content["value"] != "test-secret" {
		t.Fatalf("Content = %v, want test-secret", result.Content["value"])
	}
}

func TestSessionElicitor_Elicit_Cancel(t *testing.T) {
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "cancel"},
		},
	}

	params := &mcp.ElicitParams{
		Message: "Test message",
	}

	result, err := elicitor.Elicit(context.Background(), params)
	if err != nil {
		t.Fatalf("Elicit: %v", err)
	}

	if result.Action != "cancel" {
		t.Fatalf("Action = %q, want cancel", result.Action)
	}
}

func TestSessionElicitor_Elicit_Error(t *testing.T) {
	elicitor := &mocktest.MockElicitor{
		Errors: []error{context.Canceled},
	}

	params := &mcp.ElicitParams{
		Message: "Test message",
	}

	_, err := elicitor.Elicit(context.Background(), params)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestElicitor_Interface(t *testing.T) {
	var _ Elicitor = (*mocktest.MockElicitor)(nil)
	var _ Elicitor = (*SessionElicitor)(nil)
}