package builderpreflight_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

// These cases come from CF-builder-preflight-v1-pinned-inputs.md, not from
// production report types or a generated implementation snapshot.
const abcSHA = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
const emptySHA = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func item(id, role, path, pin string) string {
	return fmt.Sprintf(`{"id":%q,"role":%q,"path":%q,"sha256":%q}`, id, role, path, pin)
}

func manifest(items ...string) []byte {
	return []byte(`{"formatVersion":"model-builder-inputs/v1","artifacts":[` + strings.Join(items, ",") + `]}`)
}

func writeFixture(t *testing.T, root, path string, raw []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func decodeReport(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatalf("report is not newline terminated: %q", raw)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var got map[string]any
	if err := d.Decode(&got); err != nil {
		t.Fatalf("report JSON: %v", err)
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		t.Fatalf("report has trailing data: %v", err)
	}
	return got
}

func assertEnvelope(t *testing.T, report map[string]any, pin, outcome string, count int) []any {
	t.Helper()
	want := map[string]any{
		"formatVersion": "model-builder-input-report/v1", "manifestSHA256": pin,
		"outcome": outcome, "executionStatus": "not-run", "reviewStatus": "not-established", "authority": "none",
	}
	if len(report) != len(want)+1 {
		t.Fatalf("report fields: %#v", report)
	}
	for key, value := range want {
		if report[key] != value {
			t.Fatalf("report %s = %#v, want %#v", key, report[key], value)
		}
	}
	items, ok := report["artifacts"].([]any)
	if !ok || len(items) != count {
		t.Fatalf("artifacts = %#v, want %d items", report["artifacts"], count)
	}
	return items
}

func assertArtifact(t *testing.T, got any, id, role, path, expected, status string, observed any, size any) {
	t.Helper()
	want := map[string]any{
		"id": id, "role": role, "path": path, "expectedSHA256": expected,
		"status": status, "observedSHA256": observed, "observedBytes": size,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("artifact = %#v\nwant = %#v", got, want)
	}
}

func check(t *testing.T, raw []byte, root string) ([]byte, map[string]any) {
	t.Helper()
	before := bytes.Clone(raw)
	got, err := builderpreflight.Check(raw, root, digest(raw))
	if err != nil {
		t.Fatalf("valid manifest/root rejected: %v", err)
	}
	if !bytes.Equal(raw, before) {
		t.Fatal("Check mutated the manifest")
	}
	return got, decodeReport(t, got)
}

func assertInvalid(t *testing.T, raw []byte, root, pin string) {
	t.Helper()
	got, err := builderpreflight.Check(raw, root, pin)
	if !errors.Is(err, builderpreflight.ErrInvalid) || len(got) != 0 {
		t.Fatalf("rejection = %q, %v; want no report and ErrInvalid", got, err)
	}
}

func TestLiteralReportAndDeterministicOrder(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "z.txt", []byte("abc"))
	writeFixture(t, root, "a.txt", nil)
	raw := manifest(item("z", "requirements", "z.txt", abcSHA), item("a", "claim", "a.txt", emptySHA))
	first, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "matched", 2)
	assertArtifact(t, items[0], "z", "requirements", "z.txt", abcSHA, "matched", abcSHA, json.Number("3"))
	assertArtifact(t, items[1], "a", "claim", "a.txt", emptySHA, "matched", emptySHA, json.Number("0"))
	for i := 0; i < 4; i++ {
		replay, _ := check(t, raw, root)
		if !bytes.Equal(first, replay) {
			t.Fatal("report bytes changed on replay")
		}
	}
	if bytes.Contains(first, []byte(root)) || bytes.Contains(first, []byte(`"contents"`)) {
		t.Fatal("report contains an absolute root or content field")
	}
}

