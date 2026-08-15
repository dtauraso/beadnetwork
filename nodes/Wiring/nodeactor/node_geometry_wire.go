package nodeactor

import (
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

func (m *NodeGeometry) WireMessaging(
	resolveDest func(id string) (func(movemsg.Msg) bool, bool),
	sendMove func(id string, msg movemsg.Msg),
	commitLocal func(id string, newPos vec3, targetPolar *polar.Polar),
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

func (m *NodeGeometry) SetDragRule(rule *polar.DragRule) {
	m.topo.SetDragRule(rule)
}

func (m *NodeGeometry) SetDragActive(active bool) {
	m.topo.SetDragActive(active)
}

func (m *NodeGeometry) SetSceneFlags(coplanarEdges, upAxis bool) {
	m.flags.SetSceneFlags(coplanarEdges, upAxis)
}

func (m *NodeGeometry) SetQuantOffset(off quantoffset.QuantizedOffset) {
	m.quant.SetBase(off)
}

func (m *NodeGeometry) SetDragQuantOffset(off quantoffset.QuantizedOffset) {
	m.quant.SetDrag(off)
}

func (m *NodeGeometry) SetTopTiltVectorPhiIdx(idx int32) {
	m.tilt.SetTopTiltVectorPhiIdx(idx)
}

func (m *NodeGeometry) AddOutTarget(target string) {
	m.outTargets = append(m.outTargets, target)
}

func (m *NodeGeometry) AddOutWire(pw *wire.PacedWire, edgeRow int32) {
	m.anim.AddOutWire(pw, edgeRow)
}

func (m *NodeGeometry) BindOutEdgeWire(label, targetID, targetKind string, port *outport.Out, dest *wire.PacedWire) {
	m.outEdges.BindWire(label, targetID, targetKind, port, dest)
	m.outEdges.SetSrcID(m.id)
}

func (m *NodeGeometry) WireOutEdgeStream(label string, edgeRow int32, targetID, targetKind string, w io.Writer, nodeRow, dstNodeRow int32, buildFrame owners.EdgeFrameBuilder) {
	m.outEdges.AddOutEdge(label, edgeRow, targetID, targetKind, w, nodeRow, dstNodeRow, buildFrame)
}

func (m *NodeGeometry) deriveOutEdgeGeometry(tick int64) {
	m.outEdges.DeriveGeometry(tick, m.geom, &m.deltas)
}

func (m *NodeGeometry) writeOutEdgeFrames(tick int64) {
	m.outEdges.WriteFrames(tick, m.geom, &m.deltas)
}

func (m *NodeGeometry) WireInteriorStream(w io.Writer, row int32, buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte) *interior.Emitter {
	stream := interior.NewInteriorStream(w, buildFrame, row, interior.SlotsPerNode)
	mailbox := interior.NewMailbox(row)
	m.interior.SetInteriorStream(stream, mailbox)
	return interior.NewEmitter(mailbox, row)
}

func (m *NodeGeometry) writeInteriorFrames() {
	m.interior.WriteFrames()
}

func (m *NodeGeometry) WireBeadStream(w io.Writer, row int32, buildBeadFrame owners.BeadFrameBuilder) {
	m.anim.SetBeadStream(w, row, buildBeadFrame)
}

func (m *NodeGeometry) WireStream(streamOut streamclaim.StreamHandle, row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame nodeframe.NodeFrameBuilder) {
	m.stream.SetStream(streamOut, row, kindID, buildFrame)
	m.topo.SetNodeRowFor(nodeRowFor)
}

func (m *NodeGeometry) SetPersistRoot(root string) {
	m.persistRoot = root
	m.outEdges.SetPersistRoot(root)
	m.outEdges.SetSrcID(m.id)
}

func (m *NodeGeometry) CopyClockSrc() {
	m.clocks.CopyClockSrc()
}
