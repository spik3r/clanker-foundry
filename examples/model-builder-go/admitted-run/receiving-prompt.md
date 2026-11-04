# Logical receiving prompt for the admitted trial

This is the maintained logical rendering of the edit-only prompt. The coordinator
replaces @WORKSPACE@, @TRIAL@, @PYTHON@ and @PYTHON_RESOURCE_ROOT@ with the exact
private locations, records its SHA256 and obtains two reviews before dispatch.
No substitution is shell code. The actual fully bound prompt is retained privately;
this file is not itself a grant to a future builder. Current task/claim and inputs
must be independently checked again before reuse. The fresh fork_turns none
invocation initially grants only the exact bound-prompt file read/hash and named
parent metadata, using a recorded expected digest. Its reviewed outer message
requires a non-login shell and no discovery before that first read. Only the
verified complete prompt supplies the further named input and edit grants.

---

# Fresh fictional Go builder: edit-only assignment

You are the fresh builder for run CF-builder-admitted-run-v1, dispatched with
fork_turns none. Do not use any prior candidate, generated implementation,
conversation or other worker's control output. This is one private fictional
FirstDesignDraft experiment, not production or human acceptance. Q_use and visual
approval remain open. Implement the unchanged bounded Submit fixture only.

## Current authority before your first filesystem action

The coordinator has verified the public claim at commit
7311c677da96bf0346e6cd40e0dbcc2e6225b36a (PR88) and recorded a genuine start.
Its current stage is edit-only. Read the explicitly named current instructions
and claim copies below; no discovery is needed. Their hashes identify these
copies, not authenticated provenance. Parent coordination owns repository/Project
Memory bookkeeping; do not run Git or lifecycle scripts from this private exercise.
If current instructions conflict or a named input is unavailable, stop the
affected work and report the exact missing input/conflict. Do not search for a
replacement or infer permission from a copied packet.

The ONLY user-data content reads initially granted are the exact files in the
following JSON array. Read their complete applicable instructions before editing.
Use these absolute bound paths only. Relative links in those files do not grant
additional reads. The historical packet's review/exception, suggested Git reads
and draft build command are not current grants.

