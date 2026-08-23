package Startup

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Scene/Wiring"

	"github.com/dtauraso/wirefold/Categories/Scene/Topology"

	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"

	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	_ "github.com/dtauraso/wirefold/Categories/NodeKinds"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/Dispatch"
	"github.com/dtauraso/wirefold/Categories/Scene/View"
)

type Scene struct {
	Nodes      []Wiring.BuiltNode
	Dispatch   *Dispatch.MoveDispatch
	SpeedSinks SliderPanel.Sinks
}

func Load(ctx context.Context, scenePath string, clk clock.Clock) (Scene, error) {
	spec, err := Topology.ParseSpec(scenePath)
	if err != nil {
		return Scene{}, err
	}
	if err := Wiring.ValidateSpec(&spec, Wiring.KindPorts); err != nil {
		return Scene{}, err
	}

	sphere, hasScene := LoadSceneSphere(scenePath)

	nodeGeoms, baseIndices, dragIndices := Wiring.SeedGeometry(spec, Topology.Vec3(sphere.Center))
	destRun, edgeRun, edgeEndpoints := Wiring.AllocateBeadLines(spec, nodeGeoms)
	vectorOut, vectorIn := Wiring.AllocateVectorChannels(spec)

	var speedSinks SliderPanel.Sinks
	md, err := Wiring.NewFromSpec(spec, sphere, hasScene, scenePath, clk, &speedSinks,
		nodeGeoms, edgeEndpoints, baseIndices, dragIndices)
	if err != nil {
		return Scene{}, err
	}

	nodeType, kindBroadcastPorts := Wiring.BuildTypeMaps(spec)
	inbound, outbound, outboundHandle := Wiring.BuildEdgeMaps(spec, nodeType, kindBroadcastPorts)
	wiring := Wiring.EdgeWiring{
		Inbound: inbound, Outbound: outbound, OutboundHandle: outboundHandle,
		DestRun: destRun, EdgeRun: edgeRun,
	}

	nodes, outSink, err := Wiring.BuildNodes(ctx, spec, md, wiring, nodeGeoms, vectorOut, vectorIn, clk, &speedSinks)
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

func LoadSceneState(scenePath string, md *Dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	Camera.SeedInitialViewpoint(scenePath, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint)

	s := Scenes.For(scenePath)
	md.UI.SceneEditable = s.Editable
	md.UI.SceneKinds = SceneKindMask(s)

	InstallOverlays(&md.UI, scenePath)

	InstallPanels(&md.UI, scenePath)

	InstallSpeed(&md.UI, scenePath, speedSinks)

	EnableViewpointPersist(&md.UI, scenePath)

	EnableEditPersist(&md.UI, &md.Scenes, md.MR.NodeGeoms(), scenePath)

	InstallSceneSphere(&md.UI, &md.GS, scenePath)
}

func EmitStartupBreadcrumbs(md *Dispatch.MoveDispatch, scenePath string, nodeCount int) {

	md.UI.EmitBreadcrumb(View.RowEvent{
		Label: Wiring.BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func CheckRowSeedCount(md *Dispatch.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {

		md.UI.EmitBreadcrumb(View.RowEvent{
			Label: Wiring.BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(len(md.GS.NodeSeedsFn())), X: float64(nodeCount),
		})
	}
}

func SceneKindMask(s Scenes.Scene) uint32 {
	if len(s.Kinds) == 0 {
		return ^uint32(0)
	}
	var mask uint32
	for _, k := range s.Kinds {
		if id := NodeBuf.NodeKindID(k); id != NodeBuf.KindIDUnknown {
			mask |= 1 << uint(id)
		}
	}
	return mask
}
