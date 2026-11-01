package impact

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func tinyModel() model.Model {
	return model.Model{Version: 1, ID: "Example", Initial: "Draft", States: []model.State{{ID: "Draft", Label: "Draft"}, {ID: "Done", Label: "Done", Terminal: true}}, Requirements: []model.Requirement{{ID: "R", Text: "Actor supplied"}}, Guards: []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "command.actor"}}}}}, Transitions: []model.Transition{{ID: "T", From: "Draft", Trigger: "finish", To: "Done", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
}

func wire(t *testing.T, before, after model.Model, bindings []Binding) []byte {
	t.Helper()
	b, err := jsonformat.Encode(before)
	if err != nil {
		t.Fatal(err)
	}
	a, err := jsonformat.Encode(after)
	if err != nil {
		t.Fatal(err)
	}
	if bindings == nil {
		bindings = []Binding{}
	}
	raw, err := json.Marshal(requestDTO{FormatVersion, b, a, bindings})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDecodeRequestStrict(t *testing.T) {
	valid := wire(t, tinyModel(), tinyModel(), nil)
	if _, err := decodeRequest(valid); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{
		"empty": nil, "null": []byte("null"), "trailing": append(append([]byte{}, valid...), []byte(" {}")...),
		"unknown":            bytes.Replace(valid, []byte(`"bindings":`), []byte(`"extra":1,"bindings":`), 1),
		"case variant":       bytes.Replace(valid, []byte(`"bindings":`), []byte(`"Bindings":`), 1),
		"missing":            bytes.Replace(valid, []byte(`,"bindings":[]`), nil, 1),
		"null bindings":      bytes.Replace(valid, []byte(`"bindings":[]`), []byte(`"bindings":null`), 1),
		"duplicate":          bytes.Replace(valid, []byte(`"bindings":[]`), []byte(`"bindings":[],"bindings":[]`), 1),
		"escaped duplicate":  bytes.Replace(valid, []byte(`"bindings":[]`), []byte(`"bindings":[],"binding\u0073":[]`), 1),
		"wrong version":      bytes.Replace(valid, []byte(FormatVersion), []byte("future/v2"), 1),
		"nested duplicate":   bytes.Replace(valid, []byte(`"initial":"Draft"`), []byte(`"initial":"Draft","initial":"Draft"`), 1),
		"nested unknown":     bytes.Replace(valid, []byte(`"initial":"Draft"`), []byte(`"unknown":true,"initial":"Draft"`), 1),
		"nested missing":     bytes.Replace(valid, []byte(`"initial":"Draft",`), nil, 1),
		"nested profile":     bytes.Replace(valid, []byte(model.ProfileVersion), []byte("future/v2"), 1),
		"unpaired surrogate": bytes.Replace(valid, []byte(`"Actor supplied"`), []byte(`"\ud800"`), 1),
		"invalid UTF8":       append([]byte{0xff}, valid...),
		"oversized":          bytes.Repeat([]byte(" "), MaxSourceBytes+1),
		"deep":               []byte(strings.Repeat("[", MaxDepth+2) + strings.Repeat("]", MaxDepth+2)),
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeRequest(raw); err != ErrInvalidInput {
				t.Fatalf("got %v", err)
			}
		})
	}
	var d requestDTO
	if err := json.Unmarshal(valid, &d); err != nil {
		t.Fatal(err)
	}
	d.Before = append(append([]byte("{"), bytes.Repeat([]byte(" "), 1<<20)...), d.Before[1:]...)
	// Avoid Marshal compacting RawMessage; preserve the deliberately oversized value.
	raw := append([]byte(`{"formatVersion":"`+FormatVersion+`","before":`), d.Before...)
	raw = append(raw, []byte(`,"after":`)...)
	raw = append(raw, d.After...)
	raw = append(raw, []byte(`,"bindings":[]}`)...)
	if _, err := decodeRequest(raw); err != ErrInvalidInput {
		t.Fatal("per-model bound was bypassed", err)
	}
}

func TestBindingsAndSourceBytes(t *testing.T) {
	valid := Binding{"scenario", strings.Repeat("a", 64)}
	for name, bindings := range map[string][]Binding{"nil": nil, "one": {valid}, "case distinct": {valid, {"Scenario", valid.ModelDigest}}} {
		t.Run(name, func(t *testing.T) {
			if err := validateBindings(bindings); err != nil {
				t.Fatal(err)
			}
		})
	}
	for name, bindings := range map[string][]Binding{
		"duplicate": {valid, valid}, "count": make([]Binding, MaxBindings+1),
		"empty ID": {{"", valid.ModelDigest}}, "nonASCII": {{"é", valid.ModelDigest}},
		"bad start": {{"1abc", valid.ModelDigest}}, "long ID": {{strings.Repeat("a", 49), valid.ModelDigest}},
		"empty digest": {{"a", ""}}, "long digest": {{"a", strings.Repeat("a", 65)}},
		"uppercase digest": {{"a", strings.Repeat("A", 64)}}, "nonhex": {{"a", strings.Repeat("g", 64)}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateBindings(bindings); err != ErrInvalidInput {
				t.Fatal(err)
			}
		})
	}
	b, _ := jsonformat.Encode(tinyModel())
	raw := []byte(`{"formatVersion":"` + FormatVersion + `","before":` + string(b) + `,"after":` + string(bytes.TrimSpace(b)) + `,"bindings":[]}`)
	r, err := decodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(r.beforeSource, bytes.TrimSpace(b)) || !bytes.Equal(r.beforeSource, r.afterSource) {
		t.Fatal("embedded value bytes include surrounding whitespace")
	}
	if !reflect.DeepEqual(r.before, r.after) {
		t.Fatal("equal models decoded differently")
	}
	for i := range raw {
		raw[i] = 'x'
	}
	if bytes.Contains(r.beforeSource, []byte("xxxx")) {
		t.Fatal("source aliases input buffer")
	}
}
