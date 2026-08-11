// bead_advance.go — ONE JOB: the per-bead work of one drive step. advanceBead
// computes where a single in-flight bead is at this tick and whether it has hit
// its deadline; emitArrive announces the delivery once the handoff succeeded. The
// two small argument records those return/take (posEmitArgs, arriveInfo) live here
// with them. The loop that calls both, once per bead per cycle, is stepAll in
// paced_wire_drive.go.

package wire

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// arriveInfo carries the source identity a delivery must echo on the arrive trace.
// emit is false for a bead that carried no position stream.
type arriveInfo struct {
	emit       bool
	node, port string
	value      int
	gen        uint64 // the delivered bead's per-wire id (renderer bead key)
}

// posEmitArgs holds the arguments for a tr.Position call, returned by
// advanceBead so the caller can emit it.
type posEmitArgs struct {
	node, port string
	val        int
	x, y, z, t float64
	gen        uint64
}

// emitArrive sends the traversal-complete trace for a delivered bead. Called by
// this wire's own goroutine (stepAll) right after the outCh handoff succeeds.
func (pw *PacedWire) emitArrive(ai arriveInfo) {
	if ai.emit && pw.readout.StreamsActive {
		pw.readout.appendPending(pendingWireEvent{kind: T.KindArrive, value: ai.value, gen: ai.gen},
			pw.Target, pw.TargetHandle)
	}
}

// advanceBead performs one cycle's work for the in-flight bead b at clock
// reading now (the scheduled tick time). Called only by this wire's own
// goroutine (stepAll).
//
// Returns:
//   - emit=true if a Position trace should be sent (tr.Position) for this call;
//     pos contains the arguments.
//   - final=true if the bead has reached or passed its deadline at now, meaning
//     the caller should attempt the FIFO-head delivery handoff.
func (pw *PacedWire) advanceBead(b *inflightBead, nowTick float64) (emit bool, pos posEmitArgs, final bool) {
	tr := pw.readout.Trace

	steps := b.steps
	seg := b.seg
	placementTick := b.placementTick
	stream := b.streams && tr != nil && steps > 0
	crossTicks := pw.ticksToCross(steps)

	deadline := placementTick + crossTicks
	final = nowTick >= deadline

	if stream {
		// fractional progress t = elapsed ticks / ticksToCross (== steps
		// covered / steps, since both scale by the uniform per-bead dwell) — the
		// same clamp-and-divide as live_beads.go/ReviseInFlightGeometry, shared
		// via lattice.BeadFraction (nodes/wire/lattice/bead_fraction.go) rather
		// than a fourth inline copy of the same math.
		t := lattice.BeadFraction(nowTick, placementTick, crossTicks)
		p := lerp(seg.Start, seg.End, t)
		emit = true
		pos = posEmitArgs{
			node: b.node, port: b.port, val: b.val,
			x: p.X, y: p.Y, z: p.Z, t: t, gen: b.gen,
		}
	}
	return
}
