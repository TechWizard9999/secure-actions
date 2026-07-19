package mcp_config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ProjectConfigOnly(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir() // Use temp home dir to avoid picking up real ~/.claude.json

	// Create .mcp.json
	mcpConfig := `{
  "mcpServers": {
    "test-mcp": {
      "type": "stdio",
      "command": "node",
      "args": ["server.js"]
    }
  }
}`
	err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte(mcpConfig), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := &Loader{
		projectPath: tmpDir,
		homePath:    homeDir,
	}
	configs, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("got %d configs, want 1", len(configs))
	}

	cfg, ok := configs["test-mcp"]
	if !ok {
		t.Fatal("test-mcp config not found")
	}

	if cfg.Type != "stdio" {
		t.Fatalf("type = %q, want stdio", cfg.Type)
	}
	if cfg.Command != "node" {
		t.Fatalf("command = %q, want node", cfg.Command)
	}
}

func TestLoad_ProjectOverridesUser(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	// Create ~/.claude.json with user-scoped config
	claudeConfig := `{
  "projects": {
    "` + homeDir + `": {
      "mcpServers": {
        "github-mcp": {
          "type": "stdio",
          "command": "user-github"
        }
      }
    }
  }
}`
	err := os.WriteFile(filepath.Join(homeDir, ".claude.json"), []byte(claudeConfig), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Create project .mcp.json that overrides user config
	mcpConfig := `{
  "mcpServers": {
    "github-mcp": {
      "type": "stdio",
      "command": "project-github"
    }
  }
}`
	err = os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte(mcpConfig), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Override home directory for testing
	loader := &Loader{
		projectPath: tmpDir,
		homePath:    homeDir,
	}

	configs, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	cfg, ok := configs["github-mcp"]
	if !ok {
		t.Fatal("github-mcp not found")
	}

	// Project config should override user config
	if cfg.Command != "project-github" {
		t.Fatalf("command = %q, want project-github", cfg.Command)
	}
}

func TestLoad_NoConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	loader := &Loader{
		projectPath: tmpDir,
		homePath:    homeDir,
	}

	configs, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(configs) != 0 {
		t.Fatalf("got %d configs, want 0", len(configs))
	}
}

func TestGetServer_Found(t *testing.T) {
	tmpDir := t.TempDir()

	mcpConfig := `{
  "mcpServers": {
    "stripe-mcp": {
      "type": "http",
      "url": "https://stripe-mcp.example.com"
    }
  }
}`
	err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte(mcpConfig), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := &Loader{
		projectPath: tmpDir,
		homePath:    t.TempDir(),
	}

	cfg, err := loader.GetServer("stripe-mcp")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Type != "http" {
		t.Fatalf("type = %q, want http", cfg.Type)
	}
	if cfg.URL != "https://stripe-mcp.example.com" {
		t.Fatalf("url = %q", cfg.URL)
	}
}

func TestGetServer_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	loader := &Loader{
		projectPath: tmpDir,
		homePath:    homeDir,
	}

	_, err := loader.GetServer("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}

func TestLoad_WithEnv(t *testing.T) {
	tmpDir := t.TempDir()

	mcpConfig := `{
  "mcpServers": {
    "custom-mcp": {
      "type": "stdio",
      "command": "python",
      "args": ["server.py"],
      "env": {
        "API_KEY": "secret-key",
        "DEBUG": "true"
      }
    }
  }
}`
	err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte(mcpConfig), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := &Loader{
		projectPath: tmpDir,
		homePath:    t.TempDir(),
	}

	cfg, err := loader.GetServer("custom-mcp")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Env) != 2 {
		t.Fatalf("got %d env vars, want 2", len(cfg.Env))
	}
	if cfg.Env["API_KEY"] != "secret-key" {
		t.Fatalf("API_KEY = %q", cfg.Env["API_KEY"])
	}
	if cfg.Env["DEBUG"] != "true" {
		t.Fatalf("DEBUG = %q", cfg.Env["DEBUG"])
	}
}
