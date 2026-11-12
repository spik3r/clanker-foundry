package process

import (
	"fmt"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

// BPMNNamespace is the fixed targetNamespace of every exported definitions
// element.
const BPMNNamespace = "urn:foundry:diagramgen:process-diagram:v1"

var bpmnElement = map[string]string{
	KindStart:     "startEvent",
	KindEnd:       "endEvent",
	KindTask:      "task",
	KindExclusive: "exclusiveGateway",
	KindParallel:  "parallelGateway",
}

// BPMN renders the model and its layout as BPMN 2.0 XML. Model ids are used
// as element ids unchanged; derived ids contain a dot, which model ids cannot,
// so the two never collide.
func BPMN(m Model, l Layout, g contract.Generator) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL" `+
		`xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" `+
		`xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" `+
		`xmlns:di="http://www.omg.org/spec/DD/20100524/DI" `+
		`id="Definitions.%s" name="%s" targetNamespace="%s" exporter="%s" exporterVersion="%s">`+"\n",
		m.ID, esc(m.Title), BPMNNamespace, g.ID, g.Version)
	fmt.Fprintf(&b, `  <collaboration id="Collaboration.%s">`+"\n", m.ID)
	fmt.Fprintf(&b, `    <participant id="Participant.%s" name="%s" processRef="Process.%s"/>`+"\n", m.ID, esc(m.Title), m.ID)
	b.WriteString("  </collaboration>\n")
	fmt.Fprintf(&b, `  <process id="Process.%s" name="%s" isExecutable="false">`+"\n", m.ID, esc(m.Title))
	fmt.Fprintf(&b, `    <laneSet id="LaneSet.%s">`+"\n", m.ID)
	for _, ln := range m.Lanes {
		fmt.Fprintf(&b, `      <lane id="%s" name="%s">`+"\n", ln.ID, esc(ln.Label))
		for _, n := range m.Nodes {
			if n.Lane == ln.ID {
				fmt.Fprintf(&b, "        <flowNodeRef>%s</flowNodeRef>\n", n.ID)
			}
		}
		b.WriteString("      </lane>\n")
	}
	b.WriteString("    </laneSet>\n")
	for _, n := range m.Nodes {
		el := bpmnElement[n.Kind]
		fmt.Fprintf(&b, `    <%s id="%s" name="%s">`+"\n", el, n.ID, esc(n.Label))
		for _, f := range m.Flows {
			if f.To == n.ID {
				fmt.Fprintf(&b, "      <incoming>%s</incoming>\n", f.ID)
			}
		}
		for _, f := range m.Flows {
			if f.From == n.ID {
				fmt.Fprintf(&b, "      <outgoing>%s</outgoing>\n", f.ID)
			}
		}
		fmt.Fprintf(&b, "    </%s>\n", el)
	}
	for _, f := range m.Flows {
		fmt.Fprintf(&b, `    <sequenceFlow id="%s"`, f.ID)
		if f.Label != "" {
			fmt.Fprintf(&b, ` name="%s"`, esc(f.Label))
		}
		fmt.Fprintf(&b, ` sourceRef="%s" targetRef="%s"/>`+"\n", f.From, f.To)
	}
	b.WriteString("  </process>\n")

	fmt.Fprintf(&b, `  <bpmndi:BPMNDiagram id="Diagram.%s">`+"\n", m.ID)
	fmt.Fprintf(&b, `    <bpmndi:BPMNPlane id="Plane.%s" bpmnElement="Collaboration.%s">`+"\n", m.ID, m.ID)
	pool := l.Lanes[0].Box
	for _, ln := range l.Lanes[1:] {
		pool.H += ln.Box.H
	}
	shape := func(id, ref string, bx Box, extra string) {
		fmt.Fprintf(&b, `      <bpmndi:BPMNShape id="%s" bpmnElement="%s"%s>`+"\n", id, ref, extra)
		fmt.Fprintf(&b, `        <dc:Bounds x="%d" y="%d" width="%d" height="%d"/>`+"\n", bx.X, bx.Y, bx.W, bx.H)
		b.WriteString("      </bpmndi:BPMNShape>\n")
	}
	shape("Shape.Participant."+m.ID, "Participant."+m.ID, pool, ` isHorizontal="true"`)
	for _, ln := range l.Lanes {
		shape("Shape."+ln.Lane.ID, ln.Lane.ID, ln.Box, ` isHorizontal="true"`)
	}
	for _, n := range l.Nodes {
		extra := ""
		if n.Node.Kind == KindExclusive {
			extra = ` isMarkerVisible="true"`
		}
		shape("Shape."+n.Node.ID, n.Node.ID, n.Box, extra)
	}
	for _, f := range l.Flows {
		fmt.Fprintf(&b, `      <bpmndi:BPMNEdge id="Edge.%s" bpmnElement="%s">`+"\n", f.Flow.ID, f.Flow.ID)
		for _, p := range f.Points {
			fmt.Fprintf(&b, `        <di:waypoint x="%d" y="%d"/>`+"\n", p.X, p.Y)
		}
		b.WriteString("      </bpmndi:BPMNEdge>\n")
	}
	b.WriteString("    </bpmndi:BPMNPlane>\n  </bpmndi:BPMNDiagram>\n</definitions>\n")
	return []byte(b.String())
}
