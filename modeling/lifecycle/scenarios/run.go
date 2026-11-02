package scenarios

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

type transition func(lifecycle.Task, lifecycle.Command) lifecycle.Result

// Run validates all input before evaluating any step. Literal expectations are
// never computed, repaired or refreshed from observed reducer results.
func Run(s Suite) (Report, error) { return runWith(s, lifecycle.Transition) }
func runWith(s Suite, apply transition) (Report, error) {
	return runFor(s, apply, defaultBinding(), false)
}

func runFor(s Suite, apply transition, binding Binding, provider bool) (Report, error) {
	if e := validateFor(s, binding); e != nil {
		return Report{}, e
	}
	digest, e := digestFor(s, binding)
	if e != nil {
		return Report{}, e
	}
	if provider {
		// An effectful callback may retain the caller's original Suite in a
		// closure. Freeze all validated literals before the first invocation.
		s = copySuite(s)
	}
	report := Report{ModelDigest: s.ModelDigest, ScenarioDigest: digest, Matched: true, Entries: []Entry{}}
	encodedSize := 4096 // Conservative fixed report-envelope allowance.
	for _, c := range s.Cases {
		current := copyTask(c.Initial)
		for _, step := range c.Steps {
			entry := Entry{CaseID: c.ID, StepID: step.ID, Input: copyTask(current), Command: copyCommand(step.Command), Expected: copyExpected(step.Expected)}
			observed := apply(copyTask(current), copyCommand(step.Command))
			if provider && !providerResultValid(observed, binding) {
				return Report{}, ErrProviderProtocol
			}
			entry.Actual = copyResult(observed)
			entry.Differences = compare(entry.Actual, entry.Expected, s.ModelDigest)
			entry.Matched = len(entry.Differences) == 0
			// Account for each complete encoded entry before retaining it. Large valid
			// input reused across many steps cannot accumulate an unbounded report.
			if !boundedEntryFor(entry, binding) {
				return Report{}, ErrReportLimit
			}
			raw, e := encodeJSON(entryToDTO(entry))
			if e != nil {
				return Report{}, ErrReportLimit
			}
			encodedSize += len(raw) + 1
			if encodedSize > MaxReportBytes {
				return Report{}, ErrReportLimit
			}
			report.Entries = append(report.Entries, entry)
			report.StepsExecuted++
			if !entry.Matched {
				report.Matched = false
				return report, nil
			}
			if observed.OK {
				current = copyTask(*observed.Task)
			}
			// Any matching refusal retains current, including INVALID_INPUT with Task=nil.
		}
		report.CasesCompleted++
	}
	return report, nil
}
func compare(actual lifecycle.Result, want Expected, digest string) []string {
	differences := []string{}
	if actual.Mode != Mode {
		differences = append(differences, "mode")
	}
	if actual.Authority != Authority {
		differences = append(differences, "authority")
	}
	if actual.ModelDigest != digest {
		differences = append(differences, "modelDigest")
	}
	if actual.OK != want.OK {
		differences = append(differences, "ok")
	}
	if actual.TransitionID != want.TransitionID {
		differences = append(differences, "transitionId")
	}
	if !reflect.DeepEqual(actual.Task, want.Task) {
		differences = append(differences, "task")
	}
	if !reflect.DeepEqual(actual.Refusals, want.Refusals) {
		differences = append(differences, "refusals")
	}

	// Whole-struct equality is authoritative. Field diagnostics must not silently
	// ignore a future generated-result field added without a profile review.
	complete := lifecycle.Result{Mode: Mode, Authority: Authority, ModelDigest: digest, OK: want.OK, TransitionID: want.TransitionID, Task: want.Task, Refusals: want.Refusals}
	if !reflect.DeepEqual(actual, complete) && len(differences) == 0 {
		differences = append(differences, "result")
	}
	return differences
}

// Execute additionally records the hash of observed original fixture bytes. A
// typed Run cannot supply that fact. Neither hash authenticates actors or code.
func Execute(raw []byte) (Report, error) { return executeWith(raw, lifecycle.Transition) }
func executeWith(raw []byte, apply transition) (Report, error) {
	return executeFor(raw, apply, defaultBinding(), false)
}

func executeFor(raw []byte, apply transition, binding Binding, provider bool) (Report, error) {
	s, e := decodeFor(raw, binding)
	if e != nil {
		return Report{}, e
	}
	var h [sha256.Size]byte
	if provider {
		// Bind the bytes decoded above, not bytes a callback may later mutate
		// through a caller-owned closure. Concurrent mutation remains unsupported.
		h = sha256.Sum256(raw)
	}
	report, e := runFor(s, apply, binding, provider)
	if e != nil {
		return Report{}, e
	}
	if !provider {
		h = sha256.Sum256(raw)
	}
	v := hex.EncodeToString(h[:])
	report.SourceDigest = &v
	return report, nil
}

type actualDTO struct {
	Mode         string   `json:"mode"`
	Authority    string   `json:"authority"`
	ModelDigest  string   `json:"modelDigest"`
	OK           bool     `json:"ok"`
	TransitionID string   `json:"transitionId"`
	Task         *taskDTO `json:"task"`
	Refusals     []string `json:"refusals"`
}
type entryDTO struct {
	CaseID      string      `json:"caseId"`
	StepID      string      `json:"stepId"`
	Input       taskDTO     `json:"input"`
	Command     commandDTO  `json:"command"`
	Expected    expectedDTO `json:"expected"`
	Actual      actualDTO   `json:"actual"`
	Matched     bool        `json:"matched"`
	Differences []string    `json:"differences"`
}