func TestExactBytesAndStalePins(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		name string
		old  []byte
		now  []byte
	}{
		{"line-endings", []byte("a\nb\n"), []byte("a\r\nb\r\n")},
		{"json-whitespace", []byte(`{"a":1}`), []byte("{ \"a\" : 1 }\n")},
		{"json-key-order", []byte(`{"a":1,"b":2}`), []byte(`{"b":2,"a":1}`)},
		{"unicode-normalization", []byte("é"), []byte("e\u0301")},
		{"trailing-newline", []byte("abc"), []byte("abc\n")},
		{"opaque-bytes", []byte("abc"), []byte{0xff, 0x00, 0xfe, 0x80}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writeFixture(t, root, "input", tc.now)
			raw := manifest(item("input", "model", "input", digest(tc.old)))
			_, report := check(t, raw, root)
			items := assertEnvelope(t, report, digest(raw), "blocked", 1)
			assertArtifact(t, items[0], "input", "model", "input", digest(tc.old), "hash-mismatch", digest(tc.now), json.Number(fmt.Sprint(len(tc.now))))
			// A new explicitly pinned manifest can describe any bytes, including
			// invalid JSON and invalid UTF-8; content is not semantically reviewed.
			fresh := manifest(item("input", "model", "input", digest(tc.now)))
			_, report = check(t, fresh, root)
			assertEnvelope(t, report, digest(fresh), "matched", 1)
			assertInvalid(t, fresh, root, digest(raw))
		})
	}
}

func TestAllRolesOpaqueBytesAndNoCandidateExecution(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "candidate-ran")
	candidate := []byte("#!/bin/sh\nprintf unexpected > '" + marker + "'\n")
	roles := []string{"requirements", "model", "interface", "acceptance", "harness", "rubric", "packet", "instructions", "claim"}
	var entries []string
	for i, role := range roles {
		path := fmt.Sprintf("opaque-%d", i)
		writeFixture(t, root, path, candidate)
		if err := os.Chmod(filepath.Join(root, path), 0700); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, item(role, role, path, digest(candidate)))
	}
	before, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	raw := manifest(entries...)
	out, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "matched", len(roles))
	for i, role := range roles {
		path := fmt.Sprintf("opaque-%d", i)
		assertArtifact(t, items[i], role, role, path, digest(candidate), "matched", digest(candidate), json.Number(fmt.Sprint(len(candidate))))
		got, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || !bytes.Equal(got, candidate) {
			t.Fatal("input changed", err)
		}
	}
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("candidate execution marker: %v", err)
	}
	after, err := os.ReadDir(root)
	if err != nil || len(before) != len(after) {
		t.Fatal("preflight wrote a directory entry", err)
	}
	if bytes.Contains(out, []byte("printf")) || bytes.Contains(out, []byte(marker)) {
		t.Fatal("candidate contents leaked into report")
	}
}

func TestReportsEveryListedOutcomeWithoutSubstitution(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "last", []byte("abc"))
	writeFixture(t, root, "stale", []byte("different"))
	writeFixture(t, root, "substitute/absent", []byte("abc"))
	writeFixture(t, root, "plain-parent", []byte("abc"))
	if err := os.Mkdir(filepath.Join(root, "directory"), 0700); err != nil {
		t.Fatal(err)
	}
	entries := []struct{ id, path, status string }{
		{"missing-leaf", "absent", "missing"},
		{"mismatch", "stale", "hash-mismatch"},
		{"missing-parent", "no-parent/input", "missing"},
		{"directory", "directory", "refused"},
		{"file-parent", "plain-parent/input", "refused"},
		{"last", "last", "matched"},
	}
	var spec []string
	for _, tc := range entries {
		spec = append(spec, item(tc.id, "acceptance", tc.path, abcSHA))
	}
	raw := manifest(spec...)
	_, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "blocked", len(entries))
	for i, tc := range entries {
		var observed, size any
		if tc.status == "matched" {
			observed, size = abcSHA, json.Number("3")
		} else if tc.status == "hash-mismatch" {
			observed, size = digest([]byte("different")), json.Number("9")
		}
		assertArtifact(t, items[i], tc.id, "acceptance", tc.path, abcSHA, tc.status, observed, size)
	}
}

