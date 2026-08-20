package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Chrome/TiltPanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/movemsg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/polarindex"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulenode"
	"github.com/dtauraso/wirefold/src/Clock"
)

const inboxDepth = 8

type NodeGeometry struct {
	id   string
	geom nodegeom.NodeGeom

	persistRoot string
	selfKind    string

	msg owners.Messaging

	clocks owners.Clocks

	stream owners.Stream

	ui owners.UI

	tilt owners.Tilt

	channels owners.ChannelVectors

	readout owners.Readout

	outTargets []string

	anim *NodeAnimation

	topo owners.Topology

	deltas owners.Deltas

	flags owners.Flags

	beads owners.Beads

	outEdges owners.OutEdges

	interior owners.Interior

	kindPosts owners.KindPosts

	rule rulenode.Link
}

func NewNodeGeometry(id string, geom nodegeom.NodeGeom, clockSrc clock.Clock, constants polarindex.SceneConstants) *NodeGeometry {
	anim := &NodeAnimation{
		id:     id,
		clocks: owners.NewClocks(clockSrc, clock.NewRealClock()),
	}

	ng := &NodeGeometry{
		id: id, geom: geom,
		msg: owners.NewMessaging(
			make(chan movemsg.Msg, inboxDepth),
			make(chan vec3, 1),
		),
		topo:      owners.NewTopology(),
		deltas:    owners.NewDeltas(),
		clocks:    owners.NewClocks(clockSrc, clock.NewRealClock()),
		tilt:      owners.NewTilt(TiltPanel.FullTurnPhiIdx),
		anim:      anim,
		kindPosts: owners.NewKindPosts(),
	}

	ng.msg.SeedCenter(nodegeom.NodeWorldPos(geom))
	ng.outEdges.SetConstants(constants)
	ng.deltas.SetConstants(constants)

	return ng
}

func (m *NodeGeometry) Animation() *NodeAnimation {
	return m.anim
}
