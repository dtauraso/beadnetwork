package runtopology

import (
	"context"
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/src/Scene/scene"

	"github.com/dtauraso/wirefold/src/Clock"

	"github.com/dtauraso/wirefold/src/Ring/Bead"

	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
	SW "github.com/dtauraso/wirefold/src/Node/Wiring/streamwire"
	NodeShape "github.com/dtauraso/wirefold/src/Ring/NodeShape"
)

func RunTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {

	streamFDs := SW.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(SW.StreamKindView, 0)

	sceneTabNames := Tabs.TabNames()
	sceneTabSelected := Tabs.SelectedIndex(topologyPath)
	scenePath := scene.ResolvePath(topologyPath)
	nodes, slotReg, md, speedSinks, err := LoadTopology(ctx, scenePath, clk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load topology: %v\n", err)
		os.Exit(1)
	}
	wireEdgeStreams(streamFDs, md)
	wireNodeStreams(streamFDs, md)
	cols := SW.NewColumnStreams(streamFDs, len(md.RT.NodeRowTable), len(md.RT.EdgeRowTable))
	md.UI.OwnerCounts.Nodes = int32(len(md.RT.NodeRowTable))
	md.UI.OwnerCounts.Edges = int32(len(md.RT.EdgeRowTable))
	md.UI.SceneTabNames = sceneTabNames
	md.UI.SceneTabSelected = sceneTabSelected
	md.UI.SetSingletonColumns(cols.SingletonColumns())
	md.UI.WriteNodeRingSurfaceColumns(NodeShape.CanonicalRingSurfacePointsFlat())
	md.UI.WriteBeadRingSurfaceColumns(bead.CanonicalRingSurfacePointsFlat())
	wireViewStream(md, viewFile, viewStreamWired, sceneTabNames, sceneTabSelected)
	emitStartupBreadcrumbs(md, scenePath, len(nodes))
	checkRowSeedCount(md, len(nodes))
	loadSceneState(scenePath, md, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Quit = cancel
	md.Scenes.Loaded = sceneTabSelected

	moverWG := md.Start(ctx)

	stdinWG, gestureWG := startStdinReader(ctx, cancel, slotReg, md, speedSinks)
	wg := launchNodeGoroutines(ctx, nodes)
	joinAll(wg, moverWG, stdinWG, gestureWG)
}
