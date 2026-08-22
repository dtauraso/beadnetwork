package PairNode

import (
	"fmt"
	"strconv"

	clock "github.com/dtauraso/wirefold/Clock"
	"github.com/dtauraso/wirefold/NodeKinds/nodeapi"

	"github.com/dtauraso/wirefold/NodeKinds/PairNode/tiltring"
	Wiring "github.com/dtauraso/wirefold/NodeKinds/kindapi"
)

func (n *Node) wirePlumbing(a Wiring.BuildArgs) {

	if id, err := strconv.Atoi(a.Name()); err == nil {
		n.plumb.PairID = int32(id)
	}
	n.plumb.Fire = a.Fire()
	if clk := a.Clock(); clk != nil {
		n.plumb.Clock = clk
	}
	n.plumb.SpeedCh = a.SpeedCh()
	n.plumb.In = a.In("In")
	n.plumb.Out = a.Out("Out")
}

func (n *Node) wireLatticeSeed(a Wiring.BuildArgs) (latticeSeed int32, seed *tiltring.State, seedUnknown bool) {
	latticeSeed = a.LatticePointsSeed()
	n.lattice.Ring = tiltring.NewRing(latticeSeed)
	seed, seedUnknown = n.lattice.Ring.SeedState(a.TiltVectorAngleSeed())
	n.setTop(seed)
	n.tilt.TiltEditIn = a.TiltEditIn()
	n.lattice.LatticeIn = a.LatticeIn()
	return latticeSeed, seed, seedUnknown
}

func (n *Node) wireSelfDrive(a Wiring.BuildArgs, latticeSeed int32, seed *tiltring.State, seedUnknown bool) {
	self := a.ClaimSelfDrive()
	n.plumb.Self = self
	n.lattice.SyncLatticePoints = func(points int32) {
		self.SetLatticePoints(points)
	}
	n.lattice.SyncLatticePoints(latticeSeed)
	if seedUnknown {

		self.Breadcrumb("pair-seed-unknown", fmt.Sprintf(
			"node=%s persisted=%d loaded=%d", a.Name(), a.TiltVectorAngleSeed(), seed.Idx))
	}
	n.tilt.SyncTiltIndex = func(theta int32) {
		self.SetTiltIndex(theta)
	}
	n.vec.SyncReceivedVector = func(theta int32, set bool) {
		self.SetReceivedVector(theta, set)
	}
	n.plumb.ClearOutBeads = func() { self.ClearOutBeads() }
}

func (n *Node) wireVectorChannels(a Wiring.BuildArgs) {
	n.vec.VectorOut = a.VectorOut()
	n.vec.VectorIn = a.VectorIn()
}

func init() {

	Wiring.RegisterBuilder("PairNode",
		func(a Wiring.BuildArgs) (nodeapi.Node, error) {
			n := &Node{
				plumb: nodePlumbing{Clock: clock.NewRealClock()},
			}
			n.wirePlumbing(a)
			latticeSeed, seed, seedUnknown := n.wireLatticeSeed(a)
			n.wireSelfDrive(a, latticeSeed, seed, seedUnknown)
			n.wireVectorChannels(a)

			return n, nil
		})
}
