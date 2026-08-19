package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Buffer/streamframe"
)

func (m *NodeGeometry) ChannelVectorsIn() chan bool { return m.channels.In() }

func (m *NodeGeometry) pollChannelVectors() {
	_, turnedOn := m.channels.TakeOn()
	if turnedOn {
		m.channels.Forget()
	}
	center := nodegeom.NodeWorldPos(m.geom)
	if !m.channels.NeedsBroadcast(center) {
		return
	}
	if rn := m.RuleNode(); rn != nil {
		select {
		case rn.CenterIn() <- center:
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
		if shaft, head, ok := framegeom.ChannelArrow(self, peer); ok {
			out = append(out, streamframe.ChannelVector{Shaft: shaft, Head: head})
		}
		if shaft, head, ok := framegeom.ChannelArrow(peer, self); ok {
			out = append(out, streamframe.ChannelVector{Shaft: shaft, Head: head})
		}
	}
	return out
}
