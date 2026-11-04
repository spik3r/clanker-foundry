package operatorcli_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/changerehearsal"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

// Frozen before change-entrypoint production. Native package bytes and the
// independent ParcelSeal vectors remain unchanged; literal outcomes below are
// not learned from either CLI or refreshed from an execution result.
const changeDigest = "36b3cd00dafe2ecdf12c9d1c92e166f4e2f29fbc8c01beb58c89cb4a0cce6771"

type changeRequest struct {
	FormatVersion           string `json:"formatVersion"`
	PatchRequest            string `json:"patchRequest"`
	BaselineProposalRequest string `json:"baselineProposalRequest"`
	CandidateSuite          string `json:"candidateSuite"`
	CandidateManifest       string `json:"candidateManifest"`
	ExpectedCandidateDigest string `json:"expectedCandidateDigest"`
}

func changeHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func changeVectors(t *testing.T) []string {
	t.Helper()
	for path, digest := range map[string]string{
		"../../changerehearsal/testdata/observations.json": "d9b30467427539c02d3355724b0d9ac13c47f46360996b4dfcc660d1a96403e3",
		"../../changerehearsal/change.go":                  "7af39626cfb1ddce694e2baf37c0383bc9bdc89aedd494ddfafe356a2366e53f",
		"../../changerehearsal/execute.go":                 "598e3d53b740c2bd0c557fa14d5b2de8b96090a0b2edba285feea766a5add2bb",
	} {
		if changeHash(read(t, path)) != digest {
			t.Fatalf("frozen native source changed: %s", path)
		}
	}
	var corpus struct {
		Vectors []struct{ ID, Request string }
	}
	if err := json.Unmarshal(read(t, "../../changerehearsal/testdata/observations.json"), &corpus); err != nil {
		t.Fatal(err)
	}
	ids := []string{"changed-actor-guard", "baseline-mismatch-stops-candidate", "candidate-mismatch-retained", "stale-candidate-preflight", "wrong-expected-candidate", "baseline-not-patch-before"}
	if len(corpus.Vectors) != len(ids) {
		t.Fatal("frozen native vector count")
	}
	requests := make([]string, len(ids))
	for i, v := range corpus.Vectors {
		if v.ID != ids[i] {
			t.Fatal("frozen native vector order")
		}
		requests[i] = v.Request
	}
	return requests
}

func changeDecode(t *testing.T, raw []byte) changeRequest {
	t.Helper()
	var q changeRequest
	if err := json.Unmarshal(raw, &q); err != nil {
		t.Fatal(err)
	}
	return q
}

func changeFiles(t *testing.T, q changeRequest) ([]string, map[string]string) {
	t.Helper()
	dir := t.TempDir()
	paths := make(map[string]string)
	args := []string{"change", "prepare"}
	for _, entry := range []struct{ flag, text string }{
		{"patch-request", q.PatchRequest}, {"baseline-proposal", q.BaselineProposalRequest},
		{"candidate-suite", q.CandidateSuite}, {"candidate-manifest", q.CandidateManifest},
	} {
		path := filepath.Join(dir, "PRIVATE-"+entry.flag)
		if err := os.WriteFile(path, []byte(entry.text), 0600); err != nil {
			t.Fatal(err)
		}
		paths[entry.flag] = path
		args = append(args, "--"+entry.flag, path)
	}
	return append(args, "--expected-candidate-digest", q.ExpectedCandidateDigest), paths
}

func changeFixture(t *testing.T) changeRequest {
	t.Helper()
	return changeDecode(t, []byte(changeVectors(t)[0]))
}

func assertChangeRefusal(t *testing.T, args []string, diagnostic string) {
	t.Helper()
	code, out, diag := run(args, poisonReader{})
	if code != 2 || len(out) != 0 || diag != diagnostic {
		t.Fatalf("refusal: code=%d stdout=%d diagnostic=%q", code, len(out), diag)
	}
}

