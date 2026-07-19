package mcp_tool_request

type Response struct {
	Success bool         `json:"success" jsonschema:"Whether tool call succeeded"`
	Result  map[string]any `json:"result,omitempty" jsonschema:"Tool result (varies by tool)"`
	Error   string       `json:"error,omitempty" jsonschema:"Error message if tool failed"`
}
