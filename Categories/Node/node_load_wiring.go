package Node

import (
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	interior "github.com/dtauraso/beadnetwork/Categories/Node/Interior"
	"github.com/dtauraso/beadnetwork/Categories/Polar/polarindex"
)

func (m *NodeGeometry) SetSelfKind(kind string) {
	m.selfKind = kind
}

func (m *NodeGeometry) SetBaseIndex(off polarindex.Index) {
	m.geom.BaseIndex = off
	m.publishCenter()
}

func (m *NodeGeometry) SetDragIndex(off polarindex.Offset) {
	m.geom.DragIndex = off
	m.publishCenter()
}

func (m *NodeGeometry) publishCenter() {
	m.msg.PublishCenter(Vec3(NodeWorldPos(m.geom)))
}

func (m *NodeGeometry) AddOutTarget(target string) {
	m.outTargets = append(m.outTargets, target)
}

func (m *NodeGeometry) BindOutEdgeRun(label, targetID, targetKind string, port *beadanimation.Sender, dest *beadanimation.BeadLine) {
	m.outEdges.BindWire(label, targetID, targetKind, port, dest)
	m.outEdges.SetSrcID(m.id)
}

func (m *NodeGeometry) DeriveOutEdgeGeometry() {
	m.outEdges.DeriveGeometry(m.geom, &m.deltas)
}

func (m *NodeGeometry) writeOutEdgeFrames(tick int64) {
	m.outEdges.WriteFrames(tick, m.geom, &m.deltas)
}

func (m *NodeGeometry) WireInteriorStream(row int32, buildFrame func(tick uint32), sceneRoot string) *interior.Emitter {
	stream, mailbox, emitter := interior.Wire(row, buildFrame, sceneRoot)
	m.interior.SetInteriorStream(stream, mailbox)
	return emitter
}

func (m *NodeGeometry) writeInteriorFrames() {
	m.interior.WriteFrames(interior.Vec3(NodeWorldPos(m.geom)))
}

func (m *NodeGeometry) WireStream(row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame NodeFrameBuilder, sceneRoot string) {
	m.stream.SetStream(row, kindID, buildFrame)
	m.topo.SetNodeRowFor(nodeRowFor)
	m.trace.Wire(sceneRoot, row)
}

func (m *NodeGeometry) SetPersistRoot(root string) {
	m.persistRoot = root
	m.outEdges.SetPersistRoot(root)
	m.outEdges.SetSrcID(m.id)
}
