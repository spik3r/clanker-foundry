// Package rawinputacceptance holds independent, black-box expectations for the
// raw JSON adapter package rawinput. Every expectation is a literal taken from
// the written adapter contract and the lifecycle README; none is computed by
// running the reducer or the adapter. Case data lives in expected.json.
package rawinputacceptance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rawinput"
)

// ---------------------------------------------------------------- case table

type wireTask struct {
	ID         string   `json:"id"`
	State      string   `json:"state"`
	Revision   string   `json:"revision"`
	Outcome    string   `json:"outcome"`
	Acceptance []string `json:"acceptance"`
	Owner      *string  `json:"owner"`
}

func (w wireTask) task(t *testing.T) lifecycle.Task {
	t.Helper()
	rev, err := strconv.ParseUint(w.Revision, 10, 64)
	if err != nil {
		t.Fatalf("test data revision %q: %v", w.Revision, err)
	}
	return lifecycle.Task{ID: w.ID, State: lifecycle.State(w.State), Revision: rev, Outcome: w.Outcome, Acceptance: w.Acceptance, Owner: w.Owner}
}

type expectation struct {
	Kind       string          `json:"kind"`
	Transition string          `json:"transition"`
	Refusals   []string        `json:"refusals"`
	Task       json.RawMessage `json:"task"`
	TaskText   string          `json:"taskText"`
	Unchanged  bool            `json:"unchanged"`
	Input      string          `json:"input"`
	Reason     string          `json:"reason"`
}

type tcase struct {
	ID        string      `json:"id"`
	Group     string      `json:"group"`
	Note      string      `json:"note"`
	Task      string      `json:"task"`
	Command   string      `json:"command"`
	ChainFrom string      `json:"chainFrom"`
	Expect    expectation `json:"expect"`
}

type table struct {
	Schema        string            `json:"schema"`
	Digests       map[string]string `json:"digests"`
	RevisionMatrx struct {
		Rejected     []string `json:"rejected"`
		WrongTypeRaw []string `json:"wrongTypeRaw"`
	} `json:"revisionMatrix"`
	Cases []tcase `json:"cases"`
}

var (
	tokenRE     = regexp.MustCompile(`\{\{(x[0-9a-fA-F]{2}|rep:[0-9]+\|[^}]*)\}\}`)
	digestHexes map[string]string
)

func loadTable(t *testing.T) *table {
	t.Helper()
	b, err := os.ReadFile("expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var tb table
	if err := json.Unmarshal(b, &tb); err != nil {
		t.Fatal(err)
	}
	if tb.Schema != "rawinput-acceptance-v1" {
		t.Fatalf("schema %q", tb.Schema)
	}
	digestHexes = map[string]string{}
	for name, canon := range tb.Digests {
		if !json.Valid([]byte(canon)) || !strings.HasPrefix(canon, `{"id":`) {
			t.Fatalf("digest %s canonical text malformed: %s", name, canon)
		}
		sum := sha256.Sum256([]byte(canon))
		digestHexes[name] = hex.EncodeToString(sum[:])
	}
	return &tb
}

// otherModelDigest is any valid lowercase 64-hex digest different from the model digest.
func otherModelDigest(t *testing.T) string {
	t.Helper()
	m := lifecycle.ModelDigest
	if len(m) != 64 || strings.Trim(m, "0123456789abcdef") != "" {
		t.Fatalf("ModelDigest is not lowercase 64-hex: %q", m)
	}
	if m[0] == '0' {
		return "1" + m[1:]
	}
	return "0" + m[1:]
}

func expand(t *testing.T, s string) string {
	t.Helper()
	s = strings.ReplaceAll(s, "${MODEL}", lifecycle.ModelDigest)
	s = strings.ReplaceAll(s, "${OTHER_MODEL}", otherModelDigest(t))
	for name, h := range digestHexes {
		s = strings.ReplaceAll(s, "${D:"+name+"}", h)
	}
	if strings.Contains(s, "${") {
		t.Fatalf("unresolved placeholder in %.120q", s)
	}
	return tokenRE.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[2 : len(m)-2]
		if inner[0] == 'x' {
			v, err := strconv.ParseUint(inner[1:], 16, 8)
			if err != nil {
				t.Fatal(err)
			}
			return string([]byte{byte(v)})
		}
		rest := strings.TrimPrefix(inner, "rep:")
		bar := strings.Index(rest, "|")
		n, err := strconv.Atoi(rest[:bar])
		if err != nil {
			t.Fatal(err)
		}
		return strings.Repeat(rest[bar+1:], n)
	})
}

