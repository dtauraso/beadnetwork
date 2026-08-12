package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
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

func (m *NodeGeometry) PartnerCenters() map[string]vec3 { return m.topo.PartnerCenters() }

func (m *NodeGeometry) NeighborKinds() map[string]string { return m.topo.NeighborKinds() }

func (m *NodeGeometry) SendMove() func(id string, msg movemsg.Msg) { return m.msg.SendMove() }

func (m *NodeGeometry) NeighborIDs() []string { return m.msg.NeighborIDs() }

func (m *NodeGeometry) QuantOffset() (iTheta, iPhi, iR int) {
	return m.quantOffset.ITheta, m.quantOffset.IPhi, m.quantOffset.IR
}

func (m *NodeGeometry) QuantizedOffsetValue() quantoffset.QuantizedOffset { return m.quantOffset }

func (m *NodeGeometry) ReachR() float64 { return m.geom.ReachR }

func (m *NodeGeometry) CommitQuantOffset(committedPolar geom.Polar) {
	off := quantoffset.MeasureScalar(committedPolar, m.quantOffset)
	m.quantOffset = off
	m.persistQuantOffset(off, committedPolar)
}

func (m *NodeGeometry) WriteStreamFrame(events []rowevent.RowEvent) {
	m.writeStreamFrame(events)
}
