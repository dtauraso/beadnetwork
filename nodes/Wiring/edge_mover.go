// edge_mover.go — the per-edge mover actor type: its held state, construction, and its
// inbox message handler. Pure move: no logic changes. edgeMover owns one edge's
// segment/arc + in-flight bead revision, and touches MoveDispatch only via injected
// fields — no back-reference to the dispatch registry, and no shared queue/lock between
// movers. See node_mover.go's doc comment for the shared "two dedicated channels, no
// shared inbox" design this actor participates in.
//
// This edge's own on-disk file (nodes/<source>/edges/<label>.json) and its persistence
// helpers live in edge_file.go, not here — edge_mover.go is the runtime actor, edge_file.go
// is the shape on disk.
//
// This edge's own per-fd stream frame write lives in edge_mover_stream.go; its per-goroutine
// run loop lives in edge_mover_run.go.

package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// edgeMover owns one edge. It holds both endpoint geometries and recomputes its own
// SEGMENT (never a length — docs/bead-model/bead-lattice.md "Ownership": the step count is the
// SOURCE NODE's) on an endpoint move (the edge label, which keys its channels below,
// encodes the two connected nodes).
type edgeMover struct {
	edgeID  string
	srcID   string
	dstID   string
	srcH    string
	dstH    string
	srcGeom nodegeom.NodeGeom
	dstGeom nodegeom.NodeGeom
	out     *wire.Out       // source Out for this edge (this edgeMover publishes the SEGMENT onto it)
	dest    *wire.PacedWire // dest wire (in-flight bead revision)
	// extIn is this edge's dedicated channel for EXTERNAL entries (gesture.go's
	// applyRingAnchor anchor mail-sort). srcIn/dstIn are this edge's two dedicated
	// channels FROM its two endpoint nodes' own goroutines — srcIn written only by
	// srcID's nodeMover, dstIn only by dstID's — the literal "two channels" the design
	// specifies, one per direction this edge can be told about a moved endpoint.
	// Nothing else ever writes on any of the three.
	extIn chan movemsg.Msg
	srcIn chan movemsg.Msg
	dstIn chan movemsg.Msg
	// stepsIn is a buffered-1, latest-wins channel carrying this edge's freshly
	// computed bead-step count (docs/bead-model/bead-lattice.md "The count") FROM the SOURCE
	// node's own goroutine (nodeMover.chainBeads, node_mover.go/chain_beads.go —
	// the count owner) TO this edgeMover's own goroutine. Needed because
	// recomputeGeometry (below) calls ReviseInFlightGeometry, which needs the
	// CURRENT step count to re-derive an in-flight bead's remaining travel — but
	// this edgeMover cannot read the source Out's Geom() itself (that cache,
	// o.sendCur, is owned exclusively by the ONE goroutine that places beads on
	// it, per out_port.go's Geom() doc comment; a second reader would race it). A
	// dedicated delivery channel, drained non-blockingly every cycle into `steps`
	// below, is the same "producer sends, one consumer owns its copy" shape
	// speedCh already uses (per-goroutine-clock.md "Delivery") — not a lock, not
	// a shared field. Bound once per edge at construction (mover_registry.go's
	// bind, paired with nodeMover.outStepsIn); nil in bare test construction.
	stepsIn chan int
	// steps is this edgeMover's OWN cached copy of the latest step count drained
	// from stepsIn (run's per-cycle drain, edge_mover_run.go) — owned and mutated
	// exclusively by this edgeMover's own goroutine. Read by recomputeGeometry to
	// pass to ReviseInFlightGeometry. Zero until the source node's first
	// chainBeads pass publishes a value (this edge then has no in-flight beads to
	// revise yet, so a zero read here is harmless — ReviseInFlightGeometry no-ops
	// when len(inflight)==0).
	steps int
	tr    *T.Trace
	// clockSrc is the Clock this edgeMover's own goroutine (run, edge_mover_run.go)
	// Copies from EXACTLY ONCE at its own start, into clk below. Not read again
	// afterward.
	clockSrc clock.Clock
	// clk is this edgeMover's OWN clock copy, set once by run() at goroutine
	// start. Only this goroutine (handle, called from run's loop) ever reads
	// it. Defaults to a fresh, real, live-ticking
	// RealClock (see newEdgeMover) so a test that calls handle() directly
	// without launching run() as a goroutine never dereferences a nil Clock —
	// per-goroutine-clock.md's API demolition deleted the old inert/zero-Tick
	// placeholder (item 3), so the only non-nil default left is a genuine
	// clock, not a fake stand-in.
	clk clock.Clock
	// speedCh delivers a speed change to THIS edgeMover's own clk copy
	// (per-goroutine-clock.md "Delivery"). Set once, at construction
	// (newMoveDispatch), from the loader's build-wide speed-sink accumulator;
	// nil in bare test construction, which is fine — a nil channel is never
	// selected in run()'s loop below.
	speedCh chan float64

	// --- dedicated per-edge stream (memory/feedback_no_single_writer_bridge.md) ---
	// streamOut, when Ok(), is THIS edge's OWN dedicated fd (see
	// MoveDispatch.SetEdgeStreams / Buffer/streamframe/stream_fds.go's StreamKindEdge). A dead
	// claimedStream (the default — no WIREFOLD_STREAM_FDS "edge" entry, e.g. headless
	// tests, OR a rejected second claim — see stream_claim.go) means writeStreamFrame is
	// a no-op: this edge's geometry+beads are simply never written to a per-edge stream.
	// claimedStream's unexported field + unexported constructor (newClaimedStream,
	// called only from setEdgeStreams) make a SECOND claim on this edge's fd
	// structurally rejected, not just written ONLY by this edgeMover's own goroutine
	// (run/recomputeGeometry) by convention.
	streamOut claimedStream
	// edgeRow is this edge's stable buffer EDGE-ROW index (the seed order — see
	// MoveDispatch.SetEdgeStreams), carried on every Geometry event this edge's own
	// stream frame records (memory/feedback_no_single_writer_bridge.md). -1 until
	// SetEdgeStreams runs (bare test construction never sets it).
	edgeRow int32
	// nodeRowFor resolves a node id to its buffer NODE-ROW index (mirroring the old
	// central accumulator's NodeRowFor), injected via MoveDispatch.SetEdgeStreams. Used to
	// resolve the SOURCE node's row for this edge's own Geometry/Position/Arrive events.
	nodeRowFor func(id string) (int32, bool)
	// selected is this edge's OWN CURRENT click-selected bit — set only by this
	// edgeMover's own goroutine (handle's movemsg.KindSelect case, from a
	// MoveDispatch.sendEdgeSelect message), no shared map.
	selected uint8
	// buildFrame packs this edge's combined per-fd frame (edge fields + this wire's
	// live beads) using Buffer's own row-writer columns (Buffer.BuildEdgeStreamFrame),
	// injected so this package needs no Buffer import. events carries this goroutine's
	// OWN row-resolved events recorded since the last flush (memory/
	// feedback_no_single_writer_bridge.md).
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte
}

