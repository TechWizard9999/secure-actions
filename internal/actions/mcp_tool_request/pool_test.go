package mcp_tool_request

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/kotakarthik/secure-actions/internal/mcp_config"
)

func mockFactory(_ context.Context, _ mcp_config.MCPServerConfig) (*MCPClient, error) {
	_, clientW := io.Pipe()
	serverR, _ := io.Pipe()
	return newMCPClientFromPipes(clientW, serverR), nil
}

func failingFactory(_ context.Context, _ mcp_config.MCPServerConfig) (*MCPClient, error) {
	return nil, fmt.Errorf("connection refused")
}

func TestPoolKey_DifferentConfigs(t *testing.T) {
	cfg1 := mcp_config.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-github"},
		Env:     map[string]string{"TOKEN": "abc"},
	}
	cfg2 := mcp_config.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-slack"},
		Env:     map[string]string{"TOKEN": "abc"},
	}

	if poolKey(cfg1) == poolKey(cfg2) {
		t.Fatal("different configs should produce different keys")
	}
}

func TestPoolKey_SameConfig(t *testing.T) {
	cfg := mcp_config.MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "server"},
		Env:     map[string]string{"A": "1", "B": "2"},
	}

	if poolKey(cfg) != poolKey(cfg) {
		t.Fatal("same config should produce same key")
	}
}

func TestPoolKey_EnvValueIgnored(t *testing.T) {
	cfg1 := mcp_config.MCPServerConfig{
		Command: "npx",
		Args:    []string{"server"},
		Env:     map[string]string{"TOKEN": "value1"},
	}
	cfg2 := mcp_config.MCPServerConfig{
		Command: "npx",
		Args:    []string{"server"},
		Env:     map[string]string{"TOKEN": "value2"},
	}

	if poolKey(cfg1) != poolKey(cfg2) {
		t.Fatal("pool key should ignore env values, only use keys")
	}
}

func TestPool_NewPool(t *testing.T) {
	pool := NewMCPProcessPool()
	if pool.Size() != 0 {
		t.Fatalf("new pool should be empty, got size %d", pool.Size())
	}
}

func TestPool_GetCachesClient(t *testing.T) {
	pool := newTestPool(mockFactory)
	defer pool.CloseAll()

	cfg := mcp_config.MCPServerConfig{Type: "stdio", Command: "echo"}

	ctx := context.Background()
	client1, err := pool.Get(ctx, cfg)
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}

	client2, err := pool.Get(ctx, cfg)
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}

	if client1 != client2 {
		t.Fatal("second Get should return cached client")
	}

	if pool.Size() != 1 {
		t.Fatalf("pool should have 1 entry, got %d", pool.Size())
	}
}

func TestPool_EvictRemovesClient(t *testing.T) {
	pool := newTestPool(mockFactory)
	defer pool.CloseAll()

	cfg := mcp_config.MCPServerConfig{Type: "stdio", Command: "echo"}

	_, err := pool.Get(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	pool.Evict(poolKey(cfg))

	if pool.Size() != 0 {
		t.Fatalf("pool should be empty after evict, got %d", pool.Size())
	}
}

func TestPool_EvictNonexistent(t *testing.T) {
	pool := newTestPool(mockFactory)
	pool.Evict("nonexistent-key")
}

func TestPool_MaxSizeEvictsOldest(t *testing.T) {
	pool := &MCPProcessPool{
		clients:     make(map[string]*poolEntry),
		maxSize:     2,
		idleTimeout: defaultIdleTimeout,
		factory:     mockFactory,
	}
	defer pool.CloseAll()

	ctx := context.Background()

	cfgs := []mcp_config.MCPServerConfig{
		{Type: "stdio", Command: "echo", Args: []string{"a"}},
		{Type: "stdio", Command: "echo", Args: []string{"b"}},
		{Type: "stdio", Command: "echo", Args: []string{"c"}},
	}

	for _, cfg := range cfgs[:2] {
		_, err := pool.Get(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	if pool.Size() != 2 {
		t.Fatalf("expected 2, got %d", pool.Size())
	}

	_, err := pool.Get(ctx, cfgs[2])
	if err != nil {
		t.Fatal(err)
	}

	if pool.Size() != 2 {
		t.Fatalf("expected 2 after eviction, got %d", pool.Size())
	}
}

func TestPool_CloseAll(t *testing.T) {
	pool := newTestPool(mockFactory)

	ctx := context.Background()
	cfgs := []mcp_config.MCPServerConfig{
		{Type: "stdio", Command: "echo", Args: []string{"1"}},
		{Type: "stdio", Command: "echo", Args: []string{"2"}},
	}

	for _, cfg := range cfgs {
		_, err := pool.Get(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
	}

	pool.CloseAll()

	if pool.Size() != 0 {
		t.Fatalf("expected 0 after CloseAll, got %d", pool.Size())
	}
}

func TestPool_GetFactoryError(t *testing.T) {
	pool := newTestPool(failingFactory)
	defer pool.CloseAll()

	cfg := mcp_config.MCPServerConfig{Type: "stdio", Command: "echo"}

	_, err := pool.Get(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error from failing factory")
	}

	if pool.Size() != 0 {
		t.Fatalf("pool should be empty on factory error, got %d", pool.Size())
	}
}

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		msg       string
		transient bool
	}{
		{"broken pipe", true},
		{"read from stdout: EOF", true},
		{"Premature close of connection", true},
		{"connection reset by peer", true},
		{"process exited unexpectedly", true},
		{"MCP error: Method not found (-32601)", false},
		{"tool not found", false},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.msg)
			if isTransientError(err) != tt.transient {
				t.Fatalf("isTransientError(%q) = %v, want %v", tt.msg, !tt.transient, tt.transient)
			}
		})
	}
}
