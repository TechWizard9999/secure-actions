package app

import (
	"context"

	"github.com/kotakarthik/secure-actions/internal/mcp"
)

type App struct {
	server *mcp.Server
}

func New() *App {

	return &App{
		server: mcp.New(),
	}
}

func (a *App) Run() error {
	return a.server.Run(context.Background())
}