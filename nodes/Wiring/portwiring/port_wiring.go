package portwiring

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"

	T "github.com/dtauraso/wirefold/Trace"
)

const BufInteriorSlotsPerNode = 4

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

func AsEventSinkGetter(g func() *interior.InteriorStream) func() rowevent.EventSink {
	return func() rowevent.EventSink {
		s := g()
		if s == nil {
			return nil
		}
		return s
	}
}

const NoPortRow = int32(-1)

func NewInPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, getStream func() *interior.InteriorStream) *wire.In {
	if b := pb.singlePaced[portName]; b.pw != nil {
		return wire.NewInPaced(b.pw, ctx, name, portName, tr, AsEventSinkGetter(getStream), NoPortRow)
	} else {
		ch := pb.deadEndIn(portName)
		return wire.NewInChan(ch, name, portName, tr, AsEventSinkGetter(getStream))
	}
}

func NewOutPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*outport.Out, getStream func() *interior.InteriorStream) *outport.Out {
	if b := pb.singlePaced[portName]; b.pw != nil {
		targetRow := int32(-1)
		if b.pw.Target != "" {
			if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
				targetRow = r
			}
		}
		o := outport.NewOutPaced(b.pw, ctx, name, portName, tr, b.rule, b.steps, b.seg, b.label, AsEventSinkGetter(getStream), NoPortRow, targetRow, NoPortRow)
		*sourceOuts = append(*sourceOuts, o)
		if pb.OutSink != nil {
			pb.OutSink[name+"."+portName] = o
		}
		return o
	}
	ch := pb.deadEndOut(portName)
	return outport.NewOutChanDeadEnd(ch, name, portName, tr)
}

func NewBroadcastPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*outport.Out, getStream func() *interior.InteriorStream) outport.Broadcast {
	if bs := pb.broadcastPaced[portName]; len(bs) > 0 {
		outs := make(outport.Broadcast, len(bs))
		for i, b := range bs {
			targetRow := int32(-1)
			if b.pw.Target != "" {
				if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
			}
			o := outport.NewOutPaced(b.pw, ctx, name, b.handle, tr, b.rule, b.steps, b.seg, b.label, AsEventSinkGetter(getStream), NoPortRow, targetRow, NoPortRow)
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
		outs := make(outport.Broadcast, len(chs))
		for i, c := range chs {
			outs[i] = outport.NewOutChanDeadEnd(c, name, portName, tr)
		}
		return outs
	}
}