// ---------------------------------------------------------------- call helpers

var rejectedLiteral = func() lifecycle.Result {
	return lifecycle.Result{
		Mode:         "example-only",
		Authority:    "none",
		ModelDigest:  lifecycle.ModelDigest,
		OK:           false,
		TransitionID: "",
		Task:         nil,
		Refusals:     []string{"INVALID_INPUT"},
	}
}

func checkMeta(t *testing.T, r lifecycle.Result) {
	t.Helper()
	if r.Mode != "example-only" || r.Authority != "none" || r.ModelDigest != lifecycle.ModelDigest {
		t.Fatalf("metadata: mode=%q authority=%q modelDigest=%q", r.Mode, r.Authority, r.ModelDigest)
	}
}

// call runs the adapter twice on private copies of the inputs, requires equal
// outcomes, requires that the caller's slices are untouched, and checks metadata.
func call(t *testing.T, taskJSON, cmdJSON []byte) (lifecycle.Result, error) {
	t.Helper()
	beforeT := append([]byte(nil), taskJSON...)
	beforeC := append([]byte(nil), cmdJSON...)
	if taskJSON == nil {
		beforeT = nil
	}
	if cmdJSON == nil {
		beforeC = nil
	}
	r1, e1 := rawinput.Transition(taskJSON, cmdJSON)
	if string(taskJSON) != string(beforeT) || string(cmdJSON) != string(beforeC) {
		t.Fatal("adapter modified caller input bytes")
	}
	r2, e2 := rawinput.Transition(taskJSON, cmdJSON)
	if string(taskJSON) != string(beforeT) || string(cmdJSON) != string(beforeC) {
		t.Fatal("adapter modified caller input bytes on second call")
	}
	if !reflect.DeepEqual(r1, r2) || !reflect.DeepEqual(e1, e2) {
		t.Fatalf("nondeterministic: %+v / %v then %+v / %v", r1, e1, r2, e2)
	}
	checkMeta(t, r1)
	return r1, e1
}

func requireInputError(t *testing.T, err error, input, reason string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want *InputError{%s,%s}, got nil error", input, reason)
	}
	var ie *rawinput.InputError
	if !errors.As(err, &ie) || ie == nil {
		t.Fatalf("error is not *rawinput.InputError: %T %v", err, err)
	}
	if ie.Input != input || string(ie.Reason) != reason {
		t.Fatalf("got Input=%q Reason=%q, want Input=%q Reason=%q", ie.Input, ie.Reason, input, reason)
	}
	if ie.Error() == "" {
		t.Fatal("InputError.Error() is empty")
	}
}

func requireRejected(t *testing.T, r lifecycle.Result, err error, input, reason string) {
	t.Helper()
	requireInputError(t, err, input, reason)
	if !reflect.DeepEqual(r, rejectedLiteral()) {
		t.Fatalf("rejected result mismatch: %+v", r)
	}
}

func requireInvalidReducer(t *testing.T, r lifecycle.Result, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("semantic invalid input must have nil error, got %v", err)
	}
	if !reflect.DeepEqual(r, rejectedLiteral()) {
		t.Fatalf("INVALID_INPUT result mismatch: %+v", r)
	}
}

// expectedTask returns the literal expected task of an ok case.
func expectedTask(t *testing.T, e expectation) lifecycle.Task {
	t.Helper()
	text := e.TaskText
	if text == "" {
		text = string(e.Task)
	}
	if text == "" {
		t.Fatal("ok expectation without task")
	}
	var w wireTask
	if err := json.Unmarshal([]byte(expand(t, text)), &w); err != nil {
		t.Fatalf("expected task: %v", err)
	}
	return w.task(t)
}

// inputTask parses the case's own task text with encoding/json (not the adapter).
func inputTask(t *testing.T, raw string) lifecycle.Task {
	t.Helper()
	var w wireTask
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("test input task does not parse: %v", err)
	}
	return w.task(t)
}

// refusedTransition is the transition ID a guard or limit refusal reports. The
// model (examples/lifecycle.json) names each transition "T_" plus its trigger,
// and a refusal for an unavailable state/trigger pair names none.
func refusedTransition(t *testing.T, rawCmd string, refusals []string) string {
	t.Helper()
	if len(refusals) == 1 && refusals[0] == "ILLEGAL_TRANSITION" {
		return ""
	}
	var c struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(rawCmd), &c); err != nil || c.Type == "" {
		t.Fatalf("refused case needs a command type: %v", err)
	}
	return "T_" + c.Type
}

