package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodefiles"
)

func (m *NodeGeometry) takeOrbitPhiToggle() {
	rule := m.topo.OrbitRule()
	var next polar.OrbitRule
	if rule != nil {
		next = *rule
	}
	if next.Phi != nil {
		next.Phi = nil
	} else {
		zero := 0.0
		next.Phi = &zero
	}
	m.SetOrbitRule(&next)
	m.persistOrbitRule()
	m.BroadcastRule()
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) takeOrbitMaxTheta(maxTheta *float64) {
	rule := m.topo.OrbitRule()
	var next polar.OrbitRule
	if rule != nil {
		next = *rule
	}
	next.MaxTheta = maxTheta
	m.SetOrbitRule(&next)
	m.persistOrbitRule()
	m.BroadcastRule()
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) takeOrbitActiveToggle() {
	m.SetOrbitActive(!m.OrbitActive())
	m.persistOrbitActive()
	m.BroadcastRule()
	if m.tr != nil {
		m.emitGeometry()
	}
}

func (m *NodeGeometry) persistOrbitRule() {
	if m.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteOrbitRule(m.persistRoot, m.id, m.topo.OrbitRule()); err != nil {
		jsonpersist.LogPersistErr("node_geometry_orbit_edit", m.id, err)
	}
}

func (m *NodeGeometry) persistOrbitActive() {
	if m.persistRoot == "" {
		if m.tr != nil {
			m.tr.Breadcrumb("orbit-active-persist", m.id, "", "SKIPPED: persistRoot unset")
		}
		return
	}
	if m.tr != nil {
		m.tr.Breadcrumb("orbit-active-persist", m.id, "", m.persistRoot)
	}
	if err := nodefiles.WriteOrbitActive(m.persistRoot, m.id, m.OrbitActive()); err != nil {
		jsonpersist.LogPersistErr("node_geometry_orbit_edit", m.id, err)
	}
}
