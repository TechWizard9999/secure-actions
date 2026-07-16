package mcp

import (
	"context"

	"github.com/kotakarthik/secure-actions/internal/actions/ping"

	mcpSdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	server *mcpSdk.Server
}

func New() *Server {

	server := mcpSdk.NewServer(
		&mcpSdk.Implementation{
			Name:    "secure-actions",
			Version: "0.0.1",
		},
		nil,
	)

	pingHandler := ping.New()

	mcpSdk.AddTool(
		server,
		&mcpSdk.Tool{
			Name:        "ping",
			Description: "Returns pong",
		},
		pingHandler.Execute,
	)

	return &Server{
		server: server,
	}
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcpSdk.StdioTransport{})
}