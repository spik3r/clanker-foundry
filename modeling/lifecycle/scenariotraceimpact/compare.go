package scenariotraceimpact

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotrace"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

func hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func retain(request, output []byte) retained {
	return retained{hash(request), string(output), hash(output)}
}

func nativeError(err error) error {
	switch {
	case errors.Is(err, scenariotrace.ErrLimit), errors.Is(err, scenariodiff.ErrLimit), errors.Is(err, traceimpact.ErrReportLimit):
		return ErrLimit
	case errors.Is(err, scenariotrace.ErrInvalid), errors.Is(err, scenariotrace.ErrBinding), errors.Is(err, scenariodiff.ErrInvalid), errors.Is(err, traceimpact.ErrInvalid):
		return ErrInvalid
	default:
		return ErrBoundary
	}
}

func readTrace(raw string) (snapshot, traceReport, retained, error) {
	var source snapshot
	var trace traceReport
	var record retained
	output, err := scenariotrace.Check([]byte(raw))
	if err != nil {
		return source, trace, record, nativeError(err)
	}
	if len(output) > MaxTraceReportBytes {
		return source, trace, record, ErrLimit
	}
	if json.Unmarshal([]byte(raw), &source) != nil || json.Unmarshal(output, &trace) != nil {
		return source, trace, record, ErrBoundary
	}
	in := trace.Inputs
	if trace.FormatVersion != scenariotrace.ReportVersion || trace.Mode != "example-only" || trace.Authority != "none" || trace.ExecutionStatus != "not-established" || trace.ReviewStatus != "not-established" || trace.Requirements == nil || trace.Cases == nil ||
		in.ModelSHA256 != hash([]byte(source.Model)) || in.RequirementsSHA256 != hash([]byte(source.Requirements)) || in.CatalogSHA256 != hash([]byte(source.Catalog)) || in.SuiteSHA256 != hash([]byte(source.Suite)) || in.BindingsSHA256 != hash([]byte(source.Bindings)) {
		return source, trace, record, ErrBoundary
	}
	return source, trace, retain([]byte(raw), output), nil
}

func token(raw string) string { return strings.Trim(raw, " \t\r\n") }

func scenarioRequest(before, after snapshot) []byte {
	// Each token has already passed native validation. Marshal(RawMessage) would
	// compact its whitespace and change the source identity reported by diff.
	var b bytes.Buffer
	b.WriteString(`{"formatVersion":"` + scenariodiff.FormatVersion + `","beforeModel":`)
	b.WriteString(token(before.Model))
	b.WriteString(`,"beforeSuite":`)
	b.WriteString(token(before.Suite))
	b.WriteString(`,"afterModel":`)
	b.WriteString(token(after.Model))
	b.WriteString(`,"afterSuite":`)
	b.WriteString(token(after.Suite))
	b.WriteString("}\n")
	return b.Bytes()
}

func declaredTraceRequest(before, after snapshot) ([]byte, error) {
	type side struct {
		Model        string `json:"model"`
		Catalog      string `json:"catalog"`
		Requirements string `json:"requirements"`
	}
	q := struct {
		FormatVersion string `json:"formatVersion"`
		Before        side   `json:"before"`
		After         side   `json:"after"`
	}{traceimpact.RequestVersion, side{before.Model, before.Catalog, before.Requirements}, side{after.Model, after.Catalog, after.Requirements}}
	// Strings preserve the complete decoded source bytes, including whitespace.
	return json.Marshal(q)
}

func reportsAgree(before, after traceReport, bs, as snapshot, diff scenariodiff.Report, impact traceimpact.Report, diffRequest []byte) bool {
	b, a := before.Inputs, after.Inputs
	if diff.DiffVersion != scenariodiff.Version || diff.Mode != "example-only" || diff.Authority != "none" || diff.Sources == nil ||
		impact.FormatVersion != traceimpact.ReportVersion || impact.Mode != "example-only" || impact.Authority != "none" {
		return false
	}
	if diff.BeforeModelDigest != b.ModelDigest || diff.AfterModelDigest != a.ModelDigest || diff.BeforeSuiteDigest != b.ScenarioDigest || diff.AfterSuiteDigest != a.ScenarioDigest ||
		impact.Before.ModelDigest != b.ModelDigest || impact.After.ModelDigest != a.ModelDigest || impact.Before.RequirementsDigest != b.RequirementsDigest || impact.After.RequirementsDigest != a.RequirementsDigest ||
		impact.Before.CatalogSHA256 != b.CatalogSHA256 || impact.After.CatalogSHA256 != a.CatalogSHA256 {
		return false
	}
	s := diff.Sources
	if s.Request != hash(diffRequest) || s.BeforeModel != hash([]byte(token(bs.Model))) || s.AfterModel != hash([]byte(token(as.Model))) || s.BeforeSuite != hash([]byte(token(bs.Suite))) || s.AfterSuite != hash([]byte(token(as.Suite))) {
		return false
	}
	for _, c := range diff.Cases {
		if c.Kind != scenariodiff.Added && c.Kind != scenariodiff.Removed && c.Kind != scenariodiff.Modified {
			return false
		}
	}
	return true
}

// Check compares validated original trace requests. All reports describe supplied
// declarations and observations; none attests execution, coverage or approval.
// No I/O occurs. Callers must not mutate raw during a call; output owns its bytes.
func Check(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var q request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 4) != nil || q.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	if len(q.Before) > scenariotrace.MaxRequestBytes || len(q.After) > scenariotrace.MaxRequestBytes {
		return nil, ErrLimit
	}
	bs, before, br, err := readTrace(q.Before)
	if err != nil {
		return nil, err
	}
	as, after, ar, err := readTrace(q.After)
	if err != nil {
		return nil, err
	}
	diffRequest := scenarioRequest(bs, as)
	impactRequest, err := declaredTraceRequest(bs, as)
	if err != nil {
		return nil, ErrBoundary
	}
	if len(diffRequest) > scenariodiff.MaxRequestBytes || len(impactRequest) > traceimpact.MaxRequestBytes {
		return nil, ErrLimit
	}
	diffJSON, err := scenariodiff.Execute(diffRequest)
	if err != nil {
		return nil, nativeError(err)
	}
	impactJSON, err := traceimpact.Execute(impactRequest)
	if err != nil {
		return nil, nativeError(err)
	}
	var diff scenariodiff.Report
	var impact traceimpact.Report
	if json.Unmarshal(diffJSON, &diff) != nil || json.Unmarshal(impactJSON, &impact) != nil || !reportsAgree(before, after, bs, as, diff, impact, diffRequest) {
		return nil, ErrBoundary
	}
	r := report{
		FormatVersion: ReportVersion, Mode: "example-only", Authority: "none", ExecutionStatus: "not-established", ReviewStatus: "not-established",
		RequestSHA256: hash(raw), Before: br, After: ar, ScenarioDiff: retain(diffRequest, diffJSON), TraceImpact: retain(impactRequest, impactJSON),
		RequestSourceChanged: q.Before != q.After, SourceChanges: sourceChanges(bs, as), ReviewRequired: q.Before != q.After,
		Candidates: project(before, after, diff, impact),
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if enc.Encode(r) != nil {
		return nil, ErrBoundary
	}
	if out.Len() > MaxOutputBytes {
		return nil, ErrLimit
	}
	return out.Bytes(), nil
}
