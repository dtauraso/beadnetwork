package scenebuild

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	_ "github.com/dtauraso/wirefold/Categories/NodeKinds"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/scene"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepersist"
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
	"github.com/dtauraso/wirefold/Categories/Scene/viewpersist"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

type Scene struct {
	Nodes      []BuiltNode
	Dispatch   *scenerun.MoveDispatch
	SpeedSinks SliderPanel.Sinks
}

func Load(ctx context.Context, scenePath string, clk clock.Clock) (Scene, error) {
	spec, err := ParseSpec(scenePath)
	if err != nil {
		return Scene{}, err
	}
	if err := ValidateSpec(&spec, KindPorts); err != nil {
		return Scene{}, err
	}

	sphere, hasScene := scenepersist.LoadSceneSphere(scenePath)

	nodeGeoms, baseIndices, dragIndices := SeedGeometry(spec, Vec3(sphere.Center))
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

func LoadSceneState(scenePath string, md *scenerun.MoveDispatch, speedSinks SliderPanel.Sinks) {
	Camera.SeedInitialViewpoint(scenePath, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint)

	s := scene.For(scenePath)
	md.UI.SceneEditable = s.Editable
	md.UI.SceneKinds = s.KindMask()

	scenepersist.InstallOverlays(&md.UI, scenePath)

	scenepersist.InstallPanels(&md.UI, scenePath)

	scenepersist.InstallSpeed(&md.UI, scenePath, speedSinks)

	viewpersist.EnableViewpointPersist(&md.Persist, &md.UI, scenePath)

	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, md.MR.NodeGeoms(), scenePath)

	scenepersist.InstallSceneSphere(&md.UI, &md.GS, scenePath)
}

func EmitStartupBreadcrumbs(md *scenerun.MoveDispatch, scenePath string, nodeCount int) {

	md.UI.EmitBreadcrumb(viewstate.RowEvent{
		Label: BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func CheckRowSeedCount(md *scenerun.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {

		md.UI.EmitBreadcrumb(viewstate.RowEvent{
			Label: BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}
