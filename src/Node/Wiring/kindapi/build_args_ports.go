package kindapi

import (
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Wiring/portwiring"
)

func (a BuildArgs) In(portName string) *beadanimation.Receiver {
	return portwiring.NewInPort(portName, a.ctx, a.name, a.pb, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Out(portName string) *beadanimation.Sender {
	return portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Broadcast(portName string) beadanimation.Broadcast {
	return portwiring.NewBroadcastPort(portName, a.ctx, a.name, a.pb, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) DriveOut(portName string) DrivenOut {
	return newDrivenOut(a.Out(portName))
}
