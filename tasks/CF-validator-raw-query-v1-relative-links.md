---
id: CF-validator-raw-query-v1
status: completed
owner: Codex validator contributor
repositories: [spik3r/clanker-foundry]
---

# Resolve relative link paths before URL queries

## Implementation claim

Owner: Codex validator contributor. Branch `codex/validator-raw-query-v1`, reviewed
main `0f63234`, with no open PRs or competing claim at start. Scope:
`scripts/validate.sh`, `tests/test-install.sh`, `docs/getting-started.md`, this task
and new immutable Project Memory evidence.

Relative Markdown link queries currently become filesystem filename bytes. Strip
the raw query before percent-decoding the path, retaining encoded question marks
as filename data. Preserve the merged percent decoder and its malformed-escape
policy. Test present and missing destinations, a literal-question-mark decoy,
query plus fragment, encoded-question-mark filename plus query, query-only self
links, and ignored-query NUL escapes/malformed escapes. No generic Markdown parser
or other URL-policy changes.

Prove failures before repair, then run Bash 3.2 syntax, full installer and
model-map regressions, repository validation, whitespace and committed memory
coverage. Record local completion after validation; publication does not assert
future merge. Publish for independent review; the coordinator owns merge.

## Local result

The local repair is complete; this status does not claim that a future PR has
merged. The original validator failed seven new cases in the full installer
regression, including a literal-question-mark decoy falsely accepted while the
URL path was missing. After repair, the full installer and resolver regression
passed, alongside 47 model-map cases and 92 validator path checks. Bash 3.2.57
syntax, repository validation and whitespace checks passed. All installations
used the existing isolated fixture homes and targets. Next: independent frozen
head review and hosted checks before the coordinator decides whether to merge.
