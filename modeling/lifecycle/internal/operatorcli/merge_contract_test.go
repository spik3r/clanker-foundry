package operatorcli_test

// Frozen before operator merge implementation. The ArchivePacket observations
// are independent, unchanged native fixtures. These finite host tests do not
// establish safety against concurrent writers, rollback, or human approval.
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/custommergerehearsal"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

const mergeDigest = "476a955520c2265074fd359c2420f2a2c1a2bde1cd5c88a8119fbca63aab8eea"

var mergeRoles = []string{"base", "left", "right", "candidate"}
var mergeFileRoles = []string{"merge-request", "base-suite", "base-manifest", "left-suite", "left-manifest", "right-suite", "right-manifest", "candidate-suite", "candidate-manifest"}

type mergePair struct {
	Suite    string `json:"suite"`
	Manifest string `json:"manifest"`
}
type mergeRequest struct {
	FormatVersion           string    `json:"formatVersion"`
	MergeRequest            string    `json:"mergeRequest"`
	Base                    mergePair `json:"base"`
	Left                    mergePair `json:"left"`
	Right                   mergePair `json:"right"`
	Candidate               mergePair `json:"candidate"`
	ExpectedCandidateDigest string    `json:"expectedCandidateDigest"`
}
type mergeVersion struct{ Model, ModelDigest, Suite, Manifest, MismatchSuite, MismatchManifest string }
type mergeCorpus struct {
	Versions                                                                                map[string]mergeVersion
	MergeRequest, ReverseMergeRequest, ConflictMergeRequest, InvalidCombinationMergeRequest string
}

