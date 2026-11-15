---
id: CF-generator-write-root-v1
status: completed
owner: Codex generator write-root contributor
repositories: [spik3r/clanker-foundry]
---

# Write generated files into the directory actually created

## Implementation claim

Owner: Codex generator write-root contributor, delegated by the coordinator
under the owner's continued reviewed-maintenance request. Reviewed main:
`7ea60e8961089c98448059236a4c355117f9e417`. Live PRs are empty and all 174
Project Memory frontiers are complete or abandoned, with no conflicts. The
previous generator contracts and output-root check task are completed; their
records remain unchanged. This claim covers generation, not a reopened check task.

An owned actual `tuic generate` fixture reproduced false success when OUT held
an interior symlink followed by `..`. The raw mkdir created the requested
physical directory, but subsequent filepath.Join calls selected a different
existing lexical directory, overwrote its file and left the requested directory
empty. The five registered generator writers share this pattern. Lifecycle
also checks a lexically cleaned parent before creation, which can falsely
refuse a valid physical parent when the lexical parent is absent.

Preserve raw OUT existence refusal and raw mkdir semantics. After creation,
resolve that created OUT through the filesystem before joining any owned file
path. On resolution failure, report an I/O failure and possible partial output;
never fall back to the raw spelling. Keep the existing no-rollback behavior.
For lifecycle's parent preflight only, use the original parent spelling from
filepath.Split, with an empty parent treated as dot. Preserve its missing and
non-directory parent refusal code and current terminal syntax behavior.

Add a shared black-box generation rule covering an existing lexical decoy and
a valid nested physical parent with no lexical counterpart. Require success,
the exact baseline tree in the physical destination, unchanged decoy contents,
no new lexical directories and unchanged model/work directories. A tagged
reference fault must prove the harness detects the old write-root behavior.
Freeze expectations and show failures on old writers before repairing them.

Exact product scope:

- `modeling/gencontract/conformance.go`
- `modeling/gencontract/conformance_test.go`
- `modeling/gencontract/cmd/refgen/main.go`
- `modeling/tui/contract.go`
- `modeling/jsonvalidator/cmd/jsonvalidator/main.go`
- `modeling/lifecycle/cmd/lifecyclegen/main.go`
- `modeling/diagram/contract/contract.go`
- `modeling/gencontract/README.md`
- `modeling/CONTRACT.md`

Also this task and new immutable Project Memory source/event files. Do not
change emitted bytes, generator versions, profiles, dependencies, schemas,
workflows, legacy fixed-path commands, consumer pins, runtime or old evidence.
Add no shared path framework, automatic cleanup, locks or race guarantees.
Aliases remain ordinary filesystem observations. Request a scope review if an
additional test path is needed for a concrete uncovered behavior.

## Validation

Use owned scratch directories and public sample models. Run the new shared
cases against every registered producer and verify the reference fault is
detected. Preserve existing file/directory/symlink output refusals, missing and
non-directory parents, invalid-model refusal, relative roots and terminal
syntax. Compare emitted bytes with ordinary baseline generation; no version
change is justified when those bytes stay identical.

Run full race tests and vet in gencontract, tui, jsonvalidator, lifecycle,
diagram and foundrygen with the pinned cached toolchain, plus gofmt, pack
validation, whitespace and committed Project Memory validation. Report actual
symlink or optional-tool skips. Freeze for two independent reviews. The
coordinator alone stages completion evidence, commits, publishes the PR and
merges after exact-head review and hosted checks. No real-home installation,
consumer trial or runtime activation belongs to this claim.

## Observed result

The bounded implementation is complete. All five registered generators now
resolve the directory created through the original OUT spelling before joining
emitted file names. Lifecycle checks the original parent spelling without
lexical cleaning. Generator versions and emitted bytes remain unchanged.

The shared cases failed against all five old writers before the fixes. After
strengthening identity and unavailable-symlink assertions, the old-writer replay
still failed. The repaired writers pass both cases and the injected
`clean-generate-root` fault is detected. Independent actual CLI checks confirm
the physical destination contains the baseline tree, the lexical decoy stays
unchanged, and lifecycle retains its existing parent and terminal-syntax refusals.

The coordinator ran all six full race suites and vet: lifecycle and JSON
validator with Go 1.27.3; gencontract, TUI, diagram and foundrygen with Go 1.27.4.
All passed. Pack validation, gofmt across all six modules and whitespace checks
passed. The optional skills-ref tool was unavailable.

An initial lifecycle run under Go 1.27.4 failed its existing exact-toolchain
guards. An exact-Go-1.27.3 run then hit an Apple developer-tool diagnostic on a
rehearsal child's stderr. With Homebrew Git on PATH, the isolated rehearsal
rerun and the coordinator's entire lifecycle suite passed. No product workaround
was added for either environment mismatch.

This records implementation and observed local checks, not publication or merge.
The coordinator retains staging, lifecycle completion, commit and PR authority.
