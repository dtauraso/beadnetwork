package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
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

	// outTargets is every edge target this node draws a chain of beads
	// toward, declared regardless of whether an out wire is physically
	// wired yet. It stays with geometry; the wire itself lives on anim.
	outTargets []string

	// anim is the peer goroutine that owns the outgoing PacedWires and the
	// clock. Nothing outside chain_beads.go and this file's setup touches
	// it directly — the two crossing messages (stepOut/pulseIn) are the
	// only traffic between the peers.
	anim *NodeAnimation

	stepOut    chan<- map[string]int
	pulseIn    <-chan map[string][]beadindex.Pulse
	lastPulses map[string][]beadindex.Pulse

	topo owners.Topology

	flags owners.Flags

	beads owners.Beads
}

func NewNodeGeometry(id string, geom nodegeom.NodeGeom, tr *T.Trace, clockSrc clock.Clock) *NodeGeometry {
	stepCh := make(chan map[string]int, 1)
	pulseCh := make(chan map[string][]beadindex.Pulse, 1)

	anim := &NodeAnimation{
		id:       id,
		clocks:   owners.NewClocks(clockSrc, clock.NewRealClock()),
		stepIn:   stepCh,
		pulseOut: pulseCh,
	}

	ng := &NodeGeometry{
		id: id, geom: geom, tr: tr,
		msg: owners.NewMessaging(
			make(chan movemsg.Msg, inboxDepth),
			map[string]chan movemsg.Msg{},
			make(chan vec3, 1),
		),
		topo:       owners.NewTopology(map[string]vec3{}),
		clocks:     owners.NewClocks(clockSrc, clock.NewRealClock()),
		tilt:       owners.NewTilt(tiltvector.FullTurnThetaIdx),
		anim:       anim,
		stepOut:    stepCh,
		pulseIn:    pulseCh,
		lastPulses: map[string][]beadindex.Pulse{},
	}

	ng.msg.SeedCenter(nodegeom.NodeWorldPos(geom))

	ng.beads.SetBeadTickFn(clock.NewTickChan)
	return ng
}

// Animation returns this node's animation peer, so moverreg can start its
// goroutine and wire its speed channel.
func (m *NodeGeometry) Animation() *NodeAnimation {
	return m.anim
}
