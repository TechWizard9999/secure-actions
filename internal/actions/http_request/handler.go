package http_request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kotakarthik/secure-actions/internal/secrets"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var secretPlaceholder = regexp.MustCompile(`<<secret:([a-z0-9-]+)>>`)

type Request struct {
	Method  string            `json:"method" jsonschema:"HTTP method: GET, POST, PUT, PATCH, DELETE"`
	URL     string            `json:"url" jsonschema:"Target URL (may contain <<secret:identifier>> placeholders)"`
	Headers map[string]string `json:"headers,omitempty" jsonschema:"Request headers (values may contain <<secret:identifier>> placeholders)"`
	Body    string            `json:"body,omitempty" jsonschema:"Request body (may contain <<secret:identifier>> placeholders)"`
}

type Response struct {
	StatusCode int               `json:"statusCode" jsonschema:"HTTP response status code"`
	Headers    map[string]string `json:"headers" jsonschema:"Response headers"`
	Body       string            `json:"body" jsonschema:"Response body"`
}

type Dependencies struct {
	SecretManager secrets.Manager
}

type Handler struct {
	deps   Dependencies
	client *http.Client
}

func New(deps Dependencies) *Handler {
	return &Handler{
		deps: deps,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *Handler) Execute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input Request,
) (*mcp.CallToolResult, Response, error) {

	method := strings.ToUpper(input.Method)
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
	default:
		return nil, Response{}, fmt.Errorf("unsupported method %q: must be GET, POST, PUT, PATCH, or DELETE", input.Method)
	}

	url, err := h.substituteSecrets(ctx, input.URL)
	if err != nil {
		return nil, Response{}, fmt.Errorf("resolve URL secrets: %w", err)
	}

	body, err := h.substituteSecrets(ctx, input.Body)
	if err != nil {
		return nil, Response{}, fmt.Errorf("resolve body secrets: %w", err)
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, Response{}, fmt.Errorf("create request: %w", err)
	}

	for key, value := range input.Headers {
		resolved, err := h.substituteSecrets(ctx, value)
		if err != nil {
			return nil, Response{}, fmt.Errorf("resolve header %q secrets: %w", key, err)
		}
		httpReq.Header.Set(key, resolved)
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, Response{}, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, Response{}, fmt.Errorf("read response: %w", err)
	}

	respHeaders := make(map[string]string, len(resp.Header))
	for key := range resp.Header {
		respHeaders[key] = resp.Header.Get(key)
	}

	return nil, Response{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(respBody),
	}, nil
}

// substituteSecrets replaces all <<secret:identifier>> placeholders with
// the decrypted secret values from the store.
func (h *Handler) substituteSecrets(ctx context.Context, input string) (string, error) {
	if input == "" {
		return input, nil
	}

	matches := secretPlaceholder.FindAllStringSubmatchIndex(input, -1)
	if len(matches) == 0 {
		return input, nil
	}

	var b strings.Builder
	lastEnd := 0

	for _, match := range matches {
		b.WriteString(input[lastEnd:match[0]])

		identifier := input[match[2]:match[3]]
		encrypted, found, err := h.deps.SecretManager.Get(ctx, identifier)
		if err != nil {
			return "", fmt.Errorf("get secret %q: %w", identifier, err)
		}
		if !found {
			return "", fmt.Errorf("secret %q not found", identifier)
		}

		decrypted, err := secrets.Decrypt(encrypted, identifier)
		if err != nil {
			return "", fmt.Errorf("decrypt secret %q: %w", identifier, err)
		}

		b.WriteString(decrypted)
		lastEnd = match[1]
	}

	b.WriteString(input[lastEnd:])
	return b.String(), nil
}
