package helddrive

import (
	"github.com/dtauraso/wirefold/src/Node/Interior"
	Wiring "github.com/dtauraso/wirefold/src/Node/Wiring/kindapi"
	lattice "github.com/dtauraso/wirefold/src/Node/BeadAnimation/lattice"
)

const NoValue = interior.NoValue

type HeldDriver struct {
	out       Wiring.DrivenOut
	transform func(int64) int

	cur     int64
	started bool
	stopped bool

	lastPlaceTick int64
}

func New(out Wiring.DrivenOut, transform func(int64) int) *HeldDriver {
	return &HeldDriver{out: out, transform: transform, cur: NoValue}
}

func (d *HeldDriver) Set(held int64) { d.cur = held }

func (d *HeldDriver) Step(tick int64) {
	if d.stopped {
		return
	}

	paced := d.out.Paced()

	if !d.started {
		d.started = true
		if paced {
			d.lastPlaceTick = tick
		}
	}

	place := !paced
	if paced {
		if k, known := heldPeriod(d.out); known {
			place = tick-d.lastPlaceTick >= k
		}
	}
	if !place {
		return
	}

	di := d.out.PlaceDrivenAt(d.transform(d.cur), tick)
	if di.Failed() {

		d.stopped = true
		return
	}
	if !di.BufferFull() && paced {
		d.lastPlaceTick = tick
	}
}

func heldPeriod(out Wiring.DrivenOut) (k int64, known bool) {
	steps := out.Steps()
	if steps <= 0 {
		return 0, false
	}
	k = int64(float64(steps)*lattice.PulsesPerSlot + 0.999999)
	if k < 1 {
		k = 1
	}
	return k, true
}
