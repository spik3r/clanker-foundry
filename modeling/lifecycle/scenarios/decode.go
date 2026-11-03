package scenarios

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// Transport DTOs are separate from the typed scenario profile. Field order is
// also the canonical semantic-content order, except the excluded formatVersion.
type suiteDTO struct {
	FormatVersion   string    `json:"formatVersion"`
	ScenarioVersion string    `json:"scenarioVersion"`
	ProfileVersion  string    `json:"profileVersion"`
	ModelDigest     string    `json:"modelDigest"`
	ID              string    `json:"id"`
	Cases           []caseDTO `json:"cases"`
}
type caseDTO struct {
	ID      string    `json:"id"`
	Initial taskDTO   `json:"initial"`
	Steps   []stepDTO `json:"steps"`
}
type stepDTO struct {
	ID       string      `json:"id"`
	Command  commandDTO  `json:"command"`
	Expected expectedDTO `json:"expected"`
}
type taskDTO struct {
	ID         string    `json:"id"`
	State      string    `json:"state"`
	Revision   string    `json:"revision"`
	Outcome    string    `json:"outcome"`
	Acceptance *[]string `json:"acceptance"`
	Owner      *string   `json:"owner"`
}
type reviewDTO struct {
	TaskID       string `json:"taskId"`
	TaskRevision string `json:"taskRevision"`
	ModelDigest  string `json:"modelDigest"`
	TaskDigest   string `json:"taskDigest"`
	Reviewer     string `json:"reviewer"`
	Decision     string `json:"decision"`
}
type commandDTO struct {
	Type             string     `json:"type"`
	Actor            string     `json:"actor"`
	ExpectedRevision string     `json:"expectedRevision"`
	Review           *reviewDTO `json:"review"`
}
type expectedDTO struct {
	OK           bool     `json:"ok"`
	TransitionID string   `json:"transitionId"`
	Task         *taskDTO `json:"task"`
	Refusals     []string `json:"refusals"`
}

func decimal(s string) (uint64, error) {
	if s == "" || len(s) > 20 || (len(s) > 1 && s[0] == '0') {
		return 0, ErrInvalidScenario
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrInvalidScenario
		}
	}
	n, e := strconv.ParseUint(s, 10, 64)
	if e != nil {
		return 0, ErrInvalidScenario
	}
	return n, nil
}
func fromTask(d taskDTO) (lifecycle.Task, error) {
	n, e := decimal(d.Revision)
	if e != nil {
		return lifecycle.Task{}, e
	}
	t := lifecycle.Task{ID: d.ID, State: lifecycle.State(d.State), Revision: n, Outcome: d.Outcome, Owner: d.Owner}
	if d.Acceptance != nil {
		t.Acceptance = *d.Acceptance
	}
	return copyTask(t), nil
}
func fromCommand(d commandDTO) (lifecycle.Command, error) {
	n, e := decimal(d.ExpectedRevision)
	if e != nil {
		return lifecycle.Command{}, e
	}
	c := lifecycle.Command{Type: d.Type, Actor: d.Actor, ExpectedRevision: n}
	if d.Review != nil {
		r := d.Review
		n, e := decimal(r.TaskRevision)
		if e != nil {
			return lifecycle.Command{}, e
		}
		c.Review = &lifecycle.Review{TaskID: r.TaskID, TaskRevision: n, ModelDigest: r.ModelDigest, TaskDigest: r.TaskDigest, Reviewer: r.Reviewer, Decision: r.Decision}
	}
	return c, nil
}

// Decode rejects the entire malformed fixture; it never invokes the reducer.
func Decode(raw []byte) (Suite, error) { return decodeFor(raw, defaultBinding()) }

