package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
)

func (m *NodeGeometry) pollChannelVectors() {
	_, turnedOn := m.channels.TakeOn()
	if turnedOn {
		m.channels.Forget()
	}
	center := NodeWorldPos(m.geom)
	if !m.channels.NeedsBroadcast(ChannelVectors.Vec3(center)) {
		return
	}
	if rn := m.RuleNode(); rn != nil {
		select {
		case rn.CenterIn() <- rulenode.Vec3(center):
		default:
		}
	}
}

func (m *NodeGeometry) channelVectors() []ChannelVectors.ChannelVector {
	return m.channels.VectorsFrom(ChannelVectors.Vec3(NodeWorldPos(m.geom)))
}
