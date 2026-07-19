# secure-actions

An MCP (Model Context Protocol) server that gives AI assistants secure access to secrets and authenticated HTTP requests — without ever exposing credentials to the LLM.

## Why?

When AI assistants need to call APIs that require authentication, the naive approach is to paste tokens directly into the chat. This means your secrets end up in:
- The LLM's context window
- Chat logs and transcripts
- Model training data (potentially)

**secure-actions** solves this by keeping secrets encrypted at rest and decrypting them only at the moment of use — inside the MCP server process, never in the LLM context.

## How It Works

```
User ──► Claude Code ──► MCP Protocol ──► secure-actions ──► MongoDB
                                               │
                                               ├─ Encrypts with AES-256-GCM
                                               ├─ Master key at ~/.secure-actions/master.key
                                               └─ Decrypts only during HTTP request execution
```

1. **Store a secret** — The MCP server sends an elicitation request to the client (Claude Code). The user enters the value in a form. The value travels only over the local MCP connection, never through the LLM.
2. **Use a secret** — When making HTTP requests, secrets are referenced by placeholder (`<<secret:identifier>>`). The server decrypts and injects them at request time. The LLM only sees the placeholder, never the actual value.
3. **Delete a secret** — Requires explicit user confirmation via elicitation before removal.

## Security

- **AES-256-GCM encryption** — All secrets are encrypted before storage
- **Per-installation master key** — A unique 256-bit random key is auto-generated at `~/.secure-actions/master.key` (file permissions `0600`) on first run. Different for every user.
- **Secrets never enter LLM context** — Collection uses MCP elicitation (client-side form), not chat messages
- **No logging of secret values** — Only identifiers are logged, never plaintext values
- **Confirmation prompts** — Destructive operations (delete, overwrite) require explicit user approval
- **Unique index enforcement** — MongoDB unique index on `identifier` prevents duplicates at the database level

> **Note:** Everything — MongoDB, the master key, and the encrypted secrets — lives entirely on your local machine. As long as you don't share or expose `~/.secure-actions/master.key` to others or to the LLM, your secrets remain safe. No data leaves your system unless you explicitly make an HTTP request with `http_request`.

## Installation

### Prerequisites

- Docker (for MongoDB)

### macOS / Linux

```bash
curl -sSL https://raw.githubusercontent.com/Kota-Karthik/secure-actions/main/scripts/install.sh | sh
```

Or pin a specific version:

```bash
VERSION=v0.1.0 curl -sSL https://raw.githubusercontent.com/Kota-Karthik/secure-actions/main/scripts/install.sh | sh
```

### Windows