func TestChangePrepareLiteralEncodingAndExactFileBytes(t *testing.T) {
	q := changeRequest{"lifecycle-change-rehearsal-request/v1", "patch\r\n雪", "baseline\t\"quoted\"\n", "suite\\end", "", changeDigest}
	args, paths := changeFiles(t, q)
	before := make(map[string][]byte)
	for role, path := range paths {
		before[role] = read(t, path)
	}
	want := `{"formatVersion":"lifecycle-change-rehearsal-request/v1","patchRequest":"patch\r\n雪","baselineProposalRequest":"baseline\t\"quoted\"\n","candidateSuite":"suite\\end","candidateManifest":"","expectedCandidateDigest":"36b3cd00dafe2ecdf12c9d1c92e166f4e2f29fbc8c01beb58c89cb4a0cce6771"}` + "\n"
	code, out, diag := run(args, poisonReader{})
	if code != 0 || string(out) != want || diag != "" || changeDecode(t, out) != q {
		t.Fatalf("literal assembly: code=%d diagnostic=%q output=%q", code, diag, out)
	}
	// Even these non-JSON artifacts assemble. Preparation is not preflight.
	for role, path := range paths {
		if !bytes.Equal(read(t, path), before[role]) {
			t.Fatalf("preparation changed input %s", role)
		}
	}
	reversed := []string{"change", "prepare"}
	for i := len(args) - 2; i >= 2; i -= 2 {
		reversed = append(reversed, args[i:i+2]...)
	}
	code, reordered, diag := run(reversed, poisonReader{})
	if code != 0 || diag != "" || !bytes.Equal(reordered, out) {
		t.Fatal("flag order changed prepared bytes")
	}
	q.PatchRequest = "\ufeff\r\n{\"duplicate\":1,\"duplicate\":2}\r\n"
	q.BaselineProposalRequest = "<>&\u2028\u2029\x00\t\r\n"
	q.CandidateSuite = "雪 e\u0301 🧪\r\n"
	q.CandidateManifest = "not JSON and not a reviewed manifest\n"
	args, _ = changeFiles(t, q)
	code, out, diag = run(args, poisonReader{})
	if code != 0 || diag != "" || changeDecode(t, out) != q {
		t.Fatal("preparation normalized bytes or imposed semantic validation", code, diag)
	}
}

func TestChangePrepareKeepsNativeVectorsAndNeverRefreshesExpectations(t *testing.T) {
	for i, raw := range changeVectors(t) {
		t.Run([]string{"current", "baseline-mismatch", "candidate-mismatch", "stale", "wrong-digest", "wrong-baseline"}[i], func(t *testing.T) {
			q := changeDecode(t, []byte(raw))
			args, paths := changeFiles(t, q)
			before := make(map[string][]byte)
			for role, path := range paths {
				before[role] = read(t, path)
			}
			code, prepared, diag := run(args, poisonReader{})
			if code != 0 || diag != "" || changeDecode(t, prepared) != q {
				t.Fatal("native artifacts changed or refused during assembly", code, diag)
			}
			for role, path := range paths {
				if !bytes.Equal(read(t, path), before[role]) {
					t.Fatal("native source file changed", role)
				}
			}
			want, err := changerehearsal.Check(prepared)
			code, out, diag := run([]string{"change", "check"}, bytes.NewReader(prepared))
			wantCode := []int{0, 0, 0, 1, 2, 2}[i]
			wantDiag := ""
			if i >= 4 {
				wantDiag = "modelchangerehearsal: request rejected\n"
			}
			if code != wantCode || diag != wantDiag || !bytes.Equal(out, want) || changerehearsal.ExitCode(want, err) != wantCode {
				t.Fatalf("native check parity: code=%d diagnostic=%q", code, diag)
			}
			if i < 4 {
				assertChangeReport(t, prepared, out, "check", []string{"checked", "checked", "checked", "stale"}[i], "check", "check")
			}
		})
	}
}

