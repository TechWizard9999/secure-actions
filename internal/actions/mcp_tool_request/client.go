package mcp_tool_request

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/kotakarthik/secure-actions/internal/mcp_config"
)

// MCPClient communicates with an MCP server over stdio using JSON-RPC
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr io.ReadCloser
	mu     sync.Mutex
	reqID  atomic.Int32
}

// JSONRPCRequest is a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	ID      int32          `json:"id"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	Result  map[string]any `json:"result,omitempty"`
	Error   *JSONRPCError  `json:"error,omitempty"`
	ID      int32          `json:"id"`
}

// JSONRPCError is a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// newMCPClientFromPipes creates an MCPClient from pre-existing I/O (used in tests)
func newMCPClientFromPipes(stdin io.WriteCloser, stdout io.Reader) *MCPClient {
	return &MCPClient{
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}
}

// NewMCPClient creates and starts a new MCP client
func NewMCPClient(ctx context.Context, config mcp_config.MCPServerConfig) (*MCPClient, error) {
	if config.Type != "stdio" {
		return nil, fmt.Errorf("only stdio transport supported, got %q", config.Type)
	}

	// Build command
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)

	// Set up environment with config env vars
	cmd.Env = os.Environ()
	for k, v := range config.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("start process: %w", err)
	}

	client := &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		stderr: stderr,
	}

	return client, nil
}

// Initialize performs the MCP protocol handshake (initialize + initialized notification)
func (c *MCPClient) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	reqID := c.reqID.Add(1)

	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "secure-actions",
				"version": "0.0.1",
			},
		},
		ID: reqID,
	}

	if err := c.sendRequest(initReq); err != nil {
		return fmt.Errorf("send initialize: %w", err)
	}

	if _, err := c.readResponse(); err != nil {
		return fmt.Errorf("read initialize response: %w", err)
	}

	// Send initialized notification (no ID = notification)
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal initialized notification: %w", err)
	}
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("send initialized notification: %w", err)
	}

	return nil
}

// CallTool calls an MCP tool and returns the result
func (c *MCPClient) CallTool(ctx context.Context, toolName string, params map[string]any) (map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	reqID := c.reqID.Add(1)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]any{
			"name":      toolName,
			"arguments": params,
		},
		ID: reqID,
	}

	if err := c.sendRequest(req); err != nil {
		return nil, fmt.Errorf("send tools/call: %w", err)
	}

	resp, err := c.readResponse()
	if err != nil {
		return nil, fmt.Errorf("read tools/call response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s (%d)", resp.Error.Message, resp.Error.Code)
	}

	return resp.Result, nil
}

func (c *MCPClient) sendRequest(req JSONRPCRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write to stdin: %w", err)
	}
	return nil
}

func (c *MCPClient) readResponse() (*JSONRPCResponse, error) {
	line, err := c.stdout.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read from stdout: %w", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response %q: %w", line, err)
	}

	return &resp, nil
}

// Close closes the MCP client
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
	return nil
}
