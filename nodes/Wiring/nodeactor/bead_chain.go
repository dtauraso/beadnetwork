package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/spatial"
	beadchain "github.com/dtauraso/wirefold/nodes/wire/beadchain"
)

type edgeBeadChain struct {
	group *beadchain.BeadWakeGroup
	beads []*beadchain.Bead

	stops []chan struct{}

	snaps []<-chan beadchain.BeadSnapshot

	last  []beadchain.BeadSnapshot
	valid []bool

	haveAim bool
	lastAim spatial.Vec3

	haveLattice bool
	lattice     float64
}

func (nb *nodeBeads) SetBeadTickFn(fn func() <-chan struct{}) {
	nb.beadTickFn = fn
}

func (nb *nodeBeads) ReconcileBeadChain(to string, count int, offsetAt func(i int) float64, aim spatial.Vec3) *edgeBeadChain {
	if nb.beadTickFn == nil {
		return nil
	}
	if nb.beadChains == nil {
		nb.beadChains = map[string]*edgeBeadChain{}
	}
	c := nb.beadChains[to]
	if c == nil {
		c = &edgeBeadChain{group: beadchain.NewBeadWakeGroup()}
		nb.beadChains[to] = c
	}

	if lat := offsetAt(0); !c.haveLattice || lat != c.lattice {
		for i := range c.stops {
			close(c.stops[i])
		}
		c.beads, c.stops, c.snaps, c.last, c.valid = nil, nil, nil, nil, nil
		c.haveLattice, c.lattice = true, lat

		c.haveAim = false
	}

	for len(c.beads) < count {
		i := len(c.beads)
		geom, wake, settle := c.group.Current()
		stop := make(chan struct{})
		b := beadchain.NewBead(offsetAt(i), geom, wake, settle, nb.beadTickFn(), stop)
		snap := b.WithObserve()
		b.Start()
		c.beads = append(c.beads, b)
		c.stops = append(c.stops, stop)
		c.snaps = append(c.snaps, snap)
		c.last = append(c.last, beadchain.BeadSnapshot{})
		c.valid = append(c.valid, false)
	}

	for len(c.beads) > count {
		last := len(c.beads) - 1
		close(c.stops[last])
		c.beads = c.beads[:last]
		c.stops = c.stops[:last]
		c.snaps = c.snaps[:last]
		c.last = c.last[:last]
		c.valid = c.valid[:last]
	}
	if !c.haveAim || aim != c.lastAim {

		c.group.BroadcastGeometry(beadchain.BeadGeometryIn{Aim: aim})
		c.lastAim = aim
		c.haveAim = true
	}

	for i, ch := range c.snaps {
		select {
		case s := <-ch:
			c.last[i] = s
			c.valid[i] = true
		default:
		}
	}
	return c
}

func (nb *nodeBeads) StartBeadDrag() {
	for _, c := range nb.beadChains {
		c.group.StartDrag()
	}
}

func (nb *nodeBeads) EndBeadDrag() {
	for _, c := range nb.beadChains {
		c.group.EndDrag()
	}
}
