package nodeframe

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"math"
)

type FrameInputs struct {
	Geom nodegeom.NodeGeom

	Row    int32
	KindID uint8
	Tick   uint32
	ID     string

	UpAxis        bool
	CoplanarEdges bool

	TopTiltVectorPhiIdx  int32
	BottomPhiIdx         int32
	NormalPhiIdx         int32
	ReceivedVectorPhiIdx int32
	ReceivedVectorSet    bool
	LatticePoints        int32

	Selected   uint8
	Hovered    uint8
	LatchedSel uint8

	RoundsToParallel int32
	MsgsToParallel   int32

	DragRule *PolarRulesPanel.DragRule
	SelfRule *PolarRulesPanel.DragRule

	DragActive     bool
	SelfActive     bool
	KindRuleActive bool
	HasKindRule    bool

	RuleGroupID   int32
	RuleGroupSize int32

	ChannelVectors []ChannelVectors.ChannelVector
}

func BuildFrame(in FrameInputs) NodeFrameInput {
	fg := DeriveFrameGeometry(FrameGeometryInputs{
		Geom:                 in.Geom,
		UpAxis:               in.UpAxis,
		CoplanarEdges:        in.CoplanarEdges,
		TopTiltVectorPhiIdx:  in.TopTiltVectorPhiIdx,
		BottomPhiIdx:         in.BottomPhiIdx,
		NormalPhiIdx:         in.NormalPhiIdx,
		ReceivedVectorPhiIdx: in.ReceivedVectorPhiIdx,
		ReceivedVectorSet:    in.ReceivedVectorSet,
		LatticePoints:        in.LatticePoints,
		DefaultLatticePoints: TiltPanel.FullTurnPhiIdx,
	})

	composedIdx := nodegeom.ComposedIndexOf(in.Geom)

	label := in.Geom.Label
	if label == "" {
		label = in.ID
	}

	dragRLocked, dragPhiLocked, dragThetaMax := lockedAxes(in.DragRule)
	selfRLocked, selfPhiLocked, selfThetaMax := lockedAxes(in.SelfRule)

	return NodeFrameInput{
		Tick:             in.Tick,
		NodeRow:          in.Row,
		NodeID:           in.Row + 1,
		IndexR:           int32(composedIdx.R),
		IndexPhi:         int32(composedIdx.Phi),
		IndexTheta:       int32(composedIdx.Theta),
		HasPos:           boolU8(in.Geom.HasPos),
		Radius:           float32(nodegeom.NodeRadius(in.Geom.Kind)),
		NavTubeR:         float32(navTubeR(in.Geom.Kind)),
		PoleAnchorX:      float32(fg.Center.X),
		PoleAnchorY:      float32(fg.Center.Y),
		PoleAnchorZ:      float32(fg.Center.Z),
		LabelAnchorX:     float32(fg.LabelAnchor.X),
		LabelAnchorY:     float32(fg.LabelAnchor.Y),
		LabelAnchorZ:     float32(fg.LabelAnchor.Z),
		PoleRingR:        float32(nodegeom.PoleRingR()),
		PolePhi:          float32(fg.PolePhi),
		PoleTheta:        float32(fg.PoleTheta),
		RingMatrix:       fg.RingMatrix,
		TopTiltVectorLen: float32(fg.TopTiltVectorLen),
		TopTiltVectorIdx: fg.TopTiltVectorIdx,
		TiltArrows:       fg.TiltArrows,
		ChannelVectors:   in.ChannelVectors,
		Selected:         in.Selected,
		KindID:           in.KindID,
		Hovered:          in.Hovered,
		LatchedSel:       in.LatchedSel,
		LatticePoints:    uint8(fg.LatticePoints),
		RoundsToParallel: in.RoundsToParallel,
		MsgsToParallel:   in.MsgsToParallel,
		HasKindRule:      boolU8(in.HasKindRule),
		KindRuleActive:   boolU8(in.KindRuleActive),
		SelfRLocked:      selfRLocked,
		SelfPhiLocked:    selfPhiLocked,
		SelfThetaMax:     selfThetaMax,
		SelfActive:       boolU8(in.SelfActive),
		RuleGroupID:      in.RuleGroupID,
		RuleGroupSize:    in.RuleGroupSize,
		DragRLocked:      dragRLocked,
		DragPhiLocked:    dragPhiLocked,
		DragThetaMax:     dragThetaMax,
		DragActive:       boolU8(in.DragActive),
		Label:            label,
	}
}

func lockedAxes(rule *PolarRulesPanel.DragRule) (rLocked, phiLocked uint8, thetaMax float32) {
	thetaMax = -1
	if rule == nil {
		return 0, 0, thetaMax
	}
	if rule.R != nil {
		rLocked = 1
	}
	if rule.Phi != nil {
		phiLocked = 1
	}
	if rule.MaxTheta != nil {
		thetaMax = float32(*rule.MaxTheta)
	}
	return rLocked, phiLocked, thetaMax
}

func navTubeR(kind string) float64 {
	return math.Max(0.5, nodegeom.NodeRadius(kind)*0.08)
}

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
