package nodeactor

import (
	"slices"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

func (m *NodeGeometry) ID() string { return m.id }

func (m *NodeGeometry) Traced() bool { return m.tr != nil }

func (m *NodeGeometry) Breadcrumb(label, node, port, value string) {
	if m.tr != nil {
		m.tr.Breadcrumb(label, node, port, value)
	}
}

func (m *NodeGeometry) Kind() string { return m.geom.Kind }

func (m *NodeGeometry) SelfKind() string { return m.selfKind }

func (m *NodeGeometry) Tick() int64 { return m.clocks.Tick() }

func (m *NodeGeometry) Label() string { return m.geom.Label }

func (m *NodeGeometry) WorldCenter() vec3 { return nodegeom.NodeWorldPos(m.geom) }

func (m *NodeGeometry) NodeRow() int32 { return m.stream.NodeRow() }

func (m *NodeGeometry) EdgeIDs() []string { return m.topo.EdgeIDs() }

func (m *NodeGeometry) NeighborKinds() map[string]string { return m.topo.NeighborKinds() }

// IsOutTarget says whether the edge to that neighbour LEAVES this node, which
// is what separates a neighbour whose constraints this node must satisfy from
// one whose constraints it imposes.
// OutTargets is every edge target this node points at.
func (m *NodeGeometry) OutTargets() []string { return m.outTargets }

func (m *NodeGeometry) IsOutTarget(neighborID string) bool {
	return slices.Contains(m.outTargets, neighborID)
}

func (m *NodeGeometry) SendMove() func(id string, msg movemsg.Msg) { return m.msg.SendMove() }

func (m *NodeGeometry) NeighborIDs() []string { return m.msg.NeighborIDs() }

func (m *NodeGeometry) QuantOffset() (iTheta, iPhi, iR int) {
	return m.quantOffset.IPhi, m.quantOffset.ITheta, m.quantOffset.IR
}

func (m *NodeGeometry) QuantizedOffsetValue() quantoffset.QuantizedOffset { return m.quantOffset }

func (m *NodeGeometry) ReachR() float64 { return m.geom.ReachR }

func (m *NodeGeometry) CommitQuantOffset(committedPolar polar.Polar) {
	off := quantoffset.MeasureScalar(committedPolar, m.quantOffset)
	m.quantOffset = off
	m.persistQuantOffset(off, committedPolar)
}

func (m *NodeGeometry) WriteStreamFrame(events []rowevent.RowEvent) {
	m.writeStreamFrame(events)
}
