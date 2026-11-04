package boundinventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestFrozenCustomInventory(t *testing.T) { assertFrozenInventory(t, Check) }
func TestStrictAndResourceBoundaries(t *testing.T) {
	v := frozenVectors(t)[0]
	raw := frozenWire(t, v)
	for _, field := range []string{"formatVersion", "model", "suite", "report"} {
		t.Run("missing-"+field, func(t *testing.T) {
			var q map[string]any
			_ = json.Unmarshal(raw, &q)
			delete(q, field)
			out, err := Check(frozenJSON(t, q))
			if out != nil || !errors.Is(err, ErrInvalid) {
				t.Fatal("missing field", err)
			}
		})
	}
	for _, test := range []struct {
		field string
		size  int
	}{{"model", maxModelBytes + 1}, {"suite", (1 << 20) + 1}, {"report", (8 << 20) + 1}} {
		t.Run(test.field+"-limit", func(t *testing.T) {
			var q map[string]any
			_ = json.Unmarshal(raw, &q)
			q[test.field] = strings.Repeat(" ", test.size)
			out, err := Check(frozenJSON(t, q))
			if out != nil || !errors.Is(err, ErrLimit) {
				t.Fatal("source bound", err)
			}
		})
	}
	for _, bad := range [][]byte{
		bytes.Replace(raw, []byte(`"model":`), []byte(`"unknown":0,"model":`), 1),
		bytes.Replace(raw, []byte(`"report":`), []byte(`"report":false,"report":`), 1),
		bytes.Replace(raw, []byte(`"suite":`), []byte(`"suite":null,"suite":`), 1),
		[]byte("{\"formatVersion\":\"\xff\"}"),
		[]byte(`{"formatVersion":"\ud800","model":"","suite":"","report":""}`),
	} {
		if out, err := Check(bad); out != nil || !errors.Is(err, ErrInvalid) {
			t.Fatal("strict JSON", err)
		}
	}
	if out, err := Check(bytes.Repeat([]byte(" "), MaxRequestBytes+1)); out != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("outer bound", err)
	}
}
func TestIdentityErrorPrecedence(t *testing.T) {
	base := frozenVectors(t)[0]
	for _, key := range []string{"sourceDigest", "scenarioDigest", "modelDigest"} {
		t.Run(key, func(t *testing.T) {
			v := base
			var p map[string]any
			_ = json.Unmarshal([]byte(v.Report), &p)
			p[key] = strings.Repeat("a", 64)
			v.Report = string(frozenJSON(t, p))
			want := ErrBinding
			if key == "modelDigest" {
				want = ErrInvalid
			}
			if out, err := Check(frozenWire(t, v)); out != nil || !errors.Is(err, want) {
				t.Fatal("native decode precedes explicit identity checks", err)
			}
		})
	}
}
func TestOriginalSourceIdentitiesAndConcurrentOwnership(t *testing.T) {
	base := frozenVectors(t)[0]
	before, err := Check(frozenWire(t, base))
	if err != nil {
		t.Fatal(err)
	}
	var original report
	if json.Unmarshal(before, &original) != nil {
		t.Fatal("report")
	}
	// Formatting the model changes its source identity, not native semantics.
	base.Model = "\r\n" + base.Model + "\n"
	raw := frozenWire(t, base)
	saved := bytes.Clone(raw)
	expected, err := Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	var formatted report
	if json.Unmarshal(expected, &formatted) != nil {
		t.Fatal("report")
	}
	if formatted.Inputs.ModelSHA256 == original.Inputs.ModelSHA256 || formatted.Inputs.ModelDigest != original.Inputs.ModelDigest || formatted.Inputs.SuiteSHA256 != original.Inputs.SuiteSHA256 {
		t.Fatal("raw and normalized identities conflated")
	}
	var group sync.WaitGroup
	for i := 0; i < 12; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			out, err := Check(raw)
			if err != nil || !bytes.Equal(out, expected) {
				t.Error("concurrent result", err)
			}
			if len(out) > 0 {
				out[0] = '!'
			}
		}()
	}
	group.Wait()
	if !bytes.Equal(raw, saved) {
		t.Fatal("input mutated")
	}
	out, err := Check(raw)
	if err != nil || !bytes.Equal(out, expected) {
		t.Fatal("output storage shared")
	}
}
func TestBufferedOutputLimit(t *testing.T) {
	// Exercise the delivery bound directly rather than invent a native-valid
	// scenario exceeding the native entry count or identifier limits.
	value := report{Status: strings.Repeat("x", MaxOutputBytes)}
	if out, err := encode(value); out != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("output bound", err)
	}
}
