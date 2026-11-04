package custommergerehearsal

import (
	"bytes"
	"encoding/json"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"os"
	"strings"
	"testing"
)

func TestFrozenArtifactIdentities(t *testing.T) {
	for path, want := range map[string]string{
		"contract_test.go":           "0cfc28d393be36df85eecb6873b36647e83fac9f7c8f248eea6005704a5a434e",
		"testdata/observations.json": "0c1c99a345875a292d7add9e9b032d246266c01a62555f68ea84f7f7b467e93b",
	} {
		b, e := os.ReadFile(path)
		if e != nil || hash(b) != want {
			t.Fatalf("frozen artifact changed: %s", path)
		}
	}
}
func TestNestedStrictFramingAndLimits(t *testing.T) {
	f := contractLoad(t)
	checkBad := func(r contractRequest) {
		t.Helper()
		b, e := Check(contractJSON(t, r))
		if e == nil || b != nil || ExitCode(b, e) != 2 {
			t.Fatal("invalid nested input accepted", e)
		}
	}
	for role := 0; role < 4; role++ {
		for _, kind := range []string{"empty-suite", "empty-manifest", "large-suite", "large-manifest"} {
			r := contractReq(f)
			p := contractPairs(&r)[role]
			switch kind {
			case "empty-suite":
				p.Suite = ""
			case "empty-manifest":
				p.Manifest = ""
			case "large-suite":
				p.Suite = strings.Repeat(" ", scenarios.MaxSourceBytes+1)
			case "large-manifest":
				p.Manifest = strings.Repeat(" ", acceptanceauthor.MaxManifestBytes+1)
			}
			checkBad(r)
		}
	}
	for _, merge := range []string{"null", "{}", f.MergeRequest + " {}", strings.Repeat(" ", modelmerge.MaxRequestBytes+1), strings.Replace(f.MergeRequest, `"base":`, `"unknown":null,"base":`, 1)} {
		r := contractReq(f)
		r.MergeRequest = merge
		checkBad(r)
	}
	for _, digest := range []string{"", strings.Repeat("A", 64), strings.Repeat("0", 63), strings.Repeat("0", 65)} {
		r := contractReq(f)
		r.ExpectedCandidateDigest = digest
		checkBad(r)
	}
	valid := contractJSON(t, contractReq(f))
	for _, bad := range [][]byte{
		bytes.Replace(valid, []byte(`"suite":`), []byte(`"SUITE":`), 1),
		bytes.Replace(valid, []byte(`"suite":`), []byte(`"extra":true,"suite":`), 1),
		bytes.Replace(valid, []byte(`"suite":`), []byte(`"suite":null,"suite":`), 1),
	} {
		b, e := Check(bad)
		if e == nil || b != nil || ExitCode(b, e) != 2 {
			t.Fatal("nested strict framing")
		}
	}
}
func TestSemanticCheckNoFabricatedReceipts(t *testing.T) {
	f := contractLoad(t)
	for _, m := range []string{f.ConflictMergeRequest, f.InvalidCombinationMergeRequest} {
		r := contractReq(f)
		r.MergeRequest = m
		out, e := Check(contractJSON(t, r))
		if e != nil || ExitCode(out, e) != 1 {
			t.Fatal(e)
		}
		got := contractRead(t, out)
		if got.Action != "check" {
			t.Fatal("wrong action")
		}
		for _, s := range contractStages(got) {
			if s.Operation != "not-attempted" || s.ReceiptJSON != nil {
				t.Fatal("invented check")
			}
		}
	}
}
func TestReportEncodingAndExitErrors(t *testing.T) {
	if b, e := encode(map[string]string{"x": "x"}, 1); b != nil || e != ErrLimit {
		t.Fatal("report limit")
	}
	if b, e := encode(make(chan int), 100); b != nil || e != ErrBoundary {
		t.Fatal("encoding boundary")
	}
	for _, b := range [][]byte{nil, []byte("null"), []byte("{}"), []byte(`{"formatVersion":"wrong","outcome":"matched"}`)} {
		if ExitCode(b, nil) != 4 {
			t.Fatal("invalid report status")
		}
	}
	for _, outcome := range []string{"failed", "invented"} {
		b, _ := json.Marshal(Report{FormatVersion: ReportVersion, Outcome: outcome})
		if ExitCode(b, nil) != 4 {
			t.Fatal("unknown outcome")
		}
	}
}
