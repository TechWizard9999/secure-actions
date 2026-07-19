package mcp_tool_request

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kotakarthik/secure-actions/internal/mcp_config"
	"github.com/kotakarthik/secure-actions/internal/secrets"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var secretPlaceholder = regexp.MustCompile(`<<secret:([a-z0-9-]+)>>`)

type Dependencies struct {
	SecretManager secrets.Manager
	EncryptionKey []byte
}

type Handler struct {
	deps Dependencies
}

func New(deps Dependencies) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) Execute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input Request,
) (*mcp.CallToolResult, Response, error) {

	if input.MCPConfig.Type == "" {
		return nil, Response{}, fmt.Errorf("mcpConfig.type is required")
	}
	if input.Tool == "" {
		return nil, Response{}, fmt.Errorf("tool is required")
	}

	// Substitute secrets in parameters recursively
	resolvedParams, err := h.substituteSecretsInParams(ctx, input.Parameters)
	if err != nil {
		return nil, Response{Success: false, Error: err.Error()}, nil
	}

	// Call the target MCP tool
	result, err := h.callMCPTool(ctx, input.MCPConfig, input.Tool, resolvedParams)
	if err != nil {
		return nil, Response{Success: false, Error: err.Error()}, nil
	}

	return nil, Response{
		Success: true,
		Result:  result,
	}, nil
}

// substituteSecretsInParams recursively replaces <<secret:identifier>> placeholders
// with decrypted secret values in nested parameter structures
func (h *Handler) substituteSecretsInParams(ctx context.Context, params map[string]any) (map[string]any, error) {
	if params == nil {
		return nil, nil
	}

	result := make(map[string]any, len(params))
	for key, value := range params {
		resolved, err := h.substituteSecretsInValue(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("substitute %q: %w", key, err)
		}
		result[key] = resolved
	}
	return result, nil
}

// substituteSecretsInValue recursively handles different types (string, map, slice, etc.)
func (h *Handler) substituteSecretsInValue(ctx context.Context, value any) (any, error) {
	switch v := value.(type) {
	case string:
		return h.substituteSecretsInString(ctx, v)
	case map[string]any:
		resolved, err := h.substituteSecretsInParams(ctx, v)
		if err != nil {
			return nil, err
		}
		return resolved, nil
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			resolved, err := h.substituteSecretsInValue(ctx, item)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			result[i] = resolved
		}
		return result, nil
	default:
		return value, nil
	}
}

// substituteSecretsInString replaces <<secret:identifier>> with decrypted values
func (h *Handler) substituteSecretsInString(ctx context.Context, input string) (string, error) {
	if !strings.Contains(input, "<<secret:") {
		return input, nil
	}

	matches := secretPlaceholder.FindAllStringSubmatchIndex(input, -1)
	if len(matches) == 0 {
		return input, nil
	}

	var b strings.Builder
	lastEnd := 0

	for _, match := range matches {
		b.WriteString(input[lastEnd:match[0]])

		identifier := input[match[2]:match[3]]
		encrypted, found, err := h.deps.SecretManager.Get(ctx, identifier)
		if err != nil {
			return "", fmt.Errorf("get secret %q: %w", identifier, err)
		}
		if !found {
			return "", fmt.Errorf("secret %q not found", identifier)
		}

		decrypted, err := secrets.Decrypt(encrypted, h.deps.EncryptionKey)
		if err != nil {
			return "", fmt.Errorf("decrypt secret %q: %w", identifier, err)
		}

		b.WriteString(decrypted)
		lastEnd = match[1]
	}

	b.WriteString(input[lastEnd:])
	return b.String(), nil
}

// callMCPTool connects to target MCP and calls the specified tool
func (h *Handler) callMCPTool(ctx context.Context, cfg MCPConfig, tool string, params map[string]any) (map[string]any, error) {
	log.Printf("[mcp_tool_request] calling %s with tool %q", cfg.Type, tool)

	// Substitute secrets in env vars before spawning the process
	resolvedEnv, err := h.substituteSecretsInEnv(ctx, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("substitute secrets in env: %w", err)
	}

	// Convert to mcp_config.MCPServerConfig
	mcpCfg := mcp_config.MCPServerConfig{
		Type:    cfg.Type,
		Command: cfg.Command,
		Args:    cfg.Args,
		URL:     cfg.URL,
		Env:     resolvedEnv,
	}

	// Create JSON-RPC client
	client, err := NewMCPClient(ctx, mcpCfg)
	if err != nil {
		return nil, fmt.Errorf("create MCP client: %w", err)
	}
	defer client.Close()

	// Perform MCP initialization handshake
	if err := client.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("initialize MCP: %w", err)
	}

	// Call the tool
	result, err := client.CallTool(ctx, tool, params)
	if err != nil {
		return nil, fmt.Errorf("call tool %q: %w", tool, err)
	}

	return result, nil
}

// substituteSecretsInEnv decrypts <<secret:identifier>> placeholders in environment variable values
func (h *Handler) substituteSecretsInEnv(ctx context.Context, env map[string]string) (map[string]string, error) {
	if len(env) == 0 {
		return env, nil
	}

	resolved := make(map[string]string, len(env))
	for k, v := range env {
		decrypted, err := h.substituteSecretsInString(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("env %q: %w", k, err)
		}
		resolved[k] = decrypted
	}
	return resolved, nil
}
