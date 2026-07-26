// edge_mover.go — the per-edge mover actor type split out of node_mover.go. Pure move:
// no logic changes. edgeMover owns one edge's segment/arc + in-flight bead revision, and
// touches MoveDispatch only via injected fields — no back-reference to the dispatch
// registry, and no shared queue/lock between movers. See node_mover.go's doc comment for
// the shared "two dedicated channels, no shared inbox" design this actor participates in.

package Wiring

import (
	"context"
	"encoding/binary"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"io"

	T "github.com/dtauraso/wirefold/Trace"
)

// edgeMover owns one edge. It holds both endpoint geometries and recomputes its own
// segment/arc on an endpoint move (the edge label, which keys its channels below,
// encodes the two connected nodes).
type edgeMover struct {
	edgeID  string
	srcID   string
	dstID   string
	srcH    string
	dstH    string
	srcGeom nodeGeom
	dstGeom nodeGeom
	out     *wire.Out       // source Out for this edge (per-edge segment/arc/latency)
	dest    *wire.PacedWire // dest wire (in-flight revision + latency aggregate)
	// extIn is this edge's dedicated channel for EXTERNAL entries (gesture.go's
	// applyRingAnchor anchor mail-sort). srcIn/dstIn are this edge's two dedicated
	// channels FROM its two endpoint nodes' own goroutines — srcIn written only by
	// srcID's nodeMover, dstIn only by dstID's — the literal "two channels" the design
	// specifies, one per direction this edge can be told about a moved endpoint.
	// Nothing else ever writes on any of the three.
	extIn chan moveMsg
	srcIn chan moveMsg
	dstIn chan moveMsg
	tr    *T.Trace
	// clockSrc is the Clock this edgeMover's own goroutine (run) Copies from
	// EXACTLY ONCE at its own start, into clk below. Not read again afterward.
	clockSrc wire.Clock
	// clk is this edgeMover's OWN clock copy, set once by run() at goroutine
	// start. Only this goroutine (handle, called from run's loop) ever reads
	// it. Defaults to a fresh, real, live-ticking
	// RealClock (see newEdgeMover) so a test that calls handle() directly
	// without launching run() as a goroutine never dereferences a nil Clock —
	// per-goroutine-clock.md's API demolition deleted the old inert/zero-Tick
	// placeholder (item 3), so the only non-nil default left is a genuine
	// clock, not a fake stand-in.
	clk wire.Clock
	// speedCh delivers a speed change to THIS edgeMover's own clk copy
	// (per-goroutine-clock.md "Delivery"). Set once, at construction
	// (newMoveDispatch), from the loader's build-wide speed-sink accumulator;
	// nil in bare test construction, which is fine — a nil channel is never
	// selected in run()'s loop below.
	speedCh chan float64

	// --- dedicated per-edge stream (memory/feedback_no_single_writer_bridge.md) ---
	// streamOut, when non-nil, is THIS edge's OWN dedicated fd (see
	// MoveDispatch.SetEdgeStreams / Buffer/stream_fds.go's StreamKindEdge). Nil (the
	// default — no WIREFOLD_STREAM_FDS "edge" entry, e.g. headless tests) means
	// writeStreamFrame is a no-op: this edge's geometry+beads are simply never written
	// to a per-edge stream. Written ONLY by this edgeMover's own
	// goroutine (run/recomputeGeometry), mirroring every other single-
	// writer-per-goroutine field in this struct.
	streamOut io.Writer
	// edgeRow is this edge's stable buffer EDGE-ROW index (the seed order — see
	// MoveDispatch.SetEdgeStreams), carried on every Geometry event this edge's own
	// stream frame records (memory/feedback_no_single_writer_bridge.md). -1 until
	// SetEdgeStreams runs (bare test construction never sets it).
	edgeRow int32
	// portRowFor resolves (node, port, isInput) to its buffer PORT-ROW index — the
	// SAME resolution buildEdgeFrame's portRowLookup performs, injected here (rather
	// than importing Buffer) via MoveDispatch.SetEdgeStreams so this package stays
	// Buffer-independent, matching PortRowResolver/EdgeRowResolver's existing
	// interface-injection pattern.
	portRowFor func(node, port string, isInput bool) (int32, bool)
	// nodeRowFor resolves a node id to its buffer NODE-ROW index (mirroring the old
	// central accumulator's NodeRowFor), injected the same way as portRowFor. Used to
	// resolve the SOURCE node's row for this edge's own Geometry/Position/Arrive events.
	nodeRowFor func(id string) (int32, bool)
	// selected is this edge's OWN CURRENT click-selected bit — set only by this
	// edgeMover's own goroutine (handle's moveMsgKindSelect case, from a
	// MoveDispatch.sendEdgeSelect message), no shared map.
	selected uint8
	// buildFrame packs this edge's combined per-fd frame (edge fields + this wire's
	// live beads) using Buffer's own row-writer columns (Buffer.BuildEdgeStreamFrame),
	// injected so this package needs no Buffer import. events carries this goroutine's
	// OWN row-resolved events recorded since the last flush (memory/
	// feedback_no_single_writer_bridge.md).
	buildFrame func(tick uint32, srcPortRow, dstPortRow int32, selected uint8, label string, beadVal []int32, beadX, beadY, beadZ []float32, events []wire.RowEvent) []byte
}

