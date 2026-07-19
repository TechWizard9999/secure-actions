package testing

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMockSecretManager_SetGet(t *testing.T) {
	mgr := NewMockSecretManager()

	err := mgr.Set(context.Background(), "key1", "value1")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, found, err := mgr.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if val != "value1" {
		t.Fatalf("val = %q, want value1", val)
	}
}

func TestMockSecretManager_GetNotFound(t *testing.T) {
	mgr := NewMockSecretManager()

	_, found, err := mgr.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
}

func TestMockSecretManager_Delete(t *testing.T) {
	mgr := NewMockSecretManager()
	mgr.Set(context.Background(), "key1", "value1")

	err := mgr.Delete(context.Background(), "key1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, found, err := mgr.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected found=false after delete")
	}
}

func TestMockSecretManager_Keys(t *testing.T) {
	mgr := NewMockSecretManager()
	mgr.Set(context.Background(), "key1", "value1")
	mgr.Set(context.Background(), "key2", "value2")
	mgr.Set(context.Background(), "key3", "value3")

	keys, err := mgr.Keys(context.Background())
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("len(keys) = %d, want 3", len(keys))
	}
}

func TestMockSecretManager_ThreadSafe(t *testing.T) {
	mgr := NewMockSecretManager()

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			mgr.Set(context.Background(), "key"+string(rune(idx)), "value"+string(rune(idx)))
			_, _, _ = mgr.Get(context.Background(), "key"+string(rune(idx)))
			_, _ = mgr.Keys(context.Background())
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestMockElicitor_Elicit(t *testing.T) {
	el := &MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"value": "test"}},
		},
	}

	result, err := el.Elicit(context.Background(), &mcp.ElicitParams{})
	if err != nil {
		t.Fatalf("Elicit: %v", err)
	}
	if result.Action != "accept" {
		t.Fatalf("Action = %q, want accept", result.Action)
	}
	if result.Content["value"] != "test" {
		t.Fatalf("Content = %v, want test", result.Content["value"])
	}
}

func TestMockElicitor_Elicit_Cancel(t *testing.T) {
	el := &MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "cancel"},
		},
	}

	result, err := el.Elicit(context.Background(), &mcp.ElicitParams{})
	if err != nil {
		t.Fatalf("Elicit: %v", err)
	}
	if result.Action != "cancel" {
		t.Fatalf("Action = %q, want cancel", result.Action)
	}
}

func TestMockElicitor_Elicit_Error(t *testing.T) {
	el := &MockElicitor{
		Errors: []error{context.Canceled},
	}

	_, err := el.Elicit(context.Background(), &mcp.ElicitParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestMockElicitor_MultipleCalls(t *testing.T) {
	el := &MockElicitor{
		Results: []*mcp.ElicitResult{
			{Action: "accept", Content: map[string]any{"confirm": "yes"}},
			{Action: "accept", Content: map[string]any{"value": "secret"}},
		},
	}

	result1, err := el.Elicit(context.Background(), &mcp.ElicitParams{})
	if err != nil {
		t.Fatalf("Elicit 1: %v", err)
	}
	if result1.Content["confirm"] != "yes" {
		t.Fatalf("Confirm = %v, want yes", result1.Content["confirm"])
	}

	result2, err := el.Elicit(context.Background(), &mcp.ElicitParams{})
	if err != nil {
		t.Fatalf("Elicit 2: %v", err)
	}
	if result2.Content["value"] != "secret" {
		t.Fatalf("Value = %v, want secret", result2.Content["value"])
	}
}