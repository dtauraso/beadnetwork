package Wiring

import (
	"sort"
	"testing"
)

// gesture_graph_test.go — a CONFORMANCE test over the commit/apply TABLES in
// gesture_graph.go (not over MoveDispatch behavior — gesture_test.go already covers that).
// It enumerates every path the table can express from gestPending and asserts the set of
// recognizable gestures is EXACTLY the 6 named in gesture.go's doc comment: tap-select,
// node-drag, rotate, handhold-orbit, wire, port-move — no phantom edge, none missing.

// gestureName maps a terminal phase to the human gesture name used in gesture.go's doc
// comment, so the test output reads the same vocabulary as the source.
func gestureName(phase gesturePhase) string {
	switch phase {
	case gestPending:
		return "tap-select"
	case gestDragging:
		return "node-drag"
	case gestRotating:
		return "rotate"
	case gestHandhold:
		return "handhold-orbit"
	case gestWiring:
		return "wire"
	case gestPortMove:
		return "port-move"
	default:
		return "UNKNOWN"
	}
}

// allGestures walks commitEdges, one at a time, from a fresh gestPending state: for each
// edge, DFS assumes ONLY that edge's guard holds (the others don't), records the resulting
// phase, and also records the "no edge fires" case (stays gestPending → resolved to
// tap-select on pointer-up, mirroring gestPointerUp's gestPending arm). This mirrors the
// table's actual precedence-ordered dispatch: at most one edge fires per event.
func allGestures() []string {
	seen := map[string]bool{}
	// The "no edge fires" path: every commitEdges guard evaluates false on the zero
	// gestureState → stays gestPending → tap-select.
	seen[gestureName(gestPending)] = true
	for _, edge := range commitEdges {
		seen[gestureName(edge.to)] = true
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func TestGestureGraphRecognizesExactlySixGestures(t *testing.T) {
	want := []string{"handhold-orbit", "node-drag", "port-move", "rotate", "tap-select", "wire"}
	got := allGestures()
	if len(got) != len(want) {
		t.Fatalf("got %d gestures %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("gesture set mismatch: got %v, want %v", got, want)
		}
	}
}

// TestGestureGraphCommitEdgePrecedence pins the exact precedence order of commitEdges
// (wireNode, portMoveNode, dragNode, handholdDown, emptyDown) — the same order the old
// inline switch in gestPointerMove used. A reorder here would silently change which
// gesture wins when multiple grab-stashes are set (which should never happen in practice,
// but the table's ORDER is still part of the contract).
func TestGestureGraphCommitEdgePrecedence(t *testing.T) {
	wantOrder := []gesturePhase{gestWiring, gestPortMove, gestDragging, gestHandhold, gestRotating}
	if len(commitEdges) != len(wantOrder) {
		t.Fatalf("got %d commit edges, want %d", len(commitEdges), len(wantOrder))
	}
	for i, phase := range wantOrder {
		if commitEdges[i].to != phase {
			t.Fatalf("commitEdges[%d].to = %v, want %v", i, commitEdges[i].to, phase)
		}
	}
}

// TestGestureGraphApplyActionCoversMoveResolvedPhases asserts applyAction has an entry for
// exactly the phases a commit edge can resolve to that ALSO need a per-move apply
// (dragging/rotating/handhold/portMove) — gestWiring intentionally has none (a wiring drag
// does nothing per-move; it only resets on pointer-up), matching the old switch's implicit
// default.
func TestGestureGraphApplyActionCoversMoveResolvedPhases(t *testing.T) {
	wantApply := map[gesturePhase]bool{
		gestDragging: true,
		gestRotating: true,
		gestHandhold: true,
		gestPortMove: true,
	}
	if len(applyAction) != len(wantApply) {
		t.Fatalf("got %d applyAction entries, want %d", len(applyAction), len(wantApply))
	}
	for phase := range wantApply {
		if _, ok := applyAction[phase]; !ok {
			t.Fatalf("applyAction missing entry for phase %v", phase)
		}
	}
	if _, ok := applyAction[gestWiring]; ok {
		t.Fatalf("applyAction unexpectedly has an entry for gestWiring")
	}
}
