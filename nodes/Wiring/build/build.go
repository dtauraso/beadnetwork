package build

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

type buildCtx struct {
	ctx      context.Context
	spec     loadspec.TopoSpec
	tr       *T.Trace
	clk      clock.Clock
	sphere   geom.SceneSphere
	hasScene bool

	scenePath string

	nodeGeoms map[string]nodegeom.NodeGeom
	centers   map[string]wire.Vec3

	quantizedOffsets map[string]quantoffset.QuantizedOffset

	destWire      map[string]*wire.PacedWire
	edgeWire      loadspec.WireRegistry
	edgeEndpoints map[string]inputcodec.EdgeEndpoints

	edgeSteps    map[string]int
	edgeSegments map[string]wire.WireSegment

	md *dispatch.MoveDispatch

	speedSinks []chan float64

	nodeType           map[string]string
	kindBroadcastPorts map[string]map[string]bool

	inbound        map[string]map[string]string
	outbound       map[string]map[string][]string
	outboundHandle map[string]map[string][]string

	outSink map[string]*wire.Out
	nodes   []wire.Node

	vectorOutByNode map[string]chan tiltvector.TiltVectorMsg
	vectorInByNode  map[string]chan tiltvector.TiltVectorMsg
}

func buildFromSpec(ctx context.Context, spec loadspec.TopoSpec, tr *T.Trace, clk clock.Clock, sphere geom.SceneSphere, hasScene bool, scenePath string) ([]wire.Node, inputcodec.SlotRegistry, *dispatch.MoveDispatch, []chan float64, error) {
	b := &buildCtx{ctx: ctx, spec: spec, tr: tr, clk: clk, sphere: sphere, hasScene: hasScene, scenePath: scenePath}

	b.nodeGeoms, b.centers = topoderive.ComputeNodeGeometry(b.spec, b.sphere)
	b.quantizedOffsets = topoderive.ComputeQuantizedLayout(b.spec, b.sphere, b.centers, b.nodeGeoms)
	topoderive.ComputeReachRadii(b.spec, b.nodeGeoms)
	b.destWire, b.edgeWire, b.edgeEndpoints, b.edgeSteps, b.edgeSegments = topoderive.AllocateWires(b.spec, b.nodeGeoms, b.tr)
	b.vectorOutByNode, b.vectorInByNode = topoderive.AllocateVectorChannels(b.spec)
	if err := b.buildMoveDispatch(); err != nil {
		return nil, nil, nil, nil, err
	}
	b.nodeType, b.kindBroadcastPorts = kindapi.BuildTypeMaps(b.spec)
	b.inbound, b.outbound, b.outboundHandle = topoderive.BuildEdgeMaps(b.spec, b.nodeType, b.kindBroadcastPorts)
	if err := b.buildNodes(); err != nil {
		return nil, nil, nil, nil, err
	}

	b.md.MR.FinalizeActors(&b.speedSinks)
	bindDispatch(b.md, b.outSink, b.destWire)

	return b.nodes, inputcodec.SlotRegistry(b.destWire), b.md, b.speedSinks, nil
}
