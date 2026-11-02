// Package views projects validated design intent; it does not execute behavior.
package views

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

// Identity records separate model and build identities, supplied by the compiler.
// Render does not authenticate these declarations or recompute the model digest.
type Identity struct {
	ModelDigest      string `json:"modelDigest"`
	FormatVersion    string `json:"formatVersion"`
	ProfileVersion   string `json:"profileVersion"`
	GeneratorVersion string `json:"generatorVersion"`
	ToolchainVersion string `json:"toolchainVersion"`
}
type Artifacts struct{ SVG, DecisionTable []byte }

const intent = "Generated design intent; invented example; authority none; no verified execution or approval evidence."

func validateIdentity(id Identity) error {
	b, err := hex.DecodeString(id.ModelDigest)
	if err != nil || len(b) != 32 || strings.ToLower(id.ModelDigest) != id.ModelDigest {
		return fmt.Errorf("invalid canonical model digest")
	}
	if id.ProfileVersion != model.ProfileVersion {
		return fmt.Errorf("unsupported semantic profile")
	}
	for _, s := range []string{id.FormatVersion, id.GeneratorVersion, id.ToolchainVersion} {
		if strings.TrimSpace(s) == "" || !utf8.ValidString(s) || len(s) > 256 {
			return fmt.Errorf("invalid build identity")
		}
	}
	return nil
}
func normalized(m model.Model, id Identity) (model.Model, error) {
	if err := validateIdentity(id); err != nil {
		return model.Model{}, err
	}
	return model.Normalize(m)
}
func encode(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Render normalizes semantic sets through the core IR, preserving effect order.
func Render(m model.Model, id Identity) (Artifacts, error) {
	m, err := normalized(m, id)
	if err != nil {
		return Artifacts{}, err
	}
	table := struct {
		Schema           string      `json:"schema"`
		Status           string      `json:"status"`
		Identity         Identity    `json:"identity"`
		GuardCombination string      `json:"guardCombination"`
		EffectOrder      string      `json:"effectOrder"`
		Model            model.Model `json:"model"`
	}{"lifecycle-decision-table/v1", intent, id, "all referenced guards must hold; no first-match priority", "effects execute in array order atomically", m}
	data, err := encode(table)
	if err != nil {
		return Artifacts{}, err
	}
	return Artifacts{SVG: svg(m, id), DecisionTable: data}, nil
}
func escaped(s string) string {
	var b bytes.Buffer
	// Valid IR text can include XML-forbidden control characters. Quoted text
	// exposes those characters without dropping or substituting their meaning.
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
func expression(e model.Expr) string {
	switch e.Kind {
	case "ref":
		return e.Ref
	case "string":
		return strconv.Quote(e.Value)
	case "op":
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = expression(a)
		}
		return e.Op + "(" + strings.Join(args, ", ") + ")"
	}
	// Normalize rejects unknown expression kinds before this function is called.
	panic("unvalidated expression")
}
func effectText(e model.Effect, t model.Transition) string {
	switch e.Op {
	case model.AssignOwner:
		return e.Op + "(command.actor)"
	case model.SetState:
		return e.Op + "(" + t.To + ")"
	case model.IncrementRevision:
		return e.Op
	}
	panic("unvalidated effect")
}
func svg(m model.Model, id Identity) []byte {
	// A transition-per-row diagram avoids hidden edges and crossing ambiguity.
	// Every row explicitly connects the source and destination state boxes.
	var lines []string
	lines = append(lines, "Model "+m.ID+"; initial: "+m.Initial, intent,
		"Digest: "+id.ModelDigest, "Format: "+id.FormatVersion+"; profile: "+id.ProfileVersion,
		"Generator: "+id.GeneratorVersion+"; toolchain: "+id.ToolchainVersion,
		"Guards are conjunctions. Effects follow the numbered order. Text uses Go string escaping.")
	for _, s := range m.States {
		lines = append(lines, "State "+s.ID+": "+s.Label+"; terminal="+strconv.FormatBool(s.Terminal))
	}
	var b strings.Builder
	// Fixed character-cell wrapping keeps all bounded model text in the viewport.
	type row struct {
		text     string
		diagram  bool
		from, to string
	}
	rows := []row{}
	add := func(s string) {
		q := strconv.Quote(s)
		r := []rune(q[1 : len(q)-1])
		if len(r) == 0 {
			rows = append(rows, row{})
			return
		}
		for len(r) > 0 {
			n := 100
			if len(r) < n {
				n = len(r)
			}
			rows = append(rows, row{text: string(r[:n])})
			r = r[n:]
		}
	}
	for _, s := range lines {
		add(s)
	}
	for _, t := range m.Transitions {
		rows = append(rows, row{text: t.ID + " / " + t.Trigger, diagram: true, from: t.From, to: t.To})
		add("Transition " + t.ID + " guards (all): " + strings.Join(t.Guards, ", "))
		for i, e := range t.Effects {
			add(fmt.Sprintf("Effect %d: %s", i+1, effectText(e, t)))
		}
	}
	for _, g := range m.Guards {
		add("Guard " + g.ID + "; requirement " + g.Requirement + ": " + expression(g.Expr))
	}
	for _, r := range m.Requirements {
		add("Requirement " + r.ID + ": " + r.Text)
	}
	h := 40 + len(rows)*24
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"1200\" height=\"%d\" viewBox=\"0 0 1200 %d\" role=\"img\">\n<title>Lifecycle design intent: %s</title>\n<rect width=\"1200\" height=\"%d\" fill=\"white\"/>\n", h, h, escaped(m.ID), h)
	for i, r := range rows {
		y := 30 + i*24
		if r.diagram {
			fmt.Fprintf(&b, "<rect x=\"20\" y=\"%d\" width=\"350\" height=\"22\" fill=\"#eef\" stroke=\"#223\"/><rect x=\"810\" y=\"%d\" width=\"350\" height=\"22\" fill=\"#eef\" stroke=\"#223\"/><path d=\"M 375 %d H 800 l -8 -5 m 8 5 l -8 5\" fill=\"none\" stroke=\"#223\"/>\n", y-16, y-16, y-5)
			fmt.Fprintf(&b, "<text x=\"25\" y=\"%d\" font-family=\"monospace\" font-size=\"11\">%s</text><text x=\"815\" y=\"%d\" font-family=\"monospace\" font-size=\"11\">%s</text>\n", y, escaped(r.from), y, escaped(r.to))
			fmt.Fprintf(&b, "<text x=\"380\" y=\"%d\" font-family=\"monospace\" font-size=\"7\">%s</text>\n", y-7, escaped(r.text))
		} else {
			fmt.Fprintf(&b, "<text x=\"20\" y=\"%d\" font-family=\"monospace\" font-size=\"11\">%s</text>\n", y, escaped(r.text))
		}
	}
	b.WriteString("</svg>\n")
	return []byte(b.String())
}
