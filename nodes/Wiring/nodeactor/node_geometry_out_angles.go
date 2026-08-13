package nodeactor

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

// OutAngleKind is the one kind whose outgoing paths are angle-constrained.
// It is the SPEC kind name, which is PascalCase — the Go package directory
// for this kind is lowercase `nodes/input/` and the two do not have to agree.
const OutAngleKind = "Input"

// outAngleEps is how far off the constrained angles a stored path may sit
// before it counts as violating. A corrected path comes back to this node as
// a neighbour centre and is converted back to angles, so it returns carrying
// float noise rather than the exact values that were sent; without a
// tolerance that noise alone would read as a fresh violation every frame.
const outAngleEps = 1e-9

// outAngleMaxFixes bounds how many times in a row one target may be corrected
// without the correction taking.
const outAngleMaxFixes = 8

// ConstrainOutAngles holds this node's outgoing paths at the angles an input
// node is allowed to have: phi = pi/2 and |theta| <= pi/2, measured in the
// node's own pole triple. Nothing else about the path moves — its length is
// what sets this node's emission cadence.
//
// The path is the stored truth and the neighbour's centre is derived from it,
// so a path whose angles are corrected is a neighbour that has moved. This
// node does the arithmetic — it is the one that knows the constraint — and
// sends each affected neighbour ONE position, which that neighbour commits in
// its own goroutine through the ordinary move path. This node never writes
// another node's position, and the neighbour never recomputes the constraint.
func (m *NodeGeometry) ConstrainOutAngles() {
	if m.selfKind != OutAngleKind {
		return
	}
	self := m.WorldCenter()
	sharedLen := m.establishSharedOutLen()
	for _, to := range m.outTargets {
		have, ok := m.topo.PolarPathTo(to)
		if !ok {
			continue
		}
		want := polar.ClampOutAngles(have)
		want.R = sharedLen
		if math.Abs(want.Phi-have.Phi) <= outAngleEps &&
			math.Abs(want.Theta-have.Theta) <= outAngleEps &&
			math.Abs(want.R-have.R) <= outAngleEps {
			m.topo.ClearOutAngleFix(to)
			continue
		}
		m.topo.SetPolarPathTo(to, want)
		m.countOutAngleFix(to)
		target := self.Add(polar.Polar2cart(want))
		m.msg.SendMove()(to, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: to, Target: target})
		m.traceOutAngleFix(to, have, want, target)
	}
}

// establishSharedOutLen is the length every outgoing path is held at.
//
// Normally it is simply the length already declared — set by whichever
// neighbour last moved, so the node that was dragged states the distance and
// its siblings are brought to it rather than the drag being undone. Because
// it survives this node's OWN move untouched, dragging this node carries its
// neighbours along at their distance instead of stretching the paths.
//
// The first time round there is nothing declared yet, so the longest path
// wins: at load the paths disagree, and growing the short one is the choice
// that never pulls a node inward past something it was already clear of.
func (m *NodeGeometry) establishSharedOutLen() float64 {
	if r, ok := m.topo.SharedOutLen(); ok {
		return r
	}
	longest := 0.0
	for _, to := range m.outTargets {
		if p, ok := m.topo.PolarPathTo(to); ok && p.R > longest {
			longest = p.R
		}
	}
	m.topo.SetSharedOutLen(longest)
	return longest
}

// NoteOutNeighborLen takes a neighbour's own new distance as the length ALL
// the outgoing paths are now held at. It is what makes the constraint hold
// whichever node moved: the mover states the distance, and the correction
// that follows brings the others to it.
func (m *NodeGeometry) NoteOutNeighborLen(from string) {
	if m.selfKind != OutAngleKind || !m.IsOutTarget(from) {
		return
	}
	// A neighbour still working through a correction is reporting a
	// position THIS node asked for, so it has no distance of its own to
	// state. Letting it restate one is how a follower that lands slightly
	// off — bead snapping moves it again after it commits — becomes the new
	// length, which the node that was actually dragged then gets corrected
	// to, and the two trade the length back and forth forever.
	if m.topo.HasPendingOutAngleFix(from) {
		return
	}
	if p, ok := m.topo.PolarPathTo(from); ok && p.R != 0 {
		m.topo.SetSharedOutLen(p.R)
	}
}

// traceOutAngleFix records which angles were out of range and where the
// corrected ones put that neighbour, so a hold that fires when it should not
// — or never fires at all — is visible without attaching to the process.
func (m *NodeGeometry) traceOutAngleFix(to string, have, want polar.Polar, target vec3) {
	if m.tr == nil {
		return
	}
	value := fmt.Sprintf(
		"to=%s havePhi=%.6f haveTheta=%.6f haveR=%.4f wantPhi=%.6f wantTheta=%.6f wantR=%.4f target=(%.4f,%.4f,%.4f)",
		to, have.Phi, have.Theta, have.R, want.Phi, want.Theta, want.R, target.X, target.Y, target.Z)
	m.tr.Breadcrumb("out-angle-fix", m.id, to, value)
	targetRow := int32(-1)
	if r, ok := m.topo.NodeRowFor(to); ok {
		targetRow = r
	}
	m.writeStreamFrame([]rowevent.RowEvent{{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbOutAngleFix, Debug: 1,
		NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: targetRow, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

// countOutAngleFix panics once a target has been corrected this many times
// running without ever coming back in range. That means the position this
// node derives from the constrained path is not the position the target
// commits — the target's own commit is moving it somewhere else, so the two
// are trading corrections and neither the paths nor the layout will settle.
func (m *NodeGeometry) countOutAngleFix(to string) {
	if m.topo.BumpOutAngleFix(to) > outAngleMaxFixes {
		sharedLen, _ := m.topo.SharedOutLen()
		panic(fmt.Sprintf(
			"NodeGeometry(%s): outgoing path constraint to %s did not converge in %d "+
				"corrections — the centre derived from phi=pi/2, |theta|<=pi/2 and the shared "+
				"length %.4f is not the centre that target commits, so the two are trading "+
				"corrections",
			m.id, to, outAngleMaxFixes, sharedLen))
	}
}
