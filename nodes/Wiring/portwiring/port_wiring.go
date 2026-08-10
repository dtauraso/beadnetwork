// port_wiring.go — wiring each port field (In/Out/Broadcast) to its resolved
// PortBindings entry and this node's own interior-stream instance.

package portwiring

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// BufInteriorSlotsPerNode is a local copy of Buffer.BufInteriorSlotsPerNode's value
// (4 — the fixed interior-bead slot count per node), kept here rather than importing
// Buffer (see boolU8's doc comment, nodes/Wiring, for the existing precedent of
// duplicating a small Buffer constant to stay Buffer-independent). Used only to size
// NewInteriorStreamGetter's initial all-absent bead-slot cache.
const BufInteriorSlotsPerNode = 4

// NewInteriorStreamGetter returns a func() *interior.InteriorStream that lazily builds
// (exactly once) and thereafter always returns THIS node's one dedicated
// interior-stream instance from *pb.InteriorOuts — so every closure/port
// belonging to this node (EmitNodeBeads/EmitHeldBead/EmitInputBeads via
// injectClosures, and Fire/Recv/Send via the Fire closure and In/Out — see
// wirePorts) shares the SAME instance, and therefore the same cached last-known
// bead-slot snapshot (interiorStream.lastPresent's doc comment) a Fire/Recv/Send
// event needs to flush a valid frame between bead-state changes.
//
// Lazy because *pb.InteriorOuts is only populated by main.go AFTER LoadTopology
// returns (i.e. after this node's own construction runs) — see the prior
// buildInteriorStream doc comment this replaces. The returned func's first REAL
// call is always made from this node's OWN Update goroutine (after node-goroutine
// launch, by which point interiorOuts is fully populated and never mutated again):
// exactly one goroutine ever calls this closure, matching
// every other single-writer-per-goroutine field in this package.
func NewInteriorStreamGetter(name string, pb PortBindings) func() *interior.InteriorStream {
	var built bool
	var stream *interior.InteriorStream
	return func() *interior.InteriorStream {
		if built {
			return stream
		}
		built = true
		if pb.InteriorOuts == nil || *pb.InteriorOuts == nil {
			return nil
		}
		out, ok := (*pb.InteriorOuts)[name]
		if !ok || out == nil || pb.BuildInteriorFrame == nil || *pb.BuildInteriorFrame == nil {
			return nil
		}
		nodeRow := int32(-1)
		if r, ok := pb.RT.NodeRowFor(name); ok {
			nodeRow = r
		}
		stream = interior.NewInteriorStream(out, *pb.BuildInteriorFrame, nodeRow, BufInteriorSlotsPerNode)
		return stream
	}
}

// NewDriveStreamGetter is NewInteriorStreamGetter's counterpart for a gatecommon.DriveHeld
// drive goroutine's OWN dedicated stream (Buffer.StreamKindDrive; docs/interior-stream-
// framing.md) — the fix for the framing desync that getter's doc comment describes:
// a DriveHeld goroutine used to record its Send events through the SAME *interior.InteriorStream
// the node's own Update goroutine writes, which is exactly the two-goroutines-one-fd
// violation that corrupted framing. This getter instead resolves (*pb.DriveOuts)[name]
// [slot] — a DIFFERENT fd, dedicated to this one drive slot — so the Out a caller builds
// via BuildArgs.DriveOut never shares a stream instance with the node's own getStream.
// Lazy-cache-once for the SAME reason NewInteriorStreamGetter is: *pb.DriveOuts is
// only populated by main.go after LoadTopology returns, and the first real call is always
// made from this node's own Update goroutine (the goroutine that then spawns the
// DriveHeld goroutine using the *wire.Out this getter feeds — no data race, since the
// getter itself never runs concurrently: it lazily builds once before DriveHeld's own
// goroutine exists, then only returns the already-built pointer thereafter).
func NewDriveStreamGetter(name string, slot int, pb PortBindings) func() *interior.InteriorStream {
	var built bool
	var stream *interior.InteriorStream
	return func() *interior.InteriorStream {
		if built {
			return stream
		}
		built = true
		if pb.DriveOuts == nil || *pb.DriveOuts == nil {
			return nil
		}
		slots, ok := (*pb.DriveOuts)[name]
		if !ok || slot < 0 || slot >= len(slots) || slots[slot] == nil || pb.BuildInteriorFrame == nil || *pb.BuildInteriorFrame == nil {
			return nil
		}
		nodeRow := int32(-1)
		if r, ok := pb.RT.NodeRowFor(name); ok {
			nodeRow = r
		}
		stream = interior.NewInteriorStream(slots[slot], *pb.BuildInteriorFrame, nodeRow, BufInteriorSlotsPerNode)
		return stream
	}
}

