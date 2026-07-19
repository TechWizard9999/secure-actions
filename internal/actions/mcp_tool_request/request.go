package mcp_tool_request

type MCPConfig struct {
	Type    string            `json:"type" jsonschema:"Transport type: 'stdio' or 'http'"`
	Command string            `json:"command,omitempty" jsonschema:"Command to execute (for stdio transport)"`
	Args    []string          `json:"args,omitempty" jsonschema:"Command arguments (for stdio transport)"`
	URL     string            `json:"url,omitempty" jsonschema:"HTTP endpoint URL (for http transport)"`
	Env     map[string]string `json:"env,omitempty" jsonschema:"Environment variables to pass to the MCP process"`
}

type Request struct {
	MCPConfig  MCPConfig      `json:"mcpConfig" jsonschema:"MCP server configuration (type, command, args, env)"`
	Tool       string         `json:"tool" jsonschema:"Tool name in target MCP"`
	Parameters map[string]any `json:"parameters,omitempty" jsonschema:"Tool input parameters (string values may contain <<secret:identifier>> placeholders)"`
	TimeoutMs  int            `json:"timeoutMs,omitempty" jsonschema:"Request timeout in milliseconds (default: 30000)"`
}
