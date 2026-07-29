// edge_set_length_test.go — what ONE edgeMover decides when told "your length is now L".
//
// Single goroutine on purpose (docs/testing-shape.md): handle() is called directly, no
// run() launched, nothing on the far end of dstOut. This asserts the edge's own
// arithmetic and the SHAPE of what it emits — not that the message is delivered.
package Wiring

import (
	"math"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// edgeAt builds a bare edgeMover whose two endpoints sit at the given world positions.
func edgeAt(t *testing.T, src, dst vec3) *edgeMover {
	t.Helper()
	em := newEdgeMover(
		EdgeEndpoints{Source: "S", Target: "D", SourceHandle: "out", TargetHandle: "in"},
		"SToD", nodeGeom{}, nodeGeom{}, T.NewWithSinkHook(nil, nil), wire.NewRealClock(),
	)
	setNodeWorld(&em.srcGeom, src)
	setNodeWorld(&em.dstGeom, dst)
	em.dstOut = make(chan moveMsg, 1)
	return em
}

// TestEdgeSetLengthEmitsDeltaToTargetOnly pins the core of the distance-group rework: the
// edge converts a target LENGTH into a DISPLACEMENT for its target endpoint, using only
// geometry it owns, and sends a delta rather than a position.
func TestEdgeSetLengthEmitsDeltaToTargetOnly(t *testing.T) {
	// Endpoints 10 apart on +x. Asking for 15 must displace the target by +5x.
	em := edgeAt(t, vec3{X: 100, Y: 0, Z: 0}, vec3{X: 110, Y: 0, Z: 0})
	em.handle(moveMsg{Kind: moveMsgKindSetLength, TargetLen: 15})

	select {
	case msg := <-em.dstOut:
		if msg.Kind != moveMsgKindMoveDelta {
			t.Fatalf("kind = %q, want %q", msg.Kind, moveMsgKindMoveDelta)
		}
		if msg.NodeID != "D" {
			t.Fatalf("NodeID = %q, want the TARGET %q (the source does not move)", msg.NodeID, "D")
		}
		// A DELTA, not a position. 115 (the absolute answer) must never appear here:
		// that is the bug this message shape exists to prevent.
		want := vec3{X: 5, Y: 0, Z: 0}
		if math.Abs(msg.MoveDelta.X-want.X) > 1e-9 ||
			math.Abs(msg.MoveDelta.Y-want.Y) > 1e-9 ||
			math.Abs(msg.MoveDelta.Z-want.Z) > 1e-9 {
			t.Fatalf("MoveDelta = %+v, want %+v (a displacement, not the absolute 115)", msg.MoveDelta, want)
		}
	default:
		t.Fatal("edge emitted nothing to its target endpoint")
	}
}

// TestEdgeSetLengthShrinks covers the down arrow: a target length SHORTER than the current
// one must move the target back along the same direction, not flip it past the source.
func TestEdgeSetLengthShrinks(t *testing.T) {
	em := edgeAt(t, vec3{}, vec3{X: 0, Y: 10, Z: 0})
	em.handle(moveMsg{Kind: moveMsgKindSetLength, TargetLen: 4})
	msg := <-em.dstOut
	if math.Abs(msg.MoveDelta.Y-(-6)) > 1e-9 {
		t.Fatalf("MoveDelta.Y = %v, want -6 (shrink 10 -> 4 along +y)", msg.MoveDelta.Y)
	}
}

// TestEdgeSetLengthDegenerateEmitsNothing pins the cases with no meaningful direction.
// Emitting anything here would move a node somewhere nobody asked for; the old code
// guarded the same case with `if offset.Length() == 0 { continue }`.
func TestEdgeSetLengthDegenerateEmitsNothing(t *testing.T) {
	for _, tc := range []struct {
		name      string
		src, dst  vec3
		targetLen float64
	}{
		{"coincident endpoints (no direction)", vec3{X: 7}, vec3{X: 7}, 5},
		{"non-positive target length", vec3{}, vec3{X: 10}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em := edgeAt(t, tc.src, tc.dst)
			em.handle(moveMsg{Kind: moveMsgKindSetLength, TargetLen: tc.targetLen})
			select {
			case msg := <-em.dstOut:
				t.Fatalf("emitted %+v; want nothing (no meaningful displacement exists)", msg.MoveDelta)
			default:
			}
		})
	}
}

// TestEdgeSetLengthUnpositionedEndpointEmitsNothing: an endpoint with no real position
// makes nodeWorldPos return the origin, so proceeding would compute a displacement toward
// a fabricated point.
func TestEdgeSetLengthUnpositionedEndpointEmitsNothing(t *testing.T) {
	em := edgeAt(t, vec3{X: 1}, vec3{X: 5})
	em.dstGeom.HasPos = false
	em.handle(moveMsg{Kind: moveMsgKindSetLength, TargetLen: 9})
	select {
	case msg := <-em.dstOut:
		t.Fatalf("emitted %+v; want nothing (target has no real position)", msg.MoveDelta)
	default:
	}
}
