package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/streamclaim"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

func (m *NodeGeometry) WireMessaging(
	resolveDest func(id string) (func(movemsg.Msg) bool, bool),
	sendMove func(id string, msg movemsg.Msg),
	centerOf func(id string) (vec3, bool),
	commitLocal func(id string, newPos vec3),
) {
	m.msg.WireMessaging(resolveDest, sendMove, centerOf, commitLocal)
}

func (m *NodeGeometry) EnsureNeighborChannel(otherID string) {
	m.msg.EnsureNeighborChannel(otherID)
}

func (m *NodeGeometry) AddMutualTarget(target string) {
	m.topo.AddMutualTarget(target)
}

func (m *NodeGeometry) SeedPartnerCenter(neighborID string, c vec3) {
	m.topo.SetPathTo(neighborID, m.WorldCenter(), c)
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

func (m *NodeGeometry) SetSceneFlags(coplanarEdges, upAxis bool) {
	m.flags.SetSceneFlags(coplanarEdges, upAxis)
}

func (m *NodeGeometry) SetQuantOffset(off quantoffset.QuantizedOffset) {
	m.quantOffset = off
}

func (m *NodeGeometry) SetTopTiltVectorThetaIdx(idx int32) {
	m.tilt.SetTopTiltVectorThetaIdx(idx)
}

func (m *NodeGeometry) AddOutTarget(target string) {
	m.outTargets = append(m.outTargets, target)
}

func (m *NodeGeometry) AddOutWire(pw *wire.PacedWire, target string, o *outport.Out, sendSteps func(int)) {
	m.anim.AddOutWire(pw, target, o, sendSteps)
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
