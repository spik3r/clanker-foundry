// Package gencontract is a black-box conformance harness for the Foundry
// generator contract v1 (see ../CONTRACT.md). It drives a built generator
// binary from its command line and reports every contract violation it can
// observe.
package gencontract

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// ContractID is the contract identifier a generator must report.
	ContractID = "foundry-generator/v1"
	// ManifestName is the manifest file at the root of every output.
	ManifestName    = "foundry-manifest.json"
	manifestVersion = "foundry-generator-manifest/v1"
	runTimeout      = 30 * time.Second
)

var (
	hexRe     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	idRe      = regexp.MustCompile(`^[a-z0-9-]+$`)
	versionRe = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

// Generator describes one generator under test.
type Generator struct {
	ID      string   // expected manifest generator.id
	Profile string   // expected profile
	Bin     string   // absolute path to the built binary
	Model   string   // absolute path to a valid sample model
	Env     []string // extra environment for the child process (fault tests only)
}

// Finding is one observed contract violation.
type Finding struct {
	Rule   string // stable rule ID, documented in README.md
	Detail string
}

func (f Finding) String() string { return f.Rule + ": " + f.Detail }

type identity struct {
	Contract        string `json:"contract"`
	Profile         string `json:"profile"`
	SourceSHA256    string `json:"sourceSHA256"`
	CanonicalSHA256 string `json:"canonicalSHA256"`
}

type manifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type manifestDoc struct {
	ManifestVersion string `json:"manifestVersion"`
	Profile         string `json:"profile"`
	Generator       struct {
		ID        string `json:"id"`
		Version   string `json:"version"`
		Toolchain string `json:"toolchain"`
	} `json:"generator"`
	Model struct {
		CanonicalSHA256 string `json:"canonicalSHA256"`
	} `json:"model"`
	Files []manifestFile `json:"files"`
}

type result struct {
	code           int
	stdout, stderr []byte
}

// entry is one node of a directory tree read from disk.
type entry struct {
	kind byte // 'f' regular file, 'd' directory, 'l' symlink, 'o' other
	data []byte
}

type checker struct {
	g        Generator
	scratch  string
	work     string // empty working directory of every child process
	seq      int
	model    []byte
	findings []Finding
	pristine map[string]entry // generated tree, nil if generation failed
}

// Check runs every conformance rule against g. It returns the contract
// violations it observed. The error is non-nil only for harness problems such
// as an unusable scratch directory or an unstartable binary. scratch is
// created if needed; the caller removes it.
func Check(g Generator, scratch string) ([]Finding, error) {
	model, err := os.ReadFile(g.Model)
	if err != nil {
		return nil, fmt.Errorf("read sample model: %w", err)
	}
	if err := os.MkdirAll(scratch, 0o755); err != nil {
		return nil, err
	}
	c := &checker{g: g, scratch: scratch, model: model}
	if c.work, err = c.dir("cwd"); err != nil {
		return nil, err
	}

	id, idOK, err := c.stepValidate()
	if err != nil {
		return nil, err
	}
	alt, err := c.stepReformatted(id, idOK)
	if err != nil {
		return nil, err
	}
	out, tree, err := c.stepGenerate(alt, id, idOK)
	if err != nil {
		return nil, err
	}
	if tree != nil {
		steps := []func() error{
			func() error { return c.stepManifestAndHeader(tree, id, idOK) },
			func() error { return c.stepCheckClean(out, tree, id, idOK) },
			func() error { return c.stepDrift(tree) },
		}
		for _, s := range steps {
			if err := s(); err != nil {
				return nil, err
			}
		}
	}
	c.pristine = tree
	for _, s := range []func() error{c.stepExistingOut, c.stepInvalid, c.stepMissingModel} {
		if err := s(); err != nil {
			return nil, err
		}
	}
	if err := c.stepUsage(); err != nil {
		return nil, err
	}
	return c.findings, nil
}

func (c *checker) add(rule, format string, a ...any) {
	c.findings = append(c.findings, Finding{rule, fmt.Sprintf(format, a...)})
}

// dir creates a fresh directory under the scratch directory.
func (c *checker) dir(name string) (string, error) {
	c.seq++
	d := filepath.Join(c.scratch, fmt.Sprintf("%02d-%s", c.seq, name))
	return d, os.Mkdir(d, 0o755)
}

// run starts the generator with an emptied environment and a bounded
// timeout, in the empty working directory.
func (c *checker) run(args ...string) (result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.g.Bin, args...)
	cmd.Dir = c.work
	cmd.Env = append([]string{}, c.g.Env...)
	cmd.WaitDelay = 5 * time.Second
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	r := result{stdout: so.Bytes(), stderr: se.Bytes()}
	var ee *exec.ExitError
	switch {
	case err == nil:
	case ctx.Err() != nil:
		r.code = -1
		c.add("run.timeout", "%s did not finish within %s", strings.Join(args, " "), runTimeout)
	case errors.As(err, &ee):
		r.code = ee.ExitCode()
	default:
		return r, fmt.Errorf("start %s: %w", c.g.Bin, err)
	}
	return r, nil
}

