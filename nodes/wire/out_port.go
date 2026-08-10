// out_port.go — ONE JOB: the SENDING end of a port pair. The Out type, the two
// independent geometry publishers whose latest values it caches (outGeom,
// publishSteps/publishSegment and the Geom() drain that owns that cache), the
// placement it builds from that geometry, the actual hand-off onto the wire
// (placeDrivenNoWalker) and the send event it flushes, plus Out's own three
// constructors. What a CALLER of that hand-off is told about it is drive_item.go;
// the receiving end is in_port.go.

package wire

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// outGeom is an immutable snapshot of an Out's per-edge geometry: this edge's
// bead-step count plus its drawn straight-segment endpoints. It is assembled from
// TWO INDEPENDENT publishers, per docs/bead-model/bead-lattice.md "Ownership" — the source
// node owns the count, the edgeMover owns the segment — delivered to the ONE
// reading goroutine (the node's own Update goroutine, via Geom() below) over two
// SEPARATE buffered-1, latest-wins channels (geomSendSteps/geomSendSeg), never a
// shared field either publisher writes directly. This mirrors
// per-goroutine-clock.md's speedCh Delivery pattern (SendSpeedNonBlocking/
// ApplySpeedNonBlocking): each producer sends, the one consumer owns its own copy.
//
//   - Steps: this edge's own bead-step count (docs/bead-model/bead-lattice.md "The count"),
//     computed by the SOURCE NODE from its own stored LocalPolar to the target —
//     ONE integer, not a chord length divided by a speed. Published by
//     PublishSteps, called only from the source node's own goroutine (its
//     chainBeads pass, nodes/Wiring/chain_beads.go) — the SAME pass that lays the
//     chain out on this integer, so layout and the wire's own timing budget can
//     never read two different lengths. SendWire logs it so each bead animates for
//     its own edge's step count even when multiple edges fan into one destination
//     port.
//   - Start/End: this edge's straight-segment endpoints (source OUT-port world pos,
//     dest IN-port world pos) in the SAME 3-D frame the renderer draws. Published
//     by PublishSegment, called only from the edgeMover's own goroutine
//     (recomputeGeometry, nodes/Wiring/edge_mover.go) on a node move/port-anchor
//     change. They travel WITH each placed bead (beadPlacement) so the wire's
//     position stream evaluates P(t)=Start+t*(End-Start) on this edge, because the
//     shared dest wire never stores per-edge geometry.
//
// TWO PUBLISHERS ON ONE Out IS WHY THIS SPLIT EXISTS AT ALL: a single combined
// PublishGeom(steps, start, end) call would force one of the two goroutines to
// either supply a value it does not own (a race-inviting guess) or block on the
// other's latest value (a lock this repo forbids — memory/feedback_no_atomics_are_defects.md).
// Two independent non-blocking channels let each publisher write only what it
// owns, whenever it has a fresh value, with no coordination between them.
type outGeom struct {
	Steps      int
	Start, End Vec3
}

