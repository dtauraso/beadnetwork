package runtopology

import (
	"fmt"
	"os"

	T "github.com/dtauraso/wirefold/src/Trace"

	NodeKind "github.com/dtauraso/wirefold/src/Node"

	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	EdgeB "github.com/dtauraso/wirefold/src/Node/Edge"
	"github.com/dtauraso/wirefold/src/Node/framegeom"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodeframe"
	BeadB "github.com/dtauraso/wirefold/src/Ring/Bead"
	TiltB "github.com/dtauraso/wirefold/src/Scene/TiltVectors"
	VecB "github.com/dtauraso/wirefold/src/Scene/Vectors"
)

func wireNodeStreams(md *W.MoveDispatch, sceneRoot string) {
	beadValues := make([]*BeadB.ValueWriter, len(md.RT.NodeRowTable))
	nodeValues := make([]*NodeKind.ValueWriter, len(md.RT.NodeRowTable))
	tiltValues := make([]*TiltB.ValueWriter, len(md.RT.NodeRowTable))
	vecValues := make([]*VecB.ValueWriter, len(md.RT.NodeRowTable))
	for row := range beadValues {
		beadValues[row] = BeadB.NewValueWriter(sceneRoot, row)
		nodeValues[row] = NodeKind.NewValueWriter(sceneRoot, row)
		tiltValues[row] = TiltB.NewValueWriter(sceneRoot, row)
		vecValues[row] = VecB.NewValueWriter(sceneRoot, row)
	}

	md.Sw.SetNodeStreams(md.GS.NodeSeeds, md.MR.NodeGeoms(), sceneRoot,
		func(tick uint32, nodeRow int32, beads []EdgeB.EdgeBead, events []T.RowEvent) {
			if int(nodeRow) < len(beadValues) {
				if err := EdgeB.WriteEdgeBeadValues(beadValues[nodeRow], beads); err != nil {
					fmt.Fprintf(os.Stderr, "bead values write (node row %d): %v\n", nodeRow, err)
				}
			}
			T.NewLog(T.OwnerBead, nodeRow).Append(events)
		},
		md.RT.NodeRowFor,

		func(f nodeframe.NodeFrameInput) {
			frame := NodeKind.NodeState{
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

			if int(f.NodeRow) < len(tiltValues) {
				if err := NodeKind.WriteNodeValues(nodeValues[f.NodeRow], frame); err != nil {
					fmt.Fprintf(os.Stderr, "node values write (node row %d): %v\n", f.NodeRow, err)
				}
				if err := TiltB.WriteTiltArrowValues(tiltValues[f.NodeRow], frame.TiltArrows); err != nil {
					fmt.Fprintf(os.Stderr, "tilt arrow values write (node row %d): %v\n", f.NodeRow, err)
				}
				if err := VecB.WriteChannelVectorValues(vecValues[f.NodeRow], frame.ChannelVectors); err != nil {
					fmt.Fprintf(os.Stderr, "channel vector values write (node row %d): %v\n", f.NodeRow, err)
				}
			}
			T.NewLog(T.OwnerNode, f.NodeRow).Append(f.Events)
		},
		func(tick uint32, nodeRow int32, events []T.RowEvent) {
			T.NewLog(T.OwnerInterior, nodeRow).Append(events)
		},
		NodeKind.NodeKindID)
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
