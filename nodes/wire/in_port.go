// in_port.go — ONE JOB: the RECEIVING end of a port pair. The In type, the two
// ways a node reads through it (PollRecv, and the Breadcrumb a windowed node
// emits on its wire identity), the row-resolved events each of those flushes onto
// the owning node's interior stream, and In's own two constructors. The sending
// end is out_port.go; the seam both ends announce through is EventSink (ports.go).

package wire

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
)

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
	case "drag.commit":
		return T.BreadcrumbDragCommit, true
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
