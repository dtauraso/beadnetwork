package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

type NodeGeometry struct {
	id   string
	geom nodegeom.NodeGeom

	persistRoot string
	selfKind    string

	quantOffset quantoffset.QuantizedOffset
	tr          *T.Trace

	msg nodeMessaging

	clocks nodeClocks

	stream nodeStream

	ui nodeUI

	tilt nodeTilt

	readout pairReadout

	outs nodeOuts

	topo neighborTopology

	flags sceneFlags

	beads nodeBeads
}

func NewNodeGeometry(id string, geom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *NodeGeometry {
	ng := &NodeGeometry{
		id: id, geom: geom, tr: tr,
		msg: nodeMessaging{
			extIn:      make(chan movemsg.Msg, inboxDepth),
			neighborIn: map[string]chan movemsg.Msg{},
			centerOut:  make(chan vec3, 1),
		},
		topo:   neighborTopology{partnerCenters: map[string]vec3{}},
		clocks: nodeClocks{clockSrc: clockSrc, clk: clock.NewRealClock()},
		tilt:   nodeTilt{latticePoints: tiltvector.FullTurnThetaIdx},
	}

	ng.msg.centerOut <- nodegeom.NodeWorldPos(geom)

	ng.beads.beadTickFn = clock.NewTickChan
	return ng
}
