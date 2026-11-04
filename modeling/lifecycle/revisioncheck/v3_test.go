package revisioncheck

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
)

func TestV3MatchesSeparatelyFrozenContract(t *testing.T) { assertFrozenV3Contract(t, CheckV3) }

// Accepted v2 inputs already contain valid bounded models. Exercise defensive
// native-error mapping directly without claiming an unreachable input vector.
func TestV3NativeModelErrorsDiscardPartialBytes(t *testing.T) {
	for _, tc := range []struct{ native, want error }{
		{impact.ErrInvalidInput, ErrInvalid},
		{fmt.Errorf("wrapped: %w", impact.ErrInvalidInput), ErrInvalid},
		{impact.ErrReportLimit, ErrLimit},
		{fmt.Errorf("wrapped: %w", impact.ErrReportLimit), ErrLimit},
		{errors.New("private native failure"), ErrBoundary},
	} {
		got, err := modelDiffOutcome([]byte("partial report"), tc.native)
		if !errors.Is(err, tc.want) || got != (modelDiffResult{}) {
			t.Fatal("native failure retained partial bytes", got, err)
		}
	}
	if got, err := modelDiffOutcome(nil, nil); !errors.Is(err, ErrBoundary) || got != (modelDiffResult{}) {
		t.Fatal("empty native success", got, err)
	}
}

func TestV3ModelCompositionBoundsAndExtraction(t *testing.T) {
	// These private builder controls precede no public acceptance claim. The
	// public API first requires the stricter complete v2 checks.
	req := Request{}
	if got, err := modelDiffRequest(req); !errors.Is(err, ErrBoundary) || got != nil {
		t.Fatal("malformed extraction", err)
	}
	proposal, err := json.Marshal(struct{ Model, Suite string }{Model: string(bytes.Repeat([]byte(" "), impact.MaxSourceBytes)), Suite: "{}"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := json.Marshal(firstdesignreview.Request{ProposalRequest: string(proposal)})
	if err != nil {
		t.Fatal(err)
	}
	req.Before.FirstDesignRequest, req.After.FirstDesignRequest = string(first), string(first)
	if got, err := modelDiffRequest(req); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal("combined model request bound", err)
	}
	if got, err := encode(reportV3{}, 1); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal("final encoding limit returned partial bytes", err)
	}
}
