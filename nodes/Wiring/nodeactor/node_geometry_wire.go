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
	m.topo.SeedPartnerCenter(neighborID, c)
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
	m.flags.coplanarEdges = coplanarEdges
	m.flags.upAxis = upAxis
}

func (m *NodeGeometry) SetQuantOffset(off quantoffset.QuantizedOffset) {
	m.quantOffset = off
}

func (m *NodeGeometry) SetTopTiltVectorThetaIdx(idx int32) {
	m.tilt.topTiltVectorThetaIdx = idx
}

func (m *NodeGeometry) AddOutTarget(target string) {
	m.outs.outTargets = append(m.outs.outTargets, target)
}

func (m *NodeGeometry) AddOutWire(pw *wire.PacedWire, target string, o *outport.Out, sendSteps func(int)) {
	m.outs.outWires = append(m.outs.outWires, pw)
	m.outs.outWireTargets = append(m.outs.outWireTargets, target)
	m.outs.outWireOuts = append(m.outs.outWireOuts, o)
	m.outs.outStepsIn = append(m.outs.outStepsIn, sendSteps)
}

func (m *NodeGeometry) WireStream(streamOut streamclaim.StreamHandle, row int32, kindID uint8, nodeRowFor func(id string) (int32, bool), buildFrame nodeframe.NodeFrameBuilder) {
	m.stream.streamOut = streamOut
	m.stream.nodeRow = row
	m.stream.kindID = kindID
	m.topo.nodeRowFor = nodeRowFor
	m.stream.buildFrame = buildFrame
}

func (m *NodeGeometry) SetPersistRoot(root string) {
	m.persistRoot = root
}

func (m *NodeGeometry) CopyClockSrc() {
	m.clocks.CopyClockSrc()
}
