package nodeactor

import (
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
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

func (m *NodeGeometry) SetOrbitRule(rule *polar.OrbitRule) {
	m.topo.SetOrbitRule(rule)
}

func (m *NodeGeometry) SetSceneFlags(coplanarEdges, upAxis bool) {
	m.flags.SetSceneFlags(coplanarEdges, upAxis)
}

func (m *NodeGeometry) SetQuantOffset(off quantoffset.QuantizedOffset) {
	m.quantOffset = off
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

func (m *NodeGeometry) WireBeadStream(w io.Writer, row int32, buildBeadFrame owners.BeadFrameBuilder) {
	m.anim.SetBeadStream(w, row, buildBeadFrame)
}

func (m *NodeGeometry) WireStream(streamOut streamclaim.StreamHandle, row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame nodeframe.NodeFrameBuilder) {
	m.stream.SetStream(streamOut, row, kindID, buildFrame)
	m.topo.SetNodeRowFor(nodeRowFor)
}

func (m *NodeGeometry) SetPersistRoot(root string) {
	m.persistRoot = root
}

func (m *NodeGeometry) CopyClockSrc() {
	m.clocks.CopyClockSrc()
}
