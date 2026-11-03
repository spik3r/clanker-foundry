package firstdesignreview

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func firstValid(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Request  string `json:"request"`
			Expected struct {
				Valid bool `json:"valid"`
			} `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &f); err != nil || len(f.Cases) == 0 || !f.Cases[0].Expected.Valid {
		t.Fatalf("frozen fixture: %v", err)
	}
	return []byte(f.Cases[0].Request)
}

func TestBoundsAndNoPartialReport(t *testing.T) {
	if out, err := Check(bytes.Repeat([]byte("x"), MaxRequestBytes+1)); !errors.Is(err, ErrLimit) || out != nil {
		t.Fatalf("outer bound: output=%d err=%v", len(out), err)
	}
	var req Request
	if err := json.Unmarshal(firstValid(t), &req); err != nil {
		t.Fatal(err)
	}
	// Native intake's own bound is narrower than the wrapper's.
	req.IntakeRequest = strings.Repeat(" ", 4<<20+1)
	raw, err := json.Marshal(req)
	if err != nil || len(raw) > MaxRequestBytes {
		t.Fatalf("test setup: %d %v", len(raw), err)
	}
	if out, err := Check(raw); !errors.Is(err, ErrLimit) || out != nil {
		t.Fatalf("native bound: output=%d err=%v", len(out), err)
	}
	// Encoding escapes each NUL as six bytes, exceeding the final output cap.
	if out, err := encode(report{ProposalReportJSON: strings.Repeat("\x00", MaxReportBytes/6)}); !errors.Is(err, ErrLimit) || out != nil {
		t.Fatalf("output bound: output=%d err=%v", len(out), err)
	}
}

func TestWrapperEscapingDoesNotChangeDecodedComparison(t *testing.T) {
	raw := firstValid(t)
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	// Equivalent outer JSON escaping is not an artifact byte difference.
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(req); err != nil {
		t.Fatal(err)
	}
	out, err := Check(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		ModelComparison string `json:"modelComparison"`
		SuiteComparison string `json:"suiteComparison"`
	}
	if err := json.Unmarshal(out, &got); err != nil || got.ModelComparison != "matches" || got.SuiteComparison != "matches" {
		t.Fatalf("decoded comparison: %+v %v", got, err)
	}
}
