package Node

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	NodeDrag "github.com/dtauraso/wirefold/Categories/Node/Drag"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

type NodeGeometry struct {
	id   string
	geom nodegeom.NodeGeom

	kindRule NodeDrag.KindRule

	persistRoot string
	selfKind    string

	msg Messaging

	clocks Clocks

	stream Stream

	trace Trace

	ui UI

	tilt TiltVectors.Tilt

	channels ChannelVectors.PeerCenters

	readout Readout

	outTargets []string

	anim *NodeBeadAnimation

	topo Topology

	deltas Deltas

	flags Flags

	beads beadanimation.Beads

	outEdges OutEdges

	interior interior.Interior

	kindPosts KindPosts

	rule rulenode.Link
}

func NewNodeGeometry(id string, geom nodegeom.NodeGeom, clockSrc clock.Clock, constants polarindex.SceneConstants) *NodeGeometry {
	anim := NewNodeBeadAnimation(id, NewClocks(clockSrc, clock.NewRealClock()))

	ng := &NodeGeometry{
		id: id, geom: geom,
		msg: NewMessaging(
			make(chan Msg, InboxDepth),
			make(chan Vec3, 1),
		),
		topo:      NewTopology(),
		deltas:    NewDeltas(),
		clocks:    NewClocks(clockSrc, clock.NewRealClock()),
		tilt:      TiltVectors.NewTilt(TiltPanel.FullTurnPhiIdx),
		anim:      anim,
		kindPosts: NewKindPosts(),
	}

	ng.msg.SeedCenter(Vec3(nodegeom.NodeWorldPos(geom)))
	ng.outEdges.SetConstants(constants)
	ng.deltas.SetConstants(constants)

	return ng
}
