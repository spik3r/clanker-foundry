package scenariodiff

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

type request struct {
	FormatVersion string          `json:"formatVersion"`
	BeforeModel   json.RawMessage `json:"beforeModel"`
	BeforeSuite   json.RawMessage `json:"beforeSuite"`
	AfterModel    json.RawMessage `json:"afterModel"`
	AfterSuite    json.RawMessage `json:"afterSuite"`
}

// validator is private and ONLY for fixture validation/identity. The binding is
// derived from a validated input model, never asserted as compiled-code identity.
// Existing Runner requires a callback even for its pure Decode/Validate/Digest
// methods. A panic sentinel makes accidentally executing this adapter fail loud;
// it must never be exported, passed elsewhere, or used with Run/Execute.
func validator(input model.Model) (scenarios.Runner, string, error) {
	normalized, err := model.Normalize(input)
	if err != nil {
		return scenarios.Runner{}, "", ErrInvalid
	}
	if _, err = jsonformat.Encode(normalized); err != nil {
		return scenarios.Runner{}, "", ErrInvalid
	}
	digest, err := canonical.Digest(normalized)
	if err != nil {
		return scenarios.Runner{}, "", ErrInvalid
	}
	runner, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: model.ProfileVersion, ModelDigest: digest},
		func(lifecycle.Task, lifecycle.Command) lifecycle.Result {
			panic("scenario comparison must not execute")
		})
	if err != nil {
		return scenarios.Runner{}, "", ErrInvalid
	}
	return runner, digest, nil
}

// Execute compares strict bounded input declarations. It does not execute their
// commands, authenticate expectations, or encode a caller-supplied report.
func Execute(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var req request
	if err := strictjson.Decode(raw, &req, MaxRequestBytes, 66); err != nil || req.FormatVersion != FormatVersion {
		return nil, ErrInvalid
	}
	bm, bs, err := decodeSide(req.BeforeModel, req.BeforeSuite)
	if err != nil {
		return nil, ErrInvalid
	}
	am, as, err := decodeSide(req.AfterModel, req.AfterSuite)
	if err != nil {
		return nil, ErrInvalid
	}
	report, err := Compare(bm, bs, am, as)
	if err != nil {
		return nil, err
	}
	report.Sources = &SourceDigests{Request: hash(raw), BeforeModel: hash(req.BeforeModel), BeforeSuite: hash(req.BeforeSuite), AfterModel: hash(req.AfterModel), AfterSuite: hash(req.AfterSuite)}
	return encode(report)
}

func decodeSide(modelBytes, suiteBytes []byte) (model.Model, scenarios.Suite, error) {
	m, err := jsonformat.Decode(modelBytes)
	if err != nil {
		return model.Model{}, scenarios.Suite{}, ErrInvalid
	}
	runner, _, err := validator(m)
	if err != nil {
		return model.Model{}, scenarios.Suite{}, ErrInvalid
	}
	s, err := runner.Decode(suiteBytes)
	if err != nil {
		return model.Model{}, scenarios.Suite{}, ErrInvalid
	}
	return m, s, nil
}

func encode(report Report) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(report); err != nil {
		return nil, ErrInvalid
	}
	if buffer.Len() > MaxReportBytes {
		return nil, ErrLimit
	}
	return buffer.Bytes(), nil
}

func hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
