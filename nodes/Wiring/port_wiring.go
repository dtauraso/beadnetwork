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

func newInPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, getStream func() *interiorStream) *wire.In {
	if b := pb.singlePaced[portName]; b.pw != nil {
		portRow := int32(-1)
		if pb.md != nil {
			if r, ok := pb.md.PortRowFor(name, portName, true); ok {
				portRow = r
			}
		}
		return wire.NewInPaced(b.pw, ctx, name, portName, tr, asEventSinkGetter(getStream), portRow)
	} else {
		ch := pb.deadEndIn(portName)
		return wire.NewInChan(ch, name, portName, tr, asEventSinkGetter(getStream))
	}
}

func newOutPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interiorStream) *wire.Out {
	if b := pb.singlePaced[portName]; b.pw != nil {
		portRow, targetRow, targetPortRow := int32(-1), int32(-1), int32(-1)
		if pb.md != nil {
			if r, ok := pb.md.PortRowFor(name, portName, false); ok {
				portRow = r
			}
			if b.pw.Target != "" {
				if r, ok := pb.md.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
				if r, ok := pb.md.PortRowFor(b.pw.Target, b.pw.TargetHandle, true); ok {
					targetPortRow = r
				}
			}
		}
		o := wire.NewOutPaced(b.pw, ctx, name, portName, tr, b.rule, b.arc, b.latency, b.seg, b.label, asEventSinkGetter(getStream), portRow, targetRow, targetPortRow)
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
			portRow, targetRow, targetPortRow := int32(-1), int32(-1), int32(-1)
			if pb.md != nil {
				if r, ok := pb.md.PortRowFor(name, b.handle, false); ok {
					portRow = r
				}
				if b.pw.Target != "" {
					if r, ok := pb.md.NodeRowFor(b.pw.Target); ok {
						targetRow = r
					}
					if r, ok := pb.md.PortRowFor(b.pw.Target, b.pw.TargetHandle, true); ok {
						targetPortRow = r
					}
				}
			}
			o := wire.NewOutPaced(b.pw, ctx, name, b.handle, tr, b.rule, b.arc, b.latency, b.seg, b.label, asEventSinkGetter(getStream), portRow, targetRow, targetPortRow)
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
