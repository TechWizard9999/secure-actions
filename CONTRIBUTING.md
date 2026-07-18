# Contributing to secure-actions

Thanks for your interest in contributing! This document covers everything you need to get started.

## Development Setup

### Prerequisites

- Go 1.25+
- Docker (for MongoDB)
- A Claude Code installation (for end-to-end testing)

### Getting Started

```bash
# Clone the repo
git clone https://github.com/Kota-Karthik/secure-actions.git
cd secure-actions

# Start MongoDB
docker compose -f deployments/docker-compose.yml up -d

# Build and run
go build -o secure-actions ./cmd/secure-actions
./secure-actions --version
```

### Register for local testing

Add to your Claude Code config:

```json
"secure-actions": {
  "type": "stdio",
  "command": "/path/to/your/built/secure-actions",
  "env": {
    "MONGO_URI": "mongodb://localhost:27018"
  }
}
```

## Workflow

### 1. Pick or create an issue

- Check [open issues](https://github.com/Kota-Karthik/secure-actions/issues) for something to work on
- Comment on the issue to let others know you're working on it
- For new ideas, open a feature request issue first to discuss the approach

### 2. Branch

```bash
git checkout -b feat/short-description
# or
git checkout -b fix/short-description
```

### 3. Make your changes

- Keep PRs focused — one feature or fix per PR
- Follow the existing code style (no linter config needed, just match what's there)
- No comments unless the "why" is non-obvious
- No unused code, no dead imports

### 4. Test

```bash
# Must pass
go build ./...

# Test manually with Claude Code MCP if touching tools
```

### 5. Commit

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add timeout configuration for http_request
fix: handle special characters in secret identifier
chore: update mongo-driver to v2.9.0
```

### 6. Open a PR

PR title must follow this format:

```
fix|feat|chore: [#issue]: description
```

Examples:
- `feat: [#12]: add vault backend for secret storage`
- `fix: [#7]: handle empty identifier in delete_secret`
- `chore: update dependencies`

The issue number is optional for `chore` PRs. Fill out the PR template completely.

## Code Guidelines

### Architecture

- Each tool lives in `internal/actions/<tool-name>/`
- Every tool has a `Dependencies` struct, `Handler` struct, and `Execute` method
- Dependencies are injected — never construct them inside the handler
- The `Manager` interface in `internal/secrets/` is what tools depend on, not the Mongo implementation

### Adding a New Tool

1. Create `internal/actions/your_tool/handler.go` with:
   - `Request` struct (input schema)
   - `Response` struct (output schema)
   - `Dependencies` struct
   - `Handler` with `New(deps)` and `Execute(ctx, req, input)` method
2. Register in `internal/mcp/server.go`
3. Update README.md with tool documentation

### Security Rules

These are non-negotiable:

- **Never log secret values** — only identifiers
- **Never return decrypted secrets in tool responses** — the LLM sees responses
- **Always validate user input** before processing
- **Use elicitation for sensitive operations** — never ask for secrets via tool input params
- **Require confirmation for destructive actions** — delete, overwrite

### What Not to Do

- Don't add dependencies without discussion — open an issue first
- Don't change the encryption scheme without a migration plan
- Don't add features that expose secrets to the LLM context
- Don't break backward compatibility with stored secrets

## Release Process

Releases are automated. Maintainers tag and push:

```bash
git tag v0.X.0
git push origin v0.X.0
```

GitHub Actions runs GoReleaser which builds all platforms and publishes to GitHub Releases.

## Getting Help

- Open an issue with the "question" template
- Check existing issues and discussions for similar questions

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
