package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/owners"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
	streamframe "github.com/dtauraso/wirefold/Categories/Scene/Vectors"
)

func (m *NodeGeometry) pollChannelVectors() {
	_, turnedOn := m.channels.TakeOn()
	if turnedOn {
		m.channels.Forget()
	}
	center := nodegeom.NodeWorldPos(m.geom)
	if !m.channels.NeedsBroadcast(owners.Vec3(center)) {
		return
	}
	if rn := m.RuleNode(); rn != nil {
		select {
		case rn.CenterIn() <- rulenode.Vec3(center):
		default:
		}
	}
}

func (m *NodeGeometry) channelVectors() []streamframe.ChannelVector {
	if !m.channels.On() {
		return nil
	}
	peers := make(map[string]streamframe.Vec3, len(m.channels.PeerCenters()))
	for id, c := range m.channels.PeerCenters() {
		peers[id] = streamframe.Vec3(c)
	}
	return streamframe.ChannelVectorsFor(streamframe.Vec3(nodegeom.NodeWorldPos(m.geom)), peers)
}
