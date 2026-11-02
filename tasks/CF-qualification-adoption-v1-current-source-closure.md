---
id: CF-qualification-adoption-v1
status: completed
owner: Byte current-main offline qualification export adoption worker
repositories: [spik3r/clanker-foundry]
---

# Adopt and freshly qualify the pure archive and offline candidate exporter

## Implementation claim

The coordinator explicitly authorized a fresh separately claimed pure archive
subset and qualification exporter adoption after original-producer ownership
clearance. Active records and absent paths were checked. This branch is
codex/qualification-adoption-v1, rooted at current public main
3d6d17188cc2723cfe04649d0f8b4220a37b1afe, tree
9072c59f387971bb99c0166be3c22ed0463a845d. Initial work is local; exact independent
qualification and a separate coordinator publication decision remain required.

Original qualification source completed at f572a7caacfcc785181d9833f4052a8f9db33018.
Its shared artifactbundle dependency was only checkpointed at
3312a090afac526ddeda392dbd31ed49fa0e04ca under the still-held reviewpack task:
it was NOT completed by the qualifier's historical event. This new claim owns and
qualifies those four pure archive files independently. Neither old task/events nor
held ancestors are imported. The old reviewpack completion and visual gate remain
held; inert ReviewProfile metadata is not adoption or approval of its producer.
Use the original two-profile archive, not the later designchange extension.

Exact new maintained frontier under modeling/lifecycle:
- artifactbundle/{types.go,archive.go,archive_test.go,FORMAT.md}.
- qualification/{qualification.go,qualification_test.go,README.md}.
- cmd/modelqualify/{main.go,main_test.go}.
- qualification/adoption_test.go, independently authored before changed logic.
Plus this task and its own immutable Project Memory source/events only.
No existing main files, Runner/compiler/model/generated APIs, workflows, private
pins, profilecatalog, designchange, reviewpack, views or runtime changes. No new
product process runner, deployment, browser rendering or worker activation.

## Explicit current-source reconciliation

This is an intentional reviewed pin and inventory change, not byte-identical
qualification import. Offline Go dependency inspection confirms the complete
current production package closure below: eleven source files, replacing the old
nine-file checkpoint. The three old scenario files changed through already-merged
Runner work; runner.go and report_decode.go must both be present. The other six
pins stay unchanged. SourceRevision becomes the exact public main commit above.
Update the longest-name bound for scenarios/report_decode.go, the list/hashes,
domain inventory assertions and docs together. The export has seventeen payloads
plus its manifest. Existing old nine-source archives fail this new verifier's
source/inventory policy; do not silently accept, migrate or rebind them.

The archive stays canonical, bounded and extraction-free. Build decodes/compiles
the model in process, copies the supplied suite bytes opaque, and emits the
existing project contract. Verify independently rebuilds with external expected
candidate/suite hashes. Both remain consistency operations, not authentication,
execution, arbitrary-adapter correctness, review or authority. Source pins identify
expected bytes; they do not attest an executing compiler, binary or callback.
The exported default runner compiles against the candidate generated constants;
including the complete Runner package does not authenticate caller bindings.

## Preserve expectations before adoption logic

Pin original nine-file identities below. Before copying archive/exporter mapping
logic, freeze original archive types/tests/format and original qualification test
bytes with their candidate and suite literals. Candidate is 8769 bytes, SHA-256
9674eb213352232f354c2c338709baf6a3230a599ef89845e1cfac5525d4c6e1;
suite is 6656 bytes, SHA-256
53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1.
Their candidate canonical identity remains
a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c.
Never edit these literals to force new behavior or bindings.

A separate contributor freezes the eleven-source/whole-output expectations and
helpers before pin changes or new harness logic. Strengthen the EXISTING candidate
project test rather than add a parallel runner: actual Go 1.27.3 exported-module
build, generated structural tests, full literal four-case/six-step report,
wrong model binding, wrong literal expectations, malformed suite, two existing
compilable generated-code mutants, and restored unmutated build/replay. Keep
verification of a malformed opaque suite separate from its execution rejection.
Add independent exact inventory, missing/extra/tampered source and stale revision
controls; neither missing runner.go nor missing report_decode.go is accepted.

## Bounded prospective repairs

