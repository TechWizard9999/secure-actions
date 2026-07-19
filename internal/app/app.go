package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kotakarthik/secure-actions/internal/config"
	"github.com/kotakarthik/secure-actions/internal/mcp"
	"github.com/kotakarthik/secure-actions/internal/secrets"
	mongostore "github.com/kotakarthik/secure-actions/internal/storage/mongo"
)

type App struct {
	server *mcp.Server
	mongo  *mongostore.Client
}

func New() (*App, error) {
	cfg := config.Load()

	masterKey, err := secrets.LoadOrCreateMasterKey(masterKeyPath())
	if err != nil {
		return nil, fmt.Errorf("init master key: %w", err)
	}

	ctx := context.Background()
	mongoClient, err := mongostore.NewClient(ctx, cfg.MongoURI, cfg.Database, &mongostore.TLSConfig{
		Enabled:   cfg.MongoTLS,
		CAFile:    cfg.MongoTLSCAFile,
		CertFile:  cfg.MongoCertFile,
		KeyFile:   cfg.MongoKeyFile,
		AuthDB:    cfg.MongoAuthDB,
		Username:  cfg.MongoUsername,
		Password:  cfg.MongoPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("init mongo: %w", err)
	}

	secretRepo, err := mongostore.NewSecretRepository(ctx, mongoClient)
	if err != nil {
		mongoClient.Disconnect(ctx)
		return nil, fmt.Errorf("init secret repo: %w", err)
	}

	server := mcp.New(mcp.Dependencies{
		SecretManager: secretRepo,
		EncryptionKey: masterKey,
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

func masterKeyPath() string {
	if p := os.Getenv("SECURE_ACTIONS_KEY_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".secure-actions/master.key"
	}
	return filepath.Join(home, ".secure-actions", "master.key")
}
