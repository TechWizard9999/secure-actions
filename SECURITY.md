# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in secure-actions, **please do not open a public issue.**

Instead, report it privately:

1. Email: [Create a security advisory](https://github.com/Kota-Karthik/secure-actions/security/advisories/new) on GitHub
2. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide a timeline for a fix.

## Scope

The following are in scope for security reports:

- Secret value leakage (to LLM context, logs, responses, or disk)
- Encryption weaknesses or key exposure
- Authentication bypass
- Unauthorized access to stored secrets
- Injection attacks via placeholders or tool inputs
- MongoDB access control issues

## Out of Scope

- Vulnerabilities in dependencies (report upstream, but let us know)
- Issues requiring physical access to the machine
- Social engineering attacks
- Denial of service via legitimate MCP tool calls

## Security Design

- Secrets are encrypted with AES-256-GCM using a per-installation master key
- The master key is stored at `~/.secure-actions/master.key` with `0600` permissions
- Secret values never appear in tool responses, logs, or stderr
- Destructive operations require user confirmation via MCP elicitation
- All secret identifiers are validated against a strict allowlist pattern (`[a-z0-9-]`)

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
| < latest | Best effort |
