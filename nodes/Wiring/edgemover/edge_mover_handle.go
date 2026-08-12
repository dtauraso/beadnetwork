// edge_mover_handle.go — EdgeMover's own inbox message handler (handle) and the geometry
// recompute + propagate it drives (recomputeGeometry). Split out of edge_mover.go by
// concern: the struct + construction + the bound-func-value wiring surface (Select,
// TrySendFromSrc/Dst, SendSteps, the Set*/binding setters) stay there; this file is what
// runs ON this edge's own goroutine once a message has arrived — see edge_mover_run.go's
// Run loop for where handle is called from.
package edgemover

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// handle applies one inbox message to this edge. For a move message it updates the
// matching endpoint geom, recomputes the edge's segment + arc, writes them onto the
// source Out, revises any in-flight bead, emits the new edge geometry, and updates
// the dest port's latency aggregate. A move that touches neither endpoint is ignored.
func (m *EdgeMover) handle(msg movemsg.Msg) {
	if msg.Kind == movemsg.KindSelect {
		if msg.Bool {
			m.selected = 1
		} else {
			m.selected = 0
		}
		return
	}
	if msg.Kind == movemsg.KindCenter {
		// Polar re-propagation: adopt the centrally-computed center on whichever
		// endpoint this message names, then recompute the edge.
		if msg.Center == nil {
			return
		}
		switch msg.NodeID {
		case m.srcID:
			nodegeom.SetNodeWorld(&m.srcGeom, *msg.Center)
		case m.dstID:
			nodegeom.SetNodeWorld(&m.dstGeom, *msg.Center)
		default:
			return
		}
		m.recomputeGeometry()
		return
	}
	if msg.Kind == movemsg.KindCenters {
		// Batched polar re-propagation: apply every moved endpoint this edge owns,
		// then recompute ONCE. An edge whose both endpoints moved in one frame would
		// otherwise recompute (and emit) twice — the duplicate-emit source on a node-2
		// drag, where the dragged node and its sphere center both move.
		moved := false
		if c, ok := msg.Centers[m.srcID]; ok {
			nodegeom.SetNodeWorld(&m.srcGeom, c)
			moved = true
		}
		if c, ok := msg.Centers[m.dstID]; ok {
			nodegeom.SetNodeWorld(&m.dstGeom, c)
			moved = true
		}
		if moved {
			m.recomputeGeometry()
		}
		return
	}
	// Plain "move" messages have no effect under the polar layout;
	// position updates arrive as "center" messages instead.
	_ = msg
}

// recomputeGeometry re-derives this edge's SEGMENT ONLY from its held endpoint
// geoms+handles and propagates it: publish onto the source Out, revise any in-flight
// bead's remaining travel (fraction-preserving, against this EdgeMover's own cached
// step count — see stepsIn's doc comment), and emit the new segment so the renderer
// redraws the wire. Shared by node-move and port-anchor handling.
//
// NO LENGTH IS COMPUTED HERE (docs/bead-model/bead-lattice.md "Ownership"): the step count is
// the SOURCE NODE's, published separately (package Wiring's chain_beads.go PublishSteps +
// this edge's SendSteps) — the old edgeArcPolar call and its lat := arc/PulseSpeedWuPerMs
// derivation are both gone, not replaced by an edge-side re-derivation of the same
// integer from a different (and potentially disagreeing) measurement.
func (m *EdgeMover) recomputeGeometry() {
	seg := nodegeom.EdgeSegment(m.srcGeom, m.dstGeom)

	// Publish the new segment onto the source Out's own buffered-1, latest-wins
	// channel, so the next placement (on the source node's own goroutine) reads it
	// by message — no data race with recomputeGeometry's write here.
	if m.out != nil {
		m.out.PublishSegment(seg.Start, seg.End)
	}
	// Re-derive an in-flight bead on this edge from the new segment + this EdgeMover's
	// own cached step count (no-op if none in flight); this runs on the SAME goroutine
	// that owns the dest wire's bead state (this is that wire's own goroutine — see
	// EdgeMover.Run).
	if m.dest != nil {
		m.dest.ReviseInFlightGeometry(m.clk.Tick(), m.steps, seg)
	}
	// Emit this edge's own segment so the renderer redraws the wire from Go's endpoints.
	// Geometry rides THIS EdgeMover's own dedicated stream (fully decentralized — it never
	// rides the VIEW stream's fallback bucket), since this goroutine is the sole owner of
	// this edge's geometry.
	// Dedicated per-edge stream (see streamOut's doc comment): write this edge's own combined frame immediately on a
	// geometry change, in addition to the tick-driven write in Run()'s loop. Carries
	// this EdgeMover's own row-resolved Geometry event (owner_events.go, package Wiring).
	m.writeStreamFrame(m.clk.Tick(), []wire.RowEvent{{
		Kind: T.KindGeometry, EdgeRow: m.edgeRow,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1,
	}})
}
