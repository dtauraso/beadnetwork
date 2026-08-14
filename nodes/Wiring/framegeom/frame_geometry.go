package framegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

type FrameGeometryInputs struct {
	Geom           nodegeom.NodeGeom
	UpAxis         bool
	CoplanarEdges  bool
	PartnerCenters map[string]vec3

	TopTiltVectorThetaIdx  int32
	BottomThetaIdx         int32
	NormalThetaIdx         int32
	ReceivedVectorThetaIdx int32
	ReceivedVectorSet      bool
	LatticePoints          int32
	DefaultLatticePoints   int32
}

type FrameGeometryOutputs struct {
	Center  vec3
	SphereR float64

	PolePhi, PoleTheta         float64
	RingAxisPhi, RingAxisTheta float64

	LatticePoints int32

	TopTiltVectorLen      float64
	TopTiltVectorTheta    float64
	BottomTiltVectorTheta float64
	CoplanarNormalTheta   float64

	ReceivedVectorLen   float64
	ReceivedVectorTheta float64
}

func DeriveFrameGeometry(in FrameGeometryInputs) FrameGeometryOutputs {
	out := FrameGeometryOutputs{
		Center:  nodegeom.NodeWorldPos(in.Geom),
		SphereR: nodegeom.EffectiveRadius(in.Geom),
	}

	out.PolePhi, out.PoleTheta = polar.WorldAxisPole()

	out.RingAxisPhi, out.RingAxisTheta = TorusDefaultAxisAngles()

	if in.UpAxis && in.Geom.HasPos && len(in.PartnerCenters) == 1 {
		for _, partner := range in.PartnerCenters {
			if t, p, ok := UprightRingAxis(nodegeom.NodeWorldPos(in.Geom), partner); ok {
				out.RingAxisPhi, out.RingAxisTheta = t, p
			}
		}
		out.TopTiltVectorLen = nodegeom.NodeRadius(in.Geom.Kind)
	} else if in.CoplanarEdges && in.Geom.HasPos && len(in.PartnerCenters) == 1 {
		for _, partner := range in.PartnerCenters {
			if t, p, ok := PoleContainingEdge(out.PolePhi, out.PoleTheta, nodegeom.NodeWorldPos(in.Geom), partner); ok {
				out.RingAxisPhi, out.RingAxisTheta = t, p
			}
		}
	}

	points := in.LatticePoints
	if points == 0 {
		points = in.DefaultLatticePoints
	}
	out.LatticePoints = points
	latticeThetaStep := 2 * math.Pi / float64(points)

	out.TopTiltVectorTheta = float64(in.TopTiltVectorThetaIdx) * latticeThetaStep
	out.BottomTiltVectorTheta = float64(in.BottomThetaIdx) * latticeThetaStep
	out.CoplanarNormalTheta = float64(in.NormalThetaIdx) * latticeThetaStep

	if in.ReceivedVectorSet {
		out.ReceivedVectorLen = nodegeom.NodeRadius(in.Geom.Kind)
		out.ReceivedVectorTheta = float64(in.ReceivedVectorThetaIdx) * latticeThetaStep
	}

	return out
}