func runCase(t *testing.T, c tcase, prevExpected map[string]lifecycle.Task) {
	t.Helper()
	rawTask, rawCmd := expand(t, c.Task), expand(t, c.Command)
	r, err := call(t, []byte(rawTask), []byte(rawCmd))
	e := c.Expect
	switch e.Kind {
	case "ok":
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		want := expectedTask(t, e)
		wantR := lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: lifecycle.ModelDigest, OK: true,
			TransitionID: e.Transition, Task: &want, Refusals: []string{}}
		if !reflect.DeepEqual(r, wantR) {
			t.Fatalf("got %+v task=%+v\nwant %+v task=%+v", r, deref(r.Task), wantR, want)
		}
		if c.ChainFrom != "" {
			prev, found := prevExpected[c.ChainFrom]
			if !found || !reflect.DeepEqual(inputTask(t, rawTask), prev) {
				t.Fatalf("hand-written chain input differs from output of %s", c.ChainFrom)
			}
		}
		prevExpected[c.ID] = want
	case "refused":
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		var want lifecycle.Task
		if len(e.Task) > 0 || e.TaskText != "" {
			want = expectedTask(t, e)
		} else if e.Unchanged {
			want = inputTask(t, rawTask)
		} else {
			t.Fatal("refused expectation needs unchanged or a task")
		}
		wantR := lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: lifecycle.ModelDigest, OK: false,
			TransitionID: refusedTransition(t, rawCmd, e.Refusals), Task: &want, Refusals: e.Refusals}
		if len(e.Refusals) == 0 {
			t.Fatal("refused expectation without refusals")
		}
		if !reflect.DeepEqual(r, wantR) {
			t.Fatalf("got %+v task=%+v\nwant %+v task=%+v", r, deref(r.Task), wantR, want)
		}
		if c.ChainFrom != "" {
			prev, found := prevExpected[c.ChainFrom]
			if !found || !reflect.DeepEqual(inputTask(t, rawTask), prev) {
				t.Fatalf("hand-written chain input differs from output of %s", c.ChainFrom)
			}
		}
	case "invalid":
		requireInvalidReducer(t, r, err)
	case "rejected":
		requireRejected(t, r, err, e.Input, e.Reason)
	default:
		t.Fatalf("unknown expectation kind %q", e.Kind)
	}
}

func deref(t *lifecycle.Task) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

// ---------------------------------------------------------------- tests

func TestConstants(t *testing.T) {
	if rawinput.MaxTaskBytes != 262144 || rawinput.MaxCommandBytes != 16384 || rawinput.MaxDepth != 4 {
		t.Fatalf("limits: %d %d %d", rawinput.MaxTaskBytes, rawinput.MaxCommandBytes, rawinput.MaxDepth)
	}
	if string(rawinput.ReasonStrictJSON) != "strict_json" || string(rawinput.ReasonRevision) != "invalid_revision" {
		t.Fatalf("reasons: %q %q", rawinput.ReasonStrictJSON, rawinput.ReasonRevision)
	}
}

func TestCaseTable(t *testing.T) {
	tb := loadTable(t)
	prev := map[string]lifecycle.Task{}
	seen := map[string]bool{}
	groups := map[string]int{}
	for _, c := range tb.Cases {
		if seen[c.ID] {
			t.Fatalf("duplicate case id %s", c.ID)
		}
		seen[c.ID] = true
		groups[c.Group]++
		c := c
		t.Run(c.ID, func(t *testing.T) { runCase(t, c, prev) })
	}
	// Guard against a silently shrunken table.
	minGroups := map[string]int{
		"ok": 10, "seq": 7, "unusual": 15, "invalid": 30, "refused": 20, "refused-review": 15, "illegal": 12,
		"maxrev": 10, "rev-accepted": 5, "task-rejected": 80, "command-rejected": 70, "order": 8,
	}
	for g, n := range minGroups {
		if groups[g] < n {
			t.Errorf("group %s has %d cases, want at least %d", g, groups[g], n)
		}
	}
	for _, id := range []string{"command-not-json", "task-not-json", "task-surrogate-lone-high", "task-utf8-invalid-ff-in-value",
		"maxrev-prepare-limit", "refuse-acceptance-null", "refuse-acceptance-empty-list", "ok-start-assigns-actor",
		"refuse-start-owner-empty-present", "seq-6-approve", "refuse-review-stale-content-digest"} {
		if !seen[id] {
			t.Errorf("required case %s missing from table", id)
		}
	}
}

