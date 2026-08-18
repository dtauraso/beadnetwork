package runtopology

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	SF "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/streamframe"
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
						Tick:             f.Tick,
						NodeRow:          f.NodeRow,
						NodeID:           f.NodeID,
						IndexR:           f.IndexR,
						IndexPhi:         f.IndexPhi,
						IndexTheta:       f.IndexTheta,
						HasPos:           f.HasPos,
						Radius:           f.Radius,
						LabelAnchorX:     f.LabelAnchorX,
						LabelAnchorY:     f.LabelAnchorY,
						LabelAnchorZ:     f.LabelAnchorZ,
						PolePhi:          f.PolePhi,
						PoleTheta:        f.PoleTheta,
						RingMatrix:       f.RingMatrix,
						TopTiltVectorLen: f.TopTiltVectorLen,
						TopTiltVectorIdx: f.TopTiltVectorIdx,
						TiltArrows:       toStreamTiltArrows(f.TiltArrows),
						ChannelVectors:   f.ChannelVectors,
						Selected:         f.Selected,
						KindID:           f.KindID,
						Hovered:          f.Hovered,
						LatchedSel:       f.LatchedSel,
						LatticePoints:    f.LatticePoints,
						RoundsToParallel: f.RoundsToParallel,
						MsgsToParallel:   f.MsgsToParallel,
						DragRLocked:      f.DragRLocked,
						DragPhiLocked:    f.DragPhiLocked,
						DragThetaMax:     f.DragThetaMax,
						DragActive:       f.DragActive,
						HasKindRule:      f.HasKindRule,
						KindRuleActive:   f.KindRuleActive,
						PoleRingR:        f.PoleRingR,
						SelfRLocked:      f.SelfRLocked,
						SelfPhiLocked:    f.SelfPhiLocked,
						SelfThetaMax:     f.SelfThetaMax,
						SelfActive:       f.SelfActive,
						RuleGroupID:      f.RuleGroupID,
						RuleGroupSize:    f.RuleGroupSize,
						Label:            f.Label,
						Events:           toStreamEvents(f.Events),
					})
				},
				func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte {
					return SF.BuildInteriorStreamFrame(tick, present, value, ox, oy, oz, toStreamEvents(events))
				},
				B.NodeKindID)
		}
	}
}

func toStreamTiltArrows(in []framegeom.TiltArrow) []SF.TiltArrow {
	if len(in) == 0 {
		return nil
	}
	out := make([]SF.TiltArrow, 0, len(in))
	for _, a := range in {
		var r uint8
		if a.Received {
			r = 1
		}
		out = append(out, SF.TiltArrow{Received: r, Shaft: a.Shaft, Head: a.Head})
	}
	return out
}
