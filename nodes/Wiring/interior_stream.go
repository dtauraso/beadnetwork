// interior_stream.go — the interiorStream I/O type split out of emit_geometry.go, as a pure
// move (no logic changes): interiorStream, write/writeEvents, writeInteriorStreamFrame,
// boolU8. See bead_emit.go and port_geom_emit.go for emit_geometry.go's other two
// concerns; builders.go keeps the reflection-driven port-manifest/node-construction half.

package Wiring

import (
	"encoding/binary"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"io"
)

// interiorStream bundles ONE node's own dedicated interior fd + injected frame builder +
// a local monotonic tick counter, so emitNodeBeads/emitHeldBead/emitInputBeads can pass
// one small value instead of three loose params. Built once per node (injectClosures);
// nil-safe (a zero-value *interiorStream is fine — write is a no-op when out is nil).
type interiorStream struct {
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
	// buffer-decode.ts), so an events-only flush (writeEvents, for a Fire/Recv/Send
	// occurring BETWEEN bead-state changes) must still ship a full, valid 4-slot
	// snapshot — it reuses this cache rather than inventing/omitting bead state.
	// Populated to an all-absent 4-slot snapshot at construction (buildInteriorStream)
	// and refreshed by every write() call.
	lastPresent            []uint8
	lastValue              []int32
	lastOx, lastOy, lastOz []float32
}

// nodeRowOf reports this stream's stable buffer node-row index, so a port can stamp it
// onto a recv/send/breadcrumb RowEvent without naming the concrete interiorStream type —
// see the eventSink seam in ports.go. Called only after a non-nil check, so no nil guard.
// NodeRowOf and WriteEvents are exported (unlike the rest of interiorStream) solely so
// this type satisfies wire.EventSink from another package — Go requires an interface's
// methods be exported to be implementable outside the interface's own package.
func (s *interiorStream) NodeRowOf() int32 { return s.nodeRow }

// write packs and writes this node's current interior-slot arrays via
// writeInteriorStreamFrame, advancing its own local tick counter. No-op (including on a
// nil receiver) when out/buildFrame aren't wired — the fallback path. events carries
// this call's own row-resolved NodeBead events, recorded by the caller in the SAME
// function invocation (emitNodeBeads/emitHeldBead/emitInputBeads) that built them.
// Caches the passed bead-slot arrays (see lastPresent's doc comment) so a later
// writeEvents call has a valid snapshot to reuse.
func (s *interiorStream) write(present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) {
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
func (s *interiorStream) WriteEvents(events []wire.RowEvent) {
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
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = out.Write(hdr[:])
	_, _ = out.Write(frame)
}
