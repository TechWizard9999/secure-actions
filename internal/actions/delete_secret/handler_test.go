package delete_secret

import (
	"context"
	"testing"

	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestExecute_SecretNotFound(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{}
	h := New(Dependencies{SecretManager: mgr, Elicitor: elicitor})

	_, resp, err := h.Execute(context.Background(), nil, Request{Name: "nonexistent"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Deleted {
		t.Fatal("expected deleted=false")
	}
	if resp.SecretName != "nonexistent" {
		t.Fatalf("secretName = %q", resp.SecretName)
	}
}

func TestExecute_ConfirmDelete(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()
	mgr.Set(ctx, "my-secret", "encrypted-value")

	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"confirm": "yes"}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, Elicitor: elicitor})

	_, resp, err := h.Execute(ctx, nil, Request{Name: "my-secret"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !resp.Deleted {
		t.Fatal("expected deleted=true")
	}

	_, found, _ := mgr.Get(ctx, "my-secret")
	if found {
		t.Fatal("secret should have been deleted from manager")
	}
}

func TestExecute_DeclineDelete(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()
	mgr.Set(ctx, "my-secret", "encrypted-value")

	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"confirm": "no"}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, Elicitor: elicitor})

	_, resp, err := h.Execute(ctx, nil, Request{Name: "my-secret"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Deleted {
		t.Fatal("expected deleted=false")
	}
	if resp.Message != "Deletion cancelled by user" {
		t.Fatalf("message = %q", resp.Message)
	}

	_, found, _ := mgr.Get(ctx, "my-secret")
	if !found {
		t.Fatal("secret should still exist")
	}
}

func TestExecute_CancelElicitation(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()
	mgr.Set(ctx, "my-secret", "encrypted-value")

	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "cancel"},
		},
	}

	h := New(Dependencies{SecretManager: mgr, Elicitor: elicitor})

	_, resp, err := h.Execute(ctx, nil, Request{Name: "my-secret"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Deleted {
		t.Fatal("expected deleted=false on cancel")
	}
}

func TestExecute_ElicitError(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()
	mgr.Set(ctx, "my-secret", "encrypted-value")

	elicitor := &mocktest.MockElicitor{
		Errors: []error{context.Canceled},
	}

	h := New(Dependencies{SecretManager: mgr, Elicitor: elicitor})

	_, _, err := h.Execute(ctx, nil, Request{Name: "my-secret"})
	if err == nil {
		t.Fatal("expected error when elicit fails")
	}
}
