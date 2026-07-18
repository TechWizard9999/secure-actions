package secrets

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateMasterKey_CreatesNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "master.key")

	key, err := LoadOrCreateMasterKey(path)
	if err != nil {
		t.Fatalf("LoadOrCreateMasterKey: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("key length = %d, want 32", len(key))
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat key file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %o, want 600", info.Mode().Perm())
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat key dir: %v", err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("dir permissions = %o, want 700", dirInfo.Mode().Perm())
	}
}

func TestLoadOrCreateMasterKey_LoadsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "master.key")

	key, err := LoadOrCreateMasterKey(path)
	if err != nil {
		t.Fatal(err)
	}

	key2, err := LoadOrCreateMasterKey(path)
	if err != nil {
		t.Fatal(err)
	}

	if hex.EncodeToString(key) != hex.EncodeToString(key2) {
		t.Fatal("second load returned different key")
	}
}

func TestLoadOrCreateMasterKey_InvalidHex(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "master.key")

	os.WriteFile(path, []byte("not-hex-at-all\n"), 0o600)

	_, err := LoadOrCreateMasterKey(path)
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestLoadOrCreateMasterKey_WrongLength(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "master.key")

	os.WriteFile(path, []byte("aabbccdd\n"), 0o600)

	_, err := LoadOrCreateMasterKey(path)
	if err == nil {
		t.Fatal("expected error for wrong key length")
	}
}

func TestParseKeyFile_ValidKey(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	encoded := hex.EncodeToString(raw) + "\n"

	key, err := parseKeyFile([]byte(encoded))
	if err != nil {
		t.Fatalf("parseKeyFile: %v", err)
	}
	if hex.EncodeToString(key) != hex.EncodeToString(raw) {
		t.Fatal("parsed key doesn't match")
	}
}

func TestParseKeyFile_NoTrailingNewline(t *testing.T) {
	raw := make([]byte, 32)
	encoded := hex.EncodeToString(raw)

	key, err := parseKeyFile([]byte(encoded))
	if err != nil {
		t.Fatalf("parseKeyFile: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("key length = %d, want 32", len(key))
	}
}
