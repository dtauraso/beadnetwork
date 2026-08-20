package beadcrud

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/src/Node/wire/lattice"
)

type TouchingBead struct {
	NeighborID string
	Source     vec3
	Centre     vec3
	AimDir     vec3
}

func DragTouchingBeads(nodeID, selfKind string, edgeIDs []string, neighborOf map[string]string, partnerCenters map[string]vec3, neighborKinds map[string]string, prevPos vec3) []TouchingBead {
	selfTorusR := nodegeom.NodeTorusOuterR(selfKind)
	out := make([]TouchingBead, 0, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		neighborID, ok := neighborOf[edgeID]
		if !ok {
			continue
		}
		neighborCenter, ok := partnerCenters[neighborID]
		if !ok {
			continue
		}
		dist, aimDir, ok := edgegeom.EdgeCenterDistAndDir(prevPos, neighborCenter)
		if !ok {
			continue
		}
		beadCentre := prevPos.Add(aimDir.Scale(selfTorusR + lattice.BeadTorusOuterR))

		neighborKind := neighborKinds[neighborID]
		count := edgegeom.EdgeStepCount(dist, neighborKind, selfKind)
		var beadSource vec3
		if count >= 2 {
			beadSource = beadCentre.Add(aimDir.Scale(lattice.BeadStepR))
		} else {
			beadSource = neighborCenter.Sub(aimDir.Scale(nodegeom.NodeTorusOuterR(neighborKind)))
		}
		out = append(out, TouchingBead{NeighborID: neighborID, Source: beadSource, Centre: beadCentre, AimDir: aimDir})
	}
	return out
}
