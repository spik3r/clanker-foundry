---
id: CF-foundry-path-alias-v1
status: completed
owner: Codex path containment contributor
repositories: [spik3r/clanker-foundry]
---

# Refuse existing filesystem aliases across inputs and outputs

## Implementation claim

Owner: Codex path containment contributor, delegated by the coordinating session
under the owner's maintenance and reviewed-merge request. Base: main at
`f98b825`. The coordinator reviewed current work; this does not overlap the
pending JSON schema text or generator output-root changes.

On the host's case-insensitive filesystem, a project with model
`data/model.json` and output `DATA` passes lexical containment checks. With a
real generator manifest in data, plan reports replace and generate exits 0
while deleting the model as part of replacing its own ancestor directory.
This was reproduced with finite owned scratch files and the pinned generator;
no concurrent mutation or fabricated generator was involved.

Before building or invoking generators, resolve every declared path under the
existing no-symlink rules. Use existing filesystem identity and ancestry to
refuse any existing output that aliases a model, foundry.json or an ancestor
of either, and any pair of outputs whose existing roots alias or contain one
another. Preserve valid case-distinct directories on filesystems that distinguish
them. Do not case-fold strings, probe by writing files, or claim to resolve
aliases between nonexistent names without filesystem metadata.

Scope: `modeling/foundrygen/run.go`, new
`modeling/foundrygen/path_alias.go`, new
`modeling/foundrygen/path_alias_test.go`, `modeling/foundrygen/README.md`, this
task and new immutable Project Memory files. No generators, versions, consumer
or runtime changes.

Tests: reproduce the actual case-alias failure before the fix, then all verbs
must refuse without changing source/model/target bytes. Exercise other-model
protection, missing-model ancestors, output equality/nesting aliases, normalized
Unicode aliases where supported and legitimate case-distinct paths. Cover actual
filesystem identity on Linux with hardlinked files and ancestry checks, without
mocking the implementation. Platform alias fixtures may skip when the filesystem
has no such alias. Run full foundrygen vet/race, repository validation, whitespace
checks and committed Project Memory validation. Publish a PR for independent
coordinator review; this contributor must not merge it.

## Integration result

The coordinating agent reviewed `1d3e1da7faa4d1c1f0969cd6b8ea496c527f328d`;
all ten hosted workflows passed on that head. [PR #151](https://github.com/spik3r/clanker-foundry/pull/151)
merged on 14 November 2026 as `b8f729dba52381733175adcfb3eeed4ef781185a`.
This bounded implementation claim is complete and reserves no further work.
