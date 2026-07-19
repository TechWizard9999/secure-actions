package mcp_tool_request

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kotakarthik/secure-actions/internal/mcp_config"
)

const (
	defaultIdleTimeout = 60 * time.Second
	defaultMaxPoolSize = 5
)

type ClientFactory func(ctx context.Context, cfg mcp_config.MCPServerConfig) (*MCPClient, error)

func defaultClientFactory(ctx context.Context, cfg mcp_config.MCPServerConfig) (*MCPClient, error) {
	client, err := NewMCPClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Initialize(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("initialize MCP: %w", err)
	}
	return client, nil
}

type poolEntry struct {
	client    *MCPClient
	lastUsed  time.Time
	idleTimer *time.Timer
}

type MCPProcessPool struct {
	mu          sync.Mutex
	clients     map[string]*poolEntry
	maxSize     int
	idleTimeout time.Duration
	factory     ClientFactory
}

func NewMCPProcessPool() *MCPProcessPool {
	return &MCPProcessPool{
		clients:     make(map[string]*poolEntry),
		maxSize:     defaultMaxPoolSize,
		idleTimeout: defaultIdleTimeout,
		factory:     defaultClientFactory,
	}
}

func newTestPool(factory ClientFactory) *MCPProcessPool {
	return &MCPProcessPool{
		clients:     make(map[string]*poolEntry),
		maxSize:     defaultMaxPoolSize,
		idleTimeout: defaultIdleTimeout,
		factory:     factory,
	}
}

func (p *MCPProcessPool) Get(ctx context.Context, cfg mcp_config.MCPServerConfig) (*MCPClient, error) {
	key := poolKey(cfg)

	p.mu.Lock()
	if entry, ok := p.clients[key]; ok {
		entry.lastUsed = time.Now()
		entry.idleTimer.Reset(p.idleTimeout)
		p.mu.Unlock()
		return entry.client, nil
	}
	p.mu.Unlock()

	client, err := p.factory(ctx, cfg)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.clients) >= p.maxSize {
		p.evictOldestLocked()
	}

	timer := time.AfterFunc(p.idleTimeout, func() {
		p.Evict(key)
	})

	p.clients[key] = &poolEntry{
		client:    client,
		lastUsed:  time.Now(),
		idleTimer: timer,
	}

	return client, nil
}

func (p *MCPProcessPool) Evict(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.evictLocked(key)
}

func (p *MCPProcessPool) evictLocked(key string) {
	entry, ok := p.clients[key]
	if !ok {
		return
	}
	entry.idleTimer.Stop()
	entry.client.Close()
	delete(p.clients, key)
}

func (p *MCPProcessPool) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range p.clients {
		if oldestKey == "" || entry.lastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastUsed
		}
	}

	if oldestKey != "" {
		p.evictLocked(oldestKey)
	}
}

func (p *MCPProcessPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key := range p.clients {
		p.evictLocked(key)
	}
}

func (p *MCPProcessPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.clients)
}

func poolKey(cfg mcp_config.MCPServerConfig) string {
	h := sha256.New()
	h.Write([]byte(cfg.Command))
	for _, arg := range cfg.Args {
		h.Write([]byte{0})
		h.Write([]byte(arg))
	}
	h.Write([]byte{0, 0})

	keys := make([]string, 0, len(cfg.Env))
	for k := range cfg.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h.Write([]byte(strings.Join(keys, ",")))

	return fmt.Sprintf("%x", h.Sum(nil))
}