// Digest contract: the hand-written canonical strings in expected.json must
// match the two oracle vectors, and hashing them with crypto/sha256 must give
// the vector hashes. lifecycle.TaskDigest is deliberately not used.
func TestDigestVectorsAgainstOracleFixture(t *testing.T) {
	tb := loadTable(t)
	var fx struct {
		DigestVectors []struct {
			Canonical string `json:"canonical"`
			Sha256    string `json:"sha256"`
		} `json:"digestVectors"`
	}
	b, err := os.ReadFile("../examples/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &fx); err != nil {
		t.Fatal(err)
	}
	if len(fx.DigestVectors) != 2 {
		t.Fatalf("oracle digestVectors: %d", len(fx.DigestVectors))
	}
	wantSHA := []string{
		"e0ab0176010aa6044142a2c830c3a6d146f41f862393d3a54bdb410071babc94",
		"d5137867f36f4f6f8f51dfbf3ad43924b0f937d409a10b36af4cb860acf7760f",
	}
	for i, name := range []string{"vec1", "vec2"} {
		v := fx.DigestVectors[i]
		if v.Sha256 != wantSHA[i] {
			t.Fatalf("fixture drift in vector %d hash", i)
		}
		if tb.Digests[name] != v.Canonical {
			t.Fatalf("hand-written %s canonical differs from fixture vector:\n%s\n%s", name, tb.Digests[name], v.Canonical)
		}
		if digestHexes[name] != wantSHA[i] {
			t.Fatalf("sha256 of %s = %s, want %s", name, digestHexes[name], wantSHA[i])
		}
	}
}

// Oracle special-03 and special-04 are pending in the fixture because no raw
// adapter existed. They run here with the fixture's own raw strings.
func TestOracleSpecialRawCases(t *testing.T) {
	b, err := os.ReadFile("../examples/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var fx struct {
		Special []struct {
			CaseID     string          `json:"caseId"`
			RawTask    string          `json:"rawTask"`
			RawCommand string          `json:"rawCommand"`
			OK         *bool           `json:"ok"`
			Task       json.RawMessage `json:"task"`
			Transition json.RawMessage `json:"transition"`
			Code       string          `json:"code"`
		} `json:"special"`
	}
	if err := json.Unmarshal(b, &fx); err != nil {
		t.Fatal(err)
	}
	// special-04 note: its rawTask "{}" is not a valid Task under the strict wire
	// format (all six keys are required), so the first rejection is the task, not
	// the command, even though the case name says "invalid command input". The
	// stronger variant (valid task, malformed command) is table case command-not-json.
	want := map[string]struct{ task, command, input string }{
		"special-03": {"not json", "{}", "task"},
		"special-04": {"{}", "not json", "task"},
	}
	for id, w := range want {
		id, w := id, w
		t.Run(id, func(t *testing.T) {
			found := false
			for _, s := range fx.Special {
				if s.CaseID != id {
					continue
				}
				found = true
				if s.RawTask != w.task || s.RawCommand != w.command {
					t.Fatalf("fixture drift: rawTask=%q rawCommand=%q", s.RawTask, s.RawCommand)
				}
				if s.OK == nil || *s.OK || string(s.Task) != "null" || string(s.Transition) != "null" || s.Code != "INVALID_INPUT" {
					t.Fatalf("fixture drift in expected outcome: %+v", s)
				}
				r, err := call(t, []byte(s.RawTask), []byte(s.RawCommand))
				requireRejected(t, r, err, w.input, "strict_json")
				if r.OK || r.Task != nil || r.TransitionID != "" || len(r.Refusals) != 1 || r.Refusals[0] != s.Code {
					t.Fatalf("outcome does not match fixture expectation: %+v", r)
				}
			}
			if !found {
				t.Fatalf("%s not in fixture", id)
			}
		})
	}
}

func TestNilAndEmptyInputs(t *testing.T) {
	good := []byte(`{"type":"prepare","actor":"owner-1","expectedRevision":"0","review":null}`)
	goodTask := []byte(`{"id":"task-1","state":"Draft","revision":"0","outcome":"Ship outcome","acceptance":["Verified"],"owner":"owner-1"}`)
	r, err := call(t, nil, good)
	requireRejected(t, r, err, "task", "strict_json")
	r, err = call(t, []byte{}, good)
	requireRejected(t, r, err, "task", "strict_json")
	r, err = call(t, goodTask, nil)
	requireRejected(t, r, err, "command", "strict_json")
	r, err = call(t, goodTask, []byte{})
	requireRejected(t, r, err, "command", "strict_json")
	r, err = call(t, nil, nil)
	requireRejected(t, r, err, "task", "strict_json")
}

