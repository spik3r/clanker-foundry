// Package traceimpact joins declared requirement trace and model changes for
// human review. It does not infer execution, coverage or dynamic impact.
package traceimpact

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
)

const ReportVersion = "lifecycle-trace-impact/v1"
const RequestVersion = "lifecycle-trace-impact-request/v1"
const MaxRequestBytes = 16 << 20
const MaxReportBytes = 4 << 20
const MaxCatalogCases = 512
const MaxReasons = 65536

var ErrInvalid = errors.New("invalid trace comparison input")
var ErrReportLimit = errors.New("trace comparison report limit")

type Snapshot struct{ Model, Catalog, Requirements []byte }
type Identity struct {
	ModelDigest        string `json:"modelDigest"`
	RequirementsDigest string `json:"requirementsDigest"`
	CatalogSHA256      string `json:"catalogSha256"`
}
type Gap struct {
	ID    string   `json:"id"`
	Codes []string `json:"codes"`
}
type Reason struct {
	Code   string `json:"code"`
	Side   string `json:"side"`
	Target string `json:"target"`
	Cause  string `json:"cause"`
}
type Candidate struct {
	ID            string   `json:"id"`
	Change        string   `json:"change"`
	Fields        []string `json:"fields"`
	BeforePresent bool     `json:"beforePresent"`
	AfterPresent  bool     `json:"afterPresent"`
	BeforeGaps    []string `json:"beforeGaps"`
	AfterGaps     []string `json:"afterGaps"`
	Reasons       []Reason `json:"reasons"`
}
type Report struct {
	FormatVersion        string      `json:"formatVersion"`
	Mode                 string      `json:"mode"`
	Authority            string      `json:"authority"`
	Before               Identity    `json:"before"`
	After                Identity    `json:"after"`
	ModelChanged         bool        `json:"modelChanged"`
	RequirementsChanged  bool        `json:"requirementsChanged"`
	CatalogSourceChanged bool        `json:"catalogSourceChanged"`
	ReviewRequired       bool        `json:"reviewRequired"`
	BeforeGaps           []Gap       `json:"beforeGaps"`
	AfterGaps            []Gap       `json:"afterGaps"`
	Candidates           []Candidate `json:"candidates"`
}

func readSnapshot(s Snapshot) (traceability.Report, model.Model, error) {
	if len(s.Model) > traceability.MaxBytes || len(s.Requirements) > traceability.MaxBytes || len(s.Catalog) > 4*traceability.MaxBytes {
		return traceability.Report{}, model.Model{}, ErrInvalid
	}
	cases, e := traceability.LoadCatalog(s.Catalog)
	if e != nil || len(cases) > MaxCatalogCases {
		return traceability.Report{}, model.Model{}, ErrInvalid
	}
	r, e := traceability.Build(s.Model, s.Catalog, s.Requirements)
	if e != nil {
		return traceability.Report{}, model.Model{}, ErrInvalid
	}
	m, e := jsonformat.Decode(s.Model)
	if e != nil {
		return traceability.Report{}, model.Model{}, ErrInvalid
	}
	return r, m, nil
}
func identity(r traceability.Report) Identity {
	return Identity{r.Inputs.ModelDigest, r.Inputs.RequirementsDigest, r.Inputs.CatalogSHA256}
}
func gaps(r traceability.Report) []Gap {
	out := []Gap{}
	for _, q := range r.Requirements {
		if len(q.Gaps) > 0 {
			out = append(out, Gap{q.ID, append([]string{}, q.Gaps...)})
		}
	}
	return out
}
func modelLinks(r traceability.RequirementTrace) []string {
	out := make([]string, 0, len(r.ModelLinks))
	for _, l := range r.ModelLinks {
		out = append(out, l.Kind+":"+l.ID)
	}
	sort.Strings(out)
	return out
}
func caseLinks(r traceability.RequirementTrace) []string {
	out := make([]string, 0, len(r.AcceptanceLinks))
	for _, l := range r.AcceptanceLinks {
		out = append(out, l.CaseID)
	}
	sort.Strings(out)
	return out
}
func traceKind(kind string) string {
	if kind == "requirement" {
		return traceability.KindModelRequirement
	}
	return kind
}

