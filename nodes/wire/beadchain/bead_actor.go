package beadchain

import "github.com/dtauraso/wirefold/nodes/spatial"

type beadGeometryState struct {
	position spatial.Vec3
}

func (g *beadGeometryState) applyTransform(xf BeadGeometryIn, offsetR float64) {
	g.position = xf.Center.Add(xf.Aim.Scale(offsetR))
}

type beadAnimationState struct {
	lit    bool
	litVal int32
}

func (a *beadAnimationState) tick(lit bool, val int32) {
	a.lit = lit
	a.litVal = val
}

type Bead struct {
	offsetR float64

	geom   *BroadcastChain
	tickCh <-chan struct{}
	wake   *BroadcastChain
	settle *BroadcastChain
	stop   <-chan struct{}

	geomState beadGeometryState
	anim      beadAnimationState
	dragging  bool

	observe chan BeadSnapshot
}

type BeadSnapshot struct {
	Position spatial.Vec3
	Dragging bool
	Lit      bool
	LitVal   int32
}

func NewBead(offsetR float64, geom, wake, settle *BroadcastChain, tickCh <-chan struct{}, stop <-chan struct{}) *Bead {
	return &Bead{
		offsetR: offsetR,
		geom:    geom,
		wake:    wake,
		settle:  settle,
		tickCh:  tickCh,
		stop:    stop,
	}
}

func (b *Bead) WithObserve() <-chan BeadSnapshot {
	b.observe = make(chan BeadSnapshot, 1)
	return b.observe
}

func (b *Bead) pushObserve() {
	if b.observe == nil {
		return
	}
	snap := BeadSnapshot{Position: b.geomState.position, Dragging: b.dragging, Lit: b.anim.lit, LitVal: b.anim.litVal}
	select {
	case b.observe <- snap:
		return
	default:
	}
	select {
	case <-b.observe:
	default:
	}
	select {
	case b.observe <- snap:
	default:
	}
}

// BEAD-SELECT-START
func (b *Bead) run() {
	for {
		select {
		case <-b.geom.Fire:

			b.geomState.applyTransform(b.geom.Value, b.offsetR)
			b.geom = b.geom.Next
			b.pushObserve()
		case <-b.wake.Fire:

			b.dragging = true
			b.wake = b.wake.Next
			b.pushObserve()
		case <-b.settle.Fire:

			b.dragging = false
			b.settle = b.settle.Next
			b.pushObserve()
		case <-b.tickCh:

			b.anim.tick(!b.anim.lit, b.anim.litVal)
			b.pushObserve()
		case <-b.stop:
			return
		}
	}
}

// BEAD-SELECT-END

func (b *Bead) Start() { go b.run() }