// Revision strings: every rejected spelling in each of the three positions.
func TestRevisionMatrixRejected(t *testing.T) {
	tb := loadTable(t)
	const (
		taskTmpl = `{"id":"task-1","state":"Draft","revision":%s,"outcome":"Ship outcome","acceptance":["Verified"],"owner":"owner-1"}`
		cmdTmpl  = `{"type":"prepare","actor":"owner-1","expectedRevision":%s,"review":null}`
		goodTask = `{"id":"task-1","state":"Draft","revision":"0","outcome":"Ship outcome","acceptance":["Verified"],"owner":"owner-1"}`
		goodCmd  = `{"type":"prepare","actor":"owner-1","expectedRevision":"0","review":null}`
	)
	reviewCmd := func(rev string) string {
		return `{"type":"approve","actor":"reviewer-1","expectedRevision":"3","review":{"taskId":"task-1","taskRevision":` + rev +
			`,"modelDigest":"` + lifecycle.ModelDigest + `","taskDigest":"` + strings.Repeat("0", 64) + `","reviewer":"reviewer-1","decision":"approved"}}`
	}
	if len(tb.RevisionMatrx.Rejected) < 20 {
		t.Fatal("revision matrix shrank")
	}
	for _, v := range tb.RevisionMatrx.Rejected {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		lit := string(b)
		t.Run("task/"+strconv.QuoteToASCII(v), func(t *testing.T) {
			r, err := call(t, []byte(fmt.Sprintf(taskTmpl, lit)), []byte(goodCmd))
			requireRejected(t, r, err, "task", "invalid_revision")
		})
		t.Run("command/"+strconv.QuoteToASCII(v), func(t *testing.T) {
			r, err := call(t, []byte(goodTask), []byte(fmt.Sprintf(cmdTmpl, lit)))
			requireRejected(t, r, err, "command", "invalid_revision")
		})
		t.Run("review/"+strconv.QuoteToASCII(v), func(t *testing.T) {
			r, err := call(t, []byte(goodTask), []byte(reviewCmd(lit)))
			requireRejected(t, r, err, "command", "invalid_revision")
		})
	}
	for _, raw := range tb.RevisionMatrx.WrongTypeRaw {
		raw := raw
		t.Run("task-type/"+raw, func(t *testing.T) {
			r, err := call(t, []byte(fmt.Sprintf(taskTmpl, raw)), []byte(goodCmd))
			requireRejected(t, r, err, "task", "strict_json")
		})
		t.Run("command-type/"+raw, func(t *testing.T) {
			r, err := call(t, []byte(goodTask), []byte(fmt.Sprintf(cmdTmpl, raw)))
			requireRejected(t, r, err, "command", "strict_json")
		})
		t.Run("review-type/"+raw, func(t *testing.T) {
			r, err := call(t, []byte(goodTask), []byte(reviewCmd(raw)))
			requireRejected(t, r, err, "command", "strict_json")
		})
	}
}

// ---------------------------------------------------------------- limits

const (
	limTask = `{"id":"task-1","state":"Draft","revision":"0","outcome":"Ship outcome","acceptance":["Verified"],"owner":"owner-1"}`
	limCmd  = `{"type":"prepare","actor":"owner-1","expectedRevision":"0","review":null}`
)

func str(s string) *string { return &s }

func readyTask() *lifecycle.Task {
	return &lifecycle.Task{ID: "task-1", State: "Ready", Revision: 1, Outcome: "Ship outcome", Acceptance: []string{"Verified"}, Owner: str("owner-1")}
}

func requirePrepared(t *testing.T, r lifecycle.Result, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	want := lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: lifecycle.ModelDigest, OK: true,
		TransitionID: "T_prepare", Task: readyTask(), Refusals: []string{}}
	if !reflect.DeepEqual(r, want) {
		t.Fatalf("got %+v", r)
	}
}

// pad grows s to exactly n bytes with insignificant JSON whitespace, appended
// (trailing) or prepended (leading).
func pad(t *testing.T, s string, n int, leading bool) []byte {
	t.Helper()
	if len(s) > n {
		t.Fatalf("pad: %d > %d", len(s), n)
	}
	const cycle = " \n\t\r"
	var sb strings.Builder
	for i := 0; i < n-len(s); i++ {
		sb.WriteByte(cycle[i%4])
	}
	if leading {
		return []byte(sb.String() + s)
	}
	return []byte(s + sb.String())
}

