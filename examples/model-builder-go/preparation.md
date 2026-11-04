# Prepare a fresh private Go builder handoff

Use this step when repeating the existing fictional [Go model-to-builder exercise](README.md).
It removes the manual copying, checksum assembly and absolute-path substitution
between the unchanged [packet](packet.md) and a separate builder. It prepares a
reviewable draft, not a new agent invocation or an execution permission.

The new [Go preparer](../../modeling/lifecycle/builderdemo/README.md) uses the
existing [input preflight](../../modeling/lifecycle/builderpreflight/README.md).
Its 15 fixed input hashes identify the sources reviewed at main
`882b5002bd4dc70b4be0060e5e5668d1faa8bbdd`. A checkout with those files unchanged
is suitable even if unrelated files were added later. Installing the skill pack
through Stow does not install this example, native tool or Go toolchain.

## Build the trusted preparation tool

Use an explicitly located trusted checkout, installed native Go 1.27.3 and an
existing authorized output parent outside the checkout. These commands compile
the preparation tool only. They do not compile the frozen acceptance module or
run a candidate. Substitute the three absolute host locations before use.

```bash
set -euo pipefail
ROOT=/exact/trusted/clanker-foundry
WORK_PARENT=/exact/authorized/private-output-parent
GO=/exact/installed/go1.27.3/bin/go
: "${ROOT:?}" "${WORK_PARENT:?}" "${GO:?}"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
"$GO" version
test "$("$GO" env GOVERSION)" = go1.27.3
umask 077
SETUP="$(mktemp -d "$WORK_PARENT/foundry-builder-setup.XXXXXXXX")"
mkdir "$SETUP/tool-cache" "$SETUP/tool-tmp" "$SETUP/tool-modcache"
export GOCACHE="$SETUP/tool-cache" GOTMPDIR="$SETUP/tool-tmp" TMPDIR="$SETUP/tool-tmp"
export GOMODCACHE="$SETUP/tool-modcache"
cd "$ROOT/modeling/lifecycle"
"$GO" build -o "$SETUP/modelbuilderdemo" ./cmd/modelbuilderdemo
"$SETUP/modelbuilderdemo" prepare \
  --root "$ROOT" --work-parent "$SETUP" --go "$GO" > "$SETUP/preparation.json"
printf 'Preparation receipt: %s\n' "$SETUP/preparation.json"
```

The version assertion stops on a different compiler; an executable's name alone
does not verify its version. Read the complete receipt and the `handoff.md` named by
its `handoffPath`. A zero exit must pair with `outcome: prepared-draft`, three
matched reports, `executionStatus: not-run`, `reviewStatus: not-established` and
`authority: none`. The new workspace contains 15 exact input copies and five
module copies, but no candidate or executable. Retain full reports, not only exits.

If source bytes differ, the tool delivers a blocked receipt and creates no
workspace. It never chooses another source or refreshes a hash. The shell above
stops on that nonzero exit; the redirected receipt still records the outcome.
If preparation or stdout delivery fails after creating a directory, keep the
reported workspace path and inspect its partial contents before deciding what
to do. No cleanup or retry occurs automatically.

## Review the bound draft before a fresh builder receives it

The draft contains proposed initial reads, every exact source/module path and
hash, candidate-only write scope, named build locations and an offline compile-only
command represented as JSON. It refers to the unchanged original packet for
semantics rather than generating another model or replacing the frozen tests.
No command in that draft has run. Do not convert path strings into shell code.

The coordinator still has to:

1. Verify current instructions, the current live task/claim and target state.
   Historical packet review and claim records are not current authority.
2. Independently review this bound packet against the original brief, model,
   concrete Go interface and independently frozen expectations. Preserve the
   original [acceptance provenance](acceptance/README.md).
3. Verify the actual Go binary/version, host and formatter separately. Supply
   exact trusted read/edit/build scope and any additional instruction/claim reads.
   Bind Go compiler/standard-library resource reads separately, and authorize a
   check of the exclusive exercise directory before compilation. `go test -c`
   includes matching Go files in that directory, not only a prior manifest list.
4. Keep `Q_use` and visual approval open. Any allowed fictional local experiment
   is a narrow exception, not real-use acceptance.
5. Only then use the [separate-builder workflow](../../workflows/model-driven-implementation.md)
   and its receiving prompt to assign the work. Carry the initial read boundary
   in that trusted receiving prompt, before the first filesystem operation.

The preparer does not dispatch or authorize this final step. It does not inspect
a candidate, supply execution authority or change the original pilot evidence.
A fresh builder must return actual exposure, changed files, frozen hashes and
observed commands/results. Candidate execution remains a later exact-source
review and authorization decision. Even `go test -run '^$'` executes initializers;
only the proposed `go test -c` is compile-only. Compilation is not a test pass.

## Retain the distinction between preparation and a completed demo

The historical [evidence](evidence.md) remains the earlier separate-agent pilot,
not evidence that this new prepared workspace has been used. New preparation
reports establish listed bytes and local file layout only. They do not establish
source authenticity, independent authorship, semantic correctness, a live claim,
review, permission, sandboxing, visual acceptance or product readiness.

This workflow ends with the bound draft. No candidate, second-agent behavior,
Go acceptance run, integration, publication, merge or deployment occurs here.
The responsible coordinator chooses the next separately authorized action.
