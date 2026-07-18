package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("MONGO_URI")

	cfg := Load()

	if cfg.MongoURI != "mongodb://localhost:27018" {
		t.Fatalf("MongoURI = %q, want default", cfg.MongoURI)
	}
	if cfg.Database != "secure_actions" {
		t.Fatalf("Database = %q, want secure_actions", cfg.Database)
	}
}

func TestLoad_CustomURI(t *testing.T) {
	os.Setenv("MONGO_URI", "mongodb://custom:12345")
	defer os.Unsetenv("MONGO_URI")

	cfg := Load()

	if cfg.MongoURI != "mongodb://custom:12345" {
		t.Fatalf("MongoURI = %q, want custom URI", cfg.MongoURI)
	}
}
