// node_geometry_lattice_points_test.go — pins writeStreamFrame's conversion of a tilt-vector
// INDEX to a streamed angle against THIS node's own lattice size (nodeGeometry.latticePoints),
// not the fixed compile-time nodegeom.CurveParamTiltVectorAngleStep — task/pair-lattice-points.
//
// Seam used: writeStreamFrame's own injected buildFrame closure (the same seam
// node_bead_test.go's captureInteriorSnapshot uses for interiorStream) — this is the actual
// streamed frame, not a re-derivation of the formula under test. A live streamOut is required
// too (writeStreamFrame no-ops when !streamOut.Ok()), built via newClaimedStream(nil, ...)
// exactly as production wiring does, just with a discard writer and no claim registry.
package Wiring

import (
	"io"
	"math"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// latticeFrameAngles captures the four θ columns writeStreamFrame packs that this task
// changed: topTiltVectorTheta, bottomTiltVectorTheta, coplanarNormalTheta, receivedVectorTheta.
type latticeFrameAngles struct {
	top, bottom, coplanar, received float32
	points                          uint8
}

func captureLatticeAngles(snap *latticeFrameAngles) *nodeGeometry {
	return &nodeGeometry{
		id:     "n",
		clocks: nodeClocks{clk: clock.NewRealClock()},
		stream: nodeStream{
			streamOut: newClaimedStream(nil, "node", "n", io.Discard),
			buildFrame: func(f NodeFrameInput) []byte {
				snap.top = f.TopTiltVectorTheta
				snap.bottom = f.BottomTiltVectorTheta
				snap.coplanar = f.CoplanarNormalTheta
				snap.received = f.ReceivedVectorTheta
				snap.points = f.LatticePoints
				return nil
			},
		},
	}
}

// TestWriteStreamFrameDefaultLatticeMatchesOldConstant: a nodeGeometry that never calls
// SetLatticePoints (every ring node, and any pair geometry built before this task) must
// stream the SAME four angles it always did — index × nodegeom.CurveParamTiltVectorAngleStep (π/12,
// 2π/24) — because latticePoints' zero value is documented to fall back to
// Wiring.FullTurnThetaIdx (24). This is the "unchanged for everyone who never opts in" half
// of the task.
func TestWriteStreamFrameDefaultLatticeMatchesOldConstant(t *testing.T) {
	const idx = int32(5)
	var snap latticeFrameAngles
	g := captureLatticeAngles(&snap)
	g.tilt.topTiltVectorThetaIdx = idx
	g.tilt.bottomThetaIdx = idx
	g.tilt.normalThetaIdx = idx
	g.tilt.receivedVectorThetaIdx = idx
	g.tilt.receivedVectorSet = true
	g.writeStreamFrame(nil)

	want := float32(float64(idx) * nodegeom.CurveParamTiltVectorAngleStep)
	if snap.top != want || snap.bottom != want || snap.coplanar != want || snap.received != want {
		t.Fatalf("default-lattice angles = (top=%v bottom=%v coplanar=%v received=%v), want all %v (idx=%d * nodegeom.CurveParamTiltVectorAngleStep)",
			snap.top, snap.bottom, snap.coplanar, snap.received, want, idx)
	}
}

// TestWriteStreamFrameFollowsSetLatticePoints: the SAME index streams a DIFFERENT angle once
// this node's own lattice size changes — 2π/12 instead of 2π/24 (== nodegeom.CurveParamTiltVectorAngleStep)
// for index 5 on a 12-point lattice — because writeStreamFrame derives its step from
// THIS node's own latticePoints, not the fixed constant.
func TestWriteStreamFrameFollowsSetLatticePoints(t *testing.T) {
	const idx = int32(5)
	const points = int32(12)
	var snap latticeFrameAngles
	g := captureLatticeAngles(&snap)
	g.tilt.topTiltVectorThetaIdx = idx
	g.tilt.bottomThetaIdx = idx
	g.tilt.normalThetaIdx = idx
	g.tilt.receivedVectorThetaIdx = idx
	g.tilt.receivedVectorSet = true
	g.tilt.latticePoints = points
	g.writeStreamFrame(nil)

	wantStep := 2 * math.Pi / float64(points)
	want := float32(float64(idx) * wantStep)
	oldConstantAngle := float32(float64(idx) * nodegeom.CurveParamTiltVectorAngleStep)
	if want == oldConstantAngle {
		t.Fatalf("test setup error: 12-point step equals the 24-point constant for idx=%d — cases don't distinguish", idx)
	}
	if snap.top != want || snap.bottom != want || snap.coplanar != want || snap.received != want {
		t.Fatalf("12-point-lattice angles = (top=%v bottom=%v coplanar=%v received=%v), want all %v (idx=%d * 2π/12, NOT the fixed nodegeom.CurveParamTiltVectorAngleStep %v)",
			snap.top, snap.bottom, snap.coplanar, snap.received, want, idx, oldConstantAngle)
	}
	if snap.points != uint8(points) {
		t.Fatalf("streamed LatticePoints = %d, want %d — the frame's own point-count column must mirror this node's latticePoints", snap.points, points)
	}
}
