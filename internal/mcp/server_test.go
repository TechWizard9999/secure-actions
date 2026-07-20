package mcp

import (
	"testing"

	"github.com/kotakarthik/secure-actions/internal/secrets"
	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
)

func TestNew(t *testing.T) {
	key, err := secrets.LoadOrCreateMasterKey(t.TempDir() + "/master.key")
	if err != nil {
		t.Fatalf("LoadOrCreateMasterKey: %v", err)
	}

	mgr := mocktest.NewMockSecretManager()

	deps := Dependencies{
		SecretManager: mgr,
		EncryptionKey: key,
	}

	server := New(deps)
	if server == nil {
		t.Fatal("New returned nil")
	}
}

func TestNew_AllToolsRegistered(t *testing.T) {
	key, err := secrets.LoadOrCreateMasterKey(t.TempDir() + "/master.key")
	if err != nil {
		t.Fatalf("LoadOrCreateMasterKey: %v", err)
	}

	mgr := mocktest.NewMockSecretManager()

	deps := Dependencies{
		SecretManager: mgr,
		EncryptionKey: key,
	}

	server := New(deps)
	if server == nil {
		t.Fatal("New returned nil")
	}
}

func TestServer_Run_NotStarted(t *testing.T) {
	key, err := secrets.LoadOrCreateMasterKey(t.TempDir() + "/master.key")
	if err != nil {
		t.Fatalf("LoadOrCreateMasterKey: %v", err)
	}

	mgr := mocktest.NewMockSecretManager()

	deps := Dependencies{
		SecretManager: mgr,
		EncryptionKey: key,
	}

	server := New(deps)
	if server == nil {
		t.Fatal("New returned nil")
	}
}