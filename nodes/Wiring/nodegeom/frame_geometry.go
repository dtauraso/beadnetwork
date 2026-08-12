package nodegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
)

type FrameGeometryInputs struct {
	Geom           NodeGeom
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

	PoleTheta, PolePhi         float64
	RingAxisTheta, RingAxisPhi float64

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
		Center:  NodeWorldPos(in.Geom),
		SphereR: EffectiveRadius(in.Geom),
	}

	if in.Geom.HasPos {
		out.PoleTheta, out.PolePhi = geom.InwardPole(in.Geom.ScenePolar)
	}

	out.RingAxisTheta, out.RingAxisPhi = TorusDefaultAxisAngles()

	if in.UpAxis && in.Geom.HasPos && len(in.PartnerCenters) == 1 {
		for _, partner := range in.PartnerCenters {
			if t, p, ok := UprightRingAxis(NodeWorldPos(in.Geom), partner); ok {
				out.RingAxisTheta, out.RingAxisPhi = t, p
			}
		}
		out.TopTiltVectorLen = NodeRadius(in.Geom.Kind)
	} else if in.CoplanarEdges && in.Geom.HasPos && len(in.PartnerCenters) == 1 {
		for _, partner := range in.PartnerCenters {
			if t, p, ok := PoleContainingEdge(out.PoleTheta, out.PolePhi, NodeWorldPos(in.Geom), partner); ok {
				out.RingAxisTheta, out.RingAxisPhi = t, p
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
		out.ReceivedVectorLen = NodeRadius(in.Geom.Kind)
		out.ReceivedVectorTheta = float64(in.ReceivedVectorThetaIdx) * latticeThetaStep
	}

	return out
}
