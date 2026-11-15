---
id: CF-project-metadata-records-v1
status: completed
owner: Codex project metadata contributor
repositories: [spik3r/clanker-foundry]
---

# Preserve submodule metadata record boundaries

## Implementation claim

Owner: Codex project metadata contributor, delegated by the coordinator under
the owner's continued reviewed-maintenance request. Reviewed main:
`b90759abd0b2ae884ebf9d6bb4e75d688c22c527`. Live PRs are empty; all 175
Project Memory frontiers are complete or abandoned, without conflicts. Earlier
setup and index-visibility work is merged in PRs 93 and 156. Preserve those
historical task records rather than reopening their completed implementations.

An owned current-main consumer passes setup and checking, then falsely fails
checking after only its staged Git submodule subsection is renamed to
`shared baseline`. Its path, URL and pin remain valid. The helper splits a
combined Git key/value line at whitespace, losing a valid subsection name.
Real Git also accepts tab and backslash subsection names. Its value output must
not be trimmed: the current command substitutions discard trailing LF bytes.

Replace only the baseline metadata lookup with NUL-delimited key and value
records. Keep Git producer failures visible through checked pipelines, including
partial output followed by failure. Do not use process substitution to hide a
query status, command substitution for values, temporary files or a new generic
Git wrapper.
Returning a selected Git-valid key or a match/count result is allowed.

Accept exactly one mapping to the literal `.agent-baseline`, with exactly one
path value on that key. Require exactly one nonempty URL value with no LF,
including trailing LF. Preserve other URL bytes, including spaces and tabs.
Other submodule mappings remain irrelevant, including unrelated repeated values.
Refuse duplicate matching mappings, mixed or duplicate selected-key paths,
missing values, invalid metadata and failed queries. Add no subsection-name
restriction, whitespace normalization, URL authenticity claim or new CLI flag.

Exact product scope: `scripts/project-setup.sh`,
`tests/test-project-setup.sh`, `docs/project-shared-base.md`, this task and new
immutable Project Memory source/event files. Preserve staged metadata, pins,
clean-baseline checks, include suppression, read-only Git settings, destination
preflight and all wiring behavior. No global setup, real consumer data, generator,
runtime, dependency, workflow, instruction-source or historical evidence changes.

## Validation

Register this exact reviewed Git source and record a genuine start before test
or implementation edits. Freeze regression expectations before repairing the
lookup and observe old-script failures for spaced names and trailing-LF values.

Use real Git with owned local repositories and private homes. Ordinary,
space/tab/backslash subsection names must pass setup, rerun and checking with
metadata and the staged pin unchanged. Git does not accept newline in a key;
this task adds no support for such keys. Keep unrelated mappings and repeated
unrelated values working. Only the exact baseline path matches; whitespace or
LF-bearing path values must not become matches. Preserve nonempty URL spaces
and tabs while refusing empty, duplicate, embedded-LF and trailing-LF URLs.

Retain refusal for selected-key duplicate or mixed paths, duplicate baseline
sections, missing path/URL and malformed metadata. An owned forwarding Git shim
must show that partial valid output followed by failure rejects each changed
query class. Real Git fixtures remain the functional proof. Both setup refusal
and checking must preserve project contents, links, staged Git metadata and HOME.
No local clone transport is enabled beyond the owned fixture repositories.

Run native Bash 3.2 project-setup regressions, full installer tests, separate
syntax checks for both changed shell scripts, ShellCheck, pack validation,
whitespace and committed Project Memory validation. State optional-tool skips.
Freeze for two independent reviews; the coordinator owns completion staging,
commit, publication and merge after exact-head review and hosted checks.
No real-home installation, runtime activation or consumer trial belongs here.

## Observed result

The bounded implementation is complete. The helper reads NUL-delimited key
names and exact value records through checked pipelines. It accepts Git-valid
subsection whitespace, preserves nonempty URL spaces and tabs, ignores repeated
unrelated paths, and refuses ambiguous baseline mappings, LF-bearing selected
values and failed queries without changing wiring or pin selection.

Frozen regressions failed against the old helper at the spaced subsection. A
separate owned actual-helper check showed trailing-LF path and URL values were
accepted after command substitution removed LF. The repaired helper passes the
new real-Git cases and all existing project-setup regressions on native Bash
3.2 with Git 2.54. Partial valid output followed by failure is refused for key,
path-value and URL queries; refusal and checking snapshots remain unchanged.

Full installer regressions, separate syntax checks for both changed shell
scripts, ShellCheck, pack validation and whitespace checks passed. The optional
skills-ref tool was unavailable. Fixtures used local owned repositories and
private homes; no real-home setup or client loading was performed.

This records implementation and observed local checks, not publication or merge.
The coordinator retains staging, lifecycle completion, commit and PR authority.
