// port_wiring.go — wiring each port field (In/Out/Broadcast) to its resolved
// PortBindings entry and this node's own interior-stream instance.

package Wiring

import (
	"context"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// bufInteriorSlotsPerNode is a local copy of Buffer.BufInteriorSlotsPerNode's value
// (4 — the fixed interior-bead slot count per node), kept here rather than importing
// Buffer (see boolU8's doc comment for the existing precedent of this package
// duplicating a small Buffer constant to stay Buffer-independent). Used only to size
// newInteriorStreamGetter's initial all-absent bead-slot cache.
const bufInteriorSlotsPerNode = 4

// newInteriorStreamGetter returns a func() *interiorStream that lazily builds
// (exactly once) and thereafter always returns THIS node's one dedicated
// interior-stream instance from pb.md.sw.interiorOuts — so every closure/port
// belonging to this node (EmitNodeBeads/EmitHeldBead/EmitInputBeads via
// injectClosures, and Fire/Recv/Send via the Fire closure and In/Out — see
// wirePorts) shares the SAME instance, and therefore the same cached last-known
// bead-slot snapshot (interiorStream.lastPresent's doc comment) a Fire/Recv/Send
// event needs to flush a valid frame between bead-state changes.
//
// Lazy because pb.md.sw.interiorOuts is only populated by main.go AFTER LoadTopology
// returns (i.e. after this node's own construction runs) — see the prior
// buildInteriorStream doc comment this replaces. The returned func's first REAL
// call is always made from this node's OWN Update goroutine (after node-goroutine
// launch, by which point interiorOuts is fully populated and never mutated again):
// exactly one goroutine ever calls this closure, matching
// every other single-writer-per-goroutine field in this package.
func newInteriorStreamGetter(name string, pb PortBindings) func() *interiorStream {
	var built bool
	var stream *interiorStream
	return func() *interiorStream {
		if built {
			return stream
		}
		built = true
		if pb.md == nil || pb.md.sw.interiorOuts == nil {
			return nil
		}
		out, ok := pb.md.sw.interiorOuts[name]
		if !ok || out == nil || pb.md.sw.buildInteriorFrame == nil {
			return nil
		}
		nodeRow := int32(-1)
		if r, ok := pb.md.NodeRowFor(name); ok {
			nodeRow = r
		}
		absent := make([]uint8, bufInteriorSlotsPerNode)
		zeroI := make([]int32, bufInteriorSlotsPerNode)
		zeroF := make([]float32, bufInteriorSlotsPerNode)
		stream = &interiorStream{
			out: out, buildFrame: pb.md.sw.buildInteriorFrame, nodeRow: nodeRow,
			lastPresent: absent, lastValue: zeroI,
			lastOx: zeroF, lastOy: append([]float32{}, zeroF...), lastOz: append([]float32{}, zeroF...),
		}
		return stream
	}
}

// newDriveStreamGetter is newInteriorStreamGetter's counterpart for a gatecommon.DriveHeld
// drive goroutine's OWN dedicated stream (Buffer.StreamKindDrive; docs/interior-stream-
// framing.md) — the fix for the framing desync that getter's doc comment describes:
// a DriveHeld goroutine used to record its Send events through the SAME *interiorStream
// the node's own Update goroutine writes, which is exactly the two-goroutines-one-fd
// violation that corrupted framing. This getter instead resolves pb.md.sw.driveOuts[name]
// [slot] — a DIFFERENT fd, dedicated to this one drive slot — so the Out a caller builds
// via BuildArgs.DriveOut never shares a stream instance with the node's own getStream.
// Lazy-cache-once for the SAME reason newInteriorStreamGetter is: pb.md.sw.driveOuts is
// only populated by main.go after LoadTopology returns, and the first real call is always
// made from this node's own Update goroutine (the goroutine that then spawns the
// DriveHeld goroutine using the *wire.Out this getter feeds — no data race, since the
// getter itself never runs concurrently: it lazily builds once before DriveHeld's own
// goroutine exists, then only returns the already-built pointer thereafter).
func newDriveStreamGetter(name string, slot int, pb PortBindings) func() *interiorStream {
	var built bool
	var stream *interiorStream
	return func() *interiorStream {
		if built {
			return stream
		}
		built = true
		if pb.md == nil || pb.md.sw.driveOuts == nil {
			return nil
		}
		slots, ok := pb.md.sw.driveOuts[name]
		if !ok || slot < 0 || slot >= len(slots) || slots[slot] == nil || pb.md.sw.buildInteriorFrame == nil {
			return nil
		}
		nodeRow := int32(-1)
		if r, ok := pb.md.NodeRowFor(name); ok {
			nodeRow = r
		}
		absent := make([]uint8, bufInteriorSlotsPerNode)
		zeroI := make([]int32, bufInteriorSlotsPerNode)
		zeroF := make([]float32, bufInteriorSlotsPerNode)
		stream = &interiorStream{
			out: slots[slot], buildFrame: pb.md.sw.buildInteriorFrame, nodeRow: nodeRow,
			lastPresent: absent, lastValue: zeroI,
			lastOx: zeroF, lastOy: append([]float32{}, zeroF...), lastOz: append([]float32{}, zeroF...),
		}
		return stream
	}
}

// asEventSinkGetter adapts a concrete interior-stream getter into the eventSink getter a
// port holds, PRESERVING nil: when the underlying getter yields no stream (nil
// *interiorStream), this returns a TRUE nil interface, not an interface value wrapping a
// nil pointer — so a port's `if s == nil` guard still fires exactly as it did against the
// concrete pointer. The emit machinery (injectClosures/emitNodeBeads/emitHeldBead) keeps
// the concrete getter unchanged; only In/Out ports route through this seam.
func asEventSinkGetter(g func() *interiorStream) func() wire.EventSink {
	return func() wire.EventSink {
		s := g()
		if s == nil {
			return nil
		}
		return s
	}
}

// portRow is always -1 now: a port has no buffer row of its own (docs/channels-not-ports.md
// — no Port block, no port-row table). Kept as a named sentinel rather than a bare -1
// literal at each call site below so the reason reads at the call site, not just here.
const noPortRow = int32(-1)

func newInPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, getStream func() *interiorStream) *wire.In {
	if b := pb.singlePaced[portName]; b.pw != nil {
		return wire.NewInPaced(b.pw, ctx, name, portName, tr, asEventSinkGetter(getStream), noPortRow)
	} else {
		ch := pb.deadEndIn(portName)
		return wire.NewInChan(ch, name, portName, tr, asEventSinkGetter(getStream))
	}
}

func newOutPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interiorStream) *wire.Out {
	if b := pb.singlePaced[portName]; b.pw != nil {
		targetRow := int32(-1)
		if pb.md != nil && b.pw.Target != "" {
			if r, ok := pb.md.NodeRowFor(b.pw.Target); ok {
				targetRow = r
			}
		}
		o := wire.NewOutPaced(b.pw, ctx, name, portName, tr, b.rule, b.steps, b.seg, b.label, asEventSinkGetter(getStream), noPortRow, targetRow, noPortRow)
		*sourceOuts = append(*sourceOuts, o)
		if pb.outSink != nil {
			pb.outSink[name+"."+portName] = o
		}
		return o
	}
	ch := pb.deadEndOut(portName)
	return wire.NewOutChanForTest(ch, name, portName, tr)
}

func newBroadcastPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interiorStream) wire.Broadcast {
	if bs := pb.broadcastPaced[portName]; len(bs) > 0 {
		outs := make(wire.Broadcast, len(bs))
		for i, b := range bs {
			targetRow := int32(-1)
			if pb.md != nil && b.pw.Target != "" {
				if r, ok := pb.md.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
			}
			o := wire.NewOutPaced(b.pw, ctx, name, b.handle, tr, b.rule, b.steps, b.seg, b.label, asEventSinkGetter(getStream), noPortRow, targetRow, noPortRow)
			outs[i] = o
			*sourceOuts = append(*sourceOuts, o)
			if pb.outSink != nil {
				pb.outSink[name+"."+b.handle] = o
			}
		}
		return outs
	}
	{
		chs := pb.deadEndOutSlice(portName)
		outs := make(wire.Broadcast, len(chs))
		for i, c := range chs {
			outs[i] = wire.NewOutChanForTest(c, name, portName, tr)
		}
		return outs
	}
}
