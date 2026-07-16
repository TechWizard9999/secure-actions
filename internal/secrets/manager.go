package secrets

import "context"

// Manager is the interface actions use to store and retrieve secrets.
type Manager interface {
	Set(ctx context.Context, name, encryptedValue string) error
	Get(ctx context.Context, name string) (encryptedValue string, found bool, err error)
	Delete(ctx context.Context, name string) error
	Keys(ctx context.Context) ([]string, error)
}
