package model

import (
	"sort"
	"sync"
	"testing"
)

func TestReferenceCapabilitiesOwnership(t *testing.T) {
	a := ReferenceCapabilities()
	if len(a) != 15 || !sort.SliceIsSorted(a, func(i, j int) bool { return a[i].Path < a[j].Path }) {
		t.Fatal("registry inventory")
	}
	saved := a[0]
	a[0].Path = "changed"
	b := ReferenceCapabilities()
	if b[0] != saved {
		t.Fatal("returned slice aliases registry")
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); r := ReferenceCapabilities(); r[0].Path = "changed" }()
	}
	wg.Wait()
	if ReferenceCapabilities()[0] != saved {
		t.Fatal("concurrent mutation leaked")
	}
	if typ, e := TypeOf(Expr{Kind: "ref", Ref: saved.Path}); e != nil || typ != saved.Type {
		t.Fatal("registry mutated", e)
	}
}
