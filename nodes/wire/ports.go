// ports.go — typed port wrappers that bake tracing into send/recv.
//
// Nodes hold In / Out / Broadcast fields instead of raw channels.
// TryRecv / TrySend emit the corresponding trace event on success,
// so a node cannot forget to trace, nor can it mis-type a port name
// string — the port name lives in the wrapper and is set by
// a kind's own builder, which passes the port name explicitly.
//
// Two backing modes:
//   - chan mode (NewIn / NewOut): used by node unit tests. Non-blocking
//     select on the raw channel — original TryRecv/TrySend semantics.
//   - PacedWire mode (NewInPaced / NewOutPaced): used by the loader.
//     TrySend blocks until the paced wire delivers the value (always
//     returns true); TryRecv blocks until a value arrives. Ctx cancel
//     causes both to return the zero-value / false.

package wire

import (
	"context"
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
)

// eventSink is the seam between a port (transport — moving values between nodes) and the
// buffer reporting it announces recv/send/breadcrumb events to. A port holds a
// func() eventSink and calls writeEvents/nodeRowOf on the result; it never names the
// concrete *interiorStream. interiorStream implements this, but the port cannot see that —
// which is what lets the transport primitive be lifted out of the reporting machinery.
// The injected getter returns a TRUE nil interface (not a typed-nil) when the node has no
// interior stream, so the callers' `if s == nil` guards keep working (see asEventSinkGetter).
type EventSink interface {
	WriteEvents(events []RowEvent)
	NodeRowOf() int32
}

// In is a typed input port.
type In struct {
	// chan mode
	ch <-chan int
	// paced mode
	pw  *PacedWire
	ctx context.Context
	// shared
	node  string
	port  string
	trace *T.Trace
	// stream is this In's owning node's shared event sink (the interior-stream getter,
	// injected by wireInPort as an eventSink adapter over newInteriorStreamGetter) —
	// lazily resolves to the SAME sink every closure/port on this node shares. Recv flushes
	// its own row-resolved RowEvent onto it (owner_events.go). The port announces events
	// through the eventSink seam and never names the concrete interior-stream type. nil for
	// a bare chan-mode In built outside a kind's builder (e.g. gatecommon test helpers) — the
	// nil check below skips the flush in that case.
	stream func() EventSink
	// portRow is this In's own buffer PORT-ROW index (isInput=true), resolved once at
	// construction (wireInPort) from pb.md's row table — see wireInPort's doc comment.
	// -1 when unresolved (no md, or an unwired dead-end port).
	portRow int32
}

// PollRecv is the non-blocking receive used by windowed nodes. In paced mode it
// calls pw.PollRecv (returns immediately with ok=false when no value is present,
// without parking) and, on success, CONSUMES the value on read (pops the front
// delivered bead) while emitting the same trace events as TryRecv. There is no
// separate Done step — the read itself consumes. In chan mode it does a
// non-blocking select, identical to TryRecv's default branch.
//
// Each successful receive ALSO flushes a KindRecv RowEvent onto this node's own
// interior-stream frame (i.stream — KindRecv is fully decentralized, it never rides
// the VIEW stream's fallback bucket): this node's own Update goroutine
// (the SAME goroutine calling PollRecv) is the sole owner of when it receives, so it
// resolves its own NodeRow/PortRow at the call site (owner_events.go) rather than
// routing through a shared accumulator.
func (i *In) PollRecv() (int, bool) {
	if i == nil {
		return 0, false
	}
	if i.pw != nil {
		n, ok := i.pw.Recv()
		if !ok {
			return 0, false
		}
		i.flushRecvEvent(n)
		return n, true
	}
	if i.ch == nil {
		return 0, false
	}
	select {
	case v := <-i.ch:
		i.flushRecvEvent(v)
		return v, true
	default:
		return 0, false
	}
}

