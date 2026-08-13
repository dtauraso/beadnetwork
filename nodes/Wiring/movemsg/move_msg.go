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

	// Delta is how far the SENDER moved, as a polar triple, told to the node at
	// the other end of an edge so it can take that move onto its own side of
	// it by adding the three components. The receiver never reads the sender's
	// position; it is told the difference.
	Delta *polar.Polar

	Centers map[string]vec3

	ReachR float64

	FromCenter vec3

	SenderID string

	Target vec3

	Bool bool

	Axis string

	TestDone chan struct{}
}
