package main

// lifecyclegen implements the Foundry generator contract v1
// (../../../CONTRACT.md) for the lifecycle/v1 profile. It is separate from
// cmd/modelc, which keeps its fixed-path commands.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/views"
)

const (
	contractID        = "foundry-generator/v1"
	contractProfile   = "lifecycle/v1"
	contractGenerator = "lifecyclegen"
	contractVersion   = "1.0.0"
	contractToolchain = "go1.27.3"
	manifestName      = "foundry-manifest.json"
	manifestVersion   = "foundry-generator-manifest/v1"
	contractModelMax  = 1 << 20
	contractFileMax   = 16 << 20
)

// Exit codes from the contract.
const (
	exitIO      = 1
	exitRefused = 2
	exitDrift   = 3
)

type contractError struct {
	code int
	msg  string
}

func (e *contractError) Error() string { return e.msg }

func refuse(format string, a ...any) error {
	return &contractError{exitRefused, fmt.Sprintf(format, a...)}
}

func ioFail(format string, a ...any) error {
	return &contractError{exitIO, fmt.Sprintf(format, a...)}
}

func main() {
	// A closed stdout pipe must reach the contract I/O failure path.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(contractMain(os.Args[1:], os.Stdout, os.Stderr))
}

// contractMain runs contract mode and returns its exit code. It writes stdout
// after contractRun succeeds; a delivery failure may leave a prefix.
func contractMain(args []string, stdout, stderr io.Writer) int {
	out, err := contractRun(args)
	if err != nil {
		code := exitIO
		var ce *contractError
		if errors.As(err, &ce) {
			code = ce.code
		}
		if code == exitDrift {
			fmt.Fprintln(stderr, err)
		} else {
			fmt.Fprintln(stderr, "lifecyclegen:", err)
		}
		return code
	}
	if n, err := stdout.Write(out); err != nil || n != len(out) {
		fmt.Fprintln(stderr, "lifecyclegen: output write failed")
		return exitIO
	}
	return 0
}

func contractRun(args []string) ([]byte, error) {
	if len(args) == 2 && args[0] == "validate" {
		id, _, err := contractLoad(args[1])
		return id, err
	}
	if len(args) == 3 && args[0] == "generate" {
		return contractGenerate(args[1], args[2])
	}
	if len(args) == 3 && args[0] == "check" {
		return contractCheck(args[1], args[2])
	}
	return nil, refuse("usage: lifecyclegen validate MODEL | generate MODEL OUT | check MODEL OUT")
}

type contractFile struct {
	path string
	data []byte
}

// contractLoad reads and validates MODEL and returns the identity line and the
// files to emit, sorted by path with the manifest last.
func contractLoad(modelPath string) (identity []byte, files []contractFile, err error) {
	f, err := os.Open(modelPath)
	if err != nil {
		return nil, nil, refuse("model: %v", err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, nil, refuse("model: %v", err)
	}
	if !st.Mode().IsRegular() {
		return nil, nil, refuse("model: not a regular file")
	}
	if st.Size() > contractModelMax {
		return nil, nil, refuse("model: larger than %d bytes", contractModelMax)
	}
	source, err := io.ReadAll(io.LimitReader(f, contractModelMax+1))
	if err != nil {
		return nil, nil, refuse("model: %v", err)
	}
	if len(source) > contractModelMax {
		return nil, nil, refuse("model: larger than %d bytes", contractModelMax)
	}
	m, err := jsonformat.Decode(source)
	if err != nil {
		return nil, nil, refuse("model: %v", err)
	}
	digest, err := canonical.Digest(m)
	if err != nil {
		return nil, nil, refuse("model: %v", err)
	}
	src := sha256.Sum256(source)
	identity, err = contractIdentity(hex.EncodeToString(src[:]), digest)
	if err != nil {
		return nil, nil, ioFail("%v", err)
	}
	files, err = contractFiles(m, digest)
	if err != nil {
		return nil, nil, err
	}
	return identity, files, nil
}

func contractIdentity(source, canon string) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(struct {
		Contract        string `json:"contract"`
		Profile         string `json:"profile"`
		SourceSHA256    string `json:"sourceSHA256"`
		CanonicalSHA256 string `json:"canonicalSHA256"`
	}{contractID, contractProfile, source, canon})
	return b.Bytes(), err
}

