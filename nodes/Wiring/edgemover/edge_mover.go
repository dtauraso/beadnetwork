// edge_mover.go — the per-edge mover actor type: its held state, construction, and its
// inbox message handler. Pure move out of package Wiring (docs/planning/movedispatch-decomposition.md):
// no logic changes. EdgeMover owns one edge's segment/arc + in-flight bead revision, and
// touches package Wiring only via injected values passed at construction/wiring time or
// bound func values handed back to it — no back-reference to the dispatch registry, and no
// shared queue/lock between movers. See node_mover.go's doc comment (package Wiring) for
// the shared "two dedicated channels, no shared inbox" design this actor participates in.
//
// This edge's own on-disk file (nodes/<source>/edges/<label>.json) and its persistence
// helpers live in edge_file.go (package Wiring), not here — this is the runtime actor,
// edge_file.go is the shape on disk.
//
// This edge's own per-fd stream frame write lives in edge_mover_stream.go; its per-goroutine
// run loop lives in edge_mover_run.go.
package edgemover

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// EdgeMover owns one edge. It holds both endpoint geometries and recomputes its own
// SEGMENT (never a length — docs/bead-model/bead-lattice.md "Ownership": the step count is the
// SOURCE NODE's) on an endpoint move (the edge label, which keys its channels below,
// encodes the two connected nodes).
type EdgeMover struct {
	edgeID string
	srcID  string
	dstID  string
	srcH   string
	dstH   string

	srcGeom nodegeom.NodeGeom
	dstGeom nodegeom.NodeGeom
	out     *wire.Out       // source Out for this edge (this EdgeMover publishes the SEGMENT onto it)
	dest    *wire.PacedWire // dest wire (in-flight bead revision)
	// extIn is this edge's dedicated channel for EXTERNAL entries (package Wiring's
	// gesture.go applyRingAnchor anchor mail-sort, and edge click-select — see Select
	// below). srcIn/dstIn are this edge's two dedicated channels FROM its two endpoint
	// nodes' own goroutines — srcIn written only by srcID's nodeMover, dstIn only by
	// dstID's — the literal "two channels" the design specifies, one per direction this
	// edge can be told about a moved endpoint. Nothing else ever writes on any of the
	// three; package Wiring never reaches these channels directly — it goes through
	// Select/TrySendFromSrc/TrySendFromDst below, which are the bound-func-value seam
	// this package hands back to Wiring (mirrors Wiring's own
	// ng.msg.sendMove = md.mr.enqueueFor(ng) pattern in the other direction).
	extIn chan movemsg.Msg
	srcIn chan movemsg.Msg
	dstIn chan movemsg.Msg
	// stepsIn is a buffered-1, latest-wins channel carrying this edge's freshly
	// computed bead-step count (docs/bead-model/bead-lattice.md "The count") FROM the SOURCE
	// node's own goroutine (nodeMover.chainBeads, package Wiring — the count owner) TO
	// this EdgeMover's own goroutine. Needed because recomputeGeometry (below) calls
	// ReviseInFlightGeometry, which needs the CURRENT step count to re-derive an
	// in-flight bead's remaining travel — but this EdgeMover cannot read the source
	// Out's Geom() itself (that cache, o.sendCur, is owned exclusively by the ONE
	// goroutine that places beads on it, per out_port.go's Geom() doc comment; a second
	// reader would race it). A dedicated delivery channel, drained non-blockingly every
	// cycle into `steps` below, is the same "producer sends, one consumer owns its
	// copy" shape speedCh already uses (per-goroutine-clock.md "Delivery") — not a
	// lock, not a shared field. Fed only through SendSteps below, never a raw channel
	// handed to package Wiring. Bound once per edge at wiring time (SetSpeedCh/
	// package Wiring's mover_registry.go bind, paired with nodeMover.outStepsIn); nil
	// in bare test construction.
	stepsIn chan int
	// steps is this EdgeMover's OWN cached copy of the latest step count drained
	// from stepsIn (run's per-cycle drain, edge_mover_run.go) — owned and mutated
	// exclusively by this EdgeMover's own goroutine. Read by recomputeGeometry to
	// pass to ReviseInFlightGeometry. Zero until the source node's first
	// chainBeads pass publishes a value (this edge then has no in-flight beads to
	// revise yet, so a zero read here is harmless — ReviseInFlightGeometry no-ops
	// when len(inflight)==0).
	steps int
	tr    *T.Trace
	// clockSrc is the Clock this EdgeMover's own goroutine (Run, edge_mover_run.go)
	// Copies from EXACTLY ONCE at its own start, into clk below. Not read again
	// afterward.
	clockSrc clock.Clock
	// clk is this EdgeMover's OWN clock copy, set once by Run() at goroutine
	// start. Only this goroutine (handle, called from Run's loop) ever reads
	// it. Defaults to a fresh, real, live-ticking RealClock (see New) so a test that
	// calls handle() directly without launching Run() as a goroutine never
	// dereferences a nil Clock.
	clk clock.Clock
	// speedCh delivers a speed change to THIS EdgeMover's own clk copy
	// (per-goroutine-clock.md "Delivery"). Set once, at wiring time (SetSpeedCh), from
	// the loader's build-wide speed-sink accumulator; nil in bare test construction,
	// which is fine — a nil channel is never selected in Run()'s loop below.
	speedCh chan float64

	// --- dedicated per-edge stream (memory/feedback_no_single_writer_bridge.md) ---
	// streamOut, when Ok(), is THIS edge's OWN dedicated fd (see SetStream /
	// Buffer/streamframe/stream_fds.go's StreamKindEdge). A dead StreamHandle (the
	// default — no WIREFOLD_STREAM_FDS "edge" entry, e.g. headless tests, OR a rejected
	// second claim — see stream_claim.go) means writeStreamFrame is a no-op: this
	// edge's geometry+beads are simply never written to a per-edge stream.
	// StreamHandle's unexported field + Claim's collision check make a SECOND claim on
	// this edge's fd structurally rejected, not just written ONLY by this EdgeMover's
	// own goroutine (Run/recomputeGeometry) by convention.
	streamOut StreamHandle
	// edgeRow is this edge's stable buffer EDGE-ROW index (the seed order — see
	// SetStream), carried on every Geometry event this edge's own stream frame records
	// (memory/feedback_no_single_writer_bridge.md). -1 until SetStream runs (bare test
	// construction never sets it).
	edgeRow int32
	// nodeRowFor resolves a node id to its buffer NODE-ROW index (mirroring the old
	// central accumulator's NodeRowFor), injected via SetStream. Used to resolve the
	// SOURCE node's row for this edge's own Geometry/Position/Arrive events.
	nodeRowFor func(id string) (int32, bool)
	// selected is this edge's OWN CURRENT click-selected bit — set only by this
	// EdgeMover's own goroutine (handle's movemsg.KindSelect case, from a Select call),
	// no shared map.
	selected uint8
	// buildFrame packs this edge's combined per-fd frame (edge fields + this wire's
	// live beads) using Buffer's own row-writer columns (Buffer.BuildEdgeStreamFrame),
	// injected so this package needs no Buffer import. events carries this goroutine's
	// OWN row-resolved events recorded since the last flush (memory/
	// feedback_no_single_writer_bridge.md).
	buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte
}

