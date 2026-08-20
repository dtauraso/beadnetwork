package portwiring

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/Wiring/interior"
	"github.com/dtauraso/wirefold/src/Bead/inport"
	"github.com/dtauraso/wirefold/src/Bead/outport"
	"github.com/dtauraso/wirefold/src/Node/rowevent"

	T "github.com/dtauraso/wirefold/src/Trace"
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

func InteriorEventSinkGetter(g func() *interior.Emitter) func() rowevent.EventSink {
	return func() rowevent.EventSink {
		e := g()
		if e == nil {
			return nil
		}
		return e
	}
}

const NoPortRow = int32(-1)

func NewInPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, getSink func() rowevent.EventSink) *inport.In {
	if b := pb.singlePaced[portName]; b.pw != nil {
		return inport.NewInPaced(b.pw, ctx, name, portName, tr, getSink, NoPortRow)
	} else {
		ch := pb.deadEndIn(portName)
		return inport.NewInChan(ch, name, portName, tr, getSink)
	}
}

func NewOutPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*outport.Out, getSink func() rowevent.EventSink) *outport.Out {
	if b := pb.singlePaced[portName]; b.pw != nil {
		targetRow := int32(-1)
		if b.pw.Target != "" {
			if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
				targetRow = r
			}
		}
		o := outport.NewOutPaced(b.pw, ctx, name, portName, tr, b.rule, b.label, getSink, NoPortRow, targetRow, NoPortRow)
		*sourceOuts = append(*sourceOuts, o)
		if pb.OutSink != nil {
			pb.OutSink[name+"."+portName] = o
		}
		return o
	}
	ch := pb.deadEndOut(portName)
	return outport.NewOutChanDeadEnd(ch, name, portName, tr)
}

func NewBroadcastPort(portName string, ctx context.Context, name string, pb PortBindings, tr *T.Trace, sourceOuts *[]*outport.Out, getSink func() rowevent.EventSink) outport.Broadcast {
	if bs := pb.broadcastPaced[portName]; len(bs) > 0 {
		outs := make(outport.Broadcast, len(bs))
		for i, b := range bs {
			targetRow := int32(-1)
			if b.pw.Target != "" {
				if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
			}
			o := outport.NewOutPaced(b.pw, ctx, name, b.handle, tr, b.rule, b.label, getSink, NoPortRow, targetRow, NoPortRow)
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