// Out is a typed output port.
type Out struct {
	// chan mode
	ch chan<- int
	// paced mode
	pw  *PacedWire
	ctx context.Context
	// shared
	node  string
	port  string
	trace *T.Trace
	// geomSendSteps/geomSendSeg are the two INDEPENDENT buffered-1, latest-wins
	// channels outGeom's doc comment describes: geomSendSteps fed by PublishSteps
	// (the source node's own goroutine, docs/bead-model/bead-lattice.md's step count owner),
	// geomSendSeg fed by PublishSegment (the edgeMover's own goroutine, the
	// segment owner). The load-time file geometry for BOTH is NOT sent through
	// either channel — sendCur is initialized to it directly in NewOutPaced. Both
	// are drained every cycle, via Geom() below, by whichever ONE goroutine
	// actually places beads on this Out — the node's own Update goroutine for most
	// kinds, or the dedicated DriveHeld goroutine for Pulse/HoldFlip (see
	// gatecommon/drive.go). Exactly one goroutine ever calls PlaceDrivenAt/
	// placement on a given Out, so exactly one goroutine ever drains either
	// channel. sendCur is the owned local cache for both, mutated only by that one
	// goroutine inside Geom(). A nil channel (chan-mode test Outs, which never
	// publish) is safe: a non-blocking receive on a nil channel simply never
	// fires.
	geomSendSteps chan int
	geomSendSeg   chan WireSegment
	sendCur       outGeom
	// EdgeLabel is the TS edge id for this output port's wire. Set by the loader
	// so the node's EmitGeometry closure can stream the authoritative curve via
	// tr.Geometry(EdgeLabel, Start..End) on startup.
	EdgeLabel string
	// Rule is the per-edge send policy applied by the source node after a
	// successful TrySend. Empty string defaults to consumeGated (see Gated).
	Rule SendRule
	// stream is this Out's owning node's shared event sink (injected by wireOutPort/
	// wireOutMultiPort as an eventSink adapter) — see In.stream's doc comment and the
	// eventSink seam. nil for a bare chan-mode Out built outside a kind's builder
	// (NewOutChanForTest, node unit tests).
	stream func() EventSink
	// portRow is this Out's own buffer PORT-ROW index (isInput=false); targetRow/
	// targetPortRow are the destination node/port's buffer rows (b.pw.Target/
	// TargetHandle — static after wiring). All resolved once at construction
	// (wireOutPort/wireOutMultiPort's doc comment). -1 when unresolved.
	portRow, targetRow, targetPortRow int32
}

// Geom returns the current per-edge geometry snapshot as seen by THIS Out's one
// sending goroutine: it non-blockingly drains any newer value off EACH of
// geomSendSteps/geomSendSeg into the goroutine-owned sendCur cache, then returns
// sendCur. Must only be called from the single goroutine that places beads on
// this Out (see the geomSendSteps/geomSendSeg doc comment above) — calling it
// from two goroutines would race sendCur. Returns the zero outGeom when nothing
// has ever been published (chan-mode test Outs never publish, and both channels
// are nil).
func (o *Out) Geom() outGeom {
	if o == nil {
		return outGeom{}
	}
	drainStepsNonBlocking(o.geomSendSteps, &o.sendCur.Steps)
	drainSegNonBlocking(o.geomSendSeg, &o.sendCur.Start, &o.sendCur.End)
	return o.sendCur
}

// publishSteps hands a fresh bead-step count to geomSendSteps' reader,
// latest-wins (dropping any undrained stale value first). Called only on the
// SOURCE NODE's own goroutine (chainBeads, nodes/Wiring/chain_beads.go) —
// docs/bead-model/bead-lattice.md "Ownership": the source node owns the count, never the
// edgeMover. The load-time file steps is set directly on sendCur in
// NewOutPaced, not published here. geomSendSteps is nil on a chan-mode Out
// (test-only, never published to); sendIntNonBlocking on a nil channel is a
// silent no-op, matching Geom()'s "never published" zero-value return.
func (o *Out) publishSteps(steps int) {
	sendIntNonBlocking(o.geomSendSteps, steps)
}

// PublishSteps is publishSteps' exported entry point for callers in another
// package (nodeMover.chainBeads in nodes/Wiring) that need to publish a live
// bead-step count without naming the unexported field it lands on.
func (o *Out) PublishSteps(steps int) {
	o.publishSteps(steps)
}

// publishSegment hands a fresh straight-segment (source OUT-port world pos, dest
// IN-port world pos) to geomSendSeg's reader, latest-wins. Called only on the
// EDGEMOVER's own goroutine (recomputeGeometry, nodes/Wiring/edge_mover.go) —
// docs/bead-model/bead-lattice.md "Ownership": the edgeMover owns the segment, never the
// step count. The load-time file segment is set directly on sendCur in
// NewOutPaced, not published here.
func (o *Out) publishSegment(start, end Vec3) {
	sendSegNonBlocking(o.geomSendSeg, WireSegment{Start: start, End: end})
}

// PublishSegment is publishSegment's exported entry point for callers in another
// package (edgeMover.recomputeGeometry in nodes/Wiring).
func (o *Out) PublishSegment(start, end Vec3) {
	o.publishSegment(start, end)
}

