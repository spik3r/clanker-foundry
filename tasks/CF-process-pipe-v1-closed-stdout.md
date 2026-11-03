---
id: CF-process-pipe-v1
status: review
owner: Foundry process CLI repair contributor
repositories: [spik3r/clanker-foundry]
---

# Return the declared process-view status on a closed pipe

## Implementation claim

Owner: Foundry process CLI repair contributor, coordinated by the repository
owner's lead agent. The owner requested repairing PR20, completing its visual
review and merging after review. This supersedes the earlier publication-only
scope for this repair; the coordinator owns visual inspection and integration.
Branch: codex/lifecycle-process-view, original head 51822945. Current main
at a981b3ed is incorporated by an ordinary merge before this claim.

Write only modeling/lifecycle/cmd/modelprocess/main.go and main_test.go, this
record and its new immutable Project Memory source/events. Preserve the original
process-view task, frozen expectations, renderer and all other product paths.
No render or merge action belongs to this builder assignment. The existing PR20
is the matching active lane; no competing implementation is created.

## Acceptance and validation

Reproduce SIGPIPE termination with an actual compiled binary and closed stdout
before repair. Match existing CLI policy by ignoring SIGPIPE in main so a failed
write produces exit 3 and the fixed diagnostic. Add compiled-binary regressions
for both json and svg with closed stdout and complete successful output. Run
Go 1.27.3 full race tests, vet, modelc check, pack validation, installer checks,
whitespace checks and fresh committed Project Memory validation. Obtain an
independent source review and publish to PR20 for the coordinator's visual review
and merge decision. Record observed results and any unavailable checks.

## Repair and review result

The original compiled CLI terminated with SIGPIPE (return code -13 and empty
stderr) for both json and svg when its stdout pipe had no reader. The regression
commit 8272c22 reproduces both failures while complete-output cases pass. Repair
commit ba2fcf4 ignores SIGPIPE in main, following existing CLI policy: both modes
now return exit 3 with the fixed output-write diagnostic. Complete output bytes
still match the existing projection. An independent reviewer checked ba2fcf4 and
passed its focused race tests with no remaining finding.

The coordinator reports native macOS QuickLook rendering and visual inspection
of the frozen two-lane fixture, normal five-lane lifecycle and a stress fixture
with a maximum-length UTF-8 label and special characters. The complete diagrams
showed no clipping, overlap, missing refusal/effect text or direction issue.
QuickLook cropped non-square thumbnails, so square presentation wrappers embedded
the original SVG bytes at their original dimensions with white padding. Those
wrappers did not alter SVG content. These were native renders, not browser checks,
and cover only the three examples. The current owner explicitly requested this
visual review; the earlier held visual attempt did not restrict this new review.

The coordinator rebuilt ba2fcf4 with Go 1.27.3 and compared exact reviewed SVG
bytes. SHA-256 identities: fixture
8c379fd3ae36974688677ce65c1de850039887c3c88f05d41a26e71d176f70cb;
normal a69a782e53cc1ac60f15295c3b09909064f18403d70d515a60c1837c5dd7947a;
stress 6b96de6fd20bcd3f47fa6bb17191601cd0ea2beb8ccc009e315cca417cb2d0a8.
Images remain local; no rendered image is committed.

Validation on Go 1.27.3: full `go test -race -count=1 ./...` passed across all
packages; `go vet ./...` and `modelc check` passed. The focused CLI tests passed,
including both real-binary closed-pipe modes and both complete-output modes.
Pack validation, installer regressions, gofmt and whitespace checks passed.
Optional skills-ref is absent. Source files and original frozen projection
fixtures are unchanged except the two scoped CLI files. Full committed Project
Memory validation and coordinator hosted checks remain the final handoff gates.
