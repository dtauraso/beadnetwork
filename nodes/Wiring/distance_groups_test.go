// distance_groups_test.go — the "distance home button" controller: driving one
// arrow click (ApplyDistanceGroupTarget) through the REAL production topology
// (topology/ dir), mirroring abc_drag_count_target_node_test.go's harness
// (LoadTopology + md.Start + the real move entry point).
package Wiring

import (
	"context"
	"math"
	"path/filepath"
	"runtime"
	"testing"
	"time"

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

func loadProductionTopologyForDistanceGroups(t *testing.T) *MoveDispatch {
	t.Helper()
	root := filepath.Join(repoRootForDistanceGroupsTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	md.Start(ctx)
	return md
}

// waitForPairLength polls (source,target)'s live center-to-center length until it
// settles within tol of want, or fails after a deadline. RootMove is a FIRE-AND-FORGET
// message to the target node's OWN goroutine (moveMsgKindDrag) — its commit, and the
// one-hop neighborSetC re-quantize it fans to every OTHER node linked to the target,
// land asynchronously on their own goroutines, so asserting immediately after
// ApplyDistanceGroupTarget returns races the real commit. Mirrors
// ui_publish_propagation_test.go's waitForNodeDragMsg polling pattern.
func waitForPairLength(t *testing.T, md *MoveDispatch, source, target string, want, tol float64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var last float64
	for {
		last = pairLength(t, md, source, target)
		if math.Abs(last-want) <= tol {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("pair (%s,%s) length never settled to %.6f (last=%.6f)", source, target, want, last)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func pairLength(t *testing.T, md *MoveDispatch, source, target string) float64 {
	t.Helper()
	cs, ok := md.centerOfNode(source)
	if !ok {
		t.Fatalf("no center for node %q", source)
	}
	ct, ok := md.centerOfNode(target)
	if !ok {
		t.Fatalf("no center for node %q", target)
	}
	return ct.Sub(cs).Length()
}

// TestDistanceGroupUpMovesEveryPairToTargetLength drives group index 0 ("time":
// (2,5),(2,4),(4,7),(4,6)) up: currentMax*1.1 is the new target length L, and every
// pair's TARGET node must end up at distance L from its (unmoved) SOURCE — asserted
// directly off the live centers RootMove's own commit path updates.
func TestDistanceGroupUpMovesEveryPairToTargetLength(t *testing.T) {
	md := loadProductionTopologyForDistanceGroups(t)

	group := "time"
	currentMax, ok := md.distanceGroupMax(group)
	if !ok {
		t.Fatalf("distanceGroupMax(%q): no resolvable pair", group)
	}
	wantL := currentMax * 1.1

	if ok := md.ApplyDistanceGroupTarget(0, 1); !ok {
		t.Fatalf("ApplyDistanceGroupTarget(0, up) returned false")
	}

	const tol = 1e-3
	for _, p := range distanceGroups[group] {
		waitForPairLength(t, md, p.Source, p.Target, wantL, tol)
	}

	// The Overlay GroupLenTime column (read-only reflect) must now report the NEW max,
	// which is wantL (every pair in the group was set to the same L).
	gotTime, _, _ := md.DistanceGroupLens()
	if math.Abs(float64(gotTime)-wantL) > tol {
		t.Errorf("DistanceGroupLens() time = %.6f, want %.6f", gotTime, wantL)
	}
}

// TestDistanceGroupDownHalvesTowardCurrentMax mirrors the up test for the down arrow
// (÷1.1) on group index 1 ("input": (1,3),(1,2)).
func TestDistanceGroupDownHalvesTowardCurrentMax(t *testing.T) {
	md := loadProductionTopologyForDistanceGroups(t)

	group := "input"
	currentMax, ok := md.distanceGroupMax(group)
	if !ok {
		t.Fatalf("distanceGroupMax(%q): no resolvable pair", group)
	}
	wantL := currentMax / 1.1

	if ok := md.ApplyDistanceGroupTarget(1, -1); !ok {
		t.Fatalf("ApplyDistanceGroupTarget(1, down) returned false")
	}

	const tol = 1e-3
	for _, p := range distanceGroups[group] {
		waitForPairLength(t, md, p.Source, p.Target, wantL, tol)
	}
}

// TestDistanceGroupGateLastWriteWinsForSharedTargets documents and asserts the
// accepted last-write-wins behavior for the gate group's two SHARED target nodes:
// node 8 is the target of both (3,8) and (5,8); node 9 of both (5,9) and (7,9).
// The group's flat pair list is applied IN ORDER with no tree/graph solver and no
// averaging — the LAST pair touching a shared target wins, so after an up click node
// 8 ends at distance L from node 5 (not node 3), and node 9 ends at distance L from
// node 7 (not node 5). This is the agreed model (CLAUDE.md GO-LAYER MODEL), not a bug.
func TestDistanceGroupGateLastWriteWinsForSharedTargets(t *testing.T) {
	md := loadProductionTopologyForDistanceGroups(t)

	group := "gate"
	pairs := distanceGroups[group]
	if len(pairs) != 4 {
		t.Fatalf("gate group pair count = %d, want 4", len(pairs))
	}
	// Sanity-check the exact flat order the model specifies: (3,8),(5,8),(5,9),(7,9).
	want := []distancePair{{"3", "8"}, {"5", "8"}, {"5", "9"}, {"7", "9"}}
	for i, p := range want {
		if pairs[i] != p {
			t.Fatalf("gate group pair[%d] = %+v, want %+v", i, pairs[i], p)
		}
	}

	currentMax, ok := md.distanceGroupMax(group)
	if !ok {
		t.Fatalf("distanceGroupMax(gate): no resolvable pair")
	}
	wantL := currentMax * 1.1

	if ok := md.ApplyDistanceGroupTarget(2, 1); !ok {
		t.Fatalf("ApplyDistanceGroupTarget(2, up) returned false")
	}

	const tol = 1e-3
	// Node 8's FINAL length is measured against its LAST pair's source, (5,8) — not
	// (3,8), which was applied first and then overwritten by node 8's own second
	// RootMove call.
	waitForPairLength(t, md, "5", "8", wantL, tol)
	// Node 9's FINAL length is measured against its LAST pair's source, (7,9).
	waitForPairLength(t, md, "7", "9", wantL, tol)
}

// TestDistanceGroupOutOfRangeIndexIsNoOp guards ApplyDistanceGroupTarget's bounds
// check: an out-of-range group index must return false and move nothing.
func TestDistanceGroupOutOfRangeIndexIsNoOp(t *testing.T) {
	md := loadProductionTopologyForDistanceGroups(t)
	if ok := md.ApplyDistanceGroupTarget(99, 1); ok {
		t.Fatal("ApplyDistanceGroupTarget(99, up) = true, want false (out of range)")
	}
}
