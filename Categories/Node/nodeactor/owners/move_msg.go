package owners

import (
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

type Msg struct {
	NodeID string

	Body Body
}

type Body interface{ moveBody() }

type Movement interface {
	Body

	Where() *polarindex.Index
	HowFar() *polarindex.Offset

	WithHowFar(polarindex.Offset) Movement
}

type NeighborMoved struct {
	Center *Vec3

	Delta *polarindex.Offset

	SenderID string
}

func (NeighborMoved) moveBody()                    {}
func (NeighborMoved) Where() *polarindex.Index     { return nil }
func (n NeighborMoved) HowFar() *polarindex.Offset { return n.Delta }
func (n NeighborMoved) WithHowFar(d polarindex.Offset) Movement {
	n.Delta = &d
	return n
}

type Drag struct {
	Target *polarindex.Index

	Delta *polarindex.Offset
}

func (Drag) moveBody()                    {}
func (d Drag) Where() *polarindex.Index   { return d.Target }
func (d Drag) HowFar() *polarindex.Offset { return d.Delta }
func (d Drag) WithHowFar(o polarindex.Offset) Movement {
	d.Delta = &o
	return d
}

type DragStart struct{}

func (DragStart) moveBody() {}

type DragEnd struct{}

func (DragEnd) moveBody() {}

type Select struct{ On bool }

func (Select) moveBody() {}

type Hover struct {
	On bool

	Port    string
	IsInput bool
}

func (Hover) moveBody() {}

type Latched struct{ On bool }

func (Latched) moveBody() {}

type TiltVectorAngle struct{ Up bool }

func (TiltVectorAngle) moveBody() {}

type TiltVectorReset struct{}

func (TiltVectorReset) moveBody() {}

type TiltEditMsg struct {
	Up bool

	Start bool

	Reset bool
}
