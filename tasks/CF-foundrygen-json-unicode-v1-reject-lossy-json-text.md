---
id: CF-foundrygen-json-unicode-v1
status: in-progress
owner: Codex coordinator with delegated builder
repositories: [spik3r/clanker-foundry]
---

# Refuse lossy Unicode in foundrygen configuration

## Implementation claim

The owner requested continued reviewed maintenance across Foundry-related repos,
including tested PRs and normal merges. The coordinator reviewed this bounded
public-safe task against main `2ed5439c57d3abe7b07df276b260623f29a245bb`.
Live PRs and issues are empty; Project Memory reports no active task or conflict.
Use an isolated worktree. Preserve all existing claims, legacy memory bytes and
unrelated worktrees. This is ordinary source maintenance, not managed dispatch
or permission to activate a runtime, change authority or publish a release.

Actual owned CLI fixtures show that raw invalid UTF-8 and escaped lone
surrogates in a model path silently select `design/�.json` and produce a valid
plan. LoadRegistry likewise converts malformed module paths to `modeling/�`.
Go JSON decoding replaces the original text before path validation. Refuse this
loss of path identity; do not normalize paths or reject legitimate U+FFFD.

Exact implementation scope:

- `modeling/foundrygen/project.go`: a private Unicode preflight on original JSON
  bytes before ParseProject and LoadRegistry decode strings. Keep size checks
  first. Reject invalid UTF-8 and unpaired escaped surrogates in keys/values.
  Follow the existing lifecycle strictjson raw-escape pattern without importing
  its module or schema policy. Keep malformed-syntax diagnostics with the
  current decoders; scan escapes only when JSON syntax is known valid.
- `modeling/foundrygen/unicode_test.go`: frozen literal helper/intake and actual
  Run regressions using owned scratch projects and the public reference fixture.
- `modeling/foundrygen/README.md`: state the lossless text admission rule and
  valid U+FFFD/pair support without expanding path or security guarantees.
- This canonical task record, plus new immutable Project Memory source/events.

Preserve valid Unicode, literal/escaped U+FFFD, valid surrogate pairs and ordinary
JSON escaping. Escaped backslashes must not be mistaken for surrogate escapes.
Project duplicate/unknown/type rules and registry's current unknown/duplicate
handling stay unchanged for valid Unicode. Preserve all byte caps, path aliases,
symlink and pin behavior. No dependency, version, schema/language extension,
model, generated artifact, registry data, other module, runtime, workflow,
consumer pin or historical metadata change is included.

## Acceptance and validation

Commit this reviewed claim, register its exact source, independently verify the
snapshot and record a genuine Node24 v2 start before tracked tests or code.
Freeze literal tests and reproduce the wrong-path/refusal failures first.

Reject raw invalid UTF-8, lone high/low, reversed and broken surrogate pairs in
project model/out and other strings/keys, and registry module/package/key/value
text. Valid U+FFFD and supplementary-character filenames remain usable; literal
and escaped spellings select the same intended owned file. Keep combining text,
escaped quotes and literal backslash-u text where existing field/path rules
otherwise admit them. Retain independent path backslash refusal.

For plan/check/generate, malformed configuration returns exit2 and empty stdout
before generator build, model path resolution/read or output writes. Prove the
build boundary with an owned forwarding Go trap; do not claim zero metadata
reads. Preserve all owned input/output bytes and metadata. Refused registries
also stop before build. Public fixture --unpinned remains limited to this
repository's tests, with no production pin-policy change.

Malformed syntax/escapes, truncated strings, trailing data and boundary lengths
must refuse without panic. Use a finite deterministic byte/escape table, not a
new JSON framework, fuzz system, normalization policy or unrelated schema limit.

Run focused Unicode tests, full foundrygen race tests/vet/build and gofmt on
Go1.27.4. Run `scripts/validate.sh` and relevant installer/pack checks. Record
observed results, stage named paths, complete the genuine run from actual staged
blobs, commit new immutable evidence, and validate the full committed diff with
explicit code/source repository identities. Do not rewrite old memory files.

Two independent reviews bind the frozen files and final commit/tree. The
coordinator owns staging, completion, commit, push, PR and normal merge after
current hosted gates pass. Stop for scope review if another path must change.
