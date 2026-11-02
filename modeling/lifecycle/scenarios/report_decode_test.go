package scenarios

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDecodeReportRoundTrip(t *testing.T) {
	raw, e := os.ReadFile("../scenarioexplain/testdata/vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Cases []struct {
			ID     string          `json:"id"`
			Report json.RawMessage `json:"report"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("vectors")
	}
	for _, id := range []string{"fabricated_multiple", "unknown_expected_literal_escape", "actual_null_refusals", "actual_empty_refusals", "source_digest_null"} {
		t.Run(id, func(t *testing.T) {
			var b []byte
			for _, c := range v.Cases {
				if c.ID == id {
					b = c.Report
				}
			}
			if b == nil {
				t.Fatal("missing vector")
			}
			original := append([]byte(nil), b...)
			r, e := DecodeReport(b)
			if e != nil {
				t.Fatal(e)
			}
			encoded, e := EncodeReport(r)
			if e != nil {
				t.Fatal(e)
			}
			var a, c any
			if json.Unmarshal(b, &a) != nil || json.Unmarshal(encoded, &c) != nil || !reflect.DeepEqual(a, c) {
				t.Fatal("round trip changed JSON semantics")
			}
			again, e := DecodeReport(encoded)
			if e != nil || !reflect.DeepEqual(r, again) {
				t.Fatal("typed round trip changed", e)
			}
			if id == "actual_null_refusals" && r.Entries[0].Actual.Refusals != nil {
				t.Fatal("null lost")
			}
			if id == "actual_empty_refusals" && (r.Entries[0].Actual.Refusals == nil || len(r.Entries[0].Actual.Refusals) != 0) {
				t.Fatal("empty lost")
			}
			if !bytes.Equal(b, original) {
				t.Fatal("input changed")
			}
		})
	}
}