// placement builds the per-bead beadPlacement this Out hands to the wire: the
// per-edge in-flight time plus the position-stream context (segment endpoints
// + source identity). Centralized so TrySend and TryEmit stay in lockstep.
func (o *Out) placement() beadPlacement {
	return o.placementFrom(o.Geom())
}

// CurrentPlacement is placement()'s exported, read-only mirror for callers in another
// package (a concurrency-race test in nodes/Wiring) that need to read the currently
// published per-edge geometry fields without naming the unexported beadPlacement type
// and without placing a bead (unlike PlaceDrivenAt, this has no side effect).
func (o *Out) CurrentPlacement() (steps int, start, end Vec3) {
	bp := o.placement()
	return bp.Steps, bp.Start, bp.End
}

// placementFrom builds a beadPlacement from an already-loaded geometry snapshot, so
// a caller can use ONE consistent snapshot for both the placement and the SendWire
// trace (rather than two independent loads that could straddle a republish).
func (o *Out) placementFrom(g outGeom) beadPlacement {
	return beadPlacement{
		Steps: g.Steps,
		Start: g.Start,
		End:   g.End,
		Node:  o.node,
		Port:  o.port,
	}
}

// Paced reports whether this Out drives a paced wire. It is the paced-vs-chan MODE
// predicate: paced mode sleeps on the caller's own clock copy and StepOnces the wire;
// chan mode (unit tests) has no wire to advance and falls back to a wall-clock sleep.
//
// This used to say out loud what `out.Clock() != nil` said sideways — Out.Clock() is
// gone now (per-goroutine-clock.md API demolition item 1: port accessors go away, a
// goroutine gets its clock passed to it directly instead of reaching through a port),
// so Paced() is the only mode selector left. The condition is just "does this Out have
// a PacedWire": NewPacedWire (paced_wire.go) is the only construction site in the repo,
// so pw != nil is unambiguous.
func (o *Out) Paced() bool {
	return o != nil && o.pw != nil
}

// Gated reports whether the source node should wait for consumption after a
// successful send. Nil-safe; the zero value (empty Rule) is gated.
func (o *Out) Gated() bool {
	if o == nil {
		return true
	}
	return o.Rule != RuleFireAndForget
}

// placeDrivenNoWalker sends one bead placement onto the paced wire's in-channel
// (PacedWire.Send — non-blocking, never waits on the wire or the destination) and
// flushes this send as a RowEvent at placement time. tick is the CALLER's own
// clock reading (read once, at the emission site — see placeRequest's doc
// comment for why the wire itself no longer stamps this). Caller must have
// already checked o.pw != nil. Returns the wire's SendOutcome verbatim so the
// caller (PlaceDrivenAt) can distinguish a transient buffer-full from a
// genuinely terminal condition instead of collapsing both to one bool.
func (o *Out) placeDrivenNoWalker(v int, tick int64) SendOutcome {
	g := o.Geom()
	outcome := o.pw.Send(v, o.placementFrom(g), tick)
	if outcome != SendPlaced {
		return outcome
	}
	o.flushSendEvent(v, g.Steps)
	return SendPlaced
}

// flushSendEvent records this send as a row-resolved RowEvent on this Out's owning
// node's shared interior-stream frame (KindSend is fully decentralized, it never
// rides the VIEW stream's fallback bucket): this node's own Update goroutine
// (the SAME goroutine driving the send) is the sole owner, so it resolves its own
// NodeRow/PortRow/TargetRow/TargetPortRow at the call site (owner_events.go). No-op
// when stream is unset (bare chan-mode Out) or the node has no dedicated interior fd.
//
// flushSendEvent records this send as a row-resolved RowEvent on this Out's owning
// node's shared interior-stream frame. BeadSteps carries the bead-lattice length
// directly (docs/bead-model/bead-lattice.md "The count") — no chord/arc to derive it from
// anymore. SimLatencyMs stays a REPORTED diagnostic derived from steps
// (steps*DwellTicksPerBead*MsPerTick), not an independently measured value.
func (o *Out) flushSendEvent(value int, steps int) {
	if o.stream == nil {
		return
	}
	s := o.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]RowEvent{{
		Kind: T.KindSend, NodeRow: s.NodeRowOf(), PortRow: o.portRow,
		TargetRow: o.targetRow, TargetPortRow: o.targetPortRow, EdgeRow: -1,
		Value:        int32(value),
		BeadSteps:    float64(steps),
		SimLatencyMs: float64(steps) * lattice.DwellTicksPerBead * clock.MsPerTick,
	}})
}