[
  {
    "path": "@WORKSPACE@/inputs/skills/model-driven-design/examples/brief.md",
    "sha256": "ab3d47aaacdcdc8d57c2cb8a9ab432c3f5c22c42b7df11b707b8af03f32eb5de",
    "bytes": 3569,
    "role": "requirements"
  },
  {
    "path": "@WORKSPACE@/inputs/skills/model-driven-design/examples/model.json",
    "sha256": "3e7b8c44c10940a7524161dd306d022019e8fc41e793a66a11771d6dd8621239",
    "bytes": 2001,
    "role": "model"
  },
  {
    "path": "@WORKSPACE@/inputs/modeling/lifecycle/README.md",
    "sha256": "26cc7d1ea82f6f09090d15b0572560bd891081505631483a2adb3ce402a754b0",
    "bytes": 9893,
    "role": "interface"
  },
  {
    "path": "@WORKSPACE@/inputs/examples/model-builder-go/acceptance/README.md",
    "sha256": "b4ec648c38dadd63dc9799f40d6183d3e464c9dc33b83302c6f6c171a8149103",
    "bytes": 4971,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/inputs/examples/model-builder-go/acceptance/go.mod",
    "sha256": "e31637eba9a06a2105c2a12b904457b50ad0c60defbbc7cafabeae8f68d4bc3c",
    "bytes": 70,
    "role": "harness"
  },
  {
    "path": "@WORKSPACE@/inputs/examples/model-builder-go/acceptance/contract.go",
    "sha256": "82b6dcd7a05d496a31c7431dce9b61bb572dd5f53b312418dc8ff72fb7a35b0a",
    "bytes": 1111,
    "role": "interface"
  },
  {
    "path": "@WORKSPACE@/inputs/examples/model-builder-go/acceptance/acceptance_test.go",
    "sha256": "eaf6078d5703f3e3664f41d0a19d68a7a66bafcfac310bfe5c91e72a7fe02518",
    "bytes": 13371,
    "role": "acceptance"
  },
  {
    "path": "@WORKSPACE@/inputs/examples/model-builder-go/acceptance/rubric.md",
    "sha256": "ec13d76787d6ce4976487d373c8f2d3587286bdb1f1b1716419b397b46569ce5",
    "bytes": 7686,
    "role": "rubric"
  },
  {
    "path": "@WORKSPACE@/inputs/examples/model-builder-go/packet.md",
    "sha256": "56ff2a75570000320520805fe734f8e9121c580db9ef0d3c89d1da7bbebf0fdf",
    "bytes": 17261,
    "role": "packet"
  },
  {
    "path": "@WORKSPACE@/inputs/AGENTS.md",
    "sha256": "4574ee25fe89b6658ae237f3f50648ef1e542c487a2c23d7ab2833696f9455eb",
    "bytes": 7325,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/inputs/global-agents/AGENTS.md",
    "sha256": "23e6023c3dae5a6e745def8b7f12335115da3982dfd94144697ca0f0dea3d122",
    "bytes": 6167,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/inputs/workflows/model-driven-implementation.md",
    "sha256": "85a88fec7dc8443c0cef099a9604ecdd685e00cf01166ebadd8d8b845df449d0",
    "bytes": 10386,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/inputs/templates/model-driven/implementation-handoff.md",
    "sha256": "24fe9f38e2c9a767f084071d24ebadf3db0ad60556adc3f52b60f27397413c80",
    "bytes": 6393,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/inputs/templates/prompts/implement-from-model.md",
    "sha256": "c3307cc80d2400001b455d491cfef0c99b37681230428aae88ab95697ae9ffe5",
    "bytes": 4815,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/inputs/skills/project-memory/SKILL.md",
    "sha256": "7fbde5e0a9b6bbb8303b16d9d3f753edadb6e4db6cd15873f93d26d5713c3a70",
    "bytes": 8917,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/exercise/README.md",
    "sha256": "b4ec648c38dadd63dc9799f40d6183d3e464c9dc33b83302c6f6c171a8149103",
    "bytes": 4971,
    "role": "instructions"
  },
  {
    "path": "@WORKSPACE@/exercise/go.mod",
    "sha256": "e31637eba9a06a2105c2a12b904457b50ad0c60defbbc7cafabeae8f68d4bc3c",
    "bytes": 70,
    "role": "harness"
  },
  {
    "path": "@WORKSPACE@/exercise/contract.go",
    "sha256": "82b6dcd7a05d496a31c7431dce9b61bb572dd5f53b312418dc8ff72fb7a35b0a",
    "bytes": 1111,
    "role": "interface"
  },
  {
    "path": "@WORKSPACE@/exercise/acceptance_test.go",
    "sha256": "eaf6078d5703f3e3664f41d0a19d68a7a66bafcfac310bfe5c91e72a7fe02518",
    "bytes": 13371,
    "role": "acceptance"
  },
  {
    "path": "@WORKSPACE@/exercise/rubric.md",
    "sha256": "ec13d76787d6ce4976487d373c8f2d3587286bdb1f1b1716419b397b46569ce5",
    "bytes": 7686,
    "role": "rubric"
  },
  {
    "path": "@WORKSPACE@/handoff.md",
    "sha256": "8dae275ddbafcb8cd92d1e1cbf8d98c4ac67d2a806913dafce96ce5d2695a0c7",
    "bytes": 13471,
    "role": "proposed-handoff"
  },
  {
    "path": "@WORKSPACE@/source-manifest.json",
    "sha256": "6a87f4db69a9c43376b405aca0d1252a662a3a4cdc7496329f2b41fac2896032",
    "bytes": 3154,
    "role": "source-manifest"
  },
  {
    "path": "@WORKSPACE@/exercise-manifest.json",
    "sha256": "fcf4bb3dd0b8195179fe671c063d234b5a9f43a7321c38fa559ace808d8bf1cd",
    "bytes": 947,
    "role": "exercise-manifest"
  },
  {
    "path": "@WORKSPACE@/source-report.json",
    "sha256": "a2dc97345910a134317b4844db94f6a6911e1e59b2d6afa54088628233290406",
    "bytes": 5705,
    "role": "preflight-report"
  },
  {
    "path": "@WORKSPACE@/inputs-report.json",
    "sha256": "a2dc97345910a134317b4844db94f6a6911e1e59b2d6afa54088628233290406",
    "bytes": 5705,
    "role": "preflight-report"
  },
  {
    "path": "@WORKSPACE@/exercise-report.json",
    "sha256": "e5be849a4d51217e79d39d3ebb66cb2388fffc3072c4bbc22802c108fa4697a2",
    "bytes": 1936,
    "role": "preflight-report"
  },
  {
    "path": "@WORKSPACE@/authority/pinned-claim.md",
    "sha256": "3e126eeba4888221bfefe7b4d436b9061d9a13b2ced8cf625d33c51397fad73d",
    "bytes": 8236,
    "role": "current-pinned-claim"
  },
  {
    "path": "@WORKSPACE@/authority/live-claim.json",
    "sha256": "1283f10e98f0713e8e8d49fcd58a80fb0cd8efc28926c319cdbefa33c8cfc961",
    "bytes": 533,
    "role": "current-claim-observation"
  }
]

