// build_args_ports.go — BuildArgs methods that resolve a node's PORTS (In/Out/Broadcast/
// DriveOut). Split out of build_args.go, which keeps the BuildArgs struct itself,
// RegisterBuilder, and the trivial identity accessors (Name/Ctx) — see that file's header
// for why BuildArgs exists at all.

package dispatch

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// In resolves an input port by its SPEC name. Paced when the loader bound a wire to it,
// dead-end otherwise (same fallback reflectBuild's wireInPort applied).
func (a BuildArgs) In(portName string) *wire.In {
	return portwiring.NewInPort(portName, a.ctx, a.name, a.pb, a.tr, a.getStream)
}

// Out resolves a single output port by its SPEC name.
func (a BuildArgs) Out(portName string) *wire.Out {
	return portwiring.NewOutPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, a.getStream)
}

// Broadcast resolves a fan-out output port by its SPEC name.
func (a BuildArgs) Broadcast(portName string) wire.Broadcast {
	return portwiring.NewBroadcastPort(portName, a.ctx, a.name, a.pb, a.tr, a.sourceOuts, a.getStream)
}

// DriveOut resolves an output port that will be DRIVEN by its own gatecommon.DriveHeld
// goroutine (a SEPARATE goroutine from this node's own Update loop — Pulse/PulseLeft/
// PulseRight/holdflip's shape), instead of Out(). It routes the port's eventSink through
// a DEDICATED per-(node, slot) drive stream (newDriveStreamGetter, Buffer.StreamKindDrive)
// rather than this node's shared getStream — the fix for the framing desync documented in
// docs/investigations/interior-stream-framing.md: two goroutines (this node's Update loop and its
// DriveHeld goroutine) must never write the same *interiorStream/fd. slot distinguishes
// multiple DriveHeld outputs on ONE node (Pulse's Out=slot 0, OutFanout=slot 1 — see
// Buffer.DriveSlotsPerNode's doc comment for the current max) and must be a DIFFERENT
// value for each such call on the same node; passing the same slot to two driven outputs
// on one node would make them share a stream, reintroducing this exact bug. A plain
// (non-DriveHeld) Out — only ever written from this node's own Update goroutine — should
// keep using Out(), not DriveOut(): its writes already satisfy the single-writer
// invariant via the shared getStream, and giving it a drive slot would burn an fd for no
// reason.
func (a BuildArgs) DriveOut(portName string, slot int) DrivenOut {
	if a.driveSlotClaims != nil {
		if prior, claimed := a.driveSlotClaims[slot]; claimed {
			// Wiring-time failure, reported not panicked (main.go's own stream-fd-
			// mismatch posture — see its "stream-fd mismatch" Fprintf calls): this runs
			// during LoadTopology's single-threaded build phase, before this node's
			// Update or DriveHeld goroutines exist, so there is no crash-loop risk and
			// nothing to unwind. The SECOND claimant gets a dead DrivenOut (zero value:
			// nil-safe Wired/Paced/Steps/PlaceDrivenAt all degrade to "not driving"),
			// same fallback shape as every other absent-stream case in this file.
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
