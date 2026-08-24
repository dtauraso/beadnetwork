package time

import (
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
)

type DrivenOut struct {
	out *beadanimation.Sender
}

func newDrivenOut(out *beadanimation.Sender) DrivenOut { return DrivenOut{out: out} }

func (d DrivenOut) HasRun() bool { return d.out.HasRun() }

func (d DrivenOut) Paced() bool { return d.out.Paced() }

func (d DrivenOut) Steps() int { return d.out.Geom().Steps }

func (d DrivenOut) PlaceDrivenAt(v int, tick int64) beadanimation.DriveItem {
	return d.out.PlaceDrivenAt(v, tick)
}
