---
id: CF-validator-percent-path-v1
status: in-progress
owner: Codex validator maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Resolve percent-encoded relative link paths

## Implementation claim

The coordinator assigned this bounded repair after isolated fixtures showed
that the pack validator rejects existing percent-encoded destinations and can
accept an encoded-spelling decoy while the decoded destination is missing.
Reviewed main at `44db6e4`, current claims and open PRs #159 and #160; their
record closure and Project Memory repair do not overlap this scope.

Limit changes to scripts/validate.sh, its existing installer validator
regressions in tests/test-install.sh, a concise clarification in
docs/getting-started.md if needed, this record and new immutable evidence.
Decode valid percent triplets exactly once after splitting the raw fragment.
Keep plus signs, literal backslashes, encoded hash filename bytes and decoded
trailing newlines unchanged. Reject decoded NUL bytes. Leave malformed escapes
literal, preserving existing literal paths. Do not add dependencies or a generic
Markdown parser, or change external-link and reference-link handling.

Prove the regressions fail before repair. Cover present/missing destinations,
literal encoded decoys, percent/double encoding, Unicode, space, hash, lowercase
hex, plus, backslash, trailing newline, NUL and malformed-escape controls.
Run Bash 3.2 syntax, full installer and model-map regressions, repository
validator, whitespace and committed Project Memory coverage. Publish for
independent review; the coordinator owns merge.
