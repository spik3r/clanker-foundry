// Package traceability links human-stated outcome requirements to elements of
// the accepted lifecycle model and to independent acceptance cases.
//
// Links are declarations made by people. A complete trace shows that every
// requirement points somewhere and every model element and case is pointed at;
// it does not show that a requirement is correct, that the model satisfies it or
// that a linked case tests it. Invalid input (bad structure or references) is
// reported as an InvalidError. Incomplete coverage of valid input is reported in
// the Report as gaps.
package traceability

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	FormatVersion = "lifecycle-requirements-json/v1"
	ReportVersion = "lifecycle-requirements-trace/v1"
	// CatalogSchema is the only acceptance catalog this version reads.
	CatalogSchema = "lifecycle-semantic-oracle-v2"

	MaxBytes        = 1 << 20
	MaxDepth        = 8
	MaxRequirements = 256
	MaxLinks        = 64
	MaxTitle        = 120
	MaxStatement    = 1024

	StatusLinked     = "linked"
	StatusIncomplete = "incomplete"

	GapNoModelLink            = "no-model-link"
	GapNoAcceptanceLink       = "no-acceptance-link"
	GapNoExecutableAcceptance = "no-executable-acceptance"

	KindState            = "state"
	KindTransition       = "transition"
	KindGuard            = "guard"
	KindModelRequirement = "modelRequirement"

	Notice = "Links are reviewer declarations. They do not show that a requirement is correct, that the model satisfies it, or that a linked case tests it."
)

var kindOrder = []string{KindState, KindTransition, KindGuard, KindModelRequirement}

// Issue is one reason the requirements input is invalid. Requirement is empty
// for document-level issues. Target names the offending link, or the
// requirement an ambiguous statement collides with.
type Issue struct {
	Code        string `json:"code"`
	Requirement string `json:"requirement"`
	Target      string `json:"target"`
}

func (i Issue) String() string {
	s := i.Code
	if i.Requirement != "" {
		s += " " + i.Requirement
	}
	if i.Target != "" {
		s += " -> " + i.Target
	}
	return s
}

// InvalidError lists every issue found, sorted by requirement, code, target.
type InvalidError struct{ Issues []Issue }

func (e *InvalidError) Error() string {
	parts := make([]string, len(e.Issues))
	for i, x := range e.Issues {
		parts[i] = x.String()
	}
	return "invalid requirements: " + strings.Join(parts, "; ")
}

type wireModelLink struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}
type wireAcceptanceLink struct {
	CaseID string `json:"caseId"`
}
type wireRequirement struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	Statement       string               `json:"statement"`
	ModelLinks      []wireModelLink      `json:"modelLinks"`
	AcceptanceLinks []wireAcceptanceLink `json:"acceptanceLinks"`
}
type wireDocument struct {
	FormatVersion     string            `json:"formatVersion"`
	ModelID           string            `json:"modelId"`
	AcceptanceCatalog string            `json:"acceptanceCatalog"`
	Requirements      []wireRequirement `json:"requirements"`
}