// AsEventSinkGetter adapts a concrete interior-stream getter into the eventSink getter a
// port holds, PRESERVING nil: when the underlying getter yields no stream (nil
// *interior.InteriorStream), this returns a TRUE nil interface, not an interface value wrapping a
// nil pointer — so a port's `if s == nil` guard still fires exactly as it did against the
// concrete pointer. The emit machinery (injectClosures/emitNodeBeads/emitHeldBead) keeps
// the concrete getter unchanged; only In/Out ports route through this seam.
func AsEventSinkGetter(g func() *interior.InteriorStream) func() wire.EventSink {
	return func() wire.EventSink {
		s := g()
		if s == nil {
			return nil
		}
		return s
	}
}

// NoPortRow is always -1 now: a port has no buffer row of its own (docs/bead-model/channels-not-ports.md
// — no Port block, no port-row table). Kept as a named sentinel rather than a bare -1
// literal at each call site below so the reason reads at the call site, not just here.
const NoPortRow = int32(-1)

func NewInPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, getStream func() *interior.InteriorStream) *wire.In {
	if b := pb.singlePaced[portName]; b.pw != nil {
		return wire.NewInPaced(b.pw, ctx, name, portName, tr, AsEventSinkGetter(getStream), NoPortRow)
	} else {
		ch := pb.deadEndIn(portName)
		return wire.NewInChan(ch, name, portName, tr, AsEventSinkGetter(getStream))
	}
}

func NewOutPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interior.InteriorStream) *wire.Out {
	if b := pb.singlePaced[portName]; b.pw != nil {
		targetRow := int32(-1)
		if b.pw.Target != "" {
			if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
				targetRow = r
			}
		}
		o := wire.NewOutPaced(b.pw, ctx, name, portName, tr, b.rule, b.steps, b.seg, b.label, AsEventSinkGetter(getStream), NoPortRow, targetRow, NoPortRow)
		*sourceOuts = append(*sourceOuts, o)
		if pb.OutSink != nil {
			pb.OutSink[name+"."+portName] = o
		}
		return o
	}
	ch := pb.deadEndOut(portName)
	return wire.NewOutChanDeadEnd(ch, name, portName, tr)
}

func NewBroadcastPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interior.InteriorStream) wire.Broadcast {
	if bs := pb.broadcastPaced[portName]; len(bs) > 0 {
		outs := make(wire.Broadcast, len(bs))
		for i, b := range bs {
			targetRow := int32(-1)
			if b.pw.Target != "" {
				if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
			}
			o := wire.NewOutPaced(b.pw, ctx, name, b.handle, tr, b.rule, b.steps, b.seg, b.label, AsEventSinkGetter(getStream), NoPortRow, targetRow, NoPortRow)
			outs[i] = o
			*sourceOuts = append(*sourceOuts, o)
			if pb.OutSink != nil {
				pb.OutSink[name+"."+b.handle] = o
			}
		}
		return outs
	}
	{
		chs := pb.deadEndOutSlice(portName)
		outs := make(wire.Broadcast, len(chs))
		for i, c := range chs {
			outs[i] = wire.NewOutChanDeadEnd(c, name, portName, tr)
		}
		return outs
	}
}
