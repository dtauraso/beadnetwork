package movemsg

import (
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

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

	Center *Vec3

	Delta *polarindex.Offset

	Target *polarindex.Index

	Centers map[string]Vec3

	FromCenter Vec3

	SenderID string

	Bool bool

	DragMaxTheta *float64

	TestDone chan struct{}
}
