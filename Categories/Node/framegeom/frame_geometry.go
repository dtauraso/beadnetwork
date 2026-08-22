package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
)

type FrameGeometryInputs struct {
	Geom          nodegeom.NodeGeom
	UpAxis        bool
	CoplanarEdges bool

	TopTiltVectorPhiIdx  int32
	BottomPhiIdx         int32
	NormalPhiIdx         int32
	ReceivedVectorPhiIdx int32
	ReceivedVectorSet    bool
	LatticePoints        int32
	DefaultLatticePoints int32
}

type FrameGeometryOutputs struct {
	Center Vec3

	PolePhi, PoleTheta float64

	RingMatrix [16]float32

	LatticePoints int32

	LabelAnchor Vec3

	TopTiltVectorLen float64

	TopTiltVectorIdx int32

	TopTiltVectorPhi    float64
	BottomTiltVectorPhi float64
	CoplanarNormalPhi   float64

	ReceivedVectorLen float64
	ReceivedVectorPhi float64

	TiltArrows []TiltArrow
}

func DeriveFrameGeometry(in FrameGeometryInputs) FrameGeometryOutputs {
	out := FrameGeometryOutputs{
		Center: Vec3(nodegeom.NodeWorldPos(in.Geom)),
	}

	out.PolePhi, out.PoleTheta = polar.WorldAxisPole()

	ringAxisPhi, ringAxisTheta := TorusDefaultAxisAngles()
	nodeR := nodegeom.NodeRadius(in.Geom.Kind)
	out.RingMatrix = RingInstanceMatrixColumnMajor(out.Center, nodeR, ringAxisPhi, ringAxisTheta)

	out.LabelAnchor = out.Center.Add(Vec3{Y: nodeR})

	if in.UpAxis && in.Geom.HasPos {
		out.TopTiltVectorLen = nodegeom.NodeRadius(in.Geom.Kind)
	}

	points := in.LatticePoints
	if points == 0 {
		points = in.DefaultLatticePoints
	}
	out.LatticePoints = points
	latticePhiStep := 2 * math.Pi / float64(points)

	out.TopTiltVectorIdx = in.TopTiltVectorPhiIdx
	out.TopTiltVectorPhi = float64(in.TopTiltVectorPhiIdx) * latticePhiStep
	out.BottomTiltVectorPhi = float64(in.BottomPhiIdx) * latticePhiStep
	out.CoplanarNormalPhi = float64(in.NormalPhiIdx) * latticePhiStep

	if in.ReceivedVectorSet {
		out.ReceivedVectorLen = nodegeom.NodeRadius(in.Geom.Kind)
		out.ReceivedVectorPhi = float64(in.ReceivedVectorPhiIdx) * latticePhiStep
	}

	if out.TopTiltVectorLen > 0 {
		out.TiltArrows = append(out.TiltArrows,
			ArrowMatrices(out.Center, out.TopTiltVectorLen, out.TopTiltVectorPhi, false),
			ArrowMatrices(out.Center, out.TopTiltVectorLen, out.BottomTiltVectorPhi, false),
			ArrowMatrices(out.Center, out.TopTiltVectorLen, out.CoplanarNormalPhi, false),
		)
	}
	if out.ReceivedVectorLen > 0 {
		out.TiltArrows = append(out.TiltArrows,
			ArrowMatrices(out.Center, out.ReceivedVectorLen, out.ReceivedVectorPhi, true),
		)
	}

	return out
}
