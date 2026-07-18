package request_secret

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/kotakarthik/secure-actions/internal/secrets"
	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return key
}

func TestNormalizeIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"helLo World 23", "hello-world-23"},
		{"API-KEY", "api-key"},
		{"simple", "simple"},
		{"MY TOKEN", "my-token"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeIdentifier(tt.input)
			if got != tt.want {
				t.Fatalf("normalizeIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIdentifierValidation(t *testing.T) {
	valid := []string{"api-key", "token-123", "a", "my-secret-1"}
	for _, id := range valid {
		if !validIdentifier.MatchString(id) {
			t.Fatalf("%q should be valid", id)
		}
	}

	invalid := []string{"has_underscore", "has.dot", "HAS-UPPER", "has space", "", "special!char"}
	for _, id := range invalid {
		if validIdentifier.MatchString(id) {
			t.Fatalf("%q should be invalid", id)
		}
	}
}

func TestExecute_InvalidIdentifier(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, _, err := h.Execute(context.Background(), nil, Request{Name: "bad!name"})
	if err == nil {
		t.Fatal("expected error for invalid identifier")
	}
}

func TestExecute_NewSecret_Accept(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"value": "super-secret"}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, resp, err := h.Execute(context.Background(), nil, Request{Name: "My Token"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !resp.Stored {
		t.Fatal("expected stored=true")
	}
	if resp.SecretName != "my-token" {
		t.Fatalf("secretName = %q, want my-token", resp.SecretName)
	}

	encrypted, found, _ := mgr.Get(context.Background(), "my-token")
	if !found {
		t.Fatal("secret not stored in manager")
	}
	decrypted, err := secrets.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted != "super-secret" {
		t.Fatalf("decrypted = %q, want super-secret", decrypted)
	}
}

func TestExecute_NewSecret_Cancel(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "cancel"},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, resp, err := h.Execute(context.Background(), nil, Request{Name: "my-key"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Stored {
		t.Fatal("expected stored=false")
	}
	if resp.Message != "Secret entry cancelled by user" {
		t.Fatalf("message = %q", resp.Message)
	}
}

func TestExecute_ExistingSecret_ConfirmUpdate(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc, _ := secrets.Encrypt("old-value", key)
	mgr.Set(ctx, "my-key", enc)

	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"confirm": "yes"}},
			{Action: "accept", Content: map[string]any{"value": "new-value"}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, resp, err := h.Execute(ctx, nil, Request{Name: "my-key"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !resp.Stored {
		t.Fatal("expected stored=true")
	}

	encrypted, _, _ := mgr.Get(ctx, "my-key")
	decrypted, _ := secrets.Decrypt(encrypted, key)
	if decrypted != "new-value" {
		t.Fatalf("decrypted = %q, want new-value", decrypted)
	}
}

func TestExecute_ExistingSecret_DeclineUpdate(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc, _ := secrets.Encrypt("old-value", key)
	mgr.Set(ctx, "my-key", enc)

	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"confirm": "no"}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, resp, err := h.Execute(ctx, nil, Request{Name: "my-key"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Stored {
		t.Fatal("expected stored=false")
	}
	if resp.SecretName != "my-key" {
		t.Fatalf("secretName = %q", resp.SecretName)
	}
}

func TestExecute_EmptyValue(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"value": "   "}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, _, err := h.Execute(context.Background(), nil, Request{Name: "my-key"})
	if err == nil {
		t.Fatal("expected error for empty value")
	}
}

func TestExecute_MissingValueField(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"other": "field"}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, _, err := h.Execute(context.Background(), nil, Request{Name: "my-key"})
	if err == nil {
		t.Fatal("expected error for missing value field")
	}
}

func TestExecute_ValueNotString(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"value": 12345}},
		},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, _, err := h.Execute(context.Background(), nil, Request{Name: "my-key"})
	if err == nil {
		t.Fatal("expected error for non-string value")
	}
}

func TestExecute_ElicitError(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	elicitor := &mocktest.MockElicitor{
		Errors: []error{context.Canceled},
	}

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key, Elicitor: elicitor})

	_, _, err := h.Execute(context.Background(), nil, Request{Name: "my-key"})
	if err == nil {
		t.Fatal("expected error when elicit fails")
	}
}

func TestBuildElicitMessage(t *testing.T) {
	h := New(Dependencies{})

	msg := h.buildElicitMessage(Request{Name: "my-key"})
	if msg == "" {
		t.Fatal("message should not be empty")
	}

	msg = h.buildElicitMessage(Request{
		Name:        "my-key",
		Description: "GitHub PAT",
		Prompt:      "Enter your token",
	})
	if msg == "" {
		t.Fatal("message should not be empty")
	}
}