func TestSizeLimits(t *testing.T) {
	t.Run("task exactly at limit, trailing whitespace", func(t *testing.T) {
		in := pad(t, limTask, rawinput.MaxTaskBytes, false)
		if len(in) != 262144 {
			t.Fatal(len(in))
		}
		r, err := call(t, in, []byte(limCmd))
		requirePrepared(t, r, err)
	})
	t.Run("task exactly at limit, leading whitespace", func(t *testing.T) {
		r, err := call(t, pad(t, limTask, 262144, true), []byte(limCmd))
		requirePrepared(t, r, err)
	})
	t.Run("task one byte over", func(t *testing.T) {
		in := pad(t, limTask, 262145, false)
		r, err := call(t, in, []byte(limCmd))
		requireRejected(t, r, err, "task", "strict_json")
	})
	t.Run("task one byte over, leading whitespace", func(t *testing.T) {
		r, err := call(t, pad(t, limTask, 262145, true), []byte(limCmd))
		requireRejected(t, r, err, "task", "strict_json")
	})
	t.Run("command exactly at limit", func(t *testing.T) {
		in := pad(t, limCmd, rawinput.MaxCommandBytes, false)
		if len(in) != 16384 {
			t.Fatal(len(in))
		}
		r, err := call(t, []byte(limTask), in)
		requirePrepared(t, r, err)
	})
	t.Run("command exactly at limit, leading whitespace", func(t *testing.T) {
		r, err := call(t, []byte(limTask), pad(t, limCmd, 16384, true))
		requirePrepared(t, r, err)
	})
	t.Run("command one byte over", func(t *testing.T) {
		r, err := call(t, []byte(limTask), pad(t, limCmd, 16385, false))
		requireRejected(t, r, err, "command", "strict_json")
	})
	t.Run("both one byte over reports task", func(t *testing.T) {
		r, err := call(t, pad(t, limTask, 262145, false), pad(t, limCmd, 16385, false))
		requireRejected(t, r, err, "task", "strict_json")
	})
	t.Run("oversized command with oversized-but-legal task still reports command", func(t *testing.T) {
		r, err := call(t, pad(t, limTask, 262144, false), pad(t, limCmd, 16385, false))
		requireRejected(t, r, err, "command", "strict_json")
	})
	t.Run("task at limit by content: passes size check, reducer rejects long outcome", func(t *testing.T) {
		pre := `{"id":"task-1","state":"Draft","revision":"0","outcome":"`
		post := `","acceptance":["Verified"],"owner":"owner-1"}`
		in := pre + strings.Repeat("a", 262144-len(pre)-len(post)) + post
		if len(in) != 262144 {
			t.Fatal(len(in))
		}
		r, err := call(t, []byte(in), []byte(limCmd))
		requireInvalidReducer(t, r, err)
	})
	t.Run("task at limit plus one by content", func(t *testing.T) {
		pre := `{"id":"task-1","state":"Draft","revision":"0","outcome":"`
		post := `","acceptance":["Verified"],"owner":"owner-1"}`
		in := pre + strings.Repeat("a", 262145-len(pre)-len(post)) + post
		r, err := call(t, []byte(in), []byte(limCmd))
		requireRejected(t, r, err, "task", "strict_json")
	})
	t.Run("command at limit by content: passes size check, reducer rejects long actor", func(t *testing.T) {
		pre := `{"type":"prepare","actor":"`
		post := `","expectedRevision":"0","review":null}`
		in := pre + strings.Repeat("a", 16384-len(pre)-len(post)) + post
		if len(in) != 16384 {
			t.Fatal(len(in))
		}
		r, err := call(t, []byte(limTask), []byte(in))
		requireInvalidReducer(t, r, err)
	})
	t.Run("command at limit plus one by content", func(t *testing.T) {
		pre := `{"type":"prepare","actor":"`
		post := `","expectedRevision":"0","review":null}`
		in := pre + strings.Repeat("a", 16385-len(pre)-len(post)) + post
		r, err := call(t, []byte(limTask), []byte(in))
		requireRejected(t, r, err, "command", "strict_json")
	})
}

func timed(t *testing.T, f func()) {
	t.Helper()
	start := time.Now()
	f()
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("took %v", d)
	}
}

