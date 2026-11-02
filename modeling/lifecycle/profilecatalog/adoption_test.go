// Independent pre-policy oracle. Historical archive bytes are regression inputs,
// not generators of semantic expectations. No new policy output authored this file.
package profilecatalog_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
)

const policyLegacy = "3d6d17188cc2723cfe04649d0f8b4220a37b1afe"
const policyCurrent = "1e88dbebb4ab13873e2a4159dc928dd1f5f73e20"
const policyOuterSHA = "d7cffd252c57f0854257b8cfe019b88a482f2482110473343ef000cabb868c91"
const policyInnerSHA = "f5b20379088cffaec7d98bb4b965fd91438bb20ccca440770a2d0f683abae945"

type policyPin struct {
	name string
	size int
	sha  string
}

var policyLegacyPins = []policyPin{
	{"go.mod", 71, "cc128fce13193d3066bdeb08cca9db627fb6c3120967c3beff254a514f2a5fac"},
	{"model/ir.go", 9385, "13f4ad2980b1ca523b07bdbc8ebeb0e0300f8df81faa8dcd1cdec5ccef007aeb"},
	{"canonical/canonical.go", 1946, "2c3cba3fd53adf8ca5a4cc2ce3d96b51457742c47b6ab4fcedba219e19001a06"},
	{"strictjson/decode.go", 4236, "a72ded4af99ffba5f5cca2931c6cace072634ee2e42eb00cd9a5988c691f1ee2"},
	{"compiler/generate.go", 12315, "e17610a27748a5d3baedf4f59c2183cb184f7e2681b5532c6f9d8d3d1eda6a38"},
	{"scenarios/types.go", 5904, "147b12bd1726fec156ba08ec5666a9a362e1563115151ae57bf9face70f86eec"},
	{"scenarios/decode.go", 7047, "e6119229899d3084820d72a777d1771037f871b14554630067b7da022cdedc15"},
	{"scenarios/run.go", 10114, "46d41c5d27d87f71f7191f18d182a99c3ca491126afb601599067692c3379dde"},
	{"cmd/modelplay/main.go", 1613, "925ec202c74aba5d50eb2bfb5dc54ea22387ccec90e969f1a7ebe1ecbd9a0c63"},
	{"scenarios/runner.go", 4005, "e85b2146d7c1d16269ff886f6c67084de5e263db8b01a7676b8c7491ed8fb59c"},
	{"scenarios/report_decode.go", 3626, "f14816a19faacf707e23b01d5d16b26be8e8c03a9f5f9af801029b89fb2ceddf"},
}

