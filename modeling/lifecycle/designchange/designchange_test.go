package designchange

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
)

const requestHash = "f3a14ba6adfd54832d810e26dccfc929e5fa61801990fa7c52436ec51eaacde0"
const suiteHash = "53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1"
const expectedHash = "efc7cc5088e1a79210303fe410af6c74831ceab9a77d1f4075efc4fb338c47a3"

type fixtureLink struct {
	Side         string   `json:"side"`
	TransitionID string   `json:"transitionId"`
	Causes       []string `json:"causes"`
}
type fixtureExpected struct {
	External      Expected               `json:"externalExpected"`
	RawCandidate  string                 `json:"rawCandidateReferenceSHA256"`
	Names         []string               `json:"outerInventory"`
	Changes       []impact.Change        `json:"changes"`
	Links         []fixtureLink          `json:"linkedTransitions"`
	Bindings      []impact.BindingResult `json:"bindings"`
	SuiteStatus   string                 `json:"suiteStatus"`
	BindingStatus string                 `json:"suiteBindingStatus"`
}

func fixture(t *testing.T, n, h string) []byte {
	t.Helper()
	b, e := os.ReadFile(filepath.Join("testdata", n))
	if e != nil || hash(b) != h {
		t.Fatalf("frozen %s: %v", n, e)
	}
	return b
}
func inputs(t *testing.T) ([]byte, []byte, []artifactbundle.File, fixtureExpected) {
	t.Helper()
	request := fixture(t, "request.json", requestHash)
	suite := fixture(t, "suite.json", suiteHash)
	var want fixtureExpected
	if e := json.Unmarshal(fixture(t, "expected.json", expectedHash), &want); e != nil {
		t.Fatal(e)
	}
	src := []artifactbundle.File{}
	for _, n := range qualification.SourceNames() {
		b, e := os.ReadFile(filepath.Join("..", filepath.FromSlash(n)))
		if e != nil {
			t.Fatal(e)
		}
		src = append(src, artifactbundle.File{Name: n, Data: b})
	}
	return request, suite, src, want
}
func payloads(t *testing.T, b []byte, p string) []artifactbundle.File {
	t.Helper()
	f, e := artifactbundle.Verify(b, p)
	if e != nil {
		t.Fatal(e)
	}
	return f
}
func byName(f []artifactbundle.File) map[string][]byte {
	m := map[string][]byte{}
	for _, x := range f {
		m[x.Name] = x.Data
	}
	return m
}
func pack(t *testing.T, f []artifactbundle.File, p string) []byte {
	t.Helper()
	b, e := artifactbundle.Build(p, f)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func changeFile(f []artifactbundle.File, n string, b []byte) []artifactbundle.File {
	out := append([]artifactbundle.File{}, f...)
	for i := range out {
		if out[i].Name == n {
			out[i].Data = b
		}
	}
	return out
}
func buildCarrier(r, s []byte, src []artifactbundle.File) []byte {
	a, b := base64.StdEncoding.EncodeToString(r), base64.StdEncoding.EncodeToString(s)
	ss := []sourceDTO{}
	for _, f := range src {
		ss = append(ss, sourceDTO{f.Name, base64.StdEncoding.EncodeToString(f.Data)})
	}
	out, _ := encode(carrierDTO{FormatVersion: FormatVersion, Operation: "build", RequestBase64: &a, SuiteBase64: &b, Sources: &ss})
	return out
}
func verifyCarrier(a []byte, e Expected) []byte {
	s := base64.StdEncoding.EncodeToString(a)
	b, _ := encode(carrierDTO{FormatVersion: FormatVersion, Operation: "verify", ArchiveBase64: &s, Expected: &e})
	return b
}
func requireError(t *testing.T, b []byte, e, want error) {
	t.Helper()
	if !errors.Is(e, want) || b != nil {
		t.Fatalf("expected nil/%v got %d/%v", want, len(b), e)
	}
}

func TestFrozenChangeProject(t *testing.T) {
	r, s, src, w := inputs(t)
	raw, e := Build(r, s, src)
	if e != nil {
		t.Fatal(e)
	}
	if Profile != artifactbundle.DesignChangeProfile {
		t.Fatal("profile mismatch")
	}
	if e = Verify(raw, w.External); e != nil {
		t.Fatal(e)
	}
	files := payloads(t, raw, Profile)
	names := []string{}
	for _, f := range files {
		names = append(names, f.Name)
	}
	if !reflect.DeepEqual(names, w.Names) {
		t.Fatalf("inventory %v", names)
	}
	outer := byName(files)
	if !bytes.Equal(outer["change-request.json"], r) {
		t.Fatal("request rewritten")
	}
	if hash(outer["before.json"]) != w.External.BeforeSourceSHA256 || hash(outer["candidate.json"]) != w.External.CandidateSHA256 {
		t.Fatal("raw token or normalized identity lost")
	}
	if hash(outer["candidate.json"]) == w.RawCandidate {
		t.Fatal("normalized/raw identities collapsed")
	}
	var meta metadataDTO
	if json.Unmarshal(outer["design-change.json"], &meta) != nil {
		t.Fatal("metadata")
	}
	if meta.RequestSHA256 != w.External.RequestSHA256 || meta.BeforeSourceSHA256 != w.External.BeforeSourceSHA256 || meta.BeforeModelDigest != w.External.BeforeModelDigest || meta.CandidateSHA256 != w.External.CandidateSHA256 || meta.CandidateModelDigest != w.External.CandidateModelDigest || meta.SuiteSHA256 != w.External.SuiteSHA256 || meta.SuiteStatus != w.SuiteStatus || meta.SuiteBindingStatus != w.BindingStatus {
		t.Fatalf("metadata identities/status %+v", meta)
	}
	var report struct {
		Review struct {
			Changes  []impact.Change           `json:"changes"`
			Links    []impact.LinkedTransition `json:"linkedTransitions"`
			Bindings []impact.BindingResult    `json:"bindings"`
		} `json:"review"`
	}
	if json.Unmarshal(outer["planner-result.json"], &report) != nil {
		t.Fatal("report")
	}
	// Fixture transitionId is descriptive; actual ABI field is ID. Never rewrite
	// the independently frozen fixture or generate expected links from Plan output.
	links := []impact.LinkedTransition{}
	for _, l := range w.Links {
		links = append(links, impact.LinkedTransition{Side: l.Side, ID: l.TransitionID, Causes: l.Causes})
	}
	if !reflect.DeepEqual(report.Review.Changes, w.Changes) || !reflect.DeepEqual(report.Review.Links, links) || !reflect.DeepEqual(report.Review.Bindings, w.Bindings) {
		t.Fatalf("unexpected declared impact %+v", report)
	}
	inner := byName(payloads(t, outer["qualification.zip"], artifactbundle.QualificationProfile))
	if !bytes.Equal(inner["suite.json"], s) || !bytes.Equal(inner["candidate.json"], outer["candidate.json"]) {
		t.Fatal("inner substitution")
	}
	for _, term := range []string{"unvalidated and unexecuted", "not inspected", "may be private"} {
		if !strings.Contains(string(outer["GUIDE.txt"]), term) {
			t.Fatalf("missing honesty %s", term)
		}
	}
	via, e := Execute(buildCarrier(r, s, src))
	if e != nil || !bytes.Equal(raw, via) {
		t.Fatalf("wire build %v", e)
	}
	receipt, e := Execute(verifyCarrier(raw, w.External))
	if e != nil {
		t.Fatal(e)
	}
	want := `{"formatVersion":"lifecycle-designchange-verification/v1","consistent":true,"mode":"example-only","authority":"none","suiteStatus":"unvalidated and unexecuted","suiteBindingStatus":"not inspected"}` + "\n"
	if string(receipt) != want {
		t.Fatalf("receipt %s", receipt)
	}
}

func TestAllExternalIdentitiesRequired(t *testing.T) {
	r, s, src, w := inputs(t)
	raw, e := Build(r, s, src)
	if e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"RequestSHA256", "BeforeSourceSHA256", "BeforeModelDigest", "CandidateSHA256", "CandidateModelDigest", "SuiteSHA256"} {
		t.Run(field, func(t *testing.T) {
			x := w.External
			reflect.ValueOf(&x).Elem().FieldByName(field).SetString(strings.Repeat("0", 64))
			if e := Verify(raw, x); !errors.Is(e, ErrInvalid) {
				t.Fatalf("external %s accepted %v", field, e)
			}
			reflect.ValueOf(&x).Elem().FieldByName(field).SetString("")
			if e := Verify(raw, x); !errors.Is(e, ErrInvalid) {
				t.Fatal("missing accepted")
			}
		})
	}
}
func TestOpaqueSuiteAndHostileContents(t *testing.T) {
	r, _, src, w := inputs(t)
	for _, suite := range [][]byte{[]byte("not JSON\x00private <script> '&\""), []byte(`{"modelDigest":"wrong","cases":[{"later":`), []byte(`{"modelDigest":"different","cases":[]}`)} {
		raw, e := Build(r, suite, src)
		if e != nil {
			t.Fatalf("opaque suite rejected %v", e)
		}
		x := w.External
		x.SuiteSHA256 = hash(suite)
		if e = Verify(raw, x); e != nil {
			t.Fatal(e)
		}
		inner := byName(payloads(t, byName(payloads(t, raw, Profile))["qualification.zip"], artifactbundle.QualificationProfile))
		if !bytes.Equal(inner["suite.json"], suite) {
			t.Fatal("suite repaired")
		}
		if e = Verify(raw, w.External); !errors.Is(e, ErrInvalid) {
			t.Fatal("external suite hash ignored")
		}
	}
	// Literal hostile caller text is preserved without being labelled public-safe.
	var request map[string]any
	if json.Unmarshal(r, &request) != nil {
		t.Fatal("request")
	}
	label := "Invented private <script> '&\" 雪"
	ops := request["operations"].([]any)
	ops = append(ops, map[string]any{"action": "replace", "kind": "state", "id": "Draft", "root": nil, "state": map[string]any{"id": "Draft", "label": label, "terminal": false}, "requirement": nil, "guard": nil, "transition": nil})
	request["operations"] = ops
	changed, _ := json.Marshal(request)
	raw, e := Build(changed, []byte("opaque"), src)
	if e != nil {
		t.Fatal(e)
	}
	outer := byName(payloads(t, raw, Profile))
	m, e := jsonformat.Decode(outer["candidate.json"])
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, state := range m.States {
		if state.ID == "Draft" {
			found = state.Label == label
		}
	}
	if !found {
		t.Fatal("hostile label changed")
	}
	if strings.Contains(string(outer["GUIDE.txt"]), "Public-safe") || !bytes.Equal(outer["change-request.json"], changed) {
		t.Fatal("provenance overclaim or rewriting")
	}

}
func TestRequestSemanticsAndResourcePrecedence(t *testing.T) {
	r, s, src, _ := inputs(t)
	stale := bytes.Replace(r, []byte("948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed"), []byte(strings.Repeat("0", 64)), 1)
	b, e := Build(stale, s, src)
	requireError(t, b, e, modelpatch.ErrPrecondition)
	broken := bytes.Replace(r, []byte(`"to":"Ready"`), []byte(`"to":"Missing"`), 1)
	if bytes.Equal(broken, r) {
		t.Fatal("mutation did not apply")
	}
	b, e = Build(broken, s, src)
	requireError(t, b, e, modelpatch.ErrInvalid)
	corrupt := append([]artifactbundle.File{}, src...)
	corrupt[0].Data = []byte("not pinned")
	b, e = Build(r, s, corrupt)
	requireError(t, b, e, ErrInvalid)
	b, e = Build(stale, s, corrupt)
	requireError(t, b, e, modelpatch.ErrPrecondition)
	for _, n := range []int{0, len(src) - 1, len(src) + 1} {
		ss := append([]artifactbundle.File{}, src...)
		if n > len(src) {
			ss = append(ss, src[0])
		} else {
			ss = ss[:n]
		}
		b, e = Build(stale, s, ss)
		want := ErrInvalid
		if n > len(src) {
			want = ErrLimit
		}
		requireError(t, b, e, want)
	}
	bad := append([]artifactbundle.File{}, src...)
	bad[1].Name = bad[0].Name
	b, e = Build(stale, s, bad)
	requireError(t, b, e, ErrInvalid)
	bad = append([]artifactbundle.File{}, src...)
	bad[0].Name = "../source"
	b, e = Build(stale, s, bad)
	requireError(t, b, e, ErrInvalid)
	bad = append([]artifactbundle.File{}, src...)
	bad[0].Data = make([]byte, MaxSourceBytes+1)
	b, e = Build(stale, s, bad)
	requireError(t, b, e, ErrLimit)
	b, e = Execute(buildCarrier(stale, s, bad))
	requireError(t, b, e, ErrLimit)
	bad = append([]artifactbundle.File{}, src...)
	for i := 0; i < 5; i++ {
		bad[i].Data = make([]byte, MaxSourceBytes)
	}
	b, e = Build(stale, s, bad)
	requireError(t, b, e, ErrLimit)
	b, e = Execute(buildCarrier(stale, s, bad))
	requireError(t, b, e, ErrLimit)
	b, e = Build(make([]byte, MaxRequestBytes+1), s, src)
	requireError(t, b, e, ErrLimit)
	b, e = Build(r, nil, src)
	requireError(t, b, e, ErrInvalid)
	b, e = Build(r, make([]byte, MaxSuiteBytes+1), src)
	requireError(t, b, e, ErrLimit)
	// At the generic source cap, pin mismatch remains qualifier ErrInvalid after
	// a valid request, not a planner-owned limit.
	bad = append([]artifactbundle.File{}, src...)
	bad[0].Data = make([]byte, MaxSourceBytes)
	b, e = Build(r, s, bad)
	requireError(t, b, e, ErrInvalid)
}

