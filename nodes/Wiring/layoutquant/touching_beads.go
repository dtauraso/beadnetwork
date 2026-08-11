// touching_beads.go — the live-mover adapter for beadcrud.DragTouchingBeads (lifted from
// nodes/Wiring/dispatch, docs/planning/movedispatch-decomposition.md §24 — pure move, no
// logic change): DragTouchingBeads used to hold a *MoveDispatch/*nodeGeometry
// back-reference itself; the pure resolution lives in nodes/Wiring/beadcrud and takes
// exactly the values it reads, so this file's only job is resolving those values from a
// live edgeMovers directory and calling it.

package layoutquant

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// DragTouchingBeads resolves nm's own incident-edge neighbours (via edgeMovers — the
// dispatch core's own directory, passed by the caller rather than reached through an
// unexported back-reference) into an edgeID->neighborID map, then calls
// beadcrud.DragTouchingBeads with nm's own id/kind/edgeIDs/partnerCenters/neighborKinds —
// see that function's doc comment for the full placement rule. Same values, same result,
// as when this function lived in the dispatch core itself.
func DragTouchingBeads(edgeMovers map[string]*edgemover.EdgeMover, nm *nodeactor.NodeGeometry, prevPos wire.Vec3) []beadcrud.TouchingBead {
	edgeIDs := nm.EdgeIDs()
	selfID := nm.ID()
	neighborOf := make(map[string]string, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		em, ok := edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.SrcID()
		if neighborID == selfID {
			neighborID = em.DstID()
		}
		neighborOf[edgeID] = neighborID
	}
	return beadcrud.DragTouchingBeads(selfID, nm.SelfKind(), edgeIDs, neighborOf, nm.PartnerCenters(), nm.NeighborKinds(), prevPos)
}