func contractFiles(m model.Model, digest string) ([]contractFile, error) {
	code, tests, err := compiler.Generate(m)
	if err != nil {
		return nil, refuse("model: %v", err)
	}
	id := views.Identity{ModelDigest: digest, FormatVersion: jsonformat.FormatVersion, ProfileVersion: model.ProfileVersion, GeneratorVersion: compiler.GeneratorVersion, ToolchainVersion: compiler.ToolchainVersion}
	view, err := views.Render(m, id)
	if err != nil {
		return nil, refuse("model: %v", err)
	}
	header := []byte("// Code generated by " + contractGenerator + " " + contractVersion + " from model " + digest + ". DO NOT EDIT.\n")
	files := []contractFile{
		{"decision-table.json", view.DecisionTable},
		{"decision.svg", view.SVG},
		{"lifecycle.go", append(append([]byte{}, header...), code...)},
		{"model_paths_test.go", append(append([]byte{}, header...), tests...)},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })
	type entry struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Bytes  int    `json:"bytes"`
	}
	entries := []entry{}
	for _, f := range files {
		if len(f.data) > contractFileMax {
			return nil, ioFail("generated artifact limit")
		}
		h := sha256.Sum256(f.data)
		entries = append(entries, entry{f.path, hex.EncodeToString(h[:]), len(f.data)})
	}
	type gen struct {
		ID        string `json:"id"`
		Version   string `json:"version"`
		Toolchain string `json:"toolchain"`
	}
	type mod struct {
		CanonicalSHA256 string `json:"canonicalSHA256"`
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		ManifestVersion string  `json:"manifestVersion"`
		Profile         string  `json:"profile"`
		Generator       gen     `json:"generator"`
		Model           mod     `json:"model"`
		Files           []entry `json:"files"`
	}{manifestVersion, contractProfile, gen{contractGenerator, contractVersion, contractToolchain}, mod{digest}, entries}); err != nil {
		return nil, ioFail("manifest: %v", err)
	}
	return append(files, contractFile{manifestName, b.Bytes()}), nil
}

func contractGenerate(modelPath, out string) ([]byte, error) {
	identity, files, err := contractLoad(modelPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(out); err == nil {
		return nil, refuse("output path already exists: %s", out)
	} else if !os.IsNotExist(err) {
		return nil, refuse("output path: %v", err)
	}
	ps, err := os.Stat(filepath.Dir(out))
	if err != nil || !ps.IsDir() {
		return nil, refuse("parent of output path is not an existing directory")
	}
	if err := os.Mkdir(out, 0o755); err != nil {
		if os.IsExist(err) {
			return nil, refuse("output path already exists: %s", out)
		}
		return nil, ioFail("create output: %v", err)
	}
	for _, f := range files {
		if err := writeNew(filepath.Join(out, f.path), f.data); err != nil {
			return nil, ioFail("output may be partial: %v", err)
		}
	}
	return identity, nil
}

func writeNew(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// trimOutputRoot removes only terminal separators and dot components. Interior
// components retain filesystem meaning, including .. after a parent symlink.
func trimOutputRoot(out string) string {
	root := len(filepath.VolumeName(out))
	for root < len(out) && os.IsPathSeparator(out[root]) {
		root++
	}
	for {
		for len(out) > root && os.IsPathSeparator(out[len(out)-1]) {
			out = out[:len(out)-1]
		}
		n := len(out)
		if n > root && n >= 2 && out[n-1] == '.' && os.IsPathSeparator(out[n-2]) {
			out = out[:n-1]
			continue
		}
		return out
	}
}

func contractCheck(modelPath, out string) ([]byte, error) {
	out = trimOutputRoot(out)
	identity, files, err := contractLoad(modelPath)
	if err != nil {
		return nil, err
	}
	var drift []string
	st, err := os.Lstat(out)
	switch {
	case os.IsNotExist(err):
		drift = append(drift, "drift: missing .")
	case err != nil:
		return nil, ioFail("output: %v", err)
	case !st.IsDir():
		drift = append(drift, "drift: unexpected .")
	default:
		// Resolve the accepted directory before Join can collapse interior parent components.
		out, err = filepath.EvalSymlinks(out)
		if err != nil {
			return nil, ioFail("output: %v", err)
		}
		entries, err := os.ReadDir(out)
		if err != nil {
			return nil, ioFail("output: %v", err)
		}
		expected := map[string][]byte{}
		for _, f := range files {
			expected[f.path] = f.data
		}
		present := map[string]bool{}
		for _, e := range entries {
			present[e.Name()] = true
			want, ok := expected[e.Name()]
			if !ok || !e.Type().IsRegular() {
				drift = append(drift, "drift: unexpected "+e.Name())
				continue
			}
			got, err := readBounded(filepath.Join(out, e.Name()), int64(len(want)))
			if err != nil {
				return nil, ioFail("output: %v", err)
			}
			if !bytes.Equal(got, want) {
				drift = append(drift, "drift: modified "+e.Name())
			}
		}
		for _, f := range files {
			if !present[f.path] {
				drift = append(drift, "drift: missing "+f.path)
			}
		}
	}
	if len(drift) > 0 {
		sort.Strings(drift)
		return nil, &contractError{exitDrift, strings.Join(drift, "\n")}
	}
	return identity, nil
}

// readBounded reads at most want+1 bytes, enough to detect a difference.
func readBounded(path string, want int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, want+1))
}
