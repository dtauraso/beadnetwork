package Node

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/owners"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

type NodeGeometry struct {
	id   string
	geom nodegeom.NodeGeom

	kindRule owners.KindRule

	persistRoot string
	selfKind    string

	msg owners.Messaging

	clocks owners.Clocks

	stream owners.Stream

	trace owners.Trace

	ui owners.UI

	tilt owners.Tilt

	channels ChannelVectors.PeerCenters

	readout owners.Readout

	outTargets []string

	anim *owners.NodeBeadAnimation

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
	anim := owners.NewNodeBeadAnimation(id, owners.NewClocks(clockSrc, clock.NewRealClock()))

	ng := &NodeGeometry{
		id: id, geom: geom,
		msg: owners.NewMessaging(
			make(chan owners.Msg, owners.InboxDepth),
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
