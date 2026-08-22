package nodeactor

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/rulenode"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
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

	anim *NodeBeadAnimation

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
	anim := &NodeBeadAnimation{
		id:     id,
		clocks: owners.NewClocks(clockSrc, clock.NewRealClock()),
	}

	ng := &NodeGeometry{
		id: id, geom: geom,
		msg: owners.NewMessaging(
			make(chan movemsg.Msg, inboxDepth),
			make(chan owners.Vec3, 1),
		),
		topo:      owners.NewTopology(),
		deltas:    owners.NewDeltas(),
		clocks:    owners.NewClocks(clockSrc, clock.NewRealClock()),
		tilt:      owners.NewTilt(TiltPanel.FullTurnPhiIdx),
		anim:      anim,
		kindPosts: owners.NewKindPosts(),
	}

	ng.msg.SeedCenter(owners.Vec3(nodegeom.NodeWorldPos(geom)))
	ng.outEdges.SetConstants(constants)
	ng.deltas.SetConstants(constants)

	return ng
}

func (m *NodeGeometry) BeadAnimation() *NodeBeadAnimation {
	return m.anim
}