// runRO runs a command that must not write anything: it reports a "readonly"
// finding if the model's directory or the working directory changed.
func (c *checker) runRO(modelPath string, args ...string) (result, error) {
	watch := func() string {
		return fingerprint(filepath.Dir(modelPath)) + "|" + fingerprint(c.work)
	}
	before := watch()
	r, err := c.run(args...)
	if after := watch(); after != before {
		c.add("readonly", "%s changed the model directory or the working directory", strings.Join(args[:1], " "))
	}
	return r, err
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 160 {
		s = s[:160] + "..."
	}
	return fmt.Sprintf("%q", s)
}

func sha(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func encodeJSON(v any, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if indent {
		enc.SetIndent("", "  ")
	}
	err := enc.Encode(v)
	return buf.Bytes(), err
}

// expectRefused checks a nonzero exit: the code, empty stdout, and at least
// one stderr diagnostic. It reports under rule and returns whether the exit
// code was the expected one.
func (c *checker) expectRefused(rule, what string, r result, code int) bool {
	if r.code != code {
		c.add(rule, "%s: exit %d, want %d (stdout %s, stderr %s)", what, r.code, code, snippet(r.stdout), snippet(r.stderr))
		return false
	}
	if len(r.stdout) != 0 {
		c.add(rule, "%s: stdout is not empty on failure: %s", what, snippet(r.stdout))
	}
	if len(bytes.TrimSpace(r.stderr)) == 0 {
		c.add(rule, "%s: stderr has no diagnostic", what)
	}
	return true
}

// expectDrift checks an exit 3 with the wanted drift line. With exact, that
// must be the only drift line.
func (c *checker) expectDrift(rule, what string, r result, want string, exact bool) {
	if !c.expectRefused(rule, what, r, 3) {
		return
	}
	var drift []string
	for _, l := range strings.Split(string(r.stderr), "\n") {
		if strings.HasPrefix(l, "drift: ") {
			drift = append(drift, l)
		}
	}
	found := false
	for _, l := range drift {
		found = found || l == want
	}
	switch {
	case !found:
		c.add(rule, "%s: stderr lacks the line %q (drift lines %q)", what, want, drift)
	case exact && len(drift) != 1:
		c.add(rule, "%s: want only %q, got drift lines %q", what, want, drift)
	}
}

// checkIdentity verifies an identity line on stdout against the model bytes.
func (c *checker) checkIdentity(rule string, stdout, src []byte) (identity, bool) {
	var id identity
	dec := json.NewDecoder(bytes.NewReader(stdout))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&id); err != nil {
		c.add(rule, "stdout is not an identity object: %v (stdout %s)", err, snippet(stdout))
		return id, false
	}
	if want, _ := encodeJSON(id, false); !bytes.Equal(stdout, want) {
		c.add(rule, "stdout is not exactly one compact line with the four members in order: got %q, want %q", stdout, want)
	}
	if id.Contract != ContractID {
		c.add(rule, "contract is %q, want %q", id.Contract, ContractID)
	}
	if id.Profile != c.g.Profile {
		c.add(rule, "profile is %q, want %q", id.Profile, c.g.Profile)
	}
	if id.SourceSHA256 != sha(src) {
		c.add(rule, "sourceSHA256 is %q, want the SHA-256 of the model bytes %q", id.SourceSHA256, sha(src))
	}
	if !hexRe.MatchString(id.CanonicalSHA256) {
		c.add(rule, "canonicalSHA256 %q is not 64 lowercase hex digits", id.CanonicalSHA256)
	}
	return id, true
}

