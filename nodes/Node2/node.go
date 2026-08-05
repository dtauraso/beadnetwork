// Package Node2 is the "Node2" kind — the mirror of Node1: a self-paced source
// that periodically emits a white bead (value 1, the on-wire white value per
// bead-style.ts's VALUE_BEAD_STYLE) to its single output, at human speed, on
// its OWN node goroutine. Emission is UNCONDITIONAL — periodic, not a reply to
// anything received — so pairing Node1/Node2 with one edge each direction has
// no bootstrap/deadlock problem and needs no seed Input node.
//
// It also has one input port (In); receiving on it never gates or paces
// sending — the two are fully independent (same shape as Node1; kept as a
// distinct package/kind rather than parametrizing Node1, per the
// check-dep-rules guard: a node-kind package may import only the shared spine,
// never a sibling kind, so Node1 and Node2 cannot share a struct across
// packages without a third shared package neither currently needs).
package Node2

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// whiteValue is the on-wire bead value that renders white
// (tools/topology-vscode/src/webview/three/bead-style.ts VALUE_BEAD_STYLE: 0 →
// black, 1 → white).
const whiteValue = 1

type Node struct {
	Fire         func()
	EmitGeometry func()
	// Clock is this node's OWN clock storage, assigned by this kind's own
	// builder directly from the loader's origin (per-goroutine-clock.md).
	// Update() Copies it once for its own loop.
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
// derivation as input.inputCadenceTicks / Node1.cadenceTicks) — one bead
// fully crosses the wire before the next is placed. Recomputed live so a drag
// that changes the edge's step count re-paces emission.
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
	// Node2 CONSTRUCTS ITSELF (Wiring.RegisterBuilder), same self-construction
	// shape as every other kind.
	Wiring.RegisterBuilder("Node2",
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
