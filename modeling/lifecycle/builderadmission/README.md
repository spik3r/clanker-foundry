# Admit the closed Go builder exercise

The existing [builder preflight](../builderpreflight/README.md) checks named input
bytes without listing other files. Go compilation can include additional sources.
This read-only check verifies a **complete six-file frontier** before a separately
authorized build of the fixed [exercise](../../../examples/model-builder-go/acceptance/README.md).
It does not compile, parse, import, execute, format or approve a candidate.

## Exact boundary

`Check(exerciseRoot, expectedCandidateSHA256 string) ([]byte,error)` takes an
explicit absolute UTF-8 exercise directory and a separately reviewed lowercase
64-hex candidate SHA256. The root cannot contain control characters or be a final
symlink, including with trailing separators. Ancestor aliases and dot-dot retain
actual host traversal: paths are not lexically cleaned into another location.
Named/opened directory and file identities are checked through a rooted handle.
The caller must authorize these reads and keep the exercise exclusive throughout.

The directory must contain exactly these six regular leaves:

- README.md
- acceptance_test.go
- candidate.go
- contract.go
- go.mod
- rubric.md

[inputs.json](inputs.json) freezes the original five module files and the exact
native manifest pin `e4fac14e2196dee8dde00bb7e9101b3324cfd4f2f564850a71ca21b76bac9812`.
Those bytes and their digest were independently checked before implementation.
The tool calls the unchanged native preflight and retains its complete JSON/hash,
including all changed or missing original-file observations. It never refreshes
pins, rewrites inputs or substitutes another file. The candidate is checked only
against the caller's independent assertion. Its bytes are opaque: even an empty
or non-Go file can match that assertion without being a valid implementation.

There are no excluded names. Additional entries block even when Go might ignore
them: hidden files, OS-suffixed/ignored Go sources, C/assembly/embed data, go.work,
symlinks, subdirectories and earlier candidate.test binaries. Nothing is removed
or relocated. The preparer's cache/tmp/modcache directories already sit beside
exercise; they must remain outside this census root. Only direct-child names and
metadata are enumerated; no subdirectory or unexpected file contents are read.
On a case-insensitive host, an absent exact original name may still resolve to an
unexpected alias. That ambiguity refuses with ErrBoundary before native content
reads. A truly missing original still produces its complete native missing report.
An absent exact candidate.go is never opened through a differently spelled alias.

The census is capped at 64 entries, candidate reads at 1 MiB and encoded output
at 1 MiB. Candidate metadata, bounded bytes and final named identity must remain
stable. The complete census and identities are checked again after content reads.
An incomplete census or changed boundary produces no admission report. These
checks detect ordinary changes; they are not an atomic multi-file snapshot,
hostile-writer defense, mount/hardlink authentication or OS sandbox. Reads can
update access times. Admission does not verify compiler identity, Go syntax,
imports, initializer purity, expectation quality or semantic approval.

## Reports and command

`model-builder-admission-report/v1` retains `authority: none`,
`executionStatus: not-run` and `reviewStatus: not-established`. A complete report
lists sorted frontier names/kinds, missing and unexpected names, candidate
status/observed hash/byte count, the separately supplied assertion, original root
spelling, fixed manifest identity and full native preflight text/hash.
`frontierComplete: true` means the bounded census completed, not that it matched.
A matched outcome requires all six regular leaves, matched frozen originals and
a matching stable candidate hash. No build permission or code review is implied.

Candidate status is matched, hash-mismatch, missing, refused or read-failed.
Missing/stale originals still retain the accessible candidate observation.
Unread candidates have null observed identities. A complete blocked report can
include a read-failed candidate; it does not claim the candidate was inspected.

From a trusted exclusive checkout's modeling/lifecycle directory, use installed
Go 1.27.3 offline and a new private output directory outside the exercise/checkout:

```sh
set -eu
: "${ADMISSION_WORK:?set a new private absolute output directory}"
: "${EXERCISE:?set the explicit exercise directory}"
: "${CANDIDATE_SHA256:?set the independently reviewed candidate digest}"
mkdir -m 700 "$ADMISSION_WORK"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
go build -o "$ADMISSION_WORK/modelbuilderadmission" ./cmd/modelbuilderadmission
"$ADMISSION_WORK/modelbuilderadmission" check --root "$EXERCISE" \
  --candidate-sha256 "$CANDIDATE_SHA256" > "$ADMISSION_WORK/admission.json"
```

This builds/runs the trusted checker only. It does not build the exercise.
The CLI reads no stdin, accepts those two flags once each in either order, and
writes one buffered report. Exit 0 is matched; 1 delivers blocked; 2 rejects
arguments/root without a report; 3 is failed/short stdout; 4 is incomplete census
or an unexpected I/O/internal boundary, with no complete-admission report.
ErrInvalid and ErrBoundary distinguish the initial and incomplete API failures.
SIGPIPE is ignored so closed stdout exits 3. Inspect exit and full receipt together;
a failed delivery can leave a prefix. Paths and names may be private.

Admission is a point-in-time byte/frontier check. Keep original instructions,
live ownership, exact-source review and later build/execution authority separate.
The preparer and its fifteen pins remain unchanged. The tool adds no foundry route.

## Validation limits

Independent fixtures froze the original manifest, opaque/empty candidate cases,
extra-source categories, missing/changed originals, candidate refusals and root
traversal before production. Tests add exact-spelling host aliases, inclusive
limits, unreadable candidates, concurrency, preservation and actual CLI delivery.
FIFO fixture probes use an available host utility; candidates themselves are
never compiled or executed. Permission probes skip if host privileges bypass mode
bits. These finite observations do not qualify all hosts or concurrent mutations.