func TestChangePrepareRejectsMalformedFlagsAndDigestBeforeFileReads(t *testing.T) {
	args, _ := changeFiles(t, changeFixture(t))
	bad := [][]string{{"change", "prepare"}, args[:len(args)-1], append(append([]string{}, args...), "positional"), append(append([]string{}, args...), "--patch-request", "PRIVATE-missing")}
	for _, index := range []int{2, 4, 6, 8, 10} {
		missing := append([]string{}, args[:index]...)
		bad = append(bad, append(missing, args[index+2:]...))
	}
	for _, replacement := range []string{"--unknown", "--patch-request", "--candidate-suite=FILE"} {
		copy := append([]string{}, args...)
		copy[4] = replacement
		bad = append(bad, copy)
	}
	for _, digest := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.ToUpper(changeDigest), "g" + changeDigest[1:], changeDigest + "\n", " " + changeDigest[1:], strings.Repeat("雪", 64), "PRIVATE-digest-file"} {
		copy := append([]string{}, args...)
		for i := 3; i < 10; i += 2 {
			copy[i] = "PRIVATE-missing-file"
		}
		copy[11] = digest
		bad = append(bad, copy)
	}
	for _, malformed := range bad {
		assertChangeRefusal(t, malformed, "foundry change prepare: invalid arguments\n")
	}
	// Lowercase syntax is sufficient for preparation; it must not derive or
	// replace the operator's separately supplied candidate assertion.
	args[11] = strings.Repeat("0", 64)
	code, out, diag := run(args, poisonReader{})
	if code != 0 || diag != "" || changeDecode(t, out).ExpectedCandidateDigest != strings.Repeat("0", 64) {
		t.Fatal("syntactically valid digest was replaced or semantically checked")
	}
}

