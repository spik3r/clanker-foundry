package canonical_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"os"
	"slices"
	"testing"
)

func fixture(t *testing.T) model.Model {
	t.Helper()
	b, e := os.ReadFile("../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	m, e := jsonformat.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func TestSemanticSetInvarianceAndChange(t *testing.T) {
	m := fixture(t)
	original, e := canonical.Bytes(m)
	if e != nil {
		t.Fatal(e)
	}
	slices.Reverse(m.States)
	slices.Reverse(m.Requirements)
	slices.Reverse(m.Guards)
	slices.Reverse(m.Transitions)
	for i := range m.Transitions {
		slices.Reverse(m.Transitions[i].Guards)
	}
	var reverse func(*model.Expr)
	reverse = func(e *model.Expr) {
		if e.Op == "all" || e.Op == "any" {
			slices.Reverse(e.Args)
		}
		for i := range e.Args {
			reverse(&e.Args[i])
		}
	}
	for i := range m.Guards {
		reverse(&m.Guards[i].Expr)
	}
	b, e := canonical.Bytes(m)
	if e != nil || !bytes.Equal(original, b) {
		t.Fatalf("set order changed canonical bytes: %v", e)
	}
	if bytes.HasSuffix(b, []byte("\n")) {
		t.Fatal("canonical newline")
	}
	sum := sha256.Sum256(b)
	d, e := canonical.Digest(m)
	if e != nil || d != hex.EncodeToString(sum[:]) {
		t.Fatal("digest mismatch")
	}
	for _, mutate := range []func(*model.Model){func(m *model.Model) { m.States[0].Label += " edited" }, func(m *model.Model) { m.Requirements[0].Text += " edited" }, func(m *model.Model) { m.Transitions[0].Trigger += "Changed" }} {
		x := fixture(t)
		mutate(&x)
		b, e := canonical.Bytes(x)
		if e != nil || bytes.Equal(original, b) {
			t.Fatalf("semantic/editorial change lost: %v", e)
		}
	}
}
func TestEscapingAndInvalidIR(t *testing.T) {
	m := fixture(t)
	m.States[0].Label = "<>&日本語\u2028"
	b, e := canonical.Bytes(m)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(b, []byte("<>&日本語")) || !bytes.Contains(b, []byte(`\u2028`)) {
		t.Fatal("incorrect escaping")
	}
	m.Version = 99
	if b, e := canonical.Bytes(m); e == nil || b != nil {
		t.Fatal("invalid IR encoded")
	}
	if d, e := canonical.Digest(m); e == nil || d != "" {
		t.Fatal("invalid IR digested")
	}
}
