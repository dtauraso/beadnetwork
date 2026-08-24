package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	"github.com/dtauraso/beadnetwork/Categories/Node/ChannelVectors"
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	interior "github.com/dtauraso/beadnetwork/Categories/Node/Interior"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Polar/polarindex"
)

type NodeGeometry struct {
	id   string
	geom NodeGeom

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

	rule Link
}

func NewNodeGeometry(id string, geom NodeGeom, clockSrc clock.Clock, constants polarindex.SceneConstants) *NodeGeometry {
	anim := NewNodeBeadAnimation(id)

	ng := &NodeGeometry{
		id: id, geom: geom,
		msg: NewMessaging(
			make(chan Msg, InboxDepth),
			make(chan Vec3, 1),
		),
		topo:      NewTopology(),
		deltas:    NewDeltas(),
		clocks:    NewClocks(clockSrc),
		tilt:      TiltVectors.NewTilt(TiltPanel.FullTurnPhiIdx),
		anim:      anim,
		kindPosts: NewKindPosts(),
	}

	ng.msg.SeedCenter(Vec3(NodeWorldPos(geom)))
	ng.outEdges.SetConstants(constants)
	ng.deltas.SetConstants(constants)

	return ng
}
