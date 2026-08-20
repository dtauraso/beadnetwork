package runtopology

import (
	"context"
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/src/Node/Wiring/scene"

	"github.com/dtauraso/wirefold/src/Node/clock"

	"github.com/dtauraso/wirefold/src/Bead"

	NodeShape "github.com/dtauraso/wirefold/src/Node/Shape"
	Bld "github.com/dtauraso/wirefold/src/Node/Wiring/build"
	SW "github.com/dtauraso/wirefold/src/Node/Wiring/streamwire"
	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func RunTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {

	streamFDs := SW.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(SW.StreamKindView, 0)

	tr := T.New()

	sceneTabNames := Tabs.TabNames()
	sceneTabSelected := Tabs.SelectedIndex(topologyPath)
	scenePath := scene.ResolvePath(topologyPath)
	nodes, slotReg, md, speedSinks, err := Bld.LoadTopology(ctx, scenePath, tr, clk)
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
	emitStartupBreadcrumbs(tr, md, scenePath, len(nodes))
	checkRowSeedCount(tr, md, len(nodes))
	loadSceneState(scenePath, md, tr, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Quit = cancel
	md.Scenes.Loaded = sceneTabSelected

	moverWG := md.Start(ctx)

	stdinWG, gestureWG := startStdinReader(ctx, cancel, slotReg, md, tr, speedSinks)
	wg := launchNodeGoroutines(ctx, nodes)
	joinAll(wg, moverWG, stdinWG, gestureWG)
}
