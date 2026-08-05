// Package Node1 is the "Node1" kind: a self-paced source that periodically emits a
// white bead (value 1, the on-wire white value per bead-style.ts's
// VALUE_BEAD_STYLE) to its single output, at human speed, on its OWN node
// goroutine. Emission is UNCONDITIONAL — periodic, not a reply to anything
// received — so pairing two of these (Node1/Node2) with one edge each direction
// has no bootstrap/deadlock problem and needs no seed Input node.
//
// It also has one input port (In); receiving on it never gates or paces
// sending — the two are fully independent, matching MODEL.md's "receives beads
// ... holds them in node-local state until its firing rule is satisfied" node
// shape, except this node's firing rule for SENDING is simply "time has
// elapsed", not "a value arrived".
package Node1

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// whiteValue is the on-wire bead value that renders white
// (tools/topology-vscode/src/webview/three/bead-style.ts VALUE_BEAD_STYLE: 0 →
// black, 1 → white). Node1/Node2 always emit this value — "sends a white bead"
// is the whole firing rule, so there is nothing else to parametrize.
const whiteValue = 1

type Node struct {
	Fire         func()
	EmitGeometry func()
	// Clock is this node's OWN clock storage, assigned by this kind's own
	// builder directly from the loader's origin (per-goroutine-clock.md; see
	// input.Node.Clock for the fuller rationale). Update() Copies it once for
	// its own loop — the sole clock-owning goroutine this node has.
	Clock wire.Clock
	// SpeedCh delivers a speed change to this goroutine's own clk copy.
	// Assigned by this kind's own builder; nil on a test build with no loader.
	SpeedCh <-chan float64
	// In is the sole input. Received values are drained (PollRecv, non-blocking)
	// but otherwise ignored — this node's own send cadence never waits on them.
	In *wire.In
	// Out is the sole output: the periodic white-bead source.
	Out *wire.Out
}

func (n *Node) clock() wire.Clock {
	if n.Clock == nil {
		return wire.NewRealClock()
	}
	return n.Clock
}

// cadenceTicks is this node's own emission period, measured in clock ticks:
// the CROSSING TIME of the Out edge (steps * DwellTicksPerBead, same
// derivation as input.inputCadenceTicks) — one bead fully crosses the wire
// before the next is placed, so beads never overlap on it. Recomputed live so
// a drag that changes the edge's step count re-paces emission.
func cadenceTicks(out *wire.Out) int64 {
	c := int64(float64(out.Geom().Steps) * wire.DwellTicksPerBead)
	if c < 1 {
		return 1
	}
	return c
}

func (n *Node) Update(ctx context.Context) {
	wire.TryEmit(n.EmitGeometry)

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := n.clock().Copy()

	lastFireTick := clk.Tick() - cadenceTicks(n.Out) // fire on the first pass
	for {
		if ctx.Err() != nil {
			return
		}

		// Drain In non-blocking; receiving never gates or paces this node's own
		// send loop below.
		if _, ok := n.In.PollRecv(); ok {
			if n.Fire != nil {
				n.Fire()
			}
		}

		now := clk.Tick()
		if now-lastFireTick >= cadenceTicks(n.Out) {
			if n.Fire != nil {
				n.Fire()
			}
			n.Out.PlaceDrivenAt(whiteValue, now)
			lastFireTick = now
		}

		wire.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {
	// Node1 CONSTRUCTS ITSELF (Wiring.RegisterBuilder), same self-construction
	// shape as every other kind — see Pacer/Input for the general note on why
	// this replaced reflectBuild.
	Wiring.RegisterBuilder("Node1",
		[]Wiring.PortSpec{
			{Name: "In", Dir: Wiring.PortIn},
			{Name: "Out", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{
				Clock: wire.NewRealClock(),
			}
			n.Fire = a.Fire()
			if clk := a.Clock(); clk != nil {
				n.Clock = clk
			}
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.Out = a.Out("Out")
			// EmitGeometry stays nil deliberately — nodeMover/edgeMover emit the
			// same geometry from their own goroutine start.
			return n, nil
		})
}