func TestDeepNesting(t *testing.T) {
	nest := func(n int) string { return strings.Repeat("[", n) + strings.Repeat("]", n) }
	taskWith := func(acc string) string {
		return `{"id":"task-1","state":"Draft","revision":"0","outcome":"o","acceptance":` + acc + `,"owner":null}`
	}
	cmdWith := func(review string) string {
		return `{"type":"prepare","actor":"owner-1","expectedRevision":"0","review":` + review + `}`
	}
	for _, n := range []int{5, 6, 10, 100, 10000, 100000} {
		n := n
		t.Run(fmt.Sprintf("task acceptance nested arrays %d", n), func(t *testing.T) {
			timed(t, func() {
				r, err := call(t, []byte(taskWith(nest(n))), []byte(limCmd))
				requireRejected(t, r, err, "task", "strict_json")
			})
		})
	}
	t.Run("task unbalanced open brackets 10000", func(t *testing.T) {
		timed(t, func() {
			r, err := call(t, []byte(taskWith(strings.Repeat("[", 10000))), []byte(limCmd))
			requireRejected(t, r, err, "task", "strict_json")
		})
	})
	t.Run("task top level nested arrays 10000", func(t *testing.T) {
		timed(t, func() {
			r, err := call(t, []byte(nest(10000)), []byte(limCmd))
			requireRejected(t, r, err, "task", "strict_json")
		})
	})
	t.Run("task deep objects under unknown key 10000", func(t *testing.T) {
		timed(t, func() {
			deep := strings.Repeat(`{"a":`, 10000) + "1" + strings.Repeat("}", 10000)
			in := `{"id":"task-1","state":"Draft","revision":"0","outcome":"o","acceptance":[],"owner":null,"x":` + deep + `}`
			r, err := call(t, []byte(in), []byte(limCmd))
			requireRejected(t, r, err, "task", "strict_json")
		})
	})
	for _, n := range []int{5, 6, 100, 8000} {
		n := n
		t.Run(fmt.Sprintf("command review nested arrays %d", n), func(t *testing.T) {
			timed(t, func() {
				r, err := call(t, []byte(limTask), []byte(cmdWith(nest(n))))
				requireRejected(t, r, err, "command", "strict_json")
			})
		})
	}
	t.Run("command unbalanced open brackets 8000", func(t *testing.T) {
		timed(t, func() {
			r, err := call(t, []byte(limTask), []byte(cmdWith(strings.Repeat("[", 8000))))
			requireRejected(t, r, err, "command", "strict_json")
		})
	})
	t.Run("both deep: task first", func(t *testing.T) {
		timed(t, func() {
			r, err := call(t, []byte(taskWith(nest(10000))), []byte(cmdWith(nest(5000))))
			requireRejected(t, r, err, "task", "strict_json")
		})
	})
}

// ---------------------------------------------------------------- determinism, immutability, metadata

func TestDeterminismAndNoInputMutation(t *testing.T) {
	tb := loadTable(t)
	pickGroups := map[string]int{}
	for _, c := range tb.Cases {
		if pickGroups[c.Group] >= 6 {
			continue
		}
		pickGroups[c.Group]++
		c := c
		t.Run(c.ID, func(t *testing.T) {
			tj, cj := []byte(expand(t, c.Task)), []byte(expand(t, c.Command))
			tj0, cj0 := append([]byte(nil), tj...), append([]byte(nil), cj...)
			var rs []lifecycle.Result
			var es []error
			for i := 0; i < 3; i++ {
				r, err := rawinput.Transition(tj, cj)
				rs, es = append(rs, r), append(es, err)
				if string(tj) != string(tj0) || string(cj) != string(cj0) {
					t.Fatal("input bytes modified")
				}
			}
			for i := 1; i < 3; i++ {
				if !reflect.DeepEqual(rs[0], rs[i]) || !reflect.DeepEqual(es[0], es[i]) {
					t.Fatalf("call %d differs: %+v %v vs %+v %v", i, rs[0], es[0], rs[i], es[i])
				}
			}
			// A returned task must not alias another call's result.
			if rs[0].Task != nil {
				rs[0].Task.Outcome = "mutated"
				if rs[0].Task.Acceptance != nil && len(rs[0].Task.Acceptance) > 0 {
					rs[0].Task.Acceptance[0] = "mutated"
				}
				if rs[1].Task.Outcome == "mutated" || (len(rs[1].Task.Acceptance) > 0 && rs[1].Task.Acceptance[0] == "mutated") {
					t.Fatal("results from separate calls share memory")
				}
				r4, _ := rawinput.Transition(tj, cj)
				if r4.Task.Outcome == "mutated" {
					t.Fatal("adapter retained shared state across calls")
				}
			}
		})
	}
}