You may inspect metadata for those named files, the not-yet-created
@WORKSPACE@/exercise/candidate.go, and these exact parent directory entries only:

[
  "/private",
  "/private/tmp",
  "@TRIAL@",
  "@WORKSPACE@",
  "@WORKSPACE@/authority",
  "@WORKSPACE@/exercise",
  "@WORKSPACE@/inputs",
  "@WORKSPACE@/inputs/examples",
  "@WORKSPACE@/inputs/examples/model-builder-go",
  "@WORKSPACE@/inputs/examples/model-builder-go/acceptance",
  "@WORKSPACE@/inputs/global-agents",
  "@WORKSPACE@/inputs/modeling",
  "@WORKSPACE@/inputs/modeling/lifecycle",
  "@WORKSPACE@/inputs/skills",
  "@WORKSPACE@/inputs/skills/model-driven-design",
  "@WORKSPACE@/inputs/skills/model-driven-design/examples",
  "@WORKSPACE@/inputs/skills/project-memory",
  "@WORKSPACE@/inputs/templates",
  "@WORKSPACE@/inputs/templates/model-driven",
  "@WORKSPACE@/inputs/templates/prompts",
  "@WORKSPACE@/inputs/workflows"
]

A metadata grant is not a directory-content listing grant. Do not enumerate any
directory, glob names, run find/rg --files/ls/pwd, discover AGENTS files, search
ancestors/siblings or read another workspace. Do not access prior implementations,
generators, generated source, historical candidate code, network or unrelated
files. The supplied fifteen inputs and five exercise files were independently
hash-checked by the coordinator; only the five expected files exist in exercise.
You may verify the supplied named file hashes; do not enumerate to discover more.

For named read/hash/text-write operations you may use the assistant file tools,
or trusted host @PYTHON@ with -I -B, and its installed standard-library resources
under @PYTHON_RESOURCE_ROOT@. Those host interpreter/resource reads are permitted
only to perform these literal named-file operations. No third-party imports,
candidate import, module discovery, subprocess from Python, Go command, formatter
or package installation is granted. Prefer one explicit literal file list.

## Only granted write

Create NEW @WORKSPACE@/exercise/candidate.go and write the smallest agent-authored
pure Go implementation that registers the unchanged acceptance.Submit variable.
You may use normal Go standard-library import declarations in source text; do not
load/import/compile/execute that source or a candidate module. Derive behavior
from the supplied brief/model/profile/packet and unchanged frozen interface and
literals. Do not copy a prior implementation or invoke a generator.

After creating your own new @WORKSPACE@/exercise/candidate.go, you may read/hash
only that same file to inspect your edit and report its observed SHA256. This
additional grant covers no prior or other candidate and permits no code import,
compilation or execution.

No other writes are allowed: preserve all five originals, source copies,
manifests/reports and authority files. No formatter, build, test, initializer,
go run, go test (even empty selection), network, persistence/authentication,
input-validator extension, dependency or extra source file is authorized now.
Do not create caches, tmp directories, logs or executables. Return your action
ledger in your final reply rather than writing a separate log file.

## Explicitly withheld later gates

The copied draft proposes exercise/candidate.test. That proposal is superseded:
ALL later products, compiler caches and temporary files will be external to
exercise. It must remain exactly six regular files after your edit and after
any later separately authorized build/run. There is no build grant in this prompt.

After you finish, the coordinator AND independent reviewer will inspect the
entire exact source and SHA. Only then may the coordinator authorize an exact
external-output compile-only command. Only after exact binary/source/input review
may the coordinator separately authorize behavior tests or race execution.
Compilation is not test success; hashes and admission do not supply approval.

## Required return

Report candidate absolute path and observed SHA256, changed paths, semantic
requirement/model/interface links, every actual input read and metadata operation,
all commands/actions in order, any unexpected exposure or attempted forbidden
operation, inaccessible inputs or unresolved questions, and checks NOT RUN.
Do not claim an OS audit or independence solely from your conclusion. State that
compilation, formatting, imports/execution and behavior/race checks were not run.
Then stop for source review; do not continue into build or execution.
