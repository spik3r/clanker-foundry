// Package intakeimpact compares original intent and opaque artifact assertions.
// It does not validate artifacts, execute scenarios or establish review.
package intakeimpact

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const ReportVersion = "lifecycle-intake-impact/v1"
const RequestVersion = "lifecycle-intake-impact-request/v1"
const MaxRequestBytes = 64 << 20
const MaxReportBytes = 16 << 20

var ErrInvalid = errors.New("invalid intake comparison input")
var ErrReportLimit = errors.New("intake comparison report limit")

type Receipt struct {
	RequestSHA256 string `json:"requestSha256"`
	ContentSHA256 string `json:"contentSha256"`
	ReportSHA256  string `json:"reportSha256"`
	ReportJSON    string `json:"reportJSON"`
}
type Change struct {
	Kind   string   `json:"kind"`
	ID     string   `json:"id"`
	Change string   `json:"change"`
	Fields []string `json:"fields"`
}
type Gap struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Status string `json:"status"`
}
type Report struct {
	FormatVersion        string   `json:"formatVersion"`
	Mode                 string   `json:"mode"`
	Authority            string   `json:"authority"`
	Before               Receipt  `json:"before"`
	After                Receipt  `json:"after"`
	RequestSourceChanged bool     `json:"requestSourceChanged"`
	ContentChanged       bool     `json:"contentChanged"`
	Changes              []Change `json:"changes"`
	BeforeGaps           []Gap    `json:"beforeGaps"`
	AfterGaps            []Gap    `json:"afterGaps"`
	ArtifactValidation   string   `json:"artifactValidation"`
	ExecutionStatus      string   `json:"executionStatus"`
	ReviewStatus         string   `json:"reviewStatus"`
}

