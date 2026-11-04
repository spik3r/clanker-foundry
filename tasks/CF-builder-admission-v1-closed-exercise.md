---
id: CF-builder-admission-v1
status: in-progress
owner: Foundry builder admission contributor
repositories: [spik3r/clanker-foundry]
---

# Admit the exact six-file Go builder exercise before compilation

## Implementation claim

The coordinator selected this bounded modeling task at verified main
30911539fa87135ff11f85aa6ed0cea0affce868. A live public PR search found none open.
The named-file builder preflight intentionally does not enumerate directories;
the current preparation guide separately warns that Go compilation can include
additional source files. Close that explicit gap for the fixed Go exercise.
This claim does not extend the preflight into a compiler or generic workspace
manager. The coordinator reviews and merges; this contributor does not merge.

Use isolated branch codex/builder-admission-v1 and write only these new paths:

- modeling/lifecycle/builderadmission/check.go
- modeling/lifecycle/builderadmission/files.go
- modeling/lifecycle/builderadmission/inputs.json
- modeling/lifecycle/builderadmission/check_test.go
- modeling/lifecycle/builderadmission/contract_test.go
- modeling/lifecycle/builderadmission/testdata/observations.json
- modeling/lifecycle/builderadmission/README.md
- modeling/lifecycle/cmd/modelbuilderadmission/main.go
- modeling/lifecycle/cmd/modelbuilderadmission/main_test.go
- This task and its new immutable Project Memory source/events.

Do not edit existing preflight/preparer, old docs or frozen exercise files,
qualification/native APIs or source policies, operator routes, runtime, CI,
permissions, held artifacts or other task records. The preparer's fifteen input
pins must remain unchanged. No compiler, parser, Git, candidate, subprocess,
network, worker, formatter, file creation or cleanup belongs in this tool.

## Exact input and read boundary

Expose Check(exerciseRoot, expectedCandidateSHA256 string) ([]byte,error).
Both arguments are trusted host input, not inferred from a supplied manifest.
The root must be an explicit absolute UTF-8 path without control characters.
The candidate assertion is separately reviewed lowercase 64-hex SHA256; neither
hashing the arriving candidate nor a matched receipt authenticates that review.
Reject malformed arguments before filesystem reads. Preserve actual host
traversal through ancestor aliases and dot-dot; do not lexically clean a path
into another location. Strip only trailing separators for final-leaf checks.
Refuse a final symlink or non-directory root, even with trailing separators.
Anchor reads in a rooted directory handle and compare named/opened identities.

A separately authored inputs.json embeds a reviewed literal native builder-input
manifest for the existing five exercise originals with their original byte pins
and module-local paths: README.md, go.mod, contract.go, acceptance_test.go and
rubric.md. Embed and independently fix its exact SHA256 in trusted code. Call
unchanged builderpreflight.Check with those exact bytes and pin. Retain its entire
native report text and exact hash, including all missing/changed/refused file
observations; do not substitute synthetic successes or refresh any pin.

The complete direct-child census must contain exactly those five original files
plus candidate.go, all regular leaves. There are no ignored or excluded names,
cache directories, build artifacts or nested directories. The preparer already
puts cache/tmp/modcache beside exercise. Any additional child blocks, including
candidate.test from an earlier build, hidden files, build-ignored or OS-suffixed
Go files, C/assembly/embed inputs, go.work, directories, symlinks and special
files. Do not delete or relocate blocked entries. No unexpected file contents
are read and no subdirectory is descended; filename and metadata enumeration of
the explicit exercise directory is part of the required trusted read grant.

Cap the census at 64 entries and candidate reads at 1 MiB. Read candidate.go only
when it is a bounded regular named/opened leaf. Retain observed byte count and
SHA256 for a complete stable read. Check ordinary identity, size and modification
time changes while reading and final named identity; do not silently follow a
replacement. Report candidate missing, refused, read-failed or hash-mismatch as
blocked observations where a complete census remains available. Still collect
all native frozen-file observations and the accessible candidate when an original
file is missing or changed. Root/census identity failure, incomplete enumeration
or unexpected internal failure cannot produce a complete-admission assertion.

These are trusted exclusive-host checks, not an atomic multi-file snapshot,
hostile concurrent-writer defense, mount/hardlink authentication or OS sandbox.
Read-only checks can update access times. Candidate bytes are opaque: matching
does not establish Go validity, purity, allowed imports, behavior or approval.

## Report and command

The fixed model-builder-admission-report/v1 JSON report has outcome matched or
blocked, authority none, executionStatus not-run and reviewStatus not-established.
It records the supplied exerciseRoot and expectedCandidateSHA256, the fixed
frozenManifestSHA256, native preflightReportSHA256/JSON, frontierComplete true,
sorted frontier entries with name/kind, sorted missingNames/unexpectedNames, and
candidate with path candidate.go, status, nullable observedSHA256/observedBytes.
Kinds are regular, directory, symlink or special. Candidate statuses are matched,
hash-mismatch, missing, refused or read-failed. Every complete report includes
all child names; malformed UTF-8 filenames refuse rather than being replaced in
JSON. matched requires the exact six-file regular frontier, matched native
preflight and matching stable candidate hash. It grants no build permission.

Output is fully buffered and capped at 1 MiB. The CLI accepts exactly
modelbuilderadmission check --root ABSOLUTE-EXERCISE --candidate-sha256 HEX;
flag pairs may be reordered but occur once each. It reads no stdin and discovers
no default root. Exit 0 requires a delivered matched report; 1 delivers blocked;
2 refuses malformed arguments or invalid locations without a report; 3 means a
short/failed stdout write; 4 means incomplete census or unexpected I/O/internal
boundary failure with no complete-admission report. ErrInvalid identifies the
initial refusal class; ErrBoundary identifies incomplete admission. A candidate
read-failed observation can still be a complete blocked report. Ignore SIGPIPE
so closed stdout reports exit 3. Diagnostics are fixed and omit file contents;
reports can contain private paths and filenames and must be handled accordingly.

## Independent acceptance and completion

Publish this committed public-safe claim and receive coordinator review before
registering its pinned source and recording genuine Project Memory start. A
separate contributor freezes inputs.json, its digest, literal filesystem cases
and expected reports in contract_test.go and observations.json before production.
The frozen originals are compared with their actual reviewed source bytes; no
existing fixture or independently stated candidate pin is rewritten to pass.

Checks cover exact success, an extra Go initializer source, ignored/hidden and
other extra files, directories, symlinks, special leaves, stale/missing originals,
missing/wrong/oversized candidate, malformed pins before reads, case aliases,
ancestor traversal, census limits and permission failures where observable.
Validate full native report preservation, complete frontier sorting, immutable
inputs, no unexpected content reads or writes, actual compiled CLI delivery and
closed stdout. Candidate fixtures are inert bytes; no candidate is compiled,
imported or executed. Separate disposable mutations must expose bypassed census,
candidate pin, frozen pins or leaf checks through expected behavioral failures.

Run Go1.27.3 focused and full race, vet, modelc drift, gofmt, pack, installer,
whitespace and full committed Project Memory validation. Freeze all checkout
writes during full race source-snapshot guards. Obtain coordinator and independent
exact-source review, record observed completion and hand off hosted checks and
merge. Keep unobserved failure paths and platform limits explicit.
