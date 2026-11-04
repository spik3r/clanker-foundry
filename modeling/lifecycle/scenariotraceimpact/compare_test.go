package scenariotraceimpact_test

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotrace"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotraceimpact"
)

func TestConcurrentReadOnlyCalls(t *testing.T) {
	fixtures := acceptanceFixtures(t)
	tc := fixtures[len(fixtures)-3] // Added case and link, with multiple candidates.
	raw := acceptanceOuter(t, tc.Before, tc.After)
	original := bytes.Clone(raw)
	want, err := scenariotraceimpact.Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	failures := make(chan string, 12)
	for range 12 {
		wg.Go(func() {
			got, err := scenariotraceimpact.Check(raw)
			if err != nil || !bytes.Equal(got, want) {
				failures <- fmt.Sprintf("concurrent report differs: %v", err)
				return
			}
			// Other concurrent calls must not share returned mutable storage.
			for i := range got {
				got[i] = '!'
			}
		})
	}
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
	if !bytes.Equal(raw, original) {
		t.Fatal("mutated shared input")
	}
}

func TestValidSnapshotsCanExceedComparisonAdmission(t *testing.T) {
	v := acceptanceLoad(t)
	for _, kind := range []string{"catalog_cases", "trace_report", "comparison_request"} {
		t.Run(kind, func(t *testing.T) {
			q := acceptanceSnapshot(t, v, "")
			want := scenariotraceimpact.ErrLimit
			switch kind {
			case "catalog_cases":
				acceptanceEditSource(t, q, "catalog", func(c map[string]any) {
					cases := c["sequences"].([]any)
					for i := range 511 { // Two original cases plus 511 exceed 512.
						cases = append(cases, map[string]any{"caseId": fmt.Sprintf("extra-%03d", i), "name": "Declared only", "executionStatus": "pending"})
					}
					c["sequences"] = cases
				})
				want = scenariotraceimpact.ErrInvalid
			case "trace_report":
				// Valid whitespace expands when retained inside the native JSON.
				q["scenarioReport"] = strings.Repeat("\t", 5<<20) + v.Reports["matched"]
			case "comparison_request":
				catalog := q["catalog"].(string)
				acceptanceSetSource(t, q, "catalog", strings.Repeat("\t", (4<<20)-len(catalog))+catalog)
			}
			snapshot := acceptanceJSON(t, q)
			native, err := scenariotrace.Check([]byte(snapshot))
			if err != nil {
				t.Fatalf("native snapshot must remain valid: %v", err)
			}
			if kind == "trace_report" && len(native) <= scenariotraceimpact.MaxTraceReportBytes {
				t.Fatal("fixture did not cross the retained trace cap")
			}
			if kind == "comparison_request" && len(acceptanceTraceRequest(t, snapshot, snapshot)) <= 16<<20 {
				t.Fatal("fixture did not cross the native comparison request cap")
			}
			acceptanceRefuse(t, acceptanceOuter(t, snapshot, snapshot), want)
		})
	}
}
