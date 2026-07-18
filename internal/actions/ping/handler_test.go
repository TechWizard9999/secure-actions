package ping

import (
	"context"
	"testing"
)

func TestPing(t *testing.T) {
	h := New(Dependencies{})

	_, resp, err := h.Execute(context.Background(), nil, Request{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Message != "pong" {
		t.Fatalf("got %q, want %q", resp.Message, "pong")
	}
}
