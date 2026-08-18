package build

import (
	"context"
	"github.com/dtauraso/wirefold/Slider"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/bead"
	"github.com/dtauraso/wirefold/nodes/bead/outport"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/nodeapi"
	"github.com/dtauraso/wirefold/nodes/spatial"

	T "github.com/dtauraso/wirefold/Trace"
)

type buildCtx struct {
	ctx      context.Context
	spec     loadspec.TopoSpec
	tr       *T.Trace
	clk      clock.Clock
	sphere   polar.SceneSphere
	hasScene bool

	scenePath string

	nodeGeoms map[string]nodegeom.NodeGeom
	centers   map[string]spatial.Vec3

	baseIndices map[string]polarindex.Index
	dragIndices map[string]polarindex.Offset

	destRun       map[string]*bead.BeadRun
	edgeRun       loadspec.BeadRunRegistry
	edgeEndpoints map[string]inputcodec.EdgeEndpoints

	md *dispatch.MoveDispatch

	speedSinks Slider.Sinks

	nodeType           map[string]string
	kindBroadcastPorts map[string]map[string]bool

	inbound        map[string]map[string]string
	outbound       map[string]map[string][]string
	outboundHandle map[string]map[string][]string

	outSink map[string]*outport.Out
	nodes   []nodeapi.Node

	vectorOutByNode map[string]chan tiltvector.TiltVectorMsg
	vectorInByNode  map[string]chan tiltvector.TiltVectorMsg
}

func buildFromSpec(ctx context.Context, spec loadspec.TopoSpec, tr *T.Trace, clk clock.Clock, sphere polar.SceneSphere, hasScene bool, scenePath string) ([]nodeapi.Node, inputcodec.SlotRegistry, *dispatch.MoveDispatch, Slider.Sinks, error) {
	b := &buildCtx{ctx: ctx, spec: spec, tr: tr, clk: clk, sphere: sphere, hasScene: hasScene, scenePath: scenePath}

	b.nodeGeoms, b.centers = topoderive.ComputeNodeGeometry(b.spec, b.sphere)
	b.baseIndices = topoderive.ComputeBaseIndices(b.spec, b.sphere, b.centers, b.nodeGeoms)
	b.dragIndices = topoderive.ComputeDragIndices(b.spec)
	b.destRun, b.edgeRun, b.edgeEndpoints = topoderive.AllocateBeadRuns(b.spec, b.nodeGeoms, b.tr)
	b.vectorOutByNode, b.vectorInByNode = topoderive.AllocateVectorChannels(b.spec)
	if err := b.buildMoveDispatch(); err != nil {
		return nil, nil, nil, Slider.Sinks{}, err
	}
	b.nodeType, b.kindBroadcastPorts = kindreg.BuildTypeMaps(b.spec)
	b.inbound, b.outbound, b.outboundHandle = topoderive.BuildEdgeMaps(b.spec, b.nodeType, b.kindBroadcastPorts)
	if err := b.buildNodes(); err != nil {
		return nil, nil, nil, Slider.Sinks{}, err
	}

	bindDispatch(b.md, b.outSink, b.destRun)

	return b.nodes, inputcodec.SlotRegistry(b.destRun), b.md, b.speedSinks, nil
}