// This strict adapter reads the existing native report, not a second intake
// parser. Native report shape changes require explicit integration review.
type binding struct {
	Kind           string              `json:"kind"`
	Declared       *intake.Declaration `json:"declared"`
	ObservedSHA256 *string             `json:"observedSha256"`
	Status         string              `json:"status"`
}
type link struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	AssertedSHA256 *string `json:"assertedSha256"`
	Status         string  `json:"status"`
}
type nativeReport struct {
	FormatVersion string            `json:"formatVersion"`
	Authority     string            `json:"authority"`
	RequestSHA256 string            `json:"requestSha256"`
	Intent        string            `json:"intent"`
	Decisions     []intake.Decision `json:"decisions"`
	Questions     []intake.Question `json:"questions"`
	Bindings      []binding         `json:"bindings"`
	Links         []link            `json:"links"`
}
type content struct {
	FormatVersion string            `json:"formatVersion"`
	Intent        string            `json:"intent"`
	Decisions     []intake.Decision `json:"decisions"`
	Questions     []intake.Question `json:"questions"`
	Bindings      []binding         `json:"bindings"`
	Links         []link            `json:"links"`
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func encode(v any) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if enc.Encode(v) != nil {
		return nil, ErrInvalid
	}
	return out.Bytes(), nil
}
func read(raw []byte) (nativeReport, Receipt, error) {
	var n nativeReport
	var receipt Receipt
	data, e := intake.Execute(raw)
	if e != nil {
		return n, receipt, ErrInvalid
	}
	if strictjson.Decode(data, &n, MaxReportBytes, 64) != nil || n.FormatVersion != intake.ReportVersion || n.Authority != "none" || n.RequestSHA256 != hash(raw) {
		return nativeReport{}, receipt, ErrInvalid
	}
	c, e := encode(content{"lifecycle-intake-content/v1", n.Intent, n.Decisions, n.Questions, n.Bindings, n.Links})
	if e != nil {
		return nativeReport{}, receipt, e
	}
	receipt = Receipt{n.RequestSHA256, hash(c[:len(c)-1]), hash(data), string(data)}
	return n, receipt, nil
}
func gaps(n nativeReport) []Gap {
	out := []Gap{}
	for _, q := range n.Questions {
		out = append(out, Gap{"question", q.ID, "unresolved"})
	}
	for _, b := range n.Bindings {
		if b.Status != "matches-declaration" {
			out = append(out, Gap{"artifact", b.Kind, b.Status})
		}
	}
	for _, l := range n.Links {
		if l.Status != "matches-declaration" {
			out = append(out, Gap{"dependency", l.From + "->" + l.To, l.Status})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].ID < out[j].ID
	})
	return out
}
func same(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func declaration(d *intake.Declaration) (*string, *string) {
	if d == nil {
		return nil, nil
	}
	return &d.SHA256, &d.Author
}

// Compare validates both original requests through intake, preserves their exact
// native reports and compares only declared content and observed source hashes.
// Inputs must not be concurrently mutated. Errors return the zero Report.
func Compare(before, after []byte) (Report, error) {
	var zero Report
	if len(before) > intake.MaxSourceBytes || len(after) > intake.MaxSourceBytes {
		return zero, ErrInvalid
	}
	b, br, e := read(before)
	if e != nil {
		return zero, e
	}
	a, ar, e := read(after)
	if e != nil {
		return zero, e
	}
	r := Report{FormatVersion: ReportVersion, Mode: "example-only", Authority: "none", Before: br, After: ar, RequestSourceChanged: br.RequestSHA256 != ar.RequestSHA256, ContentChanged: br.ContentSHA256 != ar.ContentSHA256, Changes: []Change{}, BeforeGaps: gaps(b), AfterGaps: gaps(a), ArtifactValidation: "not-run", ExecutionStatus: "not-run", ReviewStatus: "not-established"}
	add := func(kind, id, change string, fields []string) {
		sort.Strings(fields)
		r.Changes = append(r.Changes, Change{kind, id, change, fields})
	}
	if b.Intent != a.Intent {
		add("intent", "intent", "modified", []string{"text"})
	}
	// Separate lists preserve a moved ID as removal plus addition, not resolution.
	bd, ad := map[string]intake.Decision{}, map[string]intake.Decision{}
	for _, v := range b.Decisions {
		bd[v.ID] = v
	}
	for _, v := range a.Decisions {
		ad[v.ID] = v
	}
	for id, v := range bd {
		w, ok := ad[id]
		if !ok {
			add("decision", id, "removed", []string{})
			continue
		}
		fields := []string{}
		if v.Text != w.Text {
			fields = append(fields, "text")
		}
		if v.Rationale != w.Rationale {
			fields = append(fields, "rationale")
		}
		if len(fields) > 0 {
			add("decision", id, "modified", fields)
		}
	}
	for id := range ad {
		if _, ok := bd[id]; !ok {
			add("decision", id, "added", []string{})
		}
	}
	bq, aq := map[string]intake.Question{}, map[string]intake.Question{}
	for _, v := range b.Questions {
		bq[v.ID] = v
	}
	for _, v := range a.Questions {
		aq[v.ID] = v
	}
	for id, v := range bq {
		w, ok := aq[id]
		if !ok {
			add("question", id, "removed", []string{})
			continue
		}
		if v.Text != w.Text {
			add("question", id, "modified", []string{"text"})
		}
	}
	for id := range aq {
		if _, ok := bq[id]; !ok {
			add("question", id, "added", []string{})
		}
	}
	// Native intake fixes role and dependency ordering. Refuse a future shape
	// change instead of silently ignoring a new role or dependency.
	if len(b.Bindings) != 3 || len(a.Bindings) != 3 || len(b.Links) != 3 || len(a.Links) != 3 {
		return zero, ErrInvalid
	}
	for i, kind := range []string{"requirements", "model", "scenarioSuite"} {
		v, w := b.Bindings[i], a.Bindings[i]
		if v.Kind != kind || w.Kind != kind {
			return zero, ErrInvalid
		}
		bh, ba := declaration(v.Declared)
		ah, aa := declaration(w.Declared)
		fields := []string{}
		if !same(bh, ah) {
			fields = append(fields, "declaredSha256")
		}
		if !same(ba, aa) {
			fields = append(fields, "declaredAuthor")
		}
		if !same(v.ObservedSHA256, w.ObservedSHA256) {
			fields = append(fields, "observedSha256")
		}
		if v.Status != w.Status {
			fields = append(fields, "status")
		}
		if len(fields) > 0 {
			add("artifact", kind, "modified", fields)
		}
	}
	for i, id := range []string{"model->requirements", "scenarioSuite->requirements", "scenarioSuite->model"} {
		v, w := b.Links[i], a.Links[i]
		if v.From+"->"+v.To != id || w.From+"->"+w.To != id {
			return zero, ErrInvalid
		}
		fields := []string{}
		if !same(v.AssertedSHA256, w.AssertedSHA256) {
			fields = append(fields, "assertedSha256")
		}
		if v.Status != w.Status {
			fields = append(fields, "status")
		}
		if len(fields) > 0 {
			add("dependency", id, "modified", fields)
		}
	}
	sort.Slice(r.Changes, func(i, j int) bool {
		if r.Changes[i].Kind != r.Changes[j].Kind {
			return r.Changes[i].Kind < r.Changes[j].Kind
		}
		return r.Changes[i].ID < r.Changes[j].ID
	})
	data, e := encode(r)
	if e != nil {
		return zero, e
	}
	if len(data) > MaxReportBytes {
		return zero, ErrReportLimit
	}
	return r, nil
}

// Execute decodes one bounded strict carrier of the original request strings.
// It buffers the complete deterministic report before returning any output.
func Execute(raw []byte) ([]byte, error) {
	var q struct {
		FormatVersion string `json:"formatVersion"`
		Before        string `json:"before"`
		After         string `json:"after"`
	}
	if strictjson.Decode(raw, &q, MaxRequestBytes, 8) != nil || q.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	r, e := Compare([]byte(q.Before), []byte(q.After))
	if e != nil {
		return nil, e
	}
	return encode(r)
}