func TestArtifactSizeLimitInclusive(t *testing.T) {
	root := t.TempDir()
	limit := bytes.Repeat([]byte{0xa5}, 1<<20)
	writeFixture(t, root, "limit", limit)
	writeFixture(t, root, "over", append(bytes.Clone(limit), 0xa5))
	raw := manifest(item("at", "packet", "limit", digest(limit)), item("over", "packet", "over", digest(append(bytes.Clone(limit), 0xa5))))
	_, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "blocked", 2)
	assertArtifact(t, items[0], "at", "packet", "limit", digest(limit), "matched", digest(limit), json.Number("1048576"))
	assertArtifact(t, items[1], "over", "packet", "over", digest(append(bytes.Clone(limit), 0xa5)), "refused", nil, nil)
}

func TestStrictManifestRejections(t *testing.T) {
	root := t.TempDir()
	validItem := item("a", "model", "a", abcSHA)
	valid := manifest(validItem)
	bad := map[string][]byte{
		"empty": nil, "scalar": []byte(`1`), "array": []byte(`[]`), "null": []byte(`null`),
		"unknown-top":       []byte(strings.Replace(string(valid), `"artifacts"`, `"extra":true,"artifacts"`, 1)),
		"missing-version":   []byte(`{"artifacts":[` + validItem + `]}`),
		"missing-artifacts": []byte(`{"formatVersion":"model-builder-inputs/v1"}`),
		"version":           []byte(strings.Replace(string(valid), "inputs/v1", "inputs/v2", 1)),
		"null-artifacts":    []byte(`{"formatVersion":"model-builder-inputs/v1","artifacts":null}`),
		"null-version":      []byte(`{"formatVersion":null,"artifacts":[` + validItem + `]}`),
		"wrong-artifacts":   []byte(`{"formatVersion":"model-builder-inputs/v1","artifacts":{}}`),
		"null-item":         manifest("null"), "empty-artifacts": manifest(),
		"duplicate-top":          []byte(strings.Replace(string(valid), `"artifacts"`, `"formatVersion":"model-builder-inputs/v1","artifacts"`, 1)),
		"escaped-duplicate-top":  []byte(strings.Replace(string(valid), `"artifacts"`, `"\u0066ormatVersion":"model-builder-inputs/v1","artifacts"`, 1)),
		"case-folded-top":        []byte(strings.Replace(string(valid), "formatVersion", "FormatVersion", 1)),
		"trailing-object":        append(bytes.Clone(valid), []byte(` {}`)...),
		"trailing-garbage":       append(bytes.Clone(valid), '!'),
		"bom":                    append([]byte{0xef, 0xbb, 0xbf}, valid...),
		"invalid-utf8":           manifest(strings.Replace(validItem, `"a"`, "\"\xff\"", 1)),
		"lone-high-surrogate":    manifest(strings.Replace(validItem, `"path":"a"`, `"path":"\ud800"`, 1)),
		"lone-low-surrogate":     manifest(strings.Replace(validItem, `"path":"a"`, `"path":"\udc00"`, 1)),
		"bad-surrogate-pair":     manifest(strings.Replace(validItem, `"path":"a"`, `"path":"\ud800\u0041"`, 1)),
		"unknown-item":           manifest(strings.Replace(validItem, `"id"`, `"extra":1,"id"`, 1)),
		"case-folded-item":       manifest(strings.Replace(validItem, `"id"`, `"ID"`, 1)),
		"duplicate-id":           manifest(validItem, item("a", "claim", "b", abcSHA)),
		"duplicate-path":         manifest(validItem, item("b", "claim", "a", abcSHA)),
		"escaped-duplicate-path": manifest(validItem, `{"id":"b","role":"claim","path":"\u0061","sha256":"`+abcSHA+`"}`),
	}
	for _, field := range []string{"id", "role", "path", "sha256"} {
		var value map[string]any
		if err := json.Unmarshal([]byte(validItem), &value); err != nil {
			t.Fatal(err)
		}
		delete(value, field)
		missing, _ := json.Marshal(value)
		bad["missing-"+field] = manifest(string(missing))
		value[field] = nil
		null, _ := json.Marshal(value)
		bad["null-"+field] = manifest(string(null))
		value[field] = 7
		wrong, _ := json.Marshal(value)
		bad["number-"+field] = manifest(string(wrong))
		bad["duplicate-"+field+"-key"] = manifest(strings.TrimSuffix(validItem, "}") + fmt.Sprintf(`,%q:"a"}`, field))
	}
	for name, raw := range bad {
		t.Run(name, func(t *testing.T) { assertInvalid(t, raw, root, digest(raw)) })
	}
}

