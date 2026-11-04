package builderdemo

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
)

func handoff(workspace, sourceRoot, goBinary string, inventory, exercise manifest, exercisePin string) ([]byte, error) {
	var out bytes.Buffer
	out.WriteString(`# Initial read boundary: draft for coordinator review

Do not dispatch this draft. This file grants no read, edit, build or execution
authority. After the coordinator supplies trusted scope, the first readable files
must be exactly the following files. The inherited working directory is not a
discovery root. No directory enumeration, glob expansion, ancestor/sibling search,
linked-file discovery or unspecified Git operation is supplied by this draft.

`)
	block := func(value any) error {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		out.WriteString("```json\n")
		out.Write(raw)
		out.WriteString("\n```\n\n")
		return nil
	}
	if err := block(map[string]any{"firstReadableFiles": []string{
		filepath.Join(workspace, "handoff.md"), filepath.Join(workspace, "inputs", "AGENTS.md"), filepath.Join(workspace, "inputs", "global-agents", "AGENTS.md"),
	}}); err != nil {
		return nil, err
	}
	out.WriteString(`## Fixed inputs and byte observations

Only the listed subsequent files are proposed reads. Preserve every original byte.
The three saved preflight reports matched at preparation time. Their authority is
none, review is not-established and execution is not-run. This does not authenticate
sources, establish oracle independence or semantic correctness, acquire a claim,
approve this draft or enforce a sandbox. Recheck exact bytes at the authorized
handoff boundary; reports are not atomic snapshots and may become stale.

The unchanged original packet contains the complete model, concrete contract,
trace, fixture choices and return requirements. Its reviewed status, pilot
exception, historical Git commands, source paths and earlier authority describe
the historical pilot only. They are not current grants or permission to follow
links, search for claims or execute commands. Use the bound locations below only
within separately supplied trusted scope. Current governing instructions take
precedence over these copied baseline instructions; surface conflicts.

`)
	bound := make([]artifact, 0, len(inventory.Artifacts))
	for _, item := range inventory.Artifacts {
		item.Path = filepath.Join(workspace, "inputs", filepath.FromSlash(item.Path))
		bound = append(bound, item)
	}
	if err := block(map[string]any{"sourceBaseline": "882b5002bd4dc70b4be0060e5e5668d1faa8bbdd", "sourceRootRecordedNotAuthenticated": sourceRoot,
		"sourceManifestSHA256": sourceManifestSHA256, "inputs": bound}); err != nil {
		return nil, err
	}
	copies := make([]artifact, 0, len(exercise.Artifacts))
	for _, item := range exercise.Artifacts {
		item.Path = filepath.Join(workspace, "exercise", filepath.FromSlash(item.Path))
		copies = append(copies, item)
	}
	if err := block(map[string]any{"exerciseManifestSHA256": exercisePin, "frozenExerciseFiles": copies,
		"evidenceFiles": []string{filepath.Join(workspace, "source-manifest.json"), filepath.Join(workspace, "source-report.json"), filepath.Join(workspace, "exercise-manifest.json"), filepath.Join(workspace, "inputs-report.json"), filepath.Join(workspace, "exercise-report.json")}}); err != nil {
		return nil, err
	}
	out.WriteString(`## Proposed assignment boundary, not permission

The route is agent-written Go for fictional FirstDesignDraft. Do not invoke a
generator or replace independent literals with observed output. This is the finite
lifecycle/v1 profile, not a whole product or standards adapter. No persistence,
authentication, input validator, external systems, dependencies or new files outside
the explicit frontier are included. Neither a candidate nor its executable exists
at preparation. Frozen module files are read-only; preserve all other work.

If edit/build authority is later supplied, its entire proposed write frontier is:

`)
	if err := block(map[string]any{"candidateFile": filepath.Join(workspace, "exercise", "candidate.go"), "buildProduct": filepath.Join(workspace, "exercise", "candidate.test"),
		"cacheAndTemporaryDirectories": []string{filepath.Join(workspace, "cache"), filepath.Join(workspace, "tmp"), filepath.Join(workspace, "modcache")}}); err != nil {
		return nil, err
	}
	out.WriteString(`## Proposed compile-only command, not run

This JSON object is command data, not shell text. Pass arguments and environment
literally after separate authorization; never evaluate paths as shell expressions.
The environment object replaces the inherited environment rather than overlaying
it. Do not carry unlisted Go settings or external cache/tool hooks into the build.
The named Go binary was not opened or executed. Required version is go1.27.3;
its actual identity/version and host suitability remain unverified. A gofmt binary
and formatting authority are not inferred from this path and remain coordinator
checks. The exercise working directory and output/cache/temp paths are explicit.

go test -c compiles without running package initializers. go test -run, even with
an empty selection, still executes initializers and is withheld. No candidate
execution, test execution, worker dispatch, integration, publication, merge or
deployment is granted. Candidate source must remain the pure standard-library
fixture implementation; inspect exact source and build authority before compiling.

`)
	if err := block(map[string]any{"program": goBinary, "args": []string{"test", "-c", "-o", filepath.Join(workspace, "exercise", "candidate.test")},
		"directory": filepath.Join(workspace, "exercise"), "environment": map[string]string{
			"GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOENV": "off", "GOFLAGS": "", "GOEXPERIMENT": "", "GOCACHEPROG": "", "CGO_ENABLED": "0",
			"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOCACHE": filepath.Join(workspace, "cache"), "GOTMPDIR": filepath.Join(workspace, "tmp"), "TMPDIR": filepath.Join(workspace, "tmp"), "GOMODCACHE": filepath.Join(workspace, "modcache"),
		}}); err != nil {
		return nil, err
	}
	out.WriteString(`## Required coordinator checks before dispatch

- Name the receiving builder, current live task/claim, target state and exact trusted
  read/write/build grants. Verify the historical claim separately against its pinned
  source. No live claim or historical Git identity was checked by this preparer.
- Check current governing instructions and supply any additional exact permitted
  instruction/claim reads. Missing inputs require a named resolution, never discovery.
- Independently review this complete bound draft against the original requirements,
  model and frozen literals. Input hashes do not supply semantic review or approval.
- Verify the original independent expectation provenance and freeze chronology;
  this preparer does not independently authenticate the historical account.
- Verify toolchain and formatter identities, native host and exact build boundaries.
- Keep Q_use and visual approval open. Any narrowly authorized fictional experiment
  must be explicit and must not resolve real-use decisions or promise product fitness.

## Builder return, only after a separate assignment

Return exact changed files and candidate hash; requirement/model/interface links;
original and final frozen-source hashes; actual input exposure and metadata reads;
commands, exits and compile-only artifact identity; discrepancies, unavailable
inputs, unrun checks and remaining questions with the next responsible owner.
Do not change frozen tests, hashes or unresolved decisions to make a check pass.
Candidate behavior remains not-run unless the coordinator later obtains separate
exact-source review and execution authority. Full expected/actual comparisons for
unrun tests must not be invented. Keep this private workspace and its reports private.
`)
	return out.Bytes(), nil
}
