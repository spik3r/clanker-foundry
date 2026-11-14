---
id: CF-foundry-stdout-v1
status: in-progress
owner: Codex entrypoint output contributor
repositories: [spik3r/clanker-foundry]
---

# Report failed Foundry output delivery

## Implementation claim

Owner: Codex entrypoint output contributor, assigned by the coordinating session
under the owner's continued maintenance and reviewed-merge request. Base: main
at `b8f729d`, after the filesystem alias repair. The separate generator stdout
task owns generator and contract files; this task changes only foundrygen.

Run discards stdout and stderr write errors and counts. A successful command
can report exit 0 after its result or warning was not delivered. At the process
entrypoint, a closed stdout pipe can terminate the process with SIGPIPE before
the library can report an I/O failure.

Deliver nonempty buffered warnings before stdout. An internally successful run
must return exit 1 and leave stdout untouched if warning delivery fails or is
short. Preserve an existing nonzero command result if only diagnostic delivery
fails. Then deliver nonempty stdout, treating errors and short writes as exit 1
with a fixed diagnostic. Do not expose writer error details. Ignore SIGPIPE only
in process main so failed pipe writes reach ordinary error handling.

Keep normal bytes and artifact behavior unchanged. A failed output delivery can
leave a partial stdout prefix and already-generated artifacts; do not promise
rollback. Preserve primary refusal and drift behavior and skip empty writes.

Scope: `modeling/foundrygen/run.go`,
`modeling/foundrygen/cmd/foundrygen/main.go`, new
`modeling/foundrygen/stdout_test.go`, `modeling/foundrygen/README.md`, this task
and new immutable Project Memory files. No generator, version, contract,
consumer or runtime changes.

Checks: demonstrate current false success for writer errors and short writes;
exercise normal, warning, refusal and drift controls, unchanged artifacts after
failed delivery, and finite subprocesses with preclosed pipes. Run foundrygen
vet/full race, repository validator, whitespace and committed Project Memory
validation. Push and open a PR for independent review; this contributor must
not merge.
