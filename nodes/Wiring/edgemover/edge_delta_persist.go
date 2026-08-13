package edgemover

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// SetPersistRoot arms this edge to write its own file. Until it is called the
// edge keeps D in memory only, which is what a headless run wants.
func (m *EdgeMover) SetPersistRoot(root string) { m.persistRoot = root }

// Delta is this edge's own vector, source to target.
func (m *EdgeMover) Delta() polar.Polar { return m.d }

// SetDelta seeds it at build, from the edge's own file.
func (m *EdgeMover) SetDelta(d polar.Polar) { m.d = d }

// updateDeltaFromEndpoints brings this edge's own vector up to date with a move
// it has just been TOLD about.
//
// The edgeMover is the ONLY writer of nodes/<source>/edges/<label>.json, and it
// already hears from both endpoints (srcIn and dstIn), so a target that moves
// tells this goroutine rather than writing a file it does not own. That is why
// no new message shape was needed for the target-moved case: it was already
// arriving.
//
// This is the edge updating ITS OWN state from what it was told, not a
// derivation of state someone else owns — the distinction the loader lost when
// it recomputed D on every load and then asserted the result.
func (m *EdgeMover) updateDeltaFromEndpoints() {
	if !m.srcGeom.HasPos || !m.dstGeom.HasPos {
		return
	}
	m.d = polar.Between(
		polar.Cart2polar(nodegeom.NodeWorldPos(m.srcGeom).Sub(m.srcGeom.SceneCenter)),
		polar.Cart2polar(nodegeom.NodeWorldPos(m.dstGeom).Sub(m.dstGeom.SceneCenter)),
	)
}

// persistDelta writes the vector this edge holds into the file it owns.
func (m *EdgeMover) persistDelta() {
	if m.persistRoot == "" {
		return
	}
	if err := edgefile.WriteEdgeDelta(m.persistRoot, m.srcID, m.edgeID, m.d); err != nil {
		jsonpersist.LogPersistErr("edge_delta_persist", fmt.Sprintf("%s->%s", m.srcID, m.dstID), err)
	}
}