Original archive production/test/format bytes remain unchanged. Qualification
production changes are confined to the reviewed pins/revision/name bound and any
explicitly justified project guidance. README records the new closure and old
archive incompatibility. Original fixture literals in qualification_test.go stay
byte-identical while its existing trusted harness gains finite deadlines,
WaitDelay, capped stdout/stderr, explicit offline build environment and full report
assertions plus restoration. These are test hygiene, not arbitrary-process
sandboxing, cancellation guarantees or descendant cleanup.

The original modelqualify command returns exit 2 for all failures, including short
writes. Actual closed-pipe probes for export and verify both instead exited by
SIGPIPE (-13). Preserve the existing exit 2 policy and fixed diagnostic: freeze
real subprocess regressions before adding signal.Ignore(SIGPIPE) to main. Do not
copy another command's exit 3 contract. Existing fixed bounded source-file reads
and trusted single-writer host assumptions remain; no extraction/launch action is
added to the product command.

## Qualification and evidence gates

Run original canonical ZIP oracle/adversarial tests, original and new exporter
checks, actual current-source exported candidate replay and both closed-pipe cases.
Run full actual Go 1.27.3 offline race/vet, formatting, modelc/modeltrace drift,
authoring schema and repository checks. Independently review exact source,
complete source closure and lifecycle/fixture continuity. Only the separately
observed trusted tests may be called executed; every exported suite and verify
message still says unvalidated and unexecuted. Frozen expectations must not come
from exporter/compiler/reducer output. No held visual or private/timing gate is
waived by passing these checks, and no UML/SysML conformance is claimed.

## Reviewed new copied-source inventory

All paths below are relative to modeling/lifecycle and pinned to public main
3d6d17188cc2723cfe04649d0f8b4220a37b1afe; no source file is edited.

| Path | Bytes | SHA-256 |
|---|---:|---|
| go.mod | 71 | cc128fce13193d3066bdeb08cca9db627fb6c3120967c3beff254a514f2a5fac |
| model/ir.go | 9385 | 13f4ad2980b1ca523b07bdbc8ebeb0e0300f8df81faa8dcd1cdec5ccef007aeb |
| canonical/canonical.go | 1946 | 2c3cba3fd53adf8ca5a4cc2ce3d96b51457742c47b6ab4fcedba219e19001a06 |
| strictjson/decode.go | 4236 | a72ded4af99ffba5f5cca2931c6cace072634ee2e42eb00cd9a5988c691f1ee2 |
| compiler/generate.go | 12315 | e17610a27748a5d3baedf4f59c2183cb184f7e2681b5532c6f9d8d3d1eda6a38 |
| scenarios/types.go | 5904 | 147b12bd1726fec156ba08ec5666a9a362e1563115151ae57bf9face70f86eec |
| scenarios/decode.go | 7047 | e6119229899d3084820d72a777d1771037f871b14554630067b7da022cdedc15 |
| scenarios/run.go | 10114 | 46d41c5d27d87f71f7191f18d182a99c3ca491126afb601599067692c3379dde |
| scenarios/runner.go | 4005 | e85b2146d7c1d16269ff886f6c67084de5e263db8b01a7676b8c7491ed8fb59c |
| scenarios/report_decode.go | 3626 | f14816a19faacf707e23b01d5d16b26be8e8c03a9f5f9af801029b89fb2ceddf |
| cmd/modelplay/main.go | 1613 | 925ec202c74aba5d50eb2bfb5dc54ea22387ccec90e969f1a7ebe1ecbd9a0c63 |

## Original adoption source identities

These identify original bytes at f572a7ca (archive bytes are also exact3312a090),
not original archive completion or publication authority.

