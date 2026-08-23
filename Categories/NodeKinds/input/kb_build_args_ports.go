package input

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
)

func (a BuildArgs) In(portName string) *beadanimation.Receiver {
	a.mustDeclare(portName, portwiring.PortIn)
	return NewInPort(portName, a.Ctx, a.Name, a.PB, InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Out(portName string) *beadanimation.Sender {
	a.mustDeclare(portName, portwiring.PortOut)
	return NewOutPort(portName, a.Ctx, a.Name, a.PB, a.sourceOuts, InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Broadcast(portName string) beadanimation.Broadcast {
	a.mustDeclare(portName, portwiring.PortBroadcast)
	return NewBroadcastPort(portName, a.Ctx, a.Name, a.PB, a.sourceOuts, InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) DriveOut(portName string) DrivenOut {
	return newDrivenOut(a.Out(portName))
}
