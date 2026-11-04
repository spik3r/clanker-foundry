package operatorcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
)

func revisionV2Fixture(t *testing.T) []byte {
	t.Helper()
	var data struct{ Vectors []struct{ Request string } }
	if json.Unmarshal(read(t, "../../revisioncheck/testdata/v2-observations.json"), &data) != nil || len(data.Vectors) == 0 {
		t.Fatal("independent v2 fixture")
	}
	return []byte(data.Vectors[0].Request)
}
func TestFoundryRevisionsSelectsExplicitV2AndKeepsLegacyV1(t *testing.T) {
	v2 := revisionV2Fixture(t)
	expected, err := revisioncheck.CheckV2(v2)
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag := run([]string{"revisions"}, bytes.NewReader(v2))
	if code != 0 || diag != "" || !bytes.Equal(out, expected) {
		t.Fatal("v2 route changed native result", code, diag)
	}
	var q revisioncheck.Request
	if json.Unmarshal(v2, &q) != nil {
		t.Fatal("v2 request")
	}
	q.FormatVersion = revisioncheck.RequestVersion
	v1, _ := json.Marshal(q)
	expected, err = revisioncheck.Check(v1)
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag = run([]string{"revisions"}, bytes.NewReader(v1))
	if code != 0 || diag != "" || !bytes.Equal(out, expected) {
		t.Fatal("v1 route changed native result", code, diag)
	}
	var legacy, diagnostic bytes.Buffer
	if code := operatorcli.RunRevision(nil, bytes.NewReader(v1), &legacy, &diagnostic); code != 0 || diagnostic.Len() != 0 || !bytes.Equal(legacy.Bytes(), expected) {
		t.Fatal("strict legacy v1 changed")
	}
	legacy.Reset()
	diagnostic.Reset()
	if code := operatorcli.RunRevision(nil, bytes.NewReader(v2), &legacy, &diagnostic); code != 2 || legacy.Len() != 0 || diagnostic.String() != "modelrevisioncheck: request rejected\n" {
		t.Fatal("legacy unexpectedly adopted v2", code, diagnostic.String())
	}
	q.FormatVersion = "lifecycle-revision-check-request/v99"
	unknown, _ := json.Marshal(q)
	code, out, diag = run([]string{"revisions"}, bytes.NewReader(unknown))
	if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: request rejected\n" {
		t.Fatal("unknown version", code, diag)
	}
	code, out, diag = run([]string{"revisions", "v2"}, poisonReader{})
	if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: invalid request\n" {
		t.Fatal("new flag admitted", code, diag)
	}
	code, out, diag = run([]string{"revisions"}, failedReader{})
	if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: input failed\n" {
		t.Fatal("read boundary", code, diag)
	}
	diagnostic.Reset()
	if code := operatorcli.Run(context.Background(), []string{"revisions"}, bytes.NewReader(v2), shortWriter{}, &diagnostic); code != 3 || diagnostic.String() != "modelrevisioncheck: output failed\n" {
		t.Fatal("short v2 output", code, diagnostic.String())
	}
}

func TestCompiledFoundryV2AndLegacyRefusalWithClosedStdout(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binaries := map[string]string{}
	for _, name := range []string{"foundry", "modelrevisioncheck"} {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		binary := filepath.Join(t.TempDir(), name)
		cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/"+name)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v %s", name, err, output)
		}
		binaries[name] = binary
	}
	v2 := revisionV2Fixture(t)
	expected, err := revisioncheck.CheckV2(v2)
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag := binaryResult(t, root, binaries["foundry"], []string{"revisions"}, v2, false)
	if code != 0 || diag != "" || !bytes.Equal(out, expected) {
		t.Fatal("compiled v2", code, diag)
	}
	code, out, diag = binaryResult(t, root, binaries["foundry"], []string{"revisions"}, v2, true)
	if code != 3 || len(out) != 0 || diag != "modelrevisioncheck: output failed\n" {
		t.Fatal("compiled closed v2 stdout", code, diag)
	}
	code, out, diag = binaryResult(t, root, binaries["modelrevisioncheck"], nil, v2, false)
	if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: request rejected\n" {
		t.Fatal("compiled legacy v2 refusal", code, diag)
	}
}
