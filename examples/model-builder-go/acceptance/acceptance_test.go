package acceptance

import (
	"reflect"
	"testing"
)

func owner(s string) *string { return &s }

func assertResult(t *testing.T, got, want Result) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result mismatch\ngot: %#v task=%#v\nwant: %#v task=%#v", got, got.Task, want, want.Task)
	}
}

func requireSubmit(t *testing.T) {
	t.Helper()
	if Submit == nil {
		t.Fatal("no candidate registered: acceptance is frozen, not an implementation")
	}
}

func TestLiteralSubmission(t *testing.T) {
	requireSubmit(t)
	// Positional literals follow the frozen contract. Each expected Result is complete.
	cases := []struct {
		id      string
		task    Task
		command Command
		want    Result
	}{
		{"S01_success", Task{"item-1", "Draft", 0, "Describe the change", []string{"Acceptance stated"}, nil}, Command{"submit", "author-1", 0}, Result{&Task{"item-1", "Ready", 1, "Describe the change", []string{"Acceptance stated"}, nil}, "", nil, "T_submit", "example-only", "none"}},
		{"S02_preserve_order_duplicates_owner", Task{"item-2", "Draft", 7, "  outcome\n", []string{"second", "first", "second"}, owner("different-owner")}, Command{"submit", "actor", 7}, Result{&Task{"item-2", "Ready", 8, "  outcome\n", []string{"second", "first", "second"}, owner("different-owner")}, "", nil, "T_submit", "example-only", "none"}},
		{"S03_present_empty_owner", Task{"i", "Draft", 0, "o", []string{"c"}, owner("")}, Command{"submit", "a", 0}, Result{&Task{"i", "Ready", 1, "o", []string{"c"}, owner("")}, "", nil, "T_submit", "example-only", "none"}},
		{"S04_present_blank_owner", Task{"i", "Draft", 0, "o", []string{"c"}, owner(" \u2003\t")}, Command{"submit", "a", 0}, Result{&Task{"i", "Ready", 1, "o", []string{"c"}, owner(" \u2003\t")}, "", nil, "T_submit", "example-only", "none"}},
		{"S05_unicode_surrounded_text", Task{"i", "Draft", 0, "\u00a0text\u3000", []string{"\u2003criterion\u2003"}, nil}, Command{"submit", "\u2003actor\u0085", 0}, Result{&Task{"i", "Ready", 1, "\u00a0text\u3000", []string{"\u2003criterion\u2003"}, nil}, "", nil, "T_submit", "example-only", "none"}},
		{"S06_zero_width_not_space", Task{"i", "Draft", 0, "\u200b", []string{"c"}, nil}, Command{"submit", "\u200b", 0}, Result{&Task{"i", "Ready", 1, "\u200b", []string{"c"}, nil}, "", nil, "T_submit", "example-only", "none"}},
		{"G01_actor", Task{"i", "Draft", 4, "o", []string{"c"}, nil}, Command{"submit", "", 4}, Result{&Task{"i", "Draft", 4, "o", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_actor"}, "T_submit", "example-only", "none"}},
		{"G02_criteria_nil", Task{"i", "Draft", 4, "o", nil, nil}, Command{"submit", "a", 4}, Result{&Task{"i", "Draft", 4, "o", nil, nil}, "GUARD_FAILED", []string{"G_criteria"}, "T_submit", "example-only", "none"}},
		{"G03_outcome", Task{"i", "Draft", 4, "", []string{"c"}, nil}, Command{"submit", "a", 4}, Result{&Task{"i", "Draft", 4, "", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_outcome"}, "T_submit", "example-only", "none"}},
		{"G04_revision", Task{"i", "Draft", 4, "o", []string{"c"}, nil}, Command{"submit", "a", 3}, Result{&Task{"i", "Draft", 4, "o", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_revision"}, "T_submit", "example-only", "none"}},
		{"G05_actor_criteria", Task{"i", "Draft", 4, "o", nil, nil}, Command{"submit", "", 4}, Result{&Task{"i", "Draft", 4, "o", nil, nil}, "GUARD_FAILED", []string{"G_actor", "G_criteria"}, "T_submit", "example-only", "none"}},
		{"G06_actor_outcome", Task{"i", "Draft", 4, "", []string{"c"}, nil}, Command{"submit", "", 4}, Result{&Task{"i", "Draft", 4, "", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_actor", "G_outcome"}, "T_submit", "example-only", "none"}},
		{"G07_actor_revision", Task{"i", "Draft", 4, "o", []string{"c"}, nil}, Command{"submit", "", 3}, Result{&Task{"i", "Draft", 4, "o", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_actor", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G08_criteria_outcome", Task{"i", "Draft", 4, "", nil, nil}, Command{"submit", "a", 4}, Result{&Task{"i", "Draft", 4, "", nil, nil}, "GUARD_FAILED", []string{"G_criteria", "G_outcome"}, "T_submit", "example-only", "none"}},
		{"G09_criteria_revision", Task{"i", "Draft", 4, "o", nil, nil}, Command{"submit", "a", 3}, Result{&Task{"i", "Draft", 4, "o", nil, nil}, "GUARD_FAILED", []string{"G_criteria", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G10_outcome_revision", Task{"i", "Draft", 4, "", []string{"c"}, nil}, Command{"submit", "a", 3}, Result{&Task{"i", "Draft", 4, "", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_outcome", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G11_actor_criteria_outcome", Task{"i", "Draft", 4, "", nil, nil}, Command{"submit", "", 4}, Result{&Task{"i", "Draft", 4, "", nil, nil}, "GUARD_FAILED", []string{"G_actor", "G_criteria", "G_outcome"}, "T_submit", "example-only", "none"}},
		{"G12_actor_criteria_revision", Task{"i", "Draft", 4, "o", nil, nil}, Command{"submit", "", 3}, Result{&Task{"i", "Draft", 4, "o", nil, nil}, "GUARD_FAILED", []string{"G_actor", "G_criteria", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G13_actor_outcome_revision", Task{"i", "Draft", 4, "", []string{"c"}, nil}, Command{"submit", "", 3}, Result{&Task{"i", "Draft", 4, "", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_actor", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G14_criteria_outcome_revision", Task{"i", "Draft", 4, "", nil, nil}, Command{"submit", "a", 3}, Result{&Task{"i", "Draft", 4, "", nil, nil}, "GUARD_FAILED", []string{"G_criteria", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G15_all", Task{"i", "Draft", 4, "", nil, nil}, Command{"submit", "", 3}, Result{&Task{"i", "Draft", 4, "", nil, nil}, "GUARD_FAILED", []string{"G_actor", "G_criteria", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"}},
		{"G16_criteria_empty_present_owner", Task{"i", "Draft", 4, "o", []string{}, owner("")}, Command{"submit", "a", 4}, Result{&Task{"i", "Draft", 4, "o", []string{}, owner("")}, "GUARD_FAILED", []string{"G_criteria"}, "T_submit", "example-only", "none"}},
		{"G17_unicode_blank", Task{"i", "Draft", 4, "\u00a0\u2003\u3000", []string{"c"}, nil}, Command{"submit", "\u0085\t\n\u202f", 4}, Result{&Task{"i", "Draft", 4, "\u00a0\u2003\u3000", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_actor", "G_outcome"}, "T_submit", "example-only", "none"}},
		{"N01_above_float_exact", Task{"i", "Draft", 9007199254740993, "o", []string{"c"}, nil}, Command{"submit", "a", 9007199254740993}, Result{&Task{"i", "Ready", 9007199254740994, "o", []string{"c"}, nil}, "", nil, "T_submit", "example-only", "none"}},
		{"N02_no_float_rounding", Task{"i", "Draft", 9007199254740993, "o", []string{"c"}, nil}, Command{"submit", "a", 9007199254740992}, Result{&Task{"i", "Draft", 9007199254740993, "o", []string{"c"}, nil}, "GUARD_FAILED", []string{"G_revision"}, "T_submit", "example-only", "none"}},
		{"N03_last_increment", Task{"i", "Draft", 18446744073709551614, "o", []string{"c"}, owner("owner")}, Command{"submit", "a", 18446744073709551614}, Result{&Task{"i", "Ready", 18446744073709551615, "o", []string{"c"}, owner("owner")}, "", nil, "T_submit", "example-only", "none"}},
		{"N04_revision_limit", Task{"i", "Draft", 18446744073709551615, "o", []string{"c"}, owner("owner")}, Command{"submit", "a", 18446744073709551615}, Result{&Task{"i", "Draft", 18446744073709551615, "o", []string{"c"}, owner("owner")}, "REVISION_LIMIT", nil, "T_submit", "example-only", "none"}},
		{"N05_guards_before_limit", Task{"i", "Draft", 18446744073709551615, "", []string{}, owner(" ")}, Command{"submit", "", 18446744073709551614}, Result{&Task{"i", "Draft", 18446744073709551615, "", []string{}, owner(" ")}, "GUARD_FAILED", []string{"G_actor", "G_criteria", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"}},
		{"T01_terminal", Task{"i", "Ready", 1, "o", []string{"c"}, nil}, Command{"submit", "a", 1}, Result{&Task{"i", "Ready", 1, "o", []string{"c"}, nil}, "ILLEGAL_TRANSITION", nil, "", "example-only", "none"}},
		{"T02_terminal_before_guards_limit", Task{"i", "Ready", 18446744073709551615, "", nil, owner("")}, Command{"submit", "", 0}, Result{&Task{"i", "Ready", 18446744073709551615, "", nil, owner("")}, "ILLEGAL_TRANSITION", nil, "", "example-only", "none"}},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) { assertResult(t, Submit(tc.task, tc.command), tc.want) })
	}
}

func TestInputAndSuccessIsolation(t *testing.T) {
	requireSubmit(t)
	input := Task{"isolation", "Draft", 8, "o", []string{"one", "two"}, owner("original")}
	command := Command{"submit", "other-actor", 8}
	first := Submit(input, command)
	if !reflect.DeepEqual(input, Task{"isolation", "Draft", 8, "o", []string{"one", "two"}, owner("original")}) {
		t.Fatal("candidate mutated input")
	}
	if !reflect.DeepEqual(command, Command{"submit", "other-actor", 8}) {
		t.Fatal("candidate mutated command")
	}
	assertResult(t, first, Result{&Task{"isolation", "Ready", 9, "o", []string{"one", "two"}, owner("original")}, "", nil, "T_submit", "example-only", "none"})
	input.Acceptance[0] = "input mutation"
	*input.Owner = "input owner mutation"
	assertResult(t, first, Result{&Task{"isolation", "Ready", 9, "o", []string{"one", "two"}, owner("original")}, "", nil, "T_submit", "example-only", "none"})
	later := Submit(Task{"later", "Draft", 20, "later outcome", []string{"later criterion"}, owner("later owner")}, Command{"submit", "actor", 20})
	assertResult(t, later, Result{&Task{"later", "Ready", 21, "later outcome", []string{"later criterion"}, owner("later owner")}, "", nil, "T_submit", "example-only", "none"})
	later.Task.ID = "mutated"
	later.Task.Acceptance[0] = "later mutation"
	*later.Task.Owner = "later owner mutation"
	assertResult(t, first, Result{&Task{"isolation", "Ready", 9, "o", []string{"one", "two"}, owner("original")}, "", nil, "T_submit", "example-only", "none"})
	first.Task.Acceptance[1] = "output mutation"
	*first.Task.Owner = "output owner mutation"
	if !reflect.DeepEqual(input, Task{"isolation", "Draft", 8, "o", []string{"input mutation", "two"}, owner("input owner mutation")}) {
		t.Fatal("output aliases input")
	}
}

func TestGuardRefusalIsolation(t *testing.T) {
	requireSubmit(t)
	input := Task{"refusal", "Draft", 3, "", []string{"criterion"}, owner("owner")}
	first := Submit(input, Command{"submit", "", 2})
	if !reflect.DeepEqual(input, Task{"refusal", "Draft", 3, "", []string{"criterion"}, owner("owner")}) {
		t.Fatal("refusal mutated input")
	}
	assertResult(t, first, Result{&Task{"refusal", "Draft", 3, "", []string{"criterion"}, owner("owner")}, "GUARD_FAILED", []string{"G_actor", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"})
	input.Acceptance[0] = "input changed"
	*input.Owner = "input changed"
	second := Submit(Task{"refusal", "Draft", 3, "", []string{"criterion"}, owner("owner")}, Command{"submit", "", 2})
	assertResult(t, second, Result{&Task{"refusal", "Draft", 3, "", []string{"criterion"}, owner("owner")}, "GUARD_FAILED", []string{"G_actor", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"})
	second.Task.Acceptance[0] = "later changed"
	*second.Task.Owner = "later changed"
	second.FailedGuards[0] = "later changed"
	assertResult(t, first, Result{&Task{"refusal", "Draft", 3, "", []string{"criterion"}, owner("owner")}, "GUARD_FAILED", []string{"G_actor", "G_outcome", "G_revision"}, "T_submit", "example-only", "none"})
	first.Task.Acceptance[0] = "output changed"
	*first.Task.Owner = "output changed"
	if !reflect.DeepEqual(input, Task{"refusal", "Draft", 3, "", []string{"input changed"}, owner("input changed")}) {
		t.Fatal("refusal output aliases input")
	}
}

func TestTerminalAndLimitIsolation(t *testing.T) {
	requireSubmit(t)
	terminalInput := Task{"terminal", "Ready", 10, "o", []string{"c"}, owner("owner")}
	terminal := Submit(terminalInput, Command{"submit", "actor", 10})
	limitInput := Task{"limit", "Draft", 18446744073709551615, "o", []string{"c"}, owner("owner")}
	limit := Submit(limitInput, Command{"submit", "actor", 18446744073709551615})
	if !reflect.DeepEqual(terminalInput, Task{"terminal", "Ready", 10, "o", []string{"c"}, owner("owner")}) {
		t.Fatal("terminal mutated input")
	}
	if !reflect.DeepEqual(limitInput, Task{"limit", "Draft", 18446744073709551615, "o", []string{"c"}, owner("owner")}) {
		t.Fatal("limit mutated input")
	}
	terminalInput.Acceptance[0] = "changed"
	*terminalInput.Owner = "changed"
	limitInput.Acceptance[0] = "changed"
	*limitInput.Owner = "changed"
	assertResult(t, terminal, Result{&Task{"terminal", "Ready", 10, "o", []string{"c"}, owner("owner")}, "ILLEGAL_TRANSITION", nil, "", "example-only", "none"})
	assertResult(t, limit, Result{&Task{"limit", "Draft", 18446744073709551615, "o", []string{"c"}, owner("owner")}, "REVISION_LIMIT", nil, "T_submit", "example-only", "none"})
	limit.Task.Acceptance[0] = "output changed"
	*limit.Task.Owner = "output changed"
	terminal.Task.Acceptance[0] = "output changed"
	*terminal.Task.Owner = "output changed"
	if !reflect.DeepEqual(terminalInput, Task{"terminal", "Ready", 10, "o", []string{"changed"}, owner("changed")}) {
		t.Fatal("terminal output aliases input")
	}
	if !reflect.DeepEqual(limitInput, Task{"limit", "Draft", 18446744073709551615, "o", []string{"changed"}, owner("changed")}) {
		t.Fatal("limit output aliases input")
	}
}
