---
id: CF-revision-review-v1
status: in-progress
owner: Codex revision-review contributor
repositories: [spik3r/clanker-foundry]
---

# Revision review walkthrough across merged tools

## Implementation claim

On `codex/revision-review-workflow` from main `b24962c`, add a separate,
validation-only walkthrough showing how a human compares one invented lifecycle
revision using the merged intake, intake-change, strict intake/model, trace-impact
and frozen acceptance-authoring tools. Reuse the existing tutorial's two models,
manifests and literal suites unchanged. Author two new traceable requirements
snapshots and exact intake declarations; explicitly show that the trace catalog
is not the tutorial scenario suite and leave acceptance links as gaps. The old
suite must become stale for the candidate; the candidate suite is not run by the
baseline compiled runner. Show mismatches and an unchanged unanswered question;
no change is accepted as review, authorization or approval.

Exact scope: new `modeling/lifecycle/revisionreview/README.md`,
`revision_test.go`, `testdata/requirements-v1.json`,
`testdata/requirements-v2.json`, `testdata/intake-v1.json`,
`testdata/intake-v2.json`, `testdata/expected.json`, this record and its own
Project Memory source/events. Independently state and commit literal expected
outcomes before test implementation; tests may call only existing Go APIs and
must not regenerate declarations or expected results from observed tool output.
A guide should distinguish claimed hashes, actual observations, semantic model
identity and replay status. It may instruct a human to inspect existing commands;
it adds no aggregate report, renderer, archive, comparison tool or product runner.

Current main and open PRs 5 and 20, remote branches, tutorial, intake-impact,
trace-impact, intake-model and held workflowacceptance/qualification/reviewpack
claims were inspected. This is a new tutorial for merged change-review APIs, not
an implementation of the held source-closure/evidence-dossier stack. The local
claim is not a shared lock; recheck before publication. No activation, deployment,
formal approval, visual-browser retry, production claim or silent golden update.