func entryToDTO(e Entry) entryDTO {
	r := e.Actual
	a := actualDTO{r.Mode, r.Authority, r.ModelDigest, r.OK, r.TransitionID, nil, copyStrings(r.Refusals)}
	if r.Task != nil {
		v := taskToDTO(*r.Task)
		a.Task = &v
	}
	return entryDTO{e.CaseID, e.StepID, taskToDTO(e.Input), commandToDTO(e.Command), expectedToDTO(e.Expected), a, e.Matched, copyStrings(e.Differences)}
}
func boundedEntry(e Entry) bool { return boundedEntryFor(e, defaultBinding()) }

func boundedEntryFor(e Entry, binding Binding) bool {
	b := budget{limit: MaxReportBytes}
	if !model.Identifier(e.CaseID) || !model.Identifier(e.StepID) || !b.task(e.Input) || !b.command(e.Command) || !b.expected(e.Expected) {
		return false
	}
	r := e.Actual
	for _, s := range []string{r.Mode, r.Authority, r.ModelDigest, r.TransitionID} {
		if !b.text(s, MaxTextBytes) {
			return false
		}
	}
	if !b.reasons(r.Refusals) || (r.Task != nil && !b.task(*r.Task)) || e.Differences == nil || len(e.Differences) > 8 {
		return false
	}
	for _, s := range e.Differences {
		switch s {
		case "mode", "authority", "modelDigest", "ok", "transitionId", "task", "refusals", "result":
		default:
			return false
		}
	}
	actualDifferences := compare(e.Actual, e.Expected, binding.ModelDigest)
	return reflect.DeepEqual(e.Differences, actualDifferences) && e.Matched == (len(actualDifferences) == 0)
}

// EncodeReport returns one complete bounded JSON value plus newline. It never
// writes output. The caller must check both encoding and I/O errors.
func EncodeReport(r Report) ([]byte, error) { return encodeReportFor(r, defaultBinding(), false) }

func encodeReportFor(r Report, binding Binding, provider bool) ([]byte, error) {
	if r.ModelDigest != binding.ModelDigest || !hexDigest(r.ScenarioDigest) || (r.SourceDigest != nil && !hexDigest(*r.SourceDigest)) || len(r.Entries) < 1 || len(r.Entries) > MaxSteps || r.StepsExecuted != len(r.Entries) || r.CasesCompleted < 0 || r.CasesCompleted > MaxCases || (r.Matched && r.CasesCompleted < 1) {
		return nil, ErrReportLimit
	}
	entries := make([]entryDTO, len(r.Entries))
	encodedSize := 4096
	groups := 0
	groupSteps := 0
	lastCase := ""
	seenIDs := map[string]bool{}
	for i, e := range r.Entries {
		if !boundedEntryFor(e, binding) || (provider && !providerResultValid(e.Actual, binding)) {
			return nil, ErrReportLimit
		}
		if e.CaseID == lastCase && i > 0 {
			prior := r.Entries[i-1]
			current := prior.Input
			if prior.Actual.OK {
				if prior.Actual.Task == nil {
					return nil, ErrReportLimit
				}
				current = *prior.Actual.Task
			}
			if !reflect.DeepEqual(e.Input, current) {
				return nil, ErrReportLimit
			}
		}
		if e.CaseID != lastCase {
			if seenIDs[e.CaseID] {
				return nil, ErrReportLimit
			}
			seenIDs[e.CaseID] = true
			lastCase = e.CaseID
			groups++
			groupSteps = 0
			if groups > MaxCases {
				return nil, ErrReportLimit
			}
		}
		groupSteps++
		if groupSteps > MaxStepsPerCase {
			return nil, ErrReportLimit
		}
		if seenIDs[e.StepID] {
			return nil, ErrReportLimit
		}
		seenIDs[e.StepID] = true
		if !e.Matched && (r.Matched || i != len(r.Entries)-1) {
			return nil, ErrReportLimit
		}
		entries[i] = entryToDTO(e)
		raw, err := encodeJSON(entries[i])
		if err != nil {
			return nil, ErrReportLimit
		}
		encodedSize += len(raw) + 1
		if encodedSize > MaxReportBytes {
			return nil, ErrReportLimit
		}
	}
	wantCompleted := groups
	if !r.Matched {
		wantCompleted--
	}
	if r.CasesCompleted != wantCompleted {
		return nil, ErrReportLimit
	}
	if !r.Matched && r.Entries[len(r.Entries)-1].Matched {
		return nil, ErrReportLimit
	}
	format, source := ReportVersion, ""
	if provider {
		format, source = BoundReportVersion, BindingSource
	}
	dto := struct {
		FormatVersion   string     `json:"formatVersion"`
		ScenarioVersion string     `json:"scenarioVersion"`
		ProfileVersion  string     `json:"profileVersion"`
		IdentitySource  string     `json:"identitySource,omitempty"`
		Mode            string     `json:"mode"`
		Authority       string     `json:"authority"`
		ModelDigest     string     `json:"modelDigest"`
		ScenarioDigest  string     `json:"scenarioDigest"`
		SourceDigest    *string    `json:"sourceDigest"`
		Matched         bool       `json:"matched"`
		CasesCompleted  int        `json:"casesCompleted"`
		StepsExecuted   int        `json:"stepsExecuted"`
		Entries         []entryDTO `json:"entries"`
	}{format, ScenarioVersion, binding.ProfileVersion, source, Mode, Authority, r.ModelDigest, r.ScenarioDigest, r.SourceDigest, r.Matched, r.CasesCompleted, r.StepsExecuted, entries}
	raw, e := encodeJSON(dto)
	if e != nil || len(raw) > MaxReportBytes {
		return nil, ErrReportLimit
	}
	return raw, nil
}
