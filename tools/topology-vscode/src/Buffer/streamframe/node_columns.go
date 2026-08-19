package streamframe

import (
	B "github.com/dtauraso/wirefold/tools/topology-vscode/src/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/src/Buffer/colstream"
)

func WriteNodeColumns(c *colstream.ColumnSet, f NodeStreamFrame) {
	if c == nil {
		return
	}
	c.SetI32(B.ColStreamNodeNodeId, f.NodeID)
	c.SetI32(B.ColStreamNodeIndexR, f.IndexR)
	c.SetI32(B.ColStreamNodeIndexPhi, f.IndexPhi)
	c.SetI32(B.ColStreamNodeIndexTheta, f.IndexTheta)
	c.SetU8(B.ColStreamNodeHasPos, f.HasPos)
	c.SetF32(B.ColStreamNodeRadius, f.Radius)

	c.SetF32(B.ColStreamNodeNavTubeR, f.NavTubeR)
	c.SetF32(B.ColStreamNodePoleAnchorX, f.PoleAnchorX)
	c.SetF32(B.ColStreamNodePoleAnchorY, f.PoleAnchorY)
	c.SetF32(B.ColStreamNodePoleAnchorZ, f.PoleAnchorZ)

	c.SetF32(B.ColStreamNodeLabelAnchorX, f.LabelAnchorX)
	c.SetF32(B.ColStreamNodeLabelAnchorY, f.LabelAnchorY)
	c.SetF32(B.ColStreamNodeLabelAnchorZ, f.LabelAnchorZ)

	c.SetF32(B.ColStreamNodePolePhi, f.PolePhi)
	c.SetF32(B.ColStreamNodePoleTheta, f.PoleTheta)
	c.SetF32(B.ColStreamNodePoleRingR, f.PoleRingR)

	m := f.RingMatrix
	for i, col := range []int{
		B.ColStreamNodeRingM0, B.ColStreamNodeRingM1, B.ColStreamNodeRingM2, B.ColStreamNodeRingM3,
		B.ColStreamNodeRingM4, B.ColStreamNodeRingM5, B.ColStreamNodeRingM6, B.ColStreamNodeRingM7,
		B.ColStreamNodeRingM8, B.ColStreamNodeRingM9, B.ColStreamNodeRingM10, B.ColStreamNodeRingM11,
		B.ColStreamNodeRingM12, B.ColStreamNodeRingM13, B.ColStreamNodeRingM14, B.ColStreamNodeRingM15,
	} {
		c.SetF32(col, m[i])
	}

	c.SetF32(B.ColStreamNodeTopTiltVectorLen, f.TopTiltVectorLen)
	c.SetI32(B.ColStreamNodeTopTiltVectorIdx, f.TopTiltVectorIdx)

	c.SetU8(B.ColStreamNodeSelected, f.Selected)
	c.SetU8(B.ColStreamNodeKindId, f.KindID)
	c.SetU8(B.ColStreamNodeHovered, f.Hovered)
	c.SetU8(B.ColStreamNodeLatchedSel, f.LatchedSel)
	c.SetBytes(B.ColStreamNodeLabel, []byte(f.Label))

	c.SetU8(B.ColStreamNodeLatticePoints, f.LatticePoints)
	c.SetI32(B.ColStreamNodeRoundsToParallel, f.RoundsToParallel)
	c.SetI32(B.ColStreamNodeMsgsToParallel, f.MsgsToParallel)

	c.SetU8(B.ColStreamNodeDragRLocked, f.DragRLocked)
	c.SetU8(B.ColStreamNodeDragPhiLocked, f.DragPhiLocked)
	c.SetF32(B.ColStreamNodeDragThetaMax, f.DragThetaMax)
	c.SetU8(B.ColStreamNodeDragActive, f.DragActive)
	c.SetU8(B.ColStreamNodeHasKindRule, f.HasKindRule)
	c.SetU8(B.ColStreamNodeKindRuleActive, f.KindRuleActive)

	c.SetU8(B.ColStreamNodeSelfRLocked, f.SelfRLocked)
	c.SetU8(B.ColStreamNodeSelfPhiLocked, f.SelfPhiLocked)
	c.SetF32(B.ColStreamNodeSelfThetaMax, f.SelfThetaMax)
	c.SetU8(B.ColStreamNodeSelfActive, f.SelfActive)

	c.SetI32(B.ColStreamNodeRuleGroupId, f.RuleGroupID)
	c.SetI32(B.ColStreamNodeRuleGroupSize, f.RuleGroupSize)
}
