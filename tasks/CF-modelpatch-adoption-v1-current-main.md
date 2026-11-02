---
id: CF-modelpatch-adoption-v1
status: in-progress
owner: Byte current-main modelpatch adoption worker
repositories: [spik3r/clanker-foundry]
---

# Adopt the existing bounded model patch planner locally

## Implementation claim

The coordinator authorized a fresh local adoption after independent ownership,
dependency and historical-lifecycle review. Current main is
ec1037459207b898b6cd1646cc88ba2e212ae998, tree
5f1f8b48ea7dbf4cce8d1582b01389444f6ce619. This isolated branch is
codex/modelpatch-adoption-v1. Active task records and public PRs were checked;
the historical producer reports no active editing reservation for these paths.

The original planner/CLI source is final local commit
35905f96afd73ecd287675193e63eeec6dab3d8e, unchanged in held snapshot
747f5e5994fd69639bbea75a8c99cf3130fc5575. Its genuine completed local run remains
historical evidence, not this adoption's authority or permission to publish.
Neither its old task/events nor held ancestors are copied onto this branch.
Reviewpack and process-view visual gates, source/export pins, timing publication
and private-consumer work remain separate and untouched.

Exact new frontier under modeling/lifecycle:
- modelpatch/{types.go,oracle_test.go,apply.go,apply_test.go,decode.go,
  decode_test.go,README.md}.
- cmd/modelpatch/{main.go,main_test.go}.
- modelpatch/integration_test.go, authored independently before adoption logic.
Plus this task and its own immutable Project Memory source/events only.
All existing example/tutorial requirements, models, intake records, manifests,
suites and expected values remain read-only. No compiler, runner, workflows,
views, archive engines, old tasks, private pins or runtime are modified.

## Reuse and deliberate changes

Adopt the existing Plan/Execute behavior and finite patch language without a new
runner, repair algorithm or semantic interpreter. The exact seven production
files in its five dependencies are byte-identical between held source and current
main: canonical/canonical.go, impact/{compare.go,decode.go,types.go},
jsonformat/json.go, model/ir.go and strictjson/decode.go. The unused held profile
catalog extension is not adopted. No diagnostics or copied-source closure update
is part of this task.

Pin all nine original source files below. First freeze types.go and the complete
original oracle_test.go unchanged on the new branch, then separately freeze the
new integration expectations. Only then copy existing apply/decoder/CLI source
and implementation tests. Never regenerate literal expected values from planner
or reducer output. Original oracle bytes retain SHA-256
36e90cafe0067b7afd5f9f540e823c9d05cbdd047124bd805ed9e832d9dafc30.

One prospective compatibility repair is authorized: the original CLI promises
exit 3 on failed output writes but a real closed stdout pipe produced SIGPIPE
(exit -13 in the assessment subprocess) with no diagnostic. Keep the existing
stdin/stdout/no-path interface and fixed errors; ignore SIGPIPE in main so Write
returns its error, and add an actual closed-pipe subprocess regression. Do not
count a fake writer alone. This makes adoption bytes deliberately different in
main/test, rather than claiming a byte-identical import. README may document the
adoption, identity distinctions, existing-tool integration and this correction.

## Minimal integration and acceptance

A separate contributor freezes a complete, explicitly named T_prepare transition
replacement against the original source/model digest. The expected candidate is
the existing independently frozen tutorial/model-v2.json: adding G_actor makes a
blank actor refuse, while the already stated named-actor and invalid-command
expectations remain unchanged. Compare full normalized IR and the fixed canonical
candidate digest. Do not infer equal raw source bytes after normalized encoding.

Use the existing acceptanceauthor check API with unchanged old/candidate manifests
and suites. The old manifest/model binding must be stale for the candidate; the
candidate suite/model can match while the default baseline runner remains stale.
Validation/replay stays unrun where existing binding rules require it. Reapplying
the old patch to the candidate fails its exact before digest; failure returns no
partial candidate/report. A no-op remains a proposal with authority none.

This integration does not duplicate the tutorial's subprocess orchestration or
claim that the planner executes behavior. The existing tutorial independently
builds and replays its frozen candidate; this task proves that the explicit patch
reproduces that candidate's normalized meaning and preserves stale evidence.
Actual intent review and independent expectations remain prerequisites, not
statuses inferred from matching hashes or passing tests.

## Qualification and boundaries

Run original oracle self-checks before engine adoption, then focused and full
actual Go 1.27.3 offline race/vet, modelc/modeltrace drift, schema, repository and
full committed-diff Project Memory checks. Retain expected compiler/generated and
all existing fixture bytes. Behaviorally reject isolated stale-check, omitted-edit
and partial-result mutations without counting build failures. Independently review
exact final source and lifecycle continuity. Use normal bounded test subprocesses
only; no product process runner, deployment or worker activation.

This task is local-only unless the coordinator obtains separate publication
permission after qualification. Historical local completion, current review and
fresh test evidence do not waive any visual or publication hold.

## Original source identity manifest

These are provenance pins for the reused local source, not authentication or
executed-code evidence. Main/test and README may change only as specified above.

| Path under modeling/lifecycle | Git blob | SHA-256 | Bytes |
|---|---|---|---:|
| cmd/modelpatch/main.go | 5392108c14b67a11e9eecfd8f3e4aa1421874abd | 9989b362c7390bef5e44bdd24ea12cfdaf1ba4b3026ed0d1536402459f534b84 | 859 |
| cmd/modelpatch/main_test.go | 0f87bbbaa278561f3031a6509f6fe2f2146b163b | fed8c16dd6ee6e1f478fa2ca90c424f09745398a22aa1fbd6392b80d8679900c | 3075 |
| modelpatch/README.md | b88b4aa7c7b9cbeb38db9d1711bfa9dd4957da2c | 1854e8fff18986ef1c93e6b2c9b3cd0f5ca19fb2ac805d0509be7b564ce46527 | 8896 |
| modelpatch/apply.go | c305e9199e200cdfcbeebcbcf5e816408997fcfc | 9406ad1fa1a339b95d2e3294f83cb9692f1c4b4f646831dcc37fbb72f27c5173 | 8103 |
| modelpatch/apply_test.go | 837b6e45535087d456a909ac8751b2b9dd2b8297 | 1387d8eee4a3f4c98ffc6f76321f6bc52ae259e6492ade9499e5e043d2e1a178 | 13928 |
| modelpatch/decode.go | 436d8687cbe9af371d497a37dcc3a57958ad792f | 04c9e266419d7be10d0c3b2b65e62f16ad0542822447f67caf8320bbf2bb56f4 | 8690 |
| modelpatch/decode_test.go | f2d0ccb5cc14bf3710c9cbe561b89048e68b997d | 335d50932ac545416852ce2c77e3ecca14ef7e57dc27a85e939e6a2aa3e9be39 | 10952 |
| modelpatch/oracle_test.go | fc9bb957c812056a9d20e6e5628e19aa77fe3e11 | 36e90cafe0067b7afd5f9f540e823c9d05cbdd047124bd805ed9e832d9dafc30 | 14980 |
| modelpatch/types.go | 78b3caf2c89be36f434deb3b793bd0e56e30fb78 | 783e0d89ec3213cf923f07985a82858e8877cdac580ba8bcfa4f067e81d7dbc9 | 3204 |
