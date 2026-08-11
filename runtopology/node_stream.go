package runtopology

import (
	"fmt"
	"os"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	B "github.com/dtauraso/wirefold/Buffer"
	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
)

// wireNodeStreams wires the two per-node dedicated streams (NODE and INTERIOR) plus the
// per-drive-goroutine one, and reports the "all three required together" asymmetry first.
// The prose below is the record of the silent-skip failure mode this phase exists to make
// loud — it travels with this code.
//
// The two per-node dedicated streams (memory/feedback_no_single_writer_bridge.md):
// NODE (geometry+ports+label, written by each nodeMover) and INTERIOR (interior
// beads, written by each node's OWN Update goroutine — the SECOND emitting goroutine
// per node). Both require the SAME "node" AND "interior" WIREFOLD_STREAM_FDS entries
// (a node stream with no interior counterpart, or vice versa, would leave one of the
// two goroutines with nowhere fresh to write while the other has one — so both are
// required together).
//
// The "both required together" rule above is enforced by a silent skip: one entry
// present without the other leaves BOTH streams unwired and says nothing. Same class
// as the edge case, and harder to spot because the half that IS wired looks healthy.
// "drive" joins "node"/"interior" as a THIRD entry now required in lockstep: it is
// the per-gatecommon.DriveHeld-goroutine fd (Buffer.StreamKindDrive,
// docs/investigations/interior-stream-framing.md's fix) — a node with a DriveHeld drive goroutine
// needs it exactly as much as it needs "interior", or that goroutine falls back to
// writing nothing (a quieter failure than the pre-fix framing desync, but still a
// silent one) rather than sharing the node's own interior fd (the original bug).
func wireNodeStreams(streamFDs SF.StreamFDs, md *W.MoveDispatch) {
	_, nodeFDsWired := streamFDs[SF.StreamKindNode]
	_, interiorFDsWired := streamFDs[SF.StreamKindInterior]
	_, driveFDsWired := streamFDs[SF.StreamKindDrive]
	if nodeFDsWired != interiorFDsWired || nodeFDsWired != driveFDsWired {
		fmt.Fprintf(os.Stderr,
			"stream-fd mismatch: WIREFOLD_STREAM_FDS carries %q=%t, %q=%t, %q=%t; all three are "+
				"required together, so ALL THREE per-node streams stay unwired and node geometry/"+
				"interior beads/drive-goroutine sends will not be drawn.\n",
			SF.StreamKindNode, nodeFDsWired, SF.StreamKindInterior, interiorFDsWired, SF.StreamKindDrive, driveFDsWired)
	}
	if nodeBase, ok := streamFDs[SF.StreamKindNode]; ok {
		if interiorBase, ok2 := streamFDs[SF.StreamKindInterior]; ok2 {
			driveBase, driveWired := streamFDs[SF.StreamKindDrive]
			if !driveWired {
				driveBase = 0
			}
			// Selection/hover/abc-drag/kind are no longer injected lookups: each
			// nodeMover owns its OWN selected/hovered/latchedSel bits, set via
			// moveMsgKindSelect/Hover/Latched messages the gesture goroutine sends (or,
			// for kindID, resolved once here at construction).
			// kindIDFor resolves a node's static load-time kind string to its NODE_DEFS
			// index (Buffer.NodeKindID) — injected so Wiring stays Buffer-independent.
			md.SetNodeStreams(nodeBase, interiorBase, driveBase, driveWired,
				md.RT.NodeRowFor,
				// Field-by-NAME across the Buffer-independence seam, same shape as
				// toStreamEvents above: Wiring.NodeFrameInput → Buffer.NodeStreamFrame.
				// Two structs, one per side, so a transposition is a wrong field name
				// rather than a silently wrong scene.
				func(f nodeactor.NodeFrameInput) []byte {
					return SF.BuildNodeStreamFrame(SF.NodeStreamFrame{
						Tick:                  f.Tick,
						NodeRow:               f.NodeRow,
						NodeID:                f.NodeID,
						CX:                    f.CX,
						CY:                    f.CY,
						CZ:                    f.CZ,
						Radius:                f.Radius,
						SphereR:               f.SphereR,
						VRX:                   f.VRX,
						VRY:                   f.VRY,
						VRZ:                   f.VRZ,
						FRX:                   f.FRX,
						FRY:                   f.FRY,
						FRZ:                   f.FRZ,
						PoleTheta:             f.PoleTheta,
						PolePhi:               f.PolePhi,
						RingAxisTheta:         f.RingAxisTheta,
						RingAxisPhi:           f.RingAxisPhi,
						TopTiltVectorLen:      f.TopTiltVectorLen,
						TopTiltVectorTheta:    f.TopTiltVectorTheta,
						BottomTiltVectorTheta: f.BottomTiltVectorTheta,
						CoplanarNormalTheta:   f.CoplanarNormalTheta,
						ReceivedVectorLen:     f.ReceivedVectorLen,
						ReceivedVectorTheta:   f.ReceivedVectorTheta,
						Selected:              f.Selected,
						KindID:                f.KindID,
						Hovered:               f.Hovered,
						LatchedSel:            f.LatchedSel,
						LatticePoints:         f.LatticePoints,
						RoundsToParallel:      f.RoundsToParallel,
						MsgsToParallel:        f.MsgsToParallel,
						Label:                 f.Label,
						ChainBeadOX:           f.ChainBeadOX,
						ChainBeadOY:           f.ChainBeadOY,
						ChainBeadOZ:           f.ChainBeadOZ,
						ChainBeadLit:          f.ChainBeadLit,
						ChainBeadLitValue:     f.ChainBeadLitValue,
						Events:                toStreamEvents(f.Events),
					})
				},
				func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte {
					return SF.BuildInteriorStreamFrame(tick, present, value, ox, oy, oz, toStreamEvents(events))
				},
				B.NodeKindID)
		}
	}
}
