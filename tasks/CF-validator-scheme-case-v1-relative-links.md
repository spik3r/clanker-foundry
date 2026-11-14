---
id: CF-validator-scheme-case-v1
status: in-progress
owner: Codex validator contributor
repositories: [spik3r/clanker-foundry]
---

# Recognize existing external URI schemes regardless of case

## Implementation claim

Owner: Codex validator contributor. Branch `codex/validator-scheme-case-v1`,
reviewed main `75a01c8` after PR168 merged; no open PR or competing validator
claim at start. Scope: scripts/validate.sh, tests/test-install.sh,
docs/getting-started.md, this task and new immutable Project Memory evidence.

The validator currently rejects uppercase and mixed-case HTTP, HTTPS and MAILTO
links as missing local paths. RFC3986 section3.1 defines schemes as
case-insensitive. Match only those already-skipped raw schemes using explicit
ASCII character pairs compatible with Bash3.2. Preserve angle unwrap ordering,
raw classification before percent-decoding, local path case and diagnostics.
Add no protocols, dependencies, general parser or broader URI policy.

Prove bare/wrapped upper/mixed-case cases red before repair. Cover ordinary
lowercase schemes, encoded-colon relative paths present/missing/raw-spelling
decoys, and filesystem-aware local path case. Use only isolated fixtures; make
no requests to link destinations. Run Bash syntax, full installer/wiring/model
resolver and validator/model-map regressions, repository validation, whitespace
and committed memory coverage. Complete the local record only after checks,
without claiming a future PR merge. Publish for independent review; no merge.
