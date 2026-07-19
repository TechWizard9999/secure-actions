package mcp_tool_request

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/kotakarthik/secure-actions/internal/secrets"
	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return key
}

func TestExecute_MissingMCPConfig(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, _, err := h.Execute(context.Background(), nil, Request{Tool: "list"})
	if err == nil {
		t.Fatal("expected error for missing mcpConfig")
	}
}

func TestExecute_MissingTool(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, _, err := h.Execute(context.Background(), nil, Request{MCPConfig: MCPConfig{Type: "stdio", Command: "echo"}})
	if err == nil {
		t.Fatal("expected error for missing tool")
	}
}

func TestSubstituteSecretsInString_NoPlaceholders(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecretsInString(context.Background(), "plain text")
	if err != nil {
		t.Fatal(err)
	}
	if result != "plain text" {
		t.Fatalf("got %q", result)
	}
}

func TestSubstituteSecretsInString_SinglePlaceholder(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	encrypted, _ := secrets.Encrypt("secret-value-123", key)
	mgr.Set(ctx, "my-token", encrypted)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecretsInString(ctx, "Bearer <<secret:my-token>>")
	if err != nil {
		t.Fatal(err)
	}
	if result != "Bearer secret-value-123" {
		t.Fatalf("got %q", result)
	}
}

func TestSubstituteSecretsInString_MultiplePlaceholders(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc1, _ := secrets.Encrypt("user1", key)
	enc2, _ := secrets.Encrypt("pass1", key)
	mgr.Set(ctx, "username", enc1)
	mgr.Set(ctx, "password", enc2)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecretsInString(ctx, "<<secret:username>>:<<secret:password>>")
	if err != nil {
		t.Fatal(err)
	}
	if result != "user1:pass1" {
		t.Fatalf("got %q", result)
	}
}

func TestSubstituteSecretsInString_NotFound(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, err := h.substituteSecretsInString(context.Background(), "<<secret:nonexistent>>")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestSubstituteSecretsInParams_NestedMap(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	encrypted, _ := secrets.Encrypt("token123", key)
	mgr.Set(ctx, "api-token", encrypted)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	params := map[string]any{
		"url": "https://api.example.com",
		"headers": map[string]any{
			"Authorization": "Bearer <<secret:api-token>>",
		},
	}

	result, err := h.substituteSecretsInParams(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	authHeader := result["headers"].(map[string]any)["Authorization"]
	if authHeader != "Bearer token123" {
		t.Fatalf("got %q", authHeader)
	}
}

func TestSubstituteSecretsInParams_Array(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc1, _ := secrets.Encrypt("secret1", key)
	enc2, _ := secrets.Encrypt("secret2", key)
	mgr.Set(ctx, "s1", enc1)
	mgr.Set(ctx, "s2", enc2)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	params := map[string]any{
		"items": []any{
			"<<secret:s1>>",
			"plain",
			"<<secret:s2>>",
		},
	}

	result, err := h.substituteSecretsInParams(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	items := result["items"].([]any)
	if len(items) != 3 {
		t.Fatalf("got %d items", len(items))
	}
	if items[0] != "secret1" || items[1] != "plain" || items[2] != "secret2" {
		t.Fatalf("got %v", items)
	}
}

func TestSubstituteSecretsInParams_Empty(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecretsInParams(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("got %v", result)
	}
}

func TestSubstituteSecretsInValue_NonStringTypes(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	tests := []struct {
		name  string
		value any
	}{
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.substituteSecretsInValue(context.Background(), tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if result != tt.value {
				t.Fatalf("got %v", result)
			}
		})
	}
}

func TestExecute_FailsWithInvalidMCP(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	encrypted, _ := secrets.Encrypt("github-token-xyz", key)
	mgr.Set(ctx, "gh-token", encrypted)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, resp, err := h.Execute(ctx, nil, Request{
		MCPConfig: MCPConfig{
			Type:    "stdio",
			Command: "echo",
			Args:    []string{"hello"},
		},
		Tool: "list_repos",
		Parameters: map[string]any{
			"auth": "Bearer <<secret:gh-token>>",
		},
	})

	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false since echo doesn't implement MCP protocol")
	}
	if resp.Error == "" {
		t.Fatal("expected error message in response")
	}
}

func TestSubstituteSecretsInEnv_WithPlaceholders(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	encrypted, _ := secrets.Encrypt("ghp_real_token_123", key)
	mgr.Set(ctx, "github-pat", encrypted)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	env := map[string]string{
		"GITHUB_PERSONAL_ACCESS_TOKEN": "<<secret:github-pat>>",
		"NODE_ENV":                     "production",
	}

	resolved, err := h.substituteSecretsInEnv(ctx, env)
	if err != nil {
		t.Fatal(err)
	}

	if resolved["GITHUB_PERSONAL_ACCESS_TOKEN"] != "ghp_real_token_123" {
		t.Fatalf("got %q, want ghp_real_token_123", resolved["GITHUB_PERSONAL_ACCESS_TOKEN"])
	}
	if resolved["NODE_ENV"] != "production" {
		t.Fatalf("plain env var modified: got %q", resolved["NODE_ENV"])
	}
}

func TestSubstituteSecretsInEnv_Empty(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	resolved, err := h.substituteSecretsInEnv(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != nil {
		t.Fatalf("expected nil, got %v", resolved)
	}
}

func TestSubstituteSecretsInEnv_SecretNotFound(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	env := map[string]string{
		"TOKEN": "<<secret:nonexistent>>",
	}

	_, err := h.substituteSecretsInEnv(context.Background(), env)
	if err == nil {
		t.Fatal("expected error for missing secret in env")
	}
}

func TestSubstituteSecretsInParams_DecryptError(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	// Store with wrong format to cause decrypt error
	mgr.Set(ctx, "bad-secret", "not-valid-base64")

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, err := h.substituteSecretsInParams(ctx, map[string]any{
		"token": "<<secret:bad-secret>>",
	})

	if err == nil {
		t.Fatal("expected error for decrypt failure")
	}
}
