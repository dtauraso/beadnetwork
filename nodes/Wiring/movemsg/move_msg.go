package movemsg

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type vec3 = spatial.Vec3

const (
	KindAnchor  = "anchor"
	KindCenter  = "center"
	KindCenters = "centers"

	KindDrag = "drag"

	KindDragStart = "dragStart"

	KindDragEnd = "dragEnd"

	KindSelect = "select"

	KindHover = "hover"

	KindLatched = "latched"

	KindTiltVectorAngle = "tiltVectorAngle"

	KindTiltVectorReset = "tiltVectorReset"
)

type TiltEditMsg struct {
	Axis string
	Up   bool

	Start bool

	Reset bool
}

type Msg struct {
	Kind   string
	NodeID string

	Port     string
	IsInput  bool
	AnchorId int

	Center *vec3

	// Delta is a polar triple whose meaning is the message's Kind, and both
	// meanings are the same shape: three numbers the receiver adds to something
	// of its own, never a position it reads off someone else.
	//
	// On KindCenter it is how far the SENDER moved, told to the node at the
	// other end of an edge so it can take that move onto its own side of it.
	//
	// On KindDrag it is how far a POINTER is asking THIS node to move — the hit
	// converted against this node's own centre, and nothing else. It arrives
	// UNTRIMMED on purpose: what the node will actually take of it is the
	// node's own answer, given by nodeactor.NodeGeometry.TrimOwnDrag on the
	// node's own goroutine, off its own kind, orbit rule and edge sides. A
	// sender that trimmed first would be deciding a node's constraints for it,
	// which is what this field exists to stop.
	Delta *polar.Polar

	Centers map[string]vec3

	ReachR float64

	FromCenter vec3

	SenderID string

	Target vec3

	// TargetPolar is where a KindDrag is putting the node, AS THE TRIPLE THE
	// SENDER COMPOSED. When it is set the receiver commits it verbatim and
	// Target is only what that triple looks like in the world.
	//
	// It exists because a triple does not survive the world. Compose is the
	// three numbers added and does not fold — phi may pass the pole, which is
	// the whole reason a composed constraint stays exactly the number the
	// constraint named. Cart2polar MUST fold: it answers in a canonical range,
	// and the fold pays for itself by rewriting the other two components. So a
	// held point that went out through Polar2cart and came back through
	// Cart2polar arrived as a DIFFERENT triple standing at the same place.
	//
	// Measured: the load-time hold pinned D.phi to pi/2 about node 1, whose own
	// phi is 3.0600999696029443, composing 4.630896296397841 — past the pole.
	// The round trip returned 2*pi - that = 1.652289010781745, which is node 2's
	// stored phi to the last digit, with theta shifted by pi. The pin was
	// applied correctly every time and then undone in transit, and since a drag
	// only ever trims a delta and holds `have.Phi`, nothing could correct it
	// afterwards. Sending the triple is what makes the constraint mean
	// something once the node has it.
	TargetPolar *polar.Polar

	Bool bool

	Axis string

	TestDone chan struct{}
}
