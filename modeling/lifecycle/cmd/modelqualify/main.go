// modelqualify exports or verifies an example-only offline project; it executes nothing.
package main

import (
	"encoding/base64"
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const maxRequest = 3 << 20

func readSources() ([]artifactbundle.File, error) {
	var out []artifactbundle.File
	for _, name := range qualification.SourceNames() {
		p := "."
		for _, segment := range split(name) {
			p = filepath.Join(p, segment)
			st, e := os.Lstat(p)
			if e != nil || st.Mode()&os.ModeSymlink != 0 {
				return nil, qualification.ErrInvalid
			}
		}
		st, e := os.Lstat(p)
		if e != nil || !st.Mode().IsRegular() || st.Size() > 1<<20 {
			return nil, qualification.ErrInvalid
		}
		f, e := os.Open(p)
		if e != nil {
			return nil, qualification.ErrInvalid
		}
		b, e := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		closeErr := f.Close()
		if e != nil || closeErr != nil || len(b) > 1<<20 {
			return nil, qualification.ErrInvalid
		}
		out = append(out, artifactbundle.File{Name: name, Data: b})
	}
	return out, nil
}
func split(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}
func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func() int { fmt.Fprintln(diag, "modelqualify: invalid request or project"); return 2 }
	if len(args) != 1 {
		return fail()
	}
	limit := int64(maxRequest)
	if args[0] == "verify" {
		limit = artifactbundle.MaxArchiveBytes * 2
	}
	b, e := io.ReadAll(io.LimitReader(in, limit+1))
	if e != nil || int64(len(b)) > limit {
		return fail()
	}
	switch args[0] {
	case "verify":
		var req struct {
			Archive         string `json:"archive"`
			CandidateSHA256 string `json:"candidateSHA256"`
			SuiteSHA256     string `json:"suiteSHA256"`
		}
		if strictjson.Decode(b, &req, int(limit), 8) != nil {
			return fail()
		}
		archive, e := base64.StdEncoding.DecodeString(req.Archive)
		if e != nil || len(archive) > artifactbundle.MaxArchiveBytes {
			return fail()
		}
		if qualification.Verify(archive, req.CandidateSHA256, req.SuiteSHA256) != nil {
			return fail()
		}
		msg := "verified; suite unvalidated and unexecuted\n"
		n, e := io.WriteString(out, msg)
		if e != nil || n != len(msg) {
			return fail()
		}
		return 0
	case "export":
		var req struct {
			Candidate string `json:"candidate"`
			Suite     string `json:"suite"`
		}
		if strictjson.Decode(b, &req, int(limit), 8) != nil {
			return fail()
		}
		sources, e := readSources()
		if e != nil {
			return fail()
		}
		candidate, e := base64.StdEncoding.DecodeString(req.Candidate)
		if e != nil {
			return fail()
		}
		suite, e := base64.StdEncoding.DecodeString(req.Suite)
		if e != nil {
			return fail()
		}
		result, e := qualification.Build(candidate, suite, sources)
		if e != nil {
			return fail()
		}
		n, e := out.Write(result)
		if e != nil || n != len(result) {
			return fail()
		}
		return 0
	default:
		return fail()
	}
}
func main() {
	// Preserve the existing exit-2 write-failure policy for actual closed pipes.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
