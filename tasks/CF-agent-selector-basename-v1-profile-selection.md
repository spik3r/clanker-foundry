---
id: CF-agent-selector-basename-v1
status: in-progress
owner: Foundry profile selector contributor
repositories: [spik3r/clanker-foundry]
---

# Keep agent profile selectors to one basename

## Implementation claim

The coordinator assigned a bounded next-wave candidate after current public main
`7778beb4038f081ca0f36415a286811e8627fac6` and visible ownership were checked.
The previous resolver, profile-CRLF and shared-fix tasks are done and release
only their completed outcomes. Other task mentions of the resolver are historical
context or validation references, not competing write scopes. PR101 is a separate
pending requirement-change regression; do not alter its branch or review/merge
state as part of this task. Recheck current ownership before publishing.

The receiving contributor accepts this exact prospective frontier:

- `scripts/agent-model.sh`.
- `tests/test-agent-model.sh`.
- `agents/README.md`.
- This task, `tasks/CF-agent-selector-basename-v1-profile-selection.md`.
- New immutable source and lifecycle records for this task only.

Review this claim and the candidate independently before publication. Publish and
verify the genuine task source, register that exact revision and record a real
start before maintained implementation edits. Preserve every existing task and
source/event record. The proposal check is not a maintained implementation run.

## Observed problem and intended behavior

The resolver's documented argument selects an agent name, but an invented fixture
shows `../outside` selecting a Markdown file outside `agents/` and returning its
mapped model with exit 0. It also accepts a nested path, a relative path to a
flat profile, repeated separators and an empty selector when matching fixture
files exist. No real external file, local override, HOME configuration, client or
provider was accessed for this observation.

Reject an empty selector or one containing `/` with the existing usage error
(exit 2 and empty stdout), before profile existence/read operations. Preserve
all other basename spelling, including underscores, periods and spaces. Do not
add a broader filename grammar. Unknown nonempty flat names retain their existing
unknown-profile result. Mapping, tier, override and CRLF rules stay unchanged.

This is lexical selector validation, not filesystem isolation. Existing profile
symlink behavior is unchanged. No symlink confinement, native Windows/macOS
qualification, provider safety guarantee or general path-security claim follows.
No install, runtime, model profile, JSON validator, workflow, dependency, client
configuration or security-setting change belongs to this scope.

## Acceptance and validation

Use disposable fixture repositories with invented test files and copies of tracked
public profile/map assets. Never read real HOME configuration or local overrides,
and never call clients or providers. Freeze and independently review six invalid
selector cases (`../outside`, `./fixture-fast`, `nested/profile`,
`..//outside`, empty, `/absolute`) and four preserved flat basenames
(`fixture-fast`, `fixture_name`, `fixture.name`, `fixture space`). The old resolver
must fail the new expected contract; the candidate must pass it and all existing
resolver tests. Distinguish actual old acceptance from the absolute-looking
selector's existing nonzero lookup error.

Run Bash syntax checks, the complete isolated resolver suite and repository
validation against the final maintained candidate. Run ShellCheck where available
and report an unavailable tool honestly. Verify whitespace and that only the
accepted frontier changed. Check complete committed Project Memory coverage
against the actual base, with code/source identity `spik3r/clanker-foundry`.

Stage exact paths, record observed completion and preserve prior immutable
evidence. Obtain final independent review before publishing an implementation
draft PR. Verify the actual remote commit and hosted checks separately; scratch
results do not establish that those stages passed. No release, deployment, client
launch, provider call or change to held PR101 is included.
