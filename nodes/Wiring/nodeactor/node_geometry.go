package nodeactor

import (
	"time"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulenode"
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

	quant owners.Quant
	tr    *T.Trace

	msg owners.Messaging

	clocks owners.Clocks

	stream owners.Stream

	ui owners.UI

	tilt owners.Tilt

	readout owners.Readout

	outTargets []string

	anim *NodeAnimation

	topo owners.Topology

	deltas owners.Deltas

	flags owners.Flags

	beads owners.Beads

	outEdges owners.OutEdges

	interior owners.Interior

	rule rulenode.Link
}

func NewNodeGeometry(id string, geom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *NodeGeometry {
	anim := &NodeAnimation{
		id:     id,
		clocks: owners.NewClocks(clockSrc, clock.NewRealClock()),
	}

	ng := &NodeGeometry{
		id: id, geom: geom, tr: tr,
		msg: owners.NewMessaging(
			make(chan movemsg.Msg, inboxDepth),
			map[string]chan movemsg.Msg{},
			make(chan vec3, 1),
		),
		topo:   owners.NewTopology(),
		deltas: owners.NewDeltas(),
		clocks: owners.NewClocks(clockSrc, clock.NewRealClock()),
		tilt:   owners.NewTilt(tiltvector.FullTurnPhiIdx),
		anim:   anim,
	}

	ng.msg.SeedCenter(nodegeom.NodeWorldPos(geom))

	ng.beads.SetBeadTickFn(func() *time.Ticker { return time.NewTicker(clock.TickPeriod) })
	return ng
}

func (m *NodeGeometry) Animation() *NodeAnimation {
	return m.anim
}
