package list_secrets

import (
	"context"
	"sort"
	"testing"

	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
)

type nilKeysManager struct {
	mocktest.MockSecretManager
}

func (m *nilKeysManager) Keys(_ context.Context) ([]string, error) {
	return nil, nil
}

func TestListSecrets_NilKeys(t *testing.T) {
	h := New(Dependencies{SecretManager: &nilKeysManager{}})

	_, resp, err := h.Execute(context.Background(), nil, Request{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Secrets == nil {
		t.Fatal("secrets should not be nil")
	}
	if resp.Count != 0 {
		t.Fatalf("count = %d, want 0", resp.Count)
	}
}

func TestListSecrets_Empty(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr})

	_, resp, err := h.Execute(context.Background(), nil, Request{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Count != 0 {
		t.Fatalf("count = %d, want 0", resp.Count)
	}
	if len(resp.Secrets) != 0 {
		t.Fatalf("secrets = %v, want empty", resp.Secrets)
	}
}

func TestListSecrets_WithSecrets(t *testing.T) {
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()
	mgr.Set(ctx, "api-key", "encrypted1")
	mgr.Set(ctx, "db-pass", "encrypted2")
	mgr.Set(ctx, "token", "encrypted3")

	h := New(Dependencies{SecretManager: mgr})

	_, resp, err := h.Execute(ctx, nil, Request{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Count != 3 {
		t.Fatalf("count = %d, want 3", resp.Count)
	}

	sort.Strings(resp.Secrets)
	expected := []string{"api-key", "db-pass", "token"}
	for i, s := range resp.Secrets {
		if s != expected[i] {
			t.Fatalf("secrets[%d] = %q, want %q", i, s, expected[i])
		}
	}
}
