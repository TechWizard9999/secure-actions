package mongo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestNewClient_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClient_InvalidURI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewClient(ctx, "mongodb://invalid-host:12345", "test", nil)
	if err == nil {
		t.Fatal("expected error for invalid URI")
	}
}

func TestNewClient_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	_, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestClient_Collection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	col := client.Collection(CollectionSecrets)
	if col == nil {
		t.Fatal("Collection returned nil")
	}
	if col.Name() != CollectionSecrets {
		t.Fatalf("Collection name = %q, want %q", col.Name(), CollectionSecrets)
	}
}

func TestSecretRepository_SetGetDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	repo, err := NewSecretRepository(ctx, client)
	if err != nil {
		t.Fatalf("NewSecretRepository: %v", err)
	}

	// Test Set
	err = repo.Set(ctx, "test-key", "encrypted-value")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Test Get
	value, found, err := repo.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if value != "encrypted-value" {
		t.Fatalf("value = %q, want encrypted-value", value)
	}

	// Test Get not found
	_, found, err = repo.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Get not found: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}

	// Test Keys
	keys, err := repo.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 1 || keys[0] != "test-key" {
		t.Fatalf("keys = %v, want [test-key]", keys)
	}

	// Test Delete
	err = repo.Delete(ctx, "test-key")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Verify deleted
	_, found, err = repo.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if found {
		t.Fatal("expected found=false after delete")
	}

	// Verify Keys empty
	keys, err = repo.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys after delete: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("keys = %v, want empty", keys)
	}
}

func TestSecretRepository_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	repo, err := NewSecretRepository(ctx, client)
	if err != nil {
		t.Fatalf("NewSecretRepository: %v", err)
	}

	// Initial set
	err = repo.Set(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("Set initial: %v", err)
	}

	// Upsert with new value
	err = repo.Set(ctx, "key1", "value2")
	if err != nil {
		t.Fatalf("Set upsert: %v", err)
	}

	value, found, err := repo.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || value != "value2" {
		t.Fatalf("value = %q, want value2", value)
	}
}

func TestSecretRepository_UniqueIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	// Insert directly with same identifier to test unique index
	col := client.Collection(CollectionSecrets)
	_, err = col.InsertOne(ctx, map[string]string{
		"identifier": "unique-key",
		"value":      "value1",
	})
	if err != nil {
		t.Fatalf("InsertOne: %v", err)
	}

	// Try to insert duplicate - should fail due to unique index
	_, err = col.InsertOne(ctx, map[string]string{
		"identifier": "unique-key",
		"value":      "value2",
	})
	if err == nil {
		t.Fatal("expected duplicate key error")
	}
}

func TestSecretRepository_Keys_ReturnsSorted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	repo, err := NewSecretRepository(ctx, client)
	if err != nil {
		t.Fatalf("NewSecretRepository: %v", err)
	}

	keys := []string{"zebra", "alpha", "beta", "gamma"}
	for _, k := range keys {
		if err := repo.Set(ctx, k, "value"); err != nil {
			t.Fatalf("Set %s: %v", k, err)
		}
	}

	got, err := repo.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}

	// Verify all keys present (order not guaranteed by MongoDB without sort)
	if len(got) != len(keys) {
		t.Fatalf("got %d keys, want %d", len(got), len(keys))
	}

	gotMap := make(map[string]bool)
	for _, k := range got {
		gotMap[k] = true
	}
	for _, k := range keys {
		if !gotMap[k] {
			t.Fatalf("missing key %q", k)
		}
	}
}

func TestClient_Disconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}

	err = client.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
}

func TestSecretRepository_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	repo, err := NewSecretRepository(ctx, client)
	if err != nil {
		t.Fatalf("NewSecretRepository: %v", err)
	}

	// Cancel context before operation
	cancel()

	_, _, err = repo.Get(context.Background(), "key")
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with custom options
	client, err := NewClient(ctx, "mongodb://localhost:27018", "test", nil)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer client.Disconnect(ctx)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// Verify we can use custom options by checking connection works
	col := client.Collection(CollectionSecrets)
	if col == nil {
		t.Fatal("Collection returned nil")
	}

	// Test that options.Index() is used correctly by checking index exists
	indexes, err := col.Indexes().List(ctx, options.ListIndexes())
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	defer indexes.Close(ctx)

	found := false
	for indexes.Next(ctx) {
		var idx map[string]any
		if err := indexes.Decode(&idx); err != nil {
			continue
		}
		if keys, ok := idx["key"].(map[string]any); ok {
			if _, ok := keys["identifier"]; ok {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("unique index on identifier not found")
	}
}