// Wired reports whether this Out port is bound to a real edge (paced-wire
// mode). Returns false for a nil Out or a dead-end chan port (unwired).
// Nodes gate optional feedback sends on Wired() so unwired ports are never
// written.
func (o *Out) Wired() bool {
	if o == nil {
		return false
	}
	return o.pw != nil
}

// NewPacedOutNoGeom builds a paced Out with a zero wire segment. This is the
// supported entry point for tests that need to
// exercise the paced OUTPUT drive (PlaceDrivenAt → StepOnceAt) under a
// RealClock. Only bead timing is exercised; the zero segment means position
// traces carry no geometry. Production paced Outs are built by the loader/builders
// with real segments, not through this.
func NewPacedOutNoGeom(pw *PacedWire, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, edgeLabel string) *Out {
	return NewOutPaced(pw, ctx, node, port, tr, rule, steps, WireSegment{}, edgeLabel, nil, -1, -1, -1)
}

// NewOutChanForTest builds a chan-mode Out for tests outside the Wiring
// package. Chan mode's backing channel (ch) is unexported so other packages'
// tests (e.g. gatecommon's DriveHeld regression) cannot construct one
// directly; this is the supported entry point, mirroring NewPacedOutNoGeom's
// role for paced-mode tests.
func NewOutChanForTest(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return newOutChan(ch, node, port, tr)
}

// NewOutChanDeadEnd is portwiring's own production entry point for an unwired Out field
// (a port a spec's kind declares but this node's topology never wires — deadEndOut/
// deadEndOutSlice's send-only sink): identical construction to NewOutChanForTest, under a
// name that does not read as test-only, so check-fortest-has-no-production-caller.sh does
// not flag a real production call site as a test escape hatch.
func NewOutChanDeadEnd(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return newOutChan(ch, node, port, tr)
}

func newOutChan(ch chan<- int, node, port string, tr *T.Trace) *Out {
	return &Out{ch: ch, node: node, port: port, trace: tr}
}

// NewOutPaced builds a paced Out. stream is this Out's owning node's shared
// event-sink getter (nil for the many lean per-node tests, via
// NewPacedOutNoGeom, that build an Out directly without a loader). portRow is
// this Out's own buffer PORT-ROW index (isInput=false); targetRow/
// targetPortRow are the destination node/port's buffer rows. All three are
// -1 when unresolved — see wireOutPort/wireBroadcastPort's doc comments for
// how the loader resolves them.
func NewOutPaced(pw *PacedWire, ctx context.Context, node, port string, tr *T.Trace, rule SendRule, steps int, seg WireSegment, edgeLabel string, stream func() EventSink, portRow, targetRow, targetPortRow int32) *Out {
	if rule == "" {
		rule = RuleConsumeGated
	}
	// The initial geometry is the LOAD-TIME geometry the loader derived from the
	// topology file (steps/seg) — not a synthetic seed. Initialize
	// the reader's owned cache to it directly, before the reader goroutine starts
	// (happens-before), so the first placement reads valid file geometry with no
	// channel bootstrap. geomSend then carries ONLY live edgeMover updates (drags);
	// until the first one arrives, a non-blocking drain finds it empty and leaves the
	// cache at this file value.
	fileGeom := outGeom{Steps: steps, Start: seg.Start, End: seg.End}
	o := &Out{
		pw: pw, ctx: ctx, node: node, port: port, trace: tr, Rule: rule, EdgeLabel: edgeLabel,
		geomSendSteps: make(chan int, 1),
		geomSendSeg:   make(chan WireSegment, 1),
		sendCur:       fileGeom,
		stream:        stream, portRow: portRow, targetRow: targetRow, targetPortRow: targetPortRow,
	}
	return o
}
