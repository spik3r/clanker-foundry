package scenariotrace

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"strings"
	"sync"
	"testing"
)

func TestFrozenContract(t *testing.T) {
	runFrozenContract(t, Check, func(err error) string {
		switch {
		case errors.Is(err, ErrInvalid):
			return "invalid"
		case errors.Is(err, ErrBinding):
			return "binding"
		case errors.Is(err, ErrLimit):
			return "limit"
		default:
			return "boundary"
		}
	})
}

func testRequest(t *testing.T, reportName string) []byte {
	t.Helper()
	v := frozenLoad(t)
	q := map[string]any{"formatVersion": RequestVersion, "scenarioReport": nil}
	for k, s := range v.Sources {
		q[k] = s
	}
	if reportName != "" {
		q["scenarioReport"] = v.Reports[reportName]
	}
	return []byte(frozenJSON(t, q))
}

func TestConcurrentOwnership(t *testing.T) {
	raw := testRequest(t, "matched")
	before := append([]byte(nil), raw...)
	want, err := Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 8 {
				got, err := Check(raw)
				if err != nil || !bytes.Equal(got, want) {
					t.Error("concurrent result changed")
					return
				}
				// Each result is owned by its caller, not reused on subsequent calls.
				for i := range got {
					got[i] = 0
				}
			}
		})
	}
	wg.Wait()
	if !bytes.Equal(before, raw) {
		t.Fatal("input mutated")
	}
	got, err := Check(raw)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("result aliases a previous output")
	}
}

func TestReportProtocolRefusals(t *testing.T) {
	v := frozenLoad(t)
	for _, field := range []string{"mode", "authority", "modelDigest"} {
		t.Run(field, func(t *testing.T) {
			q := frozenObject(t, string(testRequest(t, "matched")))
			r := frozenObject(t, v.Reports["matched"])
			e := r["entries"].([]any)[0].(map[string]any)
			e["actual"].(map[string]any)[field] = "invented-secret"
			q["scenarioReport"] = frozenJSON(t, r)
			assertInvalid(t, []byte(frozenJSON(t, q)))
		})
	}
	q := frozenObject(t, string(testRequest(t, "matched")))
	r := frozenObject(t, v.Reports["matched"])
	r["formatVersion"] = scenarios.BoundReportVersion
	q["scenarioReport"] = frozenJSON(t, r)
	assertInvalid(t, []byte(frozenJSON(t, q)))
}

func TestNativeStructuralLimitsAreInvalid(t *testing.T) {
	q := frozenObject(t, string(testRequest(t, "")))
	s := frozenObject(t, q["suite"].(string))
	c := s["cases"].([]any)[0]
	cs := make([]any, scenarios.MaxCases+1)
	for i := range cs {
		cs[i] = c
	}
	s["cases"] = cs
	q["suite"] = frozenJSON(t, s)
	assertInvalid(t, []byte(frozenJSON(t, q)))
}

func TestStrictAdditionalRefusals(t *testing.T) {
	raw := testRequest(t, "")
	for _, b := range [][]byte{nil, []byte("null"), []byte("[]"), append(append([]byte(nil), raw...), 0xff), []byte(`{"formatVersion":"\udc00"}`)} {
		assertInvalid(t, b)
	}
	for _, value := range []any{nil, true, 1, []any{}, map[string]any{}, ""} {
		q := frozenObject(t, string(raw))
		q["model"] = value
		assertInvalid(t, []byte(frozenJSON(t, q)))
	}
}
func assertInvalid(t *testing.T, raw []byte) {
	t.Helper()
	out, err := Check(raw)
	if out != nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("got output length %d, error %v", len(out), err)
	}
}

func TestFirstStepMismatchRetainsAllSteps(t *testing.T) {
	v := frozenLoad(t)
	q := frozenObject(t, string(testRequest(t, "")))
	r := frozenObject(t, v.Reports["early"])
	first := r["entries"].([]any)[0].(map[string]any)
	first["actual"].(map[string]any)["refusals"] = []any{"OTHER"}
	first["matched"] = false
	first["differences"] = []any{"refusals"}
	r["entries"] = []any{first}
	r["stepsExecuted"] = 1
	q["scenarioReport"] = frozenJSON(t, r)
	raw, err := Check([]byte(frozenJSON(t, q)))
	if err != nil {
		t.Fatal(err)
	}
	out := frozenObject(t, string(raw))
	cs := out["cases"].([]any)
	if cs[0].(map[string]any)["status"] != "mismatched" || cs[1].(map[string]any)["status"] != "not-observed" {
		t.Fatal("statuses")
	}
	steps := cs[0].(map[string]any)["steps"].([]any)
	if len(steps) != 2 || steps[0].(map[string]any)["entryPosition"] != float64(0) || steps[1].(map[string]any)["entryPosition"] != nil {
		t.Fatal("steps")
	}
}

func TestEncodedOutputCap(t *testing.T) {
	// Include the terminating newline in the route's encoded-byte bound.
	r := report{}
	empty, err := encodeReport(r)
	if err != nil {
		t.Fatal(err)
	}
	r.Mode = strings.Repeat("x", MaxOutputBytes-len(empty))
	exact, err := encodeReport(r)
	if err != nil || len(exact) != MaxOutputBytes {
		t.Fatalf("exact output bound: %d %v", len(exact), err)
	}
	r.Mode += "x"
	over, err := encodeReport(r)
	if over != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("oversized output accepted")
	}
}

func FuzzCheckRefusalOwnership(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"formatVersion":"\ud800"}`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > MaxRequestBytes+1 {
			t.Skip()
		}
		before := append([]byte(nil), raw...)
		out, err := Check(raw)
		if !bytes.Equal(raw, before) {
			t.Fatal("mutated input")
		}
		if err != nil && out != nil {
			t.Fatal("partial output")
		}
		if err != nil && !errors.Is(err, ErrInvalid) && !errors.Is(err, ErrBinding) && !errors.Is(err, ErrLimit) && !errors.Is(err, ErrBoundary) {
			t.Fatal("unclassified error")
		}
		if err == nil && (len(out) > MaxOutputBytes || !json.Valid(out)) {
			t.Fatal("invalid output")
		}
	})
}
