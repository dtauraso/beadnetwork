package runtopology

import (
	"context"
	"fmt"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"os"

	"github.com/dtauraso/wirefold/nodes/clock"

	"github.com/dtauraso/wirefold/nodes/bead"

	Bld "github.com/dtauraso/wirefold/nodes/Wiring/build"
	SF "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/streamframe"
	NodeShape "github.com/dtauraso/wirefold/tools/topology-vscode/Node/Shape"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Tabs"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func RunTopology(ctx context.Context, cancel context.CancelFunc, topologyPath string, clk clock.Clock) {

	streamFDs := SF.ParseStreamFDs(os.Getenv("WIREFOLD_STREAM_FDS"))
	viewFile, viewStreamWired := streamFDs.Open(SF.StreamKindView, 0)

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
	cols := SF.NewColumnStreams(streamFDs, len(md.RT.NodeRowTable), len(md.RT.EdgeRowTable))
	md.UI.SetSingletonColumns(cols.SingletonColumns())
	md.UI.WriteNodeRingSurfaceColumns(NodeShape.CanonicalRingSurfacePointsFlat())
	md.UI.WriteBeadRingSurfaceColumns(bead.CanonicalRingSurfacePointsFlat())
	wireViewStream(md, viewFile, viewStreamWired, sceneTabNames, sceneTabSelected)
	emitStartupBreadcrumbs(tr, md, scenePath, len(nodes))
	checkRowSeedCount(tr, md, len(nodes))
	loadSceneState(scenePath, md, tr, speedSinks)

	md.Scenes.AnchorPath = topologyPath
	md.Scenes.Quit = cancel

	moverWG := md.Start(ctx)

	stdinWG, gestureWG := startStdinReader(ctx, cancel, slotReg, md, tr, speedSinks)
	wg := launchNodeGoroutines(ctx, nodes)
	joinAll(wg, moverWG, stdinWG, gestureWG)
}
