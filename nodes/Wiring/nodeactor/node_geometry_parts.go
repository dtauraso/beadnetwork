package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

type nodeMessaging struct {
	extIn chan movemsg.Msg

	neighborIn map[string]chan movemsg.Msg

	centerOut chan vec3

	sendMove func(id string, msg movemsg.Msg)

	resolveDest func(id string) (func(movemsg.Msg) bool, bool)

	centerOf func(id string) (vec3, bool)

	commitLocal func(id string, newPos vec3)

	pending []pendingSend
}

type pendingSend struct {
	destID string
	msg    movemsg.Msg
}

type nodeClocks struct {
	clockSrc clock.Clock

	clk clock.Clock
}

type nodeStream struct {
	streamOut StreamHandle
	nodeRow   int32
	kindID    uint8

	buildFrame NodeFrameBuilder
}

type nodeUI struct {
	selected, hovered, latchedSel uint8
	hoverPort                     string
	hoverIsInput                  bool
}

type nodeTilt struct {
	topTiltVectorThetaIdx  int32
	normalThetaIdx         int32
	bottomThetaIdx         int32
	receivedVectorThetaIdx int32
	receivedVectorSet      bool

	latticePoints int32
}

type pairReadout struct {
	roundsToParallel int32

	msgsToParallel int32
}

type nodeOuts struct {
	outTargets     []string
	outWires       []*wire.PacedWire
	outWireTargets []string
	outWireOuts    []*wire.Out

	outStepsIn []func(int)
}

type neighborTopology struct {
	edgeIDs []string

	partnerCenters map[string]vec3
	neighborKinds  map[string]string
	mutualTargets  map[string]bool
	nodeRowFor     func(id string) (int32, bool)
}

type sceneFlags struct {
	coplanarEdges bool
	upAxis        bool
}

type nodeBeads struct {
	beadTickFn func() <-chan struct{}
	beadChains map[string]*edgeBeadChain
}
