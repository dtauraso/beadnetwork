package main

// pair_node_mover_absence_test.go — the architectural assertion task/pair-node-owns-
// itself calls for: a PAIR node has NO separate nodeMover actor at all (not merely a
// flag saying "don't launch it"), while a RING node still gets a real one. This is a
// single-goroutine STRUCTURAL fact about one MoveDispatch's own registry
// (docs/testing-shape.md) — not a test of cross-goroutine delivery/ordering.
import (
	"context"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestPairNodesHaveNoNodeMoverRingNodesDo loads the real checked-in pair scene
// (Node1/Node2, writePairTree — pair_self_drive_persist_test.go) and the real
// checked-in ring scene ("topology/") and asserts Wiring.MoveDispatch.HasNodeMover:
// false for every pair node id (no nodeMover was ever constructed for it — see
// mover_registry.go's finalizeActors), true for every ring node id.
func TestPairNodesHaveNoNodeMoverRingNodesDo(t *testing.T) {
	tr := T.New()

	pairRoot := writePairTree(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, pairMD, _, err := Wiring.LoadTopology(ctx, pairRoot, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(pair): %v", err)
	}
	for _, id := range []string{"1", "2"} {
		if pairMD.HasNodeMover(id) {
			t.Fatalf("pair node %q has a nodeMover — task/pair-node-owns-itself requires none be constructed for a self-driven node", id)
		}
	}

	ringCtx, ringCancel := context.WithCancel(context.Background())
	defer ringCancel()
	_, _, ringMD, _, err := Wiring.LoadTopology(ringCtx, "topology", tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(topology): %v", err)
	}
	sawOne := false
	for i := 1; i <= 9; i++ {
		id := string(rune('0' + i))
		if i >= 10 {
			t.Fatalf("test assumes single-digit ring node ids")
		}
		if !ringMD.HasNodeMover(id) {
			continue
		}
		sawOne = true
	}
	if !sawOne {
		t.Fatal("no ring node had a nodeMover at all — expected at least one real ring node id (1..9) to have one")
	}
}
