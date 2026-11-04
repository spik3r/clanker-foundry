package revisioncheck

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
)

func TestV2MatchesSeparatelyFrozenContract(t *testing.T) { assertFrozenV2Contract(t, CheckV2) }

// Current accepted per-side bounds cannot construct a native comparison limit
// refusal. Exercise its explicit classification without inventing such an input.
func TestV2NativeRefusalAndBoundaryClassification(t *testing.T) {
	for _, tc := range []struct {
		err    error
		reason string
	}{
		{scenariodiff.ErrInvalid, "native-input-refused"},
		{scenariodiff.ErrLimit, "native-limit-refused"},
	} {
		result, err := scenarioDiffOutcome([]byte("discarded partial bytes"), tc.err)
		if err != nil || result.Status != "not-checked" || result.Reason == nil || *result.Reason != tc.reason || result.ReportJSON != nil || result.ReportSHA256 != nil {
			t.Fatal("refusal classification", result, err)
		}
	}
	for _, err := range []error{nil, errors.New("invented private failure")} {
		result, got := scenarioDiffOutcome(nil, err)
		if !errors.Is(got, ErrBoundary) || result != (scenarioDiffResult{}) {
			t.Fatal("boundary returned partial result", result, got)
		}
	}
}

func TestV2ConcurrentReadOnlyCallsAndExplicitVersion(t *testing.T) {
	raw := []byte(v2Observations(t)[0].Request)
	expected, err := CheckV2(raw)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := CheckV2(raw)
			if err != nil || !bytes.Equal(result, expected) {
				t.Errorf("concurrent result changed: %v", err)
			}
		}()
	}
	group.Wait()
	var req Request
	if json.Unmarshal(raw, &req) != nil {
		t.Fatal("fixture")
	}
	for _, version := range []string{RequestVersion, "lifecycle-revision-check-request/v3"} {
		req.FormatVersion = version
		wrong, _ := json.Marshal(req)
		if output, err := CheckV2(wrong); !errors.Is(err, ErrInvalid) || output != nil {
			t.Fatal("v2 accepted another version", version, err)
		}
	}
}
