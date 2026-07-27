// port_wiring.go — wiring each port field (In/Out/Broadcast) to its resolved
// PortBindings entry and this node's own interior-stream instance.

package Wiring

import (
	"context"
	"reflect"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// speedChanFieldNames lists every field name a node kind may declare to receive
// a speed-delivery channel. Most kinds (input/hold/Time/pacer,
// gatecommon.GateNode) run exactly one clock-owning goroutine and declare only
// SpeedCh. Pulse/HoldFlip split into a main loop plus one-or-two
// gatecommon.DriveHeld goroutines (one per driven Out) — each is an
// INDEPENDENT clock copy, so each needs its OWN channel: sharing one channel
// across two goroutines would silently starve whichever one loses a given
// receive, which is exactly the "no goroutine left behind" failure item 3 of
// per-goroutine-clock.md guards against. DriveSpeedCh/Out1SpeedCh/Out2SpeedCh
// are those extra per-drive-goroutine channels; a struct that doesn't declare
// a given name simply doesn't get one (injectSpeedChans is a no-op per name
// when the field is absent, same contract as injectFunc).
var speedChanFieldNames = []string{"SpeedCh", "DriveSpeedCh", "Out1SpeedCh", "Out2SpeedCh"}

// injectSpeedChans creates one fresh buffered-1 speed channel per field name in
// speedChanFieldNames that the struct pointed to by v actually declares (typed
// exactly `<-chan float64`), injects its RECEIVE end into that field, and
// appends its SEND end to *pb.speedSinks — the loader's build-wide accumulator
// of every goroutine's speed channel, broadcast to on a speed change. A no-op
// when pb.speedSinks is nil (test builds with no loader): such a node's
// goroutines simply have no speed channel to poll, exactly like they had no
// shared clock to receive a speed change on before this plan either.
func injectSpeedChans(v reflect.Value, pb PortBindings) {
	if pb.speedSinks == nil {
		return
	}
	tSpeedChan := reflect.TypeFor[<-chan float64]()
	for _, fname := range speedChanFieldNames {
		f := v.FieldByName(fname)
		if !f.IsValid() || !f.CanSet() || f.Type() != tSpeedChan {
			continue
		}
		speedCh := make(chan float64, 1)
		f.Set(reflect.ValueOf((<-chan float64)(speedCh)))
		*pb.speedSinks = append(*pb.speedSinks, speedCh)
	}
}

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

// wirePorts wires every port field (In/Out/Broadcast) discovered by reflectPorts
// with traced wrappers, resolving each from pb's paced bindings when present and
// falling back to a dead-end chan/slice otherwise. sourceOuts accumulates every
// paced Out built (for EmitGeometry's closure, injected by injectClosures) and
// pb.outSink (when non-nil) is populated so the loader can index Outs by edge.
// getStream is this node's shared interior-stream getter (newInteriorStreamGetter),
// threaded through so Recv/Send can flush their own RowEvent onto the same frame
// Fire/EmitNodeBeads use.
func wirePorts(ctx context.Context, v reflect.Value, nodePtr any, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interiorStream) {
	ports := reflectPorts(nodePtr)
	for _, port := range ports {
		f := v.FieldByName(port.Name)
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		switch port.Dir {
		case PortIn:
			wireInPort(f, port.Name, ctx, name, pb, tr, getStream)
		case PortOut:
			wireOutPort(f, port.Name, ctx, name, pb, tr, sourceOuts, getStream)
		case PortBroadcast:
			wireBroadcastPort(f, port.Name, ctx, name, pb, tr, sourceOuts, getStream)
		}
	}
}

// wireInPort resolves a single PortIn field: a paced binding (NewInPaced) when
// pb has one for this port name, otherwise a dead-end chan wrapper.
//
// Neither branch carries a clock (per-goroutine-clock.md API demolition item 1: port
// accessors are gone) — an unwired In just polls a dead-end channel that never
// delivers, staying inert by precondition-gating (validate.go) exactly like a wired
// node whose peer never sends; its owning node paces off its OWN Clock field/Copy(),
// never off this port.
//
// A paced In's portRow (its own buffer PORT-ROW, isInput=true) is resolved once here
// from pb.md's row table (populated at MoveDispatch construction, before any node's
// own construction — see PortRowFor's doc comment), and stream is this node's shared
// interior-stream getter: both are read later by In.PollRecv, on this node's own
// Update goroutine, to flush a KindRecv RowEvent (owner_events.go).
func wireInPort(f reflect.Value, portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, getStream func() *interiorStream) {
	if b := pb.singlePaced[portName]; b.pw != nil {
		portRow := int32(-1)
		if pb.md != nil {
			if r, ok := pb.md.PortRowFor(name, portName, true); ok {
				portRow = r
			}
		}
		in := wire.NewInPaced(b.pw, ctx, name, portName, tr, asEventSinkGetter(getStream), portRow)
		f.Set(reflect.ValueOf(in))
	} else {
		ch := pb.deadEndIn(portName)
		in := wire.NewInChan(ch, name, portName, tr, asEventSinkGetter(getStream))
		f.Set(reflect.ValueOf(in))
	}
}

// wireOutPort resolves a single PortOut field: a paced binding
// (NewOutPaced, with the edge's own send rule/arc/latency/segment/label) when pb
// has one for this port name, otherwise a dead-end chan wrapper. The resolved
// paced Out is appended to sourceOuts and (when pb.outSink is non-nil) recorded
// under "node.port" for the loader's node-move travel-time updates.
//
// A paced Out's own portRow (isInput=false) plus its destination's targetRow/
// targetPortRow are resolved once here from pb.md's row tables (same timing as
// wireInPort's portRow) — the destination is static (b.pw.Target/TargetHandle never
// change after wiring), so resolving it once at construction and reading it later on
// this node's own Update goroutine (Out.PlaceDrivenAt/placeDrivenNoWalker) matches
// edgeMover's existing static-field-resolved-once discipline (edgeRow).
func wireOutPort(f reflect.Value, portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interiorStream) {
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
		f.Set(reflect.ValueOf(o))
	} else {
		ch := pb.deadEndOut(portName)
		f.Set(reflect.ValueOf(wire.NewOutChanForTest(ch, name, portName, tr)))
	}
}

// wireBroadcastPort resolves a PortBroadcast field: one paced Out per fan-out
// element recorded in pb.broadcastPaced (each with its own handle/rule/arc/
// latency/segment/label) when present, otherwise a dead-end chan slice. Each
// resolved paced Out is appended to sourceOuts and (when pb.outSink is
// non-nil) recorded under "node.handle". Row resolution mirrors wireOutPort's,
// per fan-out element.
func wireBroadcastPort(f reflect.Value, portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*wire.Out, getStream func() *interiorStream) {
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
		f.Set(reflect.ValueOf(outs))
	} else {
		chs := pb.deadEndOutSlice(portName)
		outs := make(wire.Broadcast, len(chs))
		for i, c := range chs {
			outs[i] = wire.NewOutChanForTest(c, name, portName, tr)
		}
		f.Set(reflect.ValueOf(outs))
	}
}
