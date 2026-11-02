package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvalidRequests(t *testing.T) {
	for _, input := range []string{"{}", `{"candidate":"","candidate":"","suite":""}`, `{"unknown":true}`, "{"} {
		var out, diag bytes.Buffer
		if run([]string{"export"}, bytes.NewBufferString(input), &out, &diag) == 0 || out.Len() != 0 {
			t.Fatal("invalid export accepted")
		}
	}
	var out, diag bytes.Buffer
	if run([]string{"verify"}, bytes.NewBufferString(`{"archive":"","candidateSHA256":"","suiteSHA256":""}`), &out, &diag) == 0 || out.Len() != 0 {
		t.Fatal("invalid verify accepted")
	}
}
func TestRejectsSymlinkSource(t *testing.T) {
	root := t.TempDir()
	if e := os.WriteFile(filepath.Join(root, "go.mod"), []byte("placeholder"), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink("/tmp", filepath.Join(root, "model")); e != nil {
		t.Fatal(e)
	}
	old, e := os.Getwd()
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Chdir(root); e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(old)
	if _, e = readSources(); e == nil {
		t.Fatal("accepted missing/symlink inventory")
	}
}

func TestExportVerifyRoundTrip(t *testing.T) {
	old, e := os.Getwd()
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Chdir("../.."); e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(old)
	candidate, e := os.ReadFile("examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	suite, e := os.ReadFile("scenarios/testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	req, e := json.Marshal(map[string]string{"candidate": base64.StdEncoding.EncodeToString(candidate), "suite": base64.StdEncoding.EncodeToString(suite)})
	if e != nil {
		t.Fatal(e)
	}
	var archive, diag bytes.Buffer
	if run([]string{"export"}, bytes.NewReader(req), &archive, &diag) != 0 {
		t.Fatalf("export failed: %s", diag.String())
	}
	c := sha256.Sum256(candidate)
	s := sha256.Sum256(suite)
	req, e = json.Marshal(map[string]string{"archive": base64.StdEncoding.EncodeToString(archive.Bytes()), "candidateSHA256": hex.EncodeToString(c[:]), "suiteSHA256": hex.EncodeToString(s[:])})
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	diag.Reset()
	if run([]string{"verify"}, bytes.NewReader(req), &out, &diag) != 0 {
		t.Fatalf("verify failed: %s", diag.String())
	}
	if run([]string{"verify"}, bytes.NewReader(req), shortWriter{}, &diag) == 0 {
		t.Fatal("accepted short write")
	}
	if !strings.Contains(out.String(), "unvalidated and unexecuted") {
		t.Fatal("missing status")
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}
func TestRejectsNonregularAndOversizedSource(t *testing.T) {
	old, e := os.Getwd()
	if e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(old)
	for _, kind := range []string{"directory", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			p := filepath.Join(root, "go.mod")
			if kind == "directory" {
				if e := os.Mkdir(p, 0755); e != nil {
					t.Fatal(e)
				}
			} else if e := os.WriteFile(p, make([]byte, (1<<20)+1), 0644); e != nil {
				t.Fatal(e)
			}
			if e := os.Chdir(root); e != nil {
				t.Fatal(e)
			}
			defer os.Chdir(old)
			if _, e := readSources(); e == nil {
				t.Fatal("accepted invalid source")
			}
		})
	}
}