func TestChangePrepareRejectsUnsafeLeavesAndKeepsParentAliases(t *testing.T) {
	q := changeFixture(t)
	for _, role := range []string{"patch-request", "baseline-proposal", "candidate-suite", "candidate-manifest"} {
		for _, kind := range []string{"missing", "symlink", "directory", "invalid-utf8"} {
			t.Run(role+"/"+kind, func(t *testing.T) {
				args, paths := changeFiles(t, q)
				path := paths[role]
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				var err error
				switch kind {
				case "symlink":
					err = os.Symlink(paths["candidate-suite"], path)
				case "directory":
					err = os.Mkdir(path, 0700)
				case "invalid-utf8":
					err = os.WriteFile(path, []byte{0xf0, 0x80, 0x80, 0x80}, 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
				assertChangeRefusal(t, args, "foundry change prepare: input rejected\n")
			})
		}
	}
	args, paths := changeFiles(t, q)
	alias := filepath.Join(t.TempDir(), "parent-alias")
	if err := os.Symlink(filepath.Dir(paths["patch-request"]), alias); err != nil {
		t.Fatal(err)
	}
	for i := 3; i < 10; i += 2 {
		args[i] = filepath.Join(alias, filepath.Base(args[i]))
	}
	code, out, diag := run(args, poisonReader{})
	if code != 0 || diag != "" || changeDecode(t, out) != q {
		t.Fatal("trusted-host parent alias refused or source bytes changed", code, diag)
	}
}

func TestChangePrepareInclusiveFileAndEncodedWrapperBounds(t *testing.T) {
	for role, limit := range map[string]int{"patch-request": 3 << 20, "baseline-proposal": 16 << 20, "candidate-suite": 1 << 20, "candidate-manifest": 64 << 10} {
		t.Run(role, func(t *testing.T) {
			q := changeRequest{FormatVersion: "lifecycle-change-rehearsal-request/v1", ExpectedCandidateDigest: changeDigest}
			args, paths := changeFiles(t, q)
			for _, extra := range []int{0, 1} {
				if err := os.WriteFile(paths[role], bytes.Repeat([]byte("x"), limit+extra), 0600); err != nil {
					t.Fatal(err)
				}
				code, out, diag := run(args, poisonReader{})
				if extra == 0 {
					if code != 0 || diag != "" {
						t.Fatalf("inclusive file limit refused: code=%d diagnostic=%q", code, diag)
					}
					var fields map[string]string
					if json.Unmarshal(out, &fields) != nil {
						t.Fatal("prepared request")
					}
					key := map[string]string{"patch-request": "patchRequest", "baseline-proposal": "baselineProposalRequest", "candidate-suite": "candidateSuite", "candidate-manifest": "candidateManifest"}[role]
					if fields[key] != strings.Repeat("x", limit) {
						t.Fatal("source truncated at limit")
					}
				} else if code != 2 || len(out) != 0 || diag != "foundry change prepare: input rejected\n" {
					t.Fatal("source limit exceeded", code, diag)
				}
			}
		})
	}
	// Independent wire assembly makes the wrapper exactly 32 MiB. Escaping one
	// source, not the sum of raw file sizes, exhausts the outer request budget.
	q := changeRequest{FormatVersion: "lifecycle-change-rehearsal-request/v1", ExpectedCandidateDigest: changeDigest}
	empty, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	budget := (32 << 20) - len(empty) - 1
	q.BaselineProposalRequest = strings.Repeat("\x01", budget/6) + strings.Repeat("x", budget%6)
	for _, extra := range []int{0, 1} {
		fixture := q
		fixture.BaselineProposalRequest += strings.Repeat("x", extra)
		wire, err := json.Marshal(fixture)
		if err != nil || len(wire)+1 != (32<<20)+extra || len(fixture.BaselineProposalRequest) > 16<<20 {
			t.Fatal("incorrect independent wrapper-bound fixture")
		}
		args, _ := changeFiles(t, fixture)
		code, out, diag := run(args, poisonReader{})
		if extra == 0 {
			if code != 0 || diag != "" || len(out) != 32<<20 || changeDecode(t, out) != fixture {
				t.Fatal("inclusive encoded limit", code, len(out), diag)
			}
		} else if code != 2 || len(out) != 0 || diag != "foundry change prepare: request limit exceeded\n" {
			t.Fatal("expanded wrapper admitted", code, len(out), diag)
		}
	}
}

type changeErrorWriter struct{}

func (changeErrorWriter) Write([]byte) (int, error) { return 0, errors.New("PRIVATE writer failure") }

func TestChangeDeliveryFailureAndStrictNativeInput(t *testing.T) {
	raw := []byte(changeVectors(t)[0])
	prepare, _ := changeFiles(t, changeDecode(t, raw))
	for _, writer := range []io.Writer{shortWriter{}, changeErrorWriter{}} {
		for _, v := range []struct {
			args []string
			in   io.Reader
			diag string
		}{
			{prepare, poisonReader{}, "foundry change prepare: output failed\n"},
			{[]string{"change", "check"}, bytes.NewReader(raw), "modelchangerehearsal: report delivery failed; execution may have started\n"},
		} {
			var diag bytes.Buffer
			if code := operatorcli.Run(context.Background(), v.args, v.in, writer, &diag); code != 3 || diag.String() != v.diag {
				t.Fatal("delivery classification", code, diag.String())
			}
		}
	}
	for _, action := range []string{"check", "execute"} {
		for _, input := range []io.Reader{
			strings.NewReader("PRIVATE malformed"), strings.NewReader("null"), strings.NewReader("{}"),
			bytes.NewReader(append(bytes.Clone(raw), raw...)),
			bytes.NewReader(bytes.Replace(raw, []byte(`"patchRequest":`), []byte(`"sourceRoot":"PRIVATE-root","patchRequest":`), 1)),
			bytes.NewReader(bytes.Replace(raw, []byte(`"patchRequest":`), []byte(`"patchRequest":"duplicate","patchRequest":`), 1)),
			bytes.NewReader([]byte{0xff}), bytes.NewReader(bytes.Repeat([]byte(" "), (32<<20)+1)),
		} {
			code, out, diag := run([]string{"change", action}, input)
			if code != 2 || len(out) != 0 || diag != "modelchangerehearsal: request rejected\n" {
				t.Fatal("strict native refusal", code, diag)
			}
		}
		code, out, diag := run([]string{"change", action}, failedReader{})
		if code != 2 || len(out) != 0 || diag != "modelchangerehearsal: input read failed\n" {
			t.Fatal("native read failure", code, diag)
		}
	}
}

func assertChangeReport(t *testing.T, input, raw []byte, action, outcome, baseline, candidate string) changerehearsal.Report {
	t.Helper()
	var r changerehearsal.Report
	if json.Unmarshal(raw, &r) != nil || r.FormatVersion != "lifecycle-change-rehearsal-report/v1" || r.Mode != "example-only" || r.Authority != "none" || r.ReviewStatus != "not-established" || r.Action != action || r.Outcome != outcome || r.Baseline.Operation != baseline || r.Candidate.Operation != candidate {
		t.Fatal("literal change status or authority", r.Action, r.Outcome, r.Baseline.Operation, r.Candidate.Operation)
	}
	q := changeDecode(t, input)
	if r.RequestSHA256 != changeHash(input) || r.PatchRequestSHA256 != changeHash([]byte(q.PatchRequest)) || r.ExpectedCandidateDigest != q.ExpectedCandidateDigest || r.PatchReportSHA256 != changeHash([]byte(r.PatchReportJSON)) || r.Baseline.ProposalRequestSHA256 != changeHash([]byte(q.BaselineProposalRequest)) {
		t.Fatal("source/report hash binding changed")
	}
	for _, side := range []changerehearsal.Side{r.Baseline, r.Candidate} {
		if side.ReceiptSHA256 != changeHash([]byte(side.ReceiptJSON)) {
			t.Fatal("native receipt bytes changed")
		}
		var receipt proposalreplay.Receipt
		if json.Unmarshal([]byte(side.ReceiptJSON), &receipt) != nil || receipt.Mode != "example-only" || receipt.Authority != "none" || receipt.Action != side.Operation {
			t.Fatal("native receipt authority changed")
		}
		if side.Operation == "check" && (receipt.CandidateStarted || receipt.Compilation != "not-requested" || receipt.Replay != "not-requested") {
			t.Fatal("check claimed execution")
		}
	}
	return r
}

func TestChangeCancellationAndHostBoundaryKeepNativePartialReports(t *testing.T) {
	raw := []byte(changeVectors(t)[0])
	for _, kind := range []string{"cancelled", "untrusted-cwd"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if kind == "cancelled" {
				cancel()
			} else {
				t.Chdir(t.TempDir())
			}
			want, err := changerehearsal.Execute(ctx, raw, ".")
			var out, diag bytes.Buffer
			code := operatorcli.Run(ctx, []string{"change", "execute"}, bytes.NewReader(raw), &out, &diag)
			if code != 4 || changerehearsal.ExitCode(want, err) != 4 || !bytes.Equal(out.Bytes(), want) || diag.String() != "modelchangerehearsal: execution boundary failed; inspect receipt\n" {
				t.Fatal("partial receipt parity", code, diag.String())
			}
			operation := "check"
			if kind == "untrusted-cwd" {
				operation = "execute"
			}
			r := assertChangeReport(t, raw, out.Bytes(), "execute", "failed", operation, "check")
			var baseline proposalreplay.Receipt
			if json.Unmarshal([]byte(r.Baseline.ReceiptJSON), &baseline) != nil || baseline.CandidateStarted {
				t.Fatal("suppressed baseline fabricated candidate start")
			}
			if kind == "untrusted-cwd" && (baseline.FailureStage == nil || *baseline.FailureStage != "sources" || baseline.Outcome != "failed") {
				t.Fatal("request overrode trusted host source root")
			}
		})
	}
}

