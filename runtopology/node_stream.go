package runtopology

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	B "github.com/dtauraso/wirefold/Buffer"
	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
)

func wireNodeStreams(streamFDs SF.StreamFDs, md *W.MoveDispatch) {
	_, nodeFDsWired := streamFDs[SF.StreamKindNode]
	_, interiorFDsWired := streamFDs[SF.StreamKindInterior]
	_, beadFDsWired := streamFDs[SF.StreamKindBead]
	if nodeFDsWired != interiorFDsWired || nodeFDsWired != beadFDsWired {
		fmt.Fprintf(os.Stderr,
			"stream-fd mismatch: WIREFOLD_STREAM_FDS carries %q=%t, %q=%t, %q=%t; all three are "+
				"required together, so the per-node streams stay unwired and node geometry/"+
				"interior beads/EDGE BEADS will not be drawn. If the editor was "+
				"open across a rebuild, run \"Developer: Reload Window\" — reopening the file restarts "+
				"only the webview, not the extension host that allocates these fds.\n",
			SF.StreamKindNode, nodeFDsWired, SF.StreamKindInterior, interiorFDsWired,
			SF.StreamKindBead, beadFDsWired)
	}
	if nodeBase, ok := streamFDs[SF.StreamKindNode]; ok {
		if interiorBase, ok2 := streamFDs[SF.StreamKindInterior]; ok2 {
			beadBase, beadWired := streamFDs[SF.StreamKindBead]
			if !beadWired {
				beadBase = 0
			}

			md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), nodeBase, interiorBase,
				beadBase, beadWired,
				func(tick uint32, nodeRow int32, beads []SF.EdgeBead, events []rowevent.RowEvent) []byte {
					return SF.BuildBeadStreamFrame(tick, nodeRow, beads, toStreamEvents(events))
				},
				md.RT.NodeRowFor,

				func(f nodeframe.NodeFrameInput) []byte {
					return SF.BuildNodeStreamFrame(SF.NodeStreamFrame{
						Tick:                f.Tick,
						NodeRow:             f.NodeRow,
						NodeID:              f.NodeID,
						CX:                  f.CX,
						CY:                  f.CY,
						CZ:                  f.CZ,
						Radius:              f.Radius,
						VRX:                 f.VRX,
						VRY:                 f.VRY,
						VRZ:                 f.VRZ,
						FRX:                 f.FRX,
						FRY:                 f.FRY,
						FRZ:                 f.FRZ,
						PolePhi:             f.PolePhi,
						PoleTheta:           f.PoleTheta,
						RingAxisPhi:         f.RingAxisPhi,
						RingAxisTheta:       f.RingAxisTheta,
						RingMatrix:          f.RingMatrix,
						TopTiltVectorLen:    f.TopTiltVectorLen,
						TopTiltVectorIdx:    f.TopTiltVectorIdx,
						TopTiltVectorPhi:    f.TopTiltVectorPhi,
						BottomTiltVectorPhi: f.BottomTiltVectorPhi,
						CoplanarNormalPhi:   f.CoplanarNormalPhi,
						ReceivedVectorLen:   f.ReceivedVectorLen,
						ReceivedVectorPhi:   f.ReceivedVectorPhi,
						Selected:            f.Selected,
						KindID:              f.KindID,
						Hovered:             f.Hovered,
						LatchedSel:          f.LatchedSel,
						LatticePoints:       f.LatticePoints,
						RoundsToParallel:    f.RoundsToParallel,
						MsgsToParallel:      f.MsgsToParallel,
						DragRLocked:         f.DragRLocked,
						DragPhiLocked:       f.DragPhiLocked,
						DragThetaMax:        f.DragThetaMax,
						DragActive:          f.DragActive,
						HasKindRule:         f.HasKindRule,
						RuleGroupID:         f.RuleGroupID,
						RuleGroupSize:       f.RuleGroupSize,
						Label:               f.Label,
						Events:              toStreamEvents(f.Events),
					})
				},
				func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte {
					return SF.BuildInteriorStreamFrame(tick, present, value, ox, oy, oz, toStreamEvents(events))
				},
				B.NodeKindID)
		}
	}
}