func policyPins(revision string) []policyPin {
	out := append([]policyPin{}, policyLegacyPins...)
	if revision == policyCurrent {
		out = append(out, policyPin{"model/capabilities.go", 654, "e78f4f9c102f8d894e50130dd0bfb8030e0c7c325b01eb2452e9839c02c27e4e"})
	}
	return out
}
func policySHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func policyExpected() designchange.Expected {
	return designchange.Expected{RequestSHA256: "f3a14ba6adfd54832d810e26dccfc929e5fa61801990fa7c52436ec51eaacde0", BeforeSourceSHA256: "95b17eb3e2ef78a9485b40076c0b085e5e542e8fa3db827ea99345abe6077e26", BeforeModelDigest: "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed", CandidateSHA256: "22e6ee6f5208b124bcebffe2ac8608920a74cd34c0eeefe8ab8519b2523ea8c6", CandidateModelDigest: "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c", SuiteSHA256: "53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1"}
}
func policyRead(t *testing.T, p string, size int, sha string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil || len(b) != size || policySHA(b) != sha {
		t.Fatal("frozen byte identity drift", p, e)
	}
	return b
}
func policySources(t *testing.T, rev string) []artifactbundle.File {
	t.Helper()
	out := []artifactbundle.File{}
	for _, p := range policyPins(rev) {
		out = append(out, artifactbundle.File{Name: p.name, Data: policyRead(t, filepath.Join("..", p.name), p.size, p.sha)})
	}
	return out
}
func policyUnpack(t *testing.T, raw []byte, profile string) map[string][]byte {
	t.Helper()
	files, e := artifactbundle.Verify(raw, profile)
	if e != nil {
		t.Fatal(e)
	}
	out := map[string][]byte{}
	for _, f := range files {
		out[f.Name] = f.Data
	}
	return out
}
func policyPack(t *testing.T, m map[string][]byte, profile string) []byte {
	t.Helper()
	files := []artifactbundle.File{}
	for n, b := range m {
		files = append(files, artifactbundle.File{Name: n, Data: b})
	}
	out, e := artifactbundle.Build(profile, files)
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func policyJSON(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if e := json.Unmarshal(b, &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func policyEncode(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return append(b, '\n')
}
func policyReplaceInner(t *testing.T, outerRaw, innerRaw []byte) []byte {
	t.Helper()
	outer := policyUnpack(t, outerRaw, designchange.Profile)
	outer["qualification.zip"] = innerRaw
	meta := policyJSON(t, outer["design-change.json"])
	meta["qualificationSHA256"] = policySHA(innerRaw)
	outer["design-change.json"] = policyEncode(t, meta)
	return policyPack(t, outer, designchange.Profile)
}
func policyFrozen(t *testing.T) ([]byte, map[string][]byte, map[string][]byte) {
	t.Helper()
	raw := policyRead(t, "testdata/public-11-source-project.zip", 137626, policyOuterSHA)
	outer := policyUnpack(t, raw, designchange.Profile)
	if len(outer["qualification.zip"]) != 101580 || policySHA(outer["qualification.zip"]) != policyInnerSHA {
		t.Fatal("historical inner archive drift")
	}
	inner := policyUnpack(t, outer["qualification.zip"], artifactbundle.QualificationProfile)
	w := policyExpected()
	if policySHA(outer["change-request.json"]) != w.RequestSHA256 || policySHA(outer["before.json"]) != w.BeforeSourceSHA256 || policySHA(outer["candidate.json"]) != w.CandidateSHA256 || policySHA(inner["suite.json"]) != w.SuiteSHA256 || !bytes.Equal(outer["candidate.json"], inner["candidate.json"]) {
		t.Fatal("historical raw inputs differ from independent six-identity reference")
	}
	return raw, outer, inner
}
func policyCheckInventory(t *testing.T, raw []byte, rev string) {
	t.Helper()
	inner := policyUnpack(t, raw, artifactbundle.QualificationProfile)
	want := []string{"GUIDE.txt", "candidate.json", "suite.json", "qualification.json", "generated/lifecycle.go", "generated/model_paths_test.go"}
	for _, p := range policyPins(rev) {
		want = append(want, p.name)
		if len(inner[p.name]) != p.size || policySHA(inner[p.name]) != p.sha {
			t.Fatal("policy payload pin drift", rev, p.name)
		}
	}
	got := []string{}
	for n := range inner {
		got = append(got, n)
	}
	sort.Strings(want)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("policy domain inventory differs", rev, got)
	}
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil || len(z.File) != len(want)+1 {
		t.Fatal("missing manifest or payload count", e)
	}
	w := policyExpected()
	expected := map[string]any{"format": "lifecycle-json/v1", "profile": "lifecycle/v1", "generator": "lifecycle-go/v1", "toolchain": "go1.27.3", "modelDigest": w.CandidateModelDigest, "sourceRevision": rev, "suiteStatus": "unvalidated and unexecuted"}
	if !reflect.DeepEqual(policyJSON(t, inner["qualification.json"]), expected) {
		t.Fatal("policy metadata/status differs")
	}
}
func TestClosedPolicyNamesOwnershipAndConcurrentCalls(t *testing.T) {
	if qualification.SourceRevision != policyCurrent {
		t.Fatal("default revision is not actual helper-freeze commit")
	}
	for _, rev := range []string{policyLegacy, policyCurrent} {
		want := []string{}
		for _, p := range policyPins(rev) {
			want = append(want, p.name)
		}
		sort.Strings(want)
		got, e := qualification.SourceNamesForRevision(rev)
		if e != nil {
			t.Fatal(e)
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, want) {
			t.Fatal("policy name set differs")
		}
		got[0] = "caller mutation"
		again, e := qualification.SourceNamesForRevision(rev)
		sort.Strings(again)
		if e != nil || !reflect.DeepEqual(again, want) {
			t.Fatal("policy names alias")
		}
		if rev == policyCurrent {
			current := qualification.SourceNames()
			sort.Strings(current)
			if !reflect.DeepEqual(current, want) {
				t.Fatal("default not current")
			}
		}
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n, e := qualification.SourceNamesForRevision(rev)
				sort.Strings(n)
				if e != nil || !reflect.DeepEqual(n, want) {
					t.Error("concurrent policy corruption")
				}
				if len(n) > 0 {
					n[0] = "mutated"
				}
			}()
		}
		wg.Wait()
	}
	for _, bad := range []string{"", strings.Repeat("0", 40), strings.ToUpper(policyCurrent), "3312a090afac526ddeda392dbd31ed49fa0e04ca", policyLegacy + "x", strings.Repeat("a", 1<<20)} {
		names, e := qualification.SourceNamesForRevision(bad)
		if !errors.Is(e, qualification.ErrInvalid) || names != nil {
			t.Fatal("unknown policy names accepted")
		}
		raw, e := qualification.BuildForRevision(bad, nil, nil, nil)
		if !errors.Is(e, qualification.ErrInvalid) || raw != nil {
			t.Fatal("unknown policy build accepted")
		}
	}
}
func TestFrozenLegacyAndCurrentPolicyBuildVerify(t *testing.T) {
	old, outer, inner := policyFrozen(t)
	w := policyExpected()
	oldBefore := bytes.Clone(old)
	if e := designchange.Verify(old, w); e != nil {
		t.Fatal("public historical outer rejected", e)
	}
	if e := qualification.Verify(outer["qualification.zip"], w.CandidateSHA256, w.SuiteSHA256); e != nil {
		t.Fatal("public historical inner rejected", e)
	}
	policyCheckInventory(t, outer["qualification.zip"], policyLegacy)
	legacySources := policySources(t, policyLegacy)
	currentSources := policySources(t, policyCurrent)
	rebuilt, e := qualification.BuildForRevision(policyLegacy, inner["candidate.json"], inner["suite.json"], legacySources)
	if e != nil || !bytes.Equal(rebuilt, outer["qualification.zip"]) {
		t.Fatal("explicit legacy rebuild changed frozen bytes", e)
	}
	current, e := qualification.Build(inner["candidate.json"], inner["suite.json"], currentSources)
	if e != nil {
		t.Fatal(e)
	}
	explicit, e := qualification.BuildForRevision(policyCurrent, inner["candidate.json"], inner["suite.json"], currentSources)
	if e != nil || !bytes.Equal(current, explicit) {
		t.Fatal("current implicit/explicit build differs", e)
	}
	if bytes.Equal(current, rebuilt) {
		t.Fatal("new policy silently reused legacy archive")
	}
	if e = qualification.Verify(current, w.CandidateSHA256, w.SuiteSHA256); e != nil {
		t.Fatal(e)
	}
	policyCheckInventory(t, current, policyCurrent)
	newOuter, e := designchange.Build(outer["change-request.json"], inner["suite.json"], currentSources)
	if e != nil {
		t.Fatal(e)
	}
	if e = designchange.Verify(newOuter, w); e != nil {
		t.Fatal(e)
	}
	newParts := policyUnpack(t, newOuter, designchange.Profile)
	if !bytes.Equal(newParts["qualification.zip"], current) || !bytes.Equal(newParts["candidate.json"], outer["candidate.json"]) {
		t.Fatal("current designchange diverged from native qualifier or rewrote candidate")
	}
	meta := policyJSON(t, newParts["design-change.json"])
	if meta["qualificationSourceRevision"] != policyCurrent || meta["authority"] != "none" || meta["mode"] != "example-only" || meta["suiteStatus"] != "unvalidated and unexecuted" || meta["suiteBindingStatus"] != "not inspected" {
		t.Fatal("current receipt policy/authority mismatch")
	}
	// Current builds must not silently upgrade the historical eleven-source input.
	for _, tc := range []struct {
		rev string
		src []artifactbundle.File
	}{{policyLegacy, currentSources}, {policyCurrent, legacySources}} {
		b, e := qualification.BuildForRevision(tc.rev, inner["candidate.json"], inner["suite.json"], tc.src)
		if e == nil || b != nil {
			t.Fatal("mixed policy source inventory accepted")
		}
	}
	b, e := qualification.Build(inner["candidate.json"], inner["suite.json"], legacySources)
	if e == nil || b != nil {
		t.Fatal("default build silently upgraded legacy inventory")
	}
	b, e = designchange.Build(outer["change-request.json"], inner["suite.json"], legacySources)
	if e == nil || b != nil {
		t.Fatal("default designchange silently upgraded legacy inventory")
	}
	for _, archive := range [][]byte{old, newOuter} {
		for i := 0; i < 6; i++ {
			wrong := w
			fields := []*string{&wrong.RequestSHA256, &wrong.BeforeSourceSHA256, &wrong.BeforeModelDigest, &wrong.CandidateSHA256, &wrong.CandidateModelDigest, &wrong.SuiteSHA256}
			*fields[i] = strings.Repeat("0", 64)
			if designchange.Verify(archive, wrong) == nil {
				t.Fatal("external identity bypass", i)
			}
		}
	}
	for _, q := range [][]byte{rebuilt, current} {
		if qualification.Verify(q, strings.Repeat("0", 64), w.SuiteSHA256) == nil || qualification.Verify(q, w.CandidateSHA256, strings.Repeat("0", 64)) == nil {
			t.Fatal("qualification external hash bypass")
		}
	}
	if !bytes.Equal(old, oldBefore) || policySHA(old) != policyOuterSHA {
		t.Fatal("historical archive migrated or mutated")
	}
}
func TestPolicyWireLegacyVerifyButCurrentOnlyBuild(t *testing.T) {
	old, outer, inner := policyFrozen(t)
	w := policyExpected()
	request := map[string]any{"formatVersion": "lifecycle-designchange-json/v1", "operation": "verify", "requestBase64": nil, "suiteBase64": nil, "sources": nil, "archiveBase64": base64.StdEncoding.EncodeToString(old), "expected": w}
	out, e := designchange.Execute(policyEncode(t, request))
	if e != nil {
		t.Fatal(e)
	}
	want := map[string]any{"formatVersion": "lifecycle-designchange-verification/v1", "consistent": true, "mode": "example-only", "authority": "none", "suiteStatus": "unvalidated and unexecuted", "suiteBindingStatus": "not inspected"}
	if !reflect.DeepEqual(policyJSON(t, out), want) {
		t.Fatal("legacy verification receipt claims migration/currentness/authority")
	}
	sources := []map[string]string{}
	for _, f := range policySources(t, policyLegacy) {
		sources = append(sources, map[string]string{"name": f.Name, "dataBase64": base64.StdEncoding.EncodeToString(f.Data)})
	}
	request = map[string]any{"formatVersion": "lifecycle-designchange-json/v1", "operation": "build", "requestBase64": base64.StdEncoding.EncodeToString(outer["change-request.json"]), "suiteBase64": base64.StdEncoding.EncodeToString(inner["suite.json"]), "sources": sources, "archiveBase64": nil, "expected": nil}
	out, e = designchange.Execute(policyEncode(t, request))
	if e == nil || out != nil {
		t.Fatal("wire build accepted old eleven-source carrier")
	}
	sources = []map[string]string{}
	for _, f := range policySources(t, policyCurrent) {
		sources = append(sources, map[string]string{"name": f.Name, "dataBase64": base64.StdEncoding.EncodeToString(f.Data)})
	}
	request["sources"] = sources
	out, e = designchange.Execute(policyEncode(t, request))
	if e != nil {
		t.Fatal(e)
	}
	if e = designchange.Verify(out, w); e != nil {
		t.Fatal(e)
	}
	meta := policyJSON(t, policyUnpack(t, out, designchange.Profile)["design-change.json"])
	if meta["qualificationSourceRevision"] != policyCurrent {
		t.Fatal("wire build not current")
	}
}
func TestPolicyMetadataRoutingIsNotTrust(t *testing.T) {
	old, outer, inner := policyFrozen(t)
	w := policyExpected()
	current, e := designchange.Build(outer["change-request.json"], inner["suite.json"], policySources(t, policyCurrent))
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		rev   string
		outer []byte
	}{{policyLegacy, old}, {policyCurrent, current}} {
		for _, mutation := range []string{"other-policy", "unknown-policy", "held-policy", "empty-policy", "uppercase-policy", "missing-policy", "null-policy", "duplicate-key", "unknown-key", "oversize-metadata", "promoted-suite", "guide-tamper", "generated-tamper", "pin-tamper", "mixed-inventory"} {
			t.Run(tc.rev[:8]+"/"+mutation, func(t *testing.T) {
				parts := policyUnpack(t, tc.outer, designchange.Profile)
				nested := policyUnpack(t, parts["qualification.zip"], artifactbundle.QualificationProfile)
				metadata := policyJSON(t, nested["qualification.json"])
				other := policyCurrent
				if tc.rev == policyCurrent {
					other = policyLegacy
				}
				switch mutation {
				case "other-policy":
					metadata["sourceRevision"] = other
				case "unknown-policy":
					metadata["sourceRevision"] = strings.Repeat("0", 40)
				case "held-policy":
					metadata["sourceRevision"] = "3312a090afac526ddeda392dbd31ed49fa0e04ca"
				case "empty-policy":
					metadata["sourceRevision"] = ""
				case "uppercase-policy":
					metadata["sourceRevision"] = strings.ToUpper(tc.rev)
				case "missing-policy":
					delete(metadata, "sourceRevision")
				case "null-policy":
					metadata["sourceRevision"] = nil
				case "unknown-key":
					metadata["unknown"] = true
				case "promoted-suite":
					metadata["suiteStatus"] = "executed and approved"
				case "guide-tamper":
					nested["GUIDE.txt"] = append(bytes.Clone(nested["GUIDE.txt"]), 'x')
				case "generated-tamper":
					nested["generated/lifecycle.go"] = append(bytes.Clone(nested["generated/lifecycle.go"]), '\n')
				case "pin-tamper":
					nested["scenarios/runner.go"][0] ^= 1
				case "mixed-inventory":
					if tc.rev == policyCurrent {
						delete(nested, "model/capabilities.go")
					} else {
						p := policyPins(policyCurrent)[11]
						nested[p.name] = policyRead(t, filepath.Join("..", p.name), p.size, p.sha)
					}
				}
				nested["qualification.json"] = policyEncode(t, metadata)
				if mutation == "duplicate-key" {
					nested["qualification.json"] = append([]byte(`{"sourceRevision":"`+tc.rev+`",`), nested["qualification.json"][1:]...)
				}
				if mutation == "oversize-metadata" {
					nested["qualification.json"] = append(nested["qualification.json"], bytes.Repeat([]byte(" "), 4097)...)
				}
				altered := policyPack(t, nested, artifactbundle.QualificationProfile)
				if qualification.Verify(altered, w.CandidateSHA256, w.SuiteSHA256) == nil {
					t.Fatal("native verifier trusted malformed/mixed metadata")
				}
				repairedOuter := policyReplaceInner(t, tc.outer, altered)
				if designchange.Verify(repairedOuter, w) == nil {
					t.Fatal("outer verifier trusted malformed/mixed inner metadata")
				}
			})
		}
		// Outer metadata cannot select or override the recognized inner policy.
		parts := policyUnpack(t, tc.outer, designchange.Profile)
		meta := policyJSON(t, parts["design-change.json"])
		if tc.rev == policyLegacy {
			meta["qualificationSourceRevision"] = policyCurrent
		} else {
			meta["qualificationSourceRevision"] = policyLegacy
		}
		parts["design-change.json"] = policyEncode(t, meta)
		if designchange.Verify(policyPack(t, parts, designchange.Profile), w) == nil {
			t.Fatal("outer/inner policy disagreement accepted")
		}
	}
}
func TestPolicyOpaqueSuitesAndNoInputMutation(t *testing.T) {
	old, outer, inner := policyFrozen(t)
	savedOld := bytes.Clone(old)
	candidate := bytes.Clone(inner["candidate.json"])
	candidateBefore := bytes.Clone(candidate)
	for _, rev := range []string{policyLegacy, policyCurrent} {
		src := policySources(t, rev)
		saved := []artifactbundle.File{}
		for _, f := range src {
			saved = append(saved, artifactbundle.File{Name: f.Name, Data: bytes.Clone(f.Data)})
		}
		opaque := []byte("{ deliberately malformed suite\n")
		before := bytes.Clone(opaque)
		a, e := qualification.BuildForRevision(rev, candidate, opaque, src)
		if e != nil {
			t.Fatal(e)
		}
		b, e := qualification.BuildForRevision(rev, candidate, opaque, src)
		if e != nil || !bytes.Equal(a, b) {
			t.Fatal("policy nondeterminism", e)
		}
		if e = qualification.Verify(a, policySHA(candidate), policySHA(opaque)); e != nil {
			t.Fatal(e)
		}
		payload := policyUnpack(t, a, artifactbundle.QualificationProfile)
		if !bytes.Equal(payload["suite.json"], opaque) || policyJSON(t, payload["qualification.json"])["suiteStatus"] != "unvalidated and unexecuted" {
			t.Fatal("opaque suite rewritten/promoted")
		}
		if !bytes.Equal(candidate, candidateBefore) || !bytes.Equal(opaque, before) || !reflect.DeepEqual(src, saved) {
			t.Fatal("caller input mutation")
		}
	}
	opaque := []byte(`{"modelDigest":"wrong","cases":null}`)
	w := policyExpected()
	w.SuiteSHA256 = policySHA(opaque)
	raw, e := designchange.Build(outer["change-request.json"], opaque, policySources(t, policyCurrent))
	if e != nil {
		t.Fatal(e)
	}
	if e = designchange.Verify(raw, w); e != nil {
		t.Fatal(e)
	}
	parts := policyUnpack(t, raw, designchange.Profile)
	meta := policyJSON(t, parts["design-change.json"])
	if meta["suiteStatus"] != "unvalidated and unexecuted" || meta["suiteBindingStatus"] != "not inspected" || meta["authority"] != "none" {
		t.Fatal("opaque designchange suite promoted")
	}
	if !bytes.Equal(old, savedOld) {
		t.Fatal("historical input mutated")
	}
}
func TestEachPolicyEnforcesEverySourcePinAndExactSet(t *testing.T) {
	_, _, inner := policyFrozen(t)
	for _, rev := range []string{policyLegacy, policyCurrent} {
		source := policySources(t, rev)
		for i := range source {
			bad := append([]artifactbundle.File{}, source...)
			bad[i].Data = bytes.Clone(bad[i].Data)
			bad[i].Data[0] ^= 1
			raw, e := qualification.BuildForRevision(rev, inner["candidate.json"], inner["suite.json"], bad)
			if e == nil || raw != nil {
				t.Fatal("ignored pinned source mutation", rev, bad[i].Name)
			}
		}
		duplicate := append([]artifactbundle.File{}, source...)
		duplicate[len(duplicate)-1] = duplicate[0]
		extra := append(append([]artifactbundle.File{}, source...), artifactbundle.File{Name: "extra.go", Data: []byte("package extra")})
		for _, bad := range [][]artifactbundle.File{source[:len(source)-1], duplicate, extra} {
			raw, e := qualification.BuildForRevision(rev, inner["candidate.json"], inner["suite.json"], bad)
			if e == nil || raw != nil {
				t.Fatal("nonexact source inventory accepted", rev)
			}
		}
	}
	n := qualification.SourceNames()
	n[0] = "caller changed default"
	fresh := qualification.SourceNames()
	want, e := qualification.SourceNamesForRevision(policyCurrent)
	if e != nil {
		t.Fatal(e)
	}
	sort.Strings(fresh)
	sort.Strings(want)
	if !reflect.DeepEqual(fresh, want) {
		t.Fatal("default names alias registry")
	}
}
