package runtopology

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	edge "github.com/dtauraso/wirefold/src/Node/Edge"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/runtopology/loadspec"
	"github.com/dtauraso/wirefold/src/runtopology/scenerun"
)

type buildCtx struct {
	ctx      context.Context
	spec     loadspec.TopoSpec
	clk      clock.Clock
	sphere   polar.SceneSphere
	hasScene bool

	scenePath string

	nodeGeoms map[string]nodegeom.NodeGeom

	baseIndices map[string]polarindex.Index
	dragIndices map[string]polarindex.Offset

	edgeEndpoints map[string]edge.EdgeEndpoints

	wiring kindreg.EdgeWiring

	md *scenerun.MoveDispatch

	speedSinks SliderPanel.Sinks

	nodeType           map[string]string
	kindBroadcastPorts map[string]map[string]bool

	outSink map[string]*beadanimation.Sender
	nodes   []nodeapi.Node

	vectorOutByNode map[string]chan TiltPanel.TiltVectorMsg
	vectorInByNode  map[string]chan TiltPanel.TiltVectorMsg
}

func buildFromSpec(ctx context.Context, spec loadspec.TopoSpec, clk clock.Clock, sphere polar.SceneSphere, hasScene bool, scenePath string) ([]nodeapi.Node, beadanimation.SlotRegistry, *scenerun.MoveDispatch, SliderPanel.Sinks, error) {
	b := &buildCtx{ctx: ctx, spec: spec, clk: clk, sphere: sphere, hasScene: hasScene, scenePath: scenePath}

	b.nodeGeoms, b.baseIndices, b.dragIndices = b.spec.SeedGeometry(b.sphere.Center)
	b.wiring.DestRun, b.wiring.EdgeRun, b.edgeEndpoints = b.spec.AllocateBeadLines(b.nodeGeoms)
	b.vectorOutByNode, b.vectorInByNode = b.spec.AllocateVectorChannels()
	if err := b.buildMoveDispatch(); err != nil {
		return nil, nil, nil, SliderPanel.Sinks{}, err
	}
	b.nodeType, b.kindBroadcastPorts = kindreg.BuildTypeMaps(b.spec)
	b.wiring.Inbound, b.wiring.Outbound, b.wiring.OutboundHandle = b.spec.BuildEdgeMaps(b.nodeType, b.kindBroadcastPorts)
	if err := b.buildNodes(); err != nil {
		return nil, nil, nil, SliderPanel.Sinks{}, err
	}

	bindDispatch(b.md, b.outSink, b.wiring.DestRun)

	return b.nodes, beadanimation.SlotRegistry(b.wiring.DestRun), b.md, b.speedSinks, nil
}