| Path under modeling/lifecycle | Git blob | Bytes | SHA-256 |
|---|---|---:|---|
| artifactbundle/types.go | 31d7dd22e8ad78eef79edc2ed9bf20c86b72ce16 | 993 | 8ec41196f19f587799656a3d0033d27220da5fb3aa9e8b14f4bc6de197701782 |
| artifactbundle/archive.go | 48a60c229c2662d128928f6be0a99b8470749634 | 8152 | 6ef8eae172ce23d1ee8e8b30384e712c93154c3dbfb042bb1b3a7cd66047247a |
| artifactbundle/archive_test.go | b42d2f486635a30a805e10de09d9586d3233132b | 9280 | 00bc39c371ea17b34fc9b8e11befe81f85886968bcb1e9cc286960020a8b0b7f |
| artifactbundle/FORMAT.md | a096320b1f47c0d83633cbce1a7d49120a02cde4 | 4412 | c1507c9a10bfc79794aeeb3c678e9f31f77fcf7e5868dcf3d906c12ffcbef3d2 |
| qualification/qualification.go | fba746a69dd2c23cfc42882f67ed943470e00af1 | 6350 | e4fc569cdd5e7c22fc37e68b8dc3c4dfb22a7e62138a2dda0745a366a50646ff |
| qualification/qualification_test.go | 021858f263571da89d5662ebce88918eb4014ef8 | 26893 | c94b4d8ecee0e4c31d6debd61238f25345d808b6f32fd0afd9b2758cbcb21c74 |
| qualification/README.md | 1466f17092d32745fcb3a2fc112af101caa1a14c | 2773 | 747192f707630cd6886eb96108728bd7baa3a0446aacf1bad94033af13a61342 |
| cmd/modelqualify/main.go | f1050c264ba8501730ee68b25002545dac34919e | 3265 | 2df33d2b18cbb1924b94c4098acf61209ea32e4a309534c73e345c7b39601d4d |
| cmd/modelqualify/main_test.go | 5221e9e64d5f91bb0ae262991f6da13b6007efac | 3312 | 8fac69c2d5132fc9e5e95bc2a2e648378c122024b8e256909d97edd38ce41c26 |


## Local completion and independent qualification

Exact reviewed source is 84f2cac542f6226fd5d4b6fb5d291e56b4d41a35, tree
db20406de65db522e6e1bfd73beaab667a392aba. The original producer lead found no
material source, ownership, closure, chronology or authority defect. The oracle
author separately reviewed its unchanged expectations, source identities,
strengthened existing harness and README claims. This newly qualifies only the
claimed pure archive/exporter subset; it does not complete the old reviewpack task.

All four original archive files remain exact3312/f572. Candidate/suite bytes retain
their hashes above. Independent adoption_test.go retains SHA-256
c03f6b4c0dd48af1e840318e3db7fd44087f98fc5cafb88c6c7b725a8177dc19,
frozen at 4623dcc before original production adoption d089ee4 and explicit pin
reconciliation 509e91d. All eleven current copied-source files are unmodified;
no existing main path changed. Public main was rechecked at 3d6d1718 before
completion. The source revision, longest-name bound and seventeen-payload
inventory were reviewed changes, not silent refreshes.

Author and original lead independently passed actual Go 1.27.3 offline full race
checks (43 tested packages), focused verbose tests, vet, formatting, modelc
validate/check, modeltrace check, schema 2/2 with a lock-matched existing installed
dependency closure, and repository validation. The oracle author independently
passed the three affected packages under race. Optional skills-ref is absent.
The wrong-toolchain-only negative is intentionally skipped under required
Go 1.27.3; no unsupported-toolchain execution result is claimed.

The existing exported-project test actually compiled the candidate-specific
current-source module and ran generated structural tests. Initial four-case/six-
step output, the deliberate wrong-expectation mismatch and restored replay matched
full independently frozen reports. Wrong binding and malformed suite rejected.
Both unchanged generated-code mutants compiled and failed behaviorally; original
code was restored, rebuilt and replayed, and all materialized payloads stayed
unchanged. Those observations concern only these trusted fixed tests, not a
product execution service or later callers' suites.

Both actual CLI closed-pipe regressions failed unmodified main at 509e91d with
SIGPIPE and passed the reviewed correction with existing exit 2. The author also
ran a separately built CLI: deterministic new archive verified with the exact
unvalidated/unexecuted status, an actual old nine-source archive rejected with
empty stdout, and both closed-output modes returned 2. The example archive has
17 payloads plus manifest, 101580 bytes and SHA-256
159cf77605273dbf86cd44b8bb43b7c1634b68fbe2cf9f272d9521ff63ad2050.
These extra CLI artifact observations are author-run, not attributed to the lead.

Final metadata-only continuity and full committed Project Memory validation follow
this completion commit. A verified complete-history bundle preserves only fresh
adoption ancestry from public main. No old held task/events are copied, no private
pin is refreshed, and no visual, timing, browser, publication or deployment gate
is waived. Every exported/verified suite remains described as opaque, unvalidated
and unexecuted unless a separate invocation supplies its own observed evidence.
