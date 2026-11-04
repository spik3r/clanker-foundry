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

func revisionV3Fixtures(t *testing.T) [][]byte {
	t.Helper()
	var data struct {
		Vectors []struct {
			Request string
			Valid   bool
		}
	}
	if json.Unmarshal(read(t, "../../revisioncheck/testdata/v3-observations.json"), &data) != nil || len(data.Vectors) == 0 {
		t.Fatal("independent v3 fixture")
	}
	var requests [][]byte
	for _, vector := range data.Vectors {
		if vector.Valid {
			requests = append(requests, []byte(vector.Request))
		}
	}
	return requests
}

func TestFoundryRevisionsSelectsExplicitV3WithoutChangingLegacy(t *testing.T) {
	for _, raw := range revisionV3Fixtures(t) {
		expected, err := revisioncheck.CheckV3(raw)
		if err != nil {
			t.Fatal(err)
		}
		code, out, diag := run([]string{"revisions"}, bytes.NewReader(raw))
		if code != 0 || diag != "" || !bytes.Equal(out, expected) {
			t.Fatal("v3 route changed native bytes", code, diag)
		}
		var legacy, diagnostic bytes.Buffer
		if code := operatorcli.RunRevision(nil, bytes.NewReader(raw), &legacy, &diagnostic); code != 2 || legacy.Len() != 0 || diagnostic.String() != "modelrevisioncheck: request rejected\n" {
			t.Fatal("legacy unexpectedly adopted v3", code, diagnostic.String())
		}
	}
	raw := revisionV3Fixtures(t)[0]
	var diagnostic bytes.Buffer
	if code := operatorcli.Run(context.Background(), []string{"revisions"}, bytes.NewReader(raw), shortWriter{}, &diagnostic); code != 3 || diagnostic.String() != "modelrevisioncheck: output failed\n" {
		t.Fatal("short v3 output", code, diagnostic.String())
	}
	code, out, diag := run([]string{"revisions", "v3"}, poisonReader{})
	if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: invalid request\n" {
		t.Fatal("version flag admitted", code, diag)
	}
	for _, invalid := range [][]byte{
		bytes.Replace(raw, []byte("lifecycle-revision-check-request/v3"), []byte("lifecycle-revision-check-request/v99"), 1),
		bytes.Replace(raw, []byte(`"before":`), []byte(`"extra":null,"before":`), 1),
		append(bytes.Clone(raw), raw...),
	} {
		code, out, diag := run([]string{"revisions"}, bytes.NewReader(invalid))
		if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: request rejected\n" {
			t.Fatal("malformed v3 route input accepted", code, diag)
		}
	}
}

func TestCompiledFoundryV3FullReportClosedPipeAndLegacyRefusal(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binaries := map[string]string{}
	for _, name := range []string{"foundry", "modelrevisioncheck"} {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		binary := filepath.Join(t.TempDir(), name+".exe")
		cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/"+name)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v %s", name, err, output)
		}
		binaries[name] = binary
	}
	raw := revisionV3Fixtures(t)[3] // Edited model with stale suite binding.
	expected, err := revisioncheck.CheckV3(raw)
	if err != nil {
		t.Fatal(err)
	}
	code, out, diag := binaryResult(t, root, binaries["foundry"], []string{"revisions"}, raw, false)
	if code != 0 || diag != "" || !bytes.Equal(out, expected) {
		t.Fatal("compiled v3 complete bytes", code, diag)
	}
	code, out, diag = binaryResult(t, root, binaries["foundry"], []string{"revisions"}, raw, true)
	if code != 3 || len(out) != 0 || diag != "modelrevisioncheck: output failed\n" {
		t.Fatal("compiled v3 closed stdout", code, diag)
	}
	code, out, diag = binaryResult(t, root, binaries["modelrevisioncheck"], nil, raw, false)
	if code != 2 || len(out) != 0 || diag != "modelrevisioncheck: request rejected\n" {
		t.Fatal("compiled legacy v3 refusal", code, diag)
	}
}
