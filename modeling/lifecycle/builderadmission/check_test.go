package builderadmission

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFrozenAdmission(t *testing.T) { assertFrozenAdmission(t, Check) }
func candidateRoot(t *testing.T, content []byte) string {
	t.Helper()
	root := frozenRoot(t)
	if os.WriteFile(filepath.Join(root, "candidate.go"), content, 0600) != nil {
		t.Fatal("candidate fixture")
	}
	return root
}
func TestExactNamesDoNotReadCaseAliases(t *testing.T) {
	for _, name := range []string{"contract.go", "candidate.go"} {
		t.Run(name, func(t *testing.T) {
			text := []byte("opaque candidate")
			root := candidateRoot(t, text)
			alias := "Contract.go"
			if name == "candidate.go" {
				alias = "Candidate.go"
			}
			if err := os.Rename(filepath.Join(root, name), filepath.Join(root, alias)); err != nil {
				t.Fatal(err)
			}
			_, resolution := os.Lstat(filepath.Join(root, name))
			out, err := Check(root, hash(text))
			if name != "candidate.go" && resolution == nil {
				if out != nil || !errors.Is(err, ErrBoundary) {
					t.Fatal("native read would resolve an unexpected child", err)
				}
				return
			}
			if resolution != nil && !os.IsNotExist(resolution) {
				t.Fatal(resolution)
			}
			var r report
			if err != nil || json.Unmarshal(out, &r) != nil || r.Outcome != "blocked" {
				t.Fatal("exact census names", err)
			}
			if name == "candidate.go" && (r.Candidate.Status != "missing" || r.Candidate.ObservedSHA256 != nil) {
				t.Fatal("unexpected candidate alias read")
			}
		})
	}
}
func TestCensusAndCandidateInclusiveBounds(t *testing.T) {
	text := bytes.Repeat([]byte{0xff}, MaxCandidateBytes)
	root := candidateRoot(t, text)
	out, err := Check(root, hash(text))
	var r report
	if err != nil || json.Unmarshal(out, &r) != nil || r.Outcome != "matched" || r.Candidate.ObservedBytes == nil || *r.Candidate.ObservedBytes != MaxCandidateBytes {
		t.Fatal("candidate byte limit is inclusive and opaque", err)
	}
	for n := 0; n < MaxEntries-6; n++ {
		if os.WriteFile(filepath.Join(root, fmt.Sprintf("extra-%02d", n)), nil, 0000) != nil {
			t.Fatal("extra fixture")
		}
	}
	out, err = Check(root, hash(text))
	if err != nil || json.Unmarshal(out, &r) != nil || r.Outcome != "blocked" || len(r.Frontier) != MaxEntries {
		t.Fatal("complete maximum census", err)
	}
	if os.WriteFile(filepath.Join(root, "one-too-many"), nil, 0000) != nil {
		t.Fatal("extra fixture")
	}
	if out, err := Check(root, hash(text)); out != nil || !errors.Is(err, ErrBoundary) {
		t.Fatal("incomplete census", err)
	}
}
func TestCandidateReadFailureAndUnlistedUnreadableFile(t *testing.T) {
	root := candidateRoot(t, []byte("opaque"))
	path := filepath.Join(root, "candidate.go")
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0600)
	if b, err := os.ReadFile(path); err == nil {
		_ = b
		t.Skip("host privileges bypass candidate permission bits")
	}
	out, err := Check(root, hash([]byte("opaque")))
	var r report
	if err != nil || json.Unmarshal(out, &r) != nil || r.Outcome != "blocked" || r.Candidate.Status != "read-failed" || r.Candidate.ObservedSHA256 != nil {
		t.Fatal("complete blocked candidate observation", err)
	}
	if os.Chmod(path, 0600) != nil {
		t.Fatal("restore fixture")
	}
	extra := filepath.Join(root, "unknown.go")
	if os.WriteFile(extra, []byte("unreadable unexpected bytes"), 0000) != nil {
		t.Fatal("extra fixture")
	}
	defer os.Chmod(extra, 0600)
	out, err = Check(root, hash([]byte("opaque")))
	if err != nil || json.Unmarshal(out, &r) != nil || r.Outcome != "blocked" || r.Candidate.Status != "matched" {
		t.Fatal("unexpected content must not be read", err)
	}
}
func TestConcurrentOwnedResultsPreserveFiles(t *testing.T) {
	root := candidateRoot(t, []byte("no Go parsing"))
	pin := hash([]byte("no Go parsing"))
	before := map[string][]byte{}
	modes := map[string]os.FileMode{}
	for _, name := range allowedNames {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		before[name] = b
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		modes[name] = info.Mode()
	}
	expected, err := Check(root, pin)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			out, err := Check(root, pin)
			if err != nil || !bytes.Equal(out, expected) {
				t.Error("concurrent output", err)
			}
			if len(out) > 0 {
				out[0] = '!'
			}
		}()
	}
	group.Wait()
	out, err := Check(root, pin)
	if err != nil || !bytes.Equal(out, expected) {
		t.Fatal("shared returned storage")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 6 {
		t.Fatal("file frontier changed")
	}
	for name, b := range before {
		after, err := os.ReadFile(filepath.Join(root, name))
		info, e := os.Stat(filepath.Join(root, name))
		if err != nil || e != nil || !bytes.Equal(b, after) || info.Mode() != modes[name] {
			t.Fatal("input changed", name)
		}
	}
}
