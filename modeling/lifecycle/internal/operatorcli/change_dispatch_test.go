package operatorcli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// These capability expectations are frozen before production routing exists.
// Harmless routes receive no candidate-execution callback, including malformed
// execute forms which must refuse before reading stdin.
func TestChangeDispatchNeverGrantsExecutionToOtherRoutes(t *testing.T) {
	for _, args := range [][]string{
		nil, {"help"}, {"--help"}, {"unknown"}, {"help", "extra"},
		{"prepare"}, {"review"}, {"proposal", "check"}, {"process", "json"},
		{"inspect"}, {"revisions"}, {"change", "prepare"}, {"change", "check"},
		{"change"}, {"change", "help"}, {"change", "--help"}, {"change", "unknown"},
		{"change", "check", "extra"}, {"change", "execute", "extra"},
		{"change", "execute", "--source-root", "PRIVATE"}, {"change", "check", "execute"},
	} {
		t.Run(strings.Join(args, "/"), func(t *testing.T) {
			calls := 0
			harmless := func(context.Context, []string, io.Reader, io.Writer, io.Writer) int { return 0 }
			execute := func(context.Context, []string, io.Reader, io.Writer, io.Writer) int { calls++; return 99 }
			c := capabilities{
				prepare: harmless, review: harmless, check: harmless, execute: execute,
				process: harmless, inspect: harmless, revisions: harmless,
				changePrepare: harmless, changeCheck: harmless, changeExecute: execute,
			}
			runWith(context.Background(), args, unreadableInput{}, io.Discard, io.Discard, c)
			if calls != 0 {
				t.Fatal("non-executing route received execution capability")
			}
		})
	}
}

func TestChangeDispatchPreservesSelectedCapabilityContextAndStreams(t *testing.T) {
	for _, action := range []string{"prepare", "check", "execute"} {
		t.Run(action, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			input := &unreadableInput{}
			var output, diagnostic bytes.Buffer
			calls := 0
			tail := []string(nil)
			if action == "prepare" {
				tail = []string{"--patch-request", "PRIVATE"}
			}
			selected := func(got context.Context, args []string, in io.Reader, out, diag io.Writer) int {
				calls++
				if got != ctx || got.Err() != context.Canceled || in != input || out != &output || diag != &diagnostic || strings.Join(args, "\x00") != strings.Join(tail, "\x00") {
					t.Fatal("capability context, arguments or streams changed")
				}
				return 73
			}
			c := capabilities{}
			switch action {
			case "prepare":
				c.changePrepare = selected
			case "check":
				c.changeCheck = selected
			case "execute":
				c.changeExecute = selected
			}
			args := append([]string{"change", action}, tail...)
			if code := runWith(ctx, args, input, &output, &diagnostic, c); code != 73 || calls != 1 {
				t.Fatal("explicit capability dispatch", code, calls)
			}
		})
	}
}

func TestChangeInvalidNativeShapesRefuseWithoutReadingInput(t *testing.T) {
	for _, args := range [][]string{
		{"change"}, {"change", "help"}, {"change", "--help"}, {"change", "bogus"},
		{"change", "check", "extra"}, {"change", "execute", "extra"},
		{"change", "execute", "--source-root", "/PRIVATE"},
	} {
		var out, diag bytes.Buffer
		if code := Run(context.Background(), args, unreadableInput{}, &out, &diag); code != 2 || out.Len() != 0 || diag.String() != "modelchangerehearsal: usage: modelchangerehearsal check|execute < request.json\n" {
			t.Fatal("argument refusal", code, diag.String())
		}
	}
}

func TestChangeHelpNamesBothExecutionRoutesWithoutGrantingCapabilities(t *testing.T) {
	want := "Usage: foundry prepare --intake FILE --requirements FILE --model FILE --suite FILE --manifest FILE\n" +
		"       foundry review\n" +
		"       foundry proposal check|execute\n" +
		"       foundry process json|svg\n" +
		"       foundry inspect\n" +
		"       foundry revisions\n" +
		"       foundry change prepare --patch-request FILE --baseline-proposal FILE --candidate-suite FILE --candidate-manifest FILE --expected-candidate-digest HEX\n" +
		"       foundry change check|execute\n\n" +
		"All reports use the native tool contracts. Only proposal execute and change execute run candidates.\n"
	for _, args := range [][]string{nil, {"help"}, {"--help"}} {
		var out, diag bytes.Buffer
		if code := runWith(context.Background(), args, unreadableInput{}, &out, &diag, capabilities{}); code != 0 || diag.Len() != 0 || out.String() != want {
			t.Fatal("honest execution help", code, diag.String(), out.String())
		}
	}
}