func (c *checker) stepValidate() (identity, bool, error) {
	r, err := c.runRO(c.g.Model, "validate", c.g.Model)
	if err != nil {
		return identity{}, false, err
	}
	if r.code != 0 {
		c.add("validate.exit", "exit %d, want 0 (stderr %s)", r.code, snippet(r.stderr))
		return identity{}, false, nil
	}
	id, ok := c.checkIdentity("validate.identity", r.stdout, c.model)
	return id, ok, nil
}

// reformat returns a copy of raw that differs in bytes but only in JSON
// whitespace: keys are not reordered and number text is preserved.
func reformat(raw []byte) ([]byte, error) {
	var cands [][]byte
	var buf bytes.Buffer
	if json.Compact(&buf, raw) == nil {
		cands = append(cands, append([]byte(nil), buf.Bytes()...))
	}
	for _, indent := range []string{"  ", "\t", "      "} {
		buf.Reset()
		if json.Indent(&buf, raw, "", indent) == nil {
			cands = append(cands, append([]byte(nil), buf.Bytes()...))
		}
	}
	cands = append(cands, append(append([]byte(nil), raw...), '\n'), append([]byte("\n"), raw...))
	for _, c := range cands {
		if !bytes.Equal(c, raw) {
			return c, nil
		}
	}
	return nil, errors.New("cannot build a whitespace-different copy of the model")
}

// stepReformatted validates a whitespace-reformatted copy at another path and
// returns that path.
func (c *checker) stepReformatted(base identity, baseOK bool) (string, error) {
	alt, err := reformat(c.model)
	if err != nil {
		return "", err
	}
	d, err := c.dir("alt")
	if err != nil {
		return "", err
	}
	name := "reformatted-" + filepath.Base(c.g.Model)
	p := filepath.Join(d, name)
	if err := os.WriteFile(p, alt, 0o644); err != nil {
		return "", err
	}
	r, err := c.runRO(p, "validate", p)
	if err != nil {
		return "", err
	}
	if r.code != 0 {
		c.add("validate.whitespace", "a whitespace-reformatted model was refused: exit %d (stderr %s)", r.code, snippet(r.stderr))
		return p, nil
	}
	id, ok := c.checkIdentity("validate.identity", r.stdout, alt)
	if ok && baseOK && id.CanonicalSHA256 != base.CanonicalSHA256 {
		c.add("validate.whitespace", "canonicalSHA256 changed with JSON whitespace: %s vs %s", base.CanonicalSHA256, id.CanonicalSHA256)
	}
	if ok && baseOK && id.SourceSHA256 == base.SourceSHA256 {
		c.add("validate.whitespace", "sourceSHA256 did not change although the model bytes did")
	}
	return p, nil
}

// readTree reads every entry below root without following symlinks.
func readTree(root string) (map[string]entry, error) {
	t := map[string]entry{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		switch {
		case d.IsDir():
			t[rel] = entry{kind: 'd'}
		case d.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			t[rel] = entry{kind: 'l', data: []byte(target)}
		case d.Type().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			t[rel] = entry{kind: 'f', data: b}
		default:
			t[rel] = entry{kind: 'o'}
		}
		return nil
	})
	return t, err
}

