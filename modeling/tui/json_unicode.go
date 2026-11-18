package tui

import (
	"encoding/json"
	"errors"
	"strconv"
)

// checkJSONUnicode keeps encoding/json from replacing malformed text. Syntax
// errors remain the caller's responsibility; only valid JSON reaches look-ahead.
func checkJSONUnicode(data []byte) error {
	if !json.Valid(data) {
		return nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] != '"' {
			continue
		}
		i++
		for i < len(data) && data[i] != '"' {
			if data[i] != '\\' {
				i++
				continue
			}
			i++
			if data[i] != 'u' {
				i++
				continue
			}
			n, err := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
			if err != nil {
				return errors.New("invalid Unicode escape in JSON")
			}
			i += 5
			if n >= 0xd800 && n <= 0xdbff {
				if i+6 > len(data) || data[i] != '\\' || data[i+1] != 'u' {
					return errors.New("unpaired surrogate escape in JSON")
				}
				low, err := strconv.ParseUint(string(data[i+2:i+6]), 16, 16)
				if err != nil || low < 0xdc00 || low > 0xdfff {
					return errors.New("unpaired surrogate escape in JSON")
				}
				i += 6
			} else if n >= 0xdc00 && n <= 0xdfff {
				return errors.New("unpaired surrogate escape in JSON")
			}
		}
	}
	return nil
}
