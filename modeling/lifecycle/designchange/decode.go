package designchange

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// encodedSize validates canonical padded base64 and computes the exact decoded
// size without allocating payload bytes. Strict pad bits are checked here too.
func encodedSize(s string, limit int) (int, error) {
	if len(s) > 4*((limit+2)/3) {
		return 0, ErrLimit
	}
	if len(s)%4 != 0 {
		return 0, ErrInvalid
	}
	pad := 0
	if strings.HasSuffix(s, "=") {
		pad++
	}
	if strings.HasSuffix(s, "==") {
		pad++
	}
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	for i := 0; i < len(s)-pad; i++ {
		if strings.IndexByte(alphabet, s[i]) < 0 {
			return 0, ErrInvalid
		}
	}
	for i := len(s) - pad; i < len(s); i++ {
		if s[i] != '=' {
			return 0, ErrInvalid
		}
	}
	n := len(s)/4*3 - pad
	if n < 0 {
		return 0, ErrInvalid
	}
	if n > limit {
		return 0, ErrLimit
	}
	if pad > 0 {
		if len(s) < 4 {
			return 0, ErrInvalid
		}
		v := strings.IndexByte(alphabet, s[len(s)-pad-1])
		if pad == 2 && v&15 != 0 || pad == 1 && v&3 != 0 {
			return 0, ErrInvalid
		}
	}
	return n, nil
}
func decode64(s string) ([]byte, error) {
	b, e := base64.StdEncoding.Strict().DecodeString(s)
	if e != nil || base64.StdEncoding.EncodeToString(b) != s {
		return nil, ErrInvalid
	}
	return b, nil
}
func carrierDepth(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	depth := 0
	for {
		v, e := d.Token()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return ErrInvalid
		}
		if x, ok := v.(json.Delim); ok {
			switch x {
			case '{', '[':
				depth++
				if depth > MaxCarrierDepth {
					return ErrLimit
				}
			case '}', ']':
				depth--
			}
		}
	}
}
func decodeBuild(c carrierDTO) ([]byte, []byte, []artifactbundle.File, error) {
	rn, e := encodedSize(*c.RequestBase64, MaxRequestBytes)
	if e != nil {
		return nil, nil, nil, e
	}
	sn, e := encodedSize(*c.SuiteBase64, MaxSuiteBytes)
	if e != nil {
		return nil, nil, nil, e
	}
	names, set, max := sourceNames()
	src := *c.Sources
	if len(src) > len(names) {
		return nil, nil, nil, ErrLimit
	}
	if rn == 0 || sn == 0 || len(src) != len(names) {
		return nil, nil, nil, ErrInvalid
	}
	seen := make(map[string]bool, len(src))
	for _, s := range src {
		if len(s.Name) > max || !set[s.Name] || seen[s.Name] {
			return nil, nil, nil, ErrInvalid
		}
		seen[s.Name] = true
	}
	total := 0
	for _, s := range src {
		n, e := encodedSize(s.DataBase64, MaxSourceBytes)
		if e != nil {
			return nil, nil, nil, e
		}
		total += n
		if total > MaxSourcesBytes {
			return nil, nil, nil, ErrLimit
		}
	}
	// All decoded byte budgets have passed before any bulk payload allocation.
	request, e := decode64(*c.RequestBase64)
	if e != nil {
		return nil, nil, nil, e
	}
	suite, e := decode64(*c.SuiteBase64)
	if e != nil {
		return nil, nil, nil, e
	}
	sources := make([]artifactbundle.File, len(src))
	for i, s := range src {
		b, e := decode64(s.DataBase64)
		if e != nil {
			return nil, nil, nil, e
		}
		sources[i] = artifactbundle.File{Name: s.Name, Data: b}
	}
	return request, suite, sources, nil
}

// Execute accepts the strict bounded base64 carrier. Build emits raw ZIP; verify
// emits only a consistency receipt, with no suite validation or execution claim.
func Execute(raw []byte) ([]byte, error) {
	if len(raw) > MaxCarrierBytes {
		return nil, ErrLimit
	}
	if e := carrierDepth(raw); e != nil {
		return nil, e
	}
	var c carrierDTO
	if strictjson.Decode(raw, &c, MaxCarrierBytes, MaxCarrierDepth) != nil || c.FormatVersion != FormatVersion {
		return nil, ErrInvalid
	}
	switch c.Operation {
	case "build":
		if c.RequestBase64 == nil || c.SuiteBase64 == nil || c.Sources == nil || c.ArchiveBase64 != nil || c.Expected != nil {
			return nil, ErrInvalid
		}
		request, suite, sources, e := decodeBuild(c)
		if e != nil {
			return nil, e
		}
		return Build(request, suite, sources)
	case "verify":
		if c.RequestBase64 != nil || c.SuiteBase64 != nil || c.Sources != nil || c.ArchiveBase64 == nil || c.Expected == nil || !validExpected(*c.Expected) {
			return nil, ErrInvalid
		}
		if _, e := encodedSize(*c.ArchiveBase64, MaxArchiveBytes); e != nil {
			return nil, e
		}
		archive, e := decode64(*c.ArchiveBase64)
		if e != nil {
			return nil, e
		}
		if e = Verify(archive, *c.Expected); e != nil {
			return nil, e
		}
		return encode(verificationDTO{VerificationVersion, true, "example-only", "none", "unvalidated and unexecuted", "not inspected"})
	default:
		return nil, ErrInvalid
	}
}