func changeBinaries(t *testing.T) (string, map[string]string) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bins := make(map[string]string)
	for _, name := range []string{"foundry", "modelchangerehearsal"} {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		filename := name
		if runtime.GOOS == "windows" {
			filename += ".exe"
		}
		bin := filepath.Join(t.TempDir(), filename)
		cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", bin, "./cmd/"+name)
		cmd.Dir, cmd.WaitDelay = root, time.Second
		output, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v %s", name, err, output)
		}
		bins[name] = bin
	}
	return root, bins
}

func assertChangeLiteralExecution(t *testing.T, r changerehearsal.Report) {
	t.Helper()
	for i, side := range []changerehearsal.Side{r.Baseline, r.Candidate} {
		if side.Operation != "execute" {
			continue
		}
		var receipt proposalreplay.Receipt
		if json.Unmarshal([]byte(side.ReceiptJSON), &receipt) != nil || !receipt.CandidateStarted || receipt.Compilation != "passed" || receipt.Cleanup != "removed" || receipt.FailureStage != nil || receipt.CandidateExitCode == nil {
			t.Fatal("actual invented candidate execution missing")
		}
		var replay struct {
			Matched                       bool
			CasesCompleted, StepsExecuted int
			Entries                       []struct {
				CaseID, StepID string
				Actual         struct {
					OK           bool
					TransitionID string
					Task         struct {
						ID, State, Revision, Outcome string
						Owner                        *string
						Acceptance                   []string
					}
					Refusals []string
				}
			}
		}
		if json.Unmarshal(receipt.CandidateReport, &replay) != nil || len(replay.Entries) == 0 {
			t.Fatal("actual replay report missing")
		}
		mismatch := (i == 0 && r.Outcome == "baseline-mismatched") || (i == 1 && r.Outcome == "candidate-mismatched")
		if replay.Matched == mismatch || *receipt.CandidateExitCode != map[bool]int{false: 0, true: 1}[mismatch] || (!mismatch && (replay.CasesCompleted != 2 || replay.StepsExecuted != 2)) {
			t.Fatal("literal match status or scenario count changed")
		}
		first := replay.Entries[0]
		actual := first.Actual
		if first.CaseID != "actor_Mira" || first.StepID != "seal_Mira" || actual.TransitionID != "T_seal" || actual.Task.ID != "parcel-Mira" || actual.Task.Outcome != "Seal invented parcel" || strings.Join(actual.Task.Acceptance, "|") != "keep order|second" {
			t.Fatal("literal ParcelSeal identity/order changed")
		}
		if i == 0 {
			if !actual.OK || actual.Task.State != "Sealed" || actual.Task.Revision != "8" || actual.Task.Owner == nil || *actual.Task.Owner != "Mira" || len(actual.Refusals) != 0 {
				t.Fatal("baseline Mira outcome changed")
			}
		} else if actual.OK || actual.Task.State != "Packed" || actual.Task.Revision != "7" || actual.Task.Owner != nil || len(actual.Refusals) != 1 || actual.Refusals[0] != "G_actor" {
			t.Fatal("candidate Mira refusal changed")
		}
	}
}

