// interior_stream.go — the InteriorStream I/O type split out of emit_geometry.go, as a pure
// move (no logic changes): InteriorStream, write/writeEvents, writeInteriorStreamFrame,
// boolU8. See bead_emit.go and port_geom_emit.go for emit_geometry.go's other two
// concerns; builders.go keeps the reflection-driven port-manifest/node-construction half.
//
// Package interior (further god-object decomposition, pure move): InteriorStream is
// exported because package Wiring's own port/build wiring (port_wiring.go, build_args.go)
// constructs and threads it through. boolU8 stays a SEPARATE unexported copy in this
// package and in package Wiring (view_stream.go/overlay_gen.go) — same precedent as the
// existing local copy of Buffer.boolU8 this file's own comment already names, rather than
// inventing a shared bool-encoding package for one trivial one-liner.
package interior

import (
	"encoding/binary"
	"io"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// InteriorStream bundles ONE node's own dedicated interior fd + injected frame builder +
// a local monotonic tick counter, so EmitNodeBeads/EmitHeldBead/EmitInputBeads can pass
// one small value instead of three loose params. Built once per node (injectClosures);
// nil-safe (a zero-value *InteriorStream is fine — write is a no-op when out is nil).
type InteriorStream struct {
	out        io.Writer
	buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte
	tick       uint32
	// nodeRow is this node's stable buffer NODE-ROW index, resolved once at
	// construction (buildInteriorStream) — carried on every NodeBead/Fire/Recv/Send
	// event this stream records (memory/feedback_no_single_writer_bridge.md).
	nodeRow int32
	// lastPresent/lastValue/lastOx/lastOy/lastOz cache the most recently written 4-slot
	// interior-bead snapshot. BuildInteriorStreamFrame's slot count is FIXED (the decoder
	// reads a constant INTERIOR_SLOTS_PER_NODE, not a length carried by the frame — see
	// buffer-decode-interior.ts), so an events-only flush (writeEvents, for a Fire/Recv/Send
	// occurring BETWEEN bead-state changes) must still ship a full, valid 4-slot
	// snapshot — it reuses this cache rather than inventing/omitting bead state.
	// Populated to an all-absent 4-slot snapshot at construction (buildInteriorStream)
	// and refreshed by every write() call.
	lastPresent            []uint8
	lastValue              []int32
	lastOx, lastOy, lastOz []float32
}

// NewInteriorStream builds a fresh *InteriorStream for one node's dedicated interior fd
// (or drive-slot fd — the two callers, newInteriorStreamGetter/newDriveStreamGetter in
// package Wiring's port_wiring.go, differ only in WHICH fd they resolve), seeded with an
// all-absent slots-slot snapshot so an events-only WriteEvents before any real write()
// still ships a valid cached snapshot (lastPresent's doc comment). Exported because
// InteriorStream's own fields stay unexported (this package's sole writer discipline) —
// construction from another package must go through here, not a struct literal.
func NewInteriorStream(out io.Writer, buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte, nodeRow int32, slots int) *InteriorStream {
	absent := make([]uint8, slots)
	zeroI := make([]int32, slots)
	zeroF := make([]float32, slots)
	return &InteriorStream{
		out: out, buildFrame: buildFrame, nodeRow: nodeRow,
		lastPresent: absent, lastValue: zeroI,
		lastOx: zeroF, lastOy: append([]float32{}, zeroF...), lastOz: append([]float32{}, zeroF...),
	}
}

// OutWriter reports this stream's underlying io.Writer, exported solely so package
// Wiring's own drive_stream_wiring_test.go can assert two streams resolve to genuinely
// DIFFERENT writers (the exact wiring mistake docs/investigations/interior-stream-framing.md
// documents) without this package exposing the field itself for general use.
func (s *InteriorStream) OutWriter() io.Writer { return s.out }

// NodeRowOf reports this stream's stable buffer node-row index, so a port can stamp it
// onto a recv/send/breadcrumb RowEvent without naming the concrete InteriorStream type —
// see the eventSink seam in ports.go. Called only after a non-nil check, so no nil guard.
// NodeRowOf and WriteEvents are exported (like the type itself now) so this type satisfies
// wire.EventSink from another package — Go requires an interface's methods be exported to
// be implementable outside the interface's own package.
func (s *InteriorStream) NodeRowOf() int32 { return s.nodeRow }

// write packs and writes this node's current interior-slot arrays via
// writeInteriorStreamFrame, advancing its own local tick counter. No-op (including on a
// nil receiver) when out/buildFrame aren't wired — the fallback path. events carries
// this call's own row-resolved NodeBead events, recorded by the caller in the SAME
// function invocation (emitNodeBeads/emitHeldBead/emitInputBeads) that built them.
// Caches the passed bead-slot arrays (see lastPresent's doc comment) so a later
// writeEvents call has a valid snapshot to reuse.
func (s *InteriorStream) write(present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) {
	if s == nil {
		return
	}
	s.lastPresent, s.lastValue = present, value
	s.lastOx, s.lastOy, s.lastOz = ox, oy, oz
	s.tick++
	writeInteriorStreamFrame(s.out, s.buildFrame, s.tick, present, value, ox, oy, oz, events)
}

// writeEvents flushes an events-only interior-stream frame: no bead-slot state has
// changed since the last write, so it reuses the cached last-known 4-slot snapshot
// (lastPresent's doc comment) and carries only the caller's new row-resolved
// RowEvents (Fire/Recv/Send — see owner_events.go). No-op on a nil receiver, same as
// write.
func (s *InteriorStream) WriteEvents(events []wire.RowEvent) {
	if s == nil {
		return
	}
	s.write(s.lastPresent, s.lastValue, s.lastOx, s.lastOy, s.lastOz, events)
}

// boolU8 converts a bool to the buffer's canonical 0/1 byte encoding — a local copy of
// Buffer.boolU8 (unexported there), avoided rather than importing Buffer into this
// Buffer-independent package.
func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// writeInteriorStreamFrame packs and writes ONE node's current fixed 4-slot interior
// state to its OWN dedicated fd (out) via buildFrame (Buffer.BuildInteriorStreamFrame,
// injected so this package needs no Buffer import) — the SECOND emitting goroutine per
// node (memory/feedback_no_single_writer_bridge.md): this node's own Update loop, called
// from the SAME goroutine as the tr.NodeBead calls beside each call site below. No-op
// when out is nil (no dedicated interior fd for this node — the fallback path) or
// buildFrame is nil (no WIREFOLD_STREAM_FDS "interior" entry). tick is a local
// monotonically-increasing counter (informational only — freshness, not correctness; the
// Interior columns themselves carry the authoritative state).
func writeInteriorStreamFrame(out io.Writer, buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte, tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) {
	if out == nil || buildFrame == nil {
		return
	}
	frame := buildFrame(tick, present, value, ox, oy, oz, events)
	// ONE Write call carrying [len:u32][payload] together, not two. The real fix for the
	// framing desync (docs/investigations/interior-stream-framing.md) is giving every emitting goroutine
	// its own fd (this file's out is now single-writer by construction — see
	// Buffer.StreamKindDrive), but a single os.File.Write per frame is cheap insurance on
	// top of that: even a single writer can, in principle, be interrupted between two
	// separate Write() calls (a short write, a signal) and desync itself, and one Write
	// closes that class entirely for the cost of one extra byte-slice allocation per
	// frame (append, not a fixed scratch buffer, since frame's own backing array is
	// buildFrame's and must not be mutated in place here).
	buf := make([]byte, 4+len(frame))
	binary.LittleEndian.PutUint32(buf[:4], uint32(len(frame)))
	copy(buf[4:], frame)
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = out.Write(buf)
}
