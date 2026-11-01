package processview

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// cappedBuffer stops retaining output immediately after its documented limit.
type cappedBuffer struct {
	b      bytes.Buffer
	failed bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if b.failed || len(p) > MaxOutputBytes-b.b.Len() {
		b.failed = true
		return 0, errors.New("output limit")
	}
	return b.b.Write(p)
}

// displayText exposes control and non-ASCII scalars as Go-style escapes before
// XML escaping. This preserves content while keeping XML 1.0 text well formed.
func displayText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\\' {
			b.WriteString(`\\`)
		} else if r >= 32 && r <= 126 {
			b.WriteRune(r)
		} else {
			q := strconv.QuoteToASCII(string(r))
			b.WriteString(q[1 : len(q)-1])
		}
	}
	return b.String()
}
func xmlText(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
func wrapped(s string, n int) []string {
	s = displayText(s)
	if s == "" {
		return []string{""}
	}
	var out []string
	for len(s) > n {
		out = append(out, s[:n])
		s = s[n:]
	}
	return append(out, s)
}
func renderSVG(d document) ([]byte, error) {
	var body cappedBuffer
	y := 30
	label := func(x, at int, s string) {
		fmt.Fprintf(&body, "<text x=\"%d\" y=\"%d\" font-family=\"monospace\" font-size=\"11\">%s</text>\n", x, at, xmlText(s))
	}
	line := func(s string) {
		for _, v := range wrapped(s, 160) {
			label(20, y, v)
			y += 18
		}
	}
	line("Lifecycle process review / canonical digest: " + d.ModelDigest)
	line("Model ID: " + d.ModelID)
	line("Example only; authority none. Not BPMN, execution or approval.")
	line(d.Semantics.Scope)
	line(d.Semantics.Guards)
	line(d.Semantics.RevisionLimit)
	line(d.Semantics.Effects)
	line(d.Semantics.Unsupported)
	line("Initial state: " + d.Initial)
	for _, s := range d.States {
		line("State " + s.ID + " / label: " + s.Label + " / terminal=" + strconv.FormatBool(s.Terminal))
	}
	y += 12
	// Each card has one vertical success path. A right-hand branch explicitly
	// marks refusals; the flow never invents parallel or first-match semantics.
	card := func(lines []string, branch string) {
		main := []string{}
		for _, s := range lines {
			main = append(main, wrapped(s, 100)...)
		}
		side := []string{}
		if branch != "" {
			side = wrapped(branch, 65)
		}
		n := len(main)
		if len(side) > n {
			n = len(side)
		}
		h := n*18 + 18
		fmt.Fprintf(&body, "<rect x=\"20\" y=\"%d\" width=\"700\" height=\"%d\" rx=\"6\" fill=\"#eef4fa\" stroke=\"#233\"/>\n", y, h)
		for i, s := range main {
			label(30, y+18+i*18, s)
		}
		if branch != "" {
			fmt.Fprintf(&body, "<path d=\"M 720 %d H 758\" stroke=\"#833\" fill=\"none\"/><rect x=\"760\" y=\"%d\" width=\"450\" height=\"%d\" rx=\"6\" fill=\"#fff2ec\" stroke=\"#833\"/>\n", y+h/2, y, h)
			for i, s := range side {
				label(770, y+18+i*18, s)
			}
		}
		y += h
		fmt.Fprintf(&body, "<path d=\"M 370 %d v 20 m -5 -6 l 5 6 l 5 -6\" fill=\"none\" stroke=\"#233\"/>\n", y)
		y += 25
	}
	for _, t := range d.Transitions {
		line("Transition " + t.ID + " / trigger " + t.Trigger)
		card([]string{"From state " + t.From}, "")
		card([]string{"Command " + t.Trigger}, "")
		card([]string{"All guards: " + strings.Join(t.Guards, ", ")}, "Guard refusal: unchanged state "+t.From+"; failed guard IDs sorted")
		card([]string{"Revision is below the uint64 limit"}, "Revision limit: unchanged state "+t.From)
		effects := make([]string, 0, len(t.Effects))
		for i, e := range t.Effects {
			effects = append(effects, fmt.Sprintf("Effect %d: %s", i+1, e))
		}
		card(effects, "")
		// The destination ends this lane rather than implying another command.
		for _, s := range wrapped("To state "+t.To, 100) {
			label(30, y+18, s)
			y += 18
		}
		y += 30
	}
	line("Guard glossary")
	for _, g := range d.Guards {
		line("Guard " + g.ID + " / requirement " + g.Requirement + ": " + g.Expression)
	}
	if body.failed || y+20 > MaxCanvasHeight {
		return nil, ErrInvalid
	}
	var out cappedBuffer
	fmt.Fprintf(&out, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"1240\" height=\"%d\" viewBox=\"0 0 1240 %d\" role=\"img\">\n<title>Lifecycle process review; not BPMN</title>\n<rect width=\"1240\" height=\"%d\" fill=\"white\"/>\n", y+20, y+20, y+20)
	_, _ = out.Write(body.b.Bytes())
	_, _ = out.Write([]byte("</svg>\n"))
	if out.failed {
		return nil, ErrInvalid
	}
	return out.b.Bytes(), nil
}
