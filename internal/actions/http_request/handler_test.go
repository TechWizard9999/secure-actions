package http_request

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
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

func TestSubstituteSecrets_NoPlaceholders(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecrets(context.Background(), "https://api.example.com/v1")
	if err != nil {
		t.Fatal(err)
	}
	if result != "https://api.example.com/v1" {
		t.Fatalf("got %q", result)
	}
}

func TestSubstituteSecrets_EmptyString(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecrets(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if result != "" {
		t.Fatalf("got %q, want empty", result)
	}
}

func TestSubstituteSecrets_SinglePlaceholder(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	encrypted, _ := secrets.Encrypt("my-token-123", key)
	mgr.Set(ctx, "github-token", encrypted)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecrets(ctx, "Bearer <<secret:github-token>>")
	if err != nil {
		t.Fatal(err)
	}
	if result != "Bearer my-token-123" {
		t.Fatalf("got %q", result)
	}
}

func TestSubstituteSecrets_MultiplePlaceholders(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc1, _ := secrets.Encrypt("user1", key)
	enc2, _ := secrets.Encrypt("pass1", key)
	mgr.Set(ctx, "username", enc1)
	mgr.Set(ctx, "password", enc2)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	result, err := h.substituteSecrets(ctx, "<<secret:username>>:<<secret:password>>@host")
	if err != nil {
		t.Fatal(err)
	}
	if result != "user1:pass1@host" {
		t.Fatalf("got %q", result)
	}
}

func TestSubstituteSecrets_NotFound(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, err := h.substituteSecrets(context.Background(), "<<secret:nonexistent>>")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestExecute_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer tok123" {
			t.Fatalf("auth header = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc, _ := secrets.Encrypt("tok123", key)
	mgr.Set(ctx, "my-token", enc)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, resp, err := h.Execute(ctx, nil, Request{
		Method:  "GET",
		URL:     server.URL + "/test",
		Headers: map[string]string{"Authorization": "Bearer <<secret:my-token>>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if resp.Body != `{"ok":true}` {
		t.Fatalf("body = %q", resp.Body)
	}
	if resp.Headers["X-Custom"] != "value" {
		t.Fatalf("X-Custom header = %q", resp.Headers["X-Custom"])
	}
}

func TestExecute_POST_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		w.WriteHeader(201)
		w.Write([]byte("created"))
	}))
	defer server.Close()

	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, resp, err := h.Execute(context.Background(), nil, Request{
		Method: "POST",
		URL:    server.URL,
		Body:   `{"name":"test"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestExecute_InvalidMethod(t *testing.T) {
	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, _, err := h.Execute(context.Background(), nil, Request{
		Method: "INVALID",
		URL:    "http://example.com",
	})
	if err == nil {
		t.Fatal("expected error for invalid method")
	}
}

func TestExecute_SecretInURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/secret-org/repo" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	key := testKey(t)
	mgr := mocktest.NewMockSecretManager()
	ctx := context.Background()

	enc, _ := secrets.Encrypt("secret-org", key)
	mgr.Set(ctx, "org-name", enc)

	h := New(Dependencies{SecretManager: mgr, EncryptionKey: key})

	_, resp, err := h.Execute(ctx, nil, Request{
		Method: "GET",
		URL:    server.URL + "/repos/<<secret:org-name>>/repo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}
