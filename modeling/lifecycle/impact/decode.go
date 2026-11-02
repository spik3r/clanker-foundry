package impact

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

func hash(raw []byte) *string {
	h := sha256.Sum256(raw)
	s := hex.EncodeToString(h[:])
	return &s
}

// Execute validates both models and recomputes the comparison before encoding.
// It does not accept caller-constructed reports or automatically rebind evidence.
func Execute(raw []byte) ([]byte, error) {
	d, err := decodeRequest(raw)
	if err != nil {
		return nil, err
	}
	r, err := Compare(d.before, d.after, d.bindings)
	if err != nil {
		return nil, err
	}
	r.RequestDigest = hash(raw)
	r.BeforeSourceDigest = hash(d.beforeSource)
	r.AfterSourceDigest = hash(d.afterSource)
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil || out.Len() > MaxReportBytes {
		return nil, ErrReportLimit
	}
	return out.Bytes(), nil
}

type requestDTO struct {
	FormatVersion string          `json:"formatVersion"`
	Before        json.RawMessage `json:"before"`
	After         json.RawMessage `json:"after"`
	Bindings      []Binding       `json:"bindings"`
}

type request struct {
	before, after             model.Model
	beforeSource, afterSource []byte
	bindings                  []Binding
}

// Each nested model passes its own adapter too: RawMessage is not a shortcut
// around required fields, unsupported profiles, Unicode or per-model limits.
func decodeRequest(raw []byte) (request, error) {
	var d requestDTO
	if err := strictjson.Decode(raw, &d, MaxSourceBytes, MaxDepth); err != nil || d.FormatVersion != FormatVersion || len(d.Bindings) > MaxBindings {
		return request{}, ErrInvalidInput
	}
	before, err := jsonformat.Decode(d.Before)
	if err != nil {
		return request{}, ErrInvalidInput
	}
	after, err := jsonformat.Decode(d.After)
	if err != nil {
		return request{}, ErrInvalidInput
	}
	if err := validateBindings(d.Bindings); err != nil {
		return request{}, err
	}
	return request{before, after, d.Before, d.After, d.Bindings}, nil
}

func validateBindings(bindings []Binding) error {
	if len(bindings) > MaxBindings {
		return ErrInvalidInput
	}
	seen := map[string]bool{}
	for _, b := range bindings {
		if !model.Identifier(b.ID) || len(b.ModelDigest) != 64 {
			return ErrInvalidInput
		}
		for _, c := range []byte(b.ModelDigest) {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
				return ErrInvalidInput
			}
		}
		if seen[b.ID] {
			return ErrInvalidInput
		}
		seen[b.ID] = true
	}
	return nil
}
