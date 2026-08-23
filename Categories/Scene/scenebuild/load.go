package scenebuild

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	_ "github.com/dtauraso/wirefold/Categories/NodeKinds"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepersist"
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
)

type Scene struct {
	Nodes      []BuiltNode
	Dispatch   *scenerun.MoveDispatch
	SpeedSinks SliderPanel.Sinks
}

func Load(ctx context.Context, scenePath string, clk clock.Clock) (Scene, error) {
	spec, err := loadspec.ParseSpec(scenePath)
	if err != nil {
		return Scene{}, err
	}
	if err := ValidateSpec(&spec, KindPorts); err != nil {
		return Scene{}, err
	}

	sphere, hasScene := scenepersist.LoadSceneSphere(scenePath)

	nodeGeoms, baseIndices, dragIndices := SeedGeometry(spec, loadspec.Vec3(sphere.Center))
	destRun, edgeRun, edgeEndpoints := AllocateBeadLines(spec, nodeGeoms)
	vectorOut, vectorIn := AllocateVectorChannels(spec)

	var speedSinks SliderPanel.Sinks
	md, err := NewFromSpec(spec, sphere, hasScene, scenePath, clk, &speedSinks,
		nodeGeoms, edgeEndpoints, baseIndices, dragIndices)
	if err != nil {
		return Scene{}, err
	}

	nodeType, kindBroadcastPorts := BuildTypeMaps(spec)
	inbound, outbound, outboundHandle := BuildEdgeMaps(spec, nodeType, kindBroadcastPorts)
	wiring := EdgeWiring{
		Inbound: inbound, Outbound: outbound, OutboundHandle: outboundHandle,
		DestRun: destRun, EdgeRun: edgeRun,
	}

	nodes, outSink, err := buildNodes(ctx, spec, md, wiring, nodeGeoms, vectorOut, vectorIn, clk, &speedSinks)
	if err != nil {
		return Scene{}, err
	}

	md.MR.Bind(outSink, destRun, md.RT.EdgeRowForPair)

	return Scene{
		Nodes:      nodes,
		Dispatch:   md,
		SpeedSinks: speedSinks,
	}, nil
}
