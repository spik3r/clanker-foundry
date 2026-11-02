package modelmerge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

type request struct {
	FormatVersion string          `json:"formatVersion"`
	Base          json.RawMessage `json:"base"`
	Left          json.RawMessage `json:"left"`
	Right         json.RawMessage `json:"right"`
}

// Execute recomputes a merge from strict bounded JSON. It never writes files or
// encodes a caller-supplied report as if it were an observed result.
func Execute(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var req request
	if err := strictjson.Decode(raw, &req, MaxRequestBytes, 66); err != nil || req.FormatVersion != FormatVersion {
		return nil, ErrInvalid
	}
	base, err := jsonformat.Decode(req.Base)
	if err != nil {
		return nil, ErrInvalid
	}
	left, err := jsonformat.Decode(req.Left)
	if err != nil {
		return nil, ErrInvalid
	}
	right, err := jsonformat.Decode(req.Right)
	if err != nil {
		return nil, ErrInvalid
	}
	result, err := Merge(base, left, right)
	if err != nil {
		return nil, err
	}
	result.Sources = &SourceDigests{Request: hash(raw), Base: hash(req.Base), Left: hash(req.Left), Right: hash(req.Right)}
	var candidate json.RawMessage
	if result.Candidate != nil {
		candidate, err = jsonformat.Encode(*result.Candidate)
		if err != nil {
			return nil, ErrInvalid
		}
	}
	dto := struct {
		MergeVersion    string          `json:"mergeVersion"`
		ProfileVersion  string          `json:"profileVersion"`
		Mode            string          `json:"mode"`
		Authority       string          `json:"authority"`
		BaseDigest      string          `json:"baseDigest"`
		LeftDigest      string          `json:"leftDigest"`
		RightDigest     string          `json:"rightDigest"`
		Sources         *SourceDigests  `json:"sources"`
		Status          string          `json:"status"`
		CandidateDigest *string         `json:"candidateDigest"`
		Candidate       json.RawMessage `json:"candidate"`
		Conflicts       []Conflict      `json:"conflicts"`
	}{result.MergeVersion, result.ProfileVersion, result.Mode, result.Authority, result.BaseDigest, result.LeftDigest, result.RightDigest, result.Sources, result.Status, result.CandidateDigest, candidate, result.Conflicts}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(dto); err != nil {
		return nil, ErrInvalid
	}
	if buffer.Len() > MaxResultBytes {
		return nil, ErrLimit
	}
	return buffer.Bytes(), nil
}

func hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