func decodeFor(raw []byte, binding Binding) (Suite, error) {
	var d suiteDTO
	if e := strictjson.Decode(raw, &d, MaxSourceBytes, MaxDepth); e != nil {
		return Suite{}, ErrInvalidScenario
	}
	if d.FormatVersion != FormatVersion || d.ScenarioVersion != ScenarioVersion {
		return Suite{}, ErrInvalidScenario
	}
	s := Suite{Version: 1, ID: d.ID, ProfileVersion: d.ProfileVersion, ModelDigest: d.ModelDigest, Cases: make([]Case, len(d.Cases))}
	for i, c := range d.Cases {
		initial, e := fromTask(c.Initial)
		if e != nil {
			return Suite{}, e
		}
		v := Case{ID: c.ID, Initial: initial, Steps: make([]Step, len(c.Steps))}
		for j, x := range c.Steps {
			cmd, e := fromCommand(x.Command)
			if e != nil {
				return Suite{}, e
			}
			want := Expected{OK: x.Expected.OK, TransitionID: x.Expected.TransitionID, Refusals: copyStrings(x.Expected.Refusals)}
			if x.Expected.Task != nil {
				task, e := fromTask(*x.Expected.Task)
				if e != nil {
					return Suite{}, e
				}
				want.Task = &task
			}
			v.Steps[j] = Step{ID: x.ID, Command: cmd, Expected: want}
		}
		s.Cases[i] = v
	}
	if e := validateFor(s, binding); e != nil {
		return Suite{}, e
	}
	return s, nil
}
func taskToDTO(t lifecycle.Task) taskDTO {
	t = copyTask(t)
	var a *[]string
	if t.Acceptance != nil {
		a = &t.Acceptance
	}
	return taskDTO{t.ID, string(t.State), strconv.FormatUint(t.Revision, 10), t.Outcome, a, t.Owner}
}
func commandToDTO(c lifecycle.Command) commandDTO {
	d := commandDTO{Type: c.Type, Actor: c.Actor, ExpectedRevision: strconv.FormatUint(c.ExpectedRevision, 10)}
	if c.Review != nil {
		r := c.Review
		d.Review = &reviewDTO{r.TaskID, strconv.FormatUint(r.TaskRevision, 10), r.ModelDigest, r.TaskDigest, r.Reviewer, r.Decision}
	}
	return d
}
func expectedToDTO(e Expected) expectedDTO {
	d := expectedDTO{OK: e.OK, TransitionID: e.TransitionID, Refusals: copyStrings(e.Refusals)}
	if e.Task != nil {
		t := taskToDTO(*e.Task)
		d.Task = &t
	}
	return d
}
func toDTO(s Suite) suiteDTO {
	d := suiteDTO{FormatVersion, ScenarioVersion, s.ProfileVersion, s.ModelDigest, s.ID, make([]caseDTO, len(s.Cases))}
	for i, c := range s.Cases {
		v := caseDTO{c.ID, taskToDTO(c.Initial), make([]stepDTO, len(c.Steps))}
		for j, x := range c.Steps {
			v.Steps[j] = stepDTO{x.ID, commandToDTO(x.Command), expectedToDTO(x.Expected)}
		}
		d.Cases[i] = v
	}
	return d
}
func encodeJSON(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Encode preserves ordered data and nil/empty distinctions, without deriving or
// changing any expected result. Successful output is decodable by JSON-v1.
func Encode(s Suite) ([]byte, error) { return encodeFor(s, defaultBinding()) }

func encodeFor(s Suite, binding Binding) ([]byte, error) {
	if e := validateFor(s, binding); e != nil {
		return nil, e
	}
	return encodeJSON(toDTO(s))
}

// Digest covers semantic scenario data, including every literal expectation.
// Original JSON formatting and its format-version label are not semantic inputs.
func Digest(s Suite) (string, error) { return digestFor(s, defaultBinding()) }

func digestFor(s Suite, binding Binding) (string, error) {
	if e := validateFor(s, binding); e != nil {
		return "", e
	}
	d := toDTO(s)
	v := struct {
		ScenarioVersion string    `json:"scenarioVersion"`
		ProfileVersion  string    `json:"profileVersion"`
		ModelDigest     string    `json:"modelDigest"`
		ID              string    `json:"id"`
		Cases           []caseDTO `json:"cases"`
	}{d.ScenarioVersion, d.ProfileVersion, d.ModelDigest, d.ID, d.Cases}
	b, e := encodeJSON(v)
	if e != nil {
		return "", ErrInvalidScenario
	}
	h := sha256.Sum256(bytes.TrimSuffix(b, []byte{'\n'}))
	return hex.EncodeToString(h[:]), nil
}
