// Package NormalSum holds two normals and draws their total.
//
// A NORMAL HERE IS A θ LATTICE INDEX, not an angle and not a vector. The whole tilt-vector
// model is θ-only integer indices times one step constant, and every direction derived in it
// — bottom is θ+12, the coplanar normal is θ+6 — is index arithmetic
// (memory/feedback_abc_times_constant_not_rederive.md: arithmetic on indices, trig only at
// the cartesian boundary). Taking a normal as an index keeps this node inside that
// arithmetic rather than introducing a second spelling of a direction that would have to be
// converted back.
//
// The node holds each input as it arrives and the TOTAL of the two, and its own drawn tilt
// vector IS that total: it calls SetTiltIndex, so the arrow every node already draws from
// its TopTiltVectorTheta column points along the sum. No new buffer column and no new
// renderer — the drawing of "this node's own direction" already exists, and a total normal
// is exactly that.
package NormalSum

import (
	"context"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// noNormal marks an input that has not arrived yet. -1 rather than 0, because 0 is a
// perfectly good normal (the lattice's own zero index) and a node that has received one 0
// must be distinguishable from one that has received nothing.
const noNormal = -1

// Node holds two normals and their total.
//
// ONE GOROUTINE OWNS ALL OF IT. a, b and total are plain fields read and written only by
// Update's own loop; nothing else in the process touches them, which is why there is no lock
// here and no channel to itself.
type Node struct {
	Fire  func()
	Clock clock.Clock
	// Self is this node's own handle for driving its own geometry — SetTiltIndex, which is
	// what makes the total DRAWN rather than merely held.
	Self *nodeactor.PairNodeSelf
	// Points is the lattice's point count, the modulus the total wraps at. Seeded at load
	// and updated live, the same as every other pair-scene node's copy.
	Points int32

	// Field names ARE the port names: the generator reflects these to build the kind's
	// NODE_DEFS entry, so a field called In would declare a port called In.
	NormalA *wire.In
	NormalB *wire.In
	Out     *wire.Out
	SpeedCh <-chan float64
	// LatticeIn carries live point-count changes from the scene setting.
	LatticeIn <-chan int32

	a, b, total int32
}

// Update polls both inputs, keeps the latest of each, and republishes the total whenever it
// changes.
func (n *Node) Update(ctx context.Context) {
	c := n.Clock.Copy()
	n.a, n.b, n.total = noNormal, noNormal, noNormal

	// THIS NODE OWNS ITS OWN GEOMETRY. Claiming self-drive means no nodeMover actor is
	// constructed for it (mover_registry.go's finalizeActors skips a claimed id), so the
	// per-cycle work a mover would do — drain this node's inbound channels, drive its
	// outgoing wires, write its own stream frame — is this loop's, and the startup emit is
	// this loop's too.
	//
	// Missing both is what made a created NormalSum INVISIBLE: it loaded, it linked, and it
	// never streamed a geometry frame, so there was nothing to draw and no row to select or
	// delete. A node that claims itself and then does not drive itself does not exist on
	// screen.
	n.Self.EmitGeometryOnce()

	for {
		if ctx.Err() != nil {
			return
		}
		clock.ApplySpeedNonBlocking(c, n.SpeedCh)
		select {
		case pts := <-n.LatticeIn:
			if pts > 0 {
				n.Points = pts
			}
		default:
		}

		changed := false
		// Drain to the LATEST on each input: a normal is a direction, not a queue of
		// directions — an older one arriving behind a newer says nothing about where the
		// source is pointing now.
		for {
			v, ok := n.NormalA.PollRecv()
			if !ok {
				break
			}
			n.a = int32(v)
			changed = true
		}
		for {
			v, ok := n.NormalB.PollRecv()
			if !ok {
				break
			}
			n.b = int32(v)
			changed = true
		}

		if changed {
			n.Fire()
			n.republish()
		}
		// One cycle of this node's own geometry work, on this loop's own clock reading —
		// the same body nodeMover.run drives for a node that has a mover. It takes no sleep
		// of its own: this loop already paces itself, and a second sleep would double-pace
		// the same node.
		n.Self.Step(ctx, c.Tick())
		if err := c.SleepCycle(ctx); err != nil {
			return
		}
	}
}

// republish recomputes the total and, when it moved, draws and emits it.
//
// Both inputs are required. A node holding one normal has nothing to total — showing the one
// it has would draw a direction nobody asked for, and this node's whole meaning is the sum.
func (n *Node) republish() {
	if n.a == noNormal || n.b == noNormal {
		return
	}
	total := sumIndex(n.a, n.b, n.Points)
	if total == n.total {
		return
	}
	n.total = total
	// The DRAWN total: this node's own tilt vector, with its derived directions kept in the
	// same relationship every other node has (the coplanar normal a quarter turn on, the
	// bottom a half turn), so the arrow reads the same way here as anywhere else.
	quarter := n.Points / 4
	half := n.Points / 2
	n.Self.SetTiltIndex(total, wrapIndex(total+quarter, n.Points), wrapIndex(total+half, n.Points))
	// …and the total travels on, so a downstream node can take it as a normal in turn. The
	// placement tick is this goroutine's own clock reading, taken once for this placement —
	// the same shape every other kind's emit has.
	n.Out.PlaceDrivenAt(int(total), n.Clock.Tick())
}

// sumIndex adds two lattice indices and wraps at the lattice's own point count. SUM, not
// average: two normals pointing the same way should read as twice as far around, and an
// average would make the output depend on how many inputs happened to have arrived.
func sumIndex(a, b, points int32) int32 {
	return wrapIndex(a+b, points)
}

// wrapIndex brings an index back into [0, points). A lattice index is modular by
// construction — the step constant times points is one full turn — so this is the lattice's
// own arithmetic, not a clamp.
func wrapIndex(i, points int32) int32 {
	if points <= 0 {
		return 0
	}
	i %= points
	if i < 0 {
		i += points
	}
	return i
}

func init() {
	Wiring.RegisterBuilder("NormalSum",
		[]portwiring.PortSpec{
			{Name: "NormalA", Dir: portwiring.PortIn},
			{Name: "NormalB", Dir: portwiring.PortIn},
			{Name: "Out", Dir: portwiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{}
			n.Fire = a.Fire()
			n.Clock = a.Clock()
			n.SpeedCh = a.SpeedCh()
			n.Self = a.ClaimSelfDrive()
			n.Points = a.LatticePointsSeed()
			n.LatticeIn = a.LatticeIn()
			n.NormalA = a.In("NormalA")
			n.NormalB = a.In("NormalB")
			n.Out = a.Out("Out")
			return n, nil
		})
}
