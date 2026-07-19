package mcp_tool_request

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/kotakarthik/secure-actions/internal/mcp_config"
)

func TestNewMCPClient_UnsupportedTransport(t *testing.T) {
	_, err := NewMCPClient(context.Background(), mcp_config.MCPServerConfig{
		Type: "http",
	})
	if err == nil {
		t.Fatal("expected error for unsupported transport")
	}
}

func TestNewMCPClient_InvalidCommand(t *testing.T) {
	_, err := NewMCPClient(context.Background(), mcp_config.MCPServerConfig{
		Type:    "stdio",
		Command: "/nonexistent/binary/that/does/not/exist",
	})
	if err == nil {
		t.Fatal("expected error for invalid command")
	}
}

func TestInitialize_Success(t *testing.T) {
	clientR, clientW := io.Pipe() // client writes to server
	serverR, serverW := io.Pipe() // server writes to client

	client := newMCPClientFromPipes(clientW, serverR)
	defer client.Close()

	// Simulate server responding to initialize
	go func() {
		buf := make([]byte, 4096)
		n, _ := clientR.Read(buf) // read initialize request
		var req JSONRPCRequest
		json.Unmarshal(buf[:n], &req)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{},
				"serverInfo":      map[string]any{"name": "test-server", "version": "1.0"},
			},
			ID: req.ID,
		}
		data, _ := json.Marshal(resp)
		serverW.Write(append(data, '\n'))

		// Read the initialized notification
		clientR.Read(buf)
	}()

	err := client.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
}

func TestInitialize_ReadError(t *testing.T) {
	clientR, clientW := io.Pipe()
	serverR, _ := io.Pipe()

	client := newMCPClientFromPipes(clientW, serverR)
	defer client.Close()

	// Consume the initialize request then close so readResponse fails
	go func() {
		buf := make([]byte, 4096)
		clientR.Read(buf)
		serverR.Close()
	}()

	err := client.Initialize(context.Background())
	if err == nil {
		t.Fatal("expected error when server closes connection")
	}
}

func TestCallTool_Success(t *testing.T) {
	clientR, clientW := io.Pipe()
	serverR, serverW := io.Pipe()

	client := newMCPClientFromPipes(clientW, serverR)
	defer client.Close()

	go func() {
		buf := make([]byte, 4096)
		n, _ := clientR.Read(buf)
		var req JSONRPCRequest
		json.Unmarshal(buf[:n], &req)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "issue created"},
				},
			},
			ID: req.ID,
		}
		data, _ := json.Marshal(resp)
		serverW.Write(append(data, '\n'))
	}()

	result, err := client.CallTool(context.Background(), "create_issue", map[string]any{
		"owner": "test",
		"repo":  "test-repo",
		"title": "Test Issue",
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestCallTool_MCPError(t *testing.T) {
	clientR, clientW := io.Pipe()
	serverR, serverW := io.Pipe()

	client := newMCPClientFromPipes(clientW, serverR)
	defer client.Close()

	go func() {
		buf := make([]byte, 4096)
		n, _ := clientR.Read(buf)
		var req JSONRPCRequest
		json.Unmarshal(buf[:n], &req)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: req.ID,
		}
		data, _ := json.Marshal(resp)
		serverW.Write(append(data, '\n'))
	}()

	_, err := client.CallTool(context.Background(), "nonexistent_tool", nil)
	if err == nil {
		t.Fatal("expected error for MCP error response")
	}
	if err.Error() != "MCP error: Method not found (-32601)" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCallTool_InvalidJSON(t *testing.T) {
	clientR, clientW := io.Pipe()
	serverR, serverW := io.Pipe()

	client := newMCPClientFromPipes(clientW, serverR)
	defer client.Close()

	go func() {
		// Drain the request so sendRequest doesn't block
		buf := make([]byte, 4096)
		clientR.Read(buf)
		// Write invalid JSON as response
		serverW.Write([]byte("not valid json\n"))
	}()

	_, err := client.CallTool(context.Background(), "some_tool", nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestCallTool_WriteError(t *testing.T) {
	_, clientW := io.Pipe()
	serverR, _ := io.Pipe()

	client := newMCPClientFromPipes(clientW, serverR)
	// Close stdin so write fails
	clientW.Close()

	_, err := client.CallTool(context.Background(), "some_tool", nil)
	if err == nil {
		t.Fatal("expected error when stdin is closed")
	}
}

func TestClose_NilFields(t *testing.T) {
	client := &MCPClient{}
	err := client.Close()
	if err != nil {
		t.Fatalf("Close on empty client should not error: %v", err)
	}
}

func TestNewMCPClient_WithEnv(t *testing.T) {
	client, err := NewMCPClient(context.Background(), mcp_config.MCPServerConfig{
		Type:    "stdio",
		Command: "cat",
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
	})
	if err != nil {
		t.Fatalf("NewMCPClient failed: %v", err)
	}
	defer client.Close()

	if client.cmd == nil {
		t.Fatal("cmd should not be nil")
	}
}
