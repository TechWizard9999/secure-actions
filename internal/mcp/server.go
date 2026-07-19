package mcp

import (
	"context"

	"github.com/kotakarthik/secure-actions/internal/actions/delete_secret"
	"github.com/kotakarthik/secure-actions/internal/actions/http_request"
	"github.com/kotakarthik/secure-actions/internal/actions/list_secrets"
	"github.com/kotakarthik/secure-actions/internal/actions/mcp_tool_request"
	"github.com/kotakarthik/secure-actions/internal/actions/ping"
	"github.com/kotakarthik/secure-actions/internal/actions/request_secret"
	"github.com/kotakarthik/secure-actions/internal/secrets"

	mcpSdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Dependencies struct {
	SecretManager secrets.Manager
	EncryptionKey []byte
}

type Server struct {
	server *mcpSdk.Server
}

func New(deps Dependencies) *Server {

	server := mcpSdk.NewServer(
		&mcpSdk.Implementation{
			Name:    "secure-actions",
			Version: "0.0.1",
		},
		nil,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "mcp_tool_request",
			Description: "Universal gateway to call other MCP tools securely with secret injection. Pass the MCP configuration (type, command, args, env) along with the tool name and parameters. Secrets using <<secret:identifier>> placeholders are decrypted within secure-actions and never exposed to the target MCP or LLM context. Example: {\"mcpConfig\": {\"type\": \"stdio\", \"command\": \"npx\", \"args\": [\"@modelcontextprotocol/server-github\"], \"env\": {\"GITHUB_PERSONAL_ACCESS_TOKEN\": \"<<secret:github-pat>>\"}}, \"tool\": \"create_issue\", \"parameters\": {\"owner\": \"user\", \"repo\": \"repo\", \"title\": \"Issue\"}}",
		},
		mcp_tool_request.New(mcp_tool_request.Dependencies{
			SecretManager: deps.SecretManager,
			EncryptionKey: deps.EncryptionKey,
		}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "ping",
			Description: "Health check tool. Returns pong to verify the MCP server is running and responsive.",
		},
		ping.New(ping.Dependencies{}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "request_secret",
			Description: "Securely collect and store a secret from the user. Prompts the user via an interactive form to enter a secret value (e.g. API key, token, password). The value is encrypted with AES-256-GCM before being persisted to the database. The identifier is normalized to lowercase with hyphens (e.g. 'My Token' becomes 'my-token'). If the identifier already exists, the user is prompted to confirm before the existing secret is updated. Use this before http_request when authentication credentials are needed.",
		},
		request_secret.New(request_secret.Dependencies{
			SecretManager: deps.SecretManager,
			EncryptionKey: deps.EncryptionKey,
		}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "list_secrets",
			Description: "List all stored secret identifiers. Returns the names of all secrets currently stored in the database. Does not expose secret values — only identifiers are returned. Use this to check which secrets are available before making an http_request with <<secret:identifier>> placeholders.",
		},
		list_secrets.New(list_secrets.Dependencies{
			SecretManager: deps.SecretManager,
		}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "delete_secret",
			Description: "Permanently delete a stored secret by its identifier. Prompts the user for confirmation before deletion. The secret is removed from the database and cannot be recovered. Use list_secrets first to verify the identifier exists.",
		},
		delete_secret.New(delete_secret.Dependencies{
			SecretManager: deps.SecretManager,
		}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "http_request",
			Description: "Execute an HTTP request with optional secret injection. Supports GET, POST, PUT, PATCH, and DELETE methods. Secrets stored via request_secret can be injected into the URL, headers, or body using the <<secret:identifier>> placeholder syntax. Secrets are decrypted at request time and never logged or exposed in responses. Example header: {\"Authorization\": \"Bearer <<secret:github-pat>>\"}.",
		},
		http_request.New(http_request.Dependencies{
			SecretManager: deps.SecretManager,
			EncryptionKey: deps.EncryptionKey,
		}).Execute,
	)

	return &Server{
		server: server,
	}
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcpSdk.StdioTransport{})
}