func TestIDsRolesPathsAndDigestGrammar(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"", strings.Repeat("a", 65), "a.b", "a b", "é", "a/b", "a\n", "a*"} {
		raw := manifest(item(id, "model", "input", abcSHA))
		assertInvalid(t, raw, root, digest(raw))
	}
	for _, role := range []string{"", "Model", "builder", "reviewed", "model ", "MODEL"} {
		raw := manifest(item("a", role, "input", abcSHA))
		assertInvalid(t, raw, root, digest(raw))
	}
	for _, path := range []string{"", ".", "..", "/absolute", "a/", "a//b", "./a", "a/./b", "a/../b", "../a", `a\b`, "C:input", "https://host/input", "a\x00b", "a\tb", "a\nb", "a\rb", "a\x7fb", "a\u0085b", strings.Repeat("a", 1025)} {
		t.Run(fmt.Sprintf("path-%q", path), func(t *testing.T) {
			raw := manifest(item("a", "model", path, abcSHA))
			assertInvalid(t, raw, root, digest(raw))
		})
	}
	for _, pin := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.ToUpper(abcSHA), strings.Repeat("g", 64), " " + abcSHA, "sha256:" + abcSHA} {
		raw := manifest(item("a", "model", "input", pin))
		assertInvalid(t, raw, root, digest(raw))
		raw = manifest(item("a", "model", "input", abcSHA))
		assertInvalid(t, raw, root, pin)
	}
	// These names are literal, UTF-8 root-relative names, not a normalizer's
	// aliases. A missing valid name must produce a report rather than ErrInvalid.
	for _, path := range []string{"with space", "日本語/😀", "a..b", ".hidden", "percent%2ftext", "a+b", "*.go", "file?", "[ab].go", "a]", strings.Repeat("a/", 511) + "ab"} {
		raw := manifest(item(strings.Repeat("a", 62)+"_-", "instructions", path, abcSHA))
		_, report := check(t, raw, root)
		items := assertEnvelope(t, report, digest(raw), "blocked", 1)
		assertArtifact(t, items[0], strings.Repeat("a", 62)+"_-", "instructions", path, abcSHA, "missing", nil, nil)
	}
}

func TestGlobCharactersAreLiteralNames(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "candidate.go", []byte("abc"))
	for _, path := range []string{"*.go", "file?", "[ab].go", "a]"} {
		raw := manifest(item("literal", "model", path, abcSHA))
		_, report := check(t, raw, root)
		items := assertEnvelope(t, report, digest(raw), "blocked", 1)
		assertArtifact(t, items[0], "literal", "model", path, abcSHA, "missing", nil, nil)
		if err := os.WriteFile(filepath.Join(root, path), []byte("abc"), 0600); err != nil {
			t.Logf("host cannot create literal name %q: %v", path, err)
			continue
		}
		_, report = check(t, raw, root)
		items = assertEnvelope(t, report, digest(raw), "matched", 1)
		assertArtifact(t, items[0], "literal", "model", path, abcSHA, "matched", abcSHA, json.Number("3"))
	}
}

func TestManifestBoundsAndExactPin(t *testing.T) {
	root := t.TempDir()
	var entries []string
	for i := 0; i < 64; i++ {
		entries = append(entries, item(fmt.Sprintf("a%d", i), "interface", fmt.Sprintf("missing%d", i), abcSHA))
	}
	raw := manifest(entries...)
	_, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "blocked", 64)
	for i := range entries {
		assertArtifact(t, items[i], fmt.Sprintf("a%d", i), "interface", fmt.Sprintf("missing%d", i), abcSHA, "missing", nil, nil)
	}
	tooMany := manifest(append(entries, item("a64", "interface", "missing64", abcSHA))...)
	assertInvalid(t, tooMany, root, digest(tooMany))
	atLimit := append(bytes.Clone(raw), bytes.Repeat([]byte(" "), (64<<10)-len(raw))...)
	_, padded := check(t, atLimit, root)
	assertEnvelope(t, padded, digest(atLimit), "blocked", 64)
	assertInvalid(t, atLimit, root, digest(raw))
	overLimit := append(bytes.Clone(atLimit), ' ')
	assertInvalid(t, overLimit, root, digest(overLimit))
}