// Report is the trace derived from one model, one case catalog and one
// requirements document. Every list is sorted, so the same inputs in any
// requirement or link order give the same report.
type Report struct {
	ReportVersion   string             `json:"reportVersion"`
	Mode            string             `json:"mode"`
	Authority       string             `json:"authority"`
	Notice          string             `json:"notice"`
	Inputs          Inputs             `json:"inputs"`
	Summary         Summary            `json:"summary"`
	Requirements    []RequirementTrace `json:"requirements"`
	ModelElements   []ElementTrace     `json:"modelElements"`
	AcceptanceCases []CaseTrace        `json:"acceptanceCases"`
}
type Inputs struct {
	RequirementsFormat string `json:"requirementsFormat"`
	// RequirementsDigest is SHA-256 of the normalized requirements, so it does
	// not change with input order or whitespace.
	RequirementsDigest string `json:"requirementsDigest"`
	ModelID            string `json:"modelId"`
	ModelDigest        string `json:"modelDigest"`
	AcceptanceCatalog  string `json:"acceptanceCatalog"`
	// CatalogSHA256 is SHA-256 of the exact catalog bytes read.
	CatalogSHA256 string `json:"catalogSha256"`
}
type Summary struct {
	Requirements            int `json:"requirements"`
	Linked                  int `json:"linked"`
	Incomplete              int `json:"incomplete"`
	ModelElements           int `json:"modelElements"`
	UntracedModelElements   int `json:"untracedModelElements"`
	AcceptanceCases         int `json:"acceptanceCases"`
	PendingAcceptanceCases  int `json:"pendingAcceptanceCases"`
	UnlinkedAcceptanceCases int `json:"unlinkedAcceptanceCases"`
}
type RequirementTrace struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	Statement       string       `json:"statement"`
	Status          string       `json:"status"`
	Gaps            []string     `json:"gaps"`
	ModelLinks      []ElementRef `json:"modelLinks"`
	AcceptanceLinks []CaseRef    `json:"acceptanceLinks"`
}
type ElementRef struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Detail string `json:"detail"`
}
type CaseRef struct {
	CaseID          string `json:"caseId"`
	Name            string `json:"name"`
	ExecutionStatus string `json:"executionStatus"`
}

// ElementTrace lists which requirements link a model element directly.
// TracedViaGuard is set only for model requirements: the requirements that
// link a guard the model declares for it.
type ElementTrace struct {
	Kind           string   `json:"kind"`
	ID             string   `json:"id"`
	Detail         string   `json:"detail"`
	TracedBy       []string `json:"tracedBy"`
	TracedViaGuard []string `json:"tracedViaGuard,omitempty"`
}
type CaseTrace struct {
	CaseID          string   `json:"caseId"`
	Name            string   `json:"name"`
	ExecutionStatus string   `json:"executionStatus"`
	TracedBy        []string `json:"tracedBy"`
}

// Untraced reports whether no requirement reaches the element.
func (e ElementTrace) Untraced() bool { return len(e.TracedBy) == 0 && len(e.TracedViaGuard) == 0 }

type element struct {
	kind, id, detail, guardRequirement string
}

// Case is one entry of the independent acceptance catalog.
type Case struct {
	CaseID, Name, ExecutionStatus string
}

type catalogEntry struct {
	CaseID          *string `json:"caseId"`
	Name            *string `json:"name"`
	ExecutionStatus *string `json:"executionStatus"`
}
type catalogDocument struct {
	Schema    string            `json:"schema"`
	Sequences []json.RawMessage `json:"sequences"`
	Refusals  []json.RawMessage `json:"refusals"`
	Special   []json.RawMessage `json:"special"`
}

// LoadCatalog reads the case IDs from the sequences, refusals and special
// sections of the oracle fixture. Other oracle sections carry no case IDs and
// are not acceptance cases for this purpose. The fixture's own structure is
// owned by the oracle; only the fields read here are checked.
func LoadCatalog(src []byte) ([]Case, error) {
	bad := func(s string) error { return errors.New("invalid acceptance catalog: " + s) }
	if len(src) > 4*MaxBytes || !utf8.Valid(src) {
		return nil, bad("size or encoding")
	}
	var doc catalogDocument
	if err := json.Unmarshal(src, &doc); err != nil {
		return nil, bad("not JSON")
	}
	if doc.Schema != CatalogSchema {
		return nil, bad("unsupported schema")
	}
	seen := map[string]bool{}
	out := []Case{}
	for _, section := range [][]json.RawMessage{doc.Sequences, doc.Refusals, doc.Special} {
		for _, raw := range section {
			var e catalogEntry
			if err := json.Unmarshal(raw, &e); err != nil || e.CaseID == nil || e.Name == nil || e.ExecutionStatus == nil {
				return nil, bad("entry without caseId, name and executionStatus")
			}
			if !caseID(*e.CaseID) {
				return nil, bad("malformed caseId")
			}
			if seen[*e.CaseID] {
				return nil, bad("duplicate caseId " + *e.CaseID)
			}
			if *e.ExecutionStatus != "executable" && *e.ExecutionStatus != "pending" {
				return nil, bad("unknown executionStatus for " + *e.CaseID)
			}
			seen[*e.CaseID] = true
			out = append(out, Case{*e.CaseID, *e.Name, *e.ExecutionStatus})
		}
	}
	if len(out) == 0 {
		return nil, bad("no cases")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CaseID < out[j].CaseID })
	return out, nil
}

