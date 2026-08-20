package kindapi

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/portwiring"
	"github.com/dtauraso/wirefold/src/Bead/inport"
	"github.com/dtauraso/wirefold/src/Bead/outport"
)

func (a BuildArgs) In(portName string) *inport.In {
	return portwiring.NewInPort(portName, a.ctx, a.name, a.pb, a.tr, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Out(portName string) *outport.Out {
	return portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) Broadcast(portName string) outport.Broadcast {
	return portwiring.NewBroadcastPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, portwiring.InteriorEventSinkGetter(a.getEmitter))
}

func (a BuildArgs) DriveOut(portName string) DrivenOut {
	return newDrivenOut(a.Out(portName))
}
