package kindapi

import (
	"github.com/dtauraso/wirefold/src/Node/wire/outport"
)

type DrivenOut struct {
	out *outport.Out
}

func newDrivenOut(out *outport.Out) DrivenOut { return DrivenOut{out: out} }

func NewDrivenOutForTest(out *outport.Out) DrivenOut { return DrivenOut{out: out} }

func (d DrivenOut) HasRun() bool { return d.out.HasRun() }

func (d DrivenOut) Paced() bool { return d.out.Paced() }

func (d DrivenOut) Steps() int { return d.out.Geom().Steps }

func (d DrivenOut) PlaceDrivenAt(v int, tick int64) outport.DriveItem {
	return d.out.PlaceDrivenAt(v, tick)
}
