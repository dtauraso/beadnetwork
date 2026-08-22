package nodeactor

import (
	"github.com/dtauraso/wirefold/Node/framegeom"
	"github.com/dtauraso/wirefold/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/Node/nodegeom"
	"github.com/dtauraso/wirefold/Node/rulenode"
	streamframe "github.com/dtauraso/wirefold/Scene/Vectors"
)

func (m *NodeGeometry) ChannelVectorsIn() chan bool { return m.channels.In() }

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
	self := nodegeom.NodeWorldPos(m.geom)
	peers := m.channels.PeerCenters()
	out := make([]streamframe.ChannelVector, 0, 2*len(peers))
	for _, peer := range peers {
		if shaft, head, ok := framegeom.ChannelArrow(framegeom.Vec3(self), framegeom.Vec3(peer)); ok {
			out = append(out, streamframe.ChannelVector{Shaft: shaft, Head: head})
		}
		if shaft, head, ok := framegeom.ChannelArrow(framegeom.Vec3(peer), framegeom.Vec3(self)); ok {
			out = append(out, streamframe.ChannelVector{Shaft: shaft, Head: head})
		}
	}
	return out
}
