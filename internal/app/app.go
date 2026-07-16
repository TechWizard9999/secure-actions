package app

import (
	"context"
	"fmt"

	"github.com/kotakarthik/secure-actions/internal/config"
	"github.com/kotakarthik/secure-actions/internal/mcp"
	mongostore "github.com/kotakarthik/secure-actions/internal/storage/mongo"
)

type App struct {
	server *mcp.Server
	mongo  *mongostore.Client
}

func New() (*App, error) {
	cfg := config.Load()

	ctx := context.Background()
	mongoClient, err := mongostore.NewClient(ctx, cfg.MongoURI, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init mongo: %w", err)
	}

	secretRepo := mongostore.NewSecretRepository(mongoClient)

	server := mcp.New(mcp.Dependencies{
		SecretManager: secretRepo,
	})

	return &App{
		server: server,
		mongo:  mongoClient,
	}, nil
}

func (a *App) Run() error {
	ctx := context.Background()
	defer a.mongo.Disconnect(ctx)
	return a.server.Run(ctx)
}
