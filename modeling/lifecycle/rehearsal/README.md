# Rehearse a design change end to end

This fixed, invented example connects existing review and execution boundaries:
**patch → model and suite review → topology → archive verification → candidate
compilation → literal scenario replay**. It starts from separately frozen data,
not a prompt or an AI call. An AI can help propose requirements or a model, but
its draft remains a proposal: this demo does not collect or confer human review,
approval, identity, ownership or production authority. A real team must review the
requirements, patch, model and expectations using its own process.

The example adds the existing `G_owner` guard to `T_prepare`. A task without an
owner previously prepared successfully; the candidate must refuse it with
`G_owner`. The shared case ID `missing_owner_refuses` is a stable comparison
identifier, not a baseline assertion: the baseline literal expects success.
A named owner still succeeds. The candidate suite also exercises
rework and stale/content-bound review refusals. The baseline and candidate
expectations were separately stated **before** this workflow was implemented;
no expectation, model digest or suite binding is refreshed from a run. The
baseline and expected-result fixtures were authored by a separate contributor
agent, not independently attested as authored by a different person or human.
The original patch, candidate suite and external identity fixture are reused
unchanged from the existing design-change example.

## Run the complete fixed example

From the repository root on a trusted, local, single-writer host, install Bash,
Python 3 and the exact Go **1.27.3** toolchain, then run:

```sh
modeling/lifecycle/rehearsal/run.sh
```

The script accepts no arguments or arbitrary model, command, path or program
input. It starts Python in isolated mode, checks the Go version, restricts the
child environment and disables Go module/network resolution and workspace/ambient
Go configuration. It makes a private temporary module copy,
invokes the fixed repository tools and attempts to remove that directory before
publishing success. A cleanup failure is reported without a success receipt; an
abrupt kill or cleanup failure can leave files behind. `TMPDIR` and Go's cache
may be set for the local host; sufficient
space and an already installed toolchain are required. It does not install or
fetch dependencies, write generated files into the checkout, start workers,
render diagrams, push Git state or deploy anything. On success it emits one
stable JSON result; on failure it returns nonzero without a success result.
Fixed stage diagnostics do not echo fixture or subprocess output. Shell/runtime
errors before the Python workflow starts can be reported by the host itself.

This is **local demo orchestration**, not a sandbox or general product runner.
It executes repository code, the installed Go compiler and generated Go in a
disposable copy; all are trusted. Each fixed subprocess has a 180-second timeout,
but the workflow does not provide OS isolation, resource containment, descendant
cleanup guarantees, authentication or protection against a hostile concurrent
host writer. It is not an arbitrary-process execution interface. A failed write
can expose a prefix of stdout. Fixture IDs and outputs may be sensitive in other
uses and should be handled accordingly.

## What the run checks

1. Verify the frozen oracle and raw request/suite hashes. Build the baseline
   `modelplay` and other fixed tools. Replay two baseline steps and compare every
   actual result to the frozen literals. The candidate-bound suite is rejected
   by the baseline CLI with exit 2 and no report. A separate test also verifies
   the existing bound-runner preflight invokes **zero** callbacks for stale
   bindings; empty CLI output alone would not prove that fact.
2. Run `modelpatch` and independent `modeldiff` on the same declared change.
   Check the candidate model digests, the literal `T_prepare` guard modification,
   direct declared links and review-required flags. These are change indicators,
   not an approval or proof of behavior.
3. Run `modelscenariodiff` on the separately bound baseline and candidate suites.
   Compare all declared case/step classifications and exact semantic suite
   digests to the frozen oracle. Its wire source hashes identify the assembled
   request tokens; they do not authenticate the original files or prove replay.
   Run `modeltopology` for both models and compare every state and component to
   frozen structural expectations. Neither comparison evaluates guards or
   establishes dynamic coverage, liveness or an executable path.
4. Build a `designchange` archive using the existing pinned source policy, then
   verify it against the separately frozen external identities. The consistency
   receipt explicitly says the archived suite is **unvalidated and unexecuted**
   and its binding is **not inspected**. Only after verification, read the exact
   candidate member into memory and check its separately frozen byte hash; the
   script does not extract an archive onto the filesystem. Archive consistency is
   not authentication, review or a replay pass.
5. In the private copy only, install the verified candidate model, generate and
   check the derived reducer and artifacts, run generated structural tests, and
   build a candidate-bound `modelplay`. The old baseline suite is rejected with
   exit 2 and no report. The original candidate suite then executes separately;
   all six actual steps must match frozen complete task/refusal literals. The
   archive's earlier unexecuted status is not rewritten by that separate replay.

The success receipt also records the archive hash and all five generated artifact
hashes. The regression test compares these hashes across two runs; equality is a
byte-consistency observation, not authentication. The generated SVG is neither
rendered nor visually reviewed. The receipt is not a promotion decision.
A passing test or matching digest cannot prove that requirements are correct,
that a human approved them, that other behavior is covered, or that the example
represents production policy. No UML, BPMN, SysML or other standards conformance
is claimed. Diagrams in other workflows are explanatory, not executed or visually
reviewed here.

## Regression checks

With the same toolchain, from `modeling/lifecycle`:

```sh
go test -race -count=1 ./rehearsal
```

The test invokes the actual documented script twice, requires identical complete
receipts, checks that every module file's bytes and mode are unchanged, and checks
temporary cleanup, including an injected cleanup failure after real removal that
must produce no success output. A poison `PYTHONPATH` check confirms isolated
Python startup ignores ambient user customization. It also verifies stale binding against a call-count
spy and compiles a candidate reducer with an intentionally removed owner guard: the
frozen candidate suite must reject that behavioral mutant. These are bounded
negative controls, not exhaustive proof or a substitute for independent review.
