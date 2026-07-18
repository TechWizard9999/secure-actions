## PR Title Format

```
fix|feat|chore: [#issue]: description
fix|feat|chore: description
```

The `[#issue]` part is optional.

Examples:
- `feat: [#12]: add vault backend for secret storage`
- `fix: [#7]: handle empty identifier in delete_secret`
- `chore: update dependencies`
- `feat: add timeout configuration for http_request`

---

## Summary

<!-- What does this PR do? 1-3 bullet points. -->

-

## Type

- [ ] `fix` — Bug fix
- [ ] `feat` — New feature
- [ ] `chore` — Maintenance (deps, CI, docs, refactor)

## Changes

<!-- List the key files/areas changed and why. -->

-

## Testing

<!-- How did you verify this works? -->

- [ ] `go build ./...` passes
- [ ] Tested manually with Claude Code MCP
- [ ] Added/updated tests (if applicable)

## Security Checklist

<!-- For changes touching secrets, encryption, or HTTP handling -->

- [ ] No secret values are logged or returned in responses
- [ ] Encryption key is not exposed in any code path
- [ ] New inputs are validated/sanitized

## Screen Recording

<!-- Attach a screen recording or screenshots showing the feature/fix working in Claude Code. -->
<!-- Drag and drop a video/gif here, or paste a link. -->

## Related Issues

<!-- Link related issues: Closes #X, Fixes #X, Related to #X -->

