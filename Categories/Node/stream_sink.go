package Node

import (
	"fmt"
	"os"

	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	BeadB "github.com/dtauraso/beadnetwork/Categories/Ring/Bead"
)

type Sinks struct {
	Beads func(tick uint32, nodeRow int32, beads []edge.EdgeBead)
	Node  NodeFrameBuilder
}

func NewSinks(sceneRoot string, rows int) Sinks {
	beadValues := make([]*BeadB.ValueWriter, rows)
	nodeValues := make([]*ValueWriter, rows)
	for row := range beadValues {
		beadValues[row] = BeadB.NewValueWriter(sceneRoot, row)
		nodeValues[row] = NewValueWriter(sceneRoot, row)
	}

	return Sinks{
		Beads: func(tick uint32, nodeRow int32, beads []edge.EdgeBead) {
			if int(nodeRow) < len(beadValues) {
				if err := edge.WriteEdgeBeadValues(beadValues[nodeRow], beads); err != nil {
					fmt.Fprintf(os.Stderr, "bead values write (node row %d): %v\n", nodeRow, err)
				}
			}
		},

		Node: func(f NodeFrameInput) {
			frame := nodeStateFrom(f)
			if int(f.NodeRow) < len(nodeValues) {
				if err := WriteNodeBlock(nodeValues[f.NodeRow], frame); err != nil {
					fmt.Fprintf(os.Stderr, "node block write (node row %d): %v\n", f.NodeRow, err)
				}
			}
		},
	}
}

func nodeStateFrom(f NodeFrameInput) NodeState {
	return NodeState{
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
		BodyMatrix:       f.BodyMatrix,
		TopTiltVectorLen: f.TopTiltVectorLen,
		TopTiltVectorIdx: f.TopTiltVectorIdx,
		TiltArrows:       f.TiltArrows,
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
	}
}
