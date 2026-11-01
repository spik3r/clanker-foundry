package rawinput

import (
	"reflect"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

const (
	goodTask    = `{"id":"task-1","state":"Draft","revision":"0","outcome":"Ship outcome","acceptance":["Verified"],"owner":null}`
	goodCommand = `{"type":"prepare","actor":"owner-1","expectedRevision":"0","review":null}`
)

// countCalls replaces the reducer with a counting wrapper for one test.
func countCalls(t *testing.T) *int {
	t.Helper()
	n := 0
	orig := transition
	transition = func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		n++
		return orig(task, c)
	}
	t.Cleanup(func() { transition = orig })
	return &n
}

func TestRejectedInputNeverCallsReducer(t *testing.T) {
	cases := []struct{ name, task, command string }{
		{"malformed task", `not json`, goodCommand},
		{"malformed command", goodTask, `not json`},
		{"trailing task", goodTask + `{}`, goodCommand},
		{"duplicate command key", goodTask, `{"type":"prepare","type":"start","actor":"a","expectedRevision":"0","review":null}`},
		{"unknown task key", `{"id":"t","state":"Draft","revision":"0","outcome":"o","acceptance":null,"owner":null,"x":1}`, goodCommand},
		{"missing review key", goodTask, `{"type":"prepare","actor":"a","expectedRevision":"0"}`},
		{"task revision overflow", `{"id":"t","state":"Draft","revision":"18446744073709551616","outcome":"o","acceptance":null,"owner":null}`, goodCommand},
		{"review revision invalid", goodTask, `{"type":"approve","actor":"a","expectedRevision":"0","review":{"taskId":"t","taskRevision":"01","modelDigest":"","taskDigest":"","reviewer":"r","decision":"approved"}}`},
		{"byte order mark", "\xef\xbb\xbf" + goodTask, goodCommand},
		{"escaped duplicate command key", goodTask, `{"type":"prepare","\u0074ype":"start","actor":"a","expectedRevision":"0","review":null}`},
		{"lone surrogate in key", goodTask, `{"\ud800":"x","type":"prepare","actor":"a","expectedRevision":"0","review":null}`},
		{"lone surrogate", `{"id":"\ud800","state":"Draft","revision":"0","outcome":"o","acceptance":null,"owner":null}`, goodCommand},
		{"invalid UTF-8", "{\"id\":\"\xff\",\"state\":\"Draft\",\"revision\":\"0\",\"outcome\":\"o\",\"acceptance\":null,\"owner\":null}", goodCommand},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := countCalls(t)
			r, err := Transition([]byte(c.task), []byte(c.command))
			if err == nil || *n != 0 {
				t.Fatalf("err=%v calls=%d", err, *n)
			}
			if !reflect.DeepEqual(r, rejected()) {
				t.Fatalf("unexpected result %+v", r)
			}
		})
	}
}

func TestRejectedShapeMatchesReducerInvalidInput(t *testing.T) {
	want := lifecycle.Transition(lifecycle.Task{}, lifecycle.Command{}) // blank task ID is typed-invalid
	if want.Task != nil || len(want.Refusals) != 1 || want.Refusals[0] != "INVALID_INPUT" {
		t.Fatalf("reducer no longer returns the expected invalid shape: %+v", want)
	}
	if got := rejected(); !reflect.DeepEqual(got, want) {
		t.Fatalf("rejected() drifted from the reducer\n got %+v\nwant %+v", got, want)
	}
}

func TestTaskIsExaminedBeforeCommand(t *testing.T) {
	_, err := Transition([]byte(`x`), []byte(`y`))
	if e, ok := err.(*InputError); !ok || e.Input != "task" || e.Reason != ReasonStrictJSON {
		t.Fatalf("got %#v", err)
	}
}

func TestStructurallyValidRequestPassesThroughReducerResult(t *testing.T) {
	n := countCalls(t)
	got, err := Transition([]byte(goodTask), []byte(goodCommand))
	if err != nil || *n != 1 {
		t.Fatalf("err=%v calls=%d", err, *n)
	}
	want := lifecycle.Transition(lifecycle.Task{ID: "task-1", State: "Draft", Outcome: "Ship outcome", Acceptance: []string{"Verified"}, Owner: nil},
		lifecycle.Command{Type: "prepare", Actor: "owner-1"})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("adapter result differs from direct reducer call\n got %+v\nwant %+v", got, want)
	}
}

func TestSemanticallyInvalidIsLeftToReducer(t *testing.T) {
	n := countCalls(t)
	task := `{"id":"task-1","state":"Bogus","revision":"0","outcome":"o","acceptance":null,"owner":null}`
	r, err := Transition([]byte(task), []byte(goodCommand))
	if err != nil || *n != 1 || r.Task != nil || r.Refusals[0] != "INVALID_INPUT" {
		t.Fatalf("err=%v calls=%d r=%+v", err, *n, r)
	}
}

func TestParseRevision(t *testing.T) {
	ok := map[string]uint64{"0": 0, "7": 7, "18446744073709551615": 18446744073709551615}
	for s, want := range ok {
		if got, good := parseRevision(s); !good || got != want {
			t.Errorf("%q: got %d %v", s, got, good)
		}
	}
	for _, s := range []string{"", "00", "01", "-0", "+1", " 1", "1 ", "1.0", "1e3", "0x1", "١", "18446744073709551616"} {
		if _, good := parseRevision(s); good {
			t.Errorf("%q accepted", s)
		}
	}
}
