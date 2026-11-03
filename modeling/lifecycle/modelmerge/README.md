# Three-way model merge proposal

Reconcile two versions of a reviewed lifecycle design against their supplied
common base. The tool uses semantic IDs rather than line positions and refuses
ambiguous concurrent edits. It never writes a model, resolves a conflict for a
reviewer, repairs links, regenerates code or runs a candidate.

```go
result, err := modelmerge.Merge(base, left, right)
```

All inputs are complete `model.Model` values. Existing normalization, validation,
canonical identity and JSON-v1 representability remain authoritative. The tool
adds no alternate lifecycle semantics. Invalid inputs return `ErrInvalid` and a
zero result, even if other inputs would conflict. Input/output envelope limits
return `ErrLimit`. All three inputs are validated before comparing edits.

## Merge decisions

The units are root `id` and `initial` separately, and each complete named state,
requirement, guard or transition. A record's ID is its identity within its kind;
renames are removal/addition, not guessed matches.

| Left/right versus base | Decision |
|---|---|
| Left equals right | Keep that value, including a shared deletion |
| Only right changed | Keep right |
| Only left changed | Keep left |
| Both changed differently | Report a conflict; expose no candidate |

Equality uses the existing normalized IR. Source formatting, collection order,
transition guard order and pure `all`/`any` operand order therefore do not create
conflicts. Effects remain ordered. Equality is not behavioral equivalence.

Named records are atomic: two different edits to different fields of the same
transition still conflict. This deliberately conservative rule avoids silently
combining a changed destination with an independently changed guard policy.
Separate records can merge, but the complete result must pass existing semantic
checks. For example, individually valid additions can together make state/trigger
dispatch ambiguous. That is an `invalid-combination`, not a partly accepted merge.

The deterministic result has one of three statuses:

- `merged`: owned normalized candidate and canonical candidate digest.
- `conflicted`: all conflict locations, no candidate or candidate digest.
- `invalid-combination`: no atomic conflicts, but the combined model is invalid
  or not representable within JSON-v1 limits; no candidate or candidate digest.

Conflict records contain `kind`, `id`, `field`, `reason`, sorted by kind, ID then
field. Root fields use `kind=model`, empty `id`, and `field=id` or `initial`.
Named records use an empty field. Reasons are `add-add`, `remove-change` and
`change-change`. Reports omit competing payloads; consult the supplied originals.
No partial candidate is returned alongside a conflict. Swapping left and right
changes their identity labels, not the chosen model, conflict list or status.

## Strict JSON and stdin CLI

From the lifecycle module, with Go 1.27.3 installed:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelmerge < reviewed-merge.json
```

The request has exactly four required keys:

```text
formatVersion: "lifecycle-model-merge-json/v1"
base:  complete lifecycle-json/v1 model object
left:  complete lifecycle-json/v1 model object
right: complete lifecycle-json/v1 model object
```

Duplicate, unknown, case-variant, missing, null, malformed Unicode and trailing
input refuse. Whole input is bounded to 4 MiB/depth 66; embedded models retain
1 MiB and their existing semantic limits. Output is capped at 2 MiB and fully
buffered before returning bytes. No arguments, file paths, network, subprocesses,
installation or model-provided code are accepted by the tool.

Output has `mergeVersion`, `profileVersion`, `mode`, `authority`, `baseDigest`,
`leftDigest`, `rightDigest`, `sources`, `status`, `candidateDigest`, `candidate`
and `conflicts`, in that order. The candidate is an existing normalized JSON-v1
model document or null. `sources` records exact request and embedded model-byte
SHA-256 digests for `Execute`; typed `Merge` leaves it null. Whitespace changes
can change those source hashes without changing canonical identities or decisions.
No public encoder turns a caller-supplied report into claimed observations.

Exit 0 means a report was produced, **not that the models merged**: inspect status.
Exit 2 means rejected input/read failure and no stdout; exit 3 means a failed or
short stdout write, including a closed pipe. Output is fixed-order JSON, without
HTML escaping, followed by one LF. Repeating identical input gives identical bytes.

## Review boundary

The base is supplied by the caller; the tool does not inspect Git ancestry or
prove it is a real common ancestor. A digest is consistency evidence, not identity
or authorship proof. Models may contain private text; output candidates preserve
that text. Do not send output to an unapproved destination.

All results say `example-only`, `authority=none`. A merged design is not an
approved design, compatible scenario suite, valid implementation, behavioral
proof or permission to continue. Existing review, [model patch](../modelpatch/README.md),
[change comparison](../impact/README.md), generation and independent acceptance
steps keep their own gates. This command does not update evidence bindings.

## Verification

The independent `oracle_test.go` freezes full expected models and complete conflict
lists before merge logic. It checks concurrent edits and invalid combinations from
invented designs, not from merge output. Separate tests cover strict input, bounds,
owned copies, deterministic parallel calls and real closed-pipe CLI behavior.
Run from the module with the pinned toolchain and no module downloads:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race ./modelmerge ./cmd/modelmerge
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race ./...
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go vet ./...
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelc check
```

Observed results and exact-source review belong in the
[task record](../../../tasks/CF-model-three-way-merge-v1-concurrent-designs.md).
No production integration, standards conformance or publication is claimed.
