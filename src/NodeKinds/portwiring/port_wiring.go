package portwiring

import (
	"context"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Interior"
	B "github.com/dtauraso/wirefold/src/Buffer"
)


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

func NewInPort(portName string, ctx context.Context, name string, pb PortBindings, getSink func() B.EventSink) *beadanimation.Receiver {
	if b := pb.singlePaced[portName]; b.pw != nil {
		return beadanimation.NewInPaced(b.pw, ctx, name, portName, getSink, NoPortRow)
	} else {
		ch := pb.deadEndIn(portName)
		return beadanimation.NewInChan(ch, name, portName, getSink)
	}
}

func NewOutPort(portName string, ctx context.Context, name string, pb PortBindings, sourceOuts *[]*beadanimation.Sender, getSink func() B.EventSink) *beadanimation.Sender {
	if b := pb.singlePaced[portName]; b.pw != nil {
		targetRow := int32(-1)
		if b.pw.Target != "" {
			if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
				targetRow = r
			}
		}
		o := beadanimation.NewOutPaced(b.pw, ctx, name, portName, b.rule, b.label, getSink, NoPortRow, targetRow, NoPortRow)
		*sourceOuts = append(*sourceOuts, o)
		if pb.OutSink != nil {
			pb.OutSink[name+"."+portName] = o
		}
		return o
	}
	ch := pb.deadEndOut(portName)
	return beadanimation.NewOutChanDeadEnd(ch, name, portName)
}

func NewBroadcastPort(portName string, ctx context.Context, name string, pb PortBindings, sourceOuts *[]*beadanimation.Sender, getSink func() B.EventSink) beadanimation.Broadcast {
	if bs := pb.broadcastPaced[portName]; len(bs) > 0 {
		outs := make(beadanimation.Broadcast, len(bs))
		for i, b := range bs {
			targetRow := int32(-1)
			if b.pw.Target != "" {
				if r, ok := pb.RT.NodeRowFor(b.pw.Target); ok {
					targetRow = r
				}
			}
			o := beadanimation.NewOutPaced(b.pw, ctx, name, b.handle, b.rule, b.label, getSink, NoPortRow, targetRow, NoPortRow)
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
		outs := make(beadanimation.Broadcast, len(chs))
		for i, c := range chs {
			outs[i] = beadanimation.NewOutChanDeadEnd(c, name, portName)
		}
		return outs
	}
}
