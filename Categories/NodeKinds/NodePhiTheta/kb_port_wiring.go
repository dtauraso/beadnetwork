package NodePhiTheta

import (
	"context"

	SliderPanel "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	interior "github.com/dtauraso/beadnetwork/Categories/Node/Interior"
)

type bindings interface {
	SinglePacedOf(name string) (*beadanimation.BeadLine, beadanimation.SendRule, string)
	BroadcastCountOf(name string) int
	BroadcastAt(name string, i int) (*beadanimation.BeadLine, string, beadanimation.SendRule, string)
	NodeRowFor(id string) (int32, bool)
	SetOutSink(key string, o *beadanimation.Sender)
	InteriorEmitterOf(name string) *interior.Emitter
	ClockOf() clock.Clock
	SpeedSinksOf() *SliderPanel.Sinks
	VectorOutOf(name string) chan<- TiltPanel.TiltVectorMsg
	VectorInOf(name string) <-chan TiltPanel.TiltVectorMsg
}

func NewInteriorEmitterGetter(name string, pb bindings) func() *interior.Emitter {
	var built bool
	var emitter *interior.Emitter
	return func() *interior.Emitter {
		if built {
			return emitter
		}
		built = true
		emitter = pb.InteriorEmitterOf(name)
		return emitter
	}
}

func InteriorEventSinkGetter(g func() *interior.Emitter) func() beadanimation.EventSink {
	return func() beadanimation.EventSink {
		e := g()
		if e == nil {
			return nil
		}
		return e
	}
}

func tryEmit(fn func()) {
	if fn != nil {
		fn()
	}
}

type PortDir int

const (
	PortIn PortDir = iota
	PortOut
	PortBroadcast
)

type PortSpec struct {
	Name string
	Dir  PortDir
}

const NoPortRow = int32(-1)

func NewInPort(portName string, ctx context.Context, name string, pb bindings, getSink func() beadanimation.EventSink) *beadanimation.Receiver {
	if pw, _, _ := pb.SinglePacedOf(portName); pw != nil {
		return beadanimation.NewInPaced(pw, ctx, name, portName, getSink, NoPortRow)
	}
	return beadanimation.NewInChan(make(chan int, 1), name, portName, getSink)
}

func NewOutPort(portName string, ctx context.Context, name string, pb bindings, sourceOuts *[]*beadanimation.Sender, getSink func() beadanimation.EventSink) *beadanimation.Sender {
	pw, rule, label := pb.SinglePacedOf(portName)
	if pw == nil {
		return beadanimation.NewOutChanDeadEnd(make(chan int, 1), name, portName)
	}
	targetRow := int32(-1)
	if pw.Target != "" {
		if r, ok := pb.NodeRowFor(pw.Target); ok {
			targetRow = r
		}
	}
	o := beadanimation.NewOutPaced(pw, ctx, name, portName, rule, label, getSink, NoPortRow, targetRow, NoPortRow)
	*sourceOuts = append(*sourceOuts, o)
	pb.SetOutSink(name+"."+portName, o)
	return o
}

func NewBroadcastPort(portName string, ctx context.Context, name string, pb bindings, sourceOuts *[]*beadanimation.Sender, getSink func() beadanimation.EventSink) beadanimation.Broadcast {
	n := pb.BroadcastCountOf(portName)
	outs := make(beadanimation.Broadcast, n)
	for i := 0; i < n; i++ {
		pw, handle, rule, label := pb.BroadcastAt(portName, i)
		targetRow := int32(-1)
		if pw.Target != "" {
			if r, ok := pb.NodeRowFor(pw.Target); ok {
				targetRow = r
			}
		}
		o := beadanimation.NewOutPaced(pw, ctx, name, handle, rule, label, getSink, NoPortRow, targetRow, NoPortRow)
		outs[i] = o
		*sourceOuts = append(*sourceOuts, o)
		pb.SetOutSink(name+"."+handle, o)
	}
	return outs
}
