package tui

import (
	"runtime"
	"testing"
)

func TestTrimOutputRoot(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""}, {".", "."}, {"./", "."}, {"././", "."},
		{"/", "/"}, {"//", "//"}, {"///", "///"}, {"/./.", "/"},
		{"out", "out"}, {"out/", "out"}, {"out/.", "out"}, {"out///././", "out"},
		{"../", ".."}, {"a/../.", "a/.."}, {"a/./b", "a/./b"},
		{"parent/link/../out/.", "parent/link/../out"},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases, []struct{ in, want string }{
			{`C:\`, `C:\`}, {`C:`, `C:`}, {`C:.`, `C:.`}, {`C:.\`, `C:.`},
			{`C:\out\.\`, `C:\out`}, {`out\.\`, `out`},
			{`\\server\share`, `\\server\share`}, {`\\server\share\.\`, `\\server\share\`},
			{`C:\parent\link\..\out\.`, `C:\parent\link\..\out`},
		}...)
	}
	for _, c := range cases {
		if got := trimOutputRoot(c.in); got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}
