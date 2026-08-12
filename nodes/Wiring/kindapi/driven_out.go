package kindapi

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

type DrivenOut struct {
	out *wire.Out
}

func newDrivenOut(out *wire.Out) DrivenOut { return DrivenOut{out: out} }

func NewDrivenOutForTest(out *wire.Out) DrivenOut { return DrivenOut{out: out} }

func (d DrivenOut) Wired() bool { return d.out.Wired() }

func (d DrivenOut) Paced() bool { return d.out.Paced() }

func (d DrivenOut) Steps() int { return d.out.Geom().Steps }

func (d DrivenOut) PlaceDrivenAt(v int, tick int64) wire.DriveItem {
	return d.out.PlaceDrivenAt(v, tick)
}
