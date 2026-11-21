package Node

import (
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	"github.com/dtauraso/beadnetwork/Categories/Node/Block"
	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"
	"github.com/dtauraso/beadnetwork/Categories/Node/Geom"
	interior "github.com/dtauraso/beadnetwork/Categories/Node/Interior"
	"github.com/dtauraso/beadnetwork/Categories/Node/Messaging"
	"github.com/dtauraso/beadnetwork/Categories/Node/OutEdges"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type NodeGeometry struct {
	id   string
	geom Geom.NodeGeom

	kindRule NodeDrag.KindRule

	persistRoot string
	selfKind    string

	msg Messaging.Messaging

	clocks Clocks

	stream Block.Stream

	trace Trace

	ui UI

	tilt TiltVectors.Tilt


	outTargets []string

	anim *NodeBeadAnimation

	topo Topology

	deltas OutEdges.Deltas

	flags Flags

	beads beadanimation.Beads

	outEdges OutEdges.OutEdges

	interior interior.Interior

	kindPosts Messaging.KindPosts

}

func NewNodeGeometry(id string, geom Geom.NodeGeom, clockSrc clock.Clock, constants polarindex.SceneConstants) *NodeGeometry {
	anim := NewNodeBeadAnimation(id)

	ng := &NodeGeometry{
		id: id, geom: geom,
		msg: Messaging.NewMessaging(
			make(chan Messaging.Msg, Messaging.InboxDepth),
			make(chan Vec3, 1),
		),
		topo:      NewTopology(),
		deltas:    OutEdges.NewDeltas(),
		clocks:    NewClocks(clockSrc),
		tilt:      TiltVectors.NewTilt(polarindex.FullTurnPhiIdx),
		anim:      anim,
		kindPosts: Messaging.NewKindPosts(),
	}

	ng.msg.SeedCenter(Vec3(Geom.NodeWorldPos(geom)))
	ng.outEdges.SetConstants(constants)
	ng.deltas.SetConstants(constants)

	return ng
}
