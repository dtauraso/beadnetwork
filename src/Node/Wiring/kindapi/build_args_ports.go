package kindapi

import (
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Wiring/portwiring"
)

func (a BuildArgs) In(portName string) *beadanimation.Receiver {
	a.mustDeclare(portName, portwiring.PortIn)
	return portwiring.NewInPort(portName, a.ctx, a.name, a.pb, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Out(portName string) *beadanimation.Sender {
	a.mustDeclare(portName, portwiring.PortOut)
	return portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Broadcast(portName string) beadanimation.Broadcast {
	a.mustDeclare(portName, portwiring.PortBroadcast)
	return portwiring.NewBroadcastPort(portName, a.ctx, a.name, a.pb, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) DriveOut(portName string) DrivenOut {
	return newDrivenOut(a.Out(portName))
}