// Compare validates both complete snapshots and projects declared links against
// model-diff facts. It never evaluates a case or accepts a prebuilt report.
func Compare(before, after Snapshot) (Report, error) {
	var zero Report
	b, bm, e := readSnapshot(before)
	if e != nil {
		return zero, e
	}
	a, am, e := readSnapshot(after)
	if e != nil {
		return zero, e
	}
	diff, e := impact.Compare(bm, am, nil)
	if e != nil {
		return zero, ErrInvalid
	}
	if diff.BeforeDigest != b.Inputs.ModelDigest || diff.AfterDigest != a.Inputs.ModelDigest {
		return zero, ErrInvalid
	}
	r := Report{FormatVersion: ReportVersion, Mode: "example-only", Authority: "none", Before: identity(b), After: identity(a), ModelChanged: diff.Changed, RequirementsChanged: b.Inputs.RequirementsDigest != a.Inputs.RequirementsDigest, CatalogSourceChanged: b.Inputs.CatalogSHA256 != a.Inputs.CatalogSHA256, BeforeGaps: gaps(b), AfterGaps: gaps(a), Candidates: []Candidate{}}
	r.ReviewRequired = r.ModelChanged || r.RequirementsChanged || r.CatalogSourceChanged
	changed := map[string]string{}
	for _, c := range diff.Changes {
		if c.Kind != "model" {
			changed[traceKind(c.Kind)+":"+c.ID] = c.Type
		}
	}
	transitions := map[string]map[string][]string{"before": {}, "after": {}}
	for _, link := range diff.LinkedTransitions {
		transitions[link.Side][link.ID] = append([]string{}, link.Causes...)
	}
	left := map[string]traceability.RequirementTrace{}
	right := map[string]traceability.RequirementTrace{}
	ids := map[string]bool{}
	for _, q := range b.Requirements {
		left[q.ID] = q
		ids[q.ID] = true
	}
	for _, q := range a.Requirements {
		right[q.ID] = q
		ids[q.ID] = true
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	totalReasons := 0
	for _, id := range ordered {
		l, lp := left[id]
		v, rp := right[id]
		c := Candidate{ID: id, Change: "unchanged", Fields: []string{}, BeforePresent: lp, AfterPresent: rp, BeforeGaps: []string{}, AfterGaps: []string{}, Reasons: []Reason{}}
		if lp {
			c.BeforeGaps = append(c.BeforeGaps, l.Gaps...)
		}
		if rp {
			c.AfterGaps = append(c.AfterGaps, v.Gaps...)
		}
		seen := map[Reason]bool{}
		over := false
		add := func(reason Reason) {
			if seen[reason] {
				return
			}
			if totalReasons >= MaxReasons {
				over = true
				return
			}
			seen[reason] = true
			totalReasons++
			c.Reasons = append(c.Reasons, reason)
		}
		switch {
		case !lp:
			c.Change = "added"
			add(Reason{"requirement-added", "after", "", ""})
		case !rp:
			c.Change = "removed"
			add(Reason{"requirement-removed", "before", "", ""})
		default:
			if l.Title != v.Title {
				c.Fields = append(c.Fields, "title")
			}
			if l.Statement != v.Statement {
				c.Fields = append(c.Fields, "statement")
			}
			if !reflect.DeepEqual(modelLinks(l), modelLinks(v)) {
				c.Fields = append(c.Fields, "modelLinks")
			}
			if !reflect.DeepEqual(caseLinks(l), caseLinks(v)) {
				c.Fields = append(c.Fields, "acceptanceLinks")
			}
			if !reflect.DeepEqual(l.Gaps, v.Gaps) {
				c.Fields = append(c.Fields, "gaps")
			}
			sort.Strings(c.Fields)
			if len(c.Fields) > 0 {
				c.Change = "modified"
				for _, f := range c.Fields {
					add(Reason{"requirement-changed", "both", f, ""})
				}
			}
		}
		visit := func(side string, q traceability.RequirementTrace) {
			for _, link := range q.ModelLinks {
				key := link.Kind + ":" + link.ID
				if typ, ok := changed[key]; ok {
					add(Reason{"model-element-changed", side, key, typ})
				}
				if link.Kind == traceability.KindTransition {
					for _, cause := range transitions[side][link.ID] {
						add(Reason{"linked-transition-changed", side, link.ID, cause})
					}
				}
			}
		}
		if lp {
			visit("before", l)
		}
		if rp {
			visit("after", v)
		}
		if r.CatalogSourceChanged && ((lp && len(l.AcceptanceLinks) > 0) || (rp && len(v.AcceptanceLinks) > 0)) {
			add(Reason{"catalog-source-changed", "both", "", ""})
		}
		if over {
			return zero, ErrReportLimit
		}
		sort.Slice(c.Reasons, func(i, j int) bool {
			x, y := c.Reasons[i], c.Reasons[j]
			if x.Code != y.Code {
				return x.Code < y.Code
			}
			if x.Side != y.Side {
				return x.Side < y.Side
			}
			if x.Target != y.Target {
				return x.Target < y.Target
			}
			return x.Cause < y.Cause
		})
		if len(c.Reasons) > 0 {
			r.Candidates = append(r.Candidates, c)
		}
	}
	encoded, e := json.Marshal(r)
	if e != nil || len(encoded) > MaxReportBytes {
		return zero, ErrReportLimit
	}
	return r, nil
}

type wireSnapshot struct {
	Model        string `json:"model"`
	Catalog      string `json:"catalog"`
	Requirements string `json:"requirements"`
}
type request struct {
	FormatVersion string       `json:"formatVersion"`
	Before        wireSnapshot `json:"before"`
	After         wireSnapshot `json:"after"`
}

func snapshot(s wireSnapshot) Snapshot {
	return Snapshot{Model: []byte(s.Model), Catalog: []byte(s.Catalog), Requirements: []byte(s.Requirements)}
}

// Execute strictly decodes literal UTF-8 artifact strings and buffers a report.
func Execute(raw []byte) ([]byte, error) {
	var q request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 8) != nil || q.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	r, e := Compare(snapshot(q.Before), snapshot(q.After))
	if e != nil {
		return nil, e
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if enc.Encode(r) != nil || b.Len() > MaxReportBytes {
		return nil, ErrReportLimit
	}
	return b.Bytes(), nil
}
