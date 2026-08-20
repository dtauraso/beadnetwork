package runtopology

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Input/dispatch"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/runtopology/loadspec"
	"github.com/dtauraso/wirefold/src/runtopology/topoderive"
	"github.com/dtauraso/wirefold/src/spatial"
)

type buildCtx struct {
	ctx      context.Context
	spec     loadspec.TopoSpec
	clk      clock.Clock
	sphere   polar.SceneSphere
	hasScene bool

	scenePath string

	nodeGeoms map[string]nodegeom.NodeGeom
	centers   map[string]spatial.Vec3

	baseIndices map[string]polarindex.Index
	dragIndices map[string]polarindex.Offset

	destRun       map[string]*beadanimation.BeadLine
	edgeRun       loadspec.BeadLineRegistry
	edgeEndpoints map[string]inputcodec.EdgeEndpoints

	md *dispatch.MoveDispatch

	speedSinks SliderPanel.Sinks

	nodeType           map[string]string
	kindBroadcastPorts map[string]map[string]bool

	inbound        map[string]map[string]string
	outbound       map[string]map[string][]string
	outboundHandle map[string]map[string][]string

	outSink map[string]*beadanimation.Sender
	nodes   []nodeapi.Node

	vectorOutByNode map[string]chan TiltPanel.TiltVectorMsg
	vectorInByNode  map[string]chan TiltPanel.TiltVectorMsg
}

func buildFromSpec(ctx context.Context, spec loadspec.TopoSpec, clk clock.Clock, sphere polar.SceneSphere, hasScene bool, scenePath string) ([]nodeapi.Node, inputcodec.SlotRegistry, *dispatch.MoveDispatch, SliderPanel.Sinks, error) {
	b := &buildCtx{ctx: ctx, spec: spec, clk: clk, sphere: sphere, hasScene: hasScene, scenePath: scenePath}

	b.nodeGeoms, b.centers = topoderive.ComputeNodeGeometry(b.spec, b.sphere)
	b.baseIndices = topoderive.ComputeBaseIndices(b.spec, b.sphere, b.centers, b.nodeGeoms)
	b.dragIndices = topoderive.ComputeDragIndices(b.spec)
	b.destRun, b.edgeRun, b.edgeEndpoints = topoderive.AllocateBeadLines(b.spec, b.nodeGeoms)
	b.vectorOutByNode, b.vectorInByNode = topoderive.AllocateVectorChannels(b.spec)
	if err := b.buildMoveDispatch(); err != nil {
		return nil, nil, nil, SliderPanel.Sinks{}, err
	}
	b.nodeType, b.kindBroadcastPorts = kindreg.BuildTypeMaps(b.spec)
	b.inbound, b.outbound, b.outboundHandle = topoderive.BuildEdgeMaps(b.spec, b.nodeType, b.kindBroadcastPorts)
	if err := b.buildNodes(); err != nil {
		return nil, nil, nil, SliderPanel.Sinks{}, err
	}

	bindDispatch(b.md, b.outSink, b.destRun)

	return b.nodes, inputcodec.SlotRegistry(b.destRun), b.md, b.speedSinks, nil
}