func TestMetadataOnEveryCase(t *testing.T) {
	tb := loadTable(t)
	counts := map[string]int{}
	for _, c := range tb.Cases {
		r, _ := rawinput.Transition([]byte(expand(t, c.Task)), []byte(expand(t, c.Command)))
		if r.Mode != "example-only" || r.Authority != "none" || r.ModelDigest != lifecycle.ModelDigest {
			t.Errorf("%s: metadata %q %q %q", c.ID, r.Mode, r.Authority, r.ModelDigest)
		}
		counts[c.Expect.Kind]++
	}
	for _, k := range []string{"ok", "refused", "invalid", "rejected"} {
		if counts[k] == 0 {
			t.Errorf("no %s cases covered", k)
		}
	}
	// nil and empty inputs as well.
	for _, in := range [][2][]byte{{nil, nil}, {{}, {}}, {[]byte("x"), nil}} {
		r, _ := rawinput.Transition(in[0], in[1])
		checkMeta(t, r)
	}
}

// Success results carry a non-nil empty Refusals; the reducer result is returned unchanged.
func TestSuccessRefusalsNonNilEmpty(t *testing.T) {
	r, err := call(t, []byte(limTask), []byte(limCmd))
	requirePrepared(t, r, err)
	if r.Refusals == nil || len(r.Refusals) != 0 {
		t.Fatalf("Refusals = %#v", r.Refusals)
	}
}

// Null versus empty must survive a refusal: nil acceptance, empty acceptance,
// null owner and present empty owner are four different returned tasks.
func TestNullVersusEmptyPreserved(t *testing.T) {
	mk := func(acc, owner string) string {
		return `{"id":"task-1","state":"Done","revision":"4","outcome":"o","acceptance":` + acc + `,"owner":` + owner + `}`
	}
	cmd := `{"type":"prepare","actor":"owner-1","expectedRevision":"4","review":null}`
	for _, tc := range []struct {
		acc, owner string
		nilAcc     bool
		nilOwner   bool
	}{
		{"null", "null", true, true},
		{"[]", "null", false, true},
		{"null", `""`, true, false},
		{"[]", `""`, false, false},
	} {
		r, err := call(t, []byte(mk(tc.acc, tc.owner)), []byte(cmd))
		if err != nil || r.OK || r.Task == nil || !reflect.DeepEqual(r.Refusals, []string{"ILLEGAL_TRANSITION"}) {
			t.Fatalf("%v %+v", err, r)
		}
		if (r.Task.Acceptance == nil) != tc.nilAcc {
			t.Fatalf("acceptance %s: nil=%v", tc.acc, r.Task.Acceptance == nil)
		}
		if !tc.nilAcc && len(r.Task.Acceptance) != 0 {
			t.Fatalf("acceptance %s not empty", tc.acc)
		}
		if (r.Task.Owner == nil) != tc.nilOwner {
			t.Fatalf("owner %s: nil=%v", tc.owner, r.Task.Owner == nil)
		}
		if !tc.nilOwner && *r.Task.Owner != "" {
			t.Fatalf("owner %s = %q", tc.owner, *r.Task.Owner)
		}
	}
	// Start: null owner is assigned the actor, present empty owner is not absent.
	ready := func(owner string) string {
		return `{"id":"task-1","state":"Ready","revision":"1","outcome":"o","acceptance":["a"],"owner":` + owner + `}`
	}
	start := `{"type":"start","actor":"owner-1","expectedRevision":"1","review":null}`
	r, err := call(t, []byte(ready("null")), []byte(start))
	if err != nil || !r.OK || r.Task == nil || r.Task.Owner == nil || *r.Task.Owner != "owner-1" || r.Task.State != "InProgress" || r.Task.Revision != 2 {
		t.Fatalf("%v %+v", err, r)
	}
	r, err = call(t, []byte(ready(`""`)), []byte(start))
	if err != nil || r.OK || r.Task == nil || r.Task.Owner == nil || *r.Task.Owner != "" || r.Task.State != "Ready" || r.Task.Revision != 1 ||
		!reflect.DeepEqual(r.Refusals, []string{"G_ownerCompatible"}) {
		t.Fatalf("%v %+v", err, r)
	}
}
