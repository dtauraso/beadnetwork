package kindapi

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func (a BuildArgs) In(portName string) *wire.In {
	return portwiring.NewInPort(portName, a.ctx, a.name, a.pb, a.tr, a.getStream)
}

func (a BuildArgs) Out(portName string) *wire.Out {
	return portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, a.getStream)
}

func (a BuildArgs) Broadcast(portName string) wire.Broadcast {
	return portwiring.NewBroadcastPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, a.getStream)
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
	out := portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, portwiring.NewDriveStreamGetter(a.name, slot, a.pb))
	return newDrivenOut(out)
}
