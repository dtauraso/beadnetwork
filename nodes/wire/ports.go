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
//
// ONE JOB for this file: the overview above, plus the ONE thing both ends of a
// port pair share — the EventSink seam they announce through. The ends
// themselves are two files, because they are two types with two owners: in_port.go
// (In, its receive and its breadcrumb) and out_port.go (Out, its geometry and its
// send). They remain two ends of ONE concept by living in this one package, not by
// living in one file. The rest of the pair's vocabulary is send_rule.go (the
// per-edge policy), drive_item.go (what a placement reports back) and broadcast.go
// (one emission across a set of Outs).

package wire

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
