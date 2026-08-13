package nodeactor

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// outAngleKind is the one kind whose outgoing paths are angle-constrained.
const outAngleKind = "input"

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
	if m.selfKind != outAngleKind {
		return
	}
	self := m.WorldCenter()
	for _, to := range m.outTargets {
		have, ok := m.topo.PolarPathTo(to)
		if !ok {
			continue
		}
		want := polar.ClampOutAngles(have)
		if math.Abs(want.Phi-have.Phi) <= outAngleEps && math.Abs(want.Theta-have.Theta) <= outAngleEps {
			m.topo.ClearOutAngleFix(to)
			continue
		}
		m.topo.SetPolarPathTo(to, want)
		m.countOutAngleFix(to)
		m.msg.SendMove()(to, movemsg.Msg{
			Kind: movemsg.KindDrag, NodeID: to, Target: self.Add(polar.Polar2cart(want)),
		})
	}
}

// countOutAngleFix panics once a target has been corrected this many times
// running without ever coming back in range. That means the position this
// node derives from the constrained angles is not the position the target
// commits — the target's own commit is moving it somewhere else, so the two
// are trading corrections and neither the angles nor the layout will settle.
func (m *NodeGeometry) countOutAngleFix(to string) {
	if m.topo.BumpOutAngleFix(to) > outAngleMaxFixes {
		panic(fmt.Sprintf(
			"NodeGeometry(%s): outgoing angle constraint to %s did not converge in %d "+
				"corrections — the centre derived from phi=pi/2, |theta|<=pi/2 is not the "+
				"centre that target commits, so the two are trading corrections",
			m.id, to, outAngleMaxFixes))
	}
}
