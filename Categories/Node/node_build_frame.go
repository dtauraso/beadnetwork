package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/AngleArcs"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
	"math"
)

type FrameInputs struct {
	Geom NodeGeom

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
	LatchedSel uint8





	TickCount int32

	StartVectors []polar.Polar
	StartPicked  []bool
	RefVectors   []Vec3
	Arcs         []AngleArcs.Angle
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
		DefaultLatticePoints: polarindex.FullTurnPhiIdx,
	})

	composedIdx := ComposedIndexOf(in.Geom)

	label := in.Geom.Label
	if label == "" {
		label = in.ID
	}


	return NodeFrameInput{
		Tick:             in.Tick,
		NodeRow:          in.Row,
		NodeID:           in.Row + 1,
		IndexR:           int32(composedIdx.R),
		IndexPhi:         int32(composedIdx.Phi),
		IndexTheta:       int32(composedIdx.Theta),
		HasPos:           boolU8(in.Geom.HasPos),
		Radius:           float32(NodeRadius(in.Geom.Kind)),
		NavTubeR:         float32(navTubeR(in.Geom.Kind)),
		PoleAnchorX:      float32(fg.Center.X),
		PoleAnchorY:      float32(fg.Center.Y),
		PoleAnchorZ:      float32(fg.Center.Z),
		LabelAnchorX:     float32(fg.LabelAnchor.X),
		LabelAnchorY:     float32(fg.LabelAnchor.Y),
		LabelAnchorZ:     float32(fg.LabelAnchor.Z),
		PoleRingR:        float32(PoleRingR()),
		PolePhi:          float32(fg.PolePhi),
		PoleTheta:        float32(fg.PoleTheta),
		RingMatrix:       fg.RingMatrix,
		BodyMatrix:       fg.BodyMatrix,
		TopTiltVectorLen: float32(fg.TopTiltVectorLen),
		TopTiltVectorIdx: fg.TopTiltVectorIdx,
		TiltArrows:       append(fg.TiltArrows, append(startArrows(in.Geom, fg.Center, in.StartVectors, in.StartPicked), refArrows(in.Geom, fg.Center, in.RefVectors)...)...),
		TickMarks:        TickMarks(fg.Center, in.Geom.Kind, in.TickCount),
		AngleArcs:        AngleArcs.Build(in.Selected != 0, fg.Center, NodeTorusOuterR(in.Geom.Kind), in.Arcs),
		Selected:         in.Selected,
		KindID:           in.KindID,
		LatchedSel:       in.LatchedSel,
		LatticePoints:    uint8(fg.LatticePoints),
		Label:            label,
	}
}

func navTubeR(kind string) float64 {
	return math.Max(0.5, NodeRadius(kind)*0.08)
}

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