// flushRecvEvent records this receive as a row-resolved RowEvent on this In's owning
// node's shared interior-stream frame. No-op when stream is unset (bare chan-mode In
// built outside a kind's builder) or the node has no dedicated interior fd.
func (i *In) flushRecvEvent(value int) {
	if i.stream == nil {
		return
	}
	s := i.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]RowEvent{{
		Kind: T.KindRecv, NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Value: int32(value),
	}})
}

// NewInChan builds a dead-end chan-mode In (no PacedWire) for a port with no
// paced binding — the unwired fallback the loader's builder machinery (a
// separate package from wire) uses. stream is this In's owning node's shared
// event-sink getter, set at construction since the field is unexported and
// the loader/builders live in a different package (nil for bare chan-mode
// Ins built outside a kind's builder, e.g. gatecommon test helpers).
func NewInChan(ch <-chan int, node, port string, tr *T.Trace, stream func() EventSink) *In {
	return &In{ch: ch, node: node, port: port, trace: tr, portRow: -1, stream: stream}
}

// Wired reports whether this In port is bound to a real edge (paced-wire
// mode). Returns false for a nil In or a dead-end chan port (unwired).
// Nodes gate optional feedback receives on Wired() so unwired ports are never
// read.
func (i *In) Wired() bool {
	if i == nil {
		return false
	}
	return i.pw != nil
}

// Breadcrumb emits a trace breadcrumb on the input port's wire identity (target
// node + handle). Used by windowed nodes for the window_clear breadcrumb.
func (i *In) Breadcrumb(event, detail string) {
	if i == nil || i.trace == nil {
		return
	}
	node, port := i.node, i.port
	if i.pw != nil {
		node, port = i.pw.Target, i.pw.TargetHandle
	}
	i.trace.Breadcrumb(event, node, port, detail)
	// Structured buffer counterpart: rides this port's owning node's own INTERIOR
	// stream frame (the SAME stream KindRecv already resolves through — see
	// flushRecvEvent above), resolved to that node's row + this port's row at the
	// call site. label maps the free-form event string to its BreadcrumbLabel*
	// index; unrecognized strings are dropped from the buffer path (still logged
	// via i.trace.Breadcrumb above) rather than silently miscoded.
	if i.stream == nil {
		return
	}
	label, ok := breadcrumbLabelFor(event)
	if !ok {
		return
	}
	s := i.stream()
	if s == nil {
		return
	}
	s.WriteEvents([]RowEvent{{
		Kind: T.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: s.NodeRowOf(), PortRow: i.portRow,
		TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
	}})
}

// breadcrumbLabelFor maps a free-form Breadcrumb event string to its
// T.BreadcrumbLabel* index for the structured buffer path. Only the closed set of
// known breadcrumb sites resolve; an unrecognized string returns ok=false — and
// check-breadcrumb-label-registered.sh fails the build if any Breadcrumb() call site's
// literal label is missing from this switch/T.BreadcrumbLabels, so a new label can no
// longer be silently dropped end to end the way probe.enterCommit/drag.jump/
// probe.commitLocal were (memory/feedback_check_the_signal_the_check_emits) — that
// temporary trio has since been removed entirely (both call sites and their
// registration), but the guard they motivated stays for the next probe.
func breadcrumbLabelFor(event string) (uint8, bool) {
	switch event {
	case "topology-loaded":
		return T.BreadcrumbTopologyLoaded, true
	case "row-seed-count-mismatch":
		return T.BreadcrumbRowSeedCountMismatch, true
	case "pole-toggle-go":
		return T.BreadcrumbPoleToggleGo, true
	case "window_clear":
		return T.BreadcrumbWindowClear, true
	case "window_open":
		return T.BreadcrumbWindowOpen, true
	case "dwell_start":
		return T.BreadcrumbDwellStart, true
	case "abc-drag":
		return T.BreadcrumbAbcDrag, true
	case "wire-send-buffer-full":
		return T.BreadcrumbWireSendBufferFull, true
	case "cascade.root":
		return T.BreadcrumbCascadeRoot, true
	case "chain-aim":
		return T.BreadcrumbChainAim, true
	case "neighbor-center-recv":
		return T.BreadcrumbNeighborCenterRecv, true
	case "neighbor-setc-recv":
		return T.BreadcrumbNeighborSetCRecv, true
	case "bead-crud":
		return T.BreadcrumbBeadCrud, true
	default:
		return 0, false
	}
}

