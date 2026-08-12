package kindapi

import (
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

type DrivenOut struct {
	out *outport.Out
}

func newDrivenOut(out *outport.Out) DrivenOut { return DrivenOut{out: out} }

func NewDrivenOutForTest(out *outport.Out) DrivenOut { return DrivenOut{out: out} }

func (d DrivenOut) Wired() bool { return d.out.Wired() }

func (d DrivenOut) Paced() bool { return d.out.Paced() }

func (d DrivenOut) Steps() int { return d.out.Geom().Steps }

func (d DrivenOut) PlaceDrivenAt(v int, tick int64) outport.DriveItem {
	return d.out.PlaceDrivenAt(v, tick)
}
