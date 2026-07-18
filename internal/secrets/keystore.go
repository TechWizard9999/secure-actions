package secrets

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const keySize = 32 // AES-256

// LoadOrCreateMasterKey reads the master key from path, or generates a new
// random 256-bit key and writes it there if the file doesn't exist.
// The file is created with 0600 permissions inside a 0700 directory.
func LoadOrCreateMasterKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return parseKeyFile(data)
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read master key: %w", err)
	}

	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate master key: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create key directory: %w", err)
	}

	encoded := hex.EncodeToString(key)
	if err := os.WriteFile(path, []byte(encoded+"\n"), 0o600); err != nil {
		return nil, fmt.Errorf("write master key: %w", err)
	}

	return key, nil
}

func parseKeyFile(data []byte) ([]byte, error) {
	// Strip trailing newline
	s := string(data)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}

	key, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("parse master key (expected 64 hex chars): %w", err)
	}
	if len(key) != keySize {
		return nil, fmt.Errorf("master key must be %d bytes, got %d", keySize, len(key))
	}
	return key, nil
}
