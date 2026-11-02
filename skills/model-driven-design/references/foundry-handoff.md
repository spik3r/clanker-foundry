# Foundry lifecycle/v1: a first-design handoff

Use this route only when the requested behavior fits the existing finite lifecycle
profile. It is not an application generator. Do not imply that this skill's
installation includes the Foundry checkout or Go toolchain.

## Find the actual tools and fit

Locate an existing trusted checkout, inspect its current revision/claims and read
its `AGENTS.md`. In that checkout, consult these repository-relative paths:

| Need | Existing source of truth | What it does not establish |
|---|---|---|
| Language fit | `modeling/lifecycle/profilecatalog/README.md`, `model/ir.go`, `cmd/modelprofile` | A capability list does not validate a draft. |
| JSON shape | `modeling/lifecycle/authoring/README.md` and its schema | Parsed schema validation misses strict-byte and semantic refusals. |
| Model plus intent identities | `modeling/lifecycle/intakemodel/README.md`, `cmd/modelintakecheck` | Accepted model bytes can coexist with gaps or false author labels. |
| Literal suite contract | `modeling/lifecycle/scenarios/README.md` | Expectations are not observed behavior. |
| Independent expectation record | `modeling/lifecycle/acceptanceauthor/README.md` | A manifest author label is not proof of independence. |
| Generated draft views/code | `modeling/lifecycle/README.md`, `cmd/modelc` | Structural tests and a diagram do not approve meaning or prove execution. |
| Separately generated candidate | `modeling/lifecycle/qualification/README.md` | Export/verify copies the suite as opaque bytes and does not run it. |

The profile has fixed Task/Command/Review data, finite states/transitions, typed
allowlisted guards and a short ordered-effect vocabulary. Map the actual brief to
those limits. Do not invent fields for payments, databases, dates, parallel work,
authentication, persistence or external effects. Either exclude such requirements
explicitly from a bounded proposal or stop for a different profile/engineering plan.
A claim of UML/SysML/BPMN/UTP/fUML support needs a separately reviewed mapping.

Go operations use the installed exact Go 1.27.3 with offline module resolution.
Do not fetch a compiler or dependencies merely to make the example run. The schema
can help an editor without installing Node/Ajv; authoritative model validation is Go.

## Keep the two authoring paths separate

The model proposer works from the agreed brief and selected profile. The expectation
author receives the same requirements/public contract and stable interface IDs, not
generated implementation or its outputs. Freeze their literal suite bytes and record
case intent before running candidate behavior. Independently review disputed policy
rather than editing expected results after seeing a run. There may be several valid
designs; choosing one is not a deterministic language-model operation.

The worked files in `../examples/` are fictional proposals. Their intake is a prior
declaration with null observations; their human-use question is intentionally open.
Do not import their choices as answers for a different person. The example manifest
was authored separately from its model, but still requires human meaning review.

## Worked byte-preserving checks

The following commands are for an already authorized local review-preparation task.
Set these three absolute paths yourself from the authorized environment:

- `FOUNDRY_ROOT`: a **disposable, exclusive checkout** of the reviewed Foundry
  revision, with a valid Git worktree and its own allowed write scope. Never point
  this at another contributor's active checkout. Preparing it is a separate local
  operation governed by the repository's instructions; no clone happens below.
- `PROPOSAL_DIR`: the approved proposal directory. For the worked example it is
  the skill's `examples` directory; it may be outside the disposable checkout.
- `HANDOFF_WORK`: a new, private, writable directory for reports/tools. Do not
  reuse an occupied directory or overwrite existing reports.

Inspect the paths first. Keep original proposal bytes untouched. The commands
below compile existing trusted tools, not candidate behavior:

