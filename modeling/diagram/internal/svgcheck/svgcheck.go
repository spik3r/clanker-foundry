// Package svgcheck holds test helpers for the UML-style diagram profiles: an
// inertness check and a small element reader built on encoding/xml. It is
// used by tests only.
package svgcheck

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Elem is one parsed SVG element with its attributes and text.
type Elem struct {
	Name  string
	Attr  map[string]string
	Text  string // character data inside this element, children's included
	Depth int
}

var onAttr = regexp.MustCompile(`^on`)
var urlRef = regexp.MustCompile(`url\([^)]*\)`)

// Parse reads every element of an SVG document in document order. It fails on
// malformed XML.
func Parse(svg []byte) ([]Elem, error) {
	d := xml.NewDecoder(bytes.NewReader(svg))
	var out []Elem
	var stack []int
	for {
		t, err := d.Token()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		switch t := t.(type) {
		case xml.StartElement:
			e := Elem{Name: t.Name.Local, Attr: map[string]string{}, Depth: len(stack)}
			for _, a := range t.Attr {
				e.Attr[a.Name.Local] = a.Value
			}
			out = append(out, e)
			stack = append(stack, len(out)-1)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			for _, i := range stack {
				out[i].Text += string(t)
			}
		}
	}
}

// Inert reports why an SVG is not inert: script, event attributes, external
// references, foreignObject, embedded images, or a reference other than a
// fragment url(#id). It returns nil for an inert picture.
func Inert(svg []byte) error {
	elems, err := Parse(svg)
	if err != nil {
		return err
	}
	for _, e := range elems {
		switch strings.ToLower(e.Name) {
		case "script", "foreignobject", "image", "use", "a", "iframe", "object", "embed", "link":
			return fmt.Errorf("element <%s> is not allowed", e.Name)
		}
		for k, v := range e.Attr {
			lk := strings.ToLower(k)
			lv := strings.ToLower(v)
			if onAttr.MatchString(lk) {
				return fmt.Errorf("event attribute %s", k)
			}
			if lk == "href" || lk == "src" || strings.Contains(lv, "javascript:") {
				return fmt.Errorf("reference attribute %s=%q", k, v)
			}
			if (strings.Contains(lv, "http:") || strings.Contains(lv, "https:") || strings.Contains(v, "//")) &&
				!(e.Name == "svg" && k == "xmlns") {
				return fmt.Errorf("external reference in %s=%q", k, v)
			}
		}
	}
	s := string(svg)
	for _, bad := range []string{"<script", "<foreignObject", "<!ENTITY", "<!DOCTYPE"} {
		if strings.Contains(s, bad) {
			return fmt.Errorf("document contains %s", bad)
		}
	}
	for _, e := range elems {
		for k, v := range e.Attr {
			for _, m := range urlRef.FindAllString(v, -1) {
				if !strings.HasPrefix(m, "url(#") {
					return fmt.Errorf("non-fragment %s in %s", m, k)
				}
			}
		}
	}
	return nil
}

// Ints parses a space-separated integer list such as a data-box value.
func Ints(s string) ([]int, error) {
	var out []int
	for _, f := range strings.Fields(s) {
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

// Rect is an axis-aligned rectangle.
type Rect struct{ X, Y, W, H int }

// Box parses an "x y w h" attribute.
func Box(s string) (Rect, error) {
	n, err := Ints(s)
	if err != nil || len(n) != 4 {
		return Rect{}, fmt.Errorf("bad box %q", s)
	}
	return Rect{n[0], n[1], n[2], n[3]}, nil
}

// Overlap reports whether the interiors of a and b intersect.
func (a Rect) Overlap(b Rect) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

// Contains reports whether b lies inside a, borders included.
func (a Rect) Contains(b Rect) bool {
	return b.X >= a.X && b.Y >= a.Y && b.X+b.W <= a.X+a.W && b.Y+b.H <= a.Y+a.H
}

// Points parses a path of "M x y L x y ..." commands made only of M and L, the
// only commands the UML profiles emit.
func Points(d string) ([][2]int, error) {
	f := strings.Fields(strings.NewReplacer("M", " ", "L", " ").Replace(d))
	if len(f)%2 != 0 || len(f) == 0 {
		return nil, fmt.Errorf("bad path %q", d)
	}
	var out [][2]int
	for i := 0; i < len(f); i += 2 {
		x, err1 := strconv.Atoi(f[i])
		y, err2 := strconv.Atoi(f[i+1])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("bad path %q", d)
		}
		out = append(out, [2]int{x, y})
	}
	return out, nil
}