func mergeLoad(t *testing.T) mergeCorpus {
	t.Helper()
	for path, digest := range map[string]string{
		"../../custommergerehearsal/merge.go":                   "c04369c53948ef4fe80804aa7c0d87bbe10d9e4e43b62cda4bc39f072c49cc83",
		"../../custommergerehearsal/execute.go":                 "4a48b45352d633785c5422683e291cfed1cb8f9b91ca0ca7a98f80ef33e9540d",
		"../../custommergerehearsal/testdata/observations.json": "0c1c99a345875a292d7add9e9b032d246266c01a62555f68ea84f7f7b467e93b",
	} {
		if changeHash(read(t, path)) != digest {
			t.Fatalf("frozen native source changed: %s", path)
		}
	}
	var f mergeCorpus
	if err := json.Unmarshal(read(t, "../../custommergerehearsal/testdata/observations.json"), &f); err != nil {
		t.Fatal(err)
	}
	if f.Versions["candidate"].ModelDigest != mergeDigest {
		t.Fatal("independent candidate assertion changed")
	}
	return f
}
func mergeFixture(f mergeCorpus) mergeRequest {
	pair := func(role string) mergePair { v := f.Versions[role]; return mergePair{v.Suite, v.Manifest} }
	return mergeRequest{"lifecycle-custom-merge-rehearsal-request/v1", f.MergeRequest, pair("base"), pair("left"), pair("right"), pair("candidate"), mergeDigest}
}
func mergePairs(q *mergeRequest) []*mergePair {
	return []*mergePair{&q.Base, &q.Left, &q.Right, &q.Candidate}
}
func mergeTexts(q *mergeRequest) []*string {
	return []*string{&q.MergeRequest, &q.Base.Suite, &q.Base.Manifest, &q.Left.Suite, &q.Left.Manifest, &q.Right.Suite, &q.Right.Manifest, &q.Candidate.Suite, &q.Candidate.Manifest}
}
func mergeEncode(t *testing.T, q mergeRequest) []byte {
	t.Helper()
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(q); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func mergeDecode(t *testing.T, b []byte) mergeRequest {
	t.Helper()
	var q mergeRequest
	if err := json.Unmarshal(b, &q); err != nil {
		t.Fatal(err)
	}
	return q
}
func mergeFiles(t *testing.T, q mergeRequest) ([]string, map[string]string) {
	t.Helper()
	dir := t.TempDir()
	paths := map[string]string{}
	args := []string{"merge", "prepare"}
	for i, text := range mergeTexts(&q) {
		role := mergeFileRoles[i]
		p := filepath.Join(dir, "PRIVATE-"+role)
		if err := os.WriteFile(p, []byte(*text), 0600); err != nil {
			t.Fatal(err)
		}
		paths[role] = p
		args = append(args, "--"+role, p)
	}
	return append(args, "--expected-candidate-digest", q.ExpectedCandidateDigest), paths
}
func mergeRefusal(t *testing.T, args []string, diagnostic string) {
	t.Helper()
	code, out, diag := run(args, poisonReader{})
	if code != 2 || len(out) != 0 || diag != diagnostic {
		t.Fatalf("refusal code=%d bytes=%d diag=%q", code, len(out), diag)
	}
}

func TestMergePrepareLiteralBytesAndIndependentStages(t *testing.T) {
	q := mergeRequest{"lifecycle-custom-merge-rehearsal-request/v1", "merge\r\n雪", mergePair{"base\t\"x\"", "bm"}, mergePair{"left\\end", "lm"}, mergePair{"right", "rm"}, mergePair{"candidate", ""}, mergeDigest}
	want := `{"formatVersion":"lifecycle-custom-merge-rehearsal-request/v1","mergeRequest":"merge\r\n雪","base":{"suite":"base\t\"x\"","manifest":"bm"},"left":{"suite":"left\\end","manifest":"lm"},"right":{"suite":"right","manifest":"rm"},"candidate":{"suite":"candidate","manifest":""},"expectedCandidateDigest":"476a955520c2265074fd359c2420f2a2c1a2bde1cd5c88a8119fbca63aab8eea"}` + "\n"
	args, paths := mergeFiles(t, q)
	code, out, diag := run(args, poisonReader{})
	if code != 0 || diag != "" || string(out) != want || mergeDecode(t, out) != q {
		t.Fatalf("literal preparation %d %q %q", code, diag, out)
	}
	for i, p := range mergeTexts(&q) {
		if string(read(t, paths[mergeFileRoles[i]])) != *p {
			t.Fatal("input mutation")
		}
	}
	// All ten flag pairs are moved through every starting position, not only reversed.
	for start := 0; start < 10; start++ {
		permuted := []string{"merge", "prepare"}
		for n := 0; n < 10; n++ {
			i := 2 + 2*((start+n)%10)
			permuted = append(permuted, args[i:i+2]...)
		}
		code, got, diag := run(permuted, poisonReader{})
		if code != 0 || diag != "" || !bytes.Equal(got, out) {
			t.Fatal("flag order changed bytes")
		}
	}
	for i, p := range mergeTexts(&q) {
		*p = fmt.Sprintf("role=%d\ufeff\r\n<>&\u2028\u2029\x00雪 e\u0301 🧪\n{\"duplicate\":1,\"duplicate\":2}", i)
	}
	args, _ = mergeFiles(t, q)
	code, out, diag = run(args, poisonReader{})
	if code != 0 || diag != "" || mergeDecode(t, out) != q || !bytes.Equal(out, mergeEncode(t, q)) {
		t.Fatal("byte normalization or semantic validation", code, diag)
	}
	q.ExpectedCandidateDigest = strings.Repeat("0", 64)
	args, _ = mergeFiles(t, q)
	code, out, diag = run(args, poisonReader{})
	if code != 0 || diag != "" || mergeDecode(t, out).ExpectedCandidateDigest != q.ExpectedCandidateDigest {
		t.Fatal("candidate identity was inferred")
	}
}

func TestMergePrepareAllMalformedFlagPositionsBeforeFiles(t *testing.T) {
	args, _ := mergeFiles(t, mergeFixture(mergeLoad(t)))
	for i := 3; i < 20; i += 2 {
		args[i] = "PRIVATE-missing"
	}
	bad := [][]string{{"merge", "prepare"}, args[:len(args)-1], append(append([]string{}, args...), "positional")}
	for i := 2; i < len(args); i += 2 {
		absent := append([]string{}, args[:i]...)
		bad = append(bad, append(absent, args[i+2:]...))
		for _, flag := range []string{"--unknown", args[i] + "=PRIVATE", "positional", args[2+2*((i/2)%10)]} {
			v := append([]string{}, args...)
			v[i] = flag
			bad = append(bad, v)
		}
		repeated := append([]string{}, args...)
		bad = append(bad, append(repeated, args[i:i+2]...))
	}
	for _, digest := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.ToUpper(mergeDigest), "g" + mergeDigest[1:], mergeDigest + "\n", " " + mergeDigest[1:], strings.Repeat("雪", 64), "PRIVATE-digest-file"} {
		v := append([]string{}, args...)
		v[len(v)-1] = digest
		bad = append(bad, v)
	}
	for _, v := range bad {
		mergeRefusal(t, v, "foundry merge prepare: invalid arguments\n")
	}
}
func TestMergePrepareAllUnsafeLeavesAndParentAliases(t *testing.T) {
	q := mergeFixture(mergeLoad(t))
	for _, role := range mergeFileRoles {
		for _, kind := range []string{"missing", "symlink", "dangling-symlink", "directory", "invalid-utf8"} {
			t.Run(role+"/"+kind, func(t *testing.T) {
				args, paths := mergeFiles(t, q)
				p := paths[role]
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
				var err error
				switch kind {
				case "symlink":
					target := filepath.Join(filepath.Dir(p), "PRIVATE-target")
					if err := os.WriteFile(target, []byte("valid target"), 0600); err != nil {
						t.Fatal(err)
					}
					err = os.Symlink(target, p)
				case "dangling-symlink":
					err = os.Symlink(filepath.Join(filepath.Dir(p), "PRIVATE-missing-target"), p)
				case "directory":
					err = os.Mkdir(p, 0700)
				case "invalid-utf8":
					err = os.WriteFile(p, []byte{0xf0, 0x80, 0x80, 0x80}, 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
				mergeRefusal(t, args, "foundry merge prepare: input rejected\n")
			})
		}
	}
	args, paths := mergeFiles(t, q)
	alias := filepath.Join(t.TempDir(), "parent-alias")
	if err := os.Symlink(filepath.Dir(paths["merge-request"]), alias); err != nil {
		t.Fatal(err)
	}
	for i := 3; i < 20; i += 2 {
		args[i] = filepath.Join(alias, filepath.Base(args[i]))
	}
	code, out, diag := run(args, poisonReader{})
	if code != 0 || diag != "" || mergeDecode(t, out) != q {
		t.Fatal("parent alias or source bytes changed", code, diag)
	}
}
func TestMergePrepareInclusiveFileAndExpandedWrapperBounds(t *testing.T) {
	empty := mergeRequest{FormatVersion: "lifecycle-custom-merge-rehearsal-request/v1", ExpectedCandidateDigest: mergeDigest}
	for i, role := range mergeFileRoles {
		t.Run(role, func(t *testing.T) {
			limit := 1 << 20
			if i == 0 {
				limit = 4 << 20
			} else if strings.HasSuffix(role, "manifest") {
				limit = 64 << 10
			}
			for _, extra := range []int{0, 1} {
				q := empty
				*mergeTexts(&q)[i] = strings.Repeat("x", limit+extra)
				args, _ := mergeFiles(t, q)
				code, out, diag := run(args, poisonReader{})
				if extra == 0 {
					if code != 0 || diag != "" || mergeDecode(t, out) != q {
						t.Fatal("inclusive limit or exact bytes", code, diag)
					}
				} else if code != 2 || len(out) != 0 || diag != "foundry merge prepare: input rejected\n" {
					t.Fatal("oversized file", code, diag)
				}
			}
		})
	}
	// Each file remains within its own bound. JSON escaping, across two stages,
	// reaches the expanded wrapper limit exactly, including the trailing newline.
	q := empty
	q.MergeRequest = strings.Repeat("\x01", 4<<20)
	q.Base.Suite = strings.Repeat("\x01", 1<<20)
	budget := (32 << 20) - len(mergeEncode(t, q))
	q.Left.Suite = strings.Repeat("\x01", budget/6) + strings.Repeat("x", budget%6)
	for _, extra := range []int{0, 1} {
		v := q
		v.Left.Suite += strings.Repeat("x", extra)
		wire := mergeEncode(t, v)
		if len(wire) != (32<<20)+extra || len(v.Left.Suite) > 1<<20 {
			t.Fatal("independent bound fixture")
		}
		args, _ := mergeFiles(t, v)
		code, out, diag := run(args, poisonReader{})
		if extra == 0 {
			if code != 0 || diag != "" || !bytes.Equal(out, wire) {
				t.Fatal("inclusive wrapper limit", code, len(out), diag)
			}
		} else if code != 2 || len(out) != 0 || diag != "foundry merge prepare: request limit exceeded\n" {
			t.Fatal("expanded wrapper admitted", code, diag)
		}
	}
}

func mergeStages(r custommergerehearsal.Report) []custommergerehearsal.Stage {
	return []custommergerehearsal.Stage{r.Base, r.Left, r.Right, r.Candidate}
}
func mergeAssertReport(t *testing.T, input, out []byte, action, outcome string, executed int) custommergerehearsal.Report {
	t.Helper()
	var r custommergerehearsal.Report
	if json.Unmarshal(out, &r) != nil {
		t.Fatal("report decoding")
	}
	q := mergeDecode(t, input)
	if r.FormatVersion != "lifecycle-custom-merge-rehearsal-report/v1" || r.Mode != "example-only" || r.Authority != "none" || r.ReviewStatus != "not-established" || r.Action != action || r.Outcome != outcome {
		t.Fatal("literal report status", r.Action, r.Outcome)
	}
	if r.RequestSHA256 != changeHash(input) || r.MergeRequestSHA256 != changeHash([]byte(q.MergeRequest)) || r.MergeRequestJSON != q.MergeRequest || r.ExpectedCandidateDigest != q.ExpectedCandidateDigest || r.MergeReportSHA256 != changeHash([]byte(r.MergeReportJSON)) {
		t.Fatal("envelope byte identity")
	}
	native, err := modelmerge.Execute([]byte(q.MergeRequest))
	if err != nil || !bytes.Equal(native, []byte(r.MergeReportJSON)) {
		t.Fatal("native merge bytes changed", err)
	}
	noCandidate := outcome == "conflicted" || outcome == "invalid-combination"
	if noCandidate {
		var result struct {
			Candidate       json.RawMessage
			CandidateDigest *string
		}
		if json.Unmarshal(native, &result) != nil || string(result.Candidate) != "null" || result.CandidateDigest != nil || r.Candidate.ModelSHA256 != nil {
			t.Fatal("invented candidate")
		}
	}
	for i, s := range mergeStages(r) {
		pair := mergePairs(&q)[i]
		if s.SuiteSHA256 != changeHash([]byte(pair.Suite)) || s.ManifestSHA256 != changeHash([]byte(pair.Manifest)) {
			t.Fatal("stage expectation pair changed", i)
		}
		op := "check"
		if i < executed {
			op = "execute"
		}
		if noCandidate {
			op = "not-attempted"
		}
		if s.Operation != op {
			t.Fatal("stage order/suppression", i, s.Operation, op)
		}
		if noCandidate {
			if s.ReceiptJSON != nil || s.ReceiptSHA256 != nil || s.ProposalRequestSHA256 != nil {
				t.Fatal("fabricated preflight")
			}
			continue
		}
		if s.ReceiptJSON == nil || s.ReceiptSHA256 == nil || *s.ReceiptSHA256 != changeHash([]byte(*s.ReceiptJSON)) || s.ProposalRequestSHA256 == nil {
			t.Fatal("receipt retention")
		}
		var receipt proposalreplay.Receipt
		if json.Unmarshal([]byte(*s.ReceiptJSON), &receipt) != nil || receipt.Action != op || receipt.Mode != "example-only" || receipt.Authority != "none" || receipt.Sources.SuiteSHA256 != s.SuiteSHA256 || receipt.Sources.ManifestSHA256 != s.ManifestSHA256 {
			t.Fatal("native receipt binding")
		}
		if op == "check" && (receipt.CandidateStarted || receipt.Compilation != "not-requested" || receipt.Replay != "not-requested") {
			t.Fatal("check claimed execution")
		}
	}
	return r
}

// Observed state values below are literal independent ArchivePacket policy
// expectations, not a comparison between two wrappers that may share a defect.
func mergeAssertExecution(t *testing.T, r custommergerehearsal.Report) {
	t.Helper()
	for stage, s := range mergeStages(r) {
		if s.Operation != "execute" {
			continue
		}
		var receipt proposalreplay.Receipt
		if s.ReceiptJSON == nil || json.Unmarshal([]byte(*s.ReceiptJSON), &receipt) != nil || !receipt.CandidateStarted || receipt.Compilation != "passed" || receipt.Cleanup != "removed" || receipt.FailureStage != nil || receipt.CandidateExitCode == nil {
			t.Fatal("missing actual execution")
		}
		mismatch := r.Outcome == mergeRoles[stage]+"-mismatched"
		wantCode := 0
		if mismatch {
			wantCode = 1
		}
		if *receipt.CandidateExitCode != wantCode {
			t.Fatal("literal candidate exit")
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
		if json.Unmarshal(receipt.CandidateReport, &replay) != nil || replay.Matched == mismatch || len(replay.Entries) == 0 {
			t.Fatal("literal replay match")
		}
		if !mismatch && (replay.CasesCompleted != 4 || replay.StepsExecuted != 4 || len(replay.Entries) != 4) {
			t.Fatal("literal case counts")
		}
		for i, e := range replay.Entries {
			a := e.Actual
			if e.CaseID != fmt.Sprintf("case_%d", i) || e.StepID != fmt.Sprintf("step_%d", i) || a.TransitionID != "T_file" || a.Task.ID != fmt.Sprintf("packet-%d", i) || strings.Join(a.Task.Acceptance, "|") != "preserve first|preserve second" {
				t.Fatal("literal ArchivePacket identity/order")
			}
			refs := []string{}
			if (stage == 1 || stage == 3) && i%2 == 1 {
				refs = append(refs, "G_actor")
			}
			if (stage == 2 || stage == 3) && i >= 2 {
				refs = append(refs, "G_outcome")
			}
			ok := len(refs) == 0
			state, rev := "Ready", "9007199254740993"
			if ok {
				state, rev = "Filed", "9007199254740994"
			}
			outcome := "indexed"
			if i >= 2 {
				outcome = "draft"
			}
			owner := ""
			if i == 2 {
				owner = "Existing"
			} else if ok {
				owner = "Clerk"
				if i%2 == 1 {
					owner = "Guest"
				}
			}
			actualOwner := ""
			if a.Task.Owner != nil {
				actualOwner = *a.Task.Owner
			}
			if a.OK != ok || a.Task.State != state || a.Task.Revision != rev || a.Task.Outcome != outcome || actualOwner != owner || strings.Join(a.Refusals, "|") != strings.Join(refs, "|") {
				t.Fatalf("literal policy stage=%d case=%d", stage, i)
			}
		}
	}
}

type mergeCase struct {
	name                         string
	q                            mergeRequest
	checkCode, executeCode       int
	checkOutcome, executeOutcome string
	executed                     int
}

func mergeCases(t *testing.T) []mergeCase {
	f := mergeLoad(t)
	q := mergeFixture(f)
	cases := []mergeCase{{"matched", q, 0, 0, "checked", "matched", 4}}
	for i, role := range mergeRoles {
		v := q
		native := f.Versions[role]
		*mergePairs(&v)[i] = mergePair{native.MismatchSuite, native.MismatchManifest}
		cases = append(cases, mergeCase{role + "-mismatched", v, 0, 1, "checked", role + "-mismatched", i + 1})
		v = q
		mergePairs(&v)[i].Suite += " "
		cases = append(cases, mergeCase{role + "-stale", v, 1, 1, "stale", "stale", 0})
	}
	for _, v := range []struct{ name, text string }{{"conflicted", f.ConflictMergeRequest}, {"invalid-combination", f.InvalidCombinationMergeRequest}} {
		r := q
		r.MergeRequest = v.text
		cases = append(cases, mergeCase{v.name, r, 1, 1, v.name, v.name, 0})
	}
	q.ExpectedCandidateDigest = strings.Repeat("0", 64)
	cases = append(cases, mergeCase{"identity-refused", q, 2, 2, "", "", 0})
	return cases
}
func TestMergeNativePreparedChecksAndLiteralExecution(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range mergeCases(t) {
		t.Run(v.name, func(t *testing.T) {
			args, paths := mergeFiles(t, v.q)
			code, raw, diag := run(args, poisonReader{})
			if code != 0 || diag != "" || mergeDecode(t, raw) != v.q {
				t.Fatal("assembly changed frozen pair", code, diag)
			}
			before := bytes.Clone(raw)
			for _, action := range []string{"check", "execute"} {
				var report []byte
				var err error
				expected, outcome, executed := v.checkCode, v.checkOutcome, 0
				if action == "check" {
					report, err = custommergerehearsal.Check(raw)
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
					report, err = custommergerehearsal.Execute(ctx, raw, root)
					cancel()
					expected, outcome, executed = v.executeCode, v.executeOutcome, v.executed
				}
				if custommergerehearsal.ExitCode(report, err) != expected || !bytes.Equal(raw, before) {
					t.Fatal("native status/input mutation", action, err)
				}
				if expected == 2 {
					if len(report) != 0 {
						t.Fatal("identity refusal report")
					}
					continue
				}
				r := mergeAssertReport(t, raw, report, action, outcome, executed)
				if action == "execute" && executed > 0 {
					mergeAssertExecution(t, r)
				}
				if action == "check" {
					code, out, diag := run([]string{"merge", "check"}, bytes.NewReader(raw))
					if code != expected || diag != "" || !bytes.Equal(out, report) {
						t.Fatal("native check parity", code, diag)
					}
				}
			}
			for i, p := range mergeTexts(&v.q) {
				if string(read(t, paths[mergeFileRoles[i]])) != *p {
					t.Fatal("input file changed")
				}
			}
		})
	}
}

func TestMergeStrictNativeInputDeliveryAndCancellation(t *testing.T) {
	q := mergeFixture(mergeLoad(t))
	raw := mergeEncode(t, q)
	args, _ := mergeFiles(t, q)
	for _, writer := range []io.Writer{shortWriter{}, changeErrorWriter{}} {
		for _, v := range []struct {
			args       []string
			input      io.Reader
			diagnostic string
		}{{args, poisonReader{}, "foundry merge prepare: output failed\n"}, {[]string{"merge", "check"}, bytes.NewReader(raw), "modelmergerehearsal: report delivery failed; execution may have started\n"}} {
			var diag bytes.Buffer
			if code := operatorcli.Run(context.Background(), v.args, v.input, writer, &diag); code != 3 || diag.String() != v.diagnostic {
				t.Fatal("delivery failure", code, diag.String())
			}
		}
	}
	for _, action := range []string{"check", "execute"} {
		for _, input := range [][]byte{[]byte("PRIVATE malformed"), []byte("null"), []byte("{}"), append(bytes.Clone(raw), raw...), bytes.Replace(raw, []byte(`"mergeRequest":`), []byte(`"sourceRoot":"PRIVATE-root","mergeRequest":`), 1), bytes.Replace(raw, []byte(`"mergeRequest":`), []byte(`"mergeRequest":"duplicate","mergeRequest":`), 1), {0xff}, bytes.Repeat([]byte(" "), (32<<20)+1)} {
			code, out, diag := run([]string{"merge", action}, bytes.NewReader(input))
			if code != 2 || len(out) != 0 || diag != "modelmergerehearsal: request rejected\n" {
				t.Fatal("strict native refusal", code, diag)
			}
		}
		code, out, diag := run([]string{"merge", action}, failedReader{})
		if code != 2 || len(out) != 0 || diag != "modelmergerehearsal: input read failed\n" {
			t.Fatal("native read failure", code, diag)
		}
	}
	for _, kind := range []string{"cancelled", "untrusted-cwd"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			executed := 0
			if kind == "cancelled" {
				cancel()
			} else {
				t.Chdir(t.TempDir())
				executed = 1
			}
			want, err := custommergerehearsal.Execute(ctx, raw, ".")
			var out, diag bytes.Buffer
			code := operatorcli.Run(ctx, []string{"merge", "execute"}, bytes.NewReader(raw), &out, &diag)
			if code != 4 || custommergerehearsal.ExitCode(want, err) != 4 || !bytes.Equal(out.Bytes(), want) || diag.String() != "modelmergerehearsal: execution boundary failed; inspect receipt\n" {
				t.Fatal("partial boundary report", code, diag.String())
			}
			r := mergeAssertReport(t, raw, out.Bytes(), "execute", "failed", executed)
			for i, s := range mergeStages(r) {
				var receipt proposalreplay.Receipt
				if json.Unmarshal([]byte(*s.ReceiptJSON), &receipt) != nil || receipt.CandidateStarted {
					t.Fatal("fabricated candidate start")
				}
				if kind == "untrusted-cwd" && i == 0 && (receipt.FailureStage == nil || *receipt.FailureStage != "sources" || receipt.Outcome != "failed") {
					t.Fatal("untrusted source root")
				}
			}
			var failedDiag bytes.Buffer
			if code := operatorcli.Run(ctx, []string{"merge", "execute"}, bytes.NewReader(raw), shortWriter{}, &failedDiag); code != 3 || failedDiag.String() != "modelmergerehearsal: report delivery failed; execution may have started\n" {
				t.Fatal("partial receipt delivery failure", code, failedDiag.String())
			}
		})
	}
}

func mergeBinaries(t *testing.T) (string, map[string]string) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bins := map[string]string{}
	for _, name := range []string{"foundry", "modelmergerehearsal"} {
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
func TestMergeCompiledNativeParityLiteralsAndClosedStdout(t *testing.T) {
	cases := mergeCases(t)
	root, bins := mergeBinaries(t)
	for _, v := range cases {
		for _, action := range []string{"check", "execute"} {
			t.Run(action+"/"+v.name, func(t *testing.T) {
				raw := mergeEncode(t, v.q)
				expected, outcome, executed := v.checkCode, v.checkOutcome, 0
				if action == "execute" {
					expected, outcome, executed = v.executeCode, v.executeOutcome, v.executed
				}
				a, old, ad := binaryResult(t, root, bins["modelmergerehearsal"], []string{action}, raw, false)
				b, out, bd := binaryResult(t, root, bins["foundry"], []string{"merge", action}, raw, false)
				diagnostic := ""
				if expected == 2 {
					diagnostic = "modelmergerehearsal: request rejected\n"
				}
				if a != expected || b != expected || ad != diagnostic || bd != diagnostic || !bytes.Equal(old, out) {
					t.Fatalf("compiled parity old=%d new=%d diag=%q/%q", a, b, ad, bd)
				}
				if expected == 2 {
					if len(out) != 0 {
						t.Fatal("identity refusal output")
					}
					return
				}
				r := mergeAssertReport(t, raw, out, action, outcome, executed)
				if executed > 0 {
					mergeAssertExecution(t, r)
				}
			})
		}
	}
	raw := mergeEncode(t, cases[0].q)
	for _, name := range []string{"foundry", "modelmergerehearsal"} {
		for _, action := range []string{"check", "execute"} {
			args := []string{action}
			if name == "foundry" {
				args = append([]string{"merge"}, args...)
			}
			code, out, diag := binaryResult(t, root, bins[name], args, raw, true)
			if code != 3 || len(out) != 0 || diag != "modelmergerehearsal: report delivery failed; execution may have started\n" {
				t.Fatal("closed stdout", name, action, code, diag)
			}
		}
		args := []string{"execute", "extra"}
		if name == "foundry" {
			args = append([]string{"merge"}, args...)
		}
		code, out, diag := binaryResult(t, root, bins[name], args, nil, false)
		if code != 2 || len(out) != 0 || diag != "modelmergerehearsal: usage: modelmergerehearsal check|execute < request.json\n" {
			t.Fatal("compiled invalid execute", code, diag)
		}
	}
	args, paths := mergeFiles(t, cases[0].q)
	code, out, diag := binaryResult(t, root, bins["foundry"], args, nil, true)
	if code != 3 || len(out) != 0 || diag != "foundry merge prepare: output failed\n" {
		t.Fatal("prepare closed stdout", code, diag)
	}
	if runtime.GOOS == "windows" {
		return
	}
	mkfifo, err := exec.LookPath("mkfifo")
	if err != nil {
		t.Log("FIFO check unavailable: mkfifo not installed")
		return
	}
	// Every file position rejects an unopened FIFO in a bounded subprocess.
	for _, role := range mergeFileRoles {
		original := read(t, paths[role])
		if err := os.Remove(paths[role]); err != nil {
			t.Fatal(err)
		}
		if b, err := exec.Command(mkfifo, paths[role]).CombinedOutput(); err != nil {
			t.Fatalf("mkfifo: %v %s", err, b)
		}
		code, out, diag := binaryResult(t, root, bins["foundry"], args, nil, false)
		if code != 2 || len(out) != 0 || diag != "foundry merge prepare: input rejected\n" {
			t.Fatal("FIFO admission", role, code, diag)
		}
		if err := os.Remove(paths[role]); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(paths[role], original, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// This control can run before any merge route exists. It checks the immutable
// native package directly, with the separately declared request and literals.
func TestMergeFrozenNativeControls(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range mergeCases(t) {
		t.Run(v.name, func(t *testing.T) {
			raw := mergeEncode(t, v.q)
			original := bytes.Clone(raw)
			for _, action := range []string{"check", "execute"} {
				var out []byte
				var err error
				code, outcome, executed := v.checkCode, v.checkOutcome, 0
				if action == "check" {
					out, err = custommergerehearsal.Check(raw)
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
					out, err = custommergerehearsal.Execute(ctx, raw, root)
					cancel()
					code, outcome, executed = v.executeCode, v.executeOutcome, v.executed
				}
				if custommergerehearsal.ExitCode(out, err) != code || !bytes.Equal(raw, original) {
					t.Fatal("frozen native status or mutation", action, err)
				}
				if code == 2 {
					if len(out) != 0 {
						t.Fatal("identity refusal output")
					}
					continue
				}
				r := mergeAssertReport(t, raw, out, action, outcome, executed)
				if executed > 0 {
					mergeAssertExecution(t, r)
				}
			}
		})
	}
}

func TestMergeSharedRunnerPreservesNativeBody(t *testing.T) {
	for _, v := range mergeCases(t) {
		raw := mergeEncode(t, v.q)
		want, err := custommergerehearsal.Check(raw)
		var out, diag bytes.Buffer
		code := operatorcli.RunMerge(context.Background(), []string{"check"}, bytes.NewReader(raw), &out, &diag)
		diagnostic := ""
		if v.checkCode == 2 {
			diagnostic = "modelmergerehearsal: request rejected\n"
		}
		if code != v.checkCode || custommergerehearsal.ExitCode(want, err) != code || !bytes.Equal(out.Bytes(), want) || diag.String() != diagnostic {
			t.Fatal("shared runner check", v.name, code, diag.String())
		}
	}
	for _, args := range [][]string{nil, {"help"}, {"prepare"}, {"execute", "extra"}, {"check", "--source-root", "PRIVATE"}} {
		var out, diag bytes.Buffer
		if code := operatorcli.RunMerge(context.Background(), args, poisonReader{}, &out, &diag); code != 2 || out.Len() != 0 || diag.String() != "modelmergerehearsal: usage: modelmergerehearsal check|execute < request.json\n" {
			t.Fatal("shared runner argument refusal", code, diag.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	raw := mergeEncode(t, mergeFixture(mergeLoad(t)))
	want, err := custommergerehearsal.Execute(ctx, raw, ".")
	var out, diag bytes.Buffer
	code := operatorcli.RunMerge(ctx, []string{"execute"}, bytes.NewReader(raw), &out, &diag)
	if code != 4 || custommergerehearsal.ExitCode(want, err) != 4 || !bytes.Equal(out.Bytes(), want) || diag.String() != "modelmergerehearsal: execution boundary failed; inspect receipt\n" {
		t.Fatal("shared runner cancelled receipt", code, diag.String())
	}
}
