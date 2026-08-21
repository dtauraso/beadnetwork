package runtopology

import (
	"fmt"
	"os"

	NodeKind "github.com/dtauraso/wirefold/src/Node"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"

	B "github.com/dtauraso/wirefold/src/Buffer"

	"github.com/dtauraso/wirefold/src/Buffer/colstream"
	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	EdgeB "github.com/dtauraso/wirefold/src/Node/Edge"
	BeadB "github.com/dtauraso/wirefold/src/Ring/Bead"
	"github.com/dtauraso/wirefold/src/Node/framegeom"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodeframe"
	TiltB "github.com/dtauraso/wirefold/src/Scene/TiltVectors"
	VecB "github.com/dtauraso/wirefold/src/Scene/Vectors"
	SW "github.com/dtauraso/wirefold/src/runtopology/streamwire"
)

func wireNodeStreams(streamFDs SW.StreamFDs, md *W.MoveDispatch, sceneRoot string) {
	cols := SW.NewColumnStreams(streamFDs, len(md.RT.NodeRowTable), len(md.RT.EdgeRowTable))

	beadValues := make([]*BeadB.ValueWriter, len(md.RT.NodeRowTable))
	tiltValues := make([]*TiltB.ValueWriter, len(md.RT.NodeRowTable))
	vecValues := make([]*VecB.ValueWriter, len(md.RT.NodeRowTable))
	for row := range beadValues {
		beadValues[row] = BeadB.NewValueWriter(sceneRoot, row)
		tiltValues[row] = TiltB.NewValueWriter(sceneRoot, row)
		vecValues[row] = VecB.NewValueWriter(sceneRoot, row)
	}

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

			md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), sceneRoot, nodeBase, interiorBase,
				beadBase, beadWired,
				func(tick uint32, nodeRow int32, beads []EdgeB.EdgeBead, events []B.RowEvent) []byte {
					if int(nodeRow) < len(beadValues) {
						if err := EdgeB.WriteEdgeBeadValues(beadValues[nodeRow], beads); err != nil {
							fmt.Fprintf(os.Stderr, "bead values write (node row %d): %v\n", nodeRow, err)
						}
					}
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
					if int(f.NodeRow) < len(tiltValues) {
						if err := TiltB.WriteTiltArrowValues(tiltValues[f.NodeRow], frame.TiltArrows); err != nil {
							fmt.Fprintf(os.Stderr, "tilt arrow values write (node row %d): %v\n", f.NodeRow, err)
						}
						if err := VecB.WriteChannelVectorValues(vecValues[f.NodeRow], frame.ChannelVectors); err != nil {
							fmt.Fprintf(os.Stderr, "channel vector values write (node row %d): %v\n", f.NodeRow, err)
						}
					}
					return NodeKind.BuildNodeStreamFrame(frame)
				},
				func(tick uint32, events []B.RowEvent) []byte {
					return NodeKind.BuildInteriorStreamFrame(tick, events)
				},
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
