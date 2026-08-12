package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

const inboxDepth = 8

type NodeGeometry struct {
	id   string
	geom nodegeom.NodeGeom

	persistRoot string
	selfKind    string

	quantOffset quantoffset.QuantizedOffset
	tr          *T.Trace

	msg owners.Messaging

	clocks owners.Clocks

	stream owners.Stream

	ui owners.UI

	tilt owners.Tilt

	readout owners.Readout

	outs owners.Outs

	topo owners.Topology

	flags owners.Flags

	beads owners.Beads
}

func NewNodeGeometry(id string, geom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *NodeGeometry {
	ng := &NodeGeometry{
		id: id, geom: geom, tr: tr,
		msg: owners.NewMessaging(
			make(chan movemsg.Msg, inboxDepth),
			map[string]chan movemsg.Msg{},
			make(chan vec3, 1),
		),
		topo:   owners.NewTopology(map[string]vec3{}),
		clocks: owners.NewClocks(clockSrc, clock.NewRealClock()),
		tilt:   owners.NewTilt(tiltvector.FullTurnThetaIdx),
	}

	ng.msg.SeedCenter(nodegeom.NodeWorldPos(geom))

	ng.beads.SetBeadTickFn(clock.NewTickChan)
	return ng
}
