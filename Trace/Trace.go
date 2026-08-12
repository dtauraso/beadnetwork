// Trace is now a thin breadcrumb writer plus the closed EVENT-KIND vocabulary shared
// with the per-owner buffer streams (memory/feedback_no_single_writer_bridge.md, memory/
// feedback_no_single_writer_bridge.md). Every domain event (recv/fire/send/geometry/
// camera/selection/overlay-toggle/...) is now written by its OWNING goroutine directly
// as a RowEvent onto that goroutine's own dedicated stream frame (nodes/Wiring's
// owner_events.go and friends) — there is no more central Trace channel, no drain
// goroutine, and no second (redundant) serialization of those events through this
// package. Buffer.KindID resolves a RowEvent's string Kind to its numeric id via
// TraceEventKinds (kind_events.go), which stays the single source of that vocabulary
// (also generated into tools/topology-vscode/src/schema/trace-kinds.ts).
//
// This file holds the Trace struct itself and its two surviving events. The closed Kind
// vocabulary is kind_events.go, the breadcrumb-label sub-vocabulary is
// breadcrumb_labels.go, and the value types those events carry (PortGeom/Event) are
// event.go — each file its own concern, all still package Trace (marshal.go keeps the
// pure JSON-shape marshalling this file's methods call).
//
// The one exception is NodeBead: emitRefillSlide's per-frame interior-refill-slide
// animation (nodes/Wiring/emit_geometry.go) calls Trace.NodeBead directly with no
// RowEvent dual of its own — kept as an actual Trace event for that reason, delivered
// synchronously (no channel) to an optional in-process onEvent hook (headless tests)
// and/or sink (in-process test buffer). Neither is wired in production (main.go passes
// none), so this is a no-op cost on the live path.
//
// Breadcrumb is the other survivor: a free-form diagnostic line (outside the closed
// Kind vocabulary), written directly — one `sink.Write` call per breadcrumb, on the
// calling goroutine, no channel. A single small write to a pipe is atomic
// per POSIX PIPE_BUF, so concurrent breadcrumbs from many goroutines never interleave
// into a fused line; breadcrumbs are short, sparse control-event lines (see
// .claude/rules/go-debugging.md's "Debugging the Go layer (probe breadcrumbs)" section),
// never a per-tick firehose, so this holds in practice.

package Trace

import "io"

// Trace holds the optional in-process test sink (headless tests only — never wired in
// production). It is set ONCE at startup (New*) before any producer goroutine exists,
// and never mutated again — read-only for the rest of the process, so every later
// caller (running in a goroutine spawned after startup) sees the write via the
// ordinary happens-before edge from goroutine creation. There is no
// second writer of the field, and NodeBead/Breadcrumb only ever READ it.
//
// There used to be a second, PRODUCTION debug sink here (os.Stdout, wired by
// SetDebugSink) that gave every Breadcrumb() call a free-form JSON stdout line, routed
// by the ext host to .probe/go-debug.jsonl. That production JSON path is RETIRED
// (task/breadcrumbs-binary-buffer): breadcrumbs now ride each emitting goroutine's own
// per-owner buffer stream as a structured Kind==KindBreadcrumb EVENT row (see
// owner_events.go's RowEvent.Label/Debug/Text and each call site's writeStreamFrame/
// writeEvents/EmitBreadcrumb call) — the ext host decodes them off that stream like any
// other event and probe-merge.sh --debug filters on the Debug flag. Breadcrumb below
// keeps writing the in-process sink ONLY — that path is unrelated to the removed
// production one and still backs headless tests (BreadcrumbLabel/BreadcrumbValue).
type Trace struct {
	sink    io.Writer
	onEvent func(Event) // optional in-process observation hook (headless tests only)
}

// New allocates a Trace with no sinks wired.
func New() *Trace {
	return NewWithSink(nil)
}

// NewWithSink is like New but wires sink as the in-process test-observation sink (see
// Breadcrumb/NodeBead's doc comments) — never wired in production.
func NewWithSink(sink io.Writer) *Trace {
	return NewWithSinkHook(sink, nil)
}

// NewWithSinkHook is like NewWithSink but also installs onEvent, called synchronously
// (on the calling goroutine) by NodeBead — the one surviving Trace event. Pass nil for
// onEvent to omit the hook (production always does).
func NewWithSinkHook(sink io.Writer, onEvent func(Event)) *Trace {
	return &Trace{sink: sink, onEvent: onEvent}
}

// SetSink wires (or replaces) the in-process TEST-OBSERVATION sink after construction —
// for tests that build a *Trace via a helper (e.g. LoadTopology) that doesn't take a
// sink at New time, rather than the (removed) production stdout path. Never wired in
// production. Set once, before any producer goroutine exists (same ordering
// requirement the old SetDebugSink had), relying on the happens-before edge
// from goroutine creation.
func (t *Trace) SetSink(w io.Writer) {
	if t == nil {
		return
	}
	t.sink = w
}

// Breadcrumb writes a free-form diagnostic line DIRECTLY to the in-process test sink —
// one `sink.Write` call, on the CALLING goroutine. No channel, no ordinal:
// breadcrumbs are outside the closed Kind vocabulary (RowEvents carry the closed
// vocabulary; this is a control-event log line). Production no longer has a stdout
// sink here (see this file's header + Trace struct doc comments) — the PRODUCTION
// observation path for a breadcrumb is now the structured Kind==KindBreadcrumb buffer
// EVENT each call site emits on its own owning stream, not this method. This method's
// only remaining reader is the in-process test sink (headless model/gate tests poll
// it via BreadcrumbLabel/BreadcrumbValue). A breadcrumb with no sink wired is a cheap
// no-op.
func (t *Trace) Breadcrumb(label, node, port, value string) {
	if t == nil || t.sink == nil {
		return
	}
	b, err := marshalBreadcrumb(label, node, port, value)
	if err != nil {
		return
	}
	b = append(b, '\n')
	_, _ = t.sink.Write(b)
}

// NodeBead is the one surviving Trace EVENT (see this file's header doc comment):
// emitRefillSlide (nodes/Wiring/emit_geometry.go) calls it directly, once per animation
// frame, with no RowEvent dual of its own. Delivered synchronously to the optional
// onEvent hook and/or sink — neither wired in production, so this is a cheap no-op on
// the live path. nodeID + (row,col) key the slot; present/value/x/y/z carry its state.
func (t *Trace) NodeBead(nodeID string, row, col int, present bool, value int, x, y, z float64) {
	if t == nil {
		return
	}
	ev := Event{Kind: KindNodeBead, Node: nodeID, Row: row, Col: col, Present: present, Value: value, X: x, Y: y, Z: z}
	if t.sink != nil {
		if b, err := marshalNodeBead(ev); err == nil {
			_, _ = t.sink.Write(append(b, '\n'))
		}
	}
	if t.onEvent != nil {
		t.onEvent(ev)
	}
}