func TestDeterministicOwnedReplay(t *testing.T) {
	r, s, src, w := inputs(t)
	r0 := append([]byte{}, r...)
	s0 := append([]byte{}, s...)
	saved := make([][]byte, len(src))
	for i := range src {
		saved[i] = append([]byte{}, src[i].Data...)
	}
	a, e := Build(r, s, src)
	if e != nil {
		t.Fatal(e)
	}
	sort.Slice(src, func(i, j int) bool { return src[i].Name > src[j].Name })
	b, e := Build(r, s, src)
	if e != nil || !bytes.Equal(a, b) {
		t.Fatalf("permutation replay %v", e)
	}
	if !bytes.Equal(r, r0) || !bytes.Equal(s, s0) {
		t.Fatal("inputs changed")
	}
	for _, f := range src {
		for i, n := range qualification.SourceNames() {
			if f.Name == n && !bytes.Equal(f.Data, saved[i]) {
				t.Fatal("source changed")
			}
		}
	}
	a[0] ^= 1
	if e = Verify(b, w.External); e != nil {
		t.Fatal("returned buffers alias")
	}
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			x, e := Build(r, s, src)
			if e != nil || !bytes.Equal(x, b) {
				t.Errorf("concurrent replay %v", e)
			}
		}()
	}
	wg.Wait()
}
func TestCanonicalTamperAndLateInnerFailure(t *testing.T) {
	r, s, src, w := inputs(t)
	raw, e := Build(r, s, src)
	if e != nil {
		t.Fatal(e)
	}
	outer := payloads(t, raw, Profile)
	for _, name := range w.Names {
		t.Run(name, func(t *testing.T) {
			old := byName(outer)[name]
			changed := append(append([]byte{}, old...), byte(' '))
			b := pack(t, changeFile(outer, name, changed), Profile)
			if e := Verify(b, w.External); e == nil {
				t.Fatal("self-consistent substitution accepted")
			}
		})
	}
	for _, f := range [][]artifactbundle.File{outer[:len(outer)-1], append(append([]artifactbundle.File{}, outer...), artifactbundle.File{Name: "extra.zip", Data: []byte("nested")})} {
		if e = Verify(pack(t, f, Profile), w.External); !errors.Is(e, ErrInvalid) {
			t.Fatal("outer inventory accepted")
		}
	}
	inner := payloads(t, byName(outer)["qualification.zip"], artifactbundle.QualificationProfile)
	for _, name := range []string{"generated/lifecycle.go", "generated/model_paths_test.go", "suite.json", "qualification.json", "strictjson/decode.go"} {
		t.Run("inner/"+name, func(t *testing.T) {
			old := byName(inner)[name]
			mut := changeFile(inner, name, append(append([]byte{}, old...), byte(' ')))
			b := pack(t, changeFile(outer, "qualification.zip", pack(t, mut, artifactbundle.QualificationProfile)), Profile)
			if Verify(b, w.External) == nil {
				t.Fatal("late inner tamper accepted")
			}
		})
	}
	extra := append(append([]artifactbundle.File{}, inner...), artifactbundle.File{Name: "nested.zip", Data: []byte("not interpreted")})
	if Verify(pack(t, changeFile(outer, "qualification.zip", pack(t, extra, artifactbundle.QualificationProfile)), Profile), w.External) == nil {
		t.Fatal("extra nesting accepted")
	}
	for _, b := range [][]byte{raw[:len(raw)-1], append(append([]byte{}, raw...), 0), []byte("not zip")} {
		if Verify(b, w.External) == nil {
			t.Fatal("noncanonical accepted")
		}
	}
	if e = Verify(make([]byte, MaxArchiveBytes+1), w.External); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	// Exact resource failures do not become malformed-archive errors.
	huge := pack(t, []artifactbundle.File{{Name: "a", Data: make([]byte, MaxInnerBytes)}}, artifactbundle.QualificationProfile)
	if len(huge) <= MaxInnerBytes {
		t.Fatal("limit fixture")
	}
	f := changeFile(outer, "qualification.zip", huge)
	if _, e = artifactbundle.Build(Profile, f); !errors.Is(e, artifactbundle.ErrLimit) {
		t.Fatal("shared limit not reached")
	}
	if !errors.Is(archiveError(e), ErrLimit) {
		t.Fatal("limit mapping")
	}
}

