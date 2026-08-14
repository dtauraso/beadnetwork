package edgemover

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

func (m *EdgeMover) SetPersistRoot(root string) { m.persistRoot = root }

func (m *EdgeMover) Delta() polar.Polar { return m.d }

func (m *EdgeMover) SetDelta(d polar.Polar) { m.d = d }

func (m *EdgeMover) updateDeltaFromEndpoints() {
	if !m.srcGeom.HasPos || !m.dstGeom.HasPos {
		return
	}
	m.d = polar.Between(
		polar.Cart2polar(nodegeom.NodeWorldPos(m.srcGeom).Sub(m.srcGeom.SceneCenter)),
		polar.Cart2polar(nodegeom.NodeWorldPos(m.dstGeom).Sub(m.dstGeom.SceneCenter)),
	)
}

func (m *EdgeMover) persistDelta() {
	if m.persistRoot == "" {
		return
	}
	if err := edgefile.WriteEdgeDelta(m.persistRoot, m.srcID, m.edgeID, m.d); err != nil {
		jsonpersist.LogPersistErr("edge_delta_persist", fmt.Sprintf("%s->%s", m.srcID, m.dstID), err)
	}
}
