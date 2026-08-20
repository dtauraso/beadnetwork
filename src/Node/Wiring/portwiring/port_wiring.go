package portwiring

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/Node/wire/inport"
	"github.com/dtauraso/wirefold/src/Node/wire/outport"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

const BufInteriorSlotsPerNode = 4

func NewInteriorEmitterGetter(name string, pb PortBindings) func() *interior.Emitter {
	var built bool
	var emitter *interior.Emitter
	return func() *interior.Emitter {
		if built {
			return emitter
		}
		built = true
		if pb.InteriorEmitters == nil || *pb.InteriorEmitters == nil {
			return nil
		}
		emitter = (*pb.InteriorEmitters)[name]
		return emitter
	}
}

func InteriorEventSinkGetter(g func() *interior.Emitter) func() B.EventSink {
	return func() B.EventSink {
		e := g()
		if e == nil {
			return nil
		}
		return e
	}
}

const NoPortRow = int32(-1)

func NewInPort(portName string, ctx context.Context, name string, pb PortBindings, getSink func() B.EventSink) *inport.In {
	if b := pb.singlePaced[portName]; b.pw != nil {
		return inport.NewInPaced(b.pw, ctx, name, portName, getSink, NoPortRow)
	} else {
		ch := pb.deadEndIn(portName)
		return inport.NewInChan(ch, name, portName, getSink)
	}
}

func NewOutPort(portName string, ctx context.Context, name string, pb PortBindings, sourceOuts *[]*outport.Out, getSink func() B.EventSink) *outport.Out {
	if b := pb.singlePaced[portName]; b.pw != nil {
		targetRow := int32(-1)
		if b.pw.Target != "" {
			if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
				targetRow = r
			}
		}
		o := outport.NewOutPaced(b.pw, ctx, name, portName, b.rule, b.label, getSink, NoPortRow, targetRow, NoPortRow)
		*sourceOuts = append(*sourceOuts, o)
		if pb.OutSink != nil {
			pb.OutSink[name+"."+portName] = o
		}
		return o
	}
	ch := pb.deadEndOut(portName)
	return outport.NewOutChanDeadEnd(ch, name, portName)
}

func NewBroadcastPort(portName string, ctx context.Context, name string, pb PortBindings, sourceOuts *[]*outport.Out, getSink func() B.EventSink) outport.Broadcast {
	if bs := pb.broadcastPaced[portName]; len(bs) > 0 {
		outs := make(outport.Broadcast, len(bs))
		for i, b := range bs {
			targetRow := int32(-1)
			if b.pw.Target != "" {
				if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
			}
			o := outport.NewOutPaced(b.pw, ctx, name, b.handle, b.rule, b.label, getSink, NoPortRow, targetRow, NoPortRow)
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
			outs[i] = outport.NewOutChanDeadEnd(c, name, portName)
		}
		return outs
	}
}
