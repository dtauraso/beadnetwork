package framegeom

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

type FrameGeometryInputs struct {
	Geom          nodegeom.NodeGeom
	UpAxis        bool
	CoplanarEdges bool

	PartnerDeltas []polar.Polar

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

func (in FrameGeometryInputs) partnerCenter(i int) vec3 {
	return nodegeom.NodeWorldPos(in.Geom).Add(polar.Polar2cart(in.PartnerDeltas[i]))
}

func DeriveFrameGeometry(in FrameGeometryInputs) FrameGeometryOutputs {
	out := FrameGeometryOutputs{
		Center:  nodegeom.NodeWorldPos(in.Geom),
		SphereR: nodegeom.EffectiveRadius(in.Geom),
	}

	out.PolePhi, out.PoleTheta = polar.WorldAxisPole()

	out.RingAxisPhi, out.RingAxisTheta = TorusDefaultAxisAngles()

	if in.UpAxis && in.Geom.HasPos && len(in.PartnerDeltas) == 0 {
		panic(fmt.Sprintf(
			"DeriveFrameGeometry: node %q is placed in an up-axis scene but holds no neighbour delta, "+
				"so its ring axis and tilt vector cannot be derived and would silently fall back to the "+
				"scene defaults. A node's delta to each neighbour is seeded at build from the edge files "+
				"it owns (build_move_dispatch's SetDeltaTo) and maintained by its own moves; reaching "+
				"here means that seeding did not happen for this node",
			in.Geom.Label))
	}

	if in.UpAxis && in.Geom.HasPos && len(in.PartnerDeltas) == 1 {
		if t, p, ok := UprightRingAxis(nodegeom.NodeWorldPos(in.Geom), in.partnerCenter(0)); ok {
			out.RingAxisPhi, out.RingAxisTheta = t, p
		}
		out.TopTiltVectorLen = nodegeom.NodeRadius(in.Geom.Kind)
	} else if in.CoplanarEdges && in.Geom.HasPos && len(in.PartnerDeltas) == 1 {
		if t, p, ok := PoleContainingEdge(out.PolePhi, out.PoleTheta, nodegeom.NodeWorldPos(in.Geom), in.partnerCenter(0)); ok {
			out.RingAxisPhi, out.RingAxisTheta = t, p
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
