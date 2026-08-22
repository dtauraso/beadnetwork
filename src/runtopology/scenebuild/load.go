package scenebuild

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/src/Clock"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
	"github.com/dtauraso/wirefold/src/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
	"github.com/dtauraso/wirefold/src/runtopology/loadspec"
	"github.com/dtauraso/wirefold/src/runtopology/scenerun"
)

type Scene struct {
	Nodes      []nodeapi.Node
	SlotReg    beadanimation.SlotRegistry
	Dispatch   *scenerun.MoveDispatch
	SpeedSinks SliderPanel.Sinks
}

func Load(ctx context.Context, scenePath string, clk clock.Clock) (Scene, error) {
	kindreg.BuildRegistry()

	spec, err := loadspec.ParseSpec(scenePath)
	if err != nil {
		return Scene{}, err
	}
	kindPorts := make(map[string][]portwiring.PortSpec, len(kindreg.Registry))
	for kind, bind := range kindreg.Registry {
		kindPorts[kind] = bind.Ports
	}
	if err := loadspec.ValidateSpec(&spec, kindPorts); err != nil {
		return Scene{}, err
	}

	sphere, hasScene := scenepersist.LoadSceneSphere(scenePath)

	nodeGeoms, baseIndices, dragIndices := spec.SeedGeometry(sphere.Center)
	destRun, edgeRun, edgeEndpoints := spec.AllocateBeadLines(nodeGeoms)
	vectorOut, vectorIn := spec.AllocateVectorChannels()

	var speedSinks SliderPanel.Sinks
	md, err := NewFromSpec(spec, sphere, hasScene, scenePath, clk, &speedSinks,
		nodeGeoms, edgeEndpoints, baseIndices, dragIndices)
	if err != nil {
		return Scene{}, err
	}

	nodeType, kindBroadcastPorts := kindreg.BuildTypeMaps(spec)
	inbound, outbound, outboundHandle := spec.BuildEdgeMaps(nodeType, kindBroadcastPorts)
	wiring := kindreg.EdgeWiring{
		Inbound: inbound, Outbound: outbound, OutboundHandle: outboundHandle,
		DestRun: destRun, EdgeRun: edgeRun,
	}

	nodes, outSink, err := buildNodes(ctx, spec, md, wiring, nodeGeoms, vectorOut, vectorIn, clk, &speedSinks)
	if err != nil {
		return Scene{}, err
	}

	md.MR.Bind(outSink, beadanimation.SlotRegistry(destRun), md.RT.EdgeRowForPair)

	return Scene{
		Nodes:      nodes,
		SlotReg:    beadanimation.SlotRegistry(destRun),
		Dispatch:   md,
		SpeedSinks: speedSinks,
	}, nil
}
