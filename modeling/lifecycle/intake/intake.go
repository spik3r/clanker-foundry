// Package intake records caller declarations and compares exact artifact bytes.
// It neither validates artifacts nor verifies authorship, execution or review.
package intake

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const FormatVersion = "lifecycle-intake-json/v1"
const ReportVersion = "lifecycle-intake-report/v1"
const MaxSourceBytes = 4 << 20
const MaxArtifactBytes = 1 << 20

var ErrInvalid = errors.New("invalid intake request")

type Decision struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Rationale string `json:"rationale"`
}
type Question struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type Declaration struct {
	SHA256 string `json:"sha256"`
	Author string `json:"author"`
}
type declarations struct {
	Requirements  *Declaration `json:"requirements"`
	Model         *Declaration `json:"model"`
	ScenarioSuite *Declaration `json:"scenarioSuite"`
}
type observations struct {
	Requirements  *string `json:"requirements"`
	Model         *string `json:"model"`
	ScenarioSuite *string `json:"scenarioSuite"`
}
type dependencies struct {
	ModelRequirements *string `json:"modelRequirements"`
	SuiteRequirements *string `json:"suiteRequirements"`
	SuiteModel        *string `json:"suiteModel"`
}
type request struct {
	FormatVersion string       `json:"formatVersion"`
	Intent        string       `json:"intent"`
	Decisions     []Decision   `json:"decisions"`
	Questions     []Question   `json:"questions"`
	Artifacts     declarations `json:"artifacts"`
	Observed      observations `json:"observed"`
	Dependencies  dependencies `json:"dependencies"`
}
type binding struct {
	Kind           string       `json:"kind"`
	Declared       *Declaration `json:"declared"`
	ObservedSHA256 *string      `json:"observedSha256"`
	Status         string       `json:"status"`
}
type link struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	AssertedSHA256 *string `json:"assertedSha256"`
	Status         string  `json:"status"`
}
type report struct {
	FormatVersion string     `json:"formatVersion"`
	Authority     string     `json:"authority"`
	RequestSHA256 string     `json:"requestSha256"`
	Intent        string     `json:"intent"`
	Decisions     []Decision `json:"decisions"`
	Questions     []Question `json:"questions"`
	Bindings      []binding  `json:"bindings"`
	Links         []link     `json:"links"`
}

func digest(b []byte) string      { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func text(s string, max int) bool { return len(s) <= max && strings.TrimSpace(s) != "" }
func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range []byte(s) {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validDeclaration(d *Declaration) bool {
	if d == nil {
		return true
	}
	if !text(d.Author, 256) || !validHash(d.SHA256) {
		return false
	}
	return true
}
func makeBinding(kind string, d *Declaration, observed *string) binding {
	b := binding{Kind: kind, Declared: d, Status: "missing-declaration"}
	if observed != nil {
		h := digest([]byte(*observed))
		b.ObservedSHA256 = &h
	}
	switch {
	case d == nil && observed != nil:
		b.Status = "undeclared"
	case d == nil:
		b.Status = "missing-declaration"
	case observed == nil:
		b.Status = "missing-observation"
	case d.SHA256 != *b.ObservedSHA256:
		b.Status = "digest-mismatch"
	default:
		b.Status = "matches-declaration"
	}
	return b
}

func makeLink(from, to string, source, target *Declaration, asserted *string) link {
	l := link{From: from, To: to, AssertedSHA256: asserted}
	switch {
	case asserted == nil:
		l.Status = "missing-assertion"
	case source == nil || target == nil:
		l.Status = "missing-artifact-declaration"
	case *asserted != target.SHA256:
		l.Status = "digest-mismatch"
	default:
		l.Status = "matches-declaration"
	}
	return l
}

// Execute strictly decodes one request, hashes the exact UTF-8 contents supplied
// for each observation and returns a deterministic report. No I/O is performed.
func Execute(raw []byte) ([]byte, error) {
	var d request
	if err := strictjson.Decode(raw, &d, MaxSourceBytes, 64); err != nil || d.FormatVersion != FormatVersion || !text(d.Intent, 16384) || len(d.Decisions) > 64 || len(d.Questions) > 64 {
		return nil, ErrInvalid
	}
	seen := map[string]bool{}
	for _, v := range d.Decisions {
		if !model.Identifier(v.ID) || seen[v.ID] || !text(v.Text, 4096) || !text(v.Rationale, 4096) {
			return nil, ErrInvalid
		}
		seen[v.ID] = true
	}
	for _, v := range d.Questions {
		if !model.Identifier(v.ID) || seen[v.ID] || !text(v.Text, 4096) {
			return nil, ErrInvalid
		}
		seen[v.ID] = true
	}
	ds := []*Declaration{d.Artifacts.Requirements, d.Artifacts.Model, d.Artifacts.ScenarioSuite}
	os := []*string{d.Observed.Requirements, d.Observed.Model, d.Observed.ScenarioSuite}
	kinds := []string{"requirements", "model", "scenarioSuite"}
	bs := make([]binding, 0, 3)
	for i := range ds {
		if !validDeclaration(ds[i]) || (os[i] != nil && len(*os[i]) > MaxArtifactBytes) {
			return nil, ErrInvalid
		}
		bs = append(bs, makeBinding(kinds[i], ds[i], os[i]))
	}

	deps := []*string{d.Dependencies.ModelRequirements, d.Dependencies.SuiteRequirements, d.Dependencies.SuiteModel}
	for _, v := range deps {
		if v != nil && !validHash(*v) {
			return nil, ErrInvalid
		}
	}
	links := []link{
		makeLink("model", "requirements", ds[1], ds[0], deps[0]),
		makeLink("scenarioSuite", "requirements", ds[2], ds[0], deps[1]),
		makeLink("scenarioSuite", "model", ds[2], ds[1], deps[2]),
	}
	sort.Slice(d.Decisions, func(i, j int) bool { return d.Decisions[i].ID < d.Decisions[j].ID })
	sort.Slice(d.Questions, func(i, j int) bool { return d.Questions[i].ID < d.Questions[j].ID })
	r := report{ReportVersion, "none", digest(raw), d.Intent, d.Decisions, d.Questions, bs, links}
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if e.Encode(r) != nil {
		return nil, ErrInvalid
	}
	return out.Bytes(), nil
}