func newEdgeMover(ep EdgeEndpoints, edgeID string, srcGeom, dstGeom nodeGeom, tr *T.Trace, clockSrc wire.Clock) *edgeMover {
	// clk defaults to a fresh RealClock (its own independent origin — fine here:
	// this default is only ever read by a test calling handle() directly, never by
	// production, where run() always overwrites it below with clockSrc.Copy() before
	// the goroutine does anything else) so a test that calls handle() directly
	// (without launching run() as a goroutine) never dereferences a nil Clock;
	// run() overwrites it with a real per-goroutine copy at start.
	return &edgeMover{
		edgeID:   edgeID,
		srcID:    ep.Source,
		dstID:    ep.Target,
		srcH:     ep.SourceHandle,
		dstH:     ep.TargetHandle,
		srcGeom:  srcGeom,
		dstGeom:  dstGeom,
		extIn:    make(chan moveMsg, 8),
		srcIn:    make(chan moveMsg, 8),
		dstIn:    make(chan moveMsg, 8),
		tr:       tr,
		clockSrc: clockSrc,
		clk:      wire.NewRealClock(),
		edgeRow:  -1,
	}
}

// handle applies one inbox message to this edge. For a move message it updates the
// matching endpoint geom, recomputes the edge's segment + arc, writes them onto the
// source Out, revises any in-flight bead, emits the new edge geometry, and updates
// the dest port's latency aggregate. A move that touches neither endpoint is ignored.
func (m *edgeMover) handle(msg moveMsg) {
	if msg.Kind == moveMsgKindSelect {
		if msg.Bool {
			m.selected = 1
		} else {
			m.selected = 0
		}
		return
	}
	if msg.Kind == moveMsgKindAnchor {
		// A port-anchor change recomputes this edge's segment/arc only if the changed
		// port is one of THIS edge's endpoints (matching node id, port name, direction).
		// Source endpoint is an OUTPUT (isInput==false); target endpoint is an INPUT.
		switch {
		case msg.NodeID == m.srcID && !msg.IsInput && msg.Port == m.srcH:
			if !setPortAnchorId(&m.srcGeom, msg.Port, false, msg.AnchorId) {
				return
			}
		case msg.NodeID == m.dstID && msg.IsInput && msg.Port == m.dstH:
			if !setPortAnchorId(&m.dstGeom, msg.Port, true, msg.AnchorId) {
				return
			}
		default:
			return
		}
		m.recomputeGeometry()
		return
	}
	if msg.Kind == moveMsgKindCenter {
		// Polar re-propagation: adopt the centrally-computed center on whichever
		// endpoint this message names, then recompute the edge.
		if msg.Center == nil {
			return
		}
		switch msg.NodeID {
		case m.srcID:
			setNodeWorld(&m.srcGeom, *msg.Center)
		case m.dstID:
			setNodeWorld(&m.dstGeom, *msg.Center)
		default:
			return
		}
		m.recomputeGeometry()
		return
	}
	if msg.Kind == moveMsgKindCenters {
		// Batched polar re-propagation: apply every moved endpoint this edge owns,
		// then recompute ONCE. An edge whose both endpoints moved in one frame would
		// otherwise recompute (and emit) twice — the duplicate-emit source on a node-2
		// drag, where the dragged node and its sphere center both move.
		moved := false
		if c, ok := msg.Centers[m.srcID]; ok {
			setNodeWorld(&m.srcGeom, c)
			moved = true
		}
		if c, ok := msg.Centers[m.dstID]; ok {
			setNodeWorld(&m.dstGeom, c)
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

// recomputeGeometry re-derives this edge's segment/arc/latency from its held endpoint
// geoms+handles and propagates them: write onto the source Out, revise any in-flight
// bead (fraction-preserving), update the dest port window aggregate, and emit the new
// segment so the renderer redraws the wire. Shared by node-move and port-anchor handling.
func (m *edgeMover) recomputeGeometry() {
	seg := edgeSegment(m.srcGeom, m.dstGeom, m.srcH, m.dstH)
	arc := edgeArcPolar(m.srcGeom, m.dstGeom, m.srcH, m.dstH)
	lat := arc / wire.PulseSpeedWuPerMs

	// Publish the new per-edge segment/arc/latency onto the source Out as an immutable
	// snapshot so the next placement (on the source node goroutine) reads the new
	// segment via an atomic load — no data race with recomputeGeometry's write here.
	if m.out != nil {
		m.out.PublishGeom(arc, lat, wire.Vec3{X: seg.Start.X, Y: seg.Start.Y, Z: seg.Start.Z}, wire.Vec3{X: seg.End.X, Y: seg.End.Y, Z: seg.End.Z})
	}
	// Re-derive an in-flight bead on this edge from the new arc + segment (no-op if
	// none in flight); this runs on the SAME goroutine that owns the dest wire's
	// bead state (this is that wire's own goroutine — see edgeMover.run).
	if m.dest != nil {
		m.dest.ReviseInFlightGeometry(m.clk.Tick(), arc, toWireSegment(seg))
	}
	// Emit this edge's own segment so the renderer redraws the wire from Go's endpoints.
	// Geometry rides THIS edgeMover's own dedicated stream (fully decentralized — it never
	// rides the VIEW stream's fallback bucket), since this goroutine is the sole owner of
	// this edge's geometry.
	// Dedicated per-edge stream (see streamOut's doc comment): write this edge's own combined frame immediately on a
	// geometry change, in addition to the tick-driven write in run()'s loop. Carries
	// this edgeMover's own row-resolved Geometry event (owner_events.go).
	m.writeStreamFrame(m.clk.Tick(), []wire.RowEvent{{
		Kind: T.KindGeometry, EdgeRow: m.edgeRow,
		NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1,
	}})
}

// writeStreamFrame packs and writes this edge's combined per-fd frame (edge fields +
// this wire's currently live in-flight beads) to its OWN dedicated fd (streamOut). No-op
// when streamOut is nil (the fallback — see its doc comment) or buildFrame was never
// injected (bare test construction). Called only by this edgeMover's own goroutine
// (recomputeGeometry and run's per-cycle loop), reading m.dest's
// live bead state via LiveBeadRows (same single-goroutine-ownership contract PacedWire's
// other methods rely on).
func (m *edgeMover) writeStreamFrame(tick int64, events []wire.RowEvent) {
	if m.streamOut == nil || m.buildFrame == nil {
		return
	}
	var srcRow, dstRow int32 = -1, -1
	if m.portRowFor != nil {
		if r, ok := m.portRowFor(m.srcID, m.srcH, false); ok {
			srcRow = r
		}
		if r, ok := m.portRowFor(m.dstID, m.dstH, true); ok {
			dstRow = r
		}
	}
	selected := m.selected
	var beadVal []int32
	var beadX, beadY, beadZ []float32
	if m.dest != nil {
		rows := m.dest.LiveBeadRows(tick)
		beadVal = make([]int32, len(rows))
		beadX = make([]float32, len(rows))
		beadY = make([]float32, len(rows))
		beadZ = make([]float32, len(rows))
		for i, r := range rows {
			beadVal[i] = int32(r.Val)
			beadX[i] = float32(r.X)
			beadY[i] = float32(r.Y)
			beadZ[i] = float32(r.Z)
		}
		// Drain this wire's own OWN-goroutine-recorded Position/Arrive events, resolved
		// to rows here (srcRow/nodeRowFor — the SAME resolvers this frame's own edge
		// columns above just used), and fold them in alongside any caller-supplied
		// events (e.g. a Geometry event from recomputeGeometry).
		nodeRow := int32(-1)
		if m.nodeRowFor != nil {
			if r, ok := m.nodeRowFor(m.srcID); ok {
				nodeRow = r
			}
		}
		for _, pe := range m.dest.DrainPendingEvents() {
			events = append(events, wire.RowEvent{
				Kind: pe.Kind, NodeRow: nodeRow, PortRow: srcRow,
				TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Value: int32(pe.Value), Bead: pe.Gen,
				X: pe.X, Y: pe.Y, Z: pe.Z, F: pe.T,
			})
		}
		// This wire's own "wire-send-buffer-full" breadcrumbs, buffered on
		// breadcrumbCh from the source node's goroutine (PacedWire.Send) and
		// resolved to rows here, on this edgeMover's own goroutine — mirrors
		// DrainPendingEvents just above.
		for _, ev := range m.dest.DrainBreadcrumbEvents() {
			ev.NodeRow = nodeRow
			ev.PortRow = srcRow
			ev.TargetRow = dstRow
			events = append(events, ev)
		}
	}
	frame := m.buildFrame(uint32(tick), srcRow, dstRow, selected, m.edgeID, beadVal, beadX, beadY, beadZ, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}

// run is the edge's per-goroutine loop. It IS the wire's own goroutine
// (MODEL.md "The network" — PacedWire is an active goroutine, and it is this
// same per-edge goroutine that already existed to revise in-flight geometry,
// not an additional one): every cycle it drains any pending move/speed
// messages without blocking, then drives its dest wire's ONE cycle of bead
// ownership (DriveOneCycle — placement drain, position-step, delivery
// handoff), then paces to the next cycle on its OWN clock copy. This is what
// lets ReviseInFlightGeometry (called from handle, below, on this SAME
// goroutine) touch pw.inflight: there is exactly one goroutine
// on either side of that call.
func (m *edgeMover) run(ctx context.Context) {
	// Copy taken ONCE at this goroutine's start (run IS the goroutine). If no clockSrc was
	// given (bare test construction), keep the inert placeholder newEdgeMover
	// seeded m.clk with.
	if m.clockSrc != nil {
		m.clk = m.clockSrc.Copy()
	}
	// ONE-TIME startup geometry emit, on THIS edge's own mover goroutine — this is now
	// the sole per-owner source of an edge's initial geometry event (replacing the old
	// source-node-Update-loop startup emit builders.go's EmitGeometry closure used to
	// make for each of its outgoing edges; that closure no longer calls tr.Geometry —
	// see its doc comment). m.tr is non-nil in production; bare test construction with
	// a nil tr just skips this, matching recomputeGeometry's own nil-guard elsewhere.
	if m.tr != nil {
		m.recomputeGeometry()
	}
	for {
		// Drain extIn/srcIn/dstIn/speedCh without blocking, so a cycle always reaches
		// the wire-drive step below even with nothing queued. Three dedicated channels,
		// not one shared inbox: extIn (external gesture entries), srcIn (this edge's
		// source node's own goroutine), dstIn (this edge's target node's own goroutine).
	drain:
		for {
			select {
			case <-ctx.Done():
				return
			case sp := <-m.speedCh:
				// Delivery (per-goroutine-clock.md): apply directly to this
				// goroutine's own clk copy — nothing else reaches it.
				if rc, ok := m.clk.(*wire.RealClock); ok {
					rc.SetSpeed(sp)
				}
			case msg := <-m.extIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
			case msg := <-m.srcIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
			case msg := <-m.dstIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
			default:
				break drain
			}
		}
		if m.dest != nil {
			m.dest.DriveOneCycle(ctx, m.clk.Tick())
			// Beads on this wire may have moved even with no geometry change this
			// cycle — write this edge's dedicated stream frame every cycle (no-op
			// when streamOut is nil, the fallback path).
			m.writeStreamFrame(m.clk.Tick(), nil)
		}
		if err := m.clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
