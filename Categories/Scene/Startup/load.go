package Startup

import (
	"context"

	Ports "github.com/dtauraso/beadnetwork/Categories/Node/Ports"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Topology"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	_ "github.com/dtauraso/beadnetwork/Categories/NodeKinds"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Dispatch"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

type Scene struct {
	Nodes      []BuiltNode
	Dispatch   *Dispatch.MoveDispatch
	SpeedSinks SliderPanel.Sinks
}

func Load(ctx context.Context, scenePath string, clk clock.Clock) (Scene, error) {
	spec, err := Topology.ParseSpec(scenePath)
	if err != nil {
		return Scene{}, err
	}
	if err := Topology.ValidateSpec(&spec, Ports.KindPortSets()); err != nil {
		return Scene{}, err
	}

	sphere, hasScene := LoadSceneSphere(scenePath)

	nodeGeoms, baseIndices, dragIndices := Topology.SeedGeometry(spec, Topology.Vec3(sphere.Center))
	destRun, edgeRun, edgeEndpoints := Topology.AllocateBeadLines(spec, nodeGeoms)
	vectorOut, vectorIn := Topology.AllocateVectorChannels(spec)

	var speedSinks SliderPanel.Sinks
	md, err := NewFromSpec(spec, sphere, hasScene, scenePath, clk, &speedSinks,
		nodeGeoms, edgeEndpoints, baseIndices, dragIndices)
	if err != nil {
		return Scene{}, err
	}

	nodeType := Topology.NodeTypes(spec)
	kindBroadcastPorts := Ports.KindPortSets().Broadcast
	inbound, outbound, outboundHandle := Topology.BuildEdgeMaps(spec, nodeType, kindBroadcastPorts)
	lines := Ports.EdgeLines{
		Inbound: inbound, Outbound: outbound, OutboundHandle: outboundHandle,
		DestRun: destRun, EdgeRun: edgeRun,
	}

	nodes, outSink, err := BuildNodes(ctx, spec, md, lines, nodeGeoms, vectorOut, vectorIn, clk, &speedSinks)
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
	md.UI.SceneNodesDraggable = s.NodesDraggable
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
		Label: BreadcrumbTopologyLoaded, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(nodeCount), Text: scenePath,
	})
}

func CheckRowSeedCount(md *Dispatch.MoveDispatch, nodeCount int) {
	if len(md.GS.NodeSeedsFn()) != nodeCount {

		md.UI.EmitBreadcrumb(View.RowEvent{
			Label: BreadcrumbRowSeedCountMismatch, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
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
