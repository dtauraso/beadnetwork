package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	interior "github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
)

func (m *NodeGeometry) WireMessaging(
	resolveDest func(id string) (owners.Deposit, bool),
	sendMove func(id string, msg movemsg.Msg),
	commitLocal func(id string, idx polarindex.Index),
) {
	m.msg.WireMessaging(resolveDest, sendMove, commitLocal)
}

func (m *NodeGeometry) EnsureNeighborChannel(otherID string) {
	m.msg.EnsureNeighborChannel(otherID)
}

func (m *NodeGeometry) AddMutualTarget(target string) {
	m.topo.AddMutualTarget(target)
}

func (m *NodeGeometry) AddEdgeID(edgeID string) {
	m.topo.AddEdgeID(edgeID)
}

func (m *NodeGeometry) AddNeighborKind(toID, kind string) {
	m.topo.AddNeighborKind(toID, kind)
}

func (m *NodeGeometry) SetSelfKind(kind string) {
	m.selfKind = kind
}

func (m *NodeGeometry) SetDragRule(rule *PolarRulesPanel.DragRule) {
	m.topo.SetDragRule(rule)
}

func (m *NodeGeometry) SetDragActive(active bool) {
	m.topo.SetDragActive(active)
}

func (m *NodeGeometry) SetSelfRule(rule *PolarRulesPanel.DragRule) {
	m.topo.SetSelfRule(rule)
}

func (m *NodeGeometry) SetSelfRuleActive(active bool) {
	m.topo.SetSelfRuleActive(active)
}

func (m *NodeGeometry) SetSceneFlags(coplanarEdges, upAxis bool) {
	m.flags.SetSceneFlags(coplanarEdges, upAxis)
}

func (m *NodeGeometry) SetBaseIndex(off polarindex.Index) {
	m.geom.BaseIndex = off
	m.msg.PublishCenter(owners.Vec3(nodegeom.NodeWorldPos(m.geom)))
}

func (m *NodeGeometry) SetDragIndex(off polarindex.Offset) {
	m.geom.DragIndex = off
	m.msg.PublishCenter(owners.Vec3(nodegeom.NodeWorldPos(m.geom)))
}

func (m *NodeGeometry) SetTopTiltVectorPhiIdx(idx int32) {
	m.tilt.SetTopTiltVectorPhiIdx(idx)
}

func (m *NodeGeometry) AddOutTarget(target string) {
	m.outTargets = append(m.outTargets, target)
}

func (m *NodeGeometry) AddBeadLine(pw *beadanimation.BeadLine, edgeRow int32) {
	m.anim.AddBeadLine(pw, edgeRow)
}

func (m *NodeGeometry) BindOutEdgeRun(label, targetID, targetKind string, port *beadanimation.Sender, dest *beadanimation.BeadLine) {
	m.outEdges.BindWire(label, targetID, targetKind, port, dest)
	m.outEdges.SetSrcID(m.id)
}

func (m *NodeGeometry) WireOutEdgeStream(label string, edgeRow int32, targetID, targetKind string, nodeRow, dstNodeRow int32, buildFrame owners.EdgeFrameBuilder) {
	m.outEdges.AddOutEdge(label, edgeRow, targetID, targetKind, nodeRow, dstNodeRow, buildFrame)
}

func (m *NodeGeometry) deriveOutEdgeGeometry() {
	m.outEdges.DeriveGeometry(m.geom, &m.deltas)
}

func (m *NodeGeometry) SetAnimSleepCh(ch <-chan int64) {
	m.anim.SetSleepCh(ch)
}

func (m *NodeGeometry) DeriveOutEdgeGeometryOnce() {
	m.deriveOutEdgeGeometry()
}

func (m *NodeGeometry) writeOutEdgeFrames(tick int64) {
	m.outEdges.WriteFrames(tick, m.geom, &m.deltas)
}

func (m *NodeGeometry) WireInteriorStream(row int32, buildFrame func(tick uint32), sceneRoot string) *interior.Emitter {
	stream := interior.NewInteriorStream(buildFrame, row, interior.SlotsPerNode)
	stream.SetValueWriter(interior.NewValueWriter(sceneRoot, int(row)))
	stream.SetSceneRoot(sceneRoot)
	mailbox := interior.NewMailbox(row)
	m.interior.SetInteriorStream(stream, mailbox)
	return interior.NewEmitter(mailbox, row)
}

func (m *NodeGeometry) writeInteriorFrames() {
	m.interior.WriteFrames(m.geom)
}

func (m *NodeGeometry) WireBeadStream(row int32, buildBeadFrame beadanimation.BeadFrameBuilder, sceneRoot string) {
	m.anim.SetBeadStream(row, buildBeadFrame, sceneRoot)
}

func (m *NodeGeometry) WireStream(row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame nodeframe.NodeFrameBuilder, sceneRoot string) {
	m.stream.SetStream(row, kindID, buildFrame)
	m.topo.SetNodeRowFor(nodeRowFor)
	m.trace.Wire(sceneRoot, row)
}

func (m *NodeGeometry) SetPersistRoot(root string) {
	m.persistRoot = root
	m.outEdges.SetPersistRoot(root)
	m.outEdges.SetSrcID(m.id)
}

func (m *NodeGeometry) CopyClockSrc() {
	m.clocks.CopyClockSrc()
}