// New constructs an EdgeMover for one edge between srcID/dstID (connected via srcHandle/
// dstHandle), seeded with each endpoint's load-time geometry. clk defaults to a fresh
// RealClock (its own independent origin — fine here: this default is only ever read by a
// test calling handle() directly, never by production, where Run() always overwrites it
// below with clockSrc.Copy() before the goroutine does anything else) so a test that calls
// handle() directly (without launching Run() as a goroutine) never dereferences a nil
// Clock; Run() overwrites it with a real per-goroutine copy at start.
func New(edgeID, srcID, dstID, srcHandle, dstHandle string, srcGeom, dstGeom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *EdgeMover {
	return &EdgeMover{
		edgeID:   edgeID,
		srcID:    srcID,
		dstID:    dstID,
		srcH:     srcHandle,
		dstH:     dstHandle,
		srcGeom:  srcGeom,
		dstGeom:  dstGeom,
		extIn:    make(chan movemsg.Msg, InboxDepth),
		srcIn:    make(chan movemsg.Msg, InboxDepth),
		dstIn:    make(chan movemsg.Msg, InboxDepth),
		stepsIn:  make(chan int, 1),
		tr:       tr,
		clockSrc: clockSrc,
		clk:      clock.NewRealClock(),
		edgeRow:  -1,
	}
}

// InboxDepth is this actor's own inbox capacity, mirroring Wiring's moverInboxDepth (the
// two packages describe the same "a few frames of drag messages" queue; see mover_registry.go's
// doc comment there for the full reasoning — not re-derived here).
const InboxDepth = 8

// SrcID/DstID/SrcHandle/DstHandle are this edge's two endpoint identities, fixed at
// construction and read-only thereafter — package Wiring's resolveDest reads SrcID/DstID
// on every retry to decide which of TrySendFromSrc/TrySendFromDst applies, and bind reads
// SrcHandle/DstHandle once, at wiring time, to look up this edge's Out/dest wire.
func (m *EdgeMover) SrcID() string     { return m.srcID }
func (m *EdgeMover) DstID() string     { return m.dstID }
func (m *EdgeMover) SrcHandle() string { return m.srcH }
func (m *EdgeMover) DstHandle() string { return m.dstH }

// SetOut binds this edge's source Out — the SEGMENT this EdgeMover publishes onto, and
// the step-count publish target chainBeads (package Wiring) already reaches via the
// source node's own outWireOuts. Called once, at wiring time (package Wiring's
// moverRegistry.bind), before any goroutine is launched.
func (m *EdgeMover) SetOut(out *wire.Out) { m.out = out }

// SetDest binds this edge's destination wire (the in-flight bead revision target). Called
// once, at wiring time (moverRegistry.bind), before any goroutine is launched.
func (m *EdgeMover) SetDest(dest *wire.PacedWire) { m.dest = dest }

// Dest returns this edge's destination wire, or nil if none is bound yet — read-only
// access for package Wiring's setEdgeStreams, which flips SetStreamsActive(true) on it
// once a stream is wired (see stream_wiring.go's setEdgeStreams).
func (m *EdgeMover) Dest() *wire.PacedWire { return m.dest }

// SetSpeedCh binds this edge's speed-change delivery channel (per-goroutine-clock.md
// "Delivery"), from the loader's build-wide speed-sink accumulator. Called once, at
// construction time, before any goroutine is launched; nil in bare test construction.
func (m *EdgeMover) SetSpeedCh(ch chan float64) { m.speedCh = ch }

// SetStream wires this edge to its own dedicated per-fd stream — see streamOut's doc
// comment. Called once, at wiring time (package Wiring's setEdgeStreams), before this
// edge's own goroutine is launched.
func (m *EdgeMover) SetStream(h StreamHandle, edgeRow int32, nodeRowFor func(id string) (int32, bool), buildFrame func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte) {
	m.streamOut = h
	m.edgeRow = edgeRow
	m.nodeRowFor = nodeRowFor
	m.buildFrame = buildFrame
}

// Select routes a select/deselect message onto this edge's own dedicated extIn channel —
// the edgeMover sets its OWN selected field on its own goroutine, no shared map. A
// blocking send with a ctx-cancel escape hatch, same reasoning as package Wiring's
// sendMove: this is a bare external-entry send with no owning goroutine to thread a ctx
// from. Moved from package Wiring's sendEdgeSelect (move_dispatch_api.go) as part of the
// edgeMover package move — extIn stays unexported, so package Wiring calls this method
// instead of writing the channel directly.
func (m *EdgeMover) Select(ctx context.Context, on bool) {
	msg := movemsg.Msg{Kind: movemsg.KindSelect, Bool: on}
	if ctx == nil {
		m.extIn <- msg
		return
	}
	select {
	case m.extIn <- msg:
	case <-ctx.Done():
	}
}

// TrySendFromSrc/TrySendFromDst attempt a non-blocking send on this edge's srcIn/dstIn —
// the two dedicated channels FROM this edge's source/target node's own goroutine,
// reporting success. Package Wiring's nodeGeometry.resolveDest returns one of these as the
// bound func value a node's own flushPending calls, in place of ever reaching srcIn/dstIn
// directly (the same "hand back a bound func value" shape as Select above; mirrors
// Wiring's own ng.msg.sendMove = md.mr.enqueueFor(ng) pattern in the other direction).
func (m *EdgeMover) TrySendFromSrc(msg movemsg.Msg) bool {
	select {
	case m.srcIn <- msg:
		return true
	default:
		return false
	}
}

func (m *EdgeMover) TrySendFromDst(msg movemsg.Msg) bool {
	select {
	case m.dstIn <- msg:
		return true
	default:
		return false
	}
}

// SendSteps delivers steps onto this edge's stepsIn, non-blocking and latest-wins (an
// undrained stale value is replaced, never queued) — the source node's own chainBeads
// pass (package Wiring) calls this instead of ever reaching stepsIn directly. This is the
// same "producer sends, one consumer owns its copy" shape speedCh already uses
// (per-goroutine-clock.md "Delivery"), and replaces the old package-level
// nodes/Wiring/stepdeliver.SendStepsNonBlocking helper — that package existed solely for
// this one call site and is deleted as part of this actor's move.
func (m *EdgeMover) SendSteps(steps int) {
	if m.stepsIn == nil {
		return
	}
	select {
	case m.stepsIn <- steps:
	default:
		select {
		case <-m.stepsIn:
		default:
		}
		select {
		case m.stepsIn <- steps:
		default:
		}
	}
}

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
