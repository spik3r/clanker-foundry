package changerehearsal

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

func TestCheckMatchesIndependentFrozenContract(t *testing.T) { assertFrozenCheck(t, Check) }
func TestCheckStrictResourcesAndIdentityClasses(t *testing.T) {
	d := frozenObservations(t)
	valid := []byte(d.Vectors[0].Request)
	for _, v := range d.Vectors[4:] {
		if raw, err := Check([]byte(v.Request)); !errors.Is(err, ErrPrecondition) || raw != nil {
			t.Fatal("identity class", v.ID, err)
		}
	}
	for _, key := range []string{"patchRequest", "baselineProposalRequest", "candidateSuite", "candidateManifest", "expectedCandidateDigest"} {
		var q map[string]any
		_ = json.Unmarshal(valid, &q)
		delete(q, key)
		bad, _ := json.Marshal(q)
		if raw, err := Check(bad); !errors.Is(err, ErrInvalid) || raw != nil {
			t.Fatal("missing field", key, err)
		}
	}
	for key, limit := range map[string]int{"patchRequest": modelpatch.MaxRequestBytes, "baselineProposalRequest": proposalreplay.MaxRequestBytes, "candidateSuite": 1 << 20, "candidateManifest": 64 << 10} {
		var q map[string]any
		_ = json.Unmarshal(valid, &q)
		q[key] = string(bytes.Repeat([]byte("x"), limit+1))
		bad, _ := json.Marshal(q)
		if raw, err := Check(bad); !errors.Is(err, ErrLimit) || raw != nil {
			t.Fatal("native bound", key, err)
		}
	}
	for _, bad := range [][]byte{
		bytes.Replace(valid, []byte(`"patchRequest":`), []byte(`"PATCHREQUEST":`), 1),
		bytes.Replace(valid, []byte(`"patchRequest":`), []byte(`"patchRequest":null,"patchRequest":`), 1),
	} {
		if raw, err := Check(bad); err == nil || raw != nil {
			t.Fatal("strict shape accepted")
		}
	}
}
func TestConcurrentCheckPreservesSourceBytes(t *testing.T) {
	raw := []byte(frozenObservations(t).Vectors[0].Request)
	saved := bytes.Clone(raw)
	expected, err := Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			out, err := Check(raw)
			if err != nil || !bytes.Equal(out, expected) {
				t.Errorf("concurrent check %v", err)
			}
		}()
	}
	group.Wait()
	if !bytes.Equal(raw, saved) {
		t.Fatal("input changed")
	}
}