func TestStrictCarrier(t *testing.T) {
	r, s, src, _ := inputs(t)
	valid := buildCarrier(r, s, src)
	var fields map[string]json.RawMessage
	if json.Unmarshal(valid, &fields) != nil {
		t.Fatal("carrier")
	}
	for k := range fields {
		copy := map[string]json.RawMessage{}
		for x, v := range fields {
			copy[x] = v
		}
		delete(copy, k)
		b, _ := json.Marshal(copy)
		out, e := Execute(b)
		requireError(t, out, e, ErrInvalid)
	}
	for _, b := range [][]byte{append(valid, []byte("{}")...), bytes.Replace(valid, []byte(`"operation":"build"`), []byte(`"operation":"build","operation":"build"`), 1), bytes.Replace(valid, []byte(`"operation"`), []byte(`"Operation"`), 1), bytes.Replace(valid, []byte(`"expected":null`), []byte(`"expected":{}`), 1), bytes.Replace(valid, []byte(`"sources":[`), []byte(`"extra":1,"sources":[`), 1), []byte("null"), []byte("\xff"), []byte(`{"formatVersion":"\ud800"}`)} {
		out, e := Execute(b)
		requireError(t, out, e, ErrInvalid)
	}
	out, e := Execute(make([]byte, MaxCarrierBytes+1))
	requireError(t, out, e, ErrLimit)
	out, e = Execute([]byte(strings.Repeat("[", MaxCarrierDepth+1) + strings.Repeat("]", MaxCarrierDepth+1)))
	requireError(t, out, e, ErrLimit)
	for _, s := range []string{"Zg", "Zg==\n", "Zh==", "Zm9=", "====", "Z===", "Zg-="} {
		if _, e := encodedSize(s, 32); e == nil {
			t.Fatalf("noncanonical base64 %q", s)
		}
	}
	for n := 0; n < 20; n++ {
		s := base64.StdEncoding.EncodeToString(make([]byte, n))
		got, e := encodedSize(s, n)
		if e != nil || got != n {
			t.Fatal("boundary", n, e)
		}
		if n > 0 {
			if _, e = encodedSize(s, n-1); !errors.Is(e, ErrLimit) {
				t.Fatal("decoded cap", n, e)
			}
		}
	}
	if _, e = encodedSize(strings.Repeat("A", 48), 1); !errors.Is(e, ErrLimit) {
		t.Fatal("encoded cap")
	}
}

