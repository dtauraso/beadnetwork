package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
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
	Center  vec3
	SphereR float64

	PolePhi, PoleTheta         float64
	RingAxisPhi, RingAxisTheta float64

	RingMatrix [16]float32

	LatticePoints int32

	TopTiltVectorLen float64

	TopTiltVectorIdx int32

	TopTiltVectorPhi    float64
	BottomTiltVectorPhi float64
	CoplanarNormalPhi   float64

	ReceivedVectorLen float64
	ReceivedVectorPhi float64
}

func DeriveFrameGeometry(in FrameGeometryInputs) FrameGeometryOutputs {
	out := FrameGeometryOutputs{
		Center:  nodegeom.NodeWorldPos(in.Geom),
		SphereR: nodegeom.EffectiveRadius(in.Geom),
	}

	out.PolePhi, out.PoleTheta = polar.WorldAxisPole()

	out.RingAxisPhi, out.RingAxisTheta = TorusDefaultAxisAngles()

	nodeR := nodegeom.NodeRadius(in.Geom.Kind)
	out.RingMatrix = RingInstanceMatrixColumnMajor(out.Center, nodeR, out.RingAxisPhi, out.RingAxisTheta)

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

	return out
}
