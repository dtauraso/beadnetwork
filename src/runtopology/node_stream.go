package runtopology

import (
	"fmt"
	"os"

	NodeKind "github.com/dtauraso/wirefold/src/Node"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"

	EdgeB "github.com/dtauraso/wirefold/src/Node/Edge"
	W "github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/framegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/nodeframe"
	SW "github.com/dtauraso/wirefold/src/Node/Wiring/streamwire"
	TiltB "github.com/dtauraso/wirefold/src/Scene/TiltVectors"
	VecB "github.com/dtauraso/wirefold/src/Scene/Vectors"
	"github.com/dtauraso/wirefold/src/schema/buffer-layout/colstream"
)

func wireNodeStreams(streamFDs SW.StreamFDs, md *W.MoveDispatch) {
	cols := SW.NewColumnStreams(streamFDs, len(md.RT.NodeRowTable), len(md.RT.EdgeRowTable))

	nodeSets := map[int32]*colstream.ColumnSet{}
	nodeCols := func(row int32) *colstream.ColumnSet {
		if s, ok := nodeSets[row]; ok {
			return s
		}
		s := cols.NodeColumns(int(row))
		nodeSets[row] = s
		return s
	}
	_, nodeFDsWired := streamFDs[SW.StreamKindNode]
	_, interiorFDsWired := streamFDs[SW.StreamKindInterior]
	_, beadFDsWired := streamFDs[SW.StreamKindBead]
	if nodeFDsWired != interiorFDsWired || nodeFDsWired != beadFDsWired {
		fmt.Fprintf(os.Stderr,
			"stream-fd mismatch: WIREFOLD_STREAM_FDS carries %q=%t, %q=%t, %q=%t; all three are "+
				"required together, so the per-node streams stay unwired and node geometry/"+
				"interior beads/EDGE BEADS will not be drawn. If the editor was "+
				"open across a rebuild, run \"Developer: Reload Window\" — reopening the file restarts "+
				"only the webview, not the extension host that allocates these fds.\n",
			SW.StreamKindNode, nodeFDsWired, SW.StreamKindInterior, interiorFDsWired,
			SW.StreamKindBead, beadFDsWired)
	}
	if nodeBase, ok := streamFDs[SW.StreamKindNode]; ok {
		if interiorBase, ok2 := streamFDs[SW.StreamKindInterior]; ok2 {
			beadBase, beadWired := streamFDs[SW.StreamKindBead]
			if !beadWired {
				beadBase = 0
			}

			md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), nodeBase, interiorBase,
				beadBase, beadWired,
				func(tick uint32, nodeRow int32, beads []EdgeB.EdgeBead, events []B.RowEvent) []byte {
					EdgeB.WriteEdgeBeadColumns(nodeCols(nodeRow), beads)
					return beadanimation.BuildBeadStreamFrame(tick, nodeRow, events)
				},
				md.RT.NodeRowFor,

				func(f nodeframe.NodeFrameInput) []byte {
					frame := NodeKind.NodeStreamFrame{
						Tick:             f.Tick,
						NodeRow:          f.NodeRow,
						NodeID:           f.NodeID,
						IndexR:           f.IndexR,
						IndexPhi:         f.IndexPhi,
						IndexTheta:       f.IndexTheta,
						HasPos:           f.HasPos,
						Radius:           f.Radius,
						NavTubeR:         f.NavTubeR,
						PoleAnchorX:      f.PoleAnchorX,
						PoleAnchorY:      f.PoleAnchorY,
						PoleAnchorZ:      f.PoleAnchorZ,
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
						Events:           f.Events,
					}

					NodeKind.WriteNodeColumns(nodeCols(f.NodeRow), frame)
					TiltB.WriteTiltArrowColumns(nodeCols(f.NodeRow), frame.TiltArrows)
					VecB.WriteChannelVectorColumns(nodeCols(f.NodeRow), frame.ChannelVectors)
					return NodeKind.BuildNodeStreamFrame(frame)
				},
				func(tick uint32, events []B.RowEvent) []byte {
					return NodeKind.BuildInteriorStreamFrame(tick, events)
				},
				nodeCols,
				NodeKind.NodeKindID)
		}
	}
}

func toStreamTiltArrows(in []framegeom.TiltArrow) []TiltB.TiltArrow {
	if len(in) == 0 {
		return nil
	}
	out := make([]TiltB.TiltArrow, 0, len(in))
	for _, a := range in {
		var r uint8
		if a.Received {
			r = 1
		}
		out = append(out, TiltB.TiltArrow{Received: r, Shaft: a.Shaft, Head: a.Head})
	}
	return out
}