func newEdgeMover(ep inputcodec.EdgeEndpoints, edgeID string, srcGeom, dstGeom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *edgeMover {
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
		extIn:    make(chan movemsg.Msg, moverInboxDepth),
		srcIn:    make(chan movemsg.Msg, moverInboxDepth),
		dstIn:    make(chan movemsg.Msg, moverInboxDepth),
		stepsIn:  make(chan int, 1),
		tr:       tr,
		clockSrc: clockSrc,
		clk:      clock.NewRealClock(),
		edgeRow:  -1,
	}
}

// handle applies one inbox message to this edge. For a move message it updates the
// matching endpoint geom, recomputes the edge's segment + arc, writes them onto the
// source Out, revises any in-flight bead, emits the new edge geometry, and updates
// the dest port's latency aggregate. A move that touches neither endpoint is ignored.
func (m *edgeMover) handle(msg movemsg.Msg) {
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
// bead's remaining travel (fraction-preserving, against this edgeMover's own cached
// step count — see stepsIn's doc comment), and emit the new segment so the renderer
// redraws the wire. Shared by node-move and port-anchor handling.
//
// NO LENGTH IS COMPUTED HERE (docs/bead-model/bead-lattice.md "Ownership"): the step count is
// the SOURCE NODE's, published separately (chain_beads.go's PublishSteps + this
// edge's stepsIn) — the old edgeArcPolar call and its lat := arc/PulseSpeedWuPerMs
// derivation are both gone, not replaced by an edge-side re-derivation of the same
// integer from a different (and potentially disagreeing) measurement.
func (m *edgeMover) recomputeGeometry() {
	seg := nodegeom.EdgeSegment(m.srcGeom, m.dstGeom)

	// Publish the new segment onto the source Out's own buffered-1, latest-wins
	// channel, so the next placement (on the source node's own goroutine) reads it
	// by message — no data race with recomputeGeometry's write here.
	if m.out != nil {
		m.out.PublishSegment(seg.Start, seg.End)
	}
	// Re-derive an in-flight bead on this edge from the new segment + this edgeMover's
	// own cached step count (no-op if none in flight); this runs on the SAME goroutine
	// that owns the dest wire's bead state (this is that wire's own goroutine — see
	// edgeMover.run).
	if m.dest != nil {
		m.dest.ReviseInFlightGeometry(m.clk.Tick(), m.steps, seg)
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