// SendRule names the per-edge send policy applied by the source node after a
// successful TrySend. The wire stays dumb transport; the node consults the rule.
type SendRule string

const (
	// RuleConsumeGated: default send rule. Kept for compatibility with persisted
	// topology JSON; the consume gate was removed (PacedWire.Done/WaitConsumed
	// are no-ops). The only meaningful distinction is Gated() which gates
	// optional feedback ports.
	RuleConsumeGated SendRule = "consumeGated"
	// RuleFireAndForget: the node sends and does not wait for consumption.
	RuleFireAndForget SendRule = "fireAndForget"
)

// ParseSendRule converts a raw string into a SendRule.
// An empty string returns RuleConsumeGated (preserving the default-when-absent
// behavior). Any string that is not a recognised constant returns an error.
func ParseSendRule(s string) (SendRule, error) {
	switch s {
	case "":
		return RuleConsumeGated, nil
	case string(RuleConsumeGated):
		return RuleConsumeGated, nil
	case string(RuleFireAndForget):
		return RuleFireAndForget, nil
	default:
		return RuleConsumeGated, fmt.Errorf("invalid sendRule %q: must be one of %q, %q",
			s, RuleConsumeGated, RuleFireAndForget)
	}
}

// outGeom is an immutable snapshot of an Out's per-edge geometry: this edge's
// bead-step count plus its drawn straight-segment endpoints. It is assembled from
// TWO INDEPENDENT publishers, per docs/bead-lattice.md "Ownership" — the source
// node owns the count, the edgeMover owns the segment — delivered to the ONE
// reading goroutine (the node's own Update goroutine, via Geom() below) over two
// SEPARATE buffered-1, latest-wins channels (geomSendSteps/geomSendSeg), never a
// shared field either publisher writes directly. This mirrors
// per-goroutine-clock.md's speedCh Delivery pattern (SendSpeedNonBlocking/
// ApplySpeedNonBlocking): each producer sends, the one consumer owns its own copy.
//
//   - Steps: this edge's own bead-step count (docs/bead-lattice.md "The count"),
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
	// (the source node's own goroutine, docs/bead-lattice.md's step count owner),
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
// docs/bead-lattice.md "Ownership": the source node owns the count, never the
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
// docs/bead-lattice.md "Ownership": the edgeMover owns the segment, never the
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

// drainStepsNonBlocking folds the latest pending value off ch (if any) into
// *cur, without blocking. A nil ch (chan-mode Out, or an unpublished port)
// simply never selects the receive case, leaving *cur at its zero value.
func drainStepsNonBlocking(ch chan int, cur *int) {
	select {
	case v := <-ch:
		*cur = v
	default:
	}
}

// sendIntNonBlocking delivers v to ch, latest-wins: if the buffer already holds
// an undrained stale value, that stale value is dropped and replaced — mirrors
// SendSpeedNonBlocking (clock.go) for the same reason (absolute state, not an
// event stream). A nil ch (chan-mode Out) makes every case here select
// `default`, so this is a silent no-op.
func sendIntNonBlocking(ch chan int, v int) {
	select {
	case ch <- v:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- v:
	default:
	}
}

// drainSegNonBlocking is drainStepsNonBlocking's WireSegment counterpart.
func drainSegNonBlocking(ch chan WireSegment, start, end *Vec3) {
	select {
	case seg := <-ch:
		*start, *end = seg.Start, seg.End
	default:
	}
}

