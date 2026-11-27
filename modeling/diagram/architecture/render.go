package architecture

import (
	"fmt"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgx"
)

const (
	margin     = 24
	nodeWidth  = 288
	portPitch  = 16
	routePitch = 28
)

type row struct{ top, height int }

func render(m model) []byte {
	var b strings.Builder
	printf := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }
	text := func(x, y int, class, value string) {
		printf("<text x=\"%d\" y=\"%d\" class=\"%s\">%s</text>\n", x, y, class, svgx.Esc(value))
	}
	// Each node has its own row and each edge gets its own track. Spread
	// each node's incoming and outgoing ports so no two horizontal legs coincide.
	degree := map[string]int{}
	for _, l := range m.Links {
		degree[l.From]++
		degree[l.To]++
	}
	titleLines := svgx.Wrap(m.Title, 28, 5)
	cursor := 42 + len(titleLines)*24
	rows := map[string]row{}
	for _, n := range m.Nodes {
		height := max(138, (degree[n.ID]+1)*portPitch)
		rows[n.ID] = row{cursor, height}
		cursor += height + 18
	}
	legendY := cursor + 14
	legend := make([][]string, len(m.Links))
	width := max(640, margin+nodeWidth+44+routePitch*len(m.Links)+margin)
	height := cursor + 12
	if len(legend) > 0 {
		height = legendY + 10
		for i, l := range m.Links {
			legend[i] = svgx.Wrap(l.From+" → "+l.To+" · "+l.Label, 40, 6)
			height += len(legend[i])*19 + 9
		}
		height += 14
	}

	printf("<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 %d %d\" role=\"img\" aria-labelledby=\"title desc\">\n", width, height)
	printf("<title id=\"title\">%s</title>\n", svgx.Esc(m.Title))
	b.WriteString(`<desc id="desc">Logical architecture map. Numbered tracks point from the caller to its dependency. The key names the endpoints and interaction.</desc>
<style>
text{font-family:system-ui,sans-serif;font-size:13px;fill:#16243a}
.heading{font-size:20px;font-weight:700}.label,.key{font-weight:600}
.kind{fill:#324869}.number{font-weight:700;text-anchor:middle}
.track{fill:none;stroke:#435879;stroke-width:2;marker-end:url(#arrow)}
</style>
<defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="5" markerHeight="5" orient="auto"><path d="M0 0 L10 5 L0 10 Z" fill="#435879"/></marker></defs>
`)
	printf("<rect width=\"%d\" height=\"%d\" rx=\"10\" fill=\"#fff\"/>\n", width, height)
	for i, line := range titleLines {
		text(margin, 32+i*24, "heading", line)
	}
	for _, n := range m.Nodes {
		r := rows[n.ID]
		printf("<g data-node=\"%s\" data-kind=\"%s\">\n", n.ID, n.Kind)
		printf("<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" rx=\"8\" fill=\"#edf3fc\" stroke=\"#647d9d\"/>\n", margin, r.top, nodeWidth, r.height)
		for i, line := range svgx.Wrap(n.Kind+" · "+n.ID, 20, 3) {
			text(margin+13, r.top+19+i*14, "kind", line)
		}
		for i, line := range svgx.Wrap(n.Label, 19, 5) {
			text(margin+13, r.top+63+i*15, "label", line)
		}
		b.WriteString("</g>\n")
	}
	next := map[string]int{}
	port := func(id string) int {
		next[id]++
		r := rows[id]
		return r.top + (r.height-(degree[id]-1)*portPitch)/2 + (next[id]-1)*portPitch
	}
	right := margin + nodeWidth
	for i, l := range m.Links {
		x := right + 42 + i*routePitch
		startY, endY := port(l.From), port(l.To)
		printf("<path data-link=\"%s\" data-from=\"%s\" data-to=\"%s\" class=\"track\" d=\"M %d %d H %d V %d H %d\"/>\n",
			l.ID, l.From, l.To, right, startY, x, endY, right)
		printf("<rect x=\"%d\" y=\"%d\" width=\"22\" height=\"14\" rx=\"6\" fill=\"#fff\"/>\n", right+4, startY-7)
		text(right+15, startY+4, "number", svgx.Itoa(i+1))
	}
	if len(legend) != 0 {
		text(margin, legendY, "key", "Dependencies (caller → dependency)")
		y := legendY + 23
		for i, lines := range legend {
			text(margin+12, y, "number", svgx.Itoa(i+1))
			for _, line := range lines {
				text(margin+33, y, "entry", line)
				y += 19
			}
			y += 9
		}
	}
	b.WriteString("</svg>\n")
	return []byte(b.String())
}