func sortedPaths(t map[string]entry) []string {
	paths := make([]string, 0, len(t))
	for p := range t {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// treeDiff lists the paths where two trees differ.
func treeDiff(a, b map[string]entry) []string {
	var out []string
	for _, p := range sortedPaths(a) {
		o, ok := b[p]
		switch {
		case !ok:
			out = append(out, p+" (only in first)")
		case o.kind != a[p].kind || !bytes.Equal(o.data, a[p].data):
			out = append(out, p+" (differs)")
		}
	}
	for _, p := range sortedPaths(b) {
		if _, ok := a[p]; !ok {
			out = append(out, p+" (only in second)")
		}
	}
	return out
}

// writeTree recreates the files and directories of t under root.
func writeTree(root string, t map[string]entry) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for _, p := range sortedPaths(t) {
		full := filepath.Join(root, filepath.FromSlash(p))
		switch t[p].kind {
		case 'd':
			if err := os.MkdirAll(full, 0o755); err != nil {
				return err
			}
		case 'f':
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(full, t[p].data, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// fingerprint describes a path and everything below it, without following
// symlinks, so that two fingerprints are equal only if nothing changed.
func fingerprint(root string) string {
	var b strings.Builder
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			t, err := os.Readlink(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(&b, "%s l %s\n", rel, t)
		case d.IsDir():
			fmt.Fprintf(&b, "%s d\n", rel)
		case d.Type().IsRegular():
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(&b, "%s f %s\n", rel, sha(data))
		default:
			fmt.Fprintf(&b, "%s o %v\n", rel, d.Type())
		}
		return nil
	})
	if err != nil {
		return "error: " + err.Error()
	}
	return b.String()
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

// stepGenerate generates twice, the second time from the reformatted model at
// another path, and compares the trees. It returns the first output and tree,
// or a nil tree when generation failed.
func (c *checker) stepGenerate(alt string, base identity, baseOK bool) (string, map[string]entry, error) {
	var outs [2]string
	var trees [2]map[string]entry
	var ids [2]identity
	for i, model := range []string{c.g.Model, alt} {
		d, err := c.dir(fmt.Sprintf("gen%d", i+1))
		if err != nil {
			return "", nil, err
		}
		outs[i] = filepath.Join(d, []string{"out-one", "out-two"}[i])
		r, err := c.run("generate", model, outs[i])
		if err != nil {
			return "", nil, err
		}
		if r.code != 0 {
			c.add("generate.exit", "generate %d: exit %d, want 0 (stderr %s)", i+1, r.code, snippet(r.stderr))
			return "", nil, nil
		}
		src := c.model
		if i == 1 {
			if src, err = os.ReadFile(alt); err != nil {
				return "", nil, err
			}
		}
		ids[i], _ = c.checkIdentity("generate.identity", r.stdout, src)
		if trees[i], err = readTree(outs[i]); err != nil {
			c.add("generate.exit", "cannot read output %d: %v", i+1, err)
			return "", nil, nil
		}
	}
	if baseOK && ids[0].CanonicalSHA256 != base.CanonicalSHA256 {
		c.add("generate.identity", "canonicalSHA256 differs from validate: %s vs %s", ids[0].CanonicalSHA256, base.CanonicalSHA256)
	}
	if ids[0].CanonicalSHA256 != ids[1].CanonicalSHA256 {
		c.add("generate.deterministic", "canonicalSHA256 differs between the original and the reformatted model")
	}
	if diff := treeDiff(trees[0], trees[1]); len(diff) > 0 {
		c.add("generate.deterministic", "output differs between two runs (other output path, other model path, other whitespace): %s", strings.Join(diff, ", "))
	}
	for _, p := range sortedPaths(trees[0]) {
		switch trees[0][p].kind {
		case 'l', 'o':
			c.add("generate.tree", "%s is not a regular file or directory", p)
		case 'd':
			empty := true
			for q := range trees[0] {
				if strings.HasPrefix(q, p+"/") {
					empty = false
				}
			}
			if empty {
				c.add("generate.tree", "%s is an empty directory, which check cannot account for", p)
			}
		}
	}
	return outs[0], trees[0], nil
}

func validPath(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return false
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}

func (c *checker) stepManifestAndHeader(tree map[string]entry, id identity, idOK bool) error {
	m := c.stepManifest(tree, id, idOK)
	if m == nil {
		return nil
	}
	header := fmt.Sprintf("// Code generated by %s %s from model %s. DO NOT EDIT.\n", c.g.ID, m.Generator.Version, id.CanonicalSHA256)
	for _, p := range sortedPaths(tree) {
		if tree[p].kind != 'f' || !(strings.HasSuffix(p, ".go") || strings.HasSuffix(p, ".ts")) {
			continue
		}
		if !bytes.HasPrefix(tree[p].data, []byte(header)) {
			first, _, _ := bytes.Cut(tree[p].data, []byte("\n"))
			c.add("header", "%s: first line is %q, want %q", p, first, strings.TrimSuffix(header, "\n"))
		}
	}
	return nil
}

// stepManifest checks the manifest and returns it, or nil if it is unusable.
func (c *checker) stepManifest(tree map[string]entry, id identity, idOK bool) *manifestDoc {
	me, ok := tree[ManifestName]
	if !ok || me.kind != 'f' {
		c.add("manifest.encoding", "%s is missing or not a regular file", ManifestName)
		return nil
	}
	var m manifestDoc
	dec := json.NewDecoder(bytes.NewReader(me.data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		c.add("manifest.encoding", "cannot decode strictly: %v", err)
		return nil
	}
	if _, err := dec.Token(); err != io.EOF {
		c.add("manifest.encoding", "data after the manifest object")
		return nil
	}
	if want, err := encodeJSON(m, true); err != nil || !bytes.Equal(want, me.data) {
		n := 0
		for n < len(want) && n < len(me.data) && want[n] == me.data[n] {
			n++
		}
		c.add("manifest.encoding", "bytes are not the exact encoding (two-space indent, member order, trailing newline); first difference at offset %d", n)
	}

	if m.ManifestVersion != manifestVersion {
		c.add("manifest.fields", "manifestVersion is %q, want %q", m.ManifestVersion, manifestVersion)
	}
	if m.Profile != c.g.Profile {
		c.add("manifest.fields", "profile is %q, want %q", m.Profile, c.g.Profile)
	}
	if m.Generator.ID != c.g.ID || !idRe.MatchString(m.Generator.ID) {
		c.add("manifest.fields", "generator.id is %q, want %q (lowercase letters, digits, hyphens)", m.Generator.ID, c.g.ID)
	}
	if !versionRe.MatchString(m.Generator.Version) {
		c.add("manifest.fields", "generator.version %q is not three dot-separated decimal numbers", m.Generator.Version)
	}
	if m.Generator.Toolchain == "" {
		c.add("manifest.fields", "generator.toolchain is empty")
	}
	if !hexRe.MatchString(m.Model.CanonicalSHA256) || (idOK && m.Model.CanonicalSHA256 != id.CanonicalSHA256) {
		c.add("manifest.fields", "model.canonicalSHA256 is %q, want the identity line's %q", m.Model.CanonicalSHA256, id.CanonicalSHA256)
	}

	if len(m.Files) == 0 {
		c.add("manifest.files", "files is empty; at least one file is required")
	}
	listed := map[string]bool{}
	for i, f := range m.Files {
		if i > 0 && f.Path <= m.Files[i-1].Path {
			c.add("manifest.files", "%q is not after %q in byte order", f.Path, m.Files[i-1].Path)
		}
		if !validPath(f.Path) {
			c.add("manifest.files", "invalid path %q", f.Path)
			continue
		}
		if f.Path == ManifestName {
			c.add("manifest.files", "the manifest lists itself")
			continue
		}
		listed[f.Path] = true
		disk, ok := tree[f.Path]
		switch {
		case !ok || disk.kind != 'f':
			c.add("manifest.files", "%s is listed but is not a regular file on disk", f.Path)
		case f.SHA256 != sha(disk.data) || f.Bytes != int64(len(disk.data)):
			c.add("manifest.files", "%s: sha256/bytes do not match the file on disk", f.Path)
		}
	}
	for _, p := range sortedPaths(tree) {
		if tree[p].kind == 'f' && p != ManifestName && !listed[p] {
			c.add("manifest.files", "%s is on disk but not listed", p)
		}
	}
	return &m
}

func (c *checker) stepCheckClean(out string, tree map[string]entry, id identity, idOK bool) error {
	r, err := c.runRO(c.g.Model, "check", c.g.Model, out)
	if err != nil {
		return err
	}
	if r.code != 0 {
		c.add("check.clean", "fresh output: exit %d, want 0 (stderr %s)", r.code, snippet(r.stderr))
	} else {
		got, ok := c.checkIdentity("check.identity", r.stdout, c.model)
		if ok && idOK && got != id {
			c.add("check.identity", "identity line differs from validate")
		}
	}
	after, err := readTree(out)
	if err != nil {
		return err
	}
	if diff := treeDiff(tree, after); len(diff) > 0 {
		c.add("check.writes", "check changed the output tree: %s", strings.Join(diff, ", "))
	}
	return nil
}

// freshCopy writes a new copy of the pristine tree and returns its root.
func (c *checker) freshCopy(tree map[string]entry) (string, error) {
	d, err := c.dir("copy")
	if err != nil {
		return "", err
	}
	out := filepath.Join(d, "out")
	return out, writeTree(out, tree)
}

// driftCase runs check on a tampered tree and verifies the drift report and
// that check left the tree as it found it.
func (c *checker) driftCase(rule, what, out, want string, exact bool, watch ...string) error {
	snap := func() string {
		s := fingerprint(out)
		for _, w := range watch {
			s += "|" + fingerprint(w)
		}
		return s
	}
	before := snap()
	r, err := c.runRO(c.g.Model, "check", c.g.Model, out)
	if err != nil {
		return err
	}
	c.expectDrift(rule, what, r, want, exact)
	if snap() != before {
		c.add("check.repaired", "%s: check changed the tampered tree", what)
	}
	return nil
}

func (c *checker) stepDrift(tree map[string]entry) error {
	var files, dirs []string
	for _, p := range sortedPaths(tree) {
		switch tree[p].kind {
		case 'f':
			files = append(files, p)
		case 'd':
			dirs = append(dirs, p)
		}
	}
	for _, p := range files {
		// (a) one flipped byte
		out, err := c.freshCopy(tree)
		if err != nil {
			return err
		}
		mutated := append([]byte(nil), tree[p].data...)
		if len(mutated) == 0 {
			mutated = []byte{'x'}
		} else {
			mutated[len(mutated)/2] ^= 0x01
		}
		if err := os.WriteFile(filepath.Join(out, filepath.FromSlash(p)), mutated, 0o644); err != nil {
			return err
		}
		if err := c.driftCase("check.modified", "modify "+p, out, "drift: modified "+p, true); err != nil {
			return err
		}

		// (b) deleted file
		if out, err = c.freshCopy(tree); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(out, filepath.FromSlash(p))); err != nil {
			return err
		}
		if err := c.driftCase("check.missing", "delete "+p, out, "drift: missing "+p, true); err != nil {
			return err
		}

		// (d) symlink to an identical copy
		if out, err = c.freshCopy(tree); err != nil {
			return err
		}
		ld, err := c.dir("link-target")
		if err != nil {
			return err
		}
		target := filepath.Join(ld, "copy")
		if err := os.WriteFile(target, tree[p].data, 0o644); err != nil {
			return err
		}
		full := filepath.Join(out, filepath.FromSlash(p))
		if err := os.Remove(full); err != nil {
			return err
		}
		if err := os.Symlink(target, full); err != nil {
			continue // platform without symlinks
		}
		if err := c.driftCase("check.symlink", "replace "+p+" with a symlink", out, "drift: unexpected "+p, false, ld); err != nil {
			return err
		}
	}

	// (c) extra entries: top level, every subdirectory, and an extra directory
	extras := []struct {
		rel string
		dir bool
	}{{"zz-extra.txt", false}, {"zz-extra-dir", true}}
	for _, d := range dirs {
		extras = append(extras, struct {
			rel string
			dir bool
		}{d + "/zz-extra.txt", false})
	}
	for _, e := range extras {
		out, err := c.freshCopy(tree)
		if err != nil {
			return err
		}
		full := filepath.Join(out, filepath.FromSlash(e.rel))
		if e.dir {
			err = os.Mkdir(full, 0o755)
		} else {
			err = os.WriteFile(full, []byte("extra\n"), 0o644)
		}
		if err != nil {
			return err
		}
		if err := c.driftCase("check.unexpected", "add "+e.rel, out, "drift: unexpected "+e.rel, true); err != nil {
			return err
		}
	}

	// (e) OUT absent
	d, err := c.dir("absent")
	if err != nil {
		return err
	}
	return c.driftCase("check.absent", "absent OUT", filepath.Join(d, "out"), "drift: missing .", true)
}

func (c *checker) stepExistingOut() error {
	d, err := c.dir("existing")
	if err != nil {
		return err
	}
	emptyDir := filepath.Join(d, "empty-dir")
	file := filepath.Join(d, "a-file")
	linkTarget := filepath.Join(d, "link-target")
	link := filepath.Join(d, "a-link")
	if err := os.Mkdir(emptyDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte("keep\n"), 0o644); err != nil {
		return err
	}
	if err := os.Mkdir(linkTarget, 0o755); err != nil {
		return err
	}
	cases := []struct {
		what string
		out  string
	}{{"empty directory", emptyDir}, {"regular file", file}}
	if os.Symlink(linkTarget, link) == nil {
		cases = append(cases, struct {
			what string
			out  string
		}{"symlink to an empty directory", link})
	}
	for _, tc := range cases {
		before := fingerprint(tc.out) + "|" + fingerprint(linkTarget)
		r, err := c.run("generate", c.g.Model, tc.out)
		if err != nil {
			return err
		}
		c.expectRefused("generate.existing-out", "generate into an existing "+tc.what, r, 2)
		if after := fingerprint(tc.out) + "|" + fingerprint(linkTarget); after != before {
			c.add("generate.existing-out", "generate into an existing %s changed it", tc.what)
		}
	}
	return nil
}

func (c *checker) stepInvalid() error {
	d, err := c.dir("invalid")
	if err != nil {
		return err
	}
	cases := []struct {
		name string
		data []byte
	}{
		{"empty.json", nil},
		{"truncated.json", c.model[:len(c.model)/2]},
		{"empty-object.json", []byte("{}")},
		{"array.json", []byte("[]")},
	}
	for _, tc := range cases {
		p := filepath.Join(d, tc.name)
		if err := os.WriteFile(p, tc.data, 0o644); err != nil {
			return err
		}
		r, err := c.runRO(p, "validate", p)
		if err != nil {
			return err
		}
		c.expectRefused("refuse.invalid-model", "validate "+tc.name, r, 2)

		od, err := c.dir("invalid-out")
		if err != nil {
			return err
		}
		out := filepath.Join(od, "out")
		if r, err = c.run("generate", p, out); err != nil {
			return err
		}
		c.expectRefused("refuse.invalid-model", "generate "+tc.name, r, 2)
		if exists(out) {
			c.add("refuse.invalid-model", "generate %s created OUT", tc.name)
		}

		if c.pristine == nil {
			continue
		}
		copyDir, err := c.freshCopy(c.pristine)
		if err != nil {
			return err
		}
		before := fingerprint(copyDir)
		if r, err = c.runRO(p, "check", p, copyDir); err != nil {
			return err
		}
		c.expectRefused("refuse.invalid-model", "check "+tc.name, r, 2)
		if fingerprint(copyDir) != before {
			c.add("refuse.invalid-model", "check %s changed OUT", tc.name)
		}
	}
	return nil
}

func (c *checker) stepMissingModel() error {
	d, err := c.dir("missing-model")
	if err != nil {
		return err
	}
	p := filepath.Join(d, "no-such-model.json")
	r, err := c.runRO(p, "validate", p)
	if err != nil {
		return err
	}
	c.expectRefused("refuse.missing-model", "validate a missing model", r, 2)
	out := filepath.Join(d, "out")
	if r, err = c.run("generate", p, out); err != nil {
		return err
	}
	c.expectRefused("refuse.missing-model", "generate a missing model", r, 2)
	if exists(out) {
		c.add("refuse.missing-model", "generate a missing model created OUT")
	}
	if c.pristine != nil {
		copyDir, err := c.freshCopy(c.pristine)
		if err != nil {
			return err
		}
		if r, err = c.runRO(p, "check", p, copyDir); err != nil {
			return err
		}
		c.expectRefused("refuse.missing-model", "check a missing model", r, 2)
	}
	return nil
}

func (c *checker) stepUsage() error {
	d, err := c.dir("usage")
	if err != nil {
		return err
	}
	m, out := c.g.Model, filepath.Join(d, "out")
	cases := [][]string{
		{},
		{"frobnicate"},
		{"validate"},
		{"validate", m, m},
		{"generate", m},
		{"generate", m, out, "extra"},
		{"check", m},
		{"--help"},
	}
	for _, args := range cases {
		r, err := c.run(args...)
		if err != nil {
			return err
		}
		c.expectRefused("refuse.usage", fmt.Sprintf("arguments %q", args), r, 2)
	}
	if exists(out) {
		c.add("refuse.usage", "a usage error created OUT")
	}
	return nil
}
