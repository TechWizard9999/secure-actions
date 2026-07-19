package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

// TLSConfig holds MongoDB TLS/SSL configuration
type TLSConfig struct {
	Enabled   bool
	CAFile    string
	CertFile  string
	KeyFile   string
	AuthDB    string
	Username  string
	Password  string
}

func NewClient(ctx context.Context, uri, database string, cfg *TLSConfig) (*Client, error) {
	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	if cfg != nil {
		if cfg.Enabled {
			tlsConfig, err := buildTLSConfig(cfg)
			if err != nil {
				return nil, fmt.Errorf("build TLS config: %w", err)
			}
			opts.SetTLSConfig(tlsConfig)
		}

		if cfg.Username != "" && cfg.Password != "" {
			opts.SetAuth(options.Credential{
				Username:   cfg.Username,
				Password:   cfg.Password,
				AuthSource: cfg.AuthDB,
			})
		}
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Client{
		client: client,
		db:     client.Database(database),
	}, nil
}

func buildTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("parse CA cert")
		}
		tlsConfig.RootCAs = caPool
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
