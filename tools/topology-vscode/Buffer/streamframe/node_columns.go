package streamframe

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/colstream"
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

	ring := make([]byte, 0, 16*4)
	for _, v := range f.RingMatrix {
		ring = binary.LittleEndian.AppendUint32(ring, math.Float32bits(v))
	}
	c.SetBytes(B.ColStreamNodeRingMatrix, ring)

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
