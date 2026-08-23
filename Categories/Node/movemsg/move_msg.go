package movemsg

import (
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

const (
	KindCenter = "center"

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
	Up bool

	Start bool

	Reset bool
}

type Msg struct {
	Kind   string
	NodeID string

	Port    string
	IsInput bool

	Center *Vec3

	Delta *polarindex.Offset

	Target *polarindex.Index

	SenderID string

	Bool bool
}