Download the latest `.zip` from [GitHub Releases](https://github.com/Kota-Karthik/secure-actions/releases), extract `secure-actions.exe`, and add it to your PATH.

### From Source (any OS)

```bash
go install github.com/kotakarthik/secure-actions/cmd/secure-actions@latest
```

### Verify Installation

```bash
secure-actions --version
```

## Setup

### 1. Start MongoDB

```bash
docker compose -f deployments/docker-compose.yml up -d
```

This starts MongoDB on port `27018` with data persisted at `~/.secure-actions/mongo`.

### 2. Register the MCP Server

Add to your Claude Code configuration (`.claude.json` or via the UI):

```json
{
  "secure-actions": {
    "type": "stdio",
    "command": "secure-actions",
    "env": {
      "MONGO_URI": "mongodb://localhost:27018"
    }
  }
}
```

### 3. First Run

On first run, the server will:
- Generate a master encryption key at `~/.secure-actions/master.key`
- Create the `secrets` collection with a unique index on `identifier`

No manual setup required.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `MONGO_URI` | `mongodb://localhost:27018` | MongoDB connection string |
| `SECURE_ACTIONS_KEY_PATH` | `~/.secure-actions/master.key` | Path to the master encryption key |
| `MONGO_TLS` | `false` | Enable TLS/SSL for MongoDB connection |
| `MONGO_TLS_CA_FILE` | | Path to CA certificate file for TLS |
| `MONGO_CERT_FILE` | | Path to client certificate file (for mTLS) |
| `MONGO_KEY_FILE` | | Path to client key file (for mTLS) |
| `MONGO_AUTH_DB` | `admin` | Authentication database |
| `MONGO_USERNAME` | | Username for MongoDB authentication |
| `MONGO_PASSWORD` | | Password for MongoDB authentication |

### MongoDB with TLS/SSL

To connect to MongoDB with TLS:

```bash
MONGO_TLS=true \
MONGO_TLS_CA_FILE=/path/to/ca.pem \
MONGO_CERT_FILE=/path/to/client.pem \
MONGO_KEY_FILE=/path/to/client.key \
secure-actions
```

### MongoDB with Username/Password Authentication

```bash
MONGO_USERNAME=myuser \
MONGO_PASSWORD=mypassword \
MONGO_AUTH_DB=admin \
secure-actions
```

### MongoDB with TLS and Authentication Combined

```bash
MONGO_TLS=true \
MONGO_TLS_CA_FILE=/path/to/ca.pem \
MONGO_USERNAME=myuser \
MONGO_PASSWORD=mypassword \
MONGO_AUTH_DB=admin \
secure-actions
```

## Tools

### `ping`

Health check. Returns `pong` to verify the server is running.

### `request_secret`

Securely collect and store a secret from the user.

**Input:**
| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Identifier for the secret (normalized to lowercase, spaces become hyphens) |
| `prompt` | No | Custom prompt message shown to the user |
| `description` | No | Description shown above the input form |

**Behavior:**
- If the identifier already exists, prompts user to confirm before updating
- Identifier validation: only lowercase letters, numbers, and hyphens allowed
- Example: `"My API Key"` becomes `"my-api-key"`

### `list_secrets`

Returns all stored secret identifiers and count. Does not expose secret values.

**Output:**
```json
{
  "secrets": ["github-pat", "aws-key", "slack-token"],
  "count": 3
}
```

### `delete_secret`

Permanently delete a stored secret.

**Input:**
| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Identifier of the secret to delete |

**Behavior:**
- Returns "not found" if the identifier doesn't exist
- Prompts user with "Are you sure?" confirmation before deletion

### `http_request`

Execute an HTTP request with automatic secret injection.

**Input:**
| Field | Required | Description |
|-------|----------|-------------|
| `method` | Yes | `GET`, `POST`, `PUT`, `PATCH`, or `DELETE` |
| `url` | Yes | Target URL |
| `headers` | No | Request headers (key-value map) |
| `body` | No | Request body |

**Secret injection:** Use `<<secret:identifier>>` anywhere in the URL, headers, or body. The placeholder is replaced with the decrypted secret value at request time.

**Example:**
```json
{
  "method": "GET",
  "url": "https://api.github.com/user",
  "headers": {
    "Authorization": "Bearer <<secret:github-pat>>",
    "Accept": "application/vnd.github+json"
  }
}
```

The LLM sees `<<secret:github-pat>>` — the actual token value never appears in the conversation.

## Usage Examples

### Store and use a GitHub token

```
You: "Store my GitHub personal access token"
→ request_secret is called, you enter the token in a secure form

You: "List my repos"
→ http_request is called with:
   GET https://api.github.com/user/repos
   Authorization: Bearer <<secret:github-pat>>
```

### Manage secrets

```
You: "What secrets do I have stored?"
→ list_secrets returns: ["github-pat", "slack-webhook"]

You: "Remove the slack webhook"
→ delete_secret prompts: "Are you sure?" → yes → deleted
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `MONGO_URI` | `mongodb://localhost:27018` | MongoDB connection string |
| `SECURE_ACTIONS_KEY_PATH` | `~/.secure-actions/master.key` | Path to the master encryption key |

## Project Structure

```
cmd/secure-actions/          Entry point
internal/
  actions/
    ping/                    Health check tool
    request_secret/          Secret collection and storage
    list_secrets/            List stored identifiers
    delete_secret/           Delete with confirmation
    http_request/            HTTP client with secret injection
  app/                       Application wiring
  config/                    Configuration loading
  mcp/                       MCP server and tool registration
  secrets/                   Encryption, key management, Manager interface
  storage/mongo/             MongoDB client and repository
deployments/
  docker-compose.yml         Local MongoDB setup
  Dockerfile                 Container build (planned)
scripts/
  install.sh                 macOS/Linux installer
```

## Development

```bash
# Run locally
docker compose -f deployments/docker-compose.yml up -d
go run ./cmd/secure-actions

# Build
go build -o secure-actions ./cmd/secure-actions

# Test the binary
./secure-actions --version
```

## Releasing

Releases are automated via GitHub Actions and GoReleaser:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This builds binaries for all platforms (linux/darwin/windows × amd64/arm64), creates a GitHub Release with checksums, and publishes the artifacts.

## License

See [LICENSE](LICENSE) for details.
