package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kotakarthik/secure-actions/internal/mcp"
	"github.com/kotakarthik/secure-actions/internal/secrets"
	mocktest "github.com/kotakarthik/secure-actions/internal/testing"
)

func TestMasterKeyPath_Default(t *testing.T) {
	os.Unsetenv("SECURE_ACTIONS_KEY_PATH")
	
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/home/testuser")
	defer os.Setenv("HOME", oldHome)

	path := masterKeyPath()
	expected := "/home/testuser/.secure-actions/master.key"
	if path != expected {
		t.Fatalf("masterKeyPath() = %q, want %q", path, expected)
	}
}

func TestMasterKeyPath_EnvOverride(t *testing.T) {
	os.Setenv("SECURE_ACTIONS_KEY_PATH", "/custom/path/key")
	defer os.Unsetenv("SECURE_ACTIONS_KEY_PATH")

	path := masterKeyPath()
	if path != "/custom/path/key" {
		t.Fatalf("masterKeyPath() = %q, want /custom/path/key", path)
	}
}

func TestNew_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	app, err := New()
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer app.Run()
}

func TestNew_WithMockDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "master.key")

	key, err := secrets.LoadOrCreateMasterKey(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreateMasterKey: %v", err)
	}

	mgr := mocktest.NewMockSecretManager()

	server := mcp.New(mcp.Dependencies{
		SecretManager: mgr,
		EncryptionKey: key,
	})

	if server == nil {
		t.Fatal("mcp.New returned nil")
	}
}