// sendSegNonBlocking is sendIntNonBlocking's WireSegment counterpart.
func sendSegNonBlocking(ch chan WireSegment, seg WireSegment) {
	select {
	case ch <- seg:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- seg:
	default:
	}
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
// directly (docs/bead-lattice.md "The count") — no chord/arc to derive it from
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
		SimLatencyMs: float64(steps) * DwellTicksPerBead * MsPerTick,
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

// DriveOutcome distinguishes the outcomes a drive placement can have.
// Collapsing them onto a single bool (the pre-fix shape) made "chan mode, sent
// fine, nothing more to drive" indistinguishable from "placement failed /
// wire torn down" — callers that stopped their loop on !Live() then stopped
// on every chan-mode send too. Keeping them explicit makes that conflation
// unrepresentable.
//
// DriveBufferFull is its own constant, split out from DriveFailed, for the
// same reason: a paced wire's inCh being momentarily full (PacedWire.
// SendBufferFull) is TRANSIENT — the wire's own goroutine drains inCh every
// cycle — and must never be treated as "stop, the wire is gone". Only a nil
// Out (no destination at all, a structural condition) is DriveFailed. Naming
// them apart means a caller who writes `.Failed()` gets ONLY the terminal
// case by construction; it cannot accidentally also catch buffer-full the way
// the pre-fix single bool did.
type DriveOutcome uint8

const (
	// DrivePlaced: a real bead was placed on a paced wire; delivery is driven
	// by subsequent StepOnce/StepOnceAt calls.
	DrivePlaced DriveOutcome = iota
	// DriveSentChan: chan mode (tests) — the value was sent (or dropped by a
	// full non-blocking select) on the raw channel. Nothing to drive further.
	DriveSentChan
	// DriveBufferFull: the paced wire's inCh was momentarily full
	// (PacedWire.SendBufferFull). TRANSIENT, NOT TERMINAL — a caller driving a
	// continuous-placement loop must NOT stop on this; skip this cycle's
	// placement and retry next cycle. A breadcrumb was already emitted by
	// PacedWire.Send.
	DriveBufferFull
	// DriveFailed: nil Out — there is no destination at all. Structural and
	// terminal; the caller should stop.
	DriveFailed
)

// DriveItem is an exported handle to one placed bead. Delivery is timed by the
// wire's own goroutine, not by the caller — this type reports which of the
// DriveOutcomes occurred.
type DriveItem struct {
	outcome DriveOutcome
}

// Live reports whether this DriveItem carries a bead actually placed on a
// paced wire (i.e. PlaceDriven succeeded in paced-wire mode) — outcome ==
// DrivePlaced. False for a nil Out, chan mode, a momentary buffer-full, or a
// failed placement. Callers that need ONLY "did this become a real, time-able
// in-flight bead" (e.g. Time's processing-window length) check
// this; callers that need "should I stop, the wire is gone" must check
// Failed() instead — Live() alone cannot distinguish chan-mode success,
// buffer-full, or true failure.
func (di DriveItem) Live() bool {
	return di.outcome == DrivePlaced
}

// Failed reports whether the placement failed for a STRUCTURAL, TERMINAL
// reason: a nil Out (no destination at all). It deliberately does NOT report
// true for DriveBufferFull — a momentarily-full paced-wire buffer is
// transient and self-clears as the wire's own goroutine drains it; treating
// it as terminal would silently and permanently kill a source's drive
// goroutine on ordinary transient load. Callers driving a continuous-
// placement loop should stop on Failed(), not on !Live() — a chan-mode
// successful send or a buffer-full retry are also !Live() but must not stop
// the loop. See BufferFull() for the transient case.
func (di DriveItem) Failed() bool {
	return di.outcome == DriveFailed
}

// BufferFull reports whether this placement did not go through because the
// paced wire's inCh was momentarily full. This is the DISTINCT, NON-TERMINAL
// counterpart to Failed(): a caller must handle this case (typically: skip
// this cycle, keep looping) rather than let it fall through to a generic
// failure branch, which is exactly the bug this type split fixes.
func (di DriveItem) BufferFull() bool {
	return di.outcome == DriveBufferFull
}

// PlaceDrivenAt places one bead on this Out WITHOUT spawning a walker, emits
// the SendWire trace, and returns a DriveItem reporting the outcome. tick is
// the CALLING goroutine's own clock reading, read once by the caller and
// stamped as this bead's placementTick — the wire itself no longer decides
// when a bead started (placeRequest's doc comment). Delivery timing (the
// bead's later position-advance) is still done by the wire's driver (the
// source node's mover — node_mover.go), not the caller. In chan mode (tests)
// it sends immediately on the raw channel and returns DriveSentChan, so unit
// tests keep their synchronous chan semantics (tick is unused in this path). A
// nil Out returns DriveFailed. A momentarily-full paced wire returns
// DriveBufferFull, never DriveFailed — see the DriveOutcome doc comment.
func (o *Out) PlaceDrivenAt(v int, tick int64) DriveItem {
	if o == nil {
		return DriveItem{outcome: DriveFailed}
	}
	if o.pw != nil {
		switch o.placeDrivenNoWalker(v, tick) {
		case SendPlaced:
			return DriveItem{outcome: DrivePlaced}
		default: // SendBufferFull
			return DriveItem{outcome: DriveBufferFull}
		}
	}
	// chan mode (tests, or a production dead-end unwired Out): no drive needed, send
	// now and return DriveSentChan. flushSendEvent no-ops when stream is unset (both
	// cases: no builder-injected getter).
	if o.ch != nil {
		select {
		case o.ch <- v:
			o.flushSendEvent(v, 0)
		default:
		}
		return DriveItem{outcome: DriveSentChan}
	}
	return DriveItem{outcome: DriveSentChan}
}

// Broadcast is a broadcast port: a slice of Outs the node emits the same
// value onto, each its own independent 1:1 wire.
type Broadcast []*Out

// PlaceDrivenAllAt places value v (no walker) on EVERY Out in the set, emitting
// the SendWire trace for each and appending a DriveItem per Out to dst. tick is
// the CALLER's own clock reading, read ONCE and passed to every Out in the
// set — that single shared reading is what guarantees every bead of this one
// broadcast emission shares the same placementTick (the bug this replaced: an
// earlier version let each wire's own drain pass stamp its own tick, so a
// broadcast could straddle a tick boundary between two of its beads). Once
// placed, each wire's own driver (its source node's mover) advances and
// delivers its bead independently, so the traversal still animates
// concurrently. Chan-mode Outs send immediately and contribute inert items.
func (outs Broadcast) PlaceDrivenAllAt(v int, dst []DriveItem, tick int64) []DriveItem {
	for _, o := range outs {
		if o == nil {
			continue
		}
		dst = append(dst, o.PlaceDrivenAt(v, tick))
	}
	return dst
}

// NewInPaced / NewOutPaced are used by the loader. Uses PacedWire mode. Neither the
// port nor the wire behind it holds a clock (per-goroutine-clock.md API demolition
// item 1: port accessors are gone) — a node's own Clock field is what its goroutine
// Copies from at startup.
//
// stream is this In's owning node's shared event-sink getter (nil for the many
// lean per-node tests across nodes/<Kind> that build an In directly without a
// loader — those never flush a RowEvent, matching the prior default). portRow
// is this In's own buffer PORT-ROW index (isInput=true); -1 when unresolved
// (no md, or an unwired dead-end port) — see wireInPort's doc comment for how
// the loader resolves it.
func NewInPaced(pw *PacedWire, ctx context.Context, node, port string, tr *T.Trace, stream func() EventSink, portRow int32) *In {
	return &In{pw: pw, ctx: ctx, node: node, port: port, trace: tr, stream: stream, portRow: portRow}
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
