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

	KindOrbitPhiToggle = "orbitPhiToggle"

	KindOrbitMaxTheta = "orbitMaxTheta"

	KindOrbitActiveToggle = "orbitActiveToggle"
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

	Delta *polar.Polar

	Centers map[string]vec3

	ReachR float64

	FromCenter vec3

	SenderID string

	Bool bool

	Axis string

	OrbitMaxTheta *float64

	TestDone chan struct{}
}