```sh
set -eu
: "${FOUNDRY_ROOT:?set the disposable checkout}"
: "${PROPOSAL_DIR:?set the proposal directory}"
: "${HANDOFF_WORK:?set a new private output directory}"
export FOUNDRY_ROOT PROPOSAL_DIR HANDOFF_WORK
cd "$FOUNDRY_ROOT/modeling/lifecycle"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off GOFLAGS=
test "$(go env GOVERSION)" = go1.27.3
go build -o "$HANDOFF_WORK/modelprofile" ./cmd/modelprofile
go build -o "$HANDOFF_WORK/modelintakecheck" ./cmd/modelintakecheck
go build -o "$HANDOFF_WORK/modelscenariodiff" ./cmd/modelscenariodiff
go build -o "$HANDOFF_WORK/modelc" ./cmd/modelc
"$HANDOFF_WORK/modelprofile" > "$HANDOFF_WORK/profile.json"

python3 - <<'PY' > "$HANDOFF_WORK/intake-request.json"
import json, os, pathlib, sys
p = pathlib.Path(os.environ['PROPOSAL_DIR'])
r = json.loads((p/'intake.json').read_bytes())
# Add observed exact UTF-8 bytes to a NEW request; never repair declarations.
r['observed'] = {role: (p/name).read_bytes().decode('utf-8') for role, name in
                 [('requirements','brief.md'), ('model','model.json'), ('scenarioSuite','suite.json')]}
sys.stdout.buffer.write(json.dumps(r, ensure_ascii=False).encode('utf-8'))
PY
"$HANDOFF_WORK/modelintakecheck" < "$HANDOFF_WORK/intake-request.json" > "$HANDOFF_WORK/intake-model-report.json"

python3 - <<'PY' > "$HANDOFF_WORK/suite-check-request.json"
import os, pathlib, sys
p = pathlib.Path(os.environ['PROPOSAL_DIR'])
m, s = (p/'model.json').read_bytes(), (p/'suite.json').read_bytes()
# Embed original JSON tokens, not parse/re-encode them and lose duplicate keys.
sys.stdout.buffer.write(b'{"formatVersion":"lifecycle-scenario-diff-json/v1","beforeModel":'+m+
    b',"beforeSuite":'+s+b',"afterModel":'+m+b',"afterSuite":'+s+b'}')
PY
"$HANDOFF_WORK/modelscenariodiff" < "$HANDOFF_WORK/suite-check-request.json" > "$HANDOFF_WORK/suite-check-report.json"
```

A no-op suite comparison reuses existing strict model/suite binding validation
without execution. `reviewRequired=false` only says the two supplied copies are
identical; it is not a review result. It does not check the acceptance manifest or
requirements meaning. A malformed/stale suite is rejected; no expectation is fixed.

Read the intake report's decoded `intakeReportJSON` alongside `model.status`.
Every artifact/dependency match and unresolved question matters independently.
`modelintakecheck` exits 0 even for a complete report with missing/rejected model;
never promote its process exit alone to success. Observations must be exact bytes:
`read_bytes().decode('utf-8')` avoids universal-newline translation. Source hashes
and canonical model identity are different, and self-declared matches are not
provenance. These reports may contain sensitive proposal text.

## Generate draft review views, not approval

Only in the authorized disposable checkout, after strict validation, replace its
fixed example input with the proposal. Preserve originals in the proposal directory.
The existing generator only owns its fixed `generated/` frontier and refuses unsafe
or inconsistent outputs; do not delete/repair that frontier to force success.

```sh
set -eu
cp "$PROPOSAL_DIR/model.json" "$FOUNDRY_ROOT/modeling/lifecycle/examples/lifecycle.json"
cd "$FOUNDRY_ROOT/modeling/lifecycle"
"$HANDOFF_WORK/modelc" validate
"$HANDOFF_WORK/modelc" generate
"$HANDOFF_WORK/modelc" check
```

Review `generated/decision.svg`, `decision-table.json`, `lifecycle.go`,
`model_paths_test.go` and `trace.json` from that exact proposal. Keep their identities
with the original brief/model/suite. Inspect the actual diagram through an allowed
viewer and compare the table/guards/effects with the requirements. If visual
inspection is unavailable or denied, record it as pending; XML parsing is not a
substitute. No renderer or standards interchange is added by this workflow.

These commands **do not run the new reducer**. Do not run the copied module's full
baseline acceptance suite and assume its default-model fixtures apply to the new
model. Model changes invalidate bindings; a stale refusal is not a behavioral test.
The worked handoff's separate Go integration check runs from the unmodified
reviewed checkout's `modeling/lifecycle` (not the modified disposable copy):
`go test -race -count=1 ./authoringhandoff`. That fixed-example test checks declarations and generated syntax/table consistency,
including the example manifest through the existing bound check-only API. It does
not check arbitrary files named by `PROPOSAL_DIR`, execute its suite or establish
visual/human review.

## Handoff and the later execution boundary

Fill the review-handoff template with actual artifact identities and checks. Keep
unanswered questions, independent semantic/visual review and execution authority
pending where appropriate. Generated tests are structural checks, not the independent
behavior oracle. A successful generation is not an accepted design.

For a custom candidate, the legacy `modelplay` and `modelacceptance` binaries remain
bound to their compiled default model; they cannot select a new reducer by changing
request metadata. Later authorized work can use the existing qualification exporter
and its explicit generated-project GUIDE, or the reviewed Go bound-runner API with
the actual generated package constants and a field-only adapter. Export/verify does
not validate or run the opaque suite, authenticate the binary or authorize execution.
Read the selected source-policy contract before using an archive.

Replay only after the required human decisions and separately reviewed expectations
are in place and candidate execution is authorized. Keep the literal expectations
frozen, record actual invoked code/toolchain and results, and obtain independent
review of discrepancies. A matching replay still does not authorize publication,
integration or deployment. Surrounding product behavior remains separate engineering.