func TestRootMustBeExplicitAbsoluteDirectory(t *testing.T) {
	root := t.TempDir()
	raw := manifest(item("a", "model", "input", abcSHA))
	writeFixture(t, root, "file", []byte("abc"))
	for _, bad := range []string{"", ".", "relative/path", filepath.Join(root, "missing"), filepath.Join(root, "file"), root + string([]byte{0})} {
		assertInvalid(t, raw, bad, digest(raw))
	}
	link := filepath.Join(t.TempDir(), "linked-root")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("host does not permit symlinks: %v", err)
	}
	assertInvalid(t, raw, link, digest(raw))
	assertInvalid(t, raw, link+string(filepath.Separator), digest(raw))
}

func TestStaticSymlinksAreRefusedEvenInsideRoot(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	writeFixture(t, root, "real/input", []byte("abc"))
	writeFixture(t, outside, "input", []byte("abc"))
	links := []struct{ name, target, listed string }{
		{"leaf-in", "real/input", "leaf-in"},
		{"leaf-out", filepath.Join(outside, "input"), "leaf-out"},
		{"parent-in", "real", "parent-in/input"},
		{"parent-out", outside, "parent-out/input"},
		{"dangling", "nowhere", "dangling"},
		{"loop", "loop", "loop"},
	}
	var entries []string
	for _, tc := range links {
		if err := os.Symlink(tc.target, filepath.Join(root, tc.name)); err != nil {
			t.Skipf("host does not permit symlinks: %v", err)
		}
		entries = append(entries, item(tc.name, "harness", tc.listed, abcSHA))
	}
	entries = append(entries, item("after", "harness", "real/input", abcSHA))
	raw := manifest(entries...)
	_, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "blocked", len(entries))
	for i, tc := range links {
		assertArtifact(t, items[i], tc.name, "harness", tc.listed, abcSHA, "refused", nil, nil)
	}
	assertArtifact(t, items[len(links)], "after", "harness", "real/input", abcSHA, "matched", abcSHA, json.Number("3"))
}

func TestPermissionFailureIsReadFailed(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "private-input", []byte("abc"))
	path := filepath.Join(root, "private-input")
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("host does not support permission fixture: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0600) })
	probe, err := os.Open(path)
	if err == nil {
		probe.Close()
		t.Skip("host can read a mode-000 file")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Skipf("host does not expose the intended permission failure: %v", err)
	}
	raw := manifest(item("a", "rubric", "private-input", abcSHA))
	_, report := check(t, raw, root)
	items := assertEnvelope(t, report, digest(raw), "blocked", 1)
	assertArtifact(t, items[0], "a", "rubric", "private-input", abcSHA, "read-failed", nil, nil)
}