func TestChangeFrozenNativeCheckAndExecutionLiterals(t *testing.T) {
	requests := changeVectors(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for i, input := range requests {
		raw := []byte(input)
		before := bytes.Clone(raw)
		report, err := changerehearsal.Check(raw)
		if code := changerehearsal.ExitCode(report, err); code != []int{0, 0, 0, 1, 2, 2}[i] || !bytes.Equal(raw, before) {
			t.Fatal("frozen native check status or input mutation", i, code)
		}
		if i >= 4 {
			if len(report) != 0 {
				t.Fatal("native identity refusal returned report")
			}
			continue
		}
		assertChangeReport(t, raw, report, "check", []string{"checked", "checked", "checked", "stale"}[i], "check", "check")
		if i == 3 {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		report, err = changerehearsal.Execute(ctx, raw, root)
		cancel()
		if code := changerehearsal.ExitCode(report, err); code != []int{0, 1, 1}[i] || !bytes.Equal(raw, before) {
			t.Fatal("frozen native execution status or input mutation", i, code, err)
		}
		candidate := "execute"
		if i == 1 {
			candidate = "check"
		}
		r := assertChangeReport(t, raw, report, "execute", []string{"matched", "baseline-mismatched", "candidate-mismatched"}[i], "execute", candidate)
		assertChangeLiteralExecution(t, r)
	}
}

func TestChangeCompiledNativeParityLiteralExecutionAndClosedPipes(t *testing.T) {
	requests := changeVectors(t)
	root, bins := changeBinaries(t)
	for _, v := range []struct {
		action, outcome, baseline, candidate string
		index, code                          int
	}{
		{"check", "checked", "check", "check", 0, 0},
		{"check", "stale", "check", "check", 3, 1},
		{"check", "", "", "", 4, 2},
		{"execute", "matched", "execute", "execute", 0, 0},
		{"execute", "baseline-mismatched", "execute", "check", 1, 1},
		{"execute", "candidate-mismatched", "execute", "execute", 2, 1},
		{"execute", "stale", "check", "check", 3, 1},
	} {
		t.Run(v.action+"/"+v.outcome, func(t *testing.T) {
			input := []byte(requests[v.index])
			oldCode, oldOut, oldDiag := binaryResult(t, root, bins["modelchangerehearsal"], []string{v.action}, input, false)
			code, out, diag := binaryResult(t, root, bins["foundry"], []string{"change", v.action}, input, false)
			wantDiag := ""
			if v.code == 2 {
				wantDiag = "modelchangerehearsal: request rejected\n"
			}
			if code != v.code || oldCode != v.code || diag != wantDiag || oldDiag != wantDiag || !bytes.Equal(oldOut, out) {
				t.Fatalf("compiled native parity: old=%d new=%d diagnostic=%q/%q", oldCode, code, oldDiag, diag)
			}
			if v.code == 2 {
				if len(out) != 0 {
					t.Fatal("initial refusal delivered a report")
				}
				return
			}
			r := assertChangeReport(t, input, out, v.action, v.outcome, v.baseline, v.candidate)
			if v.action == "execute" && v.outcome != "stale" {
				assertChangeLiteralExecution(t, r)
			}
		})
	}
	for _, action := range []string{"check", "execute"} {
		for _, name := range []string{"foundry", "modelchangerehearsal"} {
			args := []string{action}
			if name == "foundry" {
				args = append([]string{"change"}, args...)
			}
			code, out, diag := binaryResult(t, root, bins[name], args, []byte(requests[0]), true)
			if code != 3 || len(out) != 0 || diag != "modelchangerehearsal: report delivery failed; execution may have started\n" {
				t.Fatal("compiled closed stdout", name, action, code, diag)
			}
		}
	}
	args, paths := changeFiles(t, changeDecode(t, []byte(requests[0])))
	code, out, diag := binaryResult(t, root, bins["foundry"], args, nil, true)
	if code != 3 || len(out) != 0 || diag != "foundry change prepare: output failed\n" {
		t.Fatal("compiled prepare closed stdout", code, diag)
	}
	// The bounded child detects a regression that opens a FIFO before refusing
	// it. No writer is present, so accepting this leaf would otherwise block.
	if err := os.Remove(paths["patch-request"]); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		return // Windows has no POSIX FIFO leaf.
	}
	mkfifo, err := exec.LookPath("mkfifo")
	if err != nil {
		t.Log("FIFO leaf check not run: mkfifo is unavailable")
		return
	}
	if b, err := exec.Command(mkfifo, paths["patch-request"]).CombinedOutput(); err != nil {
		t.Fatalf("make FIFO: %v %s", err, b)
	}
	code, out, diag = binaryResult(t, root, bins["foundry"], args, nil, false)
	if code != 2 || len(out) != 0 || diag != "foundry change prepare: input rejected\n" {
		t.Fatal("compiled FIFO admission", code, diag)
	}
}
