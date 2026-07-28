package Wiring

// delta_forward_test.go — proves the STORED cascade-edges delta-forward model
// (nodes/<id>/cascade-edges.json, specNode.CascadeEdges, nodeMover.cascadeEdges,
// nodeMover.forwardDelta, moveMsgKindDeltaForward): each node's cascade-neighbor list is
// hand-authored/persisted FILE DATA, not derived from the domain-edge/local-polar
// adjacency at load. A node relays to its stored cascade neighbors excluding the sender,
// on EVERY move it receives, so the forwarded log stays in sync with the drag as it
// continues to move (instead of freezing at the first delta, as the old forwardedThisDrag
// guard did).
//
// Real repo topology (topology/) adjacency (edges/*.json):
//
//	1: 2,3   2: 1,4,5   3: 1,8   4: 2,6,7   5: 2,8,9   6: 4   7: 4,9   8: 3,5   9: 5,7
//
// The seeded nodes/<id>/cascade-edges.json files now carry EVERY domain edge — both
// former cycle-closers (5-8, 7-9) are restored, so cascade adjacency equals domain
// adjacency. TestCascadeEdgesLoadedFromStoredFiles pins this explicitly.
import (
	"context"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func repoRootForDeltaForwardTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// nodes/Wiring/delta_forward_test.go -> repo root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestCascadeEdgesLoadedFromStoredFiles pins the exact per-node cascade-edges the real
// topology's nodes/<id>/cascade-edges.json seed files carry: every edge except {5-8,
// 7-9}. If the topology or the seeded files ever change, this test documents and
// enforces the expected result rather than letting it silently drift.
func TestCascadeEdgesLoadedFromStoredFiles(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}

	want := map[string][]string{
		"1": {"2", "3"},
		"2": {"1", "4", "5"},
		"3": {"1", "8"},
		"4": {"2", "6", "7"},
		"5": {"2", "9", "8"},
		"6": {"4"},
		"7": {"4", "9"},
		"8": {"3", "5"},
		"9": {"5", "7"},
	}
	for id, wantEdges := range want {
		nm, ok := md.mr.nodeMovers[id]
		if !ok {
			t.Fatalf("no nodeMover for %q", id)
		}
		got := append([]string(nil), nm.cascadeEdges...)
		sort.Strings(got)
		wantSorted := append([]string(nil), wantEdges...)
		sort.Strings(wantSorted)
		if !reflect.DeepEqual(got, wantSorted) {
			t.Errorf("node %q cascadeEdges = %v, want %v", id, got, wantSorted)
		}
	}
	// Cascade adjacency now EQUALS domain adjacency: both former cycle-closers (5-8, 7-9)
	// are restored, so there is no non-cascade link left to assert about. Termination no
	// longer comes from the edge set — it comes from the per-kind rules (PulseLeft and
	// PulseRight are termini, TimeStart and Pulse route by sender kind). Measured: every
	// single-node drag settles in 1-6 forwards.
	for id, nm := range md.mr.nodeMovers {
		if len(nm.cascadeEdges) == 0 {
			t.Errorf("node %q has no cascade edges; expected cascade adjacency to cover the domain graph", id)
		}
	}
}