func caseID(s string) bool {
	if len(s) < 1 || len(s) > 64 {
		return false
	}
	for _, c := range []byte(s) {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

// RequirementID accepts an uppercase prefix of 1-8 letters, a hyphen and 1-6
// digits, such as UR-001.
func RequirementID(s string) bool {
	p, n, ok := strings.Cut(s, "-")
	if !ok || len(p) < 1 || len(p) > 8 || len(n) < 1 || len(n) > 6 {
		return false
	}
	for _, c := range []byte(p) {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	for _, c := range []byte(n) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func validText(s string, limit int) bool {
	if len(s) > limit || !utf8.ValidString(s) || strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		if r < 0x20 && r != '\t' || r == 0x7f {
			return false
		}
	}
	return true
}

// statementKey folds case and whitespace so near-identical statements collide.
func statementKey(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

func elements(m model.Model) map[string]element {
	out := map[string]element{}
	for _, s := range m.States {
		d := s.Label
		if s.Terminal {
			d += " (terminal)"
		}
		out[s.ID] = element{KindState, s.ID, d, ""}
	}
	for _, t := range m.Transitions {
		out[t.ID] = element{KindTransition, t.ID, t.From + " --" + t.Trigger + "--> " + t.To, ""}
	}
	for _, g := range m.Guards {
		out[g.ID] = element{KindGuard, g.ID, "declares " + g.Requirement, g.Requirement}
	}
	for _, r := range m.Requirements {
		out[r.ID] = element{KindModelRequirement, r.ID, r.Text, ""}
	}
	return out
}

func validKind(k string) bool {
	for _, x := range kindOrder {
		if x == k {
			return true
		}
	}
	return false
}

func kindRank(k string) int {
	for i, x := range kindOrder {
		if x == k {
			return i
		}
	}
	return len(kindOrder)
}

func sortIssues(in []Issue) []Issue {
	sort.Slice(in, func(i, j int) bool {
		a, b := in[i], in[j]
		if a.Requirement != b.Requirement {
			return a.Requirement < b.Requirement
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Target < b.Target
	})
	out := in[:0]
	for i, x := range in {
		if i == 0 || x != in[i-1] {
			out = append(out, x)
		}
	}
	return out
}

// validate checks the document against the model and catalog and collects
// every issue. It never stops at the first problem after decoding succeeds.
func validate(doc wireDocument, m model.Model, els map[string]element, cases map[string]Case) []Issue {
	issues := []Issue{}
	add := func(code, req, target string) { issues = append(issues, Issue{code, req, target}) }
	if doc.FormatVersion != FormatVersion {
		add("unsupported-format-version", "", "")
	}
	if doc.ModelID != m.ID {
		add("model-id-mismatch", "", "")
	}
	if doc.AcceptanceCatalog != CatalogSchema {
		add("catalog-mismatch", "", "")
	}
	if len(doc.Requirements) == 0 {
		add("no-requirements", "", "")
	}
	if len(doc.Requirements) > MaxRequirements {
		add("too-many-requirements", "", "")
		return sortIssues(issues)
	}
	ids := map[string]int{}
	statements := map[string][]string{}
	for _, r := range doc.Requirements {
		id := r.ID
		if !RequirementID(id) {
			add("invalid-requirement-id", id, "")
		}
		ids[id]++
		if !validText(r.Title, MaxTitle) {
			add("invalid-title", id, "")
		}
		if !validText(r.Statement, MaxStatement) {
			add("invalid-statement", id, "")
		} else {
			k := statementKey(r.Statement)
			statements[k] = append(statements[k], id)
		}
		if len(r.ModelLinks) > MaxLinks || len(r.AcceptanceLinks) > MaxLinks {
			add("too-many-links", id, "")
			continue
		}
		seen := map[string]bool{}
		for _, l := range r.ModelLinks {
			target := l.Kind + ":" + l.ID
			switch el, ok := els[l.ID]; {
			case !validKind(l.Kind):
				add("unknown-link-kind", id, target)
			case !ok:
				add("dangling-model-link", id, target)
			case el.kind != l.Kind:
				add("link-kind-mismatch", id, target)
			}
			if seen[target] {
				add("duplicate-model-link", id, target)
			}
			seen[target] = true
		}
		seenCase := map[string]bool{}
		for _, l := range r.AcceptanceLinks {
			if _, ok := cases[l.CaseID]; !ok {
				add("dangling-acceptance-link", id, l.CaseID)
			}
			if seenCase[l.CaseID] {
				add("duplicate-acceptance-link", id, l.CaseID)
			}
			seenCase[l.CaseID] = true
		}
	}
	for id, n := range ids {
		if n > 1 {
			add("duplicate-requirement-id", id, "")
		}
	}
	for _, group := range statements {
		if len(group) < 2 {
			continue
		}
		sorted := append([]string{}, group...)
		sort.Strings(sorted)
		for _, id := range sorted[1:] {
			if id != sorted[0] {
				add("ambiguous-statement", id, sorted[0])
			}
		}
	}
	return sortIssues(issues)
}

// Build validates the requirements document against the model source and the
// acceptance catalog, then derives the trace. A model or catalog that cannot be
// read returns a plain error. Requirements that fail validation return an
// *InvalidError and an empty Report.
func Build(modelSrc, catalogSrc, requirementsSrc []byte) (Report, error) {
	m, err := jsonformat.Decode(modelSrc)
	if err != nil {
		return Report{}, fmt.Errorf("invalid model: %w", err)
	}
	modelDigest, err := canonical.Digest(m)
	if err != nil {
		return Report{}, fmt.Errorf("invalid model: %w", err)
	}
	catalog, err := LoadCatalog(catalogSrc)
	if err != nil {
		return Report{}, err
	}
	cases := map[string]Case{}
	for _, c := range catalog {
		cases[c.CaseID] = c
	}
	els := elements(m)

	var doc wireDocument
	if err := strictjson.Decode(requirementsSrc, &doc, MaxBytes, MaxDepth); err != nil {
		return Report{}, &InvalidError{[]Issue{{Code: "malformed-json"}}}
	}
	if issues := validate(doc, m, els, cases); len(issues) > 0 {
		return Report{}, &InvalidError{issues}
	}

	reqs := make([]RequirementTrace, 0, len(doc.Requirements))
	direct := map[string][]string{}
	viaGuard := map[string][]string{}
	caseBy := map[string][]string{}
	for _, r := range doc.Requirements {
		rt := RequirementTrace{ID: r.ID, Title: r.Title, Statement: r.Statement, Gaps: []string{}, ModelLinks: []ElementRef{}, AcceptanceLinks: []CaseRef{}}
		for _, l := range r.ModelLinks {
			el := els[l.ID]
			rt.ModelLinks = append(rt.ModelLinks, ElementRef{el.kind, el.id, el.detail})
			direct[el.id] = append(direct[el.id], r.ID)
			if el.kind == KindGuard {
				viaGuard[el.guardRequirement] = append(viaGuard[el.guardRequirement], r.ID)
			}
		}
		executable := false
		for _, l := range r.AcceptanceLinks {
			c := cases[l.CaseID]
			rt.AcceptanceLinks = append(rt.AcceptanceLinks, CaseRef{c.CaseID, c.Name, c.ExecutionStatus})
			caseBy[c.CaseID] = append(caseBy[c.CaseID], r.ID)
			executable = executable || c.ExecutionStatus == "executable"
		}
		if len(rt.ModelLinks) == 0 {
			rt.Gaps = append(rt.Gaps, GapNoModelLink)
		}
		if len(rt.AcceptanceLinks) == 0 {
			rt.Gaps = append(rt.Gaps, GapNoAcceptanceLink)
		} else if !executable {
			rt.Gaps = append(rt.Gaps, GapNoExecutableAcceptance)
		}
		sort.Strings(rt.Gaps)
		rt.Status = StatusLinked
		if len(rt.Gaps) > 0 {
			rt.Status = StatusIncomplete
		}
		sort.Slice(rt.ModelLinks, func(i, j int) bool {
			a, b := rt.ModelLinks[i], rt.ModelLinks[j]
			if a.Kind != b.Kind {
				return kindRank(a.Kind) < kindRank(b.Kind)
			}
			return a.ID < b.ID
		})
		sort.Slice(rt.AcceptanceLinks, func(i, j int) bool { return rt.AcceptanceLinks[i].CaseID < rt.AcceptanceLinks[j].CaseID })
		reqs = append(reqs, rt)
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })

	rep := Report{
		ReportVersion: ReportVersion, Mode: "example-only", Authority: "none", Notice: Notice,
		Requirements: reqs, ModelElements: []ElementTrace{}, AcceptanceCases: []CaseTrace{},
	}
	ids := make([]string, 0, len(els))
	for id := range els {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := els[ids[i]], els[ids[j]]
		if a.kind != b.kind {
			return kindRank(a.kind) < kindRank(b.kind)
		}
		return a.id < b.id
	})
	for _, id := range ids {
		el := els[id]
		et := ElementTrace{Kind: el.kind, ID: el.id, Detail: el.detail, TracedBy: uniq(direct[id])}
		if el.kind == KindModelRequirement {
			et.TracedViaGuard = uniq(viaGuard[id])
		}
		if et.Untraced() {
			rep.Summary.UntracedModelElements++
		}
		rep.ModelElements = append(rep.ModelElements, et)
	}
	for _, c := range catalog {
		ct := CaseTrace{c.CaseID, c.Name, c.ExecutionStatus, uniq(caseBy[c.CaseID])}
		if len(ct.TracedBy) == 0 {
			rep.Summary.UnlinkedAcceptanceCases++
		}
		if c.ExecutionStatus == "pending" {
			rep.Summary.PendingAcceptanceCases++
		}
		rep.AcceptanceCases = append(rep.AcceptanceCases, ct)
	}
	for _, r := range reqs {
		if r.Status == StatusLinked {
			rep.Summary.Linked++
		} else {
			rep.Summary.Incomplete++
		}
	}
	rep.Summary.Requirements = len(reqs)
	rep.Summary.ModelElements = len(rep.ModelElements)
	rep.Summary.AcceptanceCases = len(rep.AcceptanceCases)

	norm, err := json.Marshal(normalizedRequirements(reqs))
	if err != nil {
		return Report{}, err
	}
	catSum := sha256.Sum256(catalogSrc)
	reqSum := sha256.Sum256(norm)
	rep.Inputs = Inputs{
		RequirementsFormat: FormatVersion, RequirementsDigest: hex.EncodeToString(reqSum[:]),
		ModelID: m.ID, ModelDigest: modelDigest,
		AcceptanceCatalog: CatalogSchema, CatalogSHA256: hex.EncodeToString(catSum[:]),
	}
	return rep, nil
}

// normalizedRequirements is the order-independent content the requirements
// digest covers: IDs, text and sorted links, without derived details.
func normalizedRequirements(reqs []RequirementTrace) any {
	type link struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	}
	type req struct {
		ID         string   `json:"id"`
		Title      string   `json:"title"`
		Statement  string   `json:"statement"`
		Model      []link   `json:"modelLinks"`
		Acceptance []string `json:"acceptanceLinks"`
	}
	out := []req{}
	for _, r := range reqs {
		x := req{r.ID, r.Title, r.Statement, []link{}, []string{}}
		for _, l := range r.ModelLinks {
			x.Model = append(x.Model, link{l.Kind, l.ID})
		}
		for _, c := range r.AcceptanceLinks {
			x.Acceptance = append(x.Acceptance, c.CaseID)
		}
		out = append(out, x)
	}
	return struct {
		Format       string `json:"formatVersion"`
		Requirements []req  `json:"requirements"`
	}{FormatVersion, out}
}

func uniq(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	j := 0
	for i, s := range out {
		if i == 0 || s != out[i-1] {
			out[j] = s
			j++
		}
	}
	return out[:j]
}
