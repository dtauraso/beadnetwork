package movemsg

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
	"github.com/dtauraso/wirefold/src/Node/spatial"
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

	KindDragPhiToggle = "dragPhiToggle"

	KindDragMaxTheta = "dragMaxTheta"

	KindDragActiveToggle = "dragActiveToggle"
)

type TiltEditMsg struct {
	Up bool

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

	Delta *polarindex.Offset

	Target *polarindex.Index

	Centers map[string]vec3

	FromCenter vec3

	SenderID string

	Bool bool

	DragMaxTheta *float64

	TestDone chan struct{}
}
