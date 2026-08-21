package runtopology

import (
	T "github.com/dtauraso/wirefold/src/Trace"
	"context"
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/src/Scene/scene"
	"github.com/dtauraso/wirefold/src/Scene/scenepaths"

	clock "github.com/dtauraso/wirefold/src/Clock"

	bead "github.com/dtauraso/wirefold/src/Ring/Bead"

	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
	NodeShape "github.com/dtauraso/wirefold/src/Ring/NodeShape"
	SW "github.com/dtauraso/wirefold/src/runtopology/streamwire"
)

func RunTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {

	streamFDs := SW.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(SW.StreamKindView, 0)

	sceneTabNames := Tabs.TabNames()
	sceneTabSelected := Tabs.SelectedIndex(topologyPath)
	scenePath := scene.ResolvePath(topologyPath)
	T.SetSceneRoot(scenePath)
	nodes, slotReg, md, speedSinks, err := LoadTopology(ctx, scenePath, clk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load topology: %v\n", err)
		os.Exit(1)
	}
	wireEdgeStreams(streamFDs, md, scenePath)
	wireNodeStreams(streamFDs, md, scenePath)
	md.UI.OwnerCounts.Nodes = int32(len(md.RT.NodeRowTable))
	md.UI.OwnerCounts.Edges = int32(len(md.RT.EdgeRowTable))
	md.UI.SceneTabNames = sceneTabNames
	md.UI.SceneTabSelected = sceneTabSelected
	md.UI.SetSceneRoot(scenePath)
	md.UI.WriteRingSurfaces(NodeShape.CanonicalRingSurfacePointsFlat(), bead.CanonicalRingSurfacePointsFlat())
	wireViewStream(md, viewFile, viewStreamWired)
	emitStartupBreadcrumbs(md, scenePath, len(nodes))
	checkRowSeedCount(md, len(nodes))
	loadSceneState(scenePath, md, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Quit = cancel
	md.Scenes.Loaded = sceneTabSelected

	moverWG := md.Start(ctx)

	stdinWG, gestureWG := startStdinReader(ctx, cancel, slotReg, md, speedSinks, clk, scenepaths.InputDirPath(scenePath))
	wg := launchNodeGoroutines(ctx, nodes)
	joinAll(wg, moverWG, stdinWG, gestureWG)
}
