package timeend

import (
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
)

func (a BuildArgs) In(portName string) *beadanimation.Receiver {
	a.mustDeclare(portName, PortIn)
	return NewInPort(portName, a.Ctx, a.Name, a.PB, InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Out(portName string) *beadanimation.Sender {
	a.mustDeclare(portName, PortOut)
	return NewOutPort(portName, a.Ctx, a.Name, a.PB, a.sourceOuts, InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Broadcast(portName string) beadanimation.Broadcast {
	a.mustDeclare(portName, PortBroadcast)
	return NewBroadcastPort(portName, a.Ctx, a.Name, a.PB, a.sourceOuts, InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) DriveOut(portName string) DrivenOut {
	return newDrivenOut(a.Out(portName))
}
