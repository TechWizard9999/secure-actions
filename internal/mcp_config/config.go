package mcp_config

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
)

// MCPServerConfig represents an MCP server configuration
type MCPServerConfig struct {
	Type    string            `json:"type"`    // "stdio", "http", etc.
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Config holds all MCP server configurations
type Config struct {
	Servers map[string]MCPServerConfig `json:"mcpServers"`
}

// Loader loads MCP configurations from project and user scopes
type Loader struct {
	projectPath string
	homePath    string
}

// NewLoader creates a new MCP config loader
func NewLoader(projectPath string) *Loader {
	home, _ := os.UserHomeDir()
	return &Loader{
		projectPath: projectPath,
		homePath:    home,
	}
}

// Load reads MCP configurations from both project (.mcp.json) and user (~/.claude.json) scopes
// Priority: project scope (.mcp.json) > user scope (~/.claude.json)
func (l *Loader) Load() (map[string]MCPServerConfig, error) {
	result := make(map[string]MCPServerConfig)

	// Load user-scoped config first (~/.claude.json)
	userConfig, err := l.loadUserConfig()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load user config: %w", err)
	}
	if len(userConfig) > 0 {
		maps.Copy(result, userConfig)
	}

	// Load project-scoped config and override user config (.mcp.json)
	projectConfig, err := l.loadProjectConfig()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load project config: %w", err)
	}
	if len(projectConfig) > 0 {
		maps.Copy(result, projectConfig)
	}

	return result, nil
}

// loadProjectConfig reads .mcp.json from the project directory
func (l *Loader) loadProjectConfig() (map[string]MCPServerConfig, error) {
	path := filepath.Join(l.projectPath, ".mcp.json")
	return l.readConfigFile(path)
}

// loadUserConfig reads mcpServers from ~/.claude.json
func (l *Loader) loadUserConfig() (map[string]MCPServerConfig, error) {
	path := filepath.Join(l.homePath, ".claude.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var claudeConfig struct {
		Projects map[string]struct {
			MCPServers map[string]MCPServerConfig `json:"mcpServers"`
		} `json:"projects"`
	}

	if err := json.Unmarshal(data, &claudeConfig); err != nil {
		return nil, fmt.Errorf("parse ~/.claude.json: %w", err)
	}

	// Return mcpServers from the root user directory project (which is the default scope)
	if rootProject, ok := claudeConfig.Projects[l.homePath]; ok {
		return rootProject.MCPServers, nil
	}

	return nil, nil
}

// readConfigFile reads and parses .mcp.json
func (l *Loader) readConfigFile(path string) (map[string]MCPServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return config.Servers, nil
}

// GetServer returns a specific MCP server configuration
func (l *Loader) GetServer(name string) (MCPServerConfig, error) {
	configs, err := l.Load()
	if err != nil {
		return MCPServerConfig{}, err
	}

	cfg, ok := configs[name]
	if !ok {
		return MCPServerConfig{}, fmt.Errorf("MCP server %q not found", name)
	}

	return cfg, nil
}
