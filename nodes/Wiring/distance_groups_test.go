// distance_groups_test.go — the "distance home button" controller's synchronous bounds
// check: ApplyDistanceGroupTarget with an out-of-range group index must return false and
// touch nothing, without needing the mover network running.
package Wiring

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

func repoRootForDistanceGroupsTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// nodes/Wiring/distance_groups_test.go -> repo root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestDistanceGroupOutOfRangeIndexIsNoOp guards ApplyDistanceGroupTarget's bounds
// check: an out-of-range group index must return false and move nothing. This is a pure
// bounds check (out-of-range return happens before any mover is touched), so the mover
// network never needs to be started.
func TestDistanceGroupOutOfRangeIndexIsNoOp(t *testing.T) {
	root := filepath.Join(repoRootForDistanceGroupsTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	if ok := md.ApplyDistanceGroupTarget(99, 1); ok {
		t.Fatal("ApplyDistanceGroupTarget(99, up) = true, want false (out of range)")
	}
}

// TestDistanceGroupLensMatchEndpointSeparationAtLoad pins the change of OWNER for a group's
// lengths. distanceGroupMax used to derive each pair's length on the dispatch goroutine by
// subtracting two foreign nodes' mirrored centers; it now reduces over lengths each edge
// measured and published itself (edgeMover.publishLength → moverRegistry.lengthOfPair).
//
// The failure this exists for is SILENT. lengthOfPair returning !ok for every pair — an
// edge keyed differently than the group table's (source,target), or a length mirror never
// seeded — makes distanceGroupMax report any=false, which makes ApplyDistanceGroupTarget a
// no-op and DistanceGroupLens read 0. Nothing panics and no other test notices: the button
// simply stops working and the panel reads zero.
//
// Single goroutine on purpose (docs/testing-shape.md): LoadTopology builds the movers but
// md.Start is never called, so this asserts what the DISPATCH goroutine alone computes from
// the load-time seed — not that any two goroutines communicate.
func TestDistanceGroupLensMatchEndpointSeparationAtLoad(t *testing.T) {
	root := filepath.Join(repoRootForDistanceGroupsTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	for _, group := range distanceGroupOrder {
		pairs := distanceGroups[group]
		// Independent expectation, computed the OLD way (straight from the loaded node
		// centers) so this is a real cross-check of the new path, not a restatement of it.
		want := 0.0
		for _, p := range pairs {
			cs, okS := md.centerOfNode(p.Source)
			ct, okT := md.centerOfNode(p.Target)
			if !okS || !okT {
				t.Fatalf("group %q pair %s->%s: centers unresolvable at load", group, p.Source, p.Target)
			}
			if d := ct.Sub(cs).Length(); d > want {
				want = d
			}
		}
		got, ok := md.distanceGroupMax(group)
		if !ok {
			t.Fatalf("distanceGroupMax(%q) reported unresolved; every pair must have a published length at load", group)
		}
		if want <= 0 {
			t.Fatalf("group %q: expectation is %v, so this test could not detect a zeroed length", group, want)
		}
		if diff := got - want; diff > 1e-9 || diff < -1e-9 {
			t.Fatalf("distanceGroupMax(%q) = %v, want %v (edge-published length must equal endpoint separation)", group, got, want)
		}
	}
	// Every pair must resolve individually — a group's MAX can be right while one member
	// silently reports nothing, and that member would then never be repositioned.
	for _, group := range distanceGroupOrder {
		for _, p := range distanceGroups[group] {
			if _, ok := md.mr.lengthOfPair(p.Source, p.Target); !ok {
				t.Fatalf("group %q: no edge published a length for pair %s->%s", group, p.Source, p.Target)
			}
		}
	}
}
