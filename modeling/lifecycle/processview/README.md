# Lifecycle process review view

This offline view projects an existing `lifecycle-json/v1` model into a
transition-by-transition admission flow. It complements the existing
[state/transition view](../README.md): each flow shows the matching source state
and command, the conjunction of declared guards, the guard-refusal branch, the
revision-limit refusal, ordered effects and the destination. Guard expressions
and requirement IDs are listed in a separate glossary.

It is a static review aid, not a BPMN diagram or executable process model. It
does not model pools, participants, organizational roles, parallel gateways,
timers, subprocesses, retries, events, scheduling or workers. Each flow starts
after input validation and state/trigger matching; invalid input and illegal
transitions are outside the drawn lanes. A guard refusal returns the unchanged
task and sorted failed guard IDs. After the guards pass, revision overflow refuses
without changing the task; otherwise the declared effects apply atomically in the
model's order. The display states those boundaries rather than inferring
satisfiability, execution, authentication, review, or approval.

## Run

From `modeling/lifecycle`, using Go 1.27.3 and offline module resolution:

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
outdir="$(mktemp -d "${TMPDIR:-/tmp}/foundry-process-view.XXXXXX")"
go build -o "$outdir/modelprocess" ./cmd/modelprocess
"$outdir/modelprocess" svg < examples/lifecycle.json > "$outdir/process.svg"
"$outdir/modelprocess" json < examples/lifecycle.json > "$outdir/process.json"
printf 'Review files in %s\n' "$outdir"
go test ./processview ./cmd/modelprocess
```

The command reads at most 1 MiB of model source from stdin and accepts exactly
`json` or `svg`; both projections are built before output and each is capped at
16 MiB; SVG canvas height is capped at 30,000 pixels. It refuses either
oversize projection without output, rather than emitting a cropped view. The command
opens no input or output paths. Shell redirection creates the example files.
Exit 0 emits one complete projection; exit 2 means invalid input, unsupported
arguments/toolchain or a resource bound, with empty stdout; exit 3 means an output
write failed or was short, which can leave a partial stream. Keep input and output
private when the model contains private data.

The JSON projection is `lifecycle-process-view/v1`. It includes the model ID and canonical
model digest, normalized states, transitions and guards, ordered effects, and
explicit semantic boundaries. The SVG displays the same transition and guard
facts in an inert flow layout. SVG displays control and non-ASCII characters as
Go-style escapes to preserve XML safety; literal backslashes are doubled so
they cannot be confused with escapes. JSON preserves state labels; its guard
expressions are readable strings with Go-quoted literals, not a source AST. Both
derive their identity from the validated, normalized model rather than a caller-supplied digest. Equivalent source formatting
and semantic-set order produce identical output; effect order remains meaningful.
Neither the digest nor a visually complete flow establishes correctness or a real
execution. The tool does not change `modelc` or its owned generated files, and its
outputs are not included in the existing trace manifest.
