package testing

import (
	"context"
	"sync"
)

type MockSecretManager struct {
	mu      sync.RWMutex
	secrets map[string]string
}

func NewMockSecretManager() *MockSecretManager {
	return &MockSecretManager{secrets: make(map[string]string)}
}

func (m *MockSecretManager) Set(_ context.Context, name, encryptedValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[name] = encryptedValue
	return nil
}

func (m *MockSecretManager) Get(_ context.Context, name string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.secrets[name]
	return val, ok, nil
}

func (m *MockSecretManager) Delete(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.secrets, name)
	return nil
}

func (m *MockSecretManager) Keys(_ context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.secrets))
	for k := range m.secrets {
		keys = append(keys, k)
	}
	return keys, nil
}
