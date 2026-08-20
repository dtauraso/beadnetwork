package beadchain

type BeadWakeGroup struct {
	geom   *BroadcastChain
	wake   *BroadcastChain
	settle *BroadcastChain
}

func NewBeadWakeGroup() *BeadWakeGroup {
	return &BeadWakeGroup{
		geom:   NewBroadcastChain(),
		wake:   NewBroadcastChain(),
		settle: NewBroadcastChain(),
	}
}

func (g *BeadWakeGroup) Current() (geom, wake, settle *BroadcastChain) {
	return g.geom, g.wake, g.settle
}

func (g *BeadWakeGroup) BroadcastGeometry(xf BeadGeometryIn) {
	g.geom = g.geom.AdvanceWithValue(xf)
}

func (g *BeadWakeGroup) StartDrag() {
	g.wake = g.wake.Advance()
}

func (g *BeadWakeGroup) EndDrag() {
	g.settle = g.settle.Advance()
}
