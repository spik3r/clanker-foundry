---
id: CF-validator-scheme-case-v1
status: completed
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

## Local result

The local repair and validation are complete; this status does not claim its
future PR has merged. The original validator failed nine new uppercase/mixed-case
cases, including wrapped destinations. After repair, full installer/wiring/model
resolver regressions, 47 model-map cases and 92 validator path checks passed.
Bash 3.2 syntax, repository validation and whitespace passed. Encoded-colon
present/missing/decoy controls passed. This host aliases local filename case;
the wrong-case missing-path assertion is retained for case-sensitive filesystems
and was not falsely claimed as exercised here. No link destinations were
contacted. Independent working-diff review found no issues. Next: immutable head
review and hosted checks before the coordinator decides whether to merge.