func TestRawTokenWhitespaceAndNoopDistinctions(t *testing.T) {
	r, s, src, w := inputs(t)
	// Surrounding token whitespace belongs to the carrier, not BeforeSourceSHA256.
	surrounding := bytes.Replace(r, []byte(`"before":{`), []byte("\"before\": \n {"), 1)
	raw, e := Build(surrounding, s, src)
	if e != nil {
		t.Fatal(e)
	}
	x := w.External
	x.RequestSHA256 = hash(surrounding)
	if e = Verify(raw, x); e != nil {
		t.Fatal("carrier whitespace changed token identity", e)
	}
	if Verify(raw, w.External) == nil {
		t.Fatal("raw request identity ignored")
	}
	// Interior whitespace is part of the token, despite identical semantics.
	interior := bytes.Replace(r, []byte("{\n  \"formatVersion\": \"lifecycle-json/v1\""), []byte("{\n   \"formatVersion\": \"lifecycle-json/v1\""), 1)
	if bytes.Equal(interior, r) {
		t.Fatal("literal mutation failed")
	}
	// Independently extract only the source token to form caller-held raw identity.
	var carrier struct {
		Before json.RawMessage `json:"before"`
	}
	if json.Unmarshal(interior, &carrier) != nil {
		t.Fatal("literal source")
	}
	x = w.External
	x.RequestSHA256 = hash(interior)
	x.BeforeSourceSHA256 = hash(carrier.Before)
	raw, e = Build(interior, s, src)
	if e != nil || Verify(raw, x) != nil {
		t.Fatal("interior whitespace", e)
	}
	if Verify(raw, w.External) == nil {
		t.Fatal("raw token identity ignored")
	}
	// Explicit no-op keeps baseline semantics but makes no assertion about the
	// opaque candidate-bound suite. Binding remains not inspected.
	var obj map[string]json.RawMessage
	if json.Unmarshal(r, &obj) != nil {
		t.Fatal("source")
	}
	obj["operations"] = json.RawMessage("[]")
	noop, _ := json.Marshal(obj)
	raw, e = Build(noop, s, src)
	if e != nil {
		t.Fatal(e)
	}
	outer := byName(payloads(t, raw, Profile))
	var meta metadataDTO
	if json.Unmarshal(outer["design-change.json"], &meta) != nil {
		t.Fatal("metadata")
	}
	if meta.BeforeModelDigest != "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed" || meta.CandidateModelDigest != meta.BeforeModelDigest || meta.SuiteBindingStatus != "not inspected" {
		t.Fatal("no-op or suite binding overclaim")
	}
	var report struct {
		Review struct {
			Changed        bool            `json:"changed"`
			ReviewRequired bool            `json:"reviewRequired"`
			Changes        []impact.Change `json:"changes"`
		} `json:"review"`
	}
	if json.Unmarshal(outer["planner-result.json"], &report) != nil || report.Review.Changed || report.Review.ReviewRequired || report.Review.Changes == nil || len(report.Review.Changes) != 0 {
		t.Fatal("no-op change report")
	}
}

func TestTypedWireMixedResourceParity(t *testing.T) {
	_, s, src, _ := inputs(t)
	cases := []struct {
		request, suite []byte
		sources        []artifactbundle.File
		want           error
	}{
		{nil, s, append(append([]artifactbundle.File{}, src...), src[0]), ErrLimit},
		{nil, s, src, ErrInvalid},
		{nil, make([]byte, MaxSuiteBytes+1), src, ErrLimit},
	}
	for _, c := range cases {
		b, e := Build(c.request, c.suite, c.sources)
		requireError(t, b, e, c.want)
		b, e = Execute(buildCarrier(c.request, c.suite, c.sources))
		requireError(t, b, e, c.want)
	}
}
