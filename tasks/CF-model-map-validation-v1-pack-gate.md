---
id: CF-model-map-validation-v1
status: in-progress
owner: astra-foundry-model-map-validation
repositories: [spik3r/clanker-foundry]
---

# Reject model maps that cannot be resolved

## Implementation claim

Owner: `astra-foundry-model-map-validation`, assigned the next independent
Foundry improvement after installer PR 103 review. Verified starting main is
`7778beb4038f081ca0f36415a286811e8627fac6`.

The pack validator currently accepts duplicate keys, non-ASCII or spaced model
values, NUL-containing comments and empty local maps, although the resolver and
documented mapping format reject them. Make the pack gate reject those invalid
mapping files without changing the resolver's policy or execution behavior.

Exact writable product scope:

- `scripts/validate.sh`, only model-map validation and local-map presence checks.
- `tests/test-model-map-validation.sh`, a new isolated focused regression suite.
- `tests/test-validate.sh`, only the focused-suite hook and any bounded supporting
  argument plumbing; preserve all existing filename regression assertions.
- This task at `tasks/CF-model-map-validation-v1-pack-gate.md`.
- New immutable evidence under `.project-memory/tasks/CF-model-map-validation-v1/`.

Keep the validator's existing supported client/tier allowlist. Require at least
one valid mapping in every present map. Reject duplicate keys even when equal or
unused, NUL bytes anywhere, whitespace/equal signs/non-ASCII bytes in values and
other malformed records. Detect invalid local-map paths rather than treating
them as absent. Missing local maps still mean no override; valid local keys may
repeat shared keys. Preserve accepted regular symlink behavior. Never execute
mapping contents or echo model values in diagnostics. Reader failures must fail
validation instead of being counted as a clean file.

Leave profile frontmatter/tier parsing, CRLF Markdown behavior, pointer checks,
link handling, installer behavior, resolver code, runtime, workflows and
dependencies unchanged. PRs 101, 102 and 103 and the setup-helper/runtime proposals
are separate; do not change their state or retry any held action. The shared
heavy-test hold remains. No provider/client action, installation, credential or
permission change is part of this work.

## Ownership and chronology

The completed `CF-reviewed-fixes-adoption-v1` and `CF-002` lifecycle frontiers
touch the validator paths. CF-002's older task still says in-progress, but its
claim is specifically the literal pointer/ShellCheck comment correction, its
frontier is complete and PR 7 is merged. This claim leaves that code and both
tasks' owners, text, source pins and immutable history unchanged. It does not
close or take over those task identities. Independent scope review must verify
this bounded separation and current open work before product edits.

Independently review these public-safe source bytes, create the preparatory task
commit, read back its actual commit/tree/parent and register that real source.
Record a genuine start before implementation. A preparatory source commit is
part of the repository's evidence workflow, not a claim that implementation is
complete. If publication changes commit identities, preserve accurate ancestry;
do not rewrite earlier events or pretend a local commit was published.

## Validation and delivery

Use invented maps and temporary isolated homes. Cover valid shared/local values,
cross-file overrides, selected/unused and same/different duplicate values,
empty/comment-only maps, NUL positions, printable ASCII boundaries, whitespace,
CRLF map entries, unsupported keys, local path types and finite reader failures.
Prove diagnostic output does not reveal fixture model values. Preserve positive
controls and existing assertions; check new expectations against the unchanged
baseline. Do not create sockets or run heavy aggregates.

Run focused map and existing filename suites, shell syntax, pack validation and
whitespace checks. Obtain independent exact-diff review, record completion only
with observed results, commit final code/evidence and validate the full committed
diff against the actual base. Report unavailable tools and unrun platform checks.
Publish a draft under current authorization and inspect hosted checks for its
actual head. Any later authorized integration must preserve the source commit
through normal history-preserving merge and respect other ownership and holds.
