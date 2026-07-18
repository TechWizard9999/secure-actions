package mcp

import (
	"context"

	"github.com/kotakarthik/secure-actions/internal/actions/delete_secret"
	"github.com/kotakarthik/secure-actions/internal/actions/list_secrets"
	"github.com/kotakarthik/secure-actions/internal/actions/ping"
	"github.com/kotakarthik/secure-actions/internal/actions/request_secret"
	"github.com/kotakarthik/secure-actions/internal/secrets"

	mcpSdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Dependencies struct {
	SecretManager secrets.Manager
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
			Name:        "ping",
			Description: "Returns pong",
		},
		ping.New(ping.Dependencies{}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "request_secret",
			Description: "Prompts the user for a secret value, encrypts it, and stores it for later use",
		},
		request_secret.New(request_secret.Dependencies{
			SecretManager: deps.SecretManager,
		}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "list_secrets",
			Description: "Lists all stored secret identifiers",
		},
		list_secrets.New(list_secrets.Dependencies{
			SecretManager: deps.SecretManager,
		}).Execute,
	)

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "delete_secret",
			Description: "Deletes a stored secret by its identifier",
		},
		delete_secret.New(delete_secret.Dependencies{
			SecretManager: deps.SecretManager,
		}).Execute,
	)

	return &Server{
		server: server,
	}
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcpSdk.StdioTransport{})
}