func TestSpecialFilesDoNotBlock(t *testing.T) {
	const child = "FOUNDRY_PREFLIGHT_SPECIAL_CHILD"
	if os.Getenv(child) == "1" {
		raw := manifest(item("special", "packet", "special", emptySHA))
		_, report := check(t, raw, os.Getenv("FOUNDRY_PREFLIGHT_SPECIAL_ROOT"))
		items := assertEnvelope(t, report, digest(raw), "blocked", 1)
		assertArtifact(t, items[0], "special", "packet", "special", emptySHA, "refused", nil, nil)
		return
	}
	for _, kind := range []string{"socket", "fifo"} {
		t.Run(kind, func(t *testing.T) {
			// A short name keeps Unix socket paths within the native length limit.
			root, err := os.MkdirTemp("", "bp-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.RemoveAll(root) })
			path := filepath.Join(root, "special")
			if kind == "socket" {
				listener, err := net.Listen("unix", path)
				if err != nil {
					t.Skipf("host does not support Unix socket fixture: %v", err)
				}
				defer listener.Close()
			} else {
				mkfifo, err := exec.LookPath("mkfifo")
				if err != nil {
					t.Skip("host has no mkfifo fixture tool")
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := exec.CommandContext(ctx, mkfifo, path).Run(); err != nil {
					t.Skipf("host does not permit FIFO fixture: %v", err)
				}
			}
			// Run only the trusted test binary, with a deadline so an erroneous
			// blocking open of a FIFO cannot hang the test suite.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSpecialFilesDoNotBlock$")
			cmd.Env = append(os.Environ(), child+"=1", "FOUNDRY_PREFLIGHT_SPECIAL_ROOT="+root)
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil || err != nil {
				t.Fatalf("special-file check: deadline=%v error=%v output=%s", ctx.Err(), err, out)
			}
		})
	}
}

// Access-time probes are conditional: they test reads/enumeration only on hosts
// where those operations observably advance access time. They do not claim a
// syscall audit or coverage of concurrent filesystem changes.
func accessTime(path string) (time.Time, bool) {
	info, err := os.Stat(path)
	if err != nil || info.Sys() == nil {
		return time.Time{}, false
	}
	v := reflect.ValueOf(info.Sys())
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return time.Time{}, false
	}
	for _, name := range []string{"Atim", "Atimespec"} {
		stamp := v.FieldByName(name)
		if stamp.IsValid() && stamp.Kind() == reflect.Struct {
			sec, nsec := stamp.FieldByName("Sec"), stamp.FieldByName("Nsec")
			if sec.IsValid() && nsec.IsValid() && sec.CanInt() && nsec.CanInt() {
				return time.Unix(sec.Int(), nsec.Int()), true
			}
		}
	}
	return time.Time{}, false
}

func armAccessProbe(t *testing.T, path string, probe func() error) time.Time {
	t.Helper()
	old := time.Unix(946684800, 0)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Skipf("host cannot set probe times: %v", err)
	}
	if err := probe(); err != nil {
		t.Fatal(err)
	}
	if now, ok := accessTime(path); !ok || now.Equal(old) {
		t.Skip("host does not expose access-time changes for this probe")
	}
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	return old
}

func TestEntireManifestAndFrozenPinPrecedeArtifactReads(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "input", []byte("abc"))
	path := filepath.Join(root, "input")
	old := armAccessProbe(t, path, func() error { _, err := os.ReadFile(path); return err })
	valid := manifest(item("a", "model", "input", abcSHA))
	malformedLast := manifest(item("a", "model", "input", abcSHA), item("b", "unapproved-role", "other", abcSHA))
	for _, tc := range []struct {
		raw []byte
		pin string
	}{
		{valid, emptySHA},
		{append(bytes.Clone(valid), ' '), digest(valid)},
		{malformedLast, digest(malformedLast)},
	} {
		assertInvalid(t, tc.raw, root, tc.pin)
		if now, ok := accessTime(path); !ok || !now.Equal(old) {
			t.Fatal("artifact was read before the entire manifest and pin were accepted")
		}
	}
}

func TestNoDirectoryDiscoveryOrFileWrites(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "listed", []byte("abc"))
	writeFixture(t, root, "unlisted/private", []byte("private unlisted content"))
	file := filepath.Join(root, "listed")
	before, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	old := armAccessProbe(t, root, func() error { _, err := os.ReadDir(root); return err })
	raw := manifest(item("a", "model", "listed", abcSHA))
	_, report := check(t, raw, root)
	assertEnvelope(t, report, digest(raw), "matched", 1)
	if now, ok := accessTime(root); !ok || !now.Equal(old) {
		t.Fatal("preflight enumerated the input directory")
	}
	after, err := os.Stat(file)
	if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("listed file identity, bytes, mode, or modification time changed", err)
	}
	got, err := os.ReadFile(file)
	if err != nil || string(got) != "abc" {
		t.Fatal("listed file bytes changed", err)
	}
	unlisted, err := os.ReadFile(filepath.Join(root, "unlisted/private"))
	if err != nil || string(unlisted) != "private unlisted content" {
		t.Fatal("unlisted file bytes changed", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 2 {
		t.Fatal("preflight created or removed an entry", err)
	}
}
