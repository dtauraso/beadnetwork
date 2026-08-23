package selectleft

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
)

func (a BuildArgs) In(portName string) *beadanimation.Receiver {
	a.mustDeclare(portName, portwiring.PortIn)
	return portwiring.NewInPort(portName, a.Ctx, a.Name, a.PB, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Out(portName string) *beadanimation.Sender {
	a.mustDeclare(portName, portwiring.PortOut)
	return portwiring.NewOutPort(portName, a.Ctx, a.Name, a.PB, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Broadcast(portName string) beadanimation.Broadcast {
	a.mustDeclare(portName, portwiring.PortBroadcast)
	return portwiring.NewBroadcastPort(portName, a.Ctx, a.Name, a.PB, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) DriveOut(portName string) DrivenOut {
	return newDrivenOut(a.Out(portName))
}
