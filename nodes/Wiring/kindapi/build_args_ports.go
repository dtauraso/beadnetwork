package kindapi

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/wire/inport"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
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

func (a BuildArgs) DriveOut(portName string, slot int) DrivenOut {
	if a.driveSlotClaims != nil {
		if prior, claimed := a.driveSlotClaims[slot]; claimed {

			fmt.Fprintf(os.Stderr,
				"drive-stream collision: node %q slot %d already claimed by port %q; port %q "+
					"cannot also claim it — a DriveHeld goroutine for %q would share %q's dedicated "+
					"fd, which is exactly the two-goroutines-one-fd desync docs/interior-stream-"+
					"framing.md documents. %q's driven output stays unwired (drives nothing) instead. "+
					"Give it its own slot (Buffer.DriveSlotsPerNode).\n",
				a.name, slot, prior, portName, portName, prior, portName)
			return DrivenOut{}
		}
		a.driveSlotClaims[slot] = portName
	}
	out := portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, portwiring.AsEventSinkGetter(portwiring.NewDriveStreamGetter(a.name, slot, a.pb)))
	return newDrivenOut(out)
